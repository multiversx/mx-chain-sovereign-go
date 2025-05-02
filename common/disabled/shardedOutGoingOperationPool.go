package disabled

import (
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/multiversx/mx-chain-core-go/data/sovereign/dto"
)

type shardedOutGoingOperationsPool struct {
}

// NewDisabledShardedOutGoingOperationPool creates a disabled sharded outgoing operation pool
func NewDisabledShardedOutGoingOperationPool() *shardedOutGoingOperationsPool {
	return &shardedOutGoingOperationsPool{}
}

// Add does nothing
func (op *shardedOutGoingOperationsPool) Add(_ *sovereign.BridgeOutGoingData, _ dto.ChainID) {}

// Get does nothing
func (op *shardedOutGoingOperationsPool) Get(_ []byte, _ dto.ChainID) *sovereign.BridgeOutGoingData {
	return &sovereign.BridgeOutGoingData{}
}

// Delete does nothing
func (op *shardedOutGoingOperationsPool) Delete(_ []byte, _ dto.ChainID) {}

// ConfirmOperation does nothing
func (op *shardedOutGoingOperationsPool) ConfirmOperation(_ []byte, _ []byte, _ dto.ChainID) error {
	return nil
}

// GetUnconfirmedOperations does nothing
func (op *shardedOutGoingOperationsPool) GetUnconfirmedOperations() []*sovereign.BridgeOutGoingData {
	return make([]*sovereign.BridgeOutGoingData, 0)
}

// ResetTimer does nothing
func (op *shardedOutGoingOperationsPool) ResetTimer(_ [][]byte, _ dto.ChainID) {
}

// IsInterfaceNil checks if the underlying pointer is nil
func (op *shardedOutGoingOperationsPool) IsInterfaceNil() bool {
	return op == nil
}
