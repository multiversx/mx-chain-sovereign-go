package sovereign

import (
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
)

// TopicsCheckerMock -
type TopicsCheckerMock struct {
	CheckValidityCalled func(topics [][]byte, transferData *sovereign.TransferData) error
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
