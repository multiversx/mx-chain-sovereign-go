package factory

import (
	"github.com/multiversx/mx-chain-sovereign-go/process"
	procMock "github.com/multiversx/mx-chain-sovereign-go/process/mock"
)

// PreProcessorsContainerFactoryMock -
type PreProcessorsContainerFactoryMock struct {
}

// Create -
func (mock *PreProcessorsContainerFactoryMock) Create() (process.PreProcessorsContainer, error) {
	return &procMock.PreProcessorContainerMock{}, nil
}

// IsInterfaceNil  -
func (mock *PreProcessorsContainerFactoryMock) IsInterfaceNil() bool {
	return mock == nil
}
