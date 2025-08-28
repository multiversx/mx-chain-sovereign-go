package sovereign_test

import (
	"errors"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/assert"

	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls/sovereign"
	v1 "github.com/multiversx/mx-chain-go/consensus/spos/bls/v1"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	consensusMock "github.com/multiversx/mx-chain-go/testscommon/consensus"
	"github.com/multiversx/mx-chain-go/testscommon/consensus/initializers"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
)

func defaultSubroundForSRBlock(consensusState *spos.ConsensusState, ch chan bool,
	container *spos.ConsensusCore, appStatusHandler core.AppStatusHandler) (*spos.Subround, error) {
	return spos.NewSubround(
		bls.SrStartRound,
		bls.SrBlock,
		bls.SrSignature,
		int64(5*roundTimeDuration/100),
		int64(25*roundTimeDuration/100),
		"(BLOCK)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
		chainID,
		currentPid,
		appStatusHandler,
	)
}

func defaultSubroundBlockFromSubround(sr *spos.Subround) (bls.SubRoundBlockHandler, error) {
	srBlock, err := v1.NewSubroundBlock(
		sr,
		extend,
		processingThresholdPercent,
	)

	return srBlock, err
}

func initSubroundBlock(
	blockChain data.ChainHandler,
	container *spos.ConsensusCore,
	appStatusHandler core.AppStatusHandler,
) bls.SubRoundBlockHandler {
	if blockChain == nil {
		blockChain = &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return &block.Header{}
			},
			GetGenesisHeaderCalled: func() data.HeaderHandler {
				return &block.Header{
					Nonce:     uint64(0),
					Signature: []byte("genesis signature"),
					RandSeed:  []byte{0},
				}
			},
			GetGenesisHeaderHashCalled: func() []byte {
				return []byte("genesis header hash")
			},
		}
	}

	consensusState := initializers.InitConsensusState()
	ch := make(chan bool, 1)

	container.SetBlockchain(blockChain)

	sr, _ := defaultSubroundForSRBlock(consensusState, ch, container, appStatusHandler)
	srBlock, _ := defaultSubroundBlockFromSubround(sr)
	return srBlock
}

func TestNewSubroundBlockV2_ShouldErrNilSubround(t *testing.T) {
	t.Parallel()

	srV2, err := sovereign.NewSubroundBlockV2(nil)
	assert.Nil(t, srV2)
	assert.Equal(t, spos.ErrNilSubround, err)
}

func TestNewSubroundBlockV2_ShouldWork(t *testing.T) {
	t.Parallel()

	container := consensusMock.InitConsensusCore()
	sr := initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})

	srV2, err := sovereign.NewSubroundBlockV2(sr)
	assert.NotNil(t, srV2)
	assert.Nil(t, err)
}

