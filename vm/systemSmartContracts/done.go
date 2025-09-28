package systemSmartContracts

import "github.com/nikolaydubina/fpdecimal"

// Done represents the result of a processed order.
type Done struct {
	OrderID    string            `json:"order_id"`
	Done       []*Order          `json:"done"`
	Partial    *Order            `json:"partial"`
	PartialQty fpdecimal.Decimal `json:"partial_qty"`
	Err        error             `json:"err"`
}