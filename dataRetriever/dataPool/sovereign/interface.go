package sovereign

import (
	sovereignCore "github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/multiversx/mx-chain-core-go/data/sovereign/dto"
)

// OutGoingOperationsPool defines the behavior of a timed cache for outgoing operations
type OutGoingOperationsPool interface {
	Add(data *sovereignCore.BridgeOutGoingData)
	Get(hash []byte) *sovereignCore.BridgeOutGoingData
	Delete(hash []byte)
	GetUnconfirmedOperations() []*sovereignCore.BridgeOutGoingData
	ResetTimer(hashes [][]byte)
	ConfirmOperation(hashOfHashes []byte, hash []byte) error
	IsInterfaceNil() bool
}

// ShardedOutGoingOperationPool defines a sharded(per chain) outgoing operation pool
type ShardedOutGoingOperationPool interface {
	Add(data *sovereignCore.BridgeOutGoingData, chainID dto.ChainID)
	Get(hash []byte, chainID dto.ChainID) *sovereignCore.BridgeOutGoingData
	Delete(hash []byte, chainID dto.ChainID)
	GetUnconfirmedOperations() []*sovereignCore.BridgeOutGoingData
	ConfirmOperation(hashOfHashes []byte, hash []byte, chainID dto.ChainID) error
	ResetTimer(hashes [][]byte, chainID dto.ChainID)
	IsInterfaceNil() bool
}
