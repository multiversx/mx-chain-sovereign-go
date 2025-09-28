package clob

import (
	"encoding/json"
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
			orderBook.Orders[order.GetID()] = order
		} else {
			orderBook.appendLimitOrder(order)
		}
	}

	c.OrderBook = orderBook
	return nil
}

// SaveState saves the order book state to a byte array.
func (c *CLOB) SaveState() ([]byte, error) {
	orders := make([]*Order, 0, len(c.OrderBook.Orders))
	for _, order := range c.OrderBook.Orders {
		orders = append(orders, order)
	}
	return json.Marshal(orders)
}

// ProcessOrder processes a new order.
func (c *CLOB) ProcessOrder(
	orderID string,
	side Side,
	orderType OrderType,
	quantity, price, stop Decimal,
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
func (c *CLOB) CancelOrder(orderID string) (*Order, error) {
	order := c.OrderBook.CancelOrder(orderID)
	if order == nil {
		return nil, ErrOrderNotFound
	}
	return order, nil
}

// GetOrder retrieves an order by its ID.
func (c *CLOB) GetOrder(orderID string) (*Order, error) {
	order := c.OrderBook.GetOrder(orderID)
	if order == nil {
		return nil, ErrOrderNotFound
	}
	return order, nil
}

// GetDepth retrieves the order book depth.
func (c *CLOB) GetDepth() *Depth {
	return c.OrderBook.Depth()
}

// MatchOrders is a placeholder for the matching engine.
// In a real implementation, this would be triggered by a cron job or some other mechanism.
func (c *CLOB) MatchOrders() ([]*Done, error) {
	var (
		dones     []*Done
		lastPrice Decimal
	)

	// In a real implementation, we would get the last trade price from a persistent store.
	// For now, we'll just use the best bid price.
	if c.OrderBook.Bids.Len() > 0 {
		lastPrice = c.OrderBook.Bids.Best().GetPrice()
	}

	if lastPrice.IsZero() {
		return dones, nil
	}

	c.OrderBook.Stop.Iterate(func(order *Order) {
		if order.GetSide() == SideBuy && order.GetStop().LTE(lastPrice) {
			done, err := c.OrderBook.Process(order)
			if err == nil {
				dones = append(dones, done)
			}
		}
		if order.GetSide() == SideSell && order.GetStop().GTE(lastPrice) {
			done, err := c.OrderBook.Process(order)
			if err == nil {
				dones = append(dones, done)
			}
		}
	})

	return dones, nil
}