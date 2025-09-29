package clob

import "errors"

var (
	ErrInvalidOrderType = errors.New("invalid order type")
	ErrOrderNotFound    = errors.New("order not found")
	ErrNilOrderBook     = errors.New("order book is nil")
)