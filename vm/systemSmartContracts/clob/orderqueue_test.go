package clob

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

const testPrice = "100.0"

func TestOrderQueue(t *testing.T) {
	t.Run("test queue operations", func(t *testing.T) {
		q := NewOrderQueue(testPrice)
		require.Equal(t, 0, q.Len())

		order1 := NewLimitOrder("1", SideBuy, big.NewFloat(10), big.NewFloat(100), GTC, "")
		order2 := NewLimitOrder("2", SideBuy, big.NewFloat(5), big.NewFloat(100), GTC, "")

		q.PushBack(order1)
		q.PushBack(order2)
		require.Equal(t, 2, q.Len())
		require.Equal(t, order1, q.Front())
		require.Equal(t, order2, q.Back())

		popped := q.PopFront()
		require.Equal(t, order1, popped)
		require.Equal(t, 1, q.Len())
		require.Equal(t, order2, q.Front())

		q.Remove(order2)
		require.Equal(t, 0, q.Len())
	})

	t.Run("test remove from middle", func(t *testing.T) {
		q := NewOrderQueue(testPrice)
		order1 := NewLimitOrder("1", SideBuy, big.NewFloat(10), big.NewFloat(100), GTC, "")
		order2 := NewLimitOrder("2", SideBuy, big.NewFloat(5), big.NewFloat(100), GTC, "")
		order3 := NewLimitOrder("3", SideBuy, big.NewFloat(2), big.NewFloat(100), GTC, "")

		q.PushBack(order1)
		q.PushBack(order2)
		q.PushBack(order3)

		removed := q.Remove(order2)
		require.Equal(t, order2, removed)
		require.Equal(t, 2, q.Len())
	})

	t.Run("test remove not found", func(t *testing.T) {
		q := NewOrderQueue(testPrice)
		order1 := NewLimitOrder("1", SideBuy, big.NewFloat(10), big.NewFloat(100), GTC, "")
		order2 := NewLimitOrder("2", SideBuy, big.NewFloat(5), big.NewFloat(100), GTC, "")

		q.PushBack(order1)
		removed := q.Remove(order2)
		require.Nil(t, removed)
		require.Equal(t, 1, q.Len())
	})

	t.Run("test empty queue", func(t *testing.T) {
		q := NewOrderQueue(testPrice)
		require.Nil(t, q.Front())
		require.Nil(t, q.PopFront())
		require.Nil(t, q.Back())
	})
}