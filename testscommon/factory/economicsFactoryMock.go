package factory

import (
	metachainEpochStart "github.com/multiversx/mx-chain-sovereign-go/epochStart/metachain"
	"github.com/multiversx/mx-chain-sovereign-go/integrationTests/mock"
	"github.com/multiversx/mx-chain-sovereign-go/process"
)

// EconomicsFactoryMock -
type EconomicsFactoryMock struct {
	CreateEndOfEpochEconomicsCalled func(args metachainEpochStart.ArgsNewEpochEconomics) (process.EndOfEpochEconomics, error)
}

// CreateEndOfEpochEconomics -
func (f *EconomicsFactoryMock) CreateEndOfEpochEconomics(args metachainEpochStart.ArgsNewEpochEconomics) (process.EndOfEpochEconomics, error) {
	if f.CreateEndOfEpochEconomicsCalled != nil {
		return f.CreateEndOfEpochEconomicsCalled(args)
	}

	return &mock.EpochEconomicsStub{}, nil
}

// IsInterfaceNil -
func (f *EconomicsFactoryMock) IsInterfaceNil() bool {
	return f == nil
}
