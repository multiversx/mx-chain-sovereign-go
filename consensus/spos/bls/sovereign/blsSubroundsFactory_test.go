package sovereign_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls/sovereign"
	v1 "github.com/multiversx/mx-chain-go/consensus/spos/bls/v1"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/testscommon"
	consensusMock "github.com/multiversx/mx-chain-go/testscommon/consensus"
	"github.com/multiversx/mx-chain-go/testscommon/consensus/initializers"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	sovTests "github.com/multiversx/mx-chain-go/testscommon/sovereign"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	"github.com/multiversx/mx-chain-go/testscommon/subRoundsHolder"
	"github.com/stretchr/testify/require"
)

const processingThresholdPercent = 85

var chainID = []byte("chain ID")

const currentPid = core.PeerID("pid")

const roundTimeDuration = 100 * time.Millisecond

func displayStatistics() {
}

func extend(subroundId int) {
	fmt.Println(subroundId)
}

// executeStoredMessages tries to execute all the messages received which are valid for execution
func executeStoredMessages() {
}

// resetConsensusMessages resets at the start of each round, all the previous consensus messages received
func resetConsensusMessages() {
}

func initRoundHandlerMock() *consensusMock.RoundHandlerMock {
	return &consensusMock.RoundHandlerMock{
		RoundIndex: 0,
		TimeStampCalled: func() time.Time {
			return time.Unix(0, 0)
		},
		TimeDurationCalled: func() time.Duration {
			return roundTimeDuration
		},
	}
}

func initWorker() spos.WorkerHandler {
	sposWorker := &consensusMock.SposWorkerMock{}
	sposWorker.GetConsensusStateChangedChannelsCalled = func() chan bool {
		return make(chan bool)
	}
	sposWorker.RemoveAllReceivedMessagesCallsCalled = func() {}

	sposWorker.AddReceivedMessageCallCalled =
		func(messageType consensus.MessageType, receivedMessageCall func(ctx context.Context, cnsDta *consensus.Message) bool) {
		}

	return sposWorker
}

func createArgsSovSubRoundsFactory() sovereign.ArgsSovereignSubRoundsFactory {
	container := consensusMock.InitConsensusCore()
	consensusState := initializers.InitConsensusState()
	worker := initWorker()

	baseFactory := initFactoryV1(container, consensusState, worker)

	return sovereign.ArgsSovereignSubRoundsFactory{
		ConsensusDataContainer: container,
		ConsensusState:         consensusState,
		Worker:                 worker,
		OutportHandler:         nil,
		ConsensusModel:         consensus.ConsensusModelV2,
		EnableEpochHandler:     &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		BaseSubRoundsFactory:   baseFactory,
		OutGoingOperationsPool: &sovTests.OutGoingOperationsPoolMock{},
		BridgeOpHandler:        &sovTests.BridgeOperationsHandlerMock{},
	}
}

func initFactoryV1(
	container spos.ConsensusCoreHandler,
	consensusState spos.ConsensusStateHandler,
	worker spos.WorkerHandler,
) sovereign.SubRoundsFactoryHandler {
	fct, _ := v1.NewSubroundsFactory(
		container,
		consensusState,
		worker,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
		&testscommon.SentSignatureTrackerStub{},
		nil,
		consensus.ConsensusModelV1,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&subRoundsHolder.ExtraSignersHolderMock{},
	)

	return fct
}

func initFactoryWithContainer(container *spos.ConsensusCore) sovereign.Factory {
	args := createArgsSovSubRoundsFactory()
	args.ConsensusDataContainer = container
	fct, _ := sovereign.NewSubroundsFactory(args)

	return fct
}

func initFactory() sovereign.Factory {
	container := consensusMock.InitConsensusCore()
	return initFactoryWithContainer(container)
}

func TestFactory_NewFactoryNilContainerShouldFail(t *testing.T) {
	t.Parallel()

	args := createArgsSovSubRoundsFactory()
	args.ConsensusDataContainer = nil

	fct, err := sovereign.NewSubroundsFactory(args)

	require.Nil(t, fct)
	require.Equal(t, spos.ErrNilConsensusCore, err)
}

