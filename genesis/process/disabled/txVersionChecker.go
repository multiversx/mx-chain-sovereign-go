package disabled

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
)

// TxVersionChecker implements the TxVersionChecker interface, it does nothing as it is a disabled component
type TxVersionChecker struct{}

// NewDisabledTxVersionChecker is the constructor for the disabled tx version checker
func NewDisabledTxVersionChecker() *TxVersionChecker {
	return &TxVersionChecker{}
}

// IsGuardedTransaction returns false as this is a disabled component
func (tvc *TxVersionChecker) IsGuardedTransaction(_ data.TransactionHandler) bool {
	return false
}

// IsSignedWithHash returns false as this is a disabled component
func (tvc *TxVersionChecker) IsSignedWithHash(_ data.TransactionHandler) bool {
	return false
}

// CheckTxVersion returns nil as this is a disabled component
func (tvc *TxVersionChecker) CheckTxVersion(_ *transaction.Transaction) error {
	return nil
}

// IsInterfaceNil does the nil check for the receiver
func (tvc *TxVersionChecker) IsInterfaceNil() bool {
	return tvc == nil
}
