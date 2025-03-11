package factory

import (
	"github.com/multiversx/mx-chain-sovereign-go/process"
	"github.com/multiversx/mx-chain-sovereign-go/process/block/preprocess"
	"github.com/multiversx/mx-chain-sovereign-go/process/mock"
)

// SmartContractResultPreProcessorFactoryStub -
type SmartContractResultPreProcessorFactoryStub struct {
}

// CreateSmartContractResultPreProcessor -
func (s *SmartContractResultPreProcessorFactoryStub) CreateSmartContractResultPreProcessor(_ preprocess.SmartContractResultPreProcessorCreatorArgs) (process.PreProcessor, error) {
	return &mock.PreProcessorMock{}, nil
}

// IsInterfaceNil -
func (s *SmartContractResultPreProcessorFactoryStub) IsInterfaceNil() bool {
	return s == nil
}
