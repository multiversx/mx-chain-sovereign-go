package clob

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDone(t *testing.T) {
	t.Run("test done fields", func(t *testing.T) {
		order := NewLimitOrder("1", SideBuy, big.NewFloat(10), big.NewFloat(100), GTC, "")
		trade := &Order{
			ID:       "trade1",
			Price:    big.NewFloat(100),
			Quantity: big.NewFloat(5),
		}
		done := &Done{
			Order:     order,
			Stored:    true,
			Processed: big.NewFloat(5),
			Trades:    []*Order{trade},
		}

		require.Equal(t, order, done.Order)
		require.True(t, done.Stored)
		require.Equal(t, big.NewFloat(5), done.Processed)
		require.Equal(t, []*Order{trade}, done.Trades)
	})
}