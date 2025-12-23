package crypto

import (
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/ed25519"
	"github.com/multiversx/mx-chain-crypto-go/signing/ed25519/singlesig"
)

// NewTxSignKeyGenerator creates new tx sign key generator
func NewTxSignKeyGenerator() crypto.KeyGenerator {
	return signing.NewKeyGenerator(ed25519.NewEd25519())
}

// NewTxSingleSigner creates new tx single signer
func NewTxSingleSigner() crypto.SingleSigner {
	return &singlesig.Ed25519Signer{}
}
