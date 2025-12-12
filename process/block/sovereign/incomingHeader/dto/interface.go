package dto

import dtoCore "github.com/multiversx/mx-chain-core-go/data/sovereign/dto"

// OutGoingOpNonceChainHandler holds outgoing op nonce per chain
type OutGoingOpNonceChainHandler interface {
	GetAndIncrementNonce(chainID dtoCore.ChainID) (uint64, error)
	GetNonce(chainID dtoCore.ChainID) uint64
	SetNonce(chainID dtoCore.ChainID, nonce uint64)
	IsInterfaceNil() bool
}
