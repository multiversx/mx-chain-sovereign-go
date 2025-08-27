package v1

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/outport"
)

// factory defines the data needed by this factory to create all the subrounds and give them their specific
// functionality
type factory struct {
	consensusCore  spos.ConsensusCoreHandler
	consensusState spos.ConsensusStateHandler
	worker         spos.WorkerHandler

	appStatusHandler      core.AppStatusHandler
	outportHandler        outport.OutportHandler
	sentSignaturesTracker spos.SentSignaturesTracker
	chainID               []byte
	currentPid            core.PeerID
	enableEpochHandler    common.EnableEpochsHandler
	extraSignersHolder    bls.ExtraSignersHolder
}

// NewSubroundsFactory creates a new factory object
func NewSubroundsFactory(
	consensusDataContainer spos.ConsensusCoreHandler,
	consensusState spos.ConsensusStateHandler,
	worker spos.WorkerHandler,
	chainID []byte,
	currentPid core.PeerID,
	appStatusHandler core.AppStatusHandler,
	sentSignaturesTracker spos.SentSignaturesTracker,
	outportHandler outport.OutportHandler,
	extraSignersHolder bls.ExtraSignersHolder,
) (*factory, error) {
	// no need to check the outportHandler, it can be nil
	err := checkNewFactoryParams(
		consensusDataContainer,
		consensusState,
		worker,
		chainID,
		appStatusHandler,
		sentSignaturesTracker,
		extraSignersHolder,
	)
	if err != nil {
		return nil, err
	}

	fct := factory{
		consensusCore:         consensusDataContainer,
		consensusState:        consensusState,
		worker:                worker,
		appStatusHandler:      appStatusHandler,
		chainID:               chainID,
		currentPid:            currentPid,
		sentSignaturesTracker: sentSignaturesTracker,
		outportHandler:        outportHandler,
		enableEpochHandler:    consensusDataContainer.EnableEpochsHandler(),
		extraSignersHolder:    extraSignersHolder,
	}

	return &fct, nil
}

func checkNewFactoryParams(
	container spos.ConsensusCoreHandler,
	state spos.ConsensusStateHandler,
	worker spos.WorkerHandler,
	chainID []byte,
	appStatusHandler core.AppStatusHandler,
	sentSignaturesTracker spos.SentSignaturesTracker,
	extraSignersHolder bls.ExtraSignersHolder,
) error {
	err := spos.ValidateConsensusCore(container)
	if err != nil {
		return err
	}
	if check.IfNil(state) {
		return spos.ErrNilConsensusState
	}
	if check.IfNil(worker) {
		return spos.ErrNilWorker
	}
	if check.IfNil(appStatusHandler) {
		return spos.ErrNilAppStatusHandler
	}
	if check.IfNil(sentSignaturesTracker) {
		return ErrNilSentSignatureTracker
	}
	if len(chainID) == 0 {
		return spos.ErrInvalidChainID
	}
	if check.IfNil(extraSignersHolder) {
		return errors.ErrNilExtraSignersHolder
	}

	return nil
}

// SetOutportHandler method will update the value of the factory's outport
func (fct *factory) SetOutportHandler(driver outport.OutportHandler) {
	fct.outportHandler = driver
}

// GenerateSubrounds will generate the subrounds used in BLS consensus
func (fct *factory) GenerateSubrounds(_ uint32) error {
	fct.initConsensusThreshold()
	fct.consensusCore.Chronology().RemoveAllSubrounds()
	fct.worker.RemoveAllReceivedMessagesCalls()
	fct.worker.RemoveAllReceivedHeaderHandlers()

	err := fct.generateStartRoundSubroundV1()
	if err != nil {
		return err
	}

	err = fct.generateBlockSubroundV1()
	if err != nil {
		return err
	}

	err = fct.generateSignatureSubroundV1()
	if err != nil {
		return err
	}

	err = fct.generateEndRoundSubroundV1()
	if err != nil {
		return err
	}

	return nil
}

