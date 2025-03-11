package factory

import (
	"github.com/multiversx/mx-chain-sovereign-go/epochStart/metachain"
	"github.com/multiversx/mx-chain-sovereign-go/process"
	"github.com/multiversx/mx-chain-sovereign-go/testscommon"
)

// ValidatorInfoCreatorFactoryMock -
type ValidatorInfoCreatorFactoryMock struct {
	CreateValidatorInfoCreatorCalled func(args metachain.ArgsNewValidatorInfoCreator) (process.EpochStartValidatorInfoCreator, error)
}

// CreateValidatorInfoCreator -
func (mock *ValidatorInfoCreatorFactoryMock) CreateValidatorInfoCreator(args metachain.ArgsNewValidatorInfoCreator) (process.EpochStartValidatorInfoCreator, error) {
	if mock.CreateValidatorInfoCreatorCalled != nil {
		return mock.CreateValidatorInfoCreatorCalled(args)
	}

	return &testscommon.EpochValidatorInfoCreatorStub{}, nil
}

// IsInterfaceNil -
func (mock *ValidatorInfoCreatorFactoryMock) IsInterfaceNil() bool {
	return mock == nil
}
