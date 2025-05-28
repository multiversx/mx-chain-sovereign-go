package bls_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
	v1 "github.com/multiversx/mx-chain-go/consensus/spos/bls/v1"
	"github.com/multiversx/mx-chain-go/testscommon"
	consensusMocks "github.com/multiversx/mx-chain-go/testscommon/consensus"
	"github.com/multiversx/mx-chain-go/testscommon/consensus/initializers"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	"github.com/multiversx/mx-chain-go/testscommon/subRounds"
	"github.com/stretchr/testify/assert"
)

var chainID = []byte("chain ID")

const currentPid = core.PeerID("pid")
const processingThresholdPercent = 85

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

func initSubroundEndRoundWithContainer(
	container *spos.ConsensusCore,
	appStatusHandler core.AppStatusHandler,
	enableEpochHandler common.EnableEpochsHandler,
) bls.SubRoundEndHandler {
	ch := make(chan bool, 1)
	consensusState := initializers.InitConsensusState()
	sr, _ := spos.NewSubround(
		bls.SrSignature,
		bls.SrEndRound,
		-1,
		int64(85*roundTimeDuration/100),
		int64(95*roundTimeDuration/100),
		"(END_ROUND)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
		chainID,
		currentPid,
		appStatusHandler,
		enableEpochHandler,
	)

	srEndRound, _ := v1.NewSubroundEndRound(
		sr,
		extend,
		processingThresholdPercent,
		displayStatistics,
		&subRounds.SubRoundEndExtraSignersHolderMock{},
		&statusHandler.AppStatusHandlerStub{},
		&testscommon.SentSignatureTrackerStub{},
	)

	return srEndRound
}

func initSubroundEndRound(appStatusHandler core.AppStatusHandler) bls.SubRoundEndHandler {
	container := consensusMocks.InitConsensusCore()
	return initSubroundEndRoundWithContainer(container, appStatusHandler, &enableEpochsHandlerMock.EnableEpochsHandlerStub{})
}

func TestNewSubroundEndRoundV2_ShouldErrNilSubround(t *testing.T) {
	t.Parallel()

	srV2, err := bls.NewSubroundEndRoundV2(nil)

	assert.Nil(t, srV2)
	assert.Equal(t, spos.ErrNilSubround, err)
}

func TestNewSubroundEndRoundV2_ShouldWork(t *testing.T) {
	t.Parallel()

	sr := initSubroundEndRound(&statusHandler.AppStatusHandlerStub{})
	srV2, err := bls.NewSubroundEndRoundV2(sr)

	assert.NotNil(t, srV2)
	assert.Nil(t, err)
}

func TestSubroundEndRoundV2_GetMessageToVerifySig(t *testing.T) {
	t.Parallel()

	t.Run("getMessageToVerifySig should return nil when CalculateHash method fails", func(t *testing.T) {
		t.Parallel()

		sr := initSubroundEndRound(&statusHandler.AppStatusHandlerStub{})
		srV2, _ := bls.NewSubroundEndRoundV2(sr)

		srV2.SetHeader(nil)
		msg := srV2.GetMessageToVerifySig()

		assert.Nil(t, msg)
	})

	t.Run("getMessageToVerifySig should return the message on which the signature should be verified", func(t *testing.T) {
		t.Parallel()

		sr := initSubroundEndRound(&statusHandler.AppStatusHandlerStub{})
		srV2, _ := bls.NewSubroundEndRoundV2(sr)

		srV2.SetHeader(&block.Header{Nonce: 1})
		expectedMsg, _ := core.CalculateHash(srV2.Marshalizer(), srV2.Hasher(), srV2.GetHeader())
		msg := srV2.GetMessageToVerifySig()

		assert.Equal(t, expectedMsg, msg)
	})
}
