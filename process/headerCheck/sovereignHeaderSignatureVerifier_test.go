package headerCheck

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/block"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/stretchr/testify/require"

	consensusMock "github.com/multiversx/mx-chain-go/consensus/mock"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon/cryptoMocks"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
)

func TestNewSovereignHeaderSigVerifier(t *testing.T) {
	t.Parallel()

	t.Run("nil verifier, should return error", func(t *testing.T) {
		sovVerifier, err := NewSovereignHeaderSigVerifier(nil, &enableEpochsHandlerMock.EnableEpochsHandlerStub{})
		require.Equal(t, process.ErrNilSingleSigner, err)
		require.Nil(t, sovVerifier)
	})

	t.Run("should work", func(t *testing.T) {
		sovVerifier, err := NewSovereignHeaderSigVerifier(&mock.SignerMock{}, &enableEpochsHandlerMock.EnableEpochsHandlerStub{})
		require.Nil(t, err)
		require.False(t, check.IfNil(sovVerifier))
		require.Equal(t, "sovereignHeaderSigVerifier", sovVerifier.Identifier())
	})
}

func TestSovereignHeaderSigVerifier_VerifyAggregatedSignature(t *testing.T) {
	t.Parallel()

	outGoingOpHash := []byte("outGoingOpHash")
	outGoingAggregatedSig := []byte("aggregatedSig")
	sovHdr := &block.SovereignChainHeader{
		Header: &block.Header{
			Nonce: 4,
		},
		OutGoingMiniBlockHeaders: []*block.OutGoingMiniBlockHeader{
			{
				OutGoingOperationsHash:                outGoingOpHash,
				AggregatedSignatureOutGoingOperations: outGoingAggregatedSig,
			},
		},
	}

	verifyCalledCt := 0
	expectedPubKeys := [][]byte{[]byte("pk1"), []byte("pk1")}
	multiSigner := &cryptoMocks.MultisignerMock{
		VerifyAggregatedSigCalled: func(pubKeysSigners [][]byte, message []byte, aggSig []byte) error {
			require.Equal(t, expectedPubKeys, pubKeysSigners)
			require.Equal(t, outGoingOpHash, message)
			require.Equal(t, outGoingAggregatedSig, aggSig)

			verifyCalledCt++
			return nil
		},
	}
	sovVerifier, _ := NewSovereignHeaderSigVerifier(&mock.SignerMock{}, &enableEpochsHandlerMock.EnableEpochsHandlerStub{})

	t.Run("invalid header type, should return error", func(t *testing.T) {
		err := sovVerifier.VerifyAggregatedSignature(nil, sovHdr.Header, multiSigner, expectedPubKeys)
		require.ErrorIs(t, err, errors.ErrWrongTypeAssertion)
	})

	t.Run("no outgoing mini block header", func(t *testing.T) {
		sovHdrCopy := *sovHdr
		sovHdrCopy.OutGoingMiniBlockHeaders = nil
		err := sovVerifier.VerifyAggregatedSignature(nil, &sovHdrCopy, multiSigner, expectedPubKeys)
		require.Nil(t, err)
		require.Zero(t, verifyCalledCt)
	})

	t.Run("should create sig share", func(t *testing.T) {
		err := sovVerifier.VerifyAggregatedSignature(nil, sovHdr, multiSigner, expectedPubKeys)
		require.Nil(t, err)
		require.Equal(t, 1, verifyCalledCt)
	})
}

