package coordinator

import "github.com/multiversx/mx-chain-sovereign-go/process"

// TransactionCoordinatorCreator is an interface for creating a transaction coordinator
type TransactionCoordinatorCreator interface {
	CreateTransactionCoordinator(argsTransactionCoordinator ArgTransactionCoordinator) (process.TransactionCoordinator, error)
	IsInterfaceNil() bool
}
