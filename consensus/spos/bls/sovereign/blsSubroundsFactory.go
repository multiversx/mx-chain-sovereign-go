package sovereign

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"

	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/outport"
)

// ArgsSovereignSubRoundsFactory defines a struct placeholder for args needed to create a new sovereign sub rounds factory
type ArgsSovereignSubRoundsFactory struct {
	ConsensusDataContainer spos.ConsensusCoreHandler
	ConsensusState         spos.ConsensusStateHandler
	Worker                 spos.WorkerHandler
	OutportHandler         outport.OutportHandler
	BaseSubRoundsFactory   SubRoundsFactoryHandler
	OutGoingOperationsPool bls.OutGoingOperationsPool
	BridgeOpHandler        bls.BridgeOperationsHandler
}

// factory defines the data needed by this factory to create all the subrounds and give them their specific
// functionality
type factory struct {
	consensusCore  spos.ConsensusCoreHandler
	consensusState spos.ConsensusStateHandler
	worker         spos.WorkerHandler

	outportHandler       outport.OutportHandler
	baseSubRoundsFactory SubRoundsFactoryHandler

	outGoingOperationsPool bls.OutGoingOperationsPool
	bridgeOpHandler        bls.BridgeOperationsHandler
}

// NewSubroundsFactory creates a new factory object
func NewSubroundsFactory(args ArgsSovereignSubRoundsFactory) (*factory, error) {
	// no need to check the outportHandler, it can be nil
	err := checkNewFactoryParams(
		args.ConsensusDataContainer,
		args.ConsensusState,
		args.Worker,
		args.BaseSubRoundsFactory,
		args.OutGoingOperationsPool,
		args.BridgeOpHandler,
	)
	if err != nil {
		return nil, err
	}

	fct := factory{
		consensusCore:          args.ConsensusDataContainer,
		consensusState:         args.ConsensusState,
		worker:                 args.Worker,
		outportHandler:         args.OutportHandler,
		baseSubRoundsFactory:   args.BaseSubRoundsFactory,
		outGoingOperationsPool: args.OutGoingOperationsPool,
		bridgeOpHandler:        args.BridgeOpHandler,
	}

	return &fct, nil
}

func checkNewFactoryParams(
	container spos.ConsensusCoreHandler,
	state spos.ConsensusStateHandler,
	worker spos.WorkerHandler,
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
	if baseSubRoundsFactory == nil {
		return ErrNilSubRoundsFactoryInSovereign
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
func (fct *factory) GenerateSubrounds(epoch uint32) error {
	fct.initConsensusThreshold(epoch)
	fct.consensusCore.Chronology().RemoveAllSubrounds()
	fct.worker.RemoveAllReceivedMessagesCalls()
	fct.worker.RemoveAllReceivedHeaderHandlers()

	err := fct.generateStartRoundSubround()
	if err != nil {
		return err
	}

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
}

func (fct *factory) generateStartRoundSubround() error {
	subroundStartRoundInstance, err := fct.baseSubRoundsFactory.GenerateStartRoundSubround()
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

	fct.worker.ResetHandlers(bls.MtBlockHeaderFinalInfo)

	// TODO: MX-17039 Here we need to:
	// 1. Remove handler for base received proof from cns v2
	// 2. Add aggregated sig for outgoing ops in proof to be received on sovereign sub round handler
	// 3. Implement sov sub round handler to received proofs, which should also call base receivedProof handler
	fct.worker.AddReceivedProofHandler(sovEndRound.ReceivedProof)

	fct.worker.AddReceivedMessageCall(bls.MtBlockHeaderFinalInfo, sovEndRound.receivedBlockHeaderFinalInfo)
	fct.consensusCore.Chronology().AddSubround(sovEndRound)

	return nil
}

func (fct *factory) initConsensusThreshold(epoch uint32) {
	consensusGroupSizeForEpoch := fct.consensusCore.NodesCoordinator().ConsensusGroupSizeForShardAndEpoch(fct.consensusCore.ShardCoordinator().SelfId(), epoch)
	pBFTThreshold := core.GetPBFTThreshold(consensusGroupSizeForEpoch)
	pBFTFallbackThreshold := core.GetPBFTFallbackThreshold(consensusGroupSizeForEpoch)
	fct.consensusState.SetThreshold(bls.SrBlock, 1)
	fct.consensusState.SetThreshold(bls.SrSignature, pBFTThreshold)
	fct.consensusState.SetFallbackThreshold(bls.SrBlock, 1)
	fct.consensusState.SetFallbackThreshold(bls.SrSignature, pBFTFallbackThreshold)
}

// IsInterfaceNil returns true if there is no value under the interface
func (fct *factory) IsInterfaceNil() bool {
	return fct == nil
}
