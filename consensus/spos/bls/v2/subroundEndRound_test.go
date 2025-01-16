package v2_test

import (
	"bytes"
	"context"
	"errors"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/mock"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
	v2 "github.com/multiversx/mx-chain-go/consensus/spos/bls/v2"
	"github.com/multiversx/mx-chain-go/dataRetriever/blockchain"
	dataRetrieverMocks "github.com/multiversx/mx-chain-go/dataRetriever/mock"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/p2p/factory"
	"github.com/multiversx/mx-chain-go/testscommon"
	consensusMocks "github.com/multiversx/mx-chain-go/testscommon/consensus"
	"github.com/multiversx/mx-chain-go/testscommon/consensus/initializers"
	"github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
)

func initSubroundEndRoundWithContainer(
	container *consensusMocks.ConsensusCoreMock,
	appStatusHandler core.AppStatusHandler,
) v2.SubroundEndRound {
	ch := make(chan bool, 1)
	consensusState := initializers.InitConsensusStateWithNodesCoordinator(container.NodesCoordinator())
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
	)
	sr.SetHeader(&block.HeaderV2{
		Header: createDefaultHeader(),
	})

	srEndRound, _ := v2.NewSubroundEndRound(
		sr,
		v2.ProcessingThresholdPercent,
		appStatusHandler,
		&testscommon.SentSignatureTrackerStub{},
		&consensusMocks.SposWorkerMock{},
		&dataRetrieverMocks.ThrottlerStub{},
	)

	return srEndRound
}

func initSubroundEndRoundWithContainerAndConsensusState(
	container *consensusMocks.ConsensusCoreMock,
	appStatusHandler core.AppStatusHandler,
	consensusState *spos.ConsensusState,
	signatureThrottler core.Throttler,
) v2.SubroundEndRound {
	ch := make(chan bool, 1)
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
	)
	sr.SetHeader(&block.HeaderV2{
		Header: createDefaultHeader(),
	})

	srEndRound, _ := v2.NewSubroundEndRound(
		sr,
		v2.ProcessingThresholdPercent,
		appStatusHandler,
		&testscommon.SentSignatureTrackerStub{},
		&consensusMocks.SposWorkerMock{},
		signatureThrottler,
	)

	return srEndRound
}

func initSubroundEndRound(appStatusHandler core.AppStatusHandler) v2.SubroundEndRound {
	container := consensusMocks.InitConsensusCore()
	sr := initSubroundEndRoundWithContainer(container, appStatusHandler)
	sr.SetHeader(&block.HeaderV2{
		Header: createDefaultHeader(),
	})
	return sr
}

func TestNewSubroundEndRound(t *testing.T) {
	t.Parallel()

	container := consensusMocks.InitConsensusCore()
	consensusState := initializers.InitConsensusState()
	ch := make(chan bool, 1)
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
		&statusHandler.AppStatusHandlerStub{},
	)

	t.Run("nil subround should error", func(t *testing.T) {
		t.Parallel()

		srEndRound, err := v2.NewSubroundEndRound(
			nil,
			v2.ProcessingThresholdPercent,
			&statusHandler.AppStatusHandlerStub{},
			&testscommon.SentSignatureTrackerStub{},
			&consensusMocks.SposWorkerMock{},
			&dataRetrieverMocks.ThrottlerStub{},
		)

		assert.Nil(t, srEndRound)
		assert.Equal(t, spos.ErrNilSubround, err)
	})
	t.Run("nil app status handler should error", func(t *testing.T) {
		t.Parallel()

		srEndRound, err := v2.NewSubroundEndRound(
			sr,
			v2.ProcessingThresholdPercent,
			nil,
			&testscommon.SentSignatureTrackerStub{},
			&consensusMocks.SposWorkerMock{},
			&dataRetrieverMocks.ThrottlerStub{},
		)

		assert.Nil(t, srEndRound)
		assert.Equal(t, spos.ErrNilAppStatusHandler, err)
	})
	t.Run("nil sent signatures tracker should error", func(t *testing.T) {
		t.Parallel()

		srEndRound, err := v2.NewSubroundEndRound(
			sr,
			v2.ProcessingThresholdPercent,
			&statusHandler.AppStatusHandlerStub{},
			nil,
			&consensusMocks.SposWorkerMock{},
			&dataRetrieverMocks.ThrottlerStub{},
		)

		assert.Nil(t, srEndRound)
		assert.Equal(t, v2.ErrNilSentSignatureTracker, err)
	})
	t.Run("nil worker should error", func(t *testing.T) {
		t.Parallel()

		srEndRound, err := v2.NewSubroundEndRound(
			sr,
			v2.ProcessingThresholdPercent,
			&statusHandler.AppStatusHandlerStub{},
			&testscommon.SentSignatureTrackerStub{},
			nil,
			&dataRetrieverMocks.ThrottlerStub{},
		)

		assert.Nil(t, srEndRound)
		assert.Equal(t, spos.ErrNilWorker, err)
	})
}

func TestSubroundEndRound_NewSubroundEndRoundNilBlockChainShouldFail(t *testing.T) {
	t.Parallel()

	container := consensusMocks.InitConsensusCore()
	consensusState := initializers.InitConsensusState()
	ch := make(chan bool, 1)

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
		&statusHandler.AppStatusHandlerStub{},
	)
	container.SetBlockchain(nil)
	srEndRound, err := v2.NewSubroundEndRound(
		sr,
		v2.ProcessingThresholdPercent,
		&statusHandler.AppStatusHandlerStub{},
		&testscommon.SentSignatureTrackerStub{},
		&consensusMocks.SposWorkerMock{},
		&dataRetrieverMocks.ThrottlerStub{},
	)

	assert.True(t, check.IfNil(srEndRound))
	assert.Equal(t, spos.ErrNilBlockChain, err)
}

func TestSubroundEndRound_NewSubroundEndRoundNilBlockProcessorShouldFail(t *testing.T) {
	t.Parallel()

	container := consensusMocks.InitConsensusCore()
	consensusState := initializers.InitConsensusState()
	ch := make(chan bool, 1)

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
		&statusHandler.AppStatusHandlerStub{},
	)
	container.SetBlockProcessor(nil)
	srEndRound, err := v2.NewSubroundEndRound(
		sr,
		v2.ProcessingThresholdPercent,
		&statusHandler.AppStatusHandlerStub{},
		&testscommon.SentSignatureTrackerStub{},
		&consensusMocks.SposWorkerMock{},
		&dataRetrieverMocks.ThrottlerStub{},
	)

	assert.True(t, check.IfNil(srEndRound))
	assert.Equal(t, spos.ErrNilBlockProcessor, err)
}

