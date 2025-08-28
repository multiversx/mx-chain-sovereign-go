package sovereign

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/multiversx/mx-chain-core-go/data/sovereign/dto"

	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
	sovData "github.com/multiversx/mx-chain-go/dataRetriever/dataPool/sovereign"
	"github.com/multiversx/mx-chain-go/errors"
)

type sovereignSubRoundEnd struct {
	bls.SubRoundEndHandler
	outGoingOperationsPool sovData.ShardedOutGoingOperationPool
	bridgeOpHandler        bls.BridgeOperationsHandler
}

// NewSovereignSubRoundEndRound creates a new sovereign end subround
func NewSovereignSubRoundEndRound(
	subroundBlock bls.SubRoundEndHandler,
	outGoingOperationsPool sovData.ShardedOutGoingOperationPool,
	bridgeOpHandler bls.BridgeOperationsHandler,
) (*sovereignSubRoundEnd, error) {
	if check.IfNil(subroundBlock) {
		return nil, spos.ErrNilSubround
	}
	if check.IfNil(outGoingOperationsPool) {
		return nil, errors.ErrNilOutGoingOperationsPool
	}
	if check.IfNil(bridgeOpHandler) {
		return nil, errors.ErrNilBridgeOpHandler
	}

	sr := &sovereignSubRoundEnd{
		outGoingOperationsPool: outGoingOperationsPool,
		bridgeOpHandler:        bridgeOpHandler,
		SubRoundEndHandler:     subroundBlock,
	}

	sr.SetMessageToVerifySigFunc(sr.getMessageToVerifySig)
	sr.SetBlockJob(sr.doSovereignEndRoundJob)

	return sr, nil
}

func (sr *sovereignSubRoundEnd) getMessageToVerifySig() []byte {
	headerHash, err := core.CalculateHash(sr.Marshalizer(), sr.Hasher(), sr.GetHeader())
	if err != nil {
		log.Error("sovereignSubRoundEnd.getMessageToVerifySig", "error", err.Error())
		return nil
	}

	return headerHash
}

func (sr *sovereignSubRoundEnd) receivedBlockHeaderFinalInfo(ctx context.Context, cnsDta *consensus.Message) bool {
	success := sr.SubRoundEndHandler.ReceivedBlockHeaderFinalInfo(ctx, cnsDta)
	if !success {
		return false
	}

	// TODO: MX-15502 once we have ZKProofs included in blocks for leaders which have resent the unconfirmed
	// outgoing operation we should also call resetOutGoingOpTimer here for consensus participants
	return sr.updateOutGoingPoolIfNeeded(cnsDta) == nil
}

func (sr *sovereignSubRoundEnd) ReceivedProof(proof consensus.ProofHandler) {
	sr.SubRoundEndHandler.ReceivedProof(proof)

	err := sr.updateOutGoingPoolIfNeeded(&consensus.Message{
		PubKeysBitmap:   proof.GetPubKeysBitmap(),
		ExtraSignatures: getExtraSigsOutGoingOps(proof),
	})
	if err != nil {
		log.Error("sovereignSubRoundEnd.ReceivedProof", "error", err)
	}
}

func getExtraSigsOutGoingOps(proof consensus.ProofHandler) map[string]*consensus.ExtraSignatureData {
	extraSigsOutGoingOps := make(map[string]*consensus.ExtraSignatureData)

	for id, sigData := range proof.GetExtraSignatureHandlers() {
		extraSigsOutGoingOps[id] = &consensus.ExtraSignatureData{
			AggregatedSignatureOutGoingTxData: sigData.GetAggregatedSignature(),
			LeaderSignatureOutGoingTxData:     sigData.GetLeaderSignature(),
		}
	}

	return extraSigsOutGoingOps
}

func (sr *sovereignSubRoundEnd) updateOutGoingPoolIfNeeded(cnsDta *consensus.Message) error {
	sovHeader, castOk := sr.GetHeader().(data.SovereignChainHeaderHandler)
	if !castOk {
		log.Error("sovereignSubRoundEnd.updateOutGoingPoolIfNeeded", "error", errors.ErrWrongTypeAssertion)
		return errors.ErrWrongTypeAssertion
	}

	for _, outGoingMbHdr := range sovHeader.GetOutGoingMiniBlockHeaderHandlers() {
		err := sr.updatePoolForOutGoingMiniBlock(outGoingMbHdr, cnsDta)
		if err != nil {
			return err
		}
	}

	return nil
}

func (sr *sovereignSubRoundEnd) updatePoolForOutGoingMiniBlock(
	outGoingMBHeader data.OutGoingMiniBlockHeaderHandler,
	cnsDta *consensus.Message,
) error {
	mbType := block.OutGoingMBType(outGoingMBHeader.GetOutGoingMBTypeInt32()).String()
	extraSigData, found := cnsDta.ExtraSignatures[mbType]
	if !found {
		return fmt.Errorf("%w for type %s", bls.ErrExtraSigShareDataNotFound, mbType)
	}

	err := outGoingMBHeader.SetAggregatedSignatureOutGoingOperations(extraSigData.AggregatedSignatureOutGoingTxData)
	if err != nil {
		log.Error("sovereignSubRoundEnd.updatePoolForOutGoingMiniBlock.SetAggregatedSignatureOutGoingOperations", "error", err)
		return err
	}

	err = outGoingMBHeader.SetLeaderSignatureOutGoingOperations(extraSigData.LeaderSignatureOutGoingTxData)
	if err != nil {
		log.Error("sovereignSubRoundEnd.updatePoolForOutGoingMiniBlock.SetLeaderSignatureOutGoingOperations", "error", err)
		return err
	}

	log.Debug("step 3.1: block header final info has been received with outgoing mb",
		"LeaderSignatureOutGoingTxData", extraSigData.LeaderSignatureOutGoingTxData,
		"AggregatedSignatureOutGoingTxData", extraSigData.AggregatedSignatureOutGoingTxData,
		"type", mbType,
	)

	_, err = sr.updateBridgeDataWithSignatures(outGoingMBHeader, cnsDta.PubKeysBitmap)
	if err != nil {
		log.Error("sovereignSubRoundEnd.updatePoolForOutGoingMiniBlock.updateBridgeDataWithSignatures", "error", err)
		return err
	}

	return nil
}

