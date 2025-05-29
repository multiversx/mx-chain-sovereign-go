package sovereign

import (
	"fmt"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
	v1 "github.com/multiversx/mx-chain-go/consensus/spos/bls/v1"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/outport"
)

type SubRoundsFactoryHandler interface {
	GenerateBlockSubround() (SubRoundBlockHandler, error)
	GenerateSignatureSubround() (SubRoundSignatureHandler, error)
	GenerateEndRoundSubround() (SubRoundEndHandler, error)
}

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
	consensusModel        consensus.ConsensusModel
	enableEpochHandler    common.EnableEpochsHandler
	extraSignersHolder    bls.ExtraSignersHolder
	baseSubRoundsFactory  SubRoundsFactoryHandler

	outGoingOperationsPool bls.OutGoingOperationsPool
	bridgeOpHandler        bls.BridgeOperationsHandler
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
	consensusModel consensus.ConsensusModel,
	enableEpochHandler common.EnableEpochsHandler,
	extraSignersHolder bls.ExtraSignersHolder,
	baseSubRoundsFactory SubRoundsFactoryHandler,
	outGoingOperationsPool bls.OutGoingOperationsPool,
	bridgeOpHandler bls.BridgeOperationsHandler,
) (*factory, error) {
	// no need to check the outportHandler, it can be nil
	err := checkNewFactoryParams(
		consensusDataContainer,
		consensusState,
		worker,
		chainID,
		appStatusHandler,
		sentSignaturesTracker,
		enableEpochHandler,
		extraSignersHolder,
		baseSubRoundsFactory,
		outGoingOperationsPool,
		bridgeOpHandler,
	)
	if err != nil {
		return nil, err
	}

	fct := factory{
		consensusCore:          consensusDataContainer,
		consensusState:         consensusState,
		worker:                 worker,
		appStatusHandler:       appStatusHandler,
		outportHandler:         outportHandler,
		sentSignaturesTracker:  sentSignaturesTracker,
		chainID:                chainID,
		currentPid:             currentPid,
		consensusModel:         consensusModel,
		enableEpochHandler:     enableEpochHandler,
		extraSignersHolder:     extraSignersHolder,
		baseSubRoundsFactory:   baseSubRoundsFactory,
		outGoingOperationsPool: outGoingOperationsPool,
		bridgeOpHandler:        bridgeOpHandler,
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
	enableEpochHandler common.EnableEpochsHandler,
	extraSignersHolder bls.ExtraSignersHolder,
	baseSubRoundsFactory SubRoundsFactoryHandler,
	outGoingOperationsPool bls.OutGoingOperationsPool,
	bridgeOpHandler bls.BridgeOperationsHandler,
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
		return bls.ErrNilSentSignatureTracker
	}
	if len(chainID) == 0 {
		return spos.ErrInvalidChainID
	}
	if check.IfNil(enableEpochHandler) {
		return spos.ErrNilEnableEpochHandler
	}
	if check.IfNil(extraSignersHolder) {
		return errors.ErrNilExtraSignersHolder
	}
	if baseSubRoundsFactory == nil {
		return errNilSubRoundsFactoryInSovereign
	}
	if check.IfNil(outGoingOperationsPool) {
		return errors.ErrNilOutGoingOperationsPool
	}
	if check.IfNil(bridgeOpHandler) {
		return errors.ErrNilBridgeOpHandler
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

	err := fct.generateStartRoundSubround(fct.extraSignersHolder.GetSubRoundStartExtraSignersHolder())
	if err != nil {
		return err
	}

	switch fct.consensusModel {
	case consensus.ConsensusModelV2:
		err = fct.generateBlockSubroundV2()
		if err != nil {
			return err
		}

		err = fct.generateSignatureSubroundV2()
		if err != nil {
			return err
		}

		err = fct.generateEndRoundSubroundV2()
		if err != nil {
			return err
		}

		return nil
	default:
		return fmt.Errorf("%w model %v", errors.ErrUnimplementedConsensusModel, fct.consensusModel)
	}
}

func (fct *factory) getTimeDuration() time.Duration {
	return fct.consensusCore.RoundHandler().TimeDuration()
}

func (fct *factory) generateStartRoundSubround(extraSignersHolder bls.SubRoundStartExtraSignersHolder) error {
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
		fct.enableEpochHandler,
	)
	if err != nil {
		return err
	}

	subroundStartRoundInstance, err := v1.NewSubroundStartRound(
		subround,
		fct.worker.Extend,
		processingThresholdPercent,
		fct.worker.ExecuteStoredMessages,
		fct.worker.ResetConsensusMessages,
		fct.sentSignaturesTracker,
		extraSignersHolder,
	)
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

func (fct *factory) generateBlockSubroundV2() error {
	subroundBlockInstance, err := fct.baseSubRoundsFactory.GenerateBlockSubround()
	if err != nil {
		return err
	}

	subroundBlockV2Instance, errV2 := NewSubroundBlockV2(subroundBlockInstance)
	if errV2 != nil {
		return errV2
	}

	fct.worker.AddReceivedMessageCall(bls.MtBlockBodyAndHeader, subroundBlockV2Instance.receivedBlockBodyAndHeader)
	fct.worker.AddReceivedMessageCall(bls.MtBlockBody, subroundBlockV2Instance.receivedBlockBody)
	fct.worker.AddReceivedMessageCall(bls.MtBlockHeader, subroundBlockV2Instance.receivedBlockHeader)
	fct.consensusCore.Chronology().AddSubround(subroundBlockV2Instance)

	return nil
}

func (fct *factory) generateSignatureSubroundV2() error {
	subroundSignatureInstance, err := fct.baseSubRoundsFactory.GenerateSignatureSubround()
	if err != nil {
		return err
	}

	subroundSignatureV2Instance, errV2 := NewSubroundSignatureV2(subroundSignatureInstance)
	if errV2 != nil {
		return errV2
	}

	fct.worker.AddReceivedMessageCall(bls.MtSignature, subroundSignatureV2Instance.receivedSignature)
	fct.consensusCore.Chronology().AddSubround(subroundSignatureV2Instance)

	return nil
}

func (fct *factory) generateEndRoundSubroundV2() error {
	subroundEndRoundInstance, err := fct.baseSubRoundsFactory.GenerateEndRoundSubround()
	if err != nil {
		return err
	}

	subroundEndV2Instance, err := NewSubroundEndRoundV2(subroundEndRoundInstance)
	if err != nil {
		return err
	}

	sovEndRound, err := NewSovereignSubRoundEndRound(
		subroundEndV2Instance,
		fct.outGoingOperationsPool,
		fct.bridgeOpHandler,
	)
	if err != nil {
		return err
	}

	fct.worker.AddReceivedMessageCall(bls.MtBlockHeaderFinalInfo, sovEndRound.receivedBlockHeaderFinalInfo)
	fct.worker.AddReceivedMessageCall(bls.MtInvalidSigners, sovEndRound.receivedInvalidSignersInfo)
	fct.worker.AddReceivedHeaderHandler(sovEndRound.receivedHeader)
	fct.consensusCore.Chronology().AddSubround(sovEndRound)

	return nil
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
