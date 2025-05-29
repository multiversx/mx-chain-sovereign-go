package sovereign_test

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls/sovereign"
	v1 "github.com/multiversx/mx-chain-go/consensus/spos/bls/v1"
	"github.com/multiversx/mx-chain-go/testscommon"
	consensusMocks "github.com/multiversx/mx-chain-go/testscommon/consensus"
	"github.com/multiversx/mx-chain-go/testscommon/consensus/initializers"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	"github.com/multiversx/mx-chain-go/testscommon/subRounds"
	"github.com/stretchr/testify/assert"
)

func initSubroundSignatureWithContainer(container spos.ConsensusCoreHandler, enableEpochHandler common.EnableEpochsHandler) sovereign.SubRoundSignatureHandler {
	consensusState := initializers.InitConsensusState()
	ch := make(chan bool, 1)

	sr, _ := spos.NewSubround(
		bls.SrBlock,
		bls.SrSignature,
		bls.SrEndRound,
		int64(70*roundTimeDuration/100),
		int64(85*roundTimeDuration/100),
		"(SIGNATURE)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
		enableEpochHandler,
	)

	srSignature, _ := v1.NewSubroundSignature(
		sr,
		extend,
		&statusHandler.AppStatusHandlerStub{},
		&subRounds.SubRoundSignatureExtraSignersHolderMock{},
		&testscommon.SentSignatureTrackerStub{},
	)

	return srSignature
}

func initSubroundSignature() sovereign.SubRoundSignatureHandler {
	container := consensusMocks.InitConsensusCore()
	return initSubroundSignatureWithContainer(container, &enableEpochsHandlerMock.EnableEpochsHandlerStub{})
}

func TestNewSubroundSignatureV2_ShouldErrNilSubround(t *testing.T) {
	t.Parallel()

	srV2, err := sovereign.NewSubroundSignatureV2(nil)

	assert.Nil(t, srV2)
	assert.Equal(t, spos.ErrNilSubround, err)
}

func TestNewSubroundSignatureV2_ShouldWork(t *testing.T) {
	t.Parallel()

	sr := initSubroundSignature()
	srV2, err := sovereign.NewSubroundSignatureV2(sr)

	assert.NotNil(t, srV2)
	assert.Nil(t, err)
}

func TestSubroundSignatureV2_GetMessageToSign(t *testing.T) {
	t.Parallel()

	t.Run("getMessageToSign should return nil when CalculateHash method fails", func(t *testing.T) {
		t.Parallel()

		sr := initSubroundSignature()
		srV2, _ := sovereign.NewSubroundSignatureV2(sr)

		srV2.SetHeader(nil)
		msg := srV2.GetMessageToSign()

		assert.Nil(t, msg)
	})

	t.Run("getMessageToSign should return the message that should be signed", func(t *testing.T) {
		t.Parallel()

		sr := initSubroundSignature()
		srV2, _ := sovereign.NewSubroundSignatureV2(sr)

		srV2.SetHeader(&block.Header{Nonce: 1})
		expectedMsg, _ := core.CalculateHash(srV2.Marshalizer(), srV2.Hasher(), srV2.GetHeader())
		msg := srV2.GetMessageToSign()

		assert.Equal(t, expectedMsg, msg)
	})
}
