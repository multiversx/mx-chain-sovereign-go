package disabled

import (
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
)

type topicsChecker struct {
}

// NewDisabledTopicsChecker -
func NewDisabledTopicsChecker() *topicsChecker {
	return &topicsChecker{}
}

// CheckDepositTokensValidity -
func (tc *topicsChecker) CheckDepositTokensValidity(_ [][]byte) error {
	return nil
}

// CheckScCallValidity -
func (tc *topicsChecker) CheckScCallValidity(_ [][]byte, _ *sovereign.TransferData) error {
	return nil
}

// CheckValidity -
func (tc *topicsChecker) CheckValidity(_ [][]byte, _ *sovereign.TransferData) error {
	return nil
}

// IsInterfaceNil - returns true if there is no value under the interface
func (tc *topicsChecker) IsInterfaceNil() bool {
	return tc == nil
}