func TestFactory_NewFactoryNilConsensusStateShouldFail(t *testing.T) {
	t.Parallel()

	args := createArgsSovSubRoundsFactory()
	args.ConsensusState = nil
	fct, err := sovereign.NewSubroundsFactory(args)

	require.Nil(t, fct)
	require.Equal(t, spos.ErrNilConsensusState, err)
}

func TestFactory_NewFactoryNilBlockchainShouldFail(t *testing.T) {
	t.Parallel()
	container := consensusMock.InitConsensusCore()
	container.SetBlockchain(nil)

	args := createArgsSovSubRoundsFactory()
	args.ConsensusDataContainer = container
	fct, err := sovereign.NewSubroundsFactory(args)

	require.Nil(t, fct)
	require.Equal(t, spos.ErrNilBlockChain, err)
}

func TestFactory_NewFactoryNilWorkerShouldFail(t *testing.T) {
	t.Parallel()

	args := createArgsSovSubRoundsFactory()
	args.Worker = nil
	fct, err := sovereign.NewSubroundsFactory(args)

	require.Nil(t, fct)
	require.Equal(t, spos.ErrNilWorker, err)
}

func TestFactory_NewFactoryNilEnableEpochHandlerShouldFail(t *testing.T) {
	t.Parallel()

	args := createArgsSovSubRoundsFactory()
	args.EnableEpochHandler = nil
	fct, err := sovereign.NewSubroundsFactory(args)

	require.Nil(t, fct)
	require.Equal(t, spos.ErrNilEnableEpochHandler, err)
}

func TestFactory_NewFactoryNilBaseFactoryShouldFail(t *testing.T) {
	t.Parallel()

	args := createArgsSovSubRoundsFactory()
	args.BaseSubRoundsFactory = nil
	fct, err := sovereign.NewSubroundsFactory(args)

	require.Nil(t, fct)
	require.Equal(t, sovereign.ErrNilSubRoundsFactoryInSovereign, err)
}

func TestFactory_NewFactoryNilOutGoingOperationsPoolShouldFail(t *testing.T) {
	t.Parallel()

	args := createArgsSovSubRoundsFactory()
	args.OutGoingOperationsPool = nil
	fct, err := sovereign.NewSubroundsFactory(args)

	require.Nil(t, fct)
	require.Equal(t, errors.ErrNilOutGoingOperationsPool, err)
}

func TestFactory_NewFactoryNilBridgeOpHandlerShouldFail(t *testing.T) {
	t.Parallel()

	args := createArgsSovSubRoundsFactory()
	args.BridgeOpHandler = nil
	fct, err := sovereign.NewSubroundsFactory(args)

	require.Nil(t, fct)
	require.Equal(t, errors.ErrNilBridgeOpHandler, err)
}