func TestSubroundEndRound_NewSubroundEndRoundNilConsensusStateShouldFail(t *testing.T) {
	t.Parallel()

	container := consensusMocks.InitConsensusCore()
	consensusState := initializers.InitConsensusState()
	ch := make(chan bool, 1)

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
		&statusHandler.AppStatusHandlerStub{},
	)

	sr.ConsensusStateHandler = nil
	srEndRound, err := v2.NewSubroundEndRound(
		sr,
		v2.ProcessingThresholdPercent,
		&statusHandler.AppStatusHandlerStub{},
		&testscommon.SentSignatureTrackerStub{},
		&consensusMocks.SposWorkerMock{},
		&dataRetrieverMocks.ThrottlerStub{},
	)

	assert.True(t, check.IfNil(srEndRound))
	assert.Equal(t, spos.ErrNilConsensusState, err)
}

func TestSubroundEndRound_NewSubroundEndRoundNilMultiSignerContainerShouldFail(t *testing.T) {
	t.Parallel()

	container := consensusMocks.InitConsensusCore()
	consensusState := initializers.InitConsensusState()
	ch := make(chan bool, 1)

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
		&statusHandler.AppStatusHandlerStub{},
	)
	container.SetMultiSignerContainer(nil)
	srEndRound, err := v2.NewSubroundEndRound(
		sr,
		v2.ProcessingThresholdPercent,
		&statusHandler.AppStatusHandlerStub{},
		&testscommon.SentSignatureTrackerStub{},
		&consensusMocks.SposWorkerMock{},
		&dataRetrieverMocks.ThrottlerStub{},
	)

	assert.True(t, check.IfNil(srEndRound))
	assert.Equal(t, spos.ErrNilMultiSignerContainer, err)
}

func TestSubroundEndRound_NewSubroundEndRoundNilRoundHandlerShouldFail(t *testing.T) {
	t.Parallel()

	container := consensusMocks.InitConsensusCore()
	consensusState := initializers.InitConsensusState()
	ch := make(chan bool, 1)

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
		&statusHandler.AppStatusHandlerStub{},
	)
	container.SetRoundHandler(nil)
	srEndRound, err := v2.NewSubroundEndRound(
		sr,
		v2.ProcessingThresholdPercent,
		&statusHandler.AppStatusHandlerStub{},
		&testscommon.SentSignatureTrackerStub{},
		&consensusMocks.SposWorkerMock{},
		&dataRetrieverMocks.ThrottlerStub{},
	)

	assert.True(t, check.IfNil(srEndRound))
	assert.Equal(t, spos.ErrNilRoundHandler, err)
}

func TestSubroundEndRound_NewSubroundEndRoundNilSyncTimerShouldFail(t *testing.T) {
	t.Parallel()

	container := consensusMocks.InitConsensusCore()
	consensusState := initializers.InitConsensusState()
	ch := make(chan bool, 1)

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
		&statusHandler.AppStatusHandlerStub{},
	)
	container.SetSyncTimer(nil)
	srEndRound, err := v2.NewSubroundEndRound(
		sr,
		v2.ProcessingThresholdPercent,
		&statusHandler.AppStatusHandlerStub{},
		&testscommon.SentSignatureTrackerStub{},
		&consensusMocks.SposWorkerMock{},
		&dataRetrieverMocks.ThrottlerStub{},
	)

	assert.True(t, check.IfNil(srEndRound))
	assert.Equal(t, spos.ErrNilSyncTimer, err)
}

func TestSubroundEndRound_NewSubroundEndRoundNilThrottlerShouldFail(t *testing.T) {
	t.Parallel()

	container := consensusMocks.InitConsensusCore()
	consensusState := initializers.InitConsensusState()
	ch := make(chan bool, 1)

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
		&statusHandler.AppStatusHandlerStub{},
	)

	srEndRound, err := v2.NewSubroundEndRound(
		sr,
		v2.ProcessingThresholdPercent,
		&statusHandler.AppStatusHandlerStub{},
		&testscommon.SentSignatureTrackerStub{},
		&consensusMocks.SposWorkerMock{},
		nil,
	)

	assert.True(t, check.IfNil(srEndRound))
	assert.Equal(t, err, spos.ErrNilThrottler)
}

func TestSubroundEndRound_NewSubroundEndRoundShouldWork(t *testing.T) {
	t.Parallel()

	container := consensusMocks.InitConsensusCore()
	consensusState := initializers.InitConsensusState()
	ch := make(chan bool, 1)

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
		&statusHandler.AppStatusHandlerStub{},
	)

	srEndRound, err := v2.NewSubroundEndRound(
		sr,
		v2.ProcessingThresholdPercent,
		&statusHandler.AppStatusHandlerStub{},
		&testscommon.SentSignatureTrackerStub{},
		&consensusMocks.SposWorkerMock{},
		&dataRetrieverMocks.ThrottlerStub{},
	)

	assert.False(t, check.IfNil(srEndRound))
	assert.Nil(t, err)
}

func TestSubroundEndRound_DoEndRoundJobNilHeaderShouldFail(t *testing.T) {
	t.Parallel()

	container := consensusMocks.InitConsensusCore()
	sr := initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})
	sr.SetHeader(nil)

	r := sr.DoEndRoundJob()
	assert.False(t, r)
}

func TestSubroundEndRound_DoEndRoundJobErrAggregatingSigShouldFail(t *testing.T) {
	t.Parallel()
	container := consensusMocks.InitConsensusCore()
	sr := initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})

	signingHandler := &consensusMocks.SigningHandlerStub{
		AggregateSigsCalled: func(bitmap []byte, epoch uint32) ([]byte, error) {
			return nil, crypto.ErrNilHasher
		},
	}
	container.SetSigningHandler(signingHandler)

	sr.SetHeader(&block.Header{})

	sr.SetSelfPubKey("A")

	assert.True(t, sr.IsSelfLeader())
	r := sr.DoEndRoundJob()
	assert.False(t, r)
}

func TestSubroundEndRound_DoEndRoundJobErrCommitBlockShouldFail(t *testing.T) {
	t.Parallel()

	container := consensusMocks.InitConsensusCore()
	sr := initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})
	sr.SetSelfPubKey("A")

	blProcMock := consensusMocks.InitBlockProcessorMock(container.Marshalizer())
	blProcMock.CommitBlockCalled = func(
		header data.HeaderHandler,
		body data.BodyHandler,
	) error {
		return blockchain.ErrHeaderUnitNil
	}

	container.SetBlockProcessor(blProcMock)
	sr.SetHeader(&block.Header{})

	r := sr.DoEndRoundJob()
	assert.False(t, r)
}

func TestSubroundEndRound_DoEndRoundJobErrTimeIsOutShouldFail(t *testing.T) {
	t.Parallel()

	container := consensusMocks.InitConsensusCore()
	sr := initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})
	sr.SetSelfPubKey("A")

	remainingTime := -time.Millisecond
	roundHandlerMock := &consensusMocks.RoundHandlerMock{
		RemainingTimeCalled: func(startTime time.Time, maxTime time.Duration) time.Duration {
			return remainingTime
		},
	}

	container.SetRoundHandler(roundHandlerMock)
	sr.SetHeader(&block.Header{})

	r := sr.DoEndRoundJob()
	assert.False(t, r)
}

