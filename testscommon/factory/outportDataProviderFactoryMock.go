package factory

import (
	"github.com/multiversx/mx-chain-sovereign-go/outport"
	"github.com/multiversx/mx-chain-sovereign-go/outport/process/factory"
	outportStub "github.com/multiversx/mx-chain-sovereign-go/testscommon/outport"
)

// OutportDataProviderFactoryMock -
type OutportDataProviderFactoryMock struct {
	CreateOutportDataProviderCalled func(arg factory.ArgOutportDataProviderFactory) (outport.DataProviderOutport, error)
}

// CreateOutportDataProvider -
func (f *OutportDataProviderFactoryMock) CreateOutportDataProvider(arg factory.ArgOutportDataProviderFactory) (outport.DataProviderOutport, error) {
	if f.CreateOutportDataProviderCalled != nil {
		return f.CreateOutportDataProviderCalled(arg)
	}

	return &outportStub.OutportDataProviderStub{}, nil
}

// IsInterfaceNil -
func (f *OutportDataProviderFactoryMock) IsInterfaceNil() bool {
	return f == nil
}
