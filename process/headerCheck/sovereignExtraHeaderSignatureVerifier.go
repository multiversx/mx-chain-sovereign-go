package headerCheck

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	crypto "github.com/multiversx/mx-chain-crypto-go"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process"
)

type sovereignHeaderSigVerifier struct {
	singleSigVerifier   crypto.SingleSigner
	enableEpochsHandler common.EnableEpochsHandler
}

// NewSovereignHeaderSigVerifier creates a new sovereign header sig verifier for outgoing operations
func NewSovereignHeaderSigVerifier(
	singleSigVerifier crypto.SingleSigner,
	enableEpochsHandler common.EnableEpochsHandler,
) (*sovereignHeaderSigVerifier, error) {
	if check.IfNil(singleSigVerifier) {
		return nil, process.ErrNilSingleSigner
	}
	if check.IfNil(enableEpochsHandler) {
		return nil, errors.ErrNilEnableEpochsHandler
	}

	return &sovereignHeaderSigVerifier{
		singleSigVerifier:   singleSigVerifier,
		enableEpochsHandler: enableEpochsHandler,
	}, nil
}

// VerifyAggregatedSignature verifies aggregated sig for outgoing operations
func (hsv *sovereignHeaderSigVerifier) VerifyAggregatedSignature(
	proof data.HeaderProofHandler,
	header data.HeaderHandler,
	multiSigVerifier crypto.MultiSigner,
	pubKeysSigners [][]byte,
) error {
	sovHeader, castOk := header.(data.SovereignChainHeaderHandler)
	if !castOk {
		return fmt.Errorf("%w in sovereignHeaderSigVerifier.VerifyAggregatedSignature", errors.ErrWrongTypeAssertion)
	}

	for _, outGoingMBHdr := range sovHeader.GetOutGoingMiniBlockHeaderHandlers() {
		aggregatedSig, err := hsv.getAggregatedSignature(outGoingMBHdr, proof, header.GetEpoch())
		if err != nil {
			return err
		}

		err = multiSigVerifier.VerifyAggregatedSig(
			pubKeysSigners,
			outGoingMBHdr.GetOutGoingOperationsHash(),
			aggregatedSig,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (hsv *sovereignHeaderSigVerifier) getAggregatedSignature(
	outGoingMBHdr data.OutGoingMiniBlockHeaderHandler,
	proof data.HeaderProofHandler,
	epoch uint32,
) ([]byte, error) {
	if !hsv.enableEpochsHandler.IsFlagEnabledInEpoch(common.AndromedaFlag, epoch) {
		return outGoingMBHdr.GetAggregatedSignatureOutGoingOperations(), nil
	}

	if check.IfNil(proof) {
		return nil, process.ErrNilHeaderProof
	}

	chainIDStr := outGoingMBHdr.GetChainID().String()
	extraSigHandler, found := proof.GetExtraSignatureHandlers()[chainIDStr]
	if !found {
		return nil, fmt.Errorf("%w in sovereignHeaderSigVerifier.VerifyAggregatedSignature for header hash: %x, round: %d, chain: %s",
			errNoExtraSignatureDataFoundInProof, proof.GetHeaderHash(), proof.GetHeaderRound(), chainIDStr)
	}

	return extraSigHandler.GetAggregatedSignature(), nil
}

// VerifyLeaderSignature verifies leader sig for outgoing operations
func (hsv *sovereignHeaderSigVerifier) VerifyLeaderSignature(
	header data.HeaderHandler,
	leaderPubKey crypto.PublicKey,
) error {
	sovHeader, castOk := header.(data.SovereignChainHeaderHandler)
	if !castOk {
		return fmt.Errorf("%w in sovereignHeaderSigVerifier.VerifyLeaderSignature", errors.ErrWrongTypeAssertion)
	}

	for _, outGoingMBHdr := range sovHeader.GetOutGoingMiniBlockHeaderHandlers() {
		err := hsv.singleSigVerifier.Verify(
			leaderPubKey,
			hsv.getLeaderSignedMessage(outGoingMBHdr, header.GetEpoch()),
			outGoingMBHdr.GetLeaderSignatureOutGoingOperations())
		if err != nil {
			return err
		}
	}

	return nil
}

func (hsv *sovereignHeaderSigVerifier) getLeaderSignedMessage(
	outGoingMBHdr data.OutGoingMiniBlockHeaderHandler,
	epoch uint32,
) []byte {
	// In consensus v2 leader will only sign the outgoing op hash
	if hsv.enableEpochsHandler.IsFlagEnabledInEpoch(common.AndromedaFlag, epoch) {
		return outGoingMBHdr.GetOutGoingOperationsHash()
	}

	return append(outGoingMBHdr.GetOutGoingOperationsHash(), outGoingMBHdr.GetAggregatedSignatureOutGoingOperations()...)
}

// RemoveLeaderSignature removes leader sig from outgoing operations
func (hsv *sovereignHeaderSigVerifier) RemoveLeaderSignature(header data.HeaderHandler) error {
	sovHeader, castOk := header.(data.SovereignChainHeaderHandler)
	if !castOk {
		return fmt.Errorf("%w in sovereignHeaderSigVerifier.RemoveLeaderSignature", errors.ErrWrongTypeAssertion)
	}

	for _, outGoingMBHdr := range sovHeader.GetOutGoingMiniBlockHeaderHandlers() {
		err := outGoingMBHdr.SetLeaderSignatureOutGoingOperations(nil)
		if err != nil {
			return err
		}

		err = sovHeader.SetOutGoingMiniBlockHeaderHandler(outGoingMBHdr)
		if err != nil {
			return err
		}
	}

	return nil
}

// RemoveAllSignatures removes aggregated + leader sig from outgoing operations
func (hsv *sovereignHeaderSigVerifier) RemoveAllSignatures(header data.HeaderHandler) error {
	sovHeader, castOk := header.(data.SovereignChainHeaderHandler)
	if !castOk {
		return fmt.Errorf("%w in sovereignHeaderSigVerifier.RemoveAllSignatures", errors.ErrWrongTypeAssertion)
	}

	for _, outGoingMBHdr := range sovHeader.GetOutGoingMiniBlockHeaderHandlers() {
		err := outGoingMBHdr.SetAggregatedSignatureOutGoingOperations(nil)
		if err != nil {
			return err
		}

		err = outGoingMBHdr.SetLeaderSignatureOutGoingOperations(nil)
		if err != nil {
			return err
		}

		err = sovHeader.SetOutGoingMiniBlockHeaderHandler(outGoingMBHdr)
		if err != nil {
			return err
		}
	}

	return nil
}

// Identifier returns the unique id of the header verifier
func (hsv *sovereignHeaderSigVerifier) Identifier() string {
	return "sovereignHeaderSigVerifier"
}

// IsInterfaceNil checks if the underlying pointer is nil
func (hsv *sovereignHeaderSigVerifier) IsInterfaceNil() bool {
	return hsv == nil
}
