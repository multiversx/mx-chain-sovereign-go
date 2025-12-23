package interceptors

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	crypto "github.com/multiversx/mx-chain-crypto-go"

	"github.com/multiversx/mx-chain-go/process"
)

type txSignatureVerifier struct {
	keyGenerator crypto.KeyGenerator
	singleSigner crypto.SingleSigner
}

// NewTxSignatureVerifier creates a new tx signature verifier
func NewTxSignatureVerifier(keyGen crypto.KeyGenerator, signer crypto.SingleSigner) (*txSignatureVerifier, error) {
	if check.IfNil(keyGen) {
		return nil, process.ErrNilKeyGen
	}
	if check.IfNil(signer) {
		return nil, process.ErrNilSingleSigner
	}

	return &txSignatureVerifier{
		keyGenerator: keyGen,
		singleSigner: signer,
	}, nil
}

// VerifySignature verifies the tx signature
func (tsv *txSignatureVerifier) VerifySignature(tx data.TransactionHandler, txMessageForSigVerification []byte, signature []byte) error {
	pubKey, err := tsv.keyGenerator.PublicKeyFromByteArray(tx.GetSndAddr())
	if err != nil {
		return err
	}

	return tsv.singleSigner.Verify(pubKey, txMessageForSigVerification, signature)
}

// IsInterfaceNil verifies that the underlying value is nil
func (tsv *txSignatureVerifier) IsInterfaceNil() bool {
	return tsv == nil
}
