package systemSmartContracts

import "github.com/nikolaydubina/fpdecimal"

// Depth represents the order book depth.
type Depth struct {
	Bids [][2]fpdecimal.Decimal `json:"bids"`
	Asks [][2]fpdecimal.Decimal `json:"asks"`
}

// NewDepth creates a new instance of Depth.
func NewDepth() *Depth {
	return &Depth{
		Bids: make([][2]fpdecimal.Decimal, 0),
		Asks: make([][2]fpdecimal.Decimal, 0),
	}
}