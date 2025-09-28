package systemSmartContracts

import "math/big"

// Done contains information about processed order
type Done struct {
	Order     *Order     `json:"order"`
	Trades    []*Order   `json:"trades"`
	Canceled  []string   `json:"canceled"`
	Activated []string   `json:"activated"`
	Left      *big.Float `json:"left"`
	Processed *big.Float `json:"processed"`
	Stored    bool       `json:"stored"`
}