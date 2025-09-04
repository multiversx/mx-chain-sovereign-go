package extraSigners

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	logger "github.com/multiversx/mx-chain-logger-go"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
	"github.com/multiversx/mx-chain-go/errors"
)

var log = logger.GetOrCreate("extra-signers")

type sovereignSubRoundEndOutGoingTxData struct {
	signingHandler      consensus.SigningHandler
	mbType              block.OutGoingMBType
	enableEpochsHandler common.EnableEpochsHandler
}

// NewSovereignSubRoundEndExtraSigner creates a new extra signer for sovereign outgoing mini blocks in end subround
func NewSovereignSubRoundEndExtraSigner(
	signingHandler consensus.SigningHandler,
	mbType block.OutGoingMBType,
	enableEpochsHandler common.EnableEpochsHandler,
) (*sovereignSubRoundEndOutGoingTxData, error) {
	if check.IfNil(signingHandler) {
		return nil, spos.ErrNilSigningHandler
	}
	if check.IfNil(enableEpochsHandler) {
		return nil, spos.ErrNilEnableEpochHandler
	}

	return &sovereignSubRoundEndOutGoingTxData{
		signingHandler:      signingHandler,
		mbType:              mbType,
		enableEpochsHandler: enableEpochsHandler,
	}, nil
}

// VerifyAggregatedSignatures verifies outgoing tx aggregated signatures from provided header
func (sr *sovereignSubRoundEndOutGoingTxData) VerifyAggregatedSignatures(bitmap []byte, header data.HeaderHandler) error {
	sovHeader, castOk := header.(data.SovereignChainHeaderHandler)
	if !castOk {
		return fmt.Errorf("%w in sovereignSubRoundEndOutGoingTxData.SetAggregatedSignatureInHeader", errors.ErrWrongTypeAssertion)
	}

	outGoingMBs := sovHeader.GetOutGoingMiniBlockHeaderHandlersWithType(int32(sr.mbType))
	if len(outGoingMBs) == 0 {
		return nil
	}

	err := checkAllMBsHaveSameData(outGoingMBs, getOpHashData)
	if err != nil {
		return err
	}

	return sr.signingHandler.Verify(outGoingMBs[0].GetOutGoingOperationsHash(), bitmap, header.GetEpoch())
}

// AggregateAndSetSignatures aggregates and sets signatures for outgoing tx data
func (sr *sovereignSubRoundEndOutGoingTxData) AggregateAndSetSignatures(bitmap []byte, header data.HeaderHandler) ([]byte, error) {
	sovHeader, castOk := header.(data.SovereignChainHeaderHandler)
	if !castOk {
		return nil, fmt.Errorf("%w in sovereignSubRoundEndOutGoingTxData.SetAggregatedSignatureInHeader", errors.ErrWrongTypeAssertion)
	}

	outGoingMBs := sovHeader.GetOutGoingMiniBlockHeaderHandlersWithType(int32(sr.mbType))
	if len(outGoingMBs) == 0 {
		return nil, nil
	}

	sig, err := sr.signingHandler.AggregateSigs(bitmap, header.GetEpoch())
	if err != nil {
		return nil, err
	}

	err = sr.signingHandler.SetAggregatedSig(sig)
	if err != nil {
		return nil, err
	}

	return sig, nil
}

// SetAggregatedSignatureInHeader sets aggregated signature for outgoing tx in header
func (sr *sovereignSubRoundEndOutGoingTxData) SetAggregatedSignatureInHeader(header data.HeaderHandler, aggregatedSig []byte) error {
	sovHeader, castOk := header.(data.SovereignChainHeaderHandler)
	if !castOk {
		return fmt.Errorf("%w in sovereignSubRoundEndOutGoingTxData.SetAggregatedSignatureInHeader", errors.ErrWrongTypeAssertion)
	}

	outGoingMBs := sovHeader.GetOutGoingMiniBlockHeaderHandlersWithType(int32(sr.mbType))
	if len(outGoingMBs) == 0 {
		return nil
	}

	for _, outGoingMB := range outGoingMBs {
		err := outGoingMB.SetAggregatedSignatureOutGoingOperations(aggregatedSig)
		if err != nil {
			return err
		}

		err = sovHeader.SetOutGoingMiniBlockHeaderHandler(outGoingMB)
		if err != nil {
			return err
		}
	}

	return nil
}

// SignAndSetLeaderSignature signs and sets leader signature for outgoing tx in header
func (sr *sovereignSubRoundEndOutGoingTxData) SignAndSetLeaderSignature(header data.HeaderHandler, leaderPubKey []byte) error {
	sovHeader, castOk := header.(data.SovereignChainHeaderHandler)
	if !castOk {
		return fmt.Errorf("%w in sovereignSubRoundEndOutGoingTxData.SetAggregatedSignatureInHeader", errors.ErrWrongTypeAssertion)
	}

	outGoingMBs := sovHeader.GetOutGoingMiniBlockHeaderHandlersWithType(int32(sr.mbType))
	if len(outGoingMBs) == 0 {
		return nil
	}

	leaderMsgToSign := outGoingMBs[0].GetOutGoingOperationsHash()
	err := checkAllMBsHaveSameData(outGoingMBs, getOpHashData)
	if err != nil {
		return err
	}

	// In consensus v2 leader will only sign the outgoing op hash
	if !sr.enableEpochsHandler.IsFlagEnabled(common.AndromedaFlag) {
		leaderMsgToSign = append(leaderMsgToSign, outGoingMBs[0].GetAggregatedSignatureOutGoingOperations()...)
	}

	leaderSig, err := sr.signingHandler.CreateSignatureForPublicKey(leaderMsgToSign, leaderPubKey)
	if err != nil {
		return err
	}

	for _, outGoingMB := range outGoingMBs {
		err = outGoingMB.SetLeaderSignatureOutGoingOperations(leaderSig)
		if err != nil {
			return err
		}

		err = sovHeader.SetOutGoingMiniBlockHeaderHandler(outGoingMB)
		if err != nil {
			return err
		}
	}

	return nil
}

