package sovereign

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"

	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/errors"
)

// TODO: MARIUS C: Delete this

type sovereignSubRoundEndCreator struct {
	outGoingOperationsPool bls.OutGoingOperationsPool
	bridgeOpHandler        bls.BridgeOperationsHandler
}

// NewSovereignSubRoundEndCreator creates a new sovereign subround end factory
func NewSovereignSubRoundEndCreator(
	outGoingOperationsPool bls.OutGoingOperationsPool,
	bridgeOpHandler bls.BridgeOperationsHandler,
) (*sovereignSubRoundEndCreator, error) {
	if check.IfNil(outGoingOperationsPool) {
		return nil, errors.ErrNilOutGoingOperationsPool
	}
	if check.IfNil(bridgeOpHandler) {
		return nil, errors.ErrNilBridgeOpHandler
	}

	return &sovereignSubRoundEndCreator{
		outGoingOperationsPool: outGoingOperationsPool,
		bridgeOpHandler:        bridgeOpHandler,
	}, nil
}

// CreateAndAddSubRoundEnd creates a new sovereign subround end and adds it to the consensus
func (c *sovereignSubRoundEndCreator) CreateAndAddSubRoundEnd(
	subroundEndRoundInstance SubRoundEndHandler,
	worker spos.WorkerHandler,
	consensusCore spos.ConsensusCoreHandler,
) error {
	subroundEndV2Instance, err := NewSubroundEndRoundV2(subroundEndRoundInstance)
	if err != nil {
		return err
	}

	sovEndRound, err := NewSovereignSubRoundEndRound(
		subroundEndV2Instance,
		c.outGoingOperationsPool,
		c.bridgeOpHandler,
	)
	if err != nil {
		return err
	}

	worker.AddReceivedMessageCall(bls.MtBlockHeaderFinalInfo, sovEndRound.receivedBlockHeaderFinalInfo)
	worker.AddReceivedMessageCall(bls.MtInvalidSigners, sovEndRound.receivedInvalidSignersInfo)
	worker.AddReceivedHeaderHandler(sovEndRound.receivedHeader)
	consensusCore.Chronology().AddSubround(sovEndRound)

	return nil
}

// IsInterfaceNil checks if the underlying pointer is nil
func (c *sovereignSubRoundEndCreator) IsInterfaceNil() bool {
	return c == nil
}
