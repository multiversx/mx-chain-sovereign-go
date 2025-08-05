package sovereign_test

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls/sovereign"
	v1 "github.com/multiversx/mx-chain-go/consensus/spos/bls/v1"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/outport"
	"github.com/multiversx/mx-chain-go/testscommon"
	consensusMock "github.com/multiversx/mx-chain-go/testscommon/consensus"
	"github.com/multiversx/mx-chain-go/testscommon/consensus/initializers"
	testscommonOutport "github.com/multiversx/mx-chain-go/testscommon/outport"
	sovTests "github.com/multiversx/mx-chain-go/testscommon/sovereign"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	"github.com/multiversx/mx-chain-go/testscommon/subRoundsHolder"
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

func getFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
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

func TestFactory_NewFactoryShouldWork(t *testing.T) {
	t.Parallel()

	args := createArgsSovSubRoundsFactory()
	fct, err := sovereign.NewSubroundsFactory(args)

	require.Nil(t, err)
	require.False(t, check.IfNil(fct))
}

func TestFactory_GenerateSubroundBlock(t *testing.T) {
	t.Parallel()

	t.Run("should fail when new subround fails", func(t *testing.T) {
		t.Parallel()

		args := createArgsSovSubRoundsFactory()
		args.Worker.(*consensusMock.SposWorkerMock).GetConsensusStateChangedChannelsCalled = func() chan bool {
			return nil
		}

		fct, err := sovereign.NewSubroundsFactory(args)
		require.Nil(t, err)

		err = fct.GenerateBlockSubroundV2()
		require.NotNil(t, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		var addedSubround consensus.SubroundHandler
		container := consensusMock.InitConsensusCore()
		container.SetChronology(&consensusMock.ChronologyHandlerMock{
			AddSubroundCalled: func(handler consensus.SubroundHandler) {
				addedSubround = handler
			},
		})
		fct := *initFactoryWithContainer(container)

		err := fct.GenerateBlockSubroundV2()
		assert.Nil(t, err)
		assert.Equal(t, "*sovereign.subroundBlockV2", fmt.Sprintf("%T", addedSubround))
	})
}

func TestFactory_GenerateSubroundSignature(t *testing.T) {
	t.Parallel()

	t.Run("should fail when new subround fails", func(t *testing.T) {
		t.Parallel()

		args := createArgsSovSubRoundsFactory()
		args.Worker.(*consensusMock.SposWorkerMock).GetConsensusStateChangedChannelsCalled = func() chan bool {
			return nil
		}

		fct, err := sovereign.NewSubroundsFactory(args)
		require.Nil(t, err)

		err = fct.GenerateSignatureSubroundV2()
		require.Equal(t, spos.ErrNilChannel, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		var addedSubround consensus.SubroundHandler
		container := consensusMock.InitConsensusCore()
		container.SetChronology(&consensusMock.ChronologyHandlerMock{
			AddSubroundCalled: func(handler consensus.SubroundHandler) {
				addedSubround = handler
			},
		})
		fct := *initFactoryWithContainer(container)

		err := fct.GenerateSignatureSubroundV2()
		assert.Nil(t, err)
		assert.Equal(t, "*sovereign.subroundSignatureV2", fmt.Sprintf("%T", addedSubround))
	})
}

func TestFactory_GenerateSubroundEndRound(t *testing.T) {
	t.Parallel()

	t.Run("should fail when new subround fails", func(t *testing.T) {
		t.Parallel()

		args := createArgsSovSubRoundsFactory()
		args.Worker.(*consensusMock.SposWorkerMock).GetConsensusStateChangedChannelsCalled = func() chan bool {
			return nil
		}

		fct, err := sovereign.NewSubroundsFactory(args)
		require.Nil(t, err)

		err = fct.GenerateEndRoundSubroundV2()
		require.Equal(t, spos.ErrNilChannel, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		var addedSubround consensus.SubroundHandler
		container := consensusMock.InitConsensusCore()
		container.SetChronology(&consensusMock.ChronologyHandlerMock{
			AddSubroundCalled: func(handler consensus.SubroundHandler) {
				addedSubround = handler
			},
		})

		wasReceivedMsgCalled := false
		shouldCheckAddReceivedMsg := false
		resetReceivedProofHandlerCalled := false
		shouldCheckAddReceivedProof := false
		wasAddReceivedProofHandlerCalled := false
		args := createArgsSovSubRoundsFactory()
		args.ConsensusDataContainer = container
		args.Worker = &consensusMock.SposWorkerMock{
			ResetHandlersCalled: func(messageType consensus.MessageType) {
				require.Equal(t, bls.MtBlockHeaderFinalInfo, messageType)
				shouldCheckAddReceivedMsg = true
			},
			AddReceivedMessageCallCalled: func(messageType consensus.MessageType, receivedMessageCall func(ctx context.Context, cnsDta *consensus.Message) bool) {
				if !shouldCheckAddReceivedMsg {
					return
				}

				require.Equal(t, bls.MtBlockHeaderFinalInfo, messageType)
				require.True(t, strings.Contains(getFunctionName(receivedMessageCall), "(*sovereignSubRoundEnd).receivedBlockHeaderFinalInfo"))

				wasReceivedMsgCalled = true
			},
			ResetReceivedProofHandlerCalled: func() {
				resetReceivedProofHandlerCalled = true
				shouldCheckAddReceivedProof = true
			},
			AddReceivedProofHandlerCalled: func(handler func(proofHandler consensus.ProofHandler)) {
				if !shouldCheckAddReceivedProof {
					return
				}

				require.True(t, strings.Contains(getFunctionName(handler), "(*sovereignSubRoundEnd).ReceivedProof"))
				wasAddReceivedProofHandlerCalled = true
			},
		}

		fct, _ := sovereign.NewSubroundsFactory(args)

		err := fct.GenerateEndRoundSubroundV2()
		require.Nil(t, err)
		require.Equal(t, "*sovereign.sovereignSubRoundEnd", fmt.Sprintf("%T", addedSubround))
		require.True(t, wasReceivedMsgCalled)
		require.True(t, resetReceivedProofHandlerCalled)
		require.True(t, wasAddReceivedProofHandlerCalled)
	})

}

func TestFactory_GenerateSubroundsShouldWork(t *testing.T) {
	t.Parallel()

	chrm := &consensusMock.ChronologyHandlerMock{}
	subRoundsMap := make(map[string]struct{})
	chrm.AddSubroundCalled = func(subroundHandler consensus.SubroundHandler) {
		subRoundHandlerName := fmt.Sprintf("%T", subroundHandler)
		if !strings.Contains(subRoundHandlerName, "StartRound") {
			require.True(t, strings.Contains(subRoundHandlerName, "sovereign"))
		}

		subRoundsMap[subRoundHandlerName] = struct{}{}
	}
	container := consensusMock.InitConsensusCore()
	container.SetChronology(chrm)
	fct := *initFactoryWithContainer(container)
	fct.SetOutportHandler(&testscommonOutport.OutportStub{})

	err := fct.GenerateSubrounds(0)
	require.Nil(t, err)
	require.Len(t, subRoundsMap, 4)
}

func TestFactory_GenerateSubroundsNilOutportShouldFail(t *testing.T) {
	t.Parallel()

	container := consensusMock.InitConsensusCore()
	fct := *initFactoryWithContainer(container)

	err := fct.GenerateSubrounds(0)
	assert.Equal(t, outport.ErrNilDriver, err)
}

func TestFactory_SetIndexerShouldWork(t *testing.T) {
	t.Parallel()

	container := consensusMock.InitConsensusCore()
	fct := *initFactoryWithContainer(container)

	outportHandler := &testscommonOutport.OutportStub{}
	fct.SetOutportHandler(outportHandler)

	require.Equal(t, outportHandler, fct.Outport())
}
