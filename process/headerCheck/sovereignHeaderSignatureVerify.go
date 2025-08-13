package headerCheck

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	crypto "github.com/multiversx/mx-chain-crypto-go"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/runType"
	errMx "github.com/multiversx/mx-chain-go/errors"
)

type sovereignHeaderSignatureVerify struct {
	*HeaderSigVerifier
}

// NewSovereignHeaderSignatureVerifier creates a sovereign header sig verifier
func NewSovereignHeaderSignatureVerifier(headerSigVerifier *HeaderSigVerifier) (*sovereignHeaderSignatureVerify, error) {
	if check.IfNil(headerSigVerifier) {
		return nil, errMx.ErrNilHeaderSigVerifier
	}

	sovHdrSigVerifier := &sovereignHeaderSignatureVerify{
		headerSigVerifier,
	}

	sovHdrSigVerifier.headerSigVerifierHelper = sovHdrSigVerifier
	return sovHdrSigVerifier, nil
}

func (shv *sovereignHeaderSignatureVerify) verifyProofAggregatedSignature(
	multiSigVerifier crypto.MultiSigner,
	pubKeysSigners [][]byte,
	proof data.HeaderProofHandler,
) error {
	return multiSigVerifier.VerifyAggregatedSig(
		pubKeysSigners,
		proof.GetProcessedHeaderHash(),
		proof.GetAggregatedSignature(),
	)
}

func (shv *sovereignHeaderSignatureVerify) getLeaderSignedHeader(header data.HeaderHandler) data.HeaderHandler {
	if shv.enableEpochsHandler.IsFlagEnabledInEpoch(common.AndromedaFlag, header.GetEpoch()) {
		return runType.CreateSovereignProposedInitialHeader(header)
	}

	return header
}

// IsInterfaceNil checks if the underlying interface is nil
func (shv *sovereignHeaderSignatureVerify) IsInterfaceNil() bool {
	return shv == nil
}
