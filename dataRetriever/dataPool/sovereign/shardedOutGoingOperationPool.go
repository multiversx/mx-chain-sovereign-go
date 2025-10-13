package sovereign

import (
	"fmt"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/multiversx/mx-chain-core-go/data/sovereign/dto"
)

type shardedOutGoingOperationPool struct {
	mutex   sync.RWMutex
	timeout time.Duration
	cache   map[dto.ChainID]OutGoingOperationsPool
}

// NewShardedOutGoingOperationPool creates a new sharded outgoing operation pool with a specified timeout.
// Each chain id is cached as a different "shard"
func NewShardedOutGoingOperationPool(expiryTime time.Duration) *shardedOutGoingOperationPool {
	return &shardedOutGoingOperationPool{
		timeout: expiryTime,
		cache:   map[dto.ChainID]OutGoingOperationsPool{},
	}
}

// Add inserts a new outgoing operation for a given chainID. If the chain id is not found in the cache, it creates
// a new entry for it
func (sop *shardedOutGoingOperationPool) Add(data *sovereign.BridgeOutGoingData, chainID dto.ChainID) {
	sop.mutex.RLock()
	shardedPool, exists := sop.cache[chainID]
	sop.mutex.RUnlock()

	if !exists {
		sop.mutex.Lock()
		shardedPool, exists = sop.cache[chainID]
		if !exists {
			shardedPool = NewOutGoingOperationPool(sop.timeout)
			sop.cache[chainID] = shardedPool
		}
		sop.mutex.Unlock()
	}

	shardedPool.Add(data)
}

// Get retrieves a specific outgoing operation by its hash and chainID.
func (sop *shardedOutGoingOperationPool) Get(hash []byte, chainID dto.ChainID) *sovereign.BridgeOutGoingData {
	sop.mutex.RLock()
	shardedPool, exists := sop.cache[chainID]
	sop.mutex.RUnlock()

	if !exists {
		return nil
	}

	return shardedPool.Get(hash)
}

// Delete removes an outgoing operation identified by hash and chainID, if found.
func (sop *shardedOutGoingOperationPool) Delete(hash []byte, chainID dto.ChainID) {
	sop.mutex.RLock()
	shardedPool, exists := sop.cache[chainID]
	sop.mutex.RUnlock()

	if !exists {
		return
	}

	shardedPool.Delete(hash)
}

// GetUnconfirmedOperations retrieves all unconfirmed outgoing operations from all chains.
func (sop *shardedOutGoingOperationPool) GetUnconfirmedOperations() []*sovereign.BridgeOutGoingData {
	ret := make([]*sovereign.BridgeOutGoingData, 0)

	sop.mutex.RLock()
	defer sop.mutex.RUnlock()

	for _, shardedPool := range sop.cache {
		shardedUnconfirmedOps := shardedPool.GetUnconfirmedOperations()
		if len(shardedUnconfirmedOps) != 0 {
			ret = append(ret, shardedUnconfirmedOps...)
		}
	}

	return ret
}

// ConfirmOperation marks an operation as confirmed by its hash and chainID.
func (sop *shardedOutGoingOperationPool) ConfirmOperation(hashOfHashes []byte, hash []byte, chainID dto.ChainID) error {
	sop.mutex.RLock()
	shardedPool, exists := sop.cache[chainID]
	sop.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("%w for chain: %s, hashOfHashes: %x, hash: %x ",
			errChainIDNotFound, chainID.String(), hashOfHashes, hash)
	}

	return shardedPool.ConfirmOperation(hashOfHashes, hash)
}

// ResetTimer resets the timeout for the specified hashes on a given chainID.
func (sop *shardedOutGoingOperationPool) ResetTimer(hashes [][]byte, chainID dto.ChainID) {
	sop.mutex.RLock()
	shardedPool, exists := sop.cache[chainID]
	sop.mutex.RUnlock()

	if !exists {
		return
	}

	shardedPool.ResetTimer(hashes)
}

// IsInterfaceNil returns true if the underlying interface is nil.
func (sop *shardedOutGoingOperationPool) IsInterfaceNil() bool {
	return sop == nil
}
