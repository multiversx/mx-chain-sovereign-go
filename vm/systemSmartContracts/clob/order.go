package clob

import (
	"math/big"
	"time"
)

// Order represents a single order in the order book.
type Order struct {
	ID        string     `json:"id"`
	Side      Side       `json:"side"`
	OrderType OrderType  `json:"orderType"`
	Quantity  *big.Float `json:"quantity"`
	Price     *big.Float `json:"price"`
	Stop      *big.Float `json:"stop,omitempty"`
	TIF       TIF        `json:"tif,omitempty"`
	OCO       string     `json:"oco,omitempty"`
	Timestamp int64      `json:"timestamp"`
}

// NewMarketOrder creates a new market order.
func NewMarketOrder(orderID string, side Side, quantity *big.Float) *Order {
	return &Order{
		ID:        orderID,
		Side:      side,
		OrderType: TypeMarket,
		Quantity:  quantity,
		Timestamp: time.Now().UnixNano(),
	}
}

// NewLimitOrder creates a new limit order.
func NewLimitOrder(orderID string, side Side, quantity, price *big.Float, tif TIF, oco string) *Order {
	return &Order{
		ID:        orderID,
		Side:      side,
		OrderType: TypeLimit,
		Quantity:  quantity,
		Price:     price,
		TIF:       tif,
		OCO:       oco,
		Timestamp: time.Now().UnixNano(),
	}
}

// NewStopLimitOrder creates a new stop-limit order.
func NewStopLimitOrder(orderID string, side Side, quantity, price, stop *big.Float, oco string) *Order {
	return &Order{
		ID:        orderID,
		Side:      side,
		OrderType: TypeStopLimit,
		Quantity:  quantity,
		Price:     price,
		Stop:      stop,
		OCO:       oco,
		Timestamp: time.Now().UnixNano(),
	}
}

// GetID returns the order ID.
func (o *Order) GetID() string {
	return o.ID
}

// GetSide returns the order side.
func (o *Order) GetSide() Side {
	return o.Side
}

// GetType returns the order type.
func (o *Order) GetType() OrderType {
	return o.OrderType
}

// GetQuantity returns the order quantity.
func (o *Order) GetQuantity() *big.Float {
	return o.Quantity
}

// GetPrice returns the order price.
func (o *Order) GetPrice() *big.Float {
	return o.Price
}

// GetStop returns the order stop price.
func (o *Order) GetStop() *big.Float {
	return o.Stop
}

// GetTIF returns the time in force.
func (o *Order) GetTIF() TIF {
	return o.TIF
}

// GetOCO returns the OCO order ID.
func (o *Order) GetOCO() string {
	return o.OCO
}

// GetTimestamp returns the order timestamp.
func (o *Order) GetTimestamp() int64 {
	return o.Timestamp
}

// SetTimestamp sets the order timestamp.
func (o *Order) SetTimestamp(timestamp int64) {
	o.Timestamp = timestamp
}

// IsStopOrder returns true if the order is a stop order.
func (o *Order) IsStopOrder() bool {
	return o.OrderType == TypeStopLimit
}

// IsFilled returns true if the order is completely filled.
func (o *Order) IsFilled() bool {
	return o.Quantity.Cmp(big.NewFloat(0)) == 0
}

// fill fills the order with the given quantity.
func (o *Order) fill(quantity *big.Float) {
	o.Quantity.Sub(o.Quantity, quantity)
}