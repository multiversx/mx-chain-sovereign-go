package dto

import dtoCore "github.com/multiversx/mx-chain-core-go/data/sovereign/dto"

// OutGoingOpNonceChainHandler holds outgoing op nonce per chain
type OutGoingOpNonceChainHandler interface {
	GetNonce(chainID dtoCore.ChainID) (uint64, error)
	IncrementNonce(chainID dtoCore.ChainID, delta uint64)
}