func TestSovereignHeaderSigVerifier_getAggregatedSignature(t *testing.T) {
	t.Parallel()

	isAndromedaActive := false
	enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
			return isAndromedaActive
		},
	}
	sovVerifier, _ := NewSovereignHeaderSigVerifier(&mock.SignerMock{}, enableEpochsHandler)

	outGoingOpHash := []byte("outGoingOpHash")
	outGoingAggregatedSig := []byte("aggregatedSig")
	outGoingMBHeader := &block.OutGoingMiniBlockHeader{
		Type:                                  block.OutGoingMbTx,
		OutGoingOperationsHash:                outGoingOpHash,
		AggregatedSignatureOutGoingOperations: outGoingAggregatedSig,
	}

	t.Run("andromeda not active", func(t *testing.T) {
		aggSig, err := sovVerifier.getAggregatedSignature(outGoingMBHeader, nil)
		require.Nil(t, err)
		require.Equal(t, outGoingAggregatedSig, aggSig)
	})

	isAndromedaActive = true

	t.Run("andromeda active, nil proof", func(t *testing.T) {
		aggSig, err := sovVerifier.getAggregatedSignature(outGoingMBHeader, nil)
		require.Nil(t, aggSig)
		require.Equal(t, process.ErrNilHeaderProof, err)
	})

	t.Run("andromeda active, extra sig data not found for specific outgoing mb header", func(t *testing.T) {
		proof := &block.HeaderProof{
			ExtraSignatures: map[string]*block.ExtraSignatureData{
				block.OutGoingMbChangeValidatorSet.String(): {
					AggregatedSignature: outGoingAggregatedSig,
				},
			},
		}
		aggSig, err := sovVerifier.getAggregatedSignature(outGoingMBHeader, proof)
		require.Nil(t, aggSig)
		require.ErrorIs(t, err, errNoExtraSignatureDataFoundInProof)
	})

	t.Run("should work with andromeda", func(t *testing.T) {
		aggregatedSigFromProof := []byte("aggregatedSigFromProof")
		proof := &block.HeaderProof{
			ExtraSignatures: map[string]*block.ExtraSignatureData{
				block.OutGoingMbTx.String(): {
					AggregatedSignature: aggregatedSigFromProof,
				},
			},
		}
		aggSig, err := sovVerifier.getAggregatedSignature(outGoingMBHeader, proof)
		require.Nil(t, err)
		require.Equal(t, aggregatedSigFromProof, aggSig)
	})

}

func TestSovereignHeaderSigVerifier_getLeaderSignedMessage(t *testing.T) {
	t.Parallel()

	isAndromedaActive := false
	enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
			return isAndromedaActive
		},
	}
	sovVerifier, _ := NewSovereignHeaderSigVerifier(&mock.SignerMock{}, enableEpochsHandler)

	outGoingOpHash := []byte("outGoingOpHash")
	outGoingAggregatedSig := []byte("aggregatedSig")
	outGoingMBHeader := &block.OutGoingMiniBlockHeader{
		Type:                                  block.OutGoingMbTx,
		OutGoingOperationsHash:                outGoingOpHash,
		AggregatedSignatureOutGoingOperations: outGoingAggregatedSig,
	}

	t.Run("before andromeda", func(t *testing.T) {
		signedMsg := sovVerifier.getLeaderSignedMessage(outGoingMBHeader)
		require.Equal(t, append(
			outGoingMBHeader.GetOutGoingOperationsHash(),
			outGoingMBHeader.GetAggregatedSignatureOutGoingOperations()...),
			signedMsg)
	})

	t.Run("after andromeda", func(t *testing.T) {
		isAndromedaActive = true
		signedMsg := sovVerifier.getLeaderSignedMessage(outGoingMBHeader)
		require.Equal(t, outGoingMBHeader.GetOutGoingOperationsHash(), signedMsg)
	})
}

func TestSovereignHeaderSigVerifier_VerifyLeaderSignature(t *testing.T) {
	t.Parallel()

	outGoingOpHash := []byte("outGoingOpHash")
	outGoingAggregatedSig := []byte("aggregatedSig")
	outGoingLeaderSig := []byte("leaderSig")
	sovHdr := &block.SovereignChainHeader{
		Header: &block.Header{
			Nonce: 4,
		},
		OutGoingMiniBlockHeaders: []*block.OutGoingMiniBlockHeader{
			{
				OutGoingOperationsHash:                outGoingOpHash,
				AggregatedSignatureOutGoingOperations: outGoingAggregatedSig,
				LeaderSignatureOutGoingOperations:     outGoingLeaderSig,
			},
		},
	}

	verifyCalledCt := 0
	expectedLeaderPubKey := &consensusMock.PublicKeyMock{}
	signingHandler := &mock.SignerMock{
		VerifyStub: func(public crypto.PublicKey, msg []byte, sig []byte) error {
			require.Equal(t, expectedLeaderPubKey, public)
			require.Equal(t, append(outGoingOpHash, outGoingAggregatedSig...), msg)
			require.Equal(t, outGoingLeaderSig, sig)

			verifyCalledCt++
			return nil
		},
	}
	sovVerifier, _ := NewSovereignHeaderSigVerifier(signingHandler, &enableEpochsHandlerMock.EnableEpochsHandlerStub{})

	t.Run("invalid header type, should return error", func(t *testing.T) {
		err := sovVerifier.VerifyLeaderSignature(sovHdr.Header, expectedLeaderPubKey)
		require.ErrorIs(t, err, errors.ErrWrongTypeAssertion)
	})

	t.Run("no outgoing mini block header", func(t *testing.T) {
		sovHdrCopy := *sovHdr
		sovHdrCopy.OutGoingMiniBlockHeaders = nil
		err := sovVerifier.VerifyLeaderSignature(&sovHdrCopy, expectedLeaderPubKey)
		require.Nil(t, err)
		require.Zero(t, verifyCalledCt)
	})

	t.Run("should create sig share", func(t *testing.T) {
		err := sovVerifier.VerifyLeaderSignature(sovHdr, expectedLeaderPubKey)
		require.Nil(t, err)
		require.Equal(t, 1, verifyCalledCt)
	})
}

func TestSovereignHeaderSigVerifier_RemoveLeaderSignature(t *testing.T) {
	t.Parallel()

	outGoingOpHash := []byte("outGoingOpHash")
	outGoingAggregatedSig := []byte("aggregatedSig")
	outGoingLeaderSig := []byte("leaderSig")
	sovHdr := &block.SovereignChainHeader{
		Header: &block.Header{
			Nonce: 4,
		},
		OutGoingMiniBlockHeaders: []*block.OutGoingMiniBlockHeader{
			{
				OutGoingOperationsHash:                outGoingOpHash,
				AggregatedSignatureOutGoingOperations: outGoingAggregatedSig,
				LeaderSignatureOutGoingOperations:     outGoingLeaderSig,
			},
		},
	}

	sovVerifier, _ := NewSovereignHeaderSigVerifier(&mock.SignerMock{}, &enableEpochsHandlerMock.EnableEpochsHandlerStub{})

	t.Run("invalid header type, should return error", func(t *testing.T) {
		err := sovVerifier.RemoveLeaderSignature(sovHdr.Header)
		require.ErrorIs(t, err, errors.ErrWrongTypeAssertion)
	})

	t.Run("no outgoing mini block header", func(t *testing.T) {
		sovHdrCopy := *sovHdr
		sovHdrCopy.OutGoingMiniBlockHeaders = nil
		err := sovVerifier.RemoveLeaderSignature(&sovHdrCopy)
		require.Nil(t, err)
	})

	t.Run("should create sig share", func(t *testing.T) {
		err := sovVerifier.RemoveLeaderSignature(sovHdr)
		require.Nil(t, err)
		require.Equal(t, &block.SovereignChainHeader{
			Header: &block.Header{
				Nonce: 4,
			},
			OutGoingMiniBlockHeaders: []*block.OutGoingMiniBlockHeader{
				{
					OutGoingOperationsHash:                outGoingOpHash,
					AggregatedSignatureOutGoingOperations: outGoingAggregatedSig,
					LeaderSignatureOutGoingOperations:     nil,
				},
			},
		}, sovHdr)
	})
}

func TestSovereignHeaderSigVerifier_RemoveAllSignatures(t *testing.T) {
	t.Parallel()

	outGoingOpHash := []byte("outGoingOpHash")
	outGoingAggregatedSig := []byte("aggregatedSig")
	outGoingLeaderSig := []byte("leaderSig")
	sovHdr := &block.SovereignChainHeader{
		Header: &block.Header{
			Nonce: 4,
		},
		OutGoingMiniBlockHeaders: []*block.OutGoingMiniBlockHeader{
			{
				OutGoingOperationsHash:                outGoingOpHash,
				AggregatedSignatureOutGoingOperations: outGoingAggregatedSig,
				LeaderSignatureOutGoingOperations:     outGoingLeaderSig,
			},
		},
	}

	sovVerifier, _ := NewSovereignHeaderSigVerifier(&mock.SignerMock{}, &enableEpochsHandlerMock.EnableEpochsHandlerStub{})

	t.Run("invalid header type, should return error", func(t *testing.T) {
		err := sovVerifier.RemoveAllSignatures(sovHdr.Header)
		require.ErrorIs(t, err, errors.ErrWrongTypeAssertion)
	})

	t.Run("no outgoing mini block header", func(t *testing.T) {
		sovHdrCopy := *sovHdr
		sovHdrCopy.OutGoingMiniBlockHeaders = nil
		err := sovVerifier.RemoveAllSignatures(&sovHdrCopy)
		require.Nil(t, err)
	})

	t.Run("should create sig share", func(t *testing.T) {
		err := sovVerifier.RemoveAllSignatures(sovHdr)
		require.Nil(t, err)
		require.Equal(t, &block.SovereignChainHeader{
			Header: &block.Header{
				Nonce: 4,
			},
			OutGoingMiniBlockHeaders: []*block.OutGoingMiniBlockHeader{
				{
					OutGoingOperationsHash:                outGoingOpHash,
					AggregatedSignatureOutGoingOperations: nil,
					LeaderSignatureOutGoingOperations:     nil,
				},
			},
		}, sovHdr)
	})
}