/*
func TestFactory_NewFactoryShouldWork(t *testing.T) {
	t.Parallel()

	fct := *initFactory()

	assert.False(t, check.IfNil(&fct))
}



func TestFactory_NewFactoryEmptyChainIDShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initializers.InitConsensusState()
	container := consensusMock.InitConsensusCore()
	worker := initWorker()

	fct, err := v1.NewSubroundsFactory(
		container,
		consensusState,
		worker,
		nil,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
		&testscommon.SentSignatureTrackerStub{},
		nil,
		consensus.ConsensusModelV1,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&subRoundsHolder.ExtraSignersHolderMock{},
	)

	assert.Nil(t, fct)
	assert.Equal(t, spos.ErrInvalidChainID, err)
}

func TestFactory_GenerateSubroundStartRoundShouldFailWhenNewSubroundFail(t *testing.T) {
	t.Parallel()

	fct := *initFactory()
	fct.Worker().(*consensusMock.SposWorkerMock).GetConsensusStateChangedChannelsCalled = func() chan bool {
		return nil
	}

	err := fct.GenerateStartRoundSubround()

	assert.Equal(t, spos.ErrNilChannel, err)
}

func TestFactory_GenerateSubroundStartRoundShouldFailWhenNewSubroundStartRoundFail(t *testing.T) {
	t.Parallel()

	container := consensusMock.InitConsensusCore()
	fct := *initFactoryWithContainer(container)
	container.SetSyncTimer(nil)

	err := fct.GenerateStartRoundSubround()

	assert.Equal(t, spos.ErrNilSyncTimer, err)
}

func TestFactory_GenerateSubroundBlock(t *testing.T) {
	t.Parallel()

	t.Run("should fail when new subround fails", func(t *testing.T) {
		t.Parallel()

		fct := *initFactory()
		fct.Worker().(*consensusMock.SposWorkerMock).GetConsensusStateChangedChannelsCalled = func() chan bool {
			return nil
		}

		err := fct.GenerateBlockSubroundV1()
		assert.NotNil(t, err)

		err = fct.GenerateBlockSubroundV2()
		assert.NotNil(t, err)
	})
	t.Run("should work with v1", func(t *testing.T) {
		t.Parallel()

		var addedSubround consensus.SubroundHandler
		container := consensusMock.InitConsensusCore()
		container.SetChronology(&mock.ChronologyHandlerMock{
			AddSubroundCalled: func(handler consensus.SubroundHandler) {
				addedSubround = handler
			},
		})
		fct := *initFactoryWithContainer(container)

		err := fct.GenerateBlockSubroundV1()
		assert.Nil(t, err)
		assert.Equal(t, "*bls.subroundBlock", fmt.Sprintf("%T", addedSubround))
	})
	t.Run("should work with v2", func(t *testing.T) {
		t.Parallel()

		var addedSubround consensus.SubroundHandler
		container := mock.InitConsensusCore()
		container.SetChronology(&mock.ChronologyHandlerMock{
			AddSubroundCalled: func(handler consensus.SubroundHandler) {
				addedSubround = handler
			},
		})
		worker := initWorker()
		consensusState := initConsensusState()

		fct, _ := bls.NewSubroundsFactory(
			container,
			consensusState,
			worker,
			chainID,
			currentPid,
			&statusHandler.AppStatusHandlerStub{},
			&testscommon.SentSignatureTrackerStub{},
			consensus.ConsensusModelV2,
			&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			&subRoundsHolder.ExtraSignersHolderMock{},
		)

		err := fct.GenerateBlockSubroundV2()
		assert.Nil(t, err)
		assert.Equal(t, "*bls.subroundBlockV2", fmt.Sprintf("%T", addedSubround))
	})
}

func TestFactory_GenerateSubroundSignature(t *testing.T) {
	t.Parallel()

	t.Run("should fail when new subround fails", func(t *testing.T) {
		t.Parallel()

		fct := *initFactory()
		fct.Worker().(*consensusMock.SposWorkerMock).GetConsensusStateChangedChannelsCalled = func() chan bool {
			return nil
		}

		err := fct.GenerateSignatureSubroundV1()
		assert.Equal(t, spos.ErrNilChannel, err)

		err = fct.GenerateSignatureSubroundV2()
		assert.Equal(t, spos.ErrNilChannel, err)
	})
	t.Run("should fail when new subround Signature fails", func(t *testing.T) {
		t.Parallel()

		container := consensusMock.InitConsensusCore()
		fct := *initFactoryWithContainer(container)
		container.SetSyncTimer(nil)

		err := fct.GenerateSignatureSubroundV1()
		assert.Equal(t, spos.ErrNilSyncTimer, err)

		err = fct.GenerateSignatureSubroundV2()
		assert.Equal(t, spos.ErrNilSyncTimer, err)
	})
	t.Run("should work with v1", func(t *testing.T) {
		t.Parallel()

		var addedSubround consensus.SubroundHandler
		container := mock.InitConsensusCore()
		container.SetChronology(&mock.ChronologyHandlerMock{
			AddSubroundCalled: func(handler consensus.SubroundHandler) {
				addedSubround = handler
			},
		})
		fct := *initFactoryWithContainer(container)

		err := fct.GenerateSignatureSubroundV1()
		assert.Nil(t, err)
		assert.Equal(t, "*bls.subroundSignature", fmt.Sprintf("%T", addedSubround))
	})
	t.Run("should work with v2", func(t *testing.T) {
		t.Parallel()

		var addedSubround consensus.SubroundHandler
		container := mock.InitConsensusCore()
		container.SetChronology(&mock.ChronologyHandlerMock{
			AddSubroundCalled: func(handler consensus.SubroundHandler) {
				addedSubround = handler
			},
		})
		worker := initWorker()
		consensusState := initConsensusState()

		fct, _ := bls.NewSubroundsFactory(
			container,
			consensusState,
			worker,
			chainID,
			currentPid,
			&statusHandler.AppStatusHandlerStub{},
			&testscommon.SentSignatureTrackerStub{},
			consensus.ConsensusModelV2,
			&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			&subRoundsHolder.ExtraSignersHolderMock{},
		)

		err := fct.GenerateSignatureSubroundV2()
		assert.Nil(t, err)
		assert.Equal(t, "*bls.subroundSignatureV2", fmt.Sprintf("%T", addedSubround))
	})
}

func TestFactory_GenerateSubroundEndRound(t *testing.T) {
	t.Parallel()

	t.Run("should fail when new subround fails", func(t *testing.T) {
		t.Parallel()

		fct := *initFactory()
		fct.Worker().(*consensusMock.SposWorkerMock).GetConsensusStateChangedChannelsCalled = func() chan bool {
			return nil
		}

		err := fct.GenerateEndRoundSubroundV1()
		assert.Equal(t, spos.ErrNilChannel, err)

		err = fct.GenerateEndRoundSubroundV2()
		assert.Equal(t, spos.ErrNilChannel, err)
	})
	t.Run("should fail when new subround EndRound fails", func(t *testing.T) {
		t.Parallel()

		container := consensusMock.InitConsensusCore()
		fct := *initFactoryWithContainer(container)
		container.SetSyncTimer(nil)

		err := fct.GenerateEndRoundSubroundV1()
		assert.Equal(t, spos.ErrNilSyncTimer, err)

		err = fct.GenerateEndRoundSubroundV2()
		assert.Equal(t, spos.ErrNilSyncTimer, err)
	})
	t.Run("should work with v1", func(t *testing.T) {
		t.Parallel()

		var addedSubround consensus.SubroundHandler
		container := mock.InitConsensusCore()
		container.SetChronology(&mock.ChronologyHandlerMock{
			AddSubroundCalled: func(handler consensus.SubroundHandler) {
				addedSubround = handler
			},
		})
		fct := *initFactoryWithContainer(container)

		err := fct.GenerateEndRoundSubroundV1()
		assert.Nil(t, err)
		assert.Equal(t, "*bls.subroundEndRound", fmt.Sprintf("%T", addedSubround))
	})
	t.Run("should work with v2", func(t *testing.T) {
		t.Parallel()

		var addedSubround consensus.SubroundHandler
		container := mock.InitConsensusCore()
		container.SetChronology(&mock.ChronologyHandlerMock{
			AddSubroundCalled: func(handler consensus.SubroundHandler) {
				addedSubround = handler
			},
		})
		worker := initWorker()
		consensusState := initializers.InitConsensusState()

		fct, _ := bls.NewSubroundsFactory(
			container,
			consensusState,
			worker,
			chainID,
			currentPid,
			&statusHandler.AppStatusHandlerStub{},
			&testscommon.SentSignatureTrackerStub{},
			consensus.ConsensusModelV2,
			&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			&subRoundsHolder.ExtraSignersHolderMock{},
		)

		err := fct.GenerateEndRoundSubroundV2()
		assert.Nil(t, err)
		assert.Equal(t, "*bls.subroundEndRoundV2", fmt.Sprintf("%T", addedSubround))
	})
}

func TestFactory_GenerateSubroundsShouldWork(t *testing.T) {
	t.Parallel()

	subroundHandlers := 0

	chrm := &consensusMock.ChronologyHandlerMock{}
	subRoundsMap := make(map[string]struct{})
	chrm.AddSubroundCalled = func(subroundHandler consensus.SubroundHandler) {
		subRoundsMap[fmt.Sprintf("%T", subroundHandler)] = struct{}{}
		subroundHandlers++
	}
	container := consensusMock.InitConsensusCore()
	container.SetChronology(chrm)
	fct := *initFactoryWithContainer(container)
	fct.SetOutportHandler(&testscommonOutport.OutportStub{})

	err := fct.GenerateSubrounds(0)
	require.Nil(t, err)

	require.Equal(t, 4, subroundHandlers)
	require.Len(t, subRoundsMap, 4)
}

func TestFactory_GenerateSubroundsNilOutportShouldFail(t *testing.T) {
	t.Parallel()

	container := consensusMock.InitConsensusCore()
	fct := *initFactoryWithContainer(container)

	err := fct.GenerateSubrounds(0)
	assert.Equal(t, outport.ErrNilDriver, err)
}

func TestFactory_GenerateSubroundsInvalidConsensusModelShouldFail(t *testing.T) {
	t.Parallel()

	worker := initWorker()
	consensusState := initializers.InitConsensusState()

	fct, _ := bls.NewSubroundsFactory(
		mock.InitConsensusCore(),
		consensusState,
		worker,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
		&testscommon.SentSignatureTrackerStub{},
		"invalid",
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&subRoundsHolder.ExtraSignersHolderMock{},
	)
	fct.SetOutportHandler(&testscommonOutport.OutportStub{})

	err := fct.GenerateSubrounds()
	assert.ErrorIs(t, err, errors.ErrUnimplementedConsensusModel)
}

func TestFactory_SetIndexerShouldWork(t *testing.T) {
	t.Parallel()

	container := consensusMock.InitConsensusCore()
	fct := *initFactoryWithContainer(container)

	outportHandler := &testscommonOutport.OutportStub{}
	fct.SetOutportHandler(outportHandler)

	assert.Equal(t, outportHandler, fct.Outport())
}


*/

