package systemSmartContracts

import "errors"

var (
	ErrInvalidOrderType = errors.New("invalid order type")
	ErrOrderNotFound    = errors.New("order not found")
)