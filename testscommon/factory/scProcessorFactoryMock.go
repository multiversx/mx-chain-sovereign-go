package factory

import (
	"github.com/multiversx/mx-chain-go/process/smartContract/scrCommon"
	"github.com/multiversx/mx-chain-go/testscommon"
)

// SCProcessorFactoryMock -
type SCProcessorFactoryMock struct {
	CreateSCProcessorCalled func(args scrCommon.ArgsNewSmartContractProcessor) (scrCommon.SCRProcessorHandler, error)
}

// CreateSCProcessor -
func (s *SCProcessorFactoryMock) CreateSCProcessor(args scrCommon.ArgsNewSmartContractProcessor) (scrCommon.SCRProcessorHandler, error) {
	if s.CreateSCProcessorCalled != nil {
		return s.CreateSCProcessorCalled(args)
	}
	return &testscommon.SCProcessorMock{}, nil
}

// IsInterfaceNil -
func (s *SCProcessorFactoryMock) IsInterfaceNil() bool {
	return s == nil
}
