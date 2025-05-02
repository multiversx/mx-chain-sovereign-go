package sovereign

import (
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/multiversx/mx-chain-core-go/data/sovereign/dto"
)

// ShardedOutGoingOperationsPoolMock -
type ShardedOutGoingOperationsPoolMock struct {
	AddCalled                      func(data *sovereign.BridgeOutGoingData, chainID dto.ChainID)
	GetCalled                      func(hash []byte, chainID dto.ChainID) *sovereign.BridgeOutGoingData
	DeleteCalled                   func(hash []byte, chainID dto.ChainID)
	GetUnconfirmedOperationsCalled func() []*sovereign.BridgeOutGoingData
	ConfirmOperationCalled         func(hashOfHashes []byte, hash []byte, chainID dto.ChainID) error
	ResetTimerCalled               func(hashes [][]byte, chainID dto.ChainID)
}

// Add -
func (mock *ShardedOutGoingOperationsPoolMock) Add(data *sovereign.BridgeOutGoingData, chainID dto.ChainID) {
	if mock.AddCalled != nil {
		mock.AddCalled(data, chainID)
	}
}

// Get -
func (mock *ShardedOutGoingOperationsPoolMock) Get(hash []byte, chainID dto.ChainID) *sovereign.BridgeOutGoingData {
	if mock.GetCalled != nil {
		return mock.GetCalled(hash, chainID)
	}
	return nil
}

// Delete -
func (mock *ShardedOutGoingOperationsPoolMock) Delete(hash []byte, chainID dto.ChainID) {
	if mock.DeleteCalled != nil {
		mock.DeleteCalled(hash, chainID)
	}
}

// GetUnconfirmedOperations -
func (mock *ShardedOutGoingOperationsPoolMock) GetUnconfirmedOperations() []*sovereign.BridgeOutGoingData {
	if mock.GetUnconfirmedOperationsCalled != nil {
		return mock.GetUnconfirmedOperationsCalled()
	}
	return nil
}

// ConfirmOperation -
func (mock *ShardedOutGoingOperationsPoolMock) ConfirmOperation(hashOfHashes []byte, hash []byte, chainID dto.ChainID) error {
	if mock.ConfirmOperationCalled != nil {
		return mock.ConfirmOperationCalled(hashOfHashes, hash, chainID)
	}
	return nil
}

// ResetTimer -
func (mock *ShardedOutGoingOperationsPoolMock) ResetTimer(hashes [][]byte, chainID dto.ChainID) {
	if mock.ResetTimerCalled != nil {
		mock.ResetTimerCalled(hashes, chainID)
	}
}

// IsInterfaceNil -
func (mock *ShardedOutGoingOperationsPoolMock) IsInterfaceNil() bool {
	return mock == nil
}