func TestSubroundBlockV2_DoBlockJob(t *testing.T) {
	t.Parallel()

	t.Run("if self is not leader in current round, should return false", func(t *testing.T) {
		t.Parallel()

		container := consensusMock.InitConsensusCore()
		sr := initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
		srV2, _ := sovereign.NewSubroundBlockV2(sr)

		r := srV2.DoBlockJob()
		assert.False(t, r)
	})

	t.Run("if current round is less or equal than the round in last committed block, should return false", func(t *testing.T) {
		t.Parallel()

		container := consensusMock.InitConsensusCore()
		sr := initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
		srV2, _ := sovereign.NewSubroundBlockV2(sr)
		srV2.SetSelfPubKey(srV2.ConsensusGroup()[0])
		_ = srV2.SetJobDone(srV2.SelfPubKey(), bls.SrBlock, true)

		r := srV2.DoBlockJob()
		assert.False(t, r)
	})

	t.Run("if self job is done, should return false", func(t *testing.T) {
		t.Parallel()

		container := consensusMock.InitConsensusCore()
		sr := initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
		srV2, _ := sovereign.NewSubroundBlockV2(sr)
		srV2.SetSelfPubKey(srV2.ConsensusGroup()[0])
		_ = srV2.SetJobDone(srV2.SelfPubKey(), bls.SrBlock, true)
		container.SetRoundHandler(&mock.RoundHandlerMock{
			RoundIndex: 1,
		})
		srV2.SetStatus(bls.SrBlock, spos.SsFinished)

		r := srV2.DoBlockJob()
		assert.False(t, r)
	})

	t.Run("if subround is finished, should return false", func(t *testing.T) {
		t.Parallel()

		container := consensusMock.InitConsensusCore()
		sr := initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
		srV2, _ := sovereign.NewSubroundBlockV2(sr)
		srV2.SetSelfPubKey(srV2.ConsensusGroup()[0])
		container.SetRoundHandler(&mock.RoundHandlerMock{
			RoundIndex: 1,
		})
		srV2.SetStatus(bls.SrBlock, spos.SsFinished)
		bpm := &testscommon.BlockProcessorStub{}
		err := errors.New("error")
		bpm.CreateBlockCalled = func(header data.HeaderHandler, remainingTime func() bool) (data.HeaderHandler, data.BodyHandler, error) {
			return header, nil, err
		}
		container.SetBlockProcessor(bpm)

		r := srV2.DoBlockJob()
		assert.False(t, r)
	})

	t.Run("if create header fails, should return false", func(t *testing.T) {
		t.Parallel()

		container := consensusMock.InitConsensusCore()
		sr := initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
		srV2, _ := sovereign.NewSubroundBlockV2(sr)
		srV2.SetSelfPubKey(srV2.ConsensusGroup()[0])
		container.SetRoundHandler(&mock.RoundHandlerMock{
			RoundIndex: 1,
		})
		bpm := consensusMock.InitBlockProcessorMock(container.Marshalizer())
		err := errors.New("error")
		bpm.CreateNewHeaderCalled = func(round uint64, nonce uint64) (data.HeaderHandler, error) {
			return nil, err
		}
		container.SetBlockProcessor(bpm)

		r := srV2.DoBlockJob()
		assert.False(t, r)
	})

	t.Run("if create block fails, should return false", func(t *testing.T) {
		t.Parallel()

		container := consensusMock.InitConsensusCore()
		sr := initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
		srV2, _ := sovereign.NewSubroundBlockV2(sr)
		srV2.SetSelfPubKey(srV2.ConsensusGroup()[0])
		container.SetRoundHandler(&mock.RoundHandlerMock{
			RoundIndex: 1,
		})
		bpm := consensusMock.InitBlockProcessorMock(container.Marshalizer())
		err := errors.New("error")
		bpm.CreateBlockCalled = func(initialHdrData data.HeaderHandler, haveTime func() bool) (data.HeaderHandler, data.BodyHandler, error) {
			return nil, nil, err
		}
		container.SetBlockProcessor(bpm)

		r := srV2.DoBlockJob()
		assert.False(t, r)
	})

	t.Run("if send block fails, should return false", func(t *testing.T) {
		t.Parallel()

		container := consensusMock.InitConsensusCore()
		sr := initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
		srV2, _ := sovereign.NewSubroundBlockV2(sr)
		srV2.SetSelfPubKey(srV2.ConsensusGroup()[0])
		container.SetRoundHandler(&mock.RoundHandlerMock{
			RoundIndex: 1,
		})
		bpm := consensusMock.InitBlockProcessorMock(container.Marshalizer())
		err := errors.New("error")
		container.SetBlockProcessor(bpm)
		bm := &consensusMock.BroadcastMessengerMock{
			BroadcastConsensusMessageCalled: func(message *consensus.Message) error {
				return err
			},
		}
		container.SetBroadcastMessenger(bm)

		r := srV2.DoBlockJob()
		assert.False(t, r)
	})

	t.Run("if process received block fails, should return false", func(t *testing.T) {
		t.Parallel()

		container := consensusMock.InitConsensusCore()
		sr := initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
		srV2, _ := sovereign.NewSubroundBlockV2(sr)
		srV2.SetSelfPubKey(srV2.ConsensusGroup()[0])
		container.SetRoundHandler(&mock.RoundHandlerMock{
			RoundIndex: 1,
		})
		bpm := consensusMock.InitBlockProcessorMock(container.Marshalizer())
		err := errors.New("error")
		bpm.ProcessBlockCalled = func(header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) (data.HeaderHandler, data.BodyHandler, error) {
			return nil, nil, err
		}
		container.SetBlockProcessor(bpm)

		r := srV2.DoBlockJob()
		assert.False(t, r)
	})

	t.Run("if process received block succeeds, should return true", func(t *testing.T) {
		t.Parallel()

		container := consensusMock.InitConsensusCore()
		sr := initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
		srV2, _ := sovereign.NewSubroundBlockV2(sr)
		srV2.SetSelfPubKey(srV2.ConsensusGroup()[0])
		container.SetRoundHandler(&mock.RoundHandlerMock{
			RoundIndex: 1,
		})

		r := srV2.DoBlockJob()
		assert.True(t, r)
		assert.Equal(t, uint64(1), sr.GetHeader().GetNonce())
	})
}
