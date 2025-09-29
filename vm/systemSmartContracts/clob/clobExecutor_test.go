package clob

import (
	"encoding/json"
	"testing"

	"github.com/multiversx/mx-chain-storage-go/mock"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

func TestNewClobExecutor(t *testing.T) {
	ce := NewClobExecutor()
	require.NotNil(t, ce)
}

func TestClobExecutor_Execute(t *testing.T) {
	storage := &mock.AccountTracker{}
	executor := NewClobExecutor()

	t.Run("Process a new limit buy order", func(t *testing.T) {
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

		output, err := executor.Execute(input, storage)
		require.NoError(t, err)
		require.Equal(t, vmcommon.Ok, output.ReturnCode)

		var done Done
		err = json.Unmarshal(output.ReturnData[0], &done)
		require.NoError(t, err)
		require.Equal(t, "order1", done.Order.GetID())
		require.True(t, done.Stored)
	})

	t.Run("Get the newly created order", func(t *testing.T) {
		input := &vmcommon.ContractCallInput{
			Function: []byte(GetOrderEndpoint),
			VMInput: &vmcommon.VMInput{
				Arguments: [][]byte{[]byte("order1")},
			},
		}

		output, err := executor.Execute(input, storage)
		require.NoError(t, err)
		require.Equal(t, vmcommon.Ok, output.ReturnCode)

		var order Order
		err = json.Unmarshal(output.ReturnData[0], &order)
		require.NoError(t, err)
		require.Equal(t, "order1", order.GetID())
	})

	t.Run("Process a matching limit sell order", func(t *testing.T) {
		input := &vmcommon.ContractCallInput{
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

		output, err := executor.Execute(input, storage)
		require.NoError(t, err)
		require.Equal(t, vmcommon.Ok, output.ReturnCode)

		var done Done
		err = json.Unmarshal(output.ReturnData[0], &done)
		require.NoError(t, err)
		require.Equal(t, "order2", done.Order.GetID())
		require.Len(t, done.Trades, 2)
		require.False(t, done.Stored)
	})

	t.Run("Get depth after match", func(t *testing.T) {
		input := &vmcommon.ContractCallInput{
			Function: []byte(GetDepthEndpoint),
			VMInput:  &vmcommon.VMInput{},
		}

		output, err := executor.Execute(input, storage)
		require.NoError(t, err)
		require.Equal(t, vmcommon.Ok, output.ReturnCode)

		var depth Depth
		err = json.Unmarshal(output.ReturnData[0], &depth)
		require.NoError(t, err)
		require.Len(t, depth.Bids, 1)
		require.Equal(t, "5", depth.Bids[0][1].Text('f', -1))
		require.Len(t, depth.Asks, 0)
	})

	t.Run("Cancel the remaining order", func(t *testing.T) {
		input := &vmcommon.ContractCallInput{
			Function: []byte(CancelOrderEndpoint),
			VMInput: &vmcommon.VMInput{
				Arguments: [][]byte{[]byte("order1")},
			},
		}

		output, err := executor.Execute(input, storage)
		require.NoError(t, err)
		require.Equal(t, vmcommon.Ok, output.ReturnCode)

		var order Order
		err = json.Unmarshal(output.ReturnData[0], &order)
		require.NoError(t, err)
		require.Equal(t, "order1", order.GetID())
	})

	t.Run("Get depth after cancellation", func(t *testing.T) {
		input := &vmcommon.ContractCallInput{
			Function: []byte(GetDepthEndpoint),
			VMInput:  &vmcommon.VMInput{},
		}

		output, err := executor.Execute(input, storage)
		require.NoError(t, err)
		require.Equal(t, vmcommon.Ok, output.ReturnCode)

		var depth Depth
		err = json.Unmarshal(output.ReturnData[0], &depth)
		require.NoError(t, err)
		require.Len(t, depth.Bids, 0)
		require.Len(t, depth.Asks, 0)
	})

	t.Run("Invalid function call", func(t *testing.T) {
		input := &vmcommon.ContractCallInput{
			Function: []byte("invalidFunction"),
			VMInput:  &vmcommon.VMInput{},
		}

		output, err := executor.Execute(input, storage)
		require.NoError(t, err)
		require.Equal(t, vmcommon.UserError, output.ReturnCode)
		require.Contains(t, string(output.ReturnMessage), "invalid function name")
	})
}