func (fct *factory) getTimeDuration() time.Duration {
	return fct.consensusCore.RoundHandler().TimeDuration()
}

func (fct *factory) generateStartRoundSubroundV1() error {
	subroundStartRoundInstance, err := fct.GenerateStartRoundSubround()
	if err != nil {
		return err
	}

	err = subroundStartRoundInstance.SetOutportHandler(fct.outportHandler)
	if err != nil {
		return err
	}

	fct.consensusCore.Chronology().AddSubround(subroundStartRoundInstance)

	return nil
}

// GenerateStartRoundSubround will generate the start sub round
func (fct *factory) GenerateStartRoundSubround() (bls.SubRoundStartHandler, error) {
	subround, err := spos.NewSubround(
		-1,
		bls.SrStartRound,
		bls.SrBlock,
		int64(float64(fct.getTimeDuration())*srStartStartTime),
		int64(float64(fct.getTimeDuration())*srStartEndTime),
		bls.GetSubroundName(bls.SrStartRound),
		fct.consensusState,
		fct.worker.GetConsensusStateChangedChannel(),
		fct.worker.ExecuteStoredMessages,
		fct.consensusCore,
		fct.chainID,
		fct.currentPid,
		fct.appStatusHandler,
	)
	if err != nil {
		return nil, err
	}

	subroundStartRoundInstance, err := NewSubroundStartRound(
		subround,
		fct.worker.Extend,
		processingThresholdPercent,
		fct.worker.ExecuteStoredMessages,
		fct.worker.ResetConsensusMessages,
		fct.sentSignaturesTracker,
		fct.extraSignersHolder.GetSubRoundStartExtraSignersHolder(),
	)
	if err != nil {
		return nil, err
	}

	return subroundStartRoundInstance, nil
}

func (fct *factory) generateBlockSubroundV1() error {
	subroundBlockInstance, err := fct.GenerateBlockSubround()
	if err != nil {
		return err
	}

	fct.consensusCore.Chronology().AddSubround(subroundBlockInstance)

	return nil
}

// GenerateBlockSubround will generate the block sub round
func (fct *factory) GenerateBlockSubround() (bls.SubRoundBlockHandler, error) {
	subround, err := spos.NewSubround(
		bls.SrStartRound,
		bls.SrBlock,
		bls.SrSignature,
		int64(float64(fct.getTimeDuration())*srBlockStartTime),
		int64(float64(fct.getTimeDuration())*srBlockEndTime),
		bls.GetSubroundName(bls.SrBlock),
		fct.consensusState,
		fct.worker.GetConsensusStateChangedChannel(),
		fct.worker.ExecuteStoredMessages,
		fct.consensusCore,
		fct.chainID,
		fct.currentPid,
		fct.appStatusHandler,
	)
	if err != nil {
		return nil, err
	}

	subroundBlockInstance, err := NewSubroundBlock(
		subround,
		fct.worker.Extend,
		processingThresholdPercent,
	)
	if err != nil {
		return nil, err
	}

	fct.worker.AddReceivedMessageCall(bls.MtBlockBodyAndHeader, subroundBlockInstance.receivedBlockBodyAndHeader)
	fct.worker.AddReceivedMessageCall(bls.MtBlockBody, subroundBlockInstance.receivedBlockBody)
	fct.worker.AddReceivedMessageCall(bls.MtBlockHeader, subroundBlockInstance.receivedBlockHeader)
	fct.worker.AddReceivedHeaderHandler(subroundBlockInstance.receivedFullHeader)

	return subroundBlockInstance, nil
}

func (fct *factory) generateSignatureSubroundV1() error {
	subroundSignatureInstance, err := fct.GenerateSignatureSubround()
	if err != nil {
		return err
	}

	fct.consensusCore.Chronology().AddSubround(subroundSignatureInstance)

	return nil
}

