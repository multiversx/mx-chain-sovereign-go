package factory

import (
	"github.com/multiversx/mx-chain-sovereign-go/epochStart"
	"github.com/multiversx/mx-chain-sovereign-go/factory"
	"github.com/multiversx/mx-chain-sovereign-go/testscommon"
)

// EpochStartTriggerFactoryMock -
type EpochStartTriggerFactoryMock struct {
	CreateEpochStartTriggerCalled func(args factory.ArgsEpochStartTrigger) (epochStart.TriggerHandler, error)
}

// CreateEpochStartTrigger -
func (mock *EpochStartTriggerFactoryMock) CreateEpochStartTrigger(args factory.ArgsEpochStartTrigger) (epochStart.TriggerHandler, error) {
	if mock.CreateEpochStartTriggerCalled != nil {
		return mock.CreateEpochStartTriggerCalled(args)
	}
	return &testscommon.EpochStartTriggerStub{}, nil
}

// IsInterfaceNil -
func (mock *EpochStartTriggerFactoryMock) IsInterfaceNil() bool {
	return mock == nil
}
