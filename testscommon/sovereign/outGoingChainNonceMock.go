package sovereign

import dtoCore "github.com/multiversx/mx-chain-core-go/data/sovereign/dto"

// OutGoingChainNonceMock -
type OutGoingChainNonceMock struct {
	GetAndIncrementNonceCalled func(chainID dtoCore.ChainID) (uint64, error)
	GetNonceCalled             func(chainID dtoCore.ChainID) uint64
	SetNonceCalled             func(chainID dtoCore.ChainID, nonce uint64)
}

// GetAndIncrementNonce -
func (mock *OutGoingChainNonceMock) GetAndIncrementNonce(chainID dtoCore.ChainID) (uint64, error) {
	if mock.GetAndIncrementNonceCalled != nil {
		return mock.GetAndIncrementNonceCalled(chainID)
	}
	return 0, nil
}

// GetNonce -
func (mock *OutGoingChainNonceMock) GetNonce(chainID dtoCore.ChainID) uint64 {
	if mock.GetNonceCalled != nil {
		return mock.GetNonceCalled(chainID)
	}
	return 0
}

// SetNonce -
func (mock *OutGoingChainNonceMock) SetNonce(chainID dtoCore.ChainID, nonce uint64) {
	if mock.SetNonceCalled != nil {
		mock.SetNonceCalled(chainID, nonce)
	}
}

// IsInterfaceNil -
func (mock *OutGoingChainNonceMock) IsInterfaceNil() bool {
	return mock == nil
}
