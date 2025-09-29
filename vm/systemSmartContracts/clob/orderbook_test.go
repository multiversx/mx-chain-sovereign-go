package clob

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOrderBook_RemoveOrder(t *testing.T) {
	t.Run("remove sell order", func(t *testing.T) {
		ob := NewOrderBook()
		order := NewLimitOrder("1", SideSell, big.NewFloat(10), big.NewFloat(100), GTC, "")
		ob.appendLimitOrder(order)
		require.Equal(t, 1, ob.Asks.Len())

		ob.removeOrder(order)
		require.Equal(t, 0, ob.Asks.Len())
	})

	t.Run("remove buy order", func(t *testing.T) {
		ob := NewOrderBook()
		order := NewLimitOrder("1", SideBuy, big.NewFloat(10), big.NewFloat(100), GTC, "")
		ob.appendLimitOrder(order)
		require.Equal(t, 1, ob.Bids.Len())

		ob.removeOrder(order)
		require.Equal(t, 0, ob.Bids.Len())
	})
}

func TestOrderBook_Process(t *testing.T) {
	t.Run("process with no match", func(t *testing.T) {
		ob := NewOrderBook()
		order := NewMarketOrder("1", SideBuy, big.NewFloat(10))
		done, err := ob.Process(order)
		require.NoError(t, err)
		require.NotNil(t, done)
		require.False(t, done.Stored)
	})
}

func TestOrderBook_CancelOrder(t *testing.T) {
	t.Run("cancel non-existent order", func(t *testing.T) {
		ob := NewOrderBook()
		order := ob.CancelOrder("non-existent")
		require.Nil(t, order)
	})
}

func TestOrderBook_processLimitOrder(t *testing.T) {
	t.Run("process sell limit order no match", func(t *testing.T) {
		ob := NewOrderBook()
		buyOrder := NewLimitOrder("1", SideBuy, big.NewFloat(10), big.NewFloat(99), GTC, "")
		ob.appendLimitOrder(buyOrder)

		sellOrder := NewLimitOrder("2", SideSell, big.NewFloat(5), big.NewFloat(100), GTC, "")
		done, err := ob.processLimitOrder(&Done{Order: sellOrder, Processed: big.NewFloat(0)}, sellOrder)
		require.NoError(t, err)
		require.Empty(t, done.Trades)
	})
}