// SetConsensusDataInHeader sets aggregated and leader signature in header with provided data from consensus message
func (sr *sovereignSubRoundEndOutGoingTxData) SetConsensusDataInHeader(header data.HeaderHandler, cnsMsg *consensus.Message) error {
	sovHeader, castOk := header.(data.SovereignChainHeaderHandler)
	if !castOk {
		return fmt.Errorf("%w in sovereignSubRoundEndOutGoingTxData.SetConsensusDataInHeader", errors.ErrWrongTypeAssertion)
	}

	outGoingMBs := sovHeader.GetOutGoingMiniBlockHeaderHandlersWithType(int32(sr.mbType))
	if len(outGoingMBs) == 0 {
		return nil
	}

	extraSigData, found := cnsMsg.ExtraSignatures[sr.mbType.String()]
	if !found {
		return fmt.Errorf("%w for type %s", bls.ErrExtraSigShareDataNotFound, sr.mbType.String())
	}

	for _, outGoingMB := range outGoingMBs {
		err := outGoingMB.SetAggregatedSignatureOutGoingOperations(extraSigData.AggregatedSignatureOutGoingTxData)
		if err != nil {
			return err
		}
		err = outGoingMB.SetLeaderSignatureOutGoingOperations(extraSigData.LeaderSignatureOutGoingTxData)
		if err != nil {
			return err
		}
		err = sovHeader.SetOutGoingMiniBlockHeaderHandler(outGoingMB)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetLeaderExtraSig will return the leader extra sig from the header
func (sr *sovereignSubRoundEndOutGoingTxData) GetLeaderExtraSig(header data.HeaderHandler) ([]byte, error) {
	sovHeader, castOk := header.(data.SovereignChainHeaderHandler)
	if !castOk {
		return nil, fmt.Errorf("%w in sovereignSubRoundEndOutGoingTxData.GetLeaderExtraSig", errors.ErrWrongTypeAssertion)
	}

	outGoingMBs := sovHeader.GetOutGoingMiniBlockHeaderHandlersWithType(int32(sr.mbType))
	if len(outGoingMBs) == 0 {
		return nil, nil
	}

	err := checkAllMBsHaveSameData(outGoingMBs, getLeaderSigData)
	if err != nil {
		return nil, err
	}

	return outGoingMBs[0].GetLeaderSignatureOutGoingOperations(), nil
}

// AddLeaderAndAggregatedSignatures adds aggregated and leader signature in consensus message with provided data from header
func (sr *sovereignSubRoundEndOutGoingTxData) AddLeaderAndAggregatedSignatures(header data.HeaderHandler, cnsMsg *consensus.Message) error {
	sovHeader, castOk := header.(data.SovereignChainHeaderHandler)
	if !castOk {
		return fmt.Errorf("%w in sovereignSubRoundEndOutGoingTxData.SetConsensusDataInHeader", errors.ErrWrongTypeAssertion)
	}

	outGoingMBs := sovHeader.GetOutGoingMiniBlockHeaderHandlersWithType(int32(sr.mbType))
	if len(outGoingMBs) == 0 {
		return nil
	}

	keyStr := sr.mbType.String()
	initExtraSignatureEntry(cnsMsg, keyStr)

	err := checkAllMBsHaveSameData(outGoingMBs, getAggSigData, getLeaderSigData)
	if err != nil {
		return err
	}

	cnsMsg.ExtraSignatures[keyStr].AggregatedSignatureOutGoingTxData = outGoingMBs[0].GetAggregatedSignatureOutGoingOperations()
	cnsMsg.ExtraSignatures[keyStr].LeaderSignatureOutGoingTxData = outGoingMBs[0].GetLeaderSignatureOutGoingOperations()

	log.Debug("sovereignSubRoundEndOutGoingTxData.AddLeaderAndAggregatedSignatures",
		"AggregatedSignatureOutGoingTxData", cnsMsg.ExtraSignatures[keyStr].AggregatedSignatureOutGoingTxData,
		"LeaderSignatureOutGoingTxData", cnsMsg.ExtraSignatures[keyStr].LeaderSignatureOutGoingTxData,
		"type", keyStr)

	return nil
}

// Identifier returns the unique id of the signer
func (sr *sovereignSubRoundEndOutGoingTxData) Identifier() string {
	return sr.mbType.String()
}

// IsInterfaceNil checks if the underlying pointer is nil
func (sr *sovereignSubRoundEndOutGoingTxData) IsInterfaceNil() bool {
	return sr == nil
}
