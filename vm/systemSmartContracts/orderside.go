package systemSmartContracts

import (
	"sort"

	"github.com/nikolaydubina/fpdecimal"
)

// Side represents the side of an order (buy or sell).
type Side byte

const (
	SideBuy  Side = 'B'
	SideSell Side = 'S'
)

// OrderSide represents one side of the order book (bids or asks).
type OrderSide struct {
	side         Side
	prices       []string
	priceLevels  map[string]*OrderQueue
	numOrders    int
	totalNotional fpdecimal.Decimal
}

// NewOrderSide creates a new instance of the OrderSide.
func NewOrderSide(side Side) *OrderSide {
	return &OrderSide{
		side:        side,
		prices:      make([]string, 0),
		priceLevels: make(map[string]*OrderQueue),
	}
}

// Len returns the number of price levels.
func (os *OrderSide) Len() int {
	return len(os.prices)
}

// Best returns the best order (highest bid or lowest ask).
func (os *OrderSide) Best() *Order {
	if len(os.prices) == 0 {
		return nil
	}
	return os.priceLevels[os.prices[0]].Front()
}

// Append appends a new order to the side.
func (os *OrderSide) Append(order *Order) {
	priceStr := order.GetPrice().String()
	level, ok := os.priceLevels[priceStr]
	if !ok {
		level = NewOrderQueue(priceStr)
		os.priceLevels[priceStr] = level
		os.addPrice(priceStr)
	}
	level.PushBack(order)
	os.numOrders++
	os.totalNotional += order.GetQuantity() * order.GetPrice()
}

// Remove removes an order from the side.
func (os *OrderSide) Remove(order *Order) *Order {
	priceStr := order.GetPrice().String()
	level, ok := os.priceLevels[priceStr]
	if !ok {
		return nil
	}

	removed := level.Remove(order)
	if removed != nil {
		os.numOrders--
		os.totalNotional -= removed.GetQuantity() * removed.GetPrice()
		if level.Len() == 0 {
			delete(os.priceLevels, priceStr)
			os.removePrice(priceStr)
		}
	}
	return removed
}

// Depth returns the depth of the side.
func (os *OrderSide) Depth() []struct {
	Price    fpdecimal.Decimal
	Quantity fpdecimal.Decimal
} {
	levels := make([]struct {
		Price    fpdecimal.Decimal
		Quantity fpdecimal.Decimal
	}, 0, len(os.prices))

	for _, priceStr := range os.prices {
		level, ok := os.priceLevels[priceStr]
		if !ok {
			continue
		}

		price, _ := fpdecimal.Parse([]byte(priceStr))
		var quantity fpdecimal.Decimal
		for i := 0; i < level.Len(); i++ {
			quantity += level.orders.At(i).GetQuantity()
		}
		levels = append(levels, struct {
			Price    fpdecimal.Decimal
			Quantity fpdecimal.Decimal
		}{Price: price, Quantity: quantity})
	}
	return levels
}

func (os *OrderSide) addPrice(price string) {
	os.prices = append(os.prices, price)
	sort.Slice(os.prices, func(i, j int) bool {
		a, _ := fpdecimal.Parse([]byte(os.prices[i]))
		b, _ := fpdecimal.Parse([]byte(os.prices[j]))
		if os.side == SideBuy {
			return a > b
		}
		return a < b
	})
}

func (os *OrderSide) removePrice(price string) {
	for i, p := range os.prices {
		if p == price {
			os.prices = append(os.prices[:i], os.prices[i+1:]...)
			return
		}
	}
}