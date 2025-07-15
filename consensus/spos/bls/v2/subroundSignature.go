package v2

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	atomicCore "github.com/multiversx/mx-chain-core-go/core/atomic"
	"github.com/multiversx/mx-chain-core-go/core/check"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
	"github.com/multiversx/mx-chain-go/errors"
)

const timeSpentBetweenChecks = time.Millisecond

type subroundSignature struct {
	*spos.Subround
	appStatusHandler     core.AppStatusHandler
	sentSignatureTracker spos.SentSignaturesTracker
	signatureThrottler   core.Throttler

	getMessageToSignFunc func() []byte
	extraSignersHolder   bls.SubRoundSignatureExtraSignersHolder
}

// NewSubroundSignature creates a subroundSignature object
func NewSubroundSignature(
	baseSubround *spos.Subround,
	appStatusHandler core.AppStatusHandler,
	sentSignatureTracker spos.SentSignaturesTracker,
	worker spos.WorkerHandler,
	signatureThrottler core.Throttler,
	extraSignersHolder bls.SubRoundSignatureExtraSignersHolder,
) (*subroundSignature, error) {
	err := checkNewSubroundSignatureParams(
		baseSubround,
	)
	if err != nil {
		return nil, err
	}
	if check.IfNil(appStatusHandler) {
		return nil, spos.ErrNilAppStatusHandler
	}
	if check.IfNil(sentSignatureTracker) {
		return nil, ErrNilSentSignatureTracker
	}
	if check.IfNil(worker) {
		return nil, spos.ErrNilWorker
	}
	if check.IfNil(signatureThrottler) {
		return nil, spos.ErrNilThrottler
	}
	if check.IfNil(extraSignersHolder) {
		return nil, errors.ErrNilSignatureRoundExtraSignersHolder
	}

	srSignature := subroundSignature{
		Subround:             baseSubround,
		appStatusHandler:     appStatusHandler,
		sentSignatureTracker: sentSignatureTracker,
		signatureThrottler:   signatureThrottler,
		extraSignersHolder:   extraSignersHolder,
	}

	srSignature.Job = srSignature.doSignatureJob
	srSignature.Check = srSignature.doSignatureConsensusCheck
	srSignature.Extend = worker.Extend
	srSignature.getMessageToSignFunc = srSignature.getMessageToSign

	return &srSignature, nil
}

func checkNewSubroundSignatureParams(
	baseSubround *spos.Subround,
) error {
	if baseSubround == nil {
		return spos.ErrNilSubround
	}
	if check.IfNil(baseSubround.ConsensusStateHandler) {
		return spos.ErrNilConsensusState
	}

	err := spos.ValidateConsensusCore(baseSubround.ConsensusCoreHandler)

	return err
}

// doSignatureJob method does the job of the subround Signature
func (sr *subroundSignature) doSignatureJob(ctx context.Context) bool {
	if !sr.CanDoSubroundJob(sr.Current()) {
		return false
	}
	if check.IfNil(sr.GetHeader()) {
		log.Error("doSignatureJob", "error", spos.ErrNilHeader)
		return false
	}

	proofAlreadyReceived := sr.EquivalentProofsPool().HasProof(sr.ShardCoordinator().SelfId(), sr.getMessageToSignFunc())
	if proofAlreadyReceived {
		sr.SetStatus(sr.Current(), spos.SsFinished)
		log.Debug("step 2: subround has been finished, proof already received",
			"subround", sr.Name())

		return true
	}

	isSelfSingleKeyInConsensusGroup := sr.IsNodeInConsensusGroup(sr.SelfPubKey()) && sr.ShouldConsiderSelfKeyInConsensus()
	if isSelfSingleKeyInConsensusGroup {
		if !sr.doSignatureJobForSingleKey() {
			return false
		}
	}

	if !sr.doSignatureJobForManagedKeys(ctx) {
		return false
	}

	sr.SetStatus(sr.Current(), spos.SsFinished)
	log.Debug("step 2: subround has been finished",
		"subround", sr.Name())

	return true
}

