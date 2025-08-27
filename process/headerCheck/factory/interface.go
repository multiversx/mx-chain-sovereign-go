package factory

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/headerCheck"
)

// HeaderSigVerifierFactory defines a header sig verifier handler
type HeaderSigVerifierFactory interface {
	CreateHeaderSignatureVerifier(
		arguments *headerCheck.ArgsHeaderSigVerifier,
	) (process.InterceptedHeaderSigVerifier, error)
	IsInterfaceNil() bool
}
