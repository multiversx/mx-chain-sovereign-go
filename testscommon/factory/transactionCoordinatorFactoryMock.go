package factory

import (
	"github.com/multiversx/mx-chain-sovereign-go/process"
	"github.com/multiversx/mx-chain-sovereign-go/process/coordinator"
	"github.com/multiversx/mx-chain-sovereign-go/testscommon"
)

// TransactionCoordinatorFactoryMock -
type TransactionCoordinatorFactoryMock struct {
	CreateTransactionCoordinatorCalled func(args coordinator.ArgTransactionCoordinator) (process.TransactionCoordinator, error)
}

// CreateTransactionCoordinator -
func (t *TransactionCoordinatorFactoryMock) CreateTransactionCoordinator(args coordinator.ArgTransactionCoordinator) (process.TransactionCoordinator, error) {
	if t.CreateTransactionCoordinatorCalled != nil {
		return t.CreateTransactionCoordinatorCalled(args)
	}
	return &testscommon.TransactionCoordinatorMock{}, nil
}

// IsInterfaceNil -
func (t *TransactionCoordinatorFactoryMock) IsInterfaceNil() bool {
	return t == nil
}
