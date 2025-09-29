package clob

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOrderSide(t *testing.T) {
	t.Run("test best", func(t *testing.T) {
		side := NewOrderSide(SideBuy)
		require.Nil(t, side.Best())

		order1 := NewLimitOrder("1", SideBuy, big.NewFloat(10), big.NewFloat(100), GTC, "")
		side.Append(order1)
		require.Equal(t, order1, side.Best())

		// Add a better price for a buy side (higher)
		order2 := NewLimitOrder("2", SideBuy, big.NewFloat(5), big.NewFloat(101), GTC, "")
		side.Append(order2)
		require.Equal(t, order2, side.Best())
	})

	t.Run("test remove", func(t *testing.T) {
		side := NewOrderSide(SideBuy)
		order1 := NewLimitOrder("1", SideBuy, big.NewFloat(10), big.NewFloat(100), GTC, "")
		order2 := NewLimitOrder("2", SideBuy, big.NewFloat(5), big.NewFloat(100), GTC, "")
		side.Append(order1)
		side.Append(order2)
		require.Equal(t, 1, side.Len())
		require.Equal(t, 2, side.priceLevels[order1.Price.String()].Len())

		// Remove one order, price level should remain
		removed := side.Remove(order1)
		require.Equal(t, order1, removed)
		require.Equal(t, 1, side.Len())
		require.Equal(t, 1, side.priceLevels[order1.Price.String()].Len())

		// Remove the last order, price level should be removed
		removed = side.Remove(order2)
		require.Equal(t, order2, removed)
		require.Equal(t, 0, side.Len())
		require.Nil(t, side.priceLevels[order1.Price.String()])
	})

	t.Run("test remove from non-existent price level", func(t *testing.T) {
		side := NewOrderSide(SideBuy)
		order1 := NewLimitOrder("1", SideBuy, big.NewFloat(10), big.NewFloat(100), GTC, "")
		removed := side.Remove(order1)
		require.Nil(t, removed)
	})

	t.Run("test depth with multiple orders at same price", func(t *testing.T) {
		side := NewOrderSide(SideBuy)
		side.Append(NewLimitOrder("1", SideBuy, big.NewFloat(10), big.NewFloat(100), GTC, ""))
		side.Append(NewLimitOrder("2", SideBuy, big.NewFloat(5), big.NewFloat(100), GTC, ""))

		depth := side.Depth()
		require.Len(t, depth, 1)
		require.Equal(t, "100", depth[0].Price.Text('f', -1))
		require.Equal(t, "15", depth[0].Quantity.Text('f', -1))
	})

	t.Run("test depth with no orders", func(t *testing.T) {
		side := NewOrderSide(SideBuy)
		depth := side.Depth()
		require.Empty(t, depth)
	})
}