func (sr *sovereignSubRoundEnd) doSovereignEndRoundJob(ctx context.Context) bool {
	success := sr.SubRoundEndHandler.DoEndRoundJob(ctx)
	if !success {
		return false
	}

	sovHeader, castOk := sr.GetHeader().(data.SovereignChainHeaderHandler)
	if !castOk {
		log.Error("sovereignSubRoundEnd.doSovereignEndRoundJob", "error", errors.ErrWrongTypeAssertion)
		return false
	}

	outGoingMBHeaders := sovHeader.GetOutGoingMiniBlockHeaderHandlers()
	if len(outGoingMBHeaders) == 0 {
		sr.sendUnconfirmedOperationsIfFound(ctx)
		return true
	}

	currentOperations := make([]*sovereign.BridgeOutGoingData, len(outGoingMBHeaders))
	for idx, outGoingMBHdr := range outGoingMBHeaders {
		currBridgeData, err := sr.updateBridgeDataWithSignatures(outGoingMBHdr, sovHeader.GetPubKeysBitmap())
		if err != nil {
			log.Error("sovereignSubRoundEnd.doSovereignEndRoundJob.updateBridgeDataWithSignatures", "error", err)
			return false
		}
		currentOperations[idx] = currBridgeData
	}

	if !sr.isSelfLeader() {
		return true
	}

	outGoingOperations := sr.getAllOutGoingOperations(currentOperations)
	go sr.sendOutGoingOperations(ctx, outGoingOperations)

	return true
}

func (sr *sovereignSubRoundEnd) sendUnconfirmedOperationsIfFound(ctx context.Context) {
	if !sr.isSelfLeader() {
		return
	}

	unconfirmedOperations := sr.outGoingOperationsPool.GetUnconfirmedOperations()
	if len(unconfirmedOperations) == 0 {
		return
	}

	log.Debug("found unconfirmed operations", "num unconfirmed operations", len(unconfirmedOperations))
	go sr.sendOutGoingOperations(ctx, unconfirmedOperations)
}

func (sr *sovereignSubRoundEnd) updateBridgeDataWithSignatures(
	outGoingMBHeader data.OutGoingMiniBlockHeaderHandler, pubKeysBitmap []byte,
) (*sovereign.BridgeOutGoingData, error) {
	chainID := outGoingMBHeader.GetChainID()
	hash := outGoingMBHeader.GetOutGoingOperationsHash()
	currBridgeData := sr.outGoingOperationsPool.Get(hash, chainID)
	if currBridgeData == nil {
		return nil, fmt.Errorf("%w in sovereignSubRoundEnd.updateBridgeDataWithSignatures for hash: %s",
			errors.ErrOutGoingOperationsNotFound, hex.EncodeToString(hash))
	}

	currBridgeData.LeaderSignature = outGoingMBHeader.GetLeaderSignatureOutGoingOperations()
	currBridgeData.AggregatedSignature = outGoingMBHeader.GetAggregatedSignatureOutGoingOperations()
	currBridgeData.PubKeysBitmap = pubKeysBitmap

	sr.outGoingOperationsPool.Delete(hash, chainID)
	sr.outGoingOperationsPool.Add(currBridgeData, chainID)
	return currBridgeData, nil
}

func (sr *sovereignSubRoundEnd) isSelfLeader() bool {
	return sr.IsSelfLeaderInCurrentRound() || sr.IsMultiKeyLeaderInCurrentRound()
}

func (sr *sovereignSubRoundEnd) getAllOutGoingOperations(currentOperations []*sovereign.BridgeOutGoingData) []*sovereign.BridgeOutGoingData {
	outGoingOperations := make([]*sovereign.BridgeOutGoingData, 0)
	unconfirmedOperations := sr.outGoingOperationsPool.GetUnconfirmedOperations()
	if len(unconfirmedOperations) != 0 {
		log.Debug("found unconfirmed operations", "num unconfirmed operations", len(unconfirmedOperations))
		outGoingOperations = append(unconfirmedOperations, outGoingOperations...)
	}

	log.Debug("current outgoing operations", "num", len(currentOperations))
	return append(outGoingOperations, currentOperations...)
}

func (sr *sovereignSubRoundEnd) sendOutGoingOperations(ctx context.Context, data []*sovereign.BridgeOutGoingData) {
	resp, err := sr.bridgeOpHandler.Send(ctx, &sovereign.BridgeOperations{
		Data: data,
	})
	if err != nil {
		log.Error("sovereignSubRoundEnd.doSovereignEndRoundJob.bridgeOpHandler.Send", "error", err)
		return
	}

	sr.resetOutGoingOpTimer(data)
	log.Debug("sent outgoing operations", "hashes", resp.TxHashes)
}

func (sr *sovereignSubRoundEnd) resetOutGoingOpTimer(data []*sovereign.BridgeOutGoingData) {
	for _, dta := range data {
		sr.outGoingOperationsPool.ResetTimer([][]byte{dta.Hash}, dto.ChainID(dta.ChainID))
	}
}

// IsInterfaceNil checks if the underlying pointer is nil
func (sr *sovereignSubRoundEnd) IsInterfaceNil() bool {
	return sr == nil
}
