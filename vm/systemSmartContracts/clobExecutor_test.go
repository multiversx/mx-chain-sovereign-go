package systemSmartContracts

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-go/vm"
	"github.com/multiversx/mx-chain-go/vm/mock"
	"github.com/nikolaydubina/fpdecimal"
	"github.com/stretchr/testify/require"

	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

func createClobExecutorTestObjects() (*clobExecutor, *mock.SystemEIStub) {
	storage := make(map[string][]byte)
	eei := &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			return storage[string(key)]
		},
		SetStorageCalled: func(key, value []byte) {
			storage[string(key)] = value
		},
		FinishCalled: func(value []byte) {},
	}

	executor, err := NewClobExecutor(eei)
	if err != nil {
		return nil, nil
	}

	return executor, eei
}

func TestNewClobExecutor(t *testing.T) {
	t.Parallel()

	t.Run("nil eei", func(t *testing.T) {
		t.Parallel()
		executor, err := NewClobExecutor(nil)
		require.Equal(t, vm.ErrNilSystemEnvironmentInterface, err)
		require.Nil(t, executor)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()
		_, eei := createClobExecutorTestObjects()
		executor, err := NewClobExecutor(eei)
		require.NoError(t, err)
		require.NotNil(t, executor)
	})
}

func TestClobExecutor_Execute(t *testing.T) {
	t.Parallel()

	executor, eei := createClobExecutorTestObjects()

	// Test case 1: Process a new order
	orderID := "order1"
	side := SideBuy
	orderType := TypeLimit
	quantity, _ := fpdecimal.Parse([]byte("10"))
	price, _ := fpdecimal.Parse([]byte("100"))
	stop, _ := fpdecimal.Parse([]byte("0"))
	tif := TIF_GTC
	oco := ""

	args := [][]byte{
		[]byte(orderID),
		[]byte{byte(side)},
		[]byte(orderType),
		[]byte("10"),
		[]byte("100"),
		[]byte("0"),
		[]byte(tif),
		[]byte(oco),
	}

	input := &vmcommon.ContractCallInput{
		Function:  []byte(processOrderEndpoint),
		Arguments: args,
	}

	retCode := executor.Execute(input)
	require.Equal(t, vmcommon.Ok, retCode)

	// Verify that the order was added to the order book
	input = &vmcommon.ContractCallInput{
		Function:  []byte(getOrderEndpoint),
		Arguments: [][]byte{[]byte(orderID)},
	}

	var returnedData []byte
	eei.FinishCalled = func(value []byte) {
		returnedData = value
	}

	retCode = executor.Execute(input)
	require.Equal(t, vmcommon.Ok, retCode)

	var order Order
	err := json.Unmarshal(returnedData, &order)
	require.NoError(t, err)
	require.Equal(t, orderID, order.GetID())

	// Test case 2: Cancel the order
	input = &vmcommon.ContractCallInput{
		Function:  []byte(cancelOrderEndpoint),
		Arguments: [][]byte{[]byte(orderID)},
	}

	retCode = executor.Execute(input)
	require.Equal(t, vmcommon.Ok, retCode)

	// Verify that the order was removed
	input = &vmcommon.ContractCallInput{
		Function:  []byte(getOrderEndpoint),
		Arguments: [][]byte{[]byte(orderID)},
	}

	retCode = executor.Execute(input)
	require.Equal(t, vmcommon.UserError, retCode)

	// Test case 3: Match orders
	input = &vmcommon.ContractCallInput{
		Function: []byte("matchOrders"),
	}

	retCode = executor.Execute(input)
	require.Equal(t, vmcommon.Ok, retCode)
}

func TestClobExecutor_InvalidFunction(t *testing.T) {
	t.Parallel()

	executor, eei := createClobExecutorTestObjects()

	input := &vmcommon.ContractCallInput{
		Function: []byte("invalidFunction"),
	}

	var returnMessage string
	eei.AddReturnMessageCalled = func(message string) {
		returnMessage = message
	}

	retCode := executor.Execute(input)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Equal(t, "invalid function name: invalidFunction", returnMessage)
}

func TestClobExecutor_GetDepth(t *testing.T) {
	t.Parallel()

	executor, eei := createClobExecutorTestObjects()

	// Add a buy order
	buyOrderID := "buyOrder"
	buyQuantity, _ := fpdecimal.Parse([]byte("10"))
	buyPrice, _ := fpdecimal.Parse([]byte("100"))
	buyArgs := [][]byte{
		[]byte(buyOrderID),
		[]byte{byte(SideBuy)},
		[]byte(TypeLimit),
		[]byte("10"),
		[]byte("100"),
		[]byte("0"),
		[]byte(TIF_GTC),
		[]byte(""),
	}
	input := &vmcommon.ContractCallInput{
		Function:  []byte(processOrderEndpoint),
		Arguments: buyArgs,
	}
	retCode := executor.Execute(input)
	require.Equal(t, vmcommon.Ok, retCode)

	// Add a sell order
	sellOrderID := "sellOrder"
	sellQuantity, _ := fpdecimal.Parse([]byte("5"))
	sellPrice, _ := fpdecimal.Parse([]byte("101"))
	sellArgs := [][]byte{
		[]byte(sellOrderID),
		[]byte{byte(SideSell)},
		[]byte(TypeLimit),
		[]byte("5"),
		[]byte("101"),
		[]byte("0"),
		[]byte(TIF_GTC),
		[]byte(""),
	}
	input = &vmcommon.ContractCallInput{
		Function:  []byte(processOrderEndpoint),
		Arguments: sellArgs,
	}
	retCode = executor.Execute(input)
	require.Equal(t, vmcommon.Ok, retCode)

	// Get depth
	input = &vmcommon.ContractCallInput{
		Function: []byte(getDepthEndpoint),
	}

	var returnedData []byte
	eei.FinishCalled = func(value []byte) {
		returnedData = value
	}

	retCode = executor.Execute(input)
	require.Equal(t, vmcommon.Ok, retCode)

	var depth Depth
	err := json.Unmarshal(returnedData, &depth)
	require.NoError(t, err)

	require.Len(t, depth.Bids, 1)
	require.Equal(t, buyPrice, depth.Bids[0][0])
	require.Equal(t, buyQuantity, depth.Bids[0][1])

	require.Len(t, depth.Asks, 1)
	require.Equal(t, sellPrice, depth.Asks[0][0])
	require.Equal(t, sellQuantity, depth.Asks[0][1])
}

func TestClobExecutor_Persistence(t *testing.T) {
	t.Parallel()

	executor, eei := createClobExecutorTestObjects()
	storage := make(map[string][]byte)
	eei.GetStorageCalled = func(key []byte) []byte {
		return storage[string(key)]
	}
	eei.SetStorageCalled = func(key, value []byte) {
		storage[string(key)] = value
	}

	orderID := "order1"
	args := [][]byte{
		[]byte(orderID),
		[]byte{byte(SideBuy)},
		[]byte(TypeLimit),
		[]byte("10"),
		[]byte("100"),
		[]byte("0"),
		[]byte(TIF_GTC),
		[]byte(""),
	}

	input := &vmcommon.ContractCallInput{
		Function:  []byte(processOrderEndpoint),
		Arguments: args,
		VMInput: vmcommon.VMInput{
			CallerAddr: []byte("caller"),
			CallValue:  big.NewInt(0),
		},
	}

	retCode := executor.Execute(input)
	require.Equal(t, vmcommon.Ok, retCode)

	// create a new executor and check if it loads the state correctly
	executor2, eei2 := createClobExecutorTestObjects()
	eei2.GetStorageCalled = func(key []byte) []byte {
		return storage[string(key)]
	}

	inputGet := &vmcommon.ContractCallInput{
		Function:  []byte(getOrderEndpoint),
		Arguments: [][]byte{[]byte(orderID)},
	}

	var returnedData []byte
	eei2.FinishCalled = func(value []byte) {
		returnedData = value
	}

	retCode = executor2.Execute(inputGet)
	require.Equal(t, vmcommon.Ok, retCode)

	var order Order
	err := json.Unmarshal(returnedData, &order)
	require.NoError(t, err)
	require.Equal(t, orderID, order.GetID())
}

func TestClobExecutor_ExecuteLoadStateFails(t *testing.T) {
	t.Parallel()

	executor, eei := createClobExecutorTestObjects()
	eei.GetStorageCalled = func(key []byte) []byte {
		return []byte("invalid json")
	}

	input := &vmcommon.ContractCallInput{
		Function: []byte(processOrderEndpoint),
	}

	var returnMessage string
	eei.AddReturnMessageCalled = func(message string) {
		returnMessage = message
	}

	retCode := executor.Execute(input)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Contains(t, returnMessage, "cannot load state")
}

func TestClobExecutor_ExecuteSaveStateFails(t *testing.T) {
	t.Parallel()

	executor, eei := createClobExecutorTestObjects()
	badDecimal, _ := fpdecimal.Parse([]byte("1.000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001"))
	executor.sc.clob.OrderBook.orders["bad"] = &Order{
		price: badDecimal,
	}

	input := &vmcommon.ContractCallInput{
		Function: []byte(getDepthEndpoint),
	}

	var returnMessage string
	eei.AddReturnMessageCalled = func(message string) {
		returnMessage = message
	}

	retCode := executor.Execute(input)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Contains(t, returnMessage, "cannot save state")
}