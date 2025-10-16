package factory

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/headerCheck"
)

type sovereignHeaderSignatureVerifyFactory struct {
}

// NewSovereignHeaderSignatureVerifyFactory creates a header signature verifier factory for sovereign chain run type
func NewSovereignHeaderSignatureVerifyFactory() *sovereignHeaderSignatureVerifyFactory {
	return &sovereignHeaderSignatureVerifyFactory{}
}

// CreateHeaderSignatureVerifier creates a header signature verifier for sovereign chain run type
func (f *sovereignHeaderSignatureVerifyFactory) CreateHeaderSignatureVerifier(
	arguments *headerCheck.ArgsHeaderSigVerifier,
) (process.InterceptedHeaderSigVerifier, error) {
	hsv, err := headerCheck.NewHeaderSigVerifier(arguments)
	if err != nil {
		return nil, err
	}

	return headerCheck.NewSovereignHeaderSignatureVerifier(hsv)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (f *sovereignHeaderSignatureVerifyFactory) IsInterfaceNil() bool {
	return f == nil
}
