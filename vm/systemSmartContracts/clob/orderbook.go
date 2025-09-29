package clob

import "math/big"

// OrderBook represents the order book for a single market.
type OrderBook struct {
	Bids   *OrderSide        `json:"-"`
	Asks   *OrderSide        `json:"-"`
	Stop   *StopBook         `json:"-"`
	Orders map[string]*Order `json:"orders"`
}

// NewOrderBook creates a new instance of the OrderBook.
func NewOrderBook() *OrderBook {
	return &OrderBook{
		Bids:   NewOrderSide(SideBuy),
		Asks:   NewOrderSide(SideSell),
		Stop:   NewStopBook(),
		Orders: make(map[string]*Order),
	}
}

// Process processes a new order.
func (ob *OrderBook) Process(order *Order) (*Done, error) {
	if order.IsStopOrder() {
		ob.Stop.Append(order)
		ob.Orders[order.GetID()] = order
		return &Done{Order: order, Stored: true, Processed: big.NewFloat(0)}, nil
	}

	done, err := ob.processOrder(order)
	if err != nil {
		return nil, err
	}

	if !order.IsFilled() {
		if order.GetType() != TypeMarket {
			ob.appendLimitOrder(order)
			done.Stored = true
		}
	} else {
		delete(ob.Orders, order.GetID())
		done.Stored = false
	}

	return done, nil
}

// CancelOrder cancels an existing order.
func (ob *OrderBook) CancelOrder(orderID string) *Order {
	order, ok := ob.Orders[orderID]
	if !ok {
		return nil
	}

	delete(ob.Orders, orderID)

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
	return ob.Orders[orderID]
}

// Depth returns the order book depth.
func (ob *OrderBook) Depth() *Depth {
	depth := NewDepth()
	bids := ob.Bids.Depth()
	asks := ob.Asks.Depth()

	for _, level := range bids {
		depth.Bids = append(depth.Bids, [2]*big.Float{level.Price, level.Quantity})
	}
	for _, level := range asks {
		depth.Asks = append(depth.Asks, [2]*big.Float{level.Price, level.Quantity})
	}

	return depth
}

func (ob *OrderBook) processOrder(order *Order) (*Done, error) {
	done := &Done{Order: order, Processed: big.NewFloat(0)}
	var err error

	if order.GetType() == TypeMarket {
		done, err = ob.processMarketOrder(done, order)
	} else {
		done, err = ob.processLimitOrder(done, order)
	}

	return done, err
}

func (ob *OrderBook) processLimitOrder(done *Done, order *Order) (*Done, error) {
	s := ob.oppositeSide(order.GetSide())
	quantity := order.GetQuantity()

	for s.Len() > 0 && quantity.Cmp(big.NewFloat(0)) > 0 {
		bestPrice := s.Best().GetPrice()
		if order.GetSide() == SideBuy && order.GetPrice().Cmp(bestPrice) < 0 {
			break
		}
		if order.GetSide() == SideSell && order.GetPrice().Cmp(bestPrice) > 0 {
			break
		}

		ob.match(order, s.Best(), done)
		quantity = order.GetQuantity()
	}

	return done, nil
}

func (ob *OrderBook) processMarketOrder(done *Done, order *Order) (*Done, error) {
	s := ob.oppositeSide(order.GetSide())
	quantity := order.GetQuantity()

	for s.Len() > 0 && quantity.Cmp(big.NewFloat(0)) > 0 {
		best := s.Best()
		ob.match(order, best, done)
		quantity = order.GetQuantity()
	}

	return done, nil
}

func (ob *OrderBook) match(taker, maker *Order, done *Done) {
	tradeQuantity := big.NewFloat(0).Set(taker.GetQuantity())
	if maker.GetQuantity().Cmp(tradeQuantity) < 0 {
		tradeQuantity.Set(maker.GetQuantity())
	}

	taker.fill(tradeQuantity)
	maker.fill(tradeQuantity)

	done.Trades = append(done.Trades, &Order{
		ID:       taker.GetID(),
		Price:    maker.GetPrice(),
		Quantity: tradeQuantity,
	}, &Order{
		ID:       maker.GetID(),
		Price:    maker.GetPrice(),
		Quantity: tradeQuantity,
	})

	done.Processed.Add(done.Processed, tradeQuantity)

	if maker.IsFilled() {
		ob.removeOrder(maker)
	}
}

func (ob *OrderBook) appendLimitOrder(order *Order) {
	ob.Orders[order.GetID()] = order
	if order.GetSide() == SideBuy {
		ob.Bids.Append(order)
	} else {
		ob.Asks.Append(order)
	}
}

func (ob *OrderBook) removeOrder(order *Order) {
	delete(ob.Orders, order.GetID())
	if order.GetSide() == SideBuy {
		ob.Bids.Remove(order)
	} else {
		ob.Asks.Remove(order)
	}
}

func (ob *OrderBook) oppositeSide(side Side) *OrderSide {
	if side == SideBuy {
		return ob.Asks
	}
	return ob.Bids
}