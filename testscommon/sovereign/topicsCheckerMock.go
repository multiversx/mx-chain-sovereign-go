package sovereign

import (
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
)

// TopicsCheckerMock -
type TopicsCheckerMock struct {
	CheckDepositTokensValidityCalled func(topics [][]byte) error
	CheckScCallValidityCalled        func(topics [][]byte, transferData *sovereign.TransferData) error
	CheckValidityCalled              func(topics [][]byte, transferData *sovereign.TransferData) error
}

// CheckDepositTokensValidity -
func (tc *TopicsCheckerMock) CheckDepositTokensValidity(topics [][]byte) error {
	if tc.CheckDepositTokensValidityCalled != nil {
		return tc.CheckDepositTokensValidityCalled(topics)
	}

	return nil
}

// CheckScCallValidity -
func (tc *TopicsCheckerMock) CheckScCallValidity(topics [][]byte, transferData *sovereign.TransferData) error {
	if tc.CheckScCallValidityCalled != nil {
		return tc.CheckScCallValidityCalled(topics, transferData)
	}

	return nil
}

// CheckValidity -
func (tc *TopicsCheckerMock) CheckValidity(topics [][]byte, transferData *sovereign.TransferData) error {
	if tc.CheckValidityCalled != nil {
		return tc.CheckValidityCalled(topics, transferData)
	}

	return nil
}

// IsInterfaceNil -
func (tc *TopicsCheckerMock) IsInterfaceNil() bool {
	return tc == nil
}
