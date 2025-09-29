package clob

import (
	"encoding/json"
	"math/big"
)

// CLOB represents the Central Limit Order Book.
type CLOB struct {
	OrderBook *OrderBook
	lastPrice *big.Float
}

// NewCLOB creates a new instance of the Central Limit Order Book.
func NewCLOB() *CLOB {
	return &CLOB{
		OrderBook: NewOrderBook(),
		lastPrice: nil, // Initialize as nil to indicate no trades have occurred yet
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
	if c.OrderBook == nil {
		return nil, ErrNilOrderBook
	}
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
	quantity, price, stop *big.Float,
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

	done, err := c.OrderBook.Process(order)
	if err != nil {
		return nil, err
	}

	if len(done.Trades) > 0 {
		c.lastPrice = done.Trades[0].GetPrice()
	}

	return done, nil
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

// MatchOrders activates stop orders and matches any crossed limit orders.
func (c *CLOB) MatchOrders() ([]*Done, error) {
	var (
		dones []*Done
		err   error
	)

	if c.OrderBook.Stop.Len() > 0 {
		dones, err = c.activateStopOrders(dones)
		if err != nil {
			return nil, err
		}
	}

	if c.OrderBook.Bids.Len() > 0 && c.OrderBook.Asks.Len() > 0 {
		dones, err = c.matchLimitOrders(dones)
		if err != nil {
			return nil, err
		}
	}

	return dones, nil
}

func (c *CLOB) activateStopOrders(dones []*Done) ([]*Done, error) {
	if c.lastPrice == nil {
		return dones, nil // No trades yet, so no stop orders can be activated
	}

	var activated []*Order

	c.OrderBook.Stop.Iterate(func(order *Order) {
		// A buy stop order is triggered when the last price is at or above the stop price.
		// A sell stop order is triggered when the last price is at or below the stop price.
		if (order.GetSide() == SideBuy && c.lastPrice.Cmp(order.GetStop()) >= 0) ||
			(order.GetSide() == SideSell && c.lastPrice.Cmp(order.GetStop()) <= 0) {
			activated = append(activated, order)
		}
	})

	for _, stopOrder := range activated {
		// Remove the stop order from the stop book
		c.OrderBook.CancelOrder(stopOrder.GetID())

		// Create a new limit order from the stop order
		newOrder := NewLimitOrder(
			stopOrder.GetID(),
			stopOrder.GetSide(),
			stopOrder.GetQuantity(),
			stopOrder.GetPrice(),
			stopOrder.GetTIF(),
			stopOrder.GetOCO(),
		)
		newOrder.SetTimestamp(stopOrder.GetTimestamp()) // Preserve original timestamp for priority

		// Process the new order
		done, err := c.OrderBook.Process(newOrder)
		if err != nil {
			return nil, err
		}
		dones = append(dones, done)

		if len(done.Trades) > 0 {
			c.lastPrice = done.Trades[len(done.Trades)-1].GetPrice()
		}
	}

	return dones, nil
}

func (c *CLOB) matchLimitOrders(dones []*Done) ([]*Done, error) {
	for c.OrderBook.Bids.Len() > 0 && c.OrderBook.Asks.Len() > 0 && c.OrderBook.Bids.Best().GetPrice().Cmp(c.OrderBook.Asks.Best().GetPrice()) >= 0 {
		var taker *Order
		bestBid := c.OrderBook.Bids.Best()
		bestAsk := c.OrderBook.Asks.Best()

		// Determine which order is the taker (the one that arrived later)
		if bestBid.GetTimestamp() > bestAsk.GetTimestamp() {
			taker = bestBid
		} else {
			taker = bestAsk
		}

		// Remove the taker order from the book to process it against the other side
		c.OrderBook.CancelOrder(taker.GetID())

		// Re-process the taker order to match it
		done, err := c.OrderBook.Process(taker)
		if err != nil {
			return nil, err
		}
		dones = append(dones, done)

		if len(done.Trades) > 0 {
			c.lastPrice = done.Trades[len(done.Trades)-1].GetPrice()
		}
	}

	return dones, nil
}