func TestSubroundEndRound_DoEndRoundJobAllOK(t *testing.T) {
	t.Parallel()

	container := consensusMocks.InitConsensusCore()
	sr := initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})
	sr.SetSelfPubKey("A")

	sr.SetHeader(&block.Header{})

	for _, participant := range sr.ConsensusGroup() {
		_ = sr.SetJobDone(participant, bls.SrSignature, true)
	}

	r := sr.DoEndRoundJob()
	assert.True(t, r)
}

func TestSubroundEndRound_DoEndRoundConsensusCheckShouldReturnFalseWhenRoundIsCanceled(t *testing.T) {
	t.Parallel()

	sr := initSubroundEndRound(&statusHandler.AppStatusHandlerStub{})
	sr.SetRoundCanceled(true)

	ok := sr.DoEndRoundConsensusCheck()
	assert.False(t, ok)
}

func TestSubroundEndRound_DoEndRoundConsensusCheckShouldReturnTrueWhenRoundIsFinished(t *testing.T) {
	t.Parallel()

	sr := initSubroundEndRound(&statusHandler.AppStatusHandlerStub{})
	sr.SetStatus(bls.SrEndRound, spos.SsFinished)

	ok := sr.DoEndRoundConsensusCheck()
	assert.True(t, ok)
}

func TestSubroundEndRound_DoEndRoundConsensusCheckShouldReturnFalseWhenRoundIsNotFinished(t *testing.T) {
	t.Parallel()

	sr := initSubroundEndRound(&statusHandler.AppStatusHandlerStub{})

	ok := sr.DoEndRoundConsensusCheck()
	assert.False(t, ok)
}

func TestSubroundEndRound_CheckSignaturesValidityShouldErrNilSignature(t *testing.T) {
	t.Parallel()

	sr := initSubroundEndRound(&statusHandler.AppStatusHandlerStub{})

	bitmap := make([]byte, len(sr.ConsensusGroup())/8+1)
	bitmap[0] = 0x77
	bitmap[1] = 0x01
	err := sr.CheckSignaturesValidity(bitmap)

	assert.Equal(t, spos.ErrNilSignature, err)
}

func TestSubroundEndRound_CheckSignaturesValidityShouldReturnNil(t *testing.T) {
	t.Parallel()

	sr := initSubroundEndRound(&statusHandler.AppStatusHandlerStub{})

	for _, pubKey := range sr.ConsensusGroup() {
		_ = sr.SetJobDone(pubKey, bls.SrSignature, true)
	}

	bitmap := make([]byte, len(sr.ConsensusGroup())/8+1)
	bitmap[0] = 0x77
	bitmap[1] = 0x01

	err := sr.CheckSignaturesValidity(bitmap)
	require.Nil(t, err)
}

func TestSubroundEndRound_CreateAndBroadcastProofShouldBeCalled(t *testing.T) {
	t.Parallel()

	chanRcv := make(chan bool, 1)
	leaderSigInHdr := []byte("leader sig")
	container := consensusMocks.InitConsensusCore()
	messenger := &consensusMocks.BroadcastMessengerMock{
		BroadcastEquivalentProofCalled: func(proof data.HeaderProofHandler, pkBytes []byte) error {
			chanRcv <- true
			return nil
		},
	}
	container.SetBroadcastMessenger(messenger)
	sr := initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})
	sr.SetHeader(&block.Header{LeaderSignature: leaderSigInHdr})
	sr.CreateAndBroadcastProof([]byte("sig"), []byte("bitmap"))

	select {
	case <-chanRcv:
	case <-time.After(100 * time.Millisecond):
		assert.Fail(t, "broadcast not called")
	}
}

func TestSubroundEndRound_ReceivedProof(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		hdr := &block.Header{Nonce: 37}
		sr := initSubroundEndRound(&statusHandler.AppStatusHandlerStub{})
		sr.SetHeader(hdr)
		sr.AddReceivedHeader(hdr)

		sr.SetStatus(2, spos.SsFinished)
		sr.SetStatus(3, spos.SsNotFinished)

		proof := &block.HeaderProof{}
		sr.ReceivedProof(proof)
	})
	t.Run("should work with equivalent messages flag on", func(t *testing.T) {
		t.Parallel()

		providedPrevSig := []byte("prev sig")
		providedPrevBitmap := []byte{1, 1, 1, 1}
		hdr := &block.HeaderV2{
			Header:                   createDefaultHeader(),
			ScheduledRootHash:        []byte("sch root hash"),
			ScheduledAccumulatedFees: big.NewInt(0),
			ScheduledDeveloperFees:   big.NewInt(0),
			PreviousHeaderProof:      nil,
		}
		container := consensusMocks.InitConsensusCore()
		enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return flag == common.EquivalentMessagesFlag
			},
		}
		container.SetEnableEpochsHandler(enableEpochsHandler)
		container.SetBlockchain(&testscommon.ChainHandlerStub{
			GetGenesisHeaderCalled: func() data.HeaderHandler {
				return &block.HeaderV2{}
			},
		})

		container.SetEquivalentProofsPool(&dataRetriever.ProofsPoolMock{
			GetProofCalled: func(shardID uint32, headerHash []byte) (data.HeaderProofHandler, error) {
				assert.Equal(t, hdr.GetPrevHash(), headerHash)
				return &block.HeaderProof{
					HeaderHash:          headerHash,
					AggregatedSignature: providedPrevSig,
					PubKeysBitmap:       providedPrevBitmap,
				}, nil
			},
		})

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
			&statusHandler.AppStatusHandlerStub{},
		)

		srEndRound, _ := v2.NewSubroundEndRound(
			sr,
			v2.ProcessingThresholdPercent,
			&statusHandler.AppStatusHandlerStub{},
			&testscommon.SentSignatureTrackerStub{},
			&consensusMocks.SposWorkerMock{},
			&dataRetrieverMocks.ThrottlerStub{},
		)

		srEndRound.SetHeader(hdr)
		srEndRound.AddReceivedHeader(hdr)

		srEndRound.SetStatus(2, spos.SsFinished)
		srEndRound.SetStatus(3, spos.SsNotFinished)

		proof := &block.HeaderProof{}
		srEndRound.ReceivedProof(proof)
	})
	t.Run("should return false when header is nil", func(t *testing.T) {
		t.Parallel()

		sr := initSubroundEndRound(&statusHandler.AppStatusHandlerStub{})
		sr.SetHeader(nil)

		proof := &block.HeaderProof{}

		sr.ReceivedProof(proof)
	})
	t.Run("should return false when final info is not valid", func(t *testing.T) {
		t.Parallel()

		container := consensusMocks.InitConsensusCore()

		headerSigVerifier := &consensusMocks.HeaderSigVerifierMock{
			VerifyLeaderSignatureCalled: func(header data.HeaderHandler) error {
				return errors.New("error")
			},
			VerifySignatureCalled: func(header data.HeaderHandler) error {
				return errors.New("error")
			},
		}

		container.SetHeaderSigVerifier(headerSigVerifier)
		sr := initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})

		proof := &block.HeaderProof{}
		sr.ReceivedProof(proof)
	})
	t.Run("should return false when consensus data is not set", func(t *testing.T) {
		t.Parallel()

		container := consensusMocks.InitConsensusCore()
		sr := initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})
		sr.SetData(nil)

		proof := &block.HeaderProof{}
		sr.ReceivedProof(proof)
	})
	t.Run("should return false when sender is not in consensus group", func(t *testing.T) {
		t.Parallel()

		container := consensusMocks.InitConsensusCore()
		sr := initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})
		proof := &block.HeaderProof{}
		sr.ReceivedProof(proof)
	})
	t.Run("should return false when sender is self", func(t *testing.T) {
		t.Parallel()

		container := consensusMocks.InitConsensusCore()
		sr := initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})
		sr.SetSelfPubKey("A")

		proof := &block.HeaderProof{}
		sr.ReceivedProof(proof)
	})
	t.Run("should return false when different data is received", func(t *testing.T) {
		t.Parallel()

		container := consensusMocks.InitConsensusCore()
		sr := initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})
		sr.SetData([]byte("Y"))

		proof := &block.HeaderProof{}
		sr.ReceivedProof(proof)
	})
	t.Run("should return true when final info already received", func(t *testing.T) {
		t.Parallel()

		container := consensusMocks.InitConsensusCore()
		enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return flag == common.EquivalentMessagesFlag
			},
		}
		container.SetEnableEpochsHandler(enableEpochsHandler)

		container.SetEquivalentProofsPool(&dataRetriever.ProofsPoolMock{
			HasProofCalled: func(shardID uint32, headerHash []byte) bool {
				return true
			},
		})

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
			&statusHandler.AppStatusHandlerStub{},
		)
		sr.SetHeader(&block.HeaderV2{
			Header: createDefaultHeader(),
		})

		srEndRound, _ := v2.NewSubroundEndRound(
			sr,
			v2.ProcessingThresholdPercent,
			&statusHandler.AppStatusHandlerStub{},
			&testscommon.SentSignatureTrackerStub{},
			&consensusMocks.SposWorkerMock{},
			&dataRetrieverMocks.ThrottlerStub{},
		)

		proof := &block.HeaderProof{}
		srEndRound.ReceivedProof(proof)
	})
}

