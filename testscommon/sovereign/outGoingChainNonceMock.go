package sovereign

import dtoCore "github.com/multiversx/mx-chain-core-go/data/sovereign/dto"

// OutGoingChainNonceMock -
type OutGoingChainNonceMock struct {
	GetNonceCalled       func(chainID dtoCore.ChainID) (uint64, error)
	IncrementNonceCalled func(chainID dtoCore.ChainID, delta uint64)
}

// GetNonce -
func (mock *OutGoingChainNonceMock) GetNonce(chainID dtoCore.ChainID) (uint64, error) {
	if mock.GetNonceCalled != nil {
		return mock.GetNonceCalled(chainID)
	}
	return 0, nil
}

// IncrementNonce -
func (mock *OutGoingChainNonceMock) IncrementNonce(chainID dtoCore.ChainID, delta uint64) {
	if mock.IncrementNonceCalled != nil {
		mock.IncrementNonceCalled(chainID, delta)
	}
}
