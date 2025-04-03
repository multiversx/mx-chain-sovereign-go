package disabled

import (
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"

	"github.com/multiversx/mx-chain-go/state"
)

// TxProcessor implements the TransactionProcessor interface but does nothing as it is disabled
type TxProcessor struct {
}

// ProcessTransaction does nothing as it is disabled
func (txProc *TxProcessor) ProcessTransaction(_ *transaction.Transaction) (vmcommon.ReturnCode, error) {
	return 0, nil
}

// VerifyTransaction does nothing as it is disabled
func (txProc *TxProcessor) VerifyTransaction(_ *transaction.Transaction) error {
	return nil
}

// VerifyGuardian does nothing as it is disabled
func (txProc *TxProcessor) VerifyGuardian(_ *transaction.Transaction, _ state.UserAccountHandler) error {
	return nil
}

// GetSenderAndReceiverAccounts does nothing as it is disabled
func (txProc *TxProcessor) GetSenderAndReceiverAccounts(_ *transaction.Transaction) (state.UserAccountHandler, state.UserAccountHandler, error) {
	return nil, nil, nil
}

// GetRelayerAccount does nothing as it is disabled
func (txProc *TxProcessor) GetRelayerAccount(_ *transaction.Transaction) (state.UserAccountHandler, error) {
	return nil, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (txProc *TxProcessor) IsInterfaceNil() bool {
	return txProc == nil
}
