package systemSmartContracts

import (
	"encoding/json"

	"github.com/nikolaydubina/fpdecimal"
)

// CLOB represents the Central Limit Order Book.
type CLOB struct {
	OrderBook *OrderBook
}

// NewCLOB creates a new instance of the Central Limit Order Book.
func NewCLOB() *CLOB {
	return &CLOB{
		OrderBook: NewOrderBook(),
	}
}

// LoadState loads the order book state from a byte array.
func (c *CLOB) LoadState(data []byte) error {
	if len(data) == 0 {
		c.OrderBook = NewOrderBook()
		return nil
	}

	var orders []*Order
	err := json.Unmarshal(data, &orders)
	if err != nil {
		return err
	}

	orderBook := NewOrderBook()
	for _, order := range orders {
		if order.IsStopOrder() {
			orderBook.Stop.Append(order)
			orderBook.orders[order.GetID()] = order
		} else {
			orderBook.appendLimitOrder(order)
		}
	}

	c.OrderBook = orderBook
	return nil
}

// SaveState saves the order book state to a byte array.
func (c *CLOB) SaveState() ([]byte, error) {
	orders := make([]*Order, 0, len(c.OrderBook.orders))
	for _, order := range c.OrderBook.orders {
		orders = append(orders, order)
	}
	return json.Marshal(orders)
}

// ProcessOrder processes a new order.
func (c *CLOB) ProcessOrder(
	orderID string,
	side Side,
	orderType OrderType,
	quantity, price, stop fpdecimal.Decimal,
	tif TIF,
	oco string,
) (*Done, error) {
	var order *Order
	switch orderType {
	case TypeMarket:
		order = NewMarketOrder(orderID, side, quantity)
	case TypeLimit:
		order = NewLimitOrder(orderID, side, quantity, price, tif, oco)
	case TypeStopLimit:
		order = NewStopLimitOrder(orderID, side, quantity, price, stop, oco)
	default:
		return nil, ErrInvalidOrderType
	}

	return c.OrderBook.Process(order)
}

// CancelOrder cancels an existing order.
func (c *CLOB) CancelOrder(orderID string) *Order {
	return c.OrderBook.CancelOrder(orderID)
}

// GetOrder retrieves an order by its ID.
func (c *CLOB) GetOrder(orderID string) *Order {
	return c.OrderBook.GetOrder(orderID)
}

// GetDepth retrieves the order book depth.
func (c *CLOB) GetDepth() *Depth {
	return c.OrderBook.Depth()
}