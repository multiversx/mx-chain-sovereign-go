package sovereign

import dtoCore "github.com/multiversx/mx-chain-core-go/data/sovereign/dto"

// OutGoingChainNonceMock -
type OutGoingChainNonceMock struct {
	GetAndIncrementNonceCalled func(chainID dtoCore.ChainID) (uint64, error)
	IncrementNonceCalled       func(chainID dtoCore.ChainID, delta uint64)
}

// GetAndIncrementNonce -
func (mock *OutGoingChainNonceMock) GetAndIncrementNonce(chainID dtoCore.ChainID) (uint64, error) {
	if mock.GetAndIncrementNonceCalled != nil {
		return mock.GetAndIncrementNonceCalled(chainID)
	}
	return 0, nil
}

// IncrementNonce -
func (mock *OutGoingChainNonceMock) IncrementNonce(chainID dtoCore.ChainID, delta uint64) {
	if mock.IncrementNonceCalled != nil {
		mock.IncrementNonceCalled(chainID, delta)
	}
}

// IsInterfaceNil -
func (mock *OutGoingChainNonceMock) IsInterfaceNil() bool {
	return mock == nil
}
