package systemSmartContracts

import "github.com/nikolaydubina/fpdecimal"

// OrderBook represents the order book for a single market.
type OrderBook struct {
	Bids   *OrderSide
	Asks   *OrderSide
	Stop   *StopBook
	orders map[string]*Order
}

// NewOrderBook creates a new instance of the OrderBook.
func NewOrderBook() *OrderBook {
	return &OrderBook{
		Bids:   NewOrderSide(SideBuy),
		Asks:   NewOrderSide(SideSell),
		Stop:   NewStopBook(),
		orders: make(map[string]*Order),
	}
}

// Process processes a new order.
func (ob *OrderBook) Process(order *Order) (*Done, error) {
	if order.IsStopOrder() {
		ob.Stop.Append(order)
		ob.orders[order.GetID()] = order
		return &Done{OrderID: order.GetID()}, nil
	}

	return ob.processLimitOrder(order)
}

// CancelOrder cancels an existing order.
func (ob *OrderBook) CancelOrder(orderID string) *Order {
	order, ok := ob.orders[orderID]
	if !ok {
		return nil
	}

	delete(ob.orders, orderID)

	if order.IsStopOrder() {
		return ob.Stop.Remove(order)
	}

	if order.GetSide() == SideBuy {
		return ob.Bids.Remove(order)
	}
	return ob.Asks.Remove(order)
}

// GetOrder retrieves an order by its ID.
func (ob *OrderBook) GetOrder(orderID string) *Order {
	return ob.orders[orderID]
}

// Depth returns the order book depth.
func (ob *OrderBook) Depth() *Depth {
	depth := NewDepth()
	bids := ob.Bids.Depth()
	asks := ob.Asks.Depth()

	for _, level := range bids {
		depth.Bids = append(depth.Bids, [2]fpdecimal.Decimal{level.Price, level.Quantity})
	}
	for _, level := range asks {
		depth.Asks = append(depth.Asks, [2]fpdecimal.Decimal{level.Price, level.Quantity})
	}

	return depth
}

func (ob *OrderBook) processLimitOrder(order *Order) (*Done, error) {
	done := &Done{OrderID: order.GetID()}
	if order.GetSide() == SideBuy {
		ob.processOrder(done, ob.Asks, order, func(p fpdecimal.Decimal) bool { return order.GetPrice() >= p })
	} else {
		ob.processOrder(done, ob.Bids, order, func(p fpdecimal.Decimal) bool { return order.GetPrice() <= p })
	}

	if !order.IsFilled() {
		ob.appendLimitOrder(order)
	}

	return done, nil
}

func (ob *OrderBook) processOrder(
	done *Done,
	side *OrderSide,
	order *Order,
	matchable func(fpdecimal.Decimal) bool,
) {
	for side.Len() > 0 && !order.IsFilled() {
		best := side.Best()
		if !matchable(best.GetPrice()) {
			break
		}

		matched := ob.match(order, best)
		done.Done = append(done.Done, matched...)
	}
}

func (ob *OrderBook) match(a, b *Order) []*Order {
	var done []*Order
	qty := a.GetQuantity()
	if b.GetQuantity() < qty {
		qty = b.GetQuantity()
	}

	a.fill(qty)
	b.fill(qty)

	if a.IsFilled() {
		done = append(done, a)
		delete(ob.orders, a.GetID())
	}
	if b.IsFilled() {
		done = append(done, b)
		delete(ob.orders, b.GetID())
		if b.GetSide() == SideBuy {
			ob.Bids.Remove(b)
		} else {
			ob.Asks.Remove(b)
		}
	}

	return done
}

func (ob *OrderBook) appendLimitOrder(order *Order) {
	ob.orders[order.GetID()] = order
	if order.GetSide() == SideBuy {
		ob.Bids.Append(order)
	} else {
		ob.Asks.Append(order)
	}
}