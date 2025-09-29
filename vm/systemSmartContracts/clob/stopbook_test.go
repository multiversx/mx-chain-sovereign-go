package clob

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStopBook(t *testing.T) {
	t.Run("test remove not found", func(t *testing.T) {
		sb := NewStopBook()
		order1 := NewStopLimitOrder("1", SideBuy, big.NewFloat(10), big.NewFloat(100), big.NewFloat(99), "")
		order2 := NewStopLimitOrder("2", SideBuy, big.NewFloat(5), big.NewFloat(101), big.NewFloat(100), "")
		sb.Append(order1)

		removed := sb.Remove(order2)
		require.Nil(t, removed)
		require.Equal(t, 1, sb.Len())
	})

	t.Run("test remove found", func(t *testing.T) {
		sb := NewStopBook()
		order1 := NewStopLimitOrder("1", SideBuy, big.NewFloat(10), big.NewFloat(100), big.NewFloat(99), "")
		sb.Append(order1)

		removed := sb.Remove(order1)
		require.Equal(t, order1, removed)
		require.Equal(t, 0, sb.Len())
	})
}