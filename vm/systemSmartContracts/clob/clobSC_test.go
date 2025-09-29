package clob

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClobSC_NewClobSC(t *testing.T) {
	c := NewCLOB()
	sc := NewClobSC(c)
	require.NotNil(t, sc)
	require.Equal(t, c, sc.clob)
}

func TestClobSC_MatchOrders(t *testing.T) {
	t.Run("no match", func(t *testing.T) {
		c := NewCLOB()
		sc := NewClobSC(c)
		_, _ = sc.ProcessOrder("1", SideBuy, TypeLimit, big.NewFloat(10), big.NewFloat(100), nil, GTC, "")
		_, _ = sc.ProcessOrder("2", SideSell, TypeLimit, big.NewFloat(5), big.NewFloat(101), nil, GTC, "")
		done, err := sc.MatchOrders()
		require.NoError(t, err)
		require.Nil(t, done)
	})

	t.Run("with match", func(t *testing.T) {
		c := NewCLOB()
		sc := NewClobSC(c)
		buyOrder := NewLimitOrder("1", SideBuy, big.NewFloat(10), big.NewFloat(101), GTC, "")
		sellOrder := NewLimitOrder("2", SideSell, big.NewFloat(5), big.NewFloat(100), GTC, "")
		c.OrderBook.Bids.Append(buyOrder)
		c.OrderBook.Asks.Append(sellOrder)
		c.OrderBook.Orders[buyOrder.GetID()] = buyOrder
		c.OrderBook.Orders[sellOrder.GetID()] = sellOrder

		done, err := sc.MatchOrders()
		require.NoError(t, err)
		require.NotNil(t, done)
	})
}

func TestClobSC_State(t *testing.T) {
	c := NewCLOB()
	sc := NewClobSC(c)
	_, _ = sc.ProcessOrder("1", SideBuy, TypeLimit, big.NewFloat(10), big.NewFloat(100), nil, GTC, "")

	savedState, err := sc.SaveState()
	require.NoError(t, err)
	require.NotEmpty(t, savedState)

	newState := NewCLOB()
	newSc := NewClobSC(newState)
	err = newSc.LoadState(savedState)
	require.NoError(t, err)
	require.Equal(t, 1, newSc.clob.OrderBook.Bids.Len())
}

func TestClobSC_Getters(t *testing.T) {
	c := NewCLOB()
	sc := NewClobSC(c)
	_, _ = sc.ProcessOrder("1", SideBuy, TypeLimit, big.NewFloat(10), big.NewFloat(100), nil, GTC, "")

	order, err := sc.GetOrder("1")
	require.NoError(t, err)
	require.NotNil(t, order)
	require.Equal(t, "1", order.GetID())

	depth := sc.GetDepth()
	require.NotNil(t, depth)
	require.Len(t, depth.Bids, 1)
}