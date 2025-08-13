package factory

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/headerCheck"
	"github.com/multiversx/mx-chain-go/testscommon/consensus"
)

// HeaderSignatureVerifyFactoryMock -
type HeaderSignatureVerifyFactoryMock struct {
	CreateHeaderSignatureVerifierCalled func(
		arguments *headerCheck.ArgsHeaderSigVerifier,
	) (process.InterceptedHeaderSigVerifier, error)
}

// CreateHeaderSignatureVerifier -
func (mock *HeaderSignatureVerifyFactoryMock) CreateHeaderSignatureVerifier(
	arguments *headerCheck.ArgsHeaderSigVerifier,
) (process.InterceptedHeaderSigVerifier, error) {
	if mock.CreateHeaderSignatureVerifierCalled != nil {
		return mock.CreateHeaderSignatureVerifierCalled(arguments)
	}

	return &consensus.HeaderSigVerifierMock{}, nil
}

// IsInterfaceNil checks if the underlying pointer is nil
func (mock *HeaderSignatureVerifyFactoryMock) IsInterfaceNil() bool {
	return mock == nil
}
