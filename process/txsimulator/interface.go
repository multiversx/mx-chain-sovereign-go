package txsimulator

import (
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/state"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

// TransactionProcessor defines the operations needed to be done by a transaction processor
type TransactionProcessor interface {
	ProcessTransaction(transaction *transaction.Transaction) (vmcommon.ReturnCode, error)
	VerifyTransaction(transaction *transaction.Transaction) (state.UserAccountHandler, error)
	IsInterfaceNil() bool
}
