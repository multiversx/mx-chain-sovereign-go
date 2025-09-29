package clob

import (
	"encoding/json"
	"fmt"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-go/testscommon"
	storageCommon "github.com/multiversx/mx-chain-storage-go/common"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

func createExecutorTestContext(tb testing.TB) (*clobExecutor, *testscommon.AccountHandlerStub) {
	storage := testscommon.NewAccountHandlerStub()
	executor := NewClobExecutor()
	return executor, storage
}

func TestNewClobExecutor(t *testing.T) {
	executor := NewClobExecutor()
	require.NotNil(t, executor)
}

func TestClobExecutor_Execute(t *testing.T) {
	t.Run("invalid function name", func(t *testing.T) {
		executor, storage := createExecutorTestContext(t)
		input := &vmcommon.ContractCallInput{
			Function: "invalidFunction",
		}

		output, err := executor.Execute(input, storage)
		require.NoError(t, err)
		require.Equal(t, vmcommon.UserError, output.ReturnCode)
		require.Equal(t, "invalid function name: invalidFunction", string(output.ReturnMessage))
	})

	t.Run("load clob fails", func(t *testing.T) {
		executor, storage := createExecutorTestContext(t)
		expectedErr := fmt.Errorf("expected error")
		storage.RetrieveValueCalled = func(key []byte) ([]byte, uint32, error) {
			return nil, 0, expectedErr
		}

		output, err := executor.Execute(&vmcommon.ContractCallInput{}, storage)
		require.Nil(t, output)
		require.ErrorContains(t, err, "failed to retrieve clob data: expected error")
	})

	t.Run("load clob state fails", func(t *testing.T) {
		executor, storage := createExecutorTestContext(t)
		storage.RetrieveValueCalled = func(key []byte) ([]byte, uint32, error) {
			return []byte("invalid data"), 0, nil
		}

		output, err := executor.Execute(&vmcommon.ContractCallInput{}, storage)
		require.Nil(t, output)
		require.ErrorContains(t, err, "could not load clob state")
	})

	t.Run("save clob fails", func(t *testing.T) {
		executor, storage := createExecutorTestContext(t)
		expectedErr := fmt.Errorf("expected error")
		storage.SaveKeyValueCalled = func(key, value []byte) error {
			return expectedErr
		}

		input := &vmcommon.ContractCallInput{
			Function: ProcessOrderEndpoint,
			VMInput: vmcommon.VMInput{
				Arguments: [][]byte{
					[]byte("order1"),
					{byte(SideBuy)},
					[]byte(TypeLimit),
					[]byte("10"),
					[]byte("100"),
					[]byte("0"),
					[]byte(GTC),
					[]byte(""),
				},
			},
		}

		output, err := executor.Execute(input, storage)
		require.Nil(t, output)
		require.ErrorContains(t, err, "could not save key-value: expected error")
	})
}

func TestClobExecutor_ProcessOrder(t *testing.T) {
	t.Run("valid limit buy order", func(t *testing.T) {
		executor, storage := createExecutorTestContext(t)
		input := &vmcommon.ContractCallInput{
			Function: ProcessOrderEndpoint,
			VMInput: vmcommon.VMInput{
				Arguments: [][]byte{
					[]byte("order1"),
					{byte(SideBuy)},
					[]byte(TypeLimit),
					[]byte("10.0"),
					[]byte("100.0"),
					[]byte("0.0"),
					[]byte(GTC),
					[]byte(""),
				},
			},
		}

		output, err := executor.Execute(input, storage)
		require.NoError(t, err)
		require.Equal(t, vmcommon.Ok, output.ReturnCode)

		var done Done
		err = json.Unmarshal(output.ReturnData[0], &done)
		require.NoError(t, err)
		require.Equal(t, "order1", done.Order.GetID())
		require.True(t, done.Stored)
	})

	t.Run("invalid number of arguments", func(t *testing.T) {
		executor, storage := createExecutorTestContext(t)
		input := &vmcommon.ContractCallInput{
			Function: ProcessOrderEndpoint,
			VMInput: vmcommon.VMInput{
				Arguments: [][]byte{[]byte("arg1")},
			},
		}

		output, err := executor.Execute(input, storage)
		require.NoError(t, err)
		require.Equal(t, vmcommon.UserError, output.ReturnCode)
		require.Contains(t, string(output.ReturnMessage), "invalid number of arguments")
	})

	t.Run("invalid quantity", func(t *testing.T) {
		executor, storage := createExecutorTestContext(t)
		input := &vmcommon.ContractCallInput{
			Function: ProcessOrderEndpoint,
			VMInput: vmcommon.VMInput{
				Arguments: [][]byte{
					[]byte("order1"),
					{byte(SideBuy)},
					[]byte(TypeLimit),
					[]byte("invalid"),
					[]byte("100.0"),
					[]byte("0.0"),
					[]byte(GTC),
					[]byte(""),
				},
			},
		}

		output, err := executor.Execute(input, storage)
		require.NoError(t, err)
		require.Equal(t, vmcommon.UserError, output.ReturnCode)
		require.Contains(t, string(output.ReturnMessage), "invalid quantity")
	})

	t.Run("invalid price", func(t *testing.T) {
		executor, storage := createExecutorTestContext(t)
		input := &vmcommon.ContractCallInput{
			Function: ProcessOrderEndpoint,
			VMInput: vmcommon.VMInput{
				Arguments: [][]byte{
					[]byte("order1"),
					{byte(SideBuy)},
					[]byte(TypeLimit),
					[]byte("10.0"),
					[]byte("invalid"),
					[]byte("0.0"),
					[]byte(GTC),
					[]byte(""),
				},
			},
		}

		output, err := executor.Execute(input, storage)
		require.NoError(t, err)
		require.Equal(t, vmcommon.UserError, output.ReturnCode)
		require.Contains(t, string(output.ReturnMessage), "invalid price")
	})

	t.Run("invalid stop price", func(t *testing.T) {
		executor, storage := createExecutorTestContext(t)
		input := &vmcommon.ContractCallInput{
			Function: ProcessOrderEndpoint,
			VMInput: vmcommon.VMInput{
				Arguments: [][]byte{
					[]byte("order1"),
					{byte(SideBuy)},
					[]byte(TypeLimit),
					[]byte("10.0"),
					[]byte("100.0"),
					[]byte("invalid"),
					[]byte(GTC),
					[]byte(""),
				},
			},
		}

		output, err := executor.Execute(input, storage)
		require.NoError(t, err)
		require.Equal(t, vmcommon.UserError, output.ReturnCode)
		require.Contains(t, string(output.ReturnMessage), "invalid stop price")
	})
}

func TestClobExecutor_CancelOrder(t *testing.T) {
	t.Run("valid cancel", func(t *testing.T) {
		executor, storage := createExecutorTestContext(t)

		// First, add an order
		addOrderInput := &vmcommon.ContractCallInput{
			Function: ProcessOrderEndpoint,
			VMInput: vmcommon.VMInput{
				Arguments: [][]byte{
					[]byte("order1"),
					{byte(SideBuy)},
					[]byte(TypeLimit),
					[]byte("10.0"),
					[]byte("100.0"),
					[]byte("0.0"),
					[]byte(GTC),
					[]byte(""),
				},
			},
		}
		_, err := executor.Execute(addOrderInput, storage)
		require.NoError(t, err)

		// Then, cancel it
		cancelInput := &vmcommon.ContractCallInput{
			Function: CancelOrderEndpoint,
			VMInput: vmcommon.VMInput{
				Arguments: [][]byte{[]byte("order1")},
			},
		}
		output, err := executor.Execute(cancelInput, storage)
		require.NoError(t, err)
		require.Equal(t, vmcommon.Ok, output.ReturnCode)

		var order Order
		err = json.Unmarshal(output.ReturnData[0], &order)
		require.NoError(t, err)
		require.Equal(t, "order1", order.GetID())
	})

	t.Run("invalid number of arguments", func(t *testing.T) {
		executor, storage := createExecutorTestContext(t)
		input := &vmcommon.ContractCallInput{
			Function: CancelOrderEndpoint,
			VMInput:  vmcommon.VMInput{},
		}

		output, err := executor.Execute(input, storage)
		require.NoError(t, err)
		require.Equal(t, vmcommon.UserError, output.ReturnCode)
		require.Contains(t, string(output.ReturnMessage), "invalid number of arguments")
	})

	t.Run("order not found", func(t *testing.T) {
		executor, storage := createExecutorTestContext(t)
		input := &vmcommon.ContractCallInput{
			Function: CancelOrderEndpoint,
			VMInput: vmcommon.VMInput{
				Arguments: [][]byte{[]byte("nonexistent")},
			},
		}

		output, err := executor.Execute(input, storage)
		require.NoError(t, err)
		require.Equal(t, vmcommon.UserError, output.ReturnCode)
		require.Equal(t, ErrOrderNotFound.Error(), string(output.ReturnMessage))
	})
}

func TestClobExecutor_GetOrder(t *testing.T) {
	t.Run("valid get", func(t *testing.T) {
		executor, storage := createExecutorTestContext(t)

		// First, add an order
		addOrderInput := &vmcommon.ContractCallInput{
			Function: ProcessOrderEndpoint,
			VMInput: vmcommon.VMInput{
				Arguments: [][]byte{
					[]byte("order1"),
					{byte(SideBuy)},
					[]byte(TypeLimit),
					[]byte("10.0"),
					[]byte("100.0"),
					[]byte("0.0"),
					[]byte(GTC),
					[]byte(""),
				},
			},
		}
		_, err := executor.Execute(addOrderInput, storage)
		require.NoError(t, err)

		// Then, get it
		getInput := &vmcommon.ContractCallInput{
			Function: GetOrderEndpoint,
			VMInput: vmcommon.VMInput{
				Arguments: [][]byte{[]byte("order1")},
			},
		}
		output, err := executor.Execute(getInput, storage)
		require.NoError(t, err)
		require.Equal(t, vmcommon.Ok, output.ReturnCode)

		var order Order
		err = json.Unmarshal(output.ReturnData[0], &order)
		require.NoError(t, err)
		require.Equal(t, "order1", order.GetID())
	})

	t.Run("invalid number of arguments", func(t *testing.T) {
		executor, storage := createExecutorTestContext(t)
		input := &vmcommon.ContractCallInput{
			Function: GetOrderEndpoint,
			VMInput:  vmcommon.VMInput{},
		}

		output, err := executor.Execute(input, storage)
		require.NoError(t, err)
		require.Equal(t, vmcommon.UserError, output.ReturnCode)
		require.Contains(t, string(output.ReturnMessage), "invalid number of arguments")
	})

	t.Run("order not found", func(t *testing.T) {
		executor, storage := createExecutorTestContext(t)
		input := &vmcommon.ContractCallInput{
			Function: GetOrderEndpoint,
			VMInput: vmcommon.VMInput{
				Arguments: [][]byte{[]byte("nonexistent")},
			},
		}

		output, err := executor.Execute(input, storage)
		require.NoError(t, err)
		require.Equal(t, vmcommon.UserError, output.ReturnCode)
		require.Equal(t, ErrOrderNotFound.Error(), string(output.ReturnMessage))
	})
}

func TestClobExecutor_GetDepth(t *testing.T) {
	t.Run("get depth", func(t *testing.T) {
		executor, storage := createExecutorTestContext(t)

		// Add a buy order
		buyOrderInput := &vmcommon.ContractCallInput{
			Function: ProcessOrderEndpoint,
			VMInput: vmcommon.VMInput{
				Arguments: [][]byte{
					[]byte("buy1"),
					{byte(SideBuy)},
					[]byte(TypeLimit),
					[]byte("10.0"),
					[]byte("99.0"),
					[]byte("0.0"),
					[]byte(GTC),
					[]byte(""),
				},
			},
		}
		_, err := executor.Execute(buyOrderInput, storage)
		require.NoError(t, err)

		// Add a sell order
		sellOrderInput := &vmcommon.ContractCallInput{
			Function: ProcessOrderEndpoint,
			VMInput: vmcommon.VMInput{
				Arguments: [][]byte{
					[]byte("sell1"),
					{byte(SideSell)},
					[]byte(TypeLimit),
					[]byte("5.0"),
					[]byte("101.0"),
					[]byte("0.0"),
					[]byte(GTC),
					[]byte(""),
				},
			},
		}
		_, err = executor.Execute(sellOrderInput, storage)
		require.NoError(t, err)

		// Get depth
		getDepthInput := &vmcommon.ContractCallInput{
			Function: GetDepthEndpoint,
		}
		output, err := executor.Execute(getDepthInput, storage)
		require.NoError(t, err)
		require.Equal(t, vmcommon.Ok, output.ReturnCode)

		var depth Depth
		err = json.Unmarshal(output.ReturnData[0], &depth)
		require.NoError(t, err)
		require.Len(t, depth.Bids, 1)
		require.Len(t, depth.Asks, 1)
		require.Equal(t, "99", depth.Bids[0][0].Text('f', -1))
		require.Equal(t, "10", depth.Bids[0][1].Text('f', -1))
		require.Equal(t, "101", depth.Asks[0][0].Text('f', -1))
		require.Equal(t, "5", depth.Asks[0][1].Text('f', -1))
	})
}

func TestClobExecutor_MatchOrders(t *testing.T) {
	t.Run("match orders", func(t *testing.T) {
		executor, storage := createExecutorTestContext(t)

		// Add a buy order
		buyOrderInput := &vmcommon.ContractCallInput{
			Function: ProcessOrderEndpoint,
			VMInput: vmcommon.VMInput{
				Arguments: [][]byte{
					[]byte("buy1"),
					{byte(SideBuy)},
					[]byte(TypeLimit),
					[]byte("10.0"),
					[]byte("100.0"),
					[]byte("0.0"),
					[]byte(GTC),
					[]byte(""),
				},
			},
		}
		_, err := executor.Execute(buyOrderInput, storage)
		require.NoError(t, err)

		// Add a sell order that matches
		sellOrderInput := &vmcommon.ContractCallInput{
			Function: ProcessOrderEndpoint,
			VMInput: vmcommon.VMInput{
				Arguments: [][]byte{
					[]byte("sell1"),
					{byte(SideSell)},
					[]byte(TypeLimit),
					[]byte("5.0"),
					[]byte("100.0"),
					[]byte("0.0"),
					[]byte(GTC),
					[]byte(""),
				},
			},
		}
		_, err = executor.Execute(sellOrderInput, storage)
		require.NoError(t, err)

		// Manually set last price to trigger stop orders if any
		clob, err := executor.loadClob(storage)
		require.NoError(t, err)
		clob.lastPrice = big.NewFloat(100)
		err = executor.saveClob(storage, clob, ProcessOrderEndpoint)
		require.NoError(t, err)

		// Match orders
		matchInput := &vmcommon.ContractCallInput{
			Function: MatchOrdersEndpoint,
		}
		output, err := executor.Execute(matchInput, storage)
		require.NoError(t, err)
		require.Equal(t, vmcommon.Ok, output.ReturnCode)

		var done Done
		err = json.Unmarshal(output.ReturnData[0], &done)
		require.NoError(t, err)
		// MatchOrders now returns a Done object containing trades and activated stop orders
		require.NotNil(t, done)
	})
}

func TestClobExecutor_LoadClobFailsWithErrKeyNotFound(t *testing.T) {
	executor, storage := createExecutorTestContext(t)
	storage.RetrieveValueCalled = func(key []byte) ([]byte, uint32, error) {
		return nil, 0, storageCommon.ErrKeyNotFound
	}

	clob, err := executor.loadClob(storage)
	require.NoError(t, err)
	require.NotNil(t, clob)
	require.Equal(t, 0, clob.OrderBook.Bids.Len())
	require.Equal(t, 0, clob.OrderBook.Asks.Len())
}

func TestClobExecutor_SaveClobNoOp(t *testing.T) {
	executor, storage := createExecutorTestContext(t)
	var saveCalled bool
	storage.SaveKeyValueCalled = func(key, value []byte) error {
		saveCalled = true
		return nil
	}

	err := executor.saveClob(storage, NewCLOB(), GetDepthEndpoint)
	require.NoError(t, err)
	require.False(t, saveCalled)
}

func TestClobExecutor_SaveClobSaveStateFails(t *testing.T) {
	executor, storage := createExecutorTestContext(t)
	clob := NewCLOB()
	// Corrupt the clob instance to fail SaveState
	clob.OrderBook = nil

	err := executor.saveClob(storage, clob, ProcessOrderEndpoint)
	require.ErrorContains(t, err, "could not save clob state")
}

func TestClobExecutor_ComplexMatching(t *testing.T) {
	t.Run("multiple limit and market orders", func(t *testing.T) {
		executor, storage := createExecutorTestContext(t)

		// 1. Add initial limit orders to populate the book
		_, err := executor.Execute(createLimitOrderInput("buy1", SideBuy, 10, 99), storage)
		require.NoError(t, err)
		_, err = executor.Execute(createLimitOrderInput("buy2", SideBuy, 5, 98), storage)
		require.NoError(t, err)
		_, err = executor.Execute(createLimitOrderInput("sell1", SideSell, 8, 101), storage)
		require.NoError(t, err)
		_, err = executor.Execute(createLimitOrderInput("sell2", SideSell, 12, 102), storage)
		require.NoError(t, err)

		clob, _ := executor.loadClob(storage)
		require.Equal(t, 2, clob.OrderBook.Bids.Len())
		require.Equal(t, 2, clob.OrderBook.Asks.Len())

		// 2. A market sell order that partially fills the best bid
		marketSellInput := &vmcommon.ContractCallInput{
			Function: ProcessOrderEndpoint,
			VMInput: vmcommon.VMInput{
				Arguments: [][]byte{
					[]byte("market_sell1"),
					{byte(SideSell)},
					[]byte(TypeMarket),
					[]byte("7.0"),
					[]byte("0"),
					[]byte("0"),
					[]byte(GTC),
					[]byte(""),
				},
			},
		}
		output, err := executor.Execute(marketSellInput, storage)
		require.NoError(t, err)
		var done Done
		err = json.Unmarshal(output.ReturnData[0], &done)
		require.NoError(t, err)
		require.Len(t, done.Trades, 2)
		require.False(t, done.Stored)

		clob, _ = executor.loadClob(storage)
		require.Equal(t, 2, clob.OrderBook.Bids.Len())
		bestBid := clob.OrderBook.Bids.Best()
		require.Equal(t, 0, bestBid.GetQuantity().Cmp(big.NewFloat(3)))

		// 3. A limit buy order that crosses the spread and matches
		limitBuyInput := createLimitOrderInput("buy3_cross", SideBuy, 10, 101)
		output, err = executor.Execute(limitBuyInput, storage)
		require.NoError(t, err)
		err = json.Unmarshal(output.ReturnData[0], &done)
		require.NoError(t, err)
		require.Len(t, done.Trades, 2)
		require.True(t, done.Stored)

		clob, _ = executor.loadClob(storage)
		require.Equal(t, 3, clob.OrderBook.Bids.Len())
		require.Equal(t, 1, clob.OrderBook.Asks.Len())
		bestAsk := clob.OrderBook.Asks.Best()
		require.Equal(t, "sell2", bestAsk.GetID())

		// 4. Final check with MatchOrders - should do nothing as the book is not crossed
		matchInput := &vmcommon.ContractCallInput{
			Function: MatchOrdersEndpoint,
		}
		output, err = executor.Execute(matchInput, storage)
		require.NoError(t, err)
		var doneAfterMatch Done
		err = json.Unmarshal(output.ReturnData[0], &doneAfterMatch)
		require.NoError(t, err)
		require.Nil(t, doneAfterMatch.Order)
		require.Empty(t, doneAfterMatch.Trades)
	})
}

func createLimitOrderInput(id string, side Side, quantity float64, price float64) *vmcommon.ContractCallInput {
	return &vmcommon.ContractCallInput{
		Function: ProcessOrderEndpoint,
		VMInput: vmcommon.VMInput{
			Arguments: [][]byte{
				[]byte(id),
				{byte(side)},
				[]byte(TypeLimit),
				[]byte(fmt.Sprintf("%f", quantity)),
				[]byte(fmt.Sprintf("%f", price)),
				[]byte("0"),
				[]byte(GTC),
				[]byte(""),
			},
		},
	}
}

func BenchmarkClobExecutor_MatchOrders(b *testing.B) {
	executor, storage := createExecutorTestContext(b)

	// Populate the order book with a large number of orders
	for i := 0; i < 100; i++ {
		buyPrice := 99.0 - float64(i)*0.01
		sellPrice := 101.0 + float64(i)*0.01
		_, _ = executor.Execute(createLimitOrderInput(fmt.Sprintf("buy-%d", i), SideBuy, 1, buyPrice), storage)
		_, _ = executor.Execute(createLimitOrderInput(fmt.Sprintf("sell-%d", i), SideSell, 1, sellPrice), storage)
	}

	// Add a crossing order to trigger matching
	crossingOrder := createLimitOrderInput("cross", SideBuy, 100, 102)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Create a fresh copy of storage for each run to avoid side effects
		cleanStorage := testscommon.NewAccountHandlerStub()
		savedState, _, _ := storage.RetrieveValue([]byte(clobStorageKey))
		_ = cleanStorage.SaveKeyValue([]byte(clobStorageKey), savedState)

		_, _ = executor.Execute(crossingOrder, cleanStorage)
	}
}