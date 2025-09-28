package systemSmartContracts

import (
	"math/big"
	"time"
)

// OrderType represents the type of an order.
type OrderType string

const (
	TypeMarket    OrderType = "MARKET"
	TypeLimit     OrderType = "LIMIT"
	TypeStopLimit OrderType = "STOP_LIMIT"
)

// TIF represents the time in force of an order.
type TIF string

const (
	TIF_GTC TIF = "GTC" // Good Till Cancel
	TIF_IOC TIF = "IOC" // Immediate or Cancel
	TIF_FOK TIF = "FOK" // Fill or Kill
)

// Order represents a single order in the order book.
type Order struct {
	id        string
	side      Side
	orderType OrderType
	quantity  *big.Float
	price     *big.Float
	stop      *big.Float
	tif       TIF
	oco       string
	timestamp int64
}

// NewMarketOrder creates a new market order.
func NewMarketOrder(orderID string, side Side, quantity *big.Float) *Order {
	return &Order{
		id:        orderID,
		side:      side,
		orderType: TypeMarket,
		quantity:  quantity,
		timestamp: time.Now().UnixNano(),
	}
}

// NewLimitOrder creates a new limit order.
func NewLimitOrder(orderID string, side Side, quantity, price *big.Float, tif TIF, oco string) *Order {
	return &Order{
		id:        orderID,
		side:      side,
		orderType: TypeLimit,
		quantity:  quantity,
		price:     price,
		tif:       tif,
		oco:       oco,
		timestamp: time.Now().UnixNano(),
	}
}

// NewStopLimitOrder creates a new stop-limit order.
func NewStopLimitOrder(orderID string, side Side, quantity, price, stop *big.Float, oco string) *Order {
	return &Order{
		id:        orderID,
		side:      side,
		orderType: TypeStopLimit,
		quantity:  quantity,
		price:     price,
		stop:      stop,
		oco:       oco,
		timestamp: time.Now().UnixNano(),
	}
}

// GetID returns the order ID.
func (o *Order) GetID() string {
	return o.id
}

// GetSide returns the order side.
func (o *Order) GetSide() Side {
	return o.side
}

// GetType returns the order type.
func (o *Order) GetType() OrderType {
	return o.orderType
}

// GetQuantity returns the order quantity.
func (o *Order) GetQuantity() *big.Float {
	return o.quantity
}

// GetPrice returns the order price.
func (o *Order) GetPrice() *big.Float {
	return o.price
}

// GetStop returns the order stop price.
func (o *Order) GetStop() *big.Float {
	return o.stop
}

// GetTIF returns the time in force.
func (o *Order) GetTIF() TIF {
	return o.tif
}

// GetOCO returns the OCO order ID.
func (o *Order) GetOCO() string {
	return o.oco
}

// GetTimestamp returns the order timestamp.
func (o *Order) GetTimestamp() int64 {
	return o.timestamp
}

// IsStopOrder returns true if the order is a stop order.
func (o *Order) IsStopOrder() bool {
	return o.orderType == TypeStopLimit
}

// IsFilled returns true if the order is completely filled.
func (o *Order) IsFilled() bool {
	return o.quantity.Cmp(big.NewFloat(0)) == 0
}

// fill fills the order with the given quantity.
func (o *Order) fill(quantity *big.Float) {
	o.quantity.Sub(o.quantity, quantity)
}