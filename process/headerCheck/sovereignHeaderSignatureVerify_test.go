package headerCheck

import (
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/require"

	errMx "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/testscommon/cryptoMocks"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
)

func TestNewSovereignHeaderSignatureVerifier(t *testing.T) {
	t.Parallel()

	t.Run("nil input, should return error", func(t *testing.T) {
		sovHeaderVerifier, err := NewSovereignHeaderSignatureVerifier(nil)
		require.Nil(t, sovHeaderVerifier)
		require.Equal(t, errMx.ErrNilHeaderSigVerifier, err)
	})
	t.Run("should work", func(t *testing.T) {
		args := createHeaderSigVerifierArgs()
		headerVerifier, _ := NewHeaderSigVerifier(args)
		sovHeaderVerifier, err := NewSovereignHeaderSignatureVerifier(headerVerifier)
		require.Nil(t, err)
		require.False(t, sovHeaderVerifier.IsInterfaceNil())
	})
}

func TestSovereignHeaderSigVerifier_verifyProofAggregatedSignature(t *testing.T) {
	t.Parallel()

	args := createHeaderSigVerifierArgs()
	headerVerifier, _ := NewHeaderSigVerifier(args)
	sovHeaderVerifier, _ := NewSovereignHeaderSignatureVerifier(headerVerifier)

	pubKeys := [][]byte{[]byte("pk")}
	proof := &block.HeaderProof{
		ProcessedHeaderHash: []byte("hash"),
		AggregatedSignature: []byte("aggSig"),
	}

	wasVerifyCalled := false
	multiSigner := &cryptoMocks.MultisignerMock{
		VerifyAggregatedSigCalled: func(pubKeysSigners [][]byte, message []byte, aggSig []byte) error {
			require.Equal(t, pubKeys, pubKeysSigners)
			require.Equal(t, proof.GetProcessedHeaderHash(), message)
			require.Equal(t, proof.GetAggregatedSignature(), aggSig)

			wasVerifyCalled = true
			return nil
		},
	}

	require.Nil(t, sovHeaderVerifier.verifyProofAggregatedSignature(multiSigner, pubKeys, proof))
	require.True(t, wasVerifyCalled)
}

func TestSovereignHeaderSigVerifier_getLeaderSignedHeader(t *testing.T) {
	t.Parallel()

	args := createHeaderSigVerifierArgs()

	isAndromedaActive := false
	args.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
			return isAndromedaActive
		},
	}
	headerVerifier, _ := NewHeaderSigVerifier(args)
	sovHeaderVerifier, _ := NewSovereignHeaderSignatureVerifier(headerVerifier)

	header := &block.SovereignChainHeader{
		Header: &block.Header{
			Nonce: 4,
		},
		AccumulatedFeesInEpoch: big.NewInt(123),
	}

	require.Equal(t, header, sovHeaderVerifier.getLeaderSignedHeader(header))

	isAndromedaActive = true
	require.Equal(t, &block.SovereignChainHeader{
		Header: &block.Header{
			Nonce: 4,
		},
	}, sovHeaderVerifier.getLeaderSignedHeader(header))
}
