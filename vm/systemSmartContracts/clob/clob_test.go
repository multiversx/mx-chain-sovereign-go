package clob

import (
	"encoding/json"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCLOB_LoadState(t *testing.T) {
	t.Run("load empty state", func(t *testing.T) {
		c := NewCLOB()
		err := c.LoadState([]byte{})
		require.NoError(t, err)
		require.NotNil(t, c.OrderBook)
	})

	t.Run("load invalid json", func(t *testing.T) {
		c := NewCLOB()
		err := c.LoadState([]byte("invalid json"))
		require.Error(t, err)
	})

	t.Run("load valid state", func(t *testing.T) {
		c := NewCLOB()
		orders := []*Order{
			NewLimitOrder("1", SideBuy, big.NewFloat(10), big.NewFloat(100), GTC, ""),
			NewStopLimitOrder("2", SideSell, big.NewFloat(5), big.NewFloat(105), big.NewFloat(102), ""),
		}
		data, err := json.Marshal(orders)
		require.NoError(t, err)

		err = c.LoadState(data)
		require.NoError(t, err)
		require.Len(t, c.OrderBook.Orders, 2)
		require.Equal(t, 1, c.OrderBook.Bids.Len())
		require.Equal(t, 1, c.OrderBook.Stop.Len())
	})
}

func TestCLOB_ProcessOrder(t *testing.T) {
	t.Run("process market order with match", func(t *testing.T) {
		c := NewCLOB()
		// Add a limit order to match against
		_, err := c.ProcessOrder("seller", SideSell, TypeLimit, big.NewFloat(15), big.NewFloat(100), nil, GTC, "")
		require.NoError(t, err)

		done, err := c.ProcessOrder("1", SideBuy, TypeMarket, big.NewFloat(10), nil, nil, "", "")
		require.NoError(t, err)
		require.NotNil(t, done)
		require.Equal(t, "1", done.Order.GetID())
		require.NotEmpty(t, done.Trades)
	})

	t.Run("process partially filled market order", func(t *testing.T) {
		c := NewCLOB()
		// Add a smaller limit order to match against
		_, err := c.ProcessOrder("seller", SideSell, TypeLimit, big.NewFloat(5), big.NewFloat(100), nil, GTC, "")
		require.NoError(t, err)

		done, err := c.ProcessOrder("1", SideBuy, TypeMarket, big.NewFloat(10), nil, nil, "", "")
		require.NoError(t, err)
		require.NotNil(t, done)
		require.False(t, done.Stored) // Market orders are not stored
	})

	t.Run("process stop-limit order", func(t *testing.T) {
		c := NewCLOB()
		done, err := c.ProcessOrder("2", SideSell, TypeStopLimit, big.NewFloat(5), big.NewFloat(105), big.NewFloat(102), "", "")
		require.NoError(t, err)
		require.NotNil(t, done)
		require.Equal(t, "2", done.Order.GetID())
		require.Equal(t, 1, c.OrderBook.Stop.Len())
	})

	t.Run("invalid order type", func(t *testing.T) {
		c := NewCLOB()
		_, err := c.ProcessOrder("3", SideBuy, "invalid", big.NewFloat(10), nil, nil, "", "")
		require.Equal(t, ErrInvalidOrderType, err)
	})
}

func TestCLOB_MatchOrders(t *testing.T) {
	t.Run("no match", func(t *testing.T) {
		c := NewCLOB()
		_, _ = c.ProcessOrder("1", SideBuy, TypeLimit, big.NewFloat(10), big.NewFloat(100), nil, GTC, "")
		_, _ = c.ProcessOrder("2", SideSell, TypeLimit, big.NewFloat(5), big.NewFloat(101), nil, GTC, "")
		dones, err := c.MatchOrders()
		require.NoError(t, err)
		require.Empty(t, dones)
	})

	t.Run("match limit orders via ProcessOrder", func(t *testing.T) {
		c := NewCLOB()
		// Maker order
		_, err := c.ProcessOrder("1", SideBuy, TypeLimit, big.NewFloat(10), big.NewFloat(101), nil, GTC, "")
		require.NoError(t, err)

		// Taker order that should match
		done, err := c.ProcessOrder("2", SideSell, TypeLimit, big.NewFloat(5), big.NewFloat(100), nil, GTC, "")
		require.NoError(t, err)
		require.NotNil(t, done)
		require.NotEmpty(t, done.Trades)
		require.False(t, done.Stored) // Taker is filled

		// Verify book state
		require.Equal(t, 1, c.OrderBook.Bids.Len())
		require.Equal(t, 0, c.OrderBook.Asks.Len())
		maker := c.OrderBook.GetOrder("1")
		require.NotNil(t, maker)
		require.Equal(t, 0, maker.GetQuantity().Cmp(big.NewFloat(5)))
	})

	t.Run("match limit orders via MatchOrders with crossed book", func(t *testing.T) {
		c := NewCLOB()
		// Create a crossed book via LoadState
		buyOrder := NewLimitOrder("1", SideBuy, big.NewFloat(10), big.NewFloat(101), GTC, "")
		buyOrder.Timestamp = time.Now().UnixNano()
		sellOrder := NewLimitOrder("2", SideSell, big.NewFloat(5), big.NewFloat(100), GTC, "")
		sellOrder.Timestamp = time.Now().UnixNano() - 1000 // ensure buy order is the taker

		orders := []*Order{buyOrder, sellOrder}
		data, err := json.Marshal(orders)
		require.NoError(t, err)
		err = c.LoadState(data)
		require.NoError(t, err)

		// Now the book is crossed. Call MatchOrders.
		dones, err := c.MatchOrders()
		require.NoError(t, err)
		require.NotEmpty(t, dones)
		require.Len(t, dones, 1)
		require.NotEmpty(t, dones[0].Trades)
	})

	t.Run("activate stop orders", func(t *testing.T) {
		c := NewCLOB()
		_, _ = c.ProcessOrder("1", SideBuy, TypeStopLimit, big.NewFloat(10), big.NewFloat(102), big.NewFloat(101), "", "")
		c.lastPrice = big.NewFloat(101) // Set last price to trigger stop order
		dones, err := c.MatchOrders()
		require.NoError(t, err)
		require.NotEmpty(t, dones)
	})

	t.Run("activate stop orders no last price", func(t *testing.T) {
		c := NewCLOB()
		_, _ = c.ProcessOrder("1", SideBuy, TypeStopLimit, big.NewFloat(10), big.NewFloat(102), big.NewFloat(101), "", "")
		dones, err := c.MatchOrders()
		require.NoError(t, err)
		require.Empty(t, dones)
	})
}