func TestSubroundEndRound_IsOutOfTimeShouldReturnFalse(t *testing.T) {
	t.Parallel()

	sr := initSubroundEndRound(&statusHandler.AppStatusHandlerStub{})

	res := sr.IsOutOfTime()
	assert.False(t, res)
}

func TestSubroundEndRound_IsOutOfTimeShouldReturnTrue(t *testing.T) {
	t.Parallel()

	// update roundHandler's mock, so it will calculate for real the duration
	container := consensusMocks.InitConsensusCore()
	roundHandler := consensusMocks.RoundHandlerMock{RemainingTimeCalled: func(startTime time.Time, maxTime time.Duration) time.Duration {
		currentTime := time.Now()
		elapsedTime := currentTime.Sub(startTime)
		remainingTime := maxTime - elapsedTime

		return remainingTime
	}}
	container.SetRoundHandler(&roundHandler)
	sr := initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})

	sr.SetRoundTimeStamp(time.Now().AddDate(0, 0, -1))

	res := sr.IsOutOfTime()
	assert.True(t, res)
}

func TestVerifyNodesOnAggSigVerificationFail(t *testing.T) {
	t.Parallel()

	t.Run("fail to get signature share", func(t *testing.T) {
		t.Parallel()

		container := consensusMocks.InitConsensusCore()
		sr := initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})

		signingHandler := &consensusMocks.SigningHandlerStub{
			SignatureShareCalled: func(index uint16) ([]byte, error) {
				return nil, expectedErr
			},
		}

		container.SetSigningHandler(signingHandler)

		sr.SetHeader(&block.Header{})
		leader, err := sr.GetLeader()
		require.Nil(t, err)
		_ = sr.SetJobDone(leader, bls.SrSignature, true)

		_, err = sr.VerifyNodesOnAggSigFail(context.TODO())
		require.Equal(t, expectedErr, err)
	})

	t.Run("fail to verify signature share, job done will be set to false", func(t *testing.T) {
		t.Parallel()

		container := consensusMocks.InitConsensusCore()
		sr := initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})

		signingHandler := &consensusMocks.SigningHandlerStub{
			SignatureShareCalled: func(index uint16) ([]byte, error) {
				return nil, nil
			},
			VerifySignatureShareCalled: func(index uint16, sig, msg []byte, epoch uint32) error {
				return expectedErr
			},
		}

		sr.SetHeader(&block.Header{})
		leader, err := sr.GetLeader()
		require.Nil(t, err)
		_ = sr.SetJobDone(leader, bls.SrSignature, true)
		container.SetSigningHandler(signingHandler)
		_, err = sr.VerifyNodesOnAggSigFail(context.TODO())
		require.Nil(t, err)

		isJobDone, err := sr.JobDone(leader, bls.SrSignature)
		require.Nil(t, err)
		require.False(t, isJobDone)
	})

	t.Run("fail to verify signature share, an element will return an error on SignatureShare, should not panic", func(t *testing.T) {
		t.Parallel()
		container := consensusMocks.InitConsensusCore()
		sr := initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})
		signingHandler := &consensusMocks.SigningHandlerStub{
			SignatureShareCalled: func(index uint16) ([]byte, error) {
				if index < 8 {
					return nil, nil
				}
				return nil, expectedErr
			},
			VerifySignatureShareCalled: func(index uint16, sig, msg []byte, epoch uint32) error {
				time.Sleep(100 * time.Millisecond)
				return expectedErr
			},
			VerifyCalled: func(msg, bitmap []byte, epoch uint32) error {
				return nil
			},
		}
		container.SetSigningHandler(signingHandler)

		sr.SetHeader(&block.Header{})
		_ = sr.SetJobDone(sr.ConsensusGroup()[0], bls.SrSignature, true)
		_ = sr.SetJobDone(sr.ConsensusGroup()[1], bls.SrSignature, true)
		_ = sr.SetJobDone(sr.ConsensusGroup()[2], bls.SrSignature, true)
		_ = sr.SetJobDone(sr.ConsensusGroup()[3], bls.SrSignature, true)
		_ = sr.SetJobDone(sr.ConsensusGroup()[4], bls.SrSignature, true)
		_ = sr.SetJobDone(sr.ConsensusGroup()[5], bls.SrSignature, true)
		_ = sr.SetJobDone(sr.ConsensusGroup()[6], bls.SrSignature, true)
		_ = sr.SetJobDone(sr.ConsensusGroup()[7], bls.SrSignature, true)
		_ = sr.SetJobDone(sr.ConsensusGroup()[8], bls.SrSignature, true)
		go func() {
			defer func() {
				if r := recover(); r != nil {
					t.Error("Should not panic")
				}
			}()
			invalidSigners, err := sr.VerifyNodesOnAggSigFail(context.TODO())
			time.Sleep(200 * time.Millisecond)
			require.Equal(t, err, expectedErr)
			require.Nil(t, invalidSigners)
		}()
		time.Sleep(time.Second)

	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		container := consensusMocks.InitConsensusCore()
		sr := initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})
		signingHandler := &consensusMocks.SigningHandlerStub{
			SignatureShareCalled: func(index uint16) ([]byte, error) {
				return nil, nil
			},
			VerifySignatureShareCalled: func(index uint16, sig, msg []byte, epoch uint32) error {
				return nil
			},
			VerifyCalled: func(msg, bitmap []byte, epoch uint32) error {
				return nil
			},
		}
		container.SetSigningHandler(signingHandler)

		sr.SetHeader(&block.Header{})
		_ = sr.SetJobDone(sr.ConsensusGroup()[0], bls.SrSignature, true)
		_ = sr.SetJobDone(sr.ConsensusGroup()[1], bls.SrSignature, true)
		invalidSigners, err := sr.VerifyNodesOnAggSigFail(context.TODO())
		require.Nil(t, err)
		require.NotNil(t, invalidSigners)
	})
}

