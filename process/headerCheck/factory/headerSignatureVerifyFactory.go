package factory

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/headerCheck"
)

type headerSignatureVerifyFactory struct {
}

// NewHeaderSignatureVerifyFactory creates a header signature verifier factory for normal chain run type
func NewHeaderSignatureVerifyFactory() *headerSignatureVerifyFactory {
	return &headerSignatureVerifyFactory{}
}

// CreateHeaderSignatureVerifier creates a header signature verifier for normal chain run type
func (f *headerSignatureVerifyFactory) CreateHeaderSignatureVerifier(
	arguments *headerCheck.ArgsHeaderSigVerifier,
) (process.InterceptedHeaderSigVerifier, error) {
	return headerCheck.NewHeaderSigVerifier(arguments)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (f *headerSignatureVerifyFactory) IsInterfaceNil() bool {
	return f == nil
}