// TODO: MARIUS C:
// Refactor these tests to work for this factory

/*



func TestSovereignSubRoundEndV2Creator_CreateAndAddSubRoundEnd(t *testing.T) {
	t.Parallel()

	addReceivedMessageCallCt := 0
	addReceivedHeaderHandlerCallCt := 0
	workerHandler := &cnsMock.SposWorkerMock{
		AddReceivedMessageCallCalled: func(messageType consensus.MessageType, receivedMessageCall func(ctx context.Context, cnsDta *consensus.Message) bool) {
			addReceivedMessageCallCt++
			require.True(t, messageType == bls.MtBlockHeaderFinalInfo || messageType == bls.MtInvalidSigners)
		},
		AddReceivedHeaderHandlerCalled: func(handler func(data.HeaderHandler)) {
			addReceivedHeaderHandlerCallCt++
		},
	}

	addSubRoundCalledCt := 0
	consensusCore := cnsMock.InitConsensusCore()
	consensusCore.SetChronology(&cnsMock.ChronologyHandlerMock{
		AddSubroundCalled: func(handler consensus.SubroundHandler) {
			addSubRoundCalledCt++
			require.Equal(t, "*bls.sovereignSubRoundEnd", fmt.Sprintf("%T", handler))
		},
	})

	sr := initSubroundEndRound(&statusHandler.AppStatusHandlerStub{})

	creator, _ := sovereign2.NewSovereignSubRoundEndCreator(&sovereign.OutGoingOperationsPoolMock{}, &sovereign.BridgeOperationsHandlerMock{})
	err := creator.CreateAndAddSubRoundEnd(sr, workerHandler, consensusCore)
	require.Nil(t, err)
	require.Equal(t, 2, addReceivedMessageCallCt)
	require.Equal(t, 1, addReceivedHeaderHandlerCallCt)
	require.Equal(t, 1, addSubRoundCalledCt)
}

*/