func TestComputeAddSigOnValidNodes(t *testing.T) {
	t.Parallel()

	t.Run("invalid number of valid sig shares", func(t *testing.T) {
		t.Parallel()

		container := consensusMocks.InitConsensusCore()
		sr := initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})
		sr.SetHeader(&block.Header{})
		sr.SetThreshold(bls.SrEndRound, 2)

		_, _, err := sr.ComputeAggSigOnValidNodes()
		require.True(t, errors.Is(err, spos.ErrInvalidNumSigShares))
	})

	t.Run("fail to created aggregated sig", func(t *testing.T) {
		t.Parallel()

		container := consensusMocks.InitConsensusCore()
		sr := initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})

		signingHandler := &consensusMocks.SigningHandlerStub{
			AggregateSigsCalled: func(bitmap []byte, epoch uint32) ([]byte, error) {
				return nil, expectedErr
			},
		}
		container.SetSigningHandler(signingHandler)

		sr.SetHeader(&block.Header{})
		for _, participant := range sr.ConsensusGroup() {
			_ = sr.SetJobDone(participant, bls.SrSignature, true)
		}

		_, _, err := sr.ComputeAggSigOnValidNodes()
		require.Equal(t, expectedErr, err)
	})

	t.Run("fail to set aggregated sig", func(t *testing.T) {
		t.Parallel()

		container := consensusMocks.InitConsensusCore()
		sr := initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})

		signingHandler := &consensusMocks.SigningHandlerStub{
			SetAggregatedSigCalled: func(_ []byte) error {
				return expectedErr
			},
		}
		container.SetSigningHandler(signingHandler)
		sr.SetHeader(&block.Header{})
		for _, participant := range sr.ConsensusGroup() {
			_ = sr.SetJobDone(participant, bls.SrSignature, true)
		}

		_, _, err := sr.ComputeAggSigOnValidNodes()
		require.Equal(t, expectedErr, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		container := consensusMocks.InitConsensusCore()
		sr := initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})
		sr.SetHeader(&block.Header{})
		for _, participant := range sr.ConsensusGroup() {
			_ = sr.SetJobDone(participant, bls.SrSignature, true)
		}

		bitmap, sig, err := sr.ComputeAggSigOnValidNodes()
		require.NotNil(t, bitmap)
		require.NotNil(t, sig)
		require.Nil(t, err)
	})
}

func TestSubroundEndRound_DoEndRoundJobByNode(t *testing.T) {
	t.Parallel()

	t.Run("equivalent messages flag enabled and message already received", func(t *testing.T) {
		t.Parallel()

		container := consensusMocks.InitConsensusCore()
		enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return flag == common.EquivalentMessagesFlag
			},
		}
		container.SetEnableEpochsHandler(enableEpochsHandler)

		wasHasEquivalentProofCalled := false
		container.SetEquivalentProofsPool(&dataRetriever.ProofsPoolMock{
			HasProofCalled: func(shardID uint32, headerHash []byte) bool {
				wasHasEquivalentProofCalled = true
				return true
			},
		})

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
			&statusHandler.AppStatusHandlerStub{},
		)
		sr.SetHeader(&block.HeaderV2{
			Header: createDefaultHeader(),
		})

		srEndRound, _ := v2.NewSubroundEndRound(
			sr,
			v2.ProcessingThresholdPercent,
			&statusHandler.AppStatusHandlerStub{},
			&testscommon.SentSignatureTrackerStub{},
			&consensusMocks.SposWorkerMock{},
			&dataRetrieverMocks.ThrottlerStub{},
		)

		srEndRound.SetThreshold(bls.SrSignature, 2)

		for _, participant := range srEndRound.ConsensusGroup() {
			_ = srEndRound.SetJobDone(participant, bls.SrSignature, true)
		}

		r := srEndRound.DoEndRoundJobByNode()
		require.True(t, r)
		require.True(t, wasHasEquivalentProofCalled)
	})

	t.Run("should work without equivalent messages flag active", func(t *testing.T) {
		t.Parallel()

		container := consensusMocks.InitConsensusCore()
		sr := initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})

		verifySigShareNumCalls := 0
		mutex := &sync.Mutex{}
		verifyFirstCall := true
		signingHandler := &consensusMocks.SigningHandlerStub{
			SignatureShareCalled: func(index uint16) ([]byte, error) {
				return nil, nil
			},
			VerifySignatureShareCalled: func(index uint16, sig, msg []byte, epoch uint32) error {
				mutex.Lock()
				defer mutex.Unlock()
				if verifySigShareNumCalls == 0 {
					verifySigShareNumCalls++
					return expectedErr
				}

				verifySigShareNumCalls++
				return nil
			},
			VerifyCalled: func(msg, bitmap []byte, epoch uint32) error {
				mutex.Lock()
				defer mutex.Unlock()
				if verifyFirstCall {
					verifyFirstCall = false
					return expectedErr
				}

				return nil
			},
		}

		container.SetSigningHandler(signingHandler)

		sr.SetThreshold(bls.SrEndRound, 2)

		for _, participant := range sr.ConsensusGroup() {
			_ = sr.SetJobDone(participant, bls.SrSignature, true)
		}

		sr.SetHeader(&block.Header{})

		r := sr.DoEndRoundJobByNode()
		require.True(t, r)

		assert.False(t, verifyFirstCall)
		assert.Equal(t, 9, verifySigShareNumCalls)
	})
	t.Run("should work with equivalent messages flag active", func(t *testing.T) {
		t.Parallel()

		providedPrevSig := []byte("prev sig")
		providedPrevBitmap := []byte{1, 1, 1, 1, 1, 1, 1, 1, 1}
		container := consensusMocks.InitConsensusCore()
		container.SetBlockchain(&testscommon.ChainHandlerStub{
			GetGenesisHeaderCalled: func() data.HeaderHandler {
				return &block.HeaderV2{}
			},
		})
		enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return flag == common.EquivalentMessagesFlag
			},
		}
		container.SetEnableEpochsHandler(enableEpochsHandler)

		wasSetCurrentHeaderProofCalled := false
		container.SetEquivalentProofsPool(&dataRetriever.ProofsPoolMock{
			AddProofCalled: func(headerProof data.HeaderProofHandler) bool {
				wasSetCurrentHeaderProofCalled = true
				require.NotEqual(t, providedPrevSig, headerProof.GetAggregatedSignature())
				require.NotEqual(t, providedPrevBitmap, headerProof.GetPubKeysBitmap())
				return true
			},
		})

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
			&statusHandler.AppStatusHandlerStub{},
		)

		srEndRound, _ := v2.NewSubroundEndRound(
			sr,
			v2.ProcessingThresholdPercent,
			&statusHandler.AppStatusHandlerStub{},
			&testscommon.SentSignatureTrackerStub{},
			&consensusMocks.SposWorkerMock{},
			&dataRetrieverMocks.ThrottlerStub{},
		)

		srEndRound.SetThreshold(bls.SrEndRound, 2)

		for _, participant := range srEndRound.ConsensusGroup() {
			_ = srEndRound.SetJobDone(participant, bls.SrSignature, true)
		}

		srEndRound.SetHeader(&block.HeaderV2{
			Header:                   createDefaultHeader(),
			ScheduledRootHash:        []byte("sch root hash"),
			ScheduledAccumulatedFees: big.NewInt(0),
			ScheduledDeveloperFees:   big.NewInt(0),
			PreviousHeaderProof:      nil,
		})

		r := srEndRound.DoEndRoundJobByNode()
		require.True(t, r)
		require.True(t, wasSetCurrentHeaderProofCalled)
	})
}

func TestSubroundEndRound_ReceivedInvalidSignersInfo(t *testing.T) {
	t.Parallel()

	t.Run("consensus data is not set", func(t *testing.T) {
		t.Parallel()

		container := consensusMocks.InitConsensusCore()

		sr := initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})
		sr.ConsensusStateHandler.SetData(nil)

		cnsData := consensus.Message{
			BlockHeaderHash: []byte("X"),
			PubKey:          []byte("A"),
		}

		res := sr.ReceivedInvalidSignersInfo(&cnsData)
		assert.False(t, res)
	})
	t.Run("consensus header is not set", func(t *testing.T) {
		t.Parallel()

		container := consensusMocks.InitConsensusCore()

		sr := initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})
		sr.SetHeader(nil)

		cnsData := consensus.Message{
			BlockHeaderHash: []byte("X"),
			PubKey:          []byte("A"),
		}

		res := sr.ReceivedInvalidSignersInfo(&cnsData)
		assert.False(t, res)
	})
	t.Run("received message node is not leader in current round", func(t *testing.T) {
		t.Parallel()

		container := consensusMocks.InitConsensusCore()

		sr := initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})

		cnsData := consensus.Message{
			BlockHeaderHash: []byte("X"),
			PubKey:          []byte("other node"),
		}

		res := sr.ReceivedInvalidSignersInfo(&cnsData)
		assert.False(t, res)
	})

	t.Run("received message from self leader should return false", func(t *testing.T) {
		t.Parallel()

		container := consensusMocks.InitConsensusCore()

		sr := initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})
		sr.SetSelfPubKey("A")

		cnsData := consensus.Message{
			BlockHeaderHash: []byte("X"),
			PubKey:          []byte("A"),
		}

		res := sr.ReceivedInvalidSignersInfo(&cnsData)
		assert.False(t, res)
	})

	t.Run("received message from self multikey leader should return false", func(t *testing.T) {
		t.Parallel()

		container := consensusMocks.InitConsensusCore()
		keysHandler := &testscommon.KeysHandlerStub{
			IsKeyManagedByCurrentNodeCalled: func(pkBytes []byte) bool {
				return string(pkBytes) == "A"
			},
		}
		ch := make(chan bool, 1)
		consensusState := initializers.InitConsensusStateWithKeysHandler(keysHandler)
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
			&statusHandler.AppStatusHandlerStub{},
		)

		srEndRound, _ := v2.NewSubroundEndRound(
			sr,
			v2.ProcessingThresholdPercent,
			&statusHandler.AppStatusHandlerStub{},
			&testscommon.SentSignatureTrackerStub{},
			&consensusMocks.SposWorkerMock{},
			&dataRetrieverMocks.ThrottlerStub{},
		)

		srEndRound.SetSelfPubKey("A")

		cnsData := consensus.Message{
			BlockHeaderHash: []byte("X"),
			PubKey:          []byte("A"),
		}

		res := srEndRound.ReceivedInvalidSignersInfo(&cnsData)
		assert.False(t, res)
	})

	t.Run("received hash does not match the hash from current consensus state", func(t *testing.T) {
		t.Parallel()

		container := consensusMocks.InitConsensusCore()

		sr := initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})

		cnsData := consensus.Message{
			BlockHeaderHash: []byte("Y"),
			PubKey:          []byte("A"),
		}

		res := sr.ReceivedInvalidSignersInfo(&cnsData)
		assert.False(t, res)
	})
	t.Run("process received message verification failed, different round index", func(t *testing.T) {
		t.Parallel()

		container := consensusMocks.InitConsensusCore()

		sr := initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})

		cnsData := consensus.Message{
			BlockHeaderHash: []byte("X"),
			PubKey:          []byte("A"),
			RoundIndex:      1,
		}

		res := sr.ReceivedInvalidSignersInfo(&cnsData)
		assert.False(t, res)
	})
	t.Run("empty invalid signers", func(t *testing.T) {
		t.Parallel()

		container := consensusMocks.InitConsensusCore()

		sr := initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})
		cnsData := consensus.Message{
			BlockHeaderHash: []byte("X"),
			PubKey:          []byte("A"),
			InvalidSigners:  []byte{},
		}

		res := sr.ReceivedInvalidSignersInfo(&cnsData)
		assert.False(t, res)
	})
	t.Run("invalid signers data", func(t *testing.T) {
		t.Parallel()

		messageSigningHandler := &mock.MessageSigningHandlerStub{
			DeserializeCalled: func(messagesBytes []byte) ([]p2p.MessageP2P, error) {
				return nil, expectedErr
			},
		}

		container := consensusMocks.InitConsensusCore()
		container.SetMessageSigningHandler(messageSigningHandler)

		sr := initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})
		cnsData := consensus.Message{
			BlockHeaderHash: []byte("X"),
			PubKey:          []byte("A"),
			InvalidSigners:  []byte("invalid data"),
		}

		res := sr.ReceivedInvalidSignersInfo(&cnsData)
		assert.False(t, res)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		container := consensusMocks.InitConsensusCore()

		sr := initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})
		sr.SetHeader(&block.HeaderV2{
			Header: createDefaultHeader(),
		})
		cnsData := consensus.Message{
			BlockHeaderHash: []byte("X"),
			PubKey:          []byte("A"),
			InvalidSigners:  []byte("invalidSignersData"),
		}

		res := sr.ReceivedInvalidSignersInfo(&cnsData)
		assert.True(t, res)
	})
}

