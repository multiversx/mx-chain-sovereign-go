package clob

import "math/big"

// Depth represents the order book depth.
type Depth struct {
	Bids [][2]*big.Float `json:"bids"`
	Asks [][2]*big.Float `json:"asks"`
}

// NewDepth creates a new instance of Depth.
func NewDepth() *Depth {
	return &Depth{
		Bids: make([][2]*big.Float, 0),
		Asks: make([][2]*big.Float, 0),
	}
}