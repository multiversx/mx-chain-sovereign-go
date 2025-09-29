package clob

import (
	"encoding/json"
	"testing"

	vmMock "github.com/multiversx/mx-chain-go/vm/mock"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

func createCLOBTestContext(t *testing.T) (*clobSC, *vmMock.SystemEIStub) {
	sc, err := NewClobSC()
	require.NoError(t, err)
	return sc, &vmMock.SystemEIStub{}
}

func TestClobSC_All(t *testing.T) {
	t.Run("Process a new limit buy order", func(t *testing.T) {
		sc, eei := createCLOBTestContext(t)
		var finishedData [][]byte
		eei.FinishCalled = func(data []byte) {
			finishedData = append(finishedData, data)
		}

		input := &vmcommon.ContractCallInput{
			Function: []byte(ProcessOrderEndpoint),
			VMInput: &vmcommon.VMInput{
				Arguments: [][]byte{
					[]byte("order1"),
					[]byte{byte(SideBuy)},
					[]byte(TypeLimit),
					[]byte("10.0"),
					[]byte("100.0"),
					[]byte("0.0"),
					[]byte(TIF_GTC),
					[]byte(""),
				},
			},
		}

		retCode := sc.Execute(input, eei)
		require.Equal(t, vmcommon.Ok, retCode)

		var done Done
		err := json.Unmarshal(finishedData[0], &done)
		require.NoError(t, err)
		require.Equal(t, "order1", done.Order.GetID())
		require.True(t, done.Stored)
	})

	t.Run("Get the newly created order", func(t *testing.T) {
		sc, eei := createCLOBTestContext(t)
		var finishedData [][]byte
		eei.FinishCalled = func(data []byte) {
			finishedData = append(finishedData, data)
		}

		// First, add an order
		addOrderInput := &vmcommon.ContractCallInput{
			Function: []byte(ProcessOrderEndpoint),
			VMInput: &vmcommon.VMInput{
				Arguments: [][]byte{
					[]byte("order1"),
					[]byte{byte(SideBuy)},
					[]byte(TypeLimit),
					[]byte("10.0"),
					[]byte("100.0"),
					[]byte("0.0"),
					[]byte(TIF_GTC),
					[]byte(""),
				},
			},
		}
		retCode := sc.Execute(addOrderInput, eei)
		require.Equal(t, vmcommon.Ok, retCode)
		finishedData = nil

		// Then, get the order
		getOrderInput := &vmcommon.ContractCallInput{
			Function: []byte(GetOrderEndpoint),
			VMInput: &vmcommon.VMInput{
				Arguments: [][]byte{[]byte("order1")},
			},
		}

		retCode = sc.Execute(getOrderInput, eei)
		require.Equal(t, vmcommon.Ok, retCode)

		var order Order
		err := json.Unmarshal(finishedData[0], &order)
		require.NoError(t, err)
		require.Equal(t, "order1", order.GetID())
	})

	t.Run("Process a matching limit sell order", func(t *testing.T) {
		sc, eei := createCLOBTestContext(t)
		var finishedData [][]byte
		eei.FinishCalled = func(data []byte) {
			finishedData = append(finishedData, data)
		}

		// Add buy order
		buyOrderInput := &vmcommon.ContractCallInput{
			Function: []byte(ProcessOrderEndpoint),
			VMInput: &vmcommon.VMInput{
				Arguments: [][]byte{
					[]byte("order1"),
					[]byte{byte(SideBuy)},
					[]byte(TypeLimit),
					[]byte("10.0"),
					[]byte("100.0"),
					[]byte("0.0"),
					[]byte(TIF_GTC),
					[]byte(""),
				},
			},
		}
		retCode := sc.Execute(buyOrderInput, eei)
		require.Equal(t, vmcommon.Ok, retCode)
		finishedData = nil

		// Add matching sell order
		sellOrderInput := &vmcommon.ContractCallInput{
			Function: []byte(ProcessOrderEndpoint),
			VMInput: &vmcommon.VMInput{
				Arguments: [][]byte{
					[]byte("order2"),
					[]byte{byte(SideSell)},
					[]byte(TypeLimit),
					[]byte("5.0"),
					[]byte("100.0"),
					[]byte("0.0"),
					[]byte(TIF_GTC),
					[]byte(""),
				},
			},
		}
		retCode = sc.Execute(sellOrderInput, eei)
		require.Equal(t, vmcommon.Ok, retCode)

		var done Done
		err := json.Unmarshal(finishedData[0], &done)
		require.NoError(t, err)
		require.Equal(t, "order2", done.Order.GetID())
		require.Len(t, done.Trades, 2)
		require.False(t, done.Stored)
	})

	t.Run("Get depth after match", func(t *testing.T) {
		sc, eei := createCLOBTestContext(t)
		var finishedData [][]byte
		eei.FinishCalled = func(data []byte) {
			finishedData = append(finishedData, data)
		}

		// Add buy order
		buyOrderInput := &vmcommon.ContractCallInput{
			Function: []byte(ProcessOrderEndpoint),
			VMInput: &vmcommon.VMInput{
				Arguments: [][]byte{
					[]byte("order1"),
					[]byte{byte(SideBuy)},
					[]byte(TypeLimit),
					[]byte("10.0"),
					[]byte("100.0"),
					[]byte("0.0"),
					[]byte(TIF_GTC),
					[]byte(""),
				},
			},
		}
		sc.Execute(buyOrderInput, eei)

		// Add matching sell order
		sellOrderInput := &vmcommon.ContractCallInput{
			Function: []byte(ProcessOrderEndpoint),
			VMInput: &vmcommon.VMInput{
				Arguments: [][]byte{
					[]byte("order2"),
					[]byte{byte(SideSell)},
					[]byte(TypeLimit),
					[]byte("5.0"),
					[]byte("100.0"),
					[]byte("0.0"),
					[]byte(TIF_GTC),
					[]byte(""),
				},
			},
		}
		sc.Execute(sellOrderInput, eei)
		finishedData = nil

		// Get depth
		getDepthInput := &vmcommon.ContractCallInput{
			Function: []byte(GetDepthEndpoint),
			VMInput:  &vmcommon.VMInput{},
		}
		retCode := sc.Execute(getDepthInput, eei)
		require.Equal(t, vmcommon.Ok, retCode)

		var depth Depth
		err := json.Unmarshal(finishedData[0], &depth)
		require.NoError(t, err)
		require.Len(t, depth.Bids, 1)
		require.Equal(t, "5", depth.Bids[0][1].Text('f', -1))
		require.Len(t, depth.Asks, 0)
	})

	t.Run("Cancel the remaining order", func(t *testing.T) {
		sc, eei := createCLOBTestContext(t)
		var finishedData [][]byte
		eei.FinishCalled = func(data []byte) {
			finishedData = append(finishedData, data)
		}

		// Add buy order
		buyOrderInput := &vmcommon.ContractCallInput{
			Function: []byte(ProcessOrderEndpoint),
			VMInput: &vmcommon.VMInput{
				Arguments: [][]byte{
					[]byte("order1"),
					[]byte{byte(SideBuy)},
					[]byte(TypeLimit),
					[]byte("10.0"),
					[]byte("100.0"),
					[]byte("0.0"),
					[]byte(TIF_GTC),
					[]byte(""),
				},
			},
		}
		sc.Execute(buyOrderInput, eei)
		finishedData = nil

		// Cancel order
		cancelOrderInput := &vmcommon.ContractCallInput{
			Function: []byte(CancelOrderEndpoint),
			VMInput: &vmcommon.VMInput{
				Arguments: [][]byte{[]byte("order1")},
			},
		}
		retCode := sc.Execute(cancelOrderInput, eei)
		require.Equal(t, vmcommon.Ok, retCode)

		var order Order
		err := json.Unmarshal(finishedData[0], &order)
		require.NoError(t, err)
		require.Equal(t, "order1", order.GetID())
	})

	t.Run("Get depth after cancellation", func(t *testing.T) {
		sc, eei := createCLOBTestContext(t)
		var finishedData [][]byte
		eei.FinishCalled = func(data []byte) {
			finishedData = append(finishedData, data)
		}

		// Add buy order
		buyOrderInput := &vmcommon.ContractCallInput{
			Function: []byte(ProcessOrderEndpoint),
			VMInput: &vmcommon.VMInput{
				Arguments: [][]byte{
					[]byte("order1"),
					[]byte{byte(SideBuy)},
					[]byte(TypeLimit),
					[]byte("10.0"),
					[]byte("100.0"),
					[]byte("0.0"),
					[]byte(TIF_GTC),
					[]byte(""),
				},
			},
		}
		sc.Execute(buyOrderInput, eei)

		// Cancel order
		cancelOrderInput := &vmcommon.ContractCallInput{
			Function: []byte(CancelOrderEndpoint),
			VMInput: &vmcommon.VMInput{
				Arguments: [][]byte{[]byte("order1")},
			},
		}
		sc.Execute(cancelOrderInput, eei)
		finishedData = nil

		// Get depth
		getDepthInput := &vmcommon.ContractCallInput{
			Function: []byte(GetDepthEndpoint),
			VMInput:  &vmcommon.VMInput{},
		}
		retCode := sc.Execute(getDepthInput, eei)
		require.Equal(t, vmcommon.Ok, retCode)

		var depth Depth
		err := json.Unmarshal(finishedData[0], &depth)
		require.NoError(t, err)
		require.Len(t, depth.Bids, 0)
		require.Len(t, depth.Asks, 0)
	})

	t.Run("Invalid function call", func(t *testing.T) {
		sc, eei := createCLOBTestContext(t)
		var returnMessage string
		eei.AddReturnMessageCalled = func(msg string) {
			returnMessage = msg
		}

		input := &vmcommon.ContractCallInput{
			Function: []byte("invalidFunction"),
			VMInput:  &vmcommon.VMInput{},
		}

		retCode := sc.Execute(input, eei)
		require.Equal(t, vmcommon.UserError, retCode)
		require.Contains(t, returnMessage, "invalid function name")
	})

	t.Run("Process a market buy order", func(t *testing.T) {
		sc, eei := createCLOBTestContext(t)
		var finishedData [][]byte
		eei.FinishCalled = func(data []byte) {
			finishedData = append(finishedData, data)
		}

		// Add a limit sell order to match against
		sellOrderInput := &vmcommon.ContractCallInput{
			Function: []byte(ProcessOrderEndpoint),
			VMInput: &vmcommon.VMInput{
				Arguments: [][]byte{
					[]byte("sell1"),
					[]byte{byte(SideSell)},
					[]byte(TypeLimit),
					[]byte("10.0"),
					[]byte("100.0"),
					[]byte("0.0"),
					[]byte(TIF_GTC),
					[]byte(""),
				},
			},
		}
		retCode := sc.Execute(sellOrderInput, eei)
		require.Equal(t, vmcommon.Ok, retCode)
		finishedData = nil

		// Process market buy order
		marketBuyInput := &vmcommon.ContractCallInput{
			Function: []byte(ProcessOrderEndpoint),
			VMInput: &vmcommon.VMInput{
				Arguments: [][]byte{
					[]byte("market1"),
					[]byte{byte(SideBuy)},
					[]byte(TypeMarket),
					[]byte("5.0"),
					[]byte("0.0"), // Price is ignored for market orders
					[]byte("0.0"),
					[]byte(""), // TIF is ignored for market orders
					[]byte(""),
				},
			},
		}
		retCode = sc.Execute(marketBuyInput, eei)
		require.Equal(t, vmcommon.Ok, retCode)

		var done Done
		err := json.Unmarshal(finishedData[0], &done)
		require.NoError(t, err)
		require.Equal(t, "market1", done.Order.GetID())
		require.Len(t, done.Trades, 1)
		require.False(t, done.Stored)
		require.Equal(t, "5", done.Trades[0].GetQuantity().Text('f', -1))
		require.Equal(t, "100", done.Trades[0].GetPrice().Text('f', -1))
	})

	t.Run("Process an IOC limit order - partial fill", func(t *testing.T) {
		sc, eei := createCLOBTestContext(t)
		var finishedData [][]byte
		eei.FinishCalled = func(data []byte) {
			finishedData = append(finishedData, data)
		}

		// Add a limit sell order to match against
		sellOrderInput := &vmcommon.ContractCallInput{
			Function: []byte(ProcessOrderEndpoint),
			VMInput: &vmcommon.VMInput{
				Arguments: [][]byte{
					[]byte("sell1"),
					[]byte{byte(SideSell)},
					[]byte(TypeLimit),
					[]byte("5.0"),
					[]byte("100.0"),
					[]byte("0.0"),
					[]byte(TIF_GTC),
					[]byte(""),
				},
			},
		}
		retCode := sc.Execute(sellOrderInput, eei)
		require.Equal(t, vmcommon.Ok, retCode)
		finishedData = nil

		// Process IOC buy order that will be partially filled
		iocBuyInput := &vmcommon.ContractCallInput{
			Function: []byte(ProcessOrderEndpoint),
			VMInput: &vmcommon.VMInput{
				Arguments: [][]byte{
					[]byte("ioc1"),
					[]byte{byte(SideBuy)},
					[]byte(TypeLimit),
					[]byte("10.0"),
					[]byte("100.0"),
					[]byte("0.0"),
					[]byte(TIF_IOC),
					[]byte(""),
				},
			},
		}
		retCode = sc.Execute(iocBuyInput, eei)
		require.Equal(t, vmcommon.Ok, retCode)

		var done Done
		err := json.Unmarshal(finishedData[0], &done)
		require.NoError(t, err)
		require.Equal(t, "ioc1", done.Order.GetID())
		require.Len(t, done.Trades, 1)
		require.False(t, done.Stored) // IOC orders are never stored
		require.Equal(t, "5", done.Trades[0].GetQuantity().Text('f', -1))

		// Check that the unfilled part of the IOC order was canceled
		depthInput := &vmcommon.ContractCallInput{
			Function: []byte(GetDepthEndpoint),
			VMInput:  &vmcommon.VMInput{},
		}
		finishedData = nil
		sc.Execute(depthInput, eei)
		var depth Depth
		err = json.Unmarshal(finishedData[0], &depth)
		require.NoError(t, err)
		require.Len(t, depth.Bids, 0) // No buy orders should be left
	})

	t.Run("Process a FOK limit order - should be killed", func(t *testing.T) {
		sc, eei := createCLOBTestContext(t)
		var finishedData [][]byte
		eei.FinishCalled = func(data []byte) {
			finishedData = append(finishedData, data)
		}

		// Add a limit sell order that won't fully fill the FOK order
		sellOrderInput := &vmcommon.ContractCallInput{
			Function: []byte(ProcessOrderEndpoint),
			VMInput: &vmcommon.VMInput{
				Arguments: [][]byte{
					[]byte("sell1"),
					[]byte{byte(SideSell)},
					[]byte(TypeLimit),
					[]byte("5.0"),
					[]byte("100.0"),
					[]byte("0.0"),
					[]byte(TIF_GTC),
					[]byte(""),
				},
			},
		}
		retCode := sc.Execute(sellOrderInput, eei)
		require.Equal(t, vmcommon.Ok, retCode)
		finishedData = nil

		// Process FOK buy order that cannot be fully filled
		fokBuyInput := &vmcommon.ContractCallInput{
			Function: []byte(ProcessOrderEndpoint),
			VMInput: &vmcommon.VMInput{
				Arguments: [][]byte{
					[]byte("fok1"),
					[]byte{byte(SideBuy)},
					[]byte(TypeLimit),
					[]byte("10.0"),
					[]byte("100.0"),
					[]byte("0.0"),
					[]byte(TIF_FOK),
					[]byte(""),
				},
			},
		}
		retCode = sc.Execute(fokBuyInput, eei)
		require.Equal(t, vmcommon.Ok, retCode)

		var done Done
		err := json.Unmarshal(finishedData[0], &done)
		require.NoError(t, err)
		require.Equal(t, "fok1", done.Order.GetID())
		require.Len(t, done.Trades, 0)   // No trades should occur
		require.False(t, done.Stored) // FOK orders are never stored

		// Check that the order book is not empty
		depthInput := &vmcommon.ContractCallInput{
			Function: []byte(GetDepthEndpoint),
			VMInput:  &vmcommon.VMInput{},
		}
		finishedData = nil
		sc.Execute(depthInput, eei)
		var depth Depth
		err = json.Unmarshal(finishedData[0], &depth)
		require.NoError(t, err)
		require.Len(t, depth.Asks, 1) // The sell order should still be there
	})

	t.Run("Cancel a non-existent order", func(t *testing.T) {
		sc, eei := createCLOBTestContext(t)
		var returnMessage string
		eei.AddReturnMessageCalled = func(msg string) {
			returnMessage = msg
		}

		cancelOrderInput := &vmcommon.ContractCallInput{
			Function: []byte(CancelOrderEndpoint),
			VMInput: &vmcommon.VMInput{
				Arguments: [][]byte{[]byte("nonexistent")},
			},
		}
		retCode := sc.Execute(cancelOrderInput, eei)
		require.Equal(t, vmcommon.UserError, retCode)
		require.Contains(t, returnMessage, ErrOrderNotFound.Error())
	})

	t.Run("Activate a buy stop order", func(t *testing.T) {
		sc, eei := createCLOBTestContext(t)
		var finishedData [][]byte
		eei.FinishCalled = func(data []byte) {
			finishedData = append(finishedData, data)
		}

		// Add a stop-limit buy order
		stopBuyInput := &vmcommon.ContractCallInput{
			Function: []byte(ProcessOrderEndpoint),
			VMInput: &vmcommon.VMInput{
				Arguments: [][]byte{
					[]byte("stop1"),
					[]byte{byte(SideBuy)},
					[]byte(TypeStopLimit),
					[]byte("10.0"),
					[]byte("110.0"), // Limit price
					[]byte("105.0"), // Stop price
					[]byte(TIF_GTC),
					[]byte(""),
				},
			},
		}
		retCode := sc.Execute(stopBuyInput, eei)
		require.Equal(t, vmcommon.Ok, retCode)
		finishedData = nil

		// Manually trigger a trade to set the last price
		// In a real scenario, this would happen through order matching
		sc.clob.lastPrice = big.NewFloat(105)

		// Call matchOrders to activate the stop order
		matchOrdersInput := &vmcommon.ContractCallInput{
			Function: []byte(MatchOrdersEndpoint),
			VMInput:  &vmcommon.VMInput{},
		}
		retCode = sc.Execute(matchOrdersInput, eei)
		require.Equal(t, vmcommon.Ok, retCode)

		// Check that the stop order was converted to a limit order
		getOrderInput := &vmcommon.ContractCallInput{
			Function: []byte(GetOrderEndpoint),
			VMInput: &vmcommon.VMInput{
				Arguments: [][]byte{[]byte("stop1")},
			},
		}
		finishedData = nil
		retCode = sc.Execute(getOrderInput, eei)
		require.Equal(t, vmcommon.Ok, retCode)

		var order Order
		err := json.Unmarshal(finishedData[0], &order)
		require.NoError(t, err)
		require.Equal(t, TypeLimit, order.GetType())
		require.Equal(t, "110", order.GetPrice().Text('f', -1))
	})

	t.Run("Activate a sell stop order", func(t *testing.T) {
		sc, eei := createCLOBTestContext(t)
		var finishedData [][]byte
		eei.FinishCalled = func(data []byte) {
			finishedData = append(finishedData, data)
		}

		// Add a stop-limit sell order
		stopSellInput := &vmcommon.ContractCallInput{
			Function: []byte(ProcessOrderEndpoint),
			VMInput: &vmcommon.VMInput{
				Arguments: [][]byte{
					[]byte("stop1"),
					[]byte{byte(SideSell)},
					[]byte(TypeStopLimit),
					[]byte("10.0"),
					[]byte("90.0"),  // Limit price
					[]byte("95.0"),  // Stop price
					[]byte(TIF_GTC),
					[]byte(""),
				},
			},
		}
		retCode := sc.Execute(stopSellInput, eei)
		require.Equal(t, vmcommon.Ok, retCode)
		finishedData = nil

		// Manually trigger a trade to set the last price
		sc.clob.lastPrice = big.NewFloat(95)

		// Call matchOrders to activate the stop order
		matchOrdersInput := &vmcommon.ContractCallInput{
			Function: []byte(MatchOrdersEndpoint),
			VMInput:  &vmcommon.VMInput{},
		}
		retCode = sc.Execute(matchOrdersInput, eei)
		require.Equal(t, vmcommon.Ok, retCode)

		// Check that the stop order was converted to a limit order
		getOrderInput := &vmcommon.ContractCallInput{
			Function: []byte(GetOrderEndpoint),
			VMInput: &vmcommon.VMInput{
				Arguments: [][]byte{[]byte("stop1")},
			},
		}
		finishedData = nil
		retCode = sc.Execute(getOrderInput, eei)
		require.Equal(t, vmcommon.Ok, retCode)

		var order Order
		err := json.Unmarshal(finishedData[0], &order)
		require.NoError(t, err)
		require.Equal(t, TypeLimit, order.GetType())
		require.Equal(t, "90", order.GetPrice().Text('f', -1))
	})
}