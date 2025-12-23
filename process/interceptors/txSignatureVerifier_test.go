package interceptors

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/transaction"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/consensus/mock"
	"github.com/multiversx/mx-chain-go/process"
)

func TestNewTxSignatureVerifier(t *testing.T) {
	t.Parallel()

	t.Run("nil key generator, should error", func(t *testing.T) {
		tsv, err := NewTxSignatureVerifier(nil, &mock.SingleSignerMock{})
		require.Error(t, err, process.ErrNilKeyGen)
		require.True(t, tsv.IsInterfaceNil())
	})
	t.Run("nil single signer, should error", func(t *testing.T) {
		tsv, err := NewTxSignatureVerifier(&mock.KeyGenMock{}, nil)
		require.Error(t, err, process.ErrNilSingleSigner)
		require.True(t, tsv.IsInterfaceNil())
	})
	t.Run("should work", func(t *testing.T) {
		tsv, err := NewTxSignatureVerifier(&mock.KeyGenMock{}, &mock.SingleSignerMock{})
		require.NoError(t, err)
		require.False(t, tsv.IsInterfaceNil())
		require.Equal(t, "*interceptors.txSignatureVerifier", fmt.Sprintf("%T", tsv))
	})
}

func TestNewTxSignatureVerifier_VerifySignature(t *testing.T) {
	t.Parallel()

	keyGen := &mock.KeyGenMock{
		PublicKeyFromByteArrayMock: func(b []byte) (crypto.PublicKey, error) {
			return &mock.PublicKeyMock{}, nil
		},
	}
	signer := &mock.SingleSignerMock{
		VerifyStub: func(public crypto.PublicKey, msg []byte, sig []byte) error {
			return nil
		},
	}

	tsv, err := NewTxSignatureVerifier(keyGen, signer)
	require.NoError(t, err)
	require.False(t, tsv.IsInterfaceNil())

	err = tsv.VerifySignature(&transaction.Transaction{}, nil, []byte("signature"))
	require.Nil(t, err)
}