func (sr *subroundSignature) createAndSendSignatureMessage(
	signatureShare []byte,
	extraSigShares map[string][]byte,
	pkBytes []byte,
) bool {
	cnsMsg := consensus.NewConsensusMessage(
		sr.GetData(),
		signatureShare,
		nil,
		nil,
		pkBytes,
		nil,
		int(bls.MtSignature),
		sr.RoundHandler().Index(),
		sr.ChainID(),
		nil,
		nil,
		nil,
		sr.GetAssociatedPid(pkBytes),
		nil,
		sr.getProcessedHeaderHash(),
	)

	err := sr.extraSignersHolder.AddExtraSigSharesToConsensusMessage(extraSigShares, cnsMsg)
	if err != nil {
		log.Debug("createAndSendSignatureMessage.extraSignersHolder.addExtraSigSharesToConsensusMessage",
			"error", err.Error(), "pk", pkBytes)
		return false
	}

	err = sr.BroadcastMessenger().BroadcastConsensusMessage(cnsMsg)
	if err != nil {
		log.Debug("createAndSendSignatureMessage.BroadcastConsensusMessage",
			"error", err.Error(), "pk", pkBytes)
		return false
	}

	log.Debug("step 2: signature has been sent", "pk", pkBytes)

	return true
}

func (sr *subroundSignature) getProcessedHeaderHash() []byte {
	if sr.EnableEpochHandler().IsFlagEnabled(common.ConsensusModelSovereignFlag) {
		return sr.getMessageToSignFunc()
	}

	return nil
}

func (sr *subroundSignature) completeSignatureSubRound(
	pk string,
	index int,
	processedHeaderHash []byte,
) bool {
	err := sr.SetJobDone(pk, sr.Current(), true)
	if err != nil {
		log.Debug("doSignatureJob.SetSelfJobDone",
			"subround", sr.Name(),
			"error", err.Error(),
			"pk", []byte(pk),
		)
		return false
	}

	if sr.EnableEpochHandler().IsFlagEnabled(common.ConsensusModelSovereignFlag) {
		sr.AddProcessedHeadersHashes(processedHeaderHash, index)
	}

	return true
}

// doSignatureConsensusCheck method checks if the consensus in the subround Signature is achieved
func (sr *subroundSignature) doSignatureConsensusCheck() bool {
	if sr.GetRoundCanceled() {
		return false
	}

	if sr.IsSubroundFinished(sr.Current()) {
		return true
	}

	if check.IfNil(sr.GetHeader()) {
		return false
	}

	isSelfInConsensusGroup := sr.IsSelfInConsensusGroup()
	if !isSelfInConsensusGroup {
		log.Debug("step 2: subround has been finished",
			"subround", sr.Name())
		sr.SetStatus(sr.Current(), spos.SsFinished)

		return true
	}

	if sr.IsSelfJobDone(sr.Current()) {
		log.Debug("step 2: subround has been finished",
			"subround", sr.Name())
		sr.SetStatus(sr.Current(), spos.SsFinished)
		sr.appStatusHandler.SetStringValue(common.MetricConsensusRoundState, "signed")

		return true
	}

	return false
}

func (sr *subroundSignature) doSignatureJobForManagedKeys(ctx context.Context) bool {
	numMultiKeysSignaturesSent := int32(0)
	sentSigForAllKeys := atomicCore.Flag{}
	sentSigForAllKeys.SetValue(true)

	wg := sync.WaitGroup{}

	for idx, pk := range sr.ConsensusGroup() {
		pkBytes := []byte(pk)
		if !sr.IsKeyManagedBySelf(pkBytes) {
			continue
		}

		if sr.IsJobDone(pk, sr.Current()) {
			continue
		}

		err := sr.checkGoRoutinesThrottler(ctx)
		if err != nil {
			return false
		}
		sr.signatureThrottler.StartProcessing()
		wg.Add(1)

		go func(idx int, pk string) {
			defer sr.signatureThrottler.EndProcessing()

			signatureSent := sr.sendSignatureForManagedKey(idx, pk)
			if signatureSent {
				atomic.AddInt32(&numMultiKeysSignaturesSent, 1)
			} else {
				sentSigForAllKeys.SetValue(false)
			}
			wg.Done()
		}(idx, pk)
	}

	wg.Wait()

	if numMultiKeysSignaturesSent > 0 {
		log.Debug("step 2: multi keys signatures have been sent", "num", numMultiKeysSignaturesSent)
	}

	return sentSigForAllKeys.IsSet()
}