// GenerateSignatureSubround will generate the signature sub round
func (fct *factory) GenerateSignatureSubround() (bls.SubRoundSignatureHandler, error) {
	subround, err := spos.NewSubround(
		bls.SrBlock,
		bls.SrSignature,
		bls.SrEndRound,
		int64(float64(fct.getTimeDuration())*srSignatureStartTime),
		int64(float64(fct.getTimeDuration())*srSignatureEndTime),
		bls.GetSubroundName(bls.SrSignature),
		fct.consensusState,
		fct.worker.GetConsensusStateChangedChannel(),
		fct.worker.ExecuteStoredMessages,
		fct.consensusCore,
		fct.chainID,
		fct.currentPid,
		fct.appStatusHandler,
	)
	if err != nil {
		return nil, err
	}

	subroundSignatureInstance, err := NewSubroundSignature(
		subround,
		fct.worker.Extend,
		fct.appStatusHandler,
		fct.extraSignersHolder.GetSubRoundSignatureExtraSignersHolder(),
		fct.sentSignaturesTracker,
	)
	if err != nil {
		return nil, err
	}

	fct.worker.AddReceivedMessageCall(bls.MtSignature, subroundSignatureInstance.receivedSignature)

	return subroundSignatureInstance, nil
}

func (fct *factory) generateEndRoundSubroundV1() error {
	subroundEndRoundInstance, err := fct.GenerateEndRoundSubround()
	if err != nil {
		return err
	}

	fct.consensusCore.Chronology().AddSubround(subroundEndRoundInstance)

	return nil
}

// GenerateEndRoundSubround will generate the end sub round
func (fct *factory) GenerateEndRoundSubround() (bls.SubRoundEndHandler, error) {
	subround, err := spos.NewSubround(
		bls.SrSignature,
		bls.SrEndRound,
		-1,
		int64(float64(fct.getTimeDuration())*srEndStartTime),
		int64(float64(fct.getTimeDuration())*srEndEndTime),
		bls.GetSubroundName(bls.SrEndRound),
		fct.consensusState,
		fct.worker.GetConsensusStateChangedChannel(),
		fct.worker.ExecuteStoredMessages,
		fct.consensusCore,
		fct.chainID,
		fct.currentPid,
		fct.appStatusHandler,
	)
	if err != nil {
		return nil, err
	}

	subroundEndRoundInstance, err := NewSubroundEndRound(
		subround,
		fct.worker.Extend,
		spos.MaxThresholdPercent,
		fct.worker.DisplayStatistics,
		fct.extraSignersHolder.GetSubRoundEndExtraSignersHolder(),
		fct.appStatusHandler,
		fct.sentSignaturesTracker,
	)
	if err != nil {
		return nil, err
	}

	fct.worker.AddReceivedMessageCall(bls.MtBlockHeaderFinalInfo, subroundEndRoundInstance.receivedBlockHeaderFinalInfo)
	fct.worker.AddReceivedMessageCall(bls.MtInvalidSigners, subroundEndRoundInstance.receivedInvalidSignersInfo)
	fct.worker.AddReceivedHeaderHandler(subroundEndRoundInstance.receivedHeader)

	return subroundEndRoundInstance, nil
}

func (fct *factory) initConsensusThreshold() {
	pBFTThreshold := core.GetPBFTThreshold(fct.consensusState.ConsensusGroupSize())
	pBFTFallbackThreshold := core.GetPBFTFallbackThreshold(fct.consensusState.ConsensusGroupSize())
	fct.consensusState.SetThreshold(bls.SrBlock, 1)
	fct.consensusState.SetThreshold(bls.SrSignature, pBFTThreshold)
	fct.consensusState.SetFallbackThreshold(bls.SrBlock, 1)
	fct.consensusState.SetFallbackThreshold(bls.SrSignature, pBFTFallbackThreshold)
}

// IsInterfaceNil returns true if there is no value under the interface
func (fct *factory) IsInterfaceNil() bool {
	return fct == nil
}
