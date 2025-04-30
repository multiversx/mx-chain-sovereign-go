package sovereign

import (
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

func NewShardedOutGoingOperationPool(expiryTime time.Duration) *shardedOutGoingOperationPool {
	return &shardedOutGoingOperationPool{
		timeout: expiryTime,
		cache:   map[dto.ChainID]OutGoingOperationsPool{},
	}
}

// Add -
func (sop *shardedOutGoingOperationPool) Add(data *sovereign.BridgeOutGoingData, chainID dto.ChainID) {
	sop.mutex.Lock()
	defer sop.mutex.Unlock()

	shardedPool, exists := sop.cache[chainID]
	if !exists {
		shardedPool = NewOutGoingOperationPool(sop.timeout)
		sop.cache[chainID] = shardedPool
	}

	shardedPool.Add(data)
}

// Get -
func (sop *shardedOutGoingOperationPool) Get(hash []byte, chainID dto.ChainID) *sovereign.BridgeOutGoingData {
	sop.mutex.Lock()
	defer sop.mutex.Unlock()

	shardedPool, exists := sop.cache[chainID]
	if !exists {
		return nil
	}

	return shardedPool.Get(hash)
}

// Delete -
func (sop *shardedOutGoingOperationPool) Delete(hash []byte, chainID dto.ChainID) {
	sop.mutex.Lock()
	defer sop.mutex.Unlock()

	shardedPool, exists := sop.cache[chainID]
	if !exists {
		return
	}

	shardedPool.Delete(hash)
}

// GetUnconfirmedOperations -
func (sop *shardedOutGoingOperationPool) GetUnconfirmedOperations() []*sovereign.BridgeOutGoingData {
	ret := make([]*sovereign.BridgeOutGoingData, 0)

	sop.mutex.Lock()
	defer sop.mutex.Unlock()

	for _, shardedPool := range sop.cache {
		shardedUnconfirmedOps := shardedPool.GetUnconfirmedOperations()
		if len(shardedUnconfirmedOps) != 0 {
			ret = append(ret, shardedUnconfirmedOps...)
		}
	}

	return ret
}

// ConfirmOperation -
func (sop *shardedOutGoingOperationPool) ConfirmOperation(hashOfHashes []byte, hash []byte, chainID dto.ChainID) error {
	sop.mutex.Lock()
	defer sop.mutex.Unlock()

	shardedPool, exists := sop.cache[chainID]
	if !exists {
		// todo: here error maybe + optimize everywhere mutex usage
		return nil
	}

	return shardedPool.ConfirmOperation(hashOfHashes, hash)
}

// ResetTimer -
func (sop *shardedOutGoingOperationPool) ResetTimer(hashes [][]byte, chainID dto.ChainID) {
	sop.mutex.Lock()
	defer sop.mutex.Unlock()

	shardedPool, exists := sop.cache[chainID]
	if !exists {
		return
	}

	shardedPool.ResetTimer(hashes)
}

// IsInterfaceNil -
func (sop *shardedOutGoingOperationPool) IsInterfaceNil() bool {
	return sop == nil
}