func TestVerifyInvalidSigners(t *testing.T) {
	t.Parallel()

	t.Run("failed to deserialize invalidSigners field, should error", func(t *testing.T) {
		t.Parallel()

		container := consensusMocks.InitConsensusCore()

		expectedErr := errors.New("expected err")
		messageSigningHandler := &mock.MessageSigningHandlerStub{
			DeserializeCalled: func(messagesBytes []byte) ([]p2p.MessageP2P, error) {
				return nil, expectedErr
			},
		}

		container.SetMessageSigningHandler(messageSigningHandler)

		sr := initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})

		err := sr.VerifyInvalidSigners([]byte{})
		require.Equal(t, expectedErr, err)
	})

	t.Run("failed to verify low level p2p message, should error", func(t *testing.T) {
		t.Parallel()

		container := consensusMocks.InitConsensusCore()

		invalidSigners := []p2p.MessageP2P{&factory.Message{
			FromField: []byte("from"),
		}}
		invalidSignersBytes, _ := container.Marshalizer().Marshal(invalidSigners)

		expectedErr := errors.New("expected err")
		messageSigningHandler := &mock.MessageSigningHandlerStub{
			DeserializeCalled: func(messagesBytes []byte) ([]p2p.MessageP2P, error) {
				require.Equal(t, invalidSignersBytes, messagesBytes)
				return invalidSigners, nil
			},
			VerifyCalled: func(message p2p.MessageP2P) error {
				return expectedErr
			},
		}

		container.SetMessageSigningHandler(messageSigningHandler)

		sr := initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})

		err := sr.VerifyInvalidSigners(invalidSignersBytes)
		require.Equal(t, expectedErr, err)
	})

	t.Run("failed to verify signature share", func(t *testing.T) {
		t.Parallel()

		container := consensusMocks.InitConsensusCore()

		pubKey := []byte("A") // it's in consensus

		consensusMsg := &consensus.Message{
			PubKey: pubKey,
		}
		consensusMsgBytes, _ := container.Marshalizer().Marshal(consensusMsg)

		invalidSigners := []p2p.MessageP2P{&factory.Message{
			FromField: []byte("from"),
			DataField: consensusMsgBytes,
		}}
		invalidSignersBytes, _ := container.Marshalizer().Marshal(invalidSigners)

		messageSigningHandler := &mock.MessageSigningHandlerStub{
			DeserializeCalled: func(messagesBytes []byte) ([]p2p.MessageP2P, error) {
				require.Equal(t, invalidSignersBytes, messagesBytes)
				return invalidSigners, nil
			},
		}

		wasCalled := false
		signingHandler := &consensusMocks.SigningHandlerStub{
			VerifySingleSignatureCalled: func(publicKeyBytes []byte, message []byte, signature []byte) error {
				wasCalled = true
				return errors.New("expected err")
			},
		}

		container.SetSigningHandler(signingHandler)
		container.SetMessageSigningHandler(messageSigningHandler)

		sr := initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})

		err := sr.VerifyInvalidSigners(invalidSignersBytes)
		require.Nil(t, err)
		require.True(t, wasCalled)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		container := consensusMocks.InitConsensusCore()

		pubKey := []byte("A") // it's in consensus

		consensusMsg := &consensus.Message{
			PubKey: pubKey,
		}
		consensusMsgBytes, _ := container.Marshalizer().Marshal(consensusMsg)

		invalidSigners := []p2p.MessageP2P{&factory.Message{
			FromField: []byte("from"),
			DataField: consensusMsgBytes,
		}}
		invalidSignersBytes, _ := container.Marshalizer().Marshal(invalidSigners)

		messageSigningHandler := &mock.MessageSignerMock{}
		container.SetMessageSigningHandler(messageSigningHandler)

		sr := initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})

		err := sr.VerifyInvalidSigners(invalidSignersBytes)
		require.Nil(t, err)
	})
}

func TestSubroundEndRound_CreateAndBroadcastInvalidSigners(t *testing.T) {
	t.Parallel()

	t.Run("redundancy node should not send while main is active", func(t *testing.T) {
		t.Parallel()

		expectedInvalidSigners := []byte("invalid signers")

		container := consensusMocks.InitConsensusCore()
		nodeRedundancy := &mock.NodeRedundancyHandlerStub{
			IsRedundancyNodeCalled: func() bool {
				return true
			},
			IsMainMachineActiveCalled: func() bool {
				return true
			},
		}
		container.SetNodeRedundancyHandler(nodeRedundancy)
		messenger := &consensusMocks.BroadcastMessengerMock{
			BroadcastConsensusMessageCalled: func(message *consensus.Message) error {
				assert.Fail(t, "should have not been called")
				return nil
			},
		}
		container.SetBroadcastMessenger(messenger)
		sr := initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})

		sr.CreateAndBroadcastInvalidSigners(expectedInvalidSigners)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		wg := &sync.WaitGroup{}
		wg.Add(1)

		expectedInvalidSigners := []byte("invalid signers")

		wasCalled := false
		container := consensusMocks.InitConsensusCore()
		messenger := &consensusMocks.BroadcastMessengerMock{
			BroadcastConsensusMessageCalled: func(message *consensus.Message) error {
				assert.Equal(t, expectedInvalidSigners, message.InvalidSigners)
				wasCalled = true
				wg.Done()
				return nil
			},
		}
		container.SetBroadcastMessenger(messenger)
		sr := initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})
		sr.SetSelfPubKey("A")

		sr.CreateAndBroadcastInvalidSigners(expectedInvalidSigners)

		wg.Wait()

		require.True(t, wasCalled)
	})
}

func TestGetFullMessagesForInvalidSigners(t *testing.T) {
	t.Parallel()

	t.Run("empty p2p messages slice if not in state", func(t *testing.T) {
		t.Parallel()

		container := consensusMocks.InitConsensusCore()

		messageSigningHandler := &mock.MessageSigningHandlerStub{
			SerializeCalled: func(messages []p2p.MessageP2P) ([]byte, error) {
				require.Equal(t, 0, len(messages))

				return []byte{}, nil
			},
		}

		container.SetMessageSigningHandler(messageSigningHandler)

		sr := initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})
		invalidSigners := []string{"B", "C"}

		invalidSignersBytes, err := sr.GetFullMessagesForInvalidSigners(invalidSigners)
		require.Nil(t, err)
		require.Equal(t, []byte{}, invalidSignersBytes)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		container := consensusMocks.InitConsensusCore()

		expectedInvalidSigners := []byte("expectedInvalidSigners")

		messageSigningHandler := &mock.MessageSigningHandlerStub{
			SerializeCalled: func(messages []p2p.MessageP2P) ([]byte, error) {
				require.Equal(t, 2, len(messages))

				return expectedInvalidSigners, nil
			},
		}

		container.SetMessageSigningHandler(messageSigningHandler)

		sr := initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})
		sr.AddMessageWithSignature("B", &p2pmocks.P2PMessageMock{})
		sr.AddMessageWithSignature("C", &p2pmocks.P2PMessageMock{})

		invalidSigners := []string{"B", "C"}

		invalidSignersBytes, err := sr.GetFullMessagesForInvalidSigners(invalidSigners)
		require.Nil(t, err)
		require.Equal(t, expectedInvalidSigners, invalidSignersBytes)
	})
}