func (sr *subroundSignature) sendSignatureForManagedKey(idx int, pk string) bool {
	pkBytes := []byte(pk)

	processedHeaderHash := sr.getMessageToSignFunc()
	signatureShare, err := sr.SigningHandler().CreateSignatureShareForPublicKey(
		processedHeaderHash,
		uint16(idx),
		sr.GetHeader().GetEpoch(),
		pkBytes,
	)
	if err != nil {
		log.Debug("sendSignatureForManagedKey.CreateSignatureShareForPublicKey", "error", err.Error())
		return false
	}

	extraSigShares, err := sr.extraSignersHolder.CreateExtraSignatureShares(sr.GetHeader(), uint16(idx), pkBytes)
	if err != nil {
		log.Debug("doSignatureJobForManagedKeys.extraSignersHolder.createExtraSignatureShares", "error", err.Error())
		return false
	}

	// with the equivalent messages feature on, signatures from all managed keys must be broadcast, as the aggregation is done by any participant
	ok := sr.createAndSendSignatureMessage(signatureShare, extraSigShares, pkBytes)
	if !ok {
		return false
	}
	sr.sentSignatureTracker.SignatureSent(pkBytes)

	// TODO: MX-16954 check at the end if this idx is ok or we should use selfIndex, err := sr.ConsensusGroupIndex(pk)
	return sr.completeSignatureSubRound(pk, idx, processedHeaderHash)
}

func (sr *subroundSignature) checkGoRoutinesThrottler(ctx context.Context) error {
	for {
		if sr.signatureThrottler.CanProcess() {
			break
		}
		select {
		case <-time.After(timeSpentBetweenChecks):
			continue
		case <-ctx.Done():
			return fmt.Errorf("%w while checking the throttler", spos.ErrTimeIsOut)
		}
	}
	return nil
}

func (sr *subroundSignature) doSignatureJobForSingleKey() bool {
	selfIndex, err := sr.SelfConsensusGroupIndex()
	if err != nil {
		log.Debug("doSignatureJobForSingleKey.SelfConsensusGroupIndex: not in consensus group")
		return false
	}

	processedHeaderHash := sr.getMessageToSignFunc()
	signatureShare, err := sr.SigningHandler().CreateSignatureShareForPublicKey(
		processedHeaderHash,
		uint16(selfIndex),
		sr.GetHeader().GetEpoch(),
		[]byte(sr.SelfPubKey()),
	)
	if err != nil {
		log.Debug("doSignatureJobForSingleKey.CreateSignatureShareForPublicKey", "error", err.Error())
		return false
	}

	extraSigShares, err := sr.extraSignersHolder.CreateExtraSignatureShares(sr.GetHeader(), uint16(selfIndex), []byte(sr.SelfPubKey()))
	if err != nil {
		log.Debug("doSignatureJobForSingleKey.extraSignersHolder.createExtraSignatureShares", "error", err.Error())
		return false
	}

	// leader also sends his signature here
	ok := sr.createAndSendSignatureMessage(signatureShare, extraSigShares, []byte(sr.SelfPubKey()))
	if !ok {
		return false
	}

	return sr.completeSignatureSubRound(sr.SelfPubKey(), selfIndex, processedHeaderHash)
}

func (sr *subroundSignature) getMessageToSign() []byte {
	return sr.GetData()
}

// SetMessageToSignFunc should set the message to sign func
func (sr *subroundSignature) SetMessageToSignFunc(_ func() []byte) {
	// TODO: MX-16954 Analyse if we will ever use this func, since it doesn't work for now
}

// IsInterfaceNil returns true if there is no value under the interface
func (sr *subroundSignature) IsInterfaceNil() bool {
	return sr == nil
}
