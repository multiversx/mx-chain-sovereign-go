package dto

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
)

// TokenProperties defines token properties of a token that is going to be registered from sovereign chain to main chain
type TokenProperties struct {
	TokenIdentifier []byte
	TokenType       core.ESDTType
	Name            []byte
	Ticker          []byte
	NumDecimals     uint64
	EventData       *sovereign.EventData
}

// RegisteredBlsKey defines operation data for a newly registered validator in sovereign
type RegisteredBlsKey struct {
	ID    []byte
	Key   []byte
	Owner []byte
}