func TestSubroundEndRound_getMinConsensusGroupIndexOfManagedKeys(t *testing.T) {
	t.Parallel()

	container := consensusMocks.InitConsensusCore()
	keysHandler := &testscommon.KeysHandlerStub{}
	ch := make(chan bool, 1)
	consensusState := initializers.InitConsensusStateWithKeysHandler(keysHandler)
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
		&statusHandler.AppStatusHandlerStub{},
	)

	srEndRound, _ := v2.NewSubroundEndRound(
		sr,
		v2.ProcessingThresholdPercent,
		&statusHandler.AppStatusHandlerStub{},
		&testscommon.SentSignatureTrackerStub{},
		&consensusMocks.SposWorkerMock{},
		&dataRetrieverMocks.ThrottlerStub{},
	)

	t.Run("no managed keys from consensus group", func(t *testing.T) {
		keysHandler.IsKeyManagedByCurrentNodeCalled = func(pkBytes []byte) bool {
			return false
		}

		assert.Equal(t, 9, srEndRound.GetMinConsensusGroupIndexOfManagedKeys())
	})
	t.Run("first managed key in consensus group should return 0", func(t *testing.T) {
		keysHandler.IsKeyManagedByCurrentNodeCalled = func(pkBytes []byte) bool {
			return bytes.Equal([]byte("A"), pkBytes)
		}

		assert.Equal(t, 0, srEndRound.GetMinConsensusGroupIndexOfManagedKeys())
	})
	t.Run("third managed key in consensus group should return 2", func(t *testing.T) {
		keysHandler.IsKeyManagedByCurrentNodeCalled = func(pkBytes []byte) bool {
			return bytes.Equal([]byte("C"), pkBytes)
		}

		assert.Equal(t, 2, srEndRound.GetMinConsensusGroupIndexOfManagedKeys())
	})
	t.Run("last managed key in consensus group should return 8", func(t *testing.T) {
		keysHandler.IsKeyManagedByCurrentNodeCalled = func(pkBytes []byte) bool {
			return bytes.Equal([]byte("I"), pkBytes)
		}

		assert.Equal(t, 8, srEndRound.GetMinConsensusGroupIndexOfManagedKeys())
	})
}

func TestSubroundSignature_ReceivedSignature(t *testing.T) {
	t.Parallel()

	sr := initSubroundEndRound(&statusHandler.AppStatusHandlerStub{})
	signature := []byte("signature")
	cnsMsg := consensus.NewConsensusMessage(
		sr.GetData(),
		signature,
		nil,
		nil,
		[]byte(sr.ConsensusGroup()[1]),
		[]byte("sig"),
		int(bls.MtSignature),
		0,
		chainID,
		nil,
		nil,
		nil,
		currentPid,
		nil,
	)

	sr.SetHeader(&block.Header{})
	sr.SetData(nil)
	r := sr.ReceivedSignature(cnsMsg)
	assert.False(t, r)

	sr.SetData([]byte("Y"))
	r = sr.ReceivedSignature(cnsMsg)
	assert.False(t, r)

	sr.SetData([]byte("X"))
	r = sr.ReceivedSignature(cnsMsg)
	assert.False(t, r)
	leader, err := sr.GetLeader()
	assert.Nil(t, err)

	sr.SetSelfPubKey(leader)

	cnsMsg.PubKey = []byte("X")
	r = sr.ReceivedSignature(cnsMsg)
	assert.False(t, r)

	cnsMsg.PubKey = []byte(sr.ConsensusGroup()[1])
	maxCount := len(sr.ConsensusGroup()) * 2 / 3
	count := 0
	for i := 0; i < len(sr.ConsensusGroup()); i++ {
		if sr.ConsensusGroup()[i] != string(cnsMsg.PubKey) {
			_ = sr.SetJobDone(sr.ConsensusGroup()[i], bls.SrSignature, true)
			count++
			if count == maxCount {
				break
			}
		}
	}
	r = sr.ReceivedSignature(cnsMsg)
	assert.True(t, r)
}

func TestSubroundSignature_ReceivedSignatureStoreShareFailed(t *testing.T) {
	t.Parallel()

	errStore := errors.New("signature share store failed")
	storeSigShareCalled := false
	signingHandler := &consensusMocks.SigningHandlerStub{
		VerifySignatureShareCalled: func(index uint16, sig, msg []byte, epoch uint32) error {
			return nil
		},
		StoreSignatureShareCalled: func(index uint16, sig []byte) error {
			storeSigShareCalled = true
			return errStore
		},
	}

	container := consensusMocks.InitConsensusCore()
	container.SetSigningHandler(signingHandler)
	sr := initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})
	sr.SetHeader(&block.Header{})

	signature := []byte("signature")
	cnsMsg := consensus.NewConsensusMessage(
		sr.GetData(),
		signature,
		nil,
		nil,
		[]byte(sr.ConsensusGroup()[1]),
		[]byte("sig"),
		int(bls.MtSignature),
		0,
		chainID,
		nil,
		nil,
		nil,
		currentPid,
		nil,
	)

	sr.SetData(nil)
	r := sr.ReceivedSignature(cnsMsg)
	assert.False(t, r)

	sr.SetData([]byte("Y"))
	r = sr.ReceivedSignature(cnsMsg)
	assert.False(t, r)

	sr.SetData([]byte("X"))
	r = sr.ReceivedSignature(cnsMsg)
	assert.False(t, r)

	leader, err := sr.GetLeader()
	assert.Nil(t, err)
	sr.SetSelfPubKey(leader)

	cnsMsg.PubKey = []byte("X")
	r = sr.ReceivedSignature(cnsMsg)
	assert.False(t, r)

	cnsMsg.PubKey = []byte(sr.ConsensusGroup()[1])
	maxCount := len(sr.ConsensusGroup()) * 2 / 3
	count := 0
	for i := 0; i < len(sr.ConsensusGroup()); i++ {
		if sr.ConsensusGroup()[i] != string(cnsMsg.PubKey) {
			_ = sr.SetJobDone(sr.ConsensusGroup()[i], bls.SrSignature, true)
			count++
			if count == maxCount {
				break
			}
		}
	}
	r = sr.ReceivedSignature(cnsMsg)
	assert.False(t, r)
	assert.True(t, storeSigShareCalled)
}
