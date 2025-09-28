package systemSmartContracts

import (
	"fmt"

	"github.com/multiversx/mx-chain-go/vm/systemSmartContracts/clob"
	storageCommon "github.com/multiversx/mx-chain-storage-go/common"
	vmcommon "github.comcom/multiversx/mx-chain-vm-common-go"
	"github.com/nikolaydubina/fpdecimal"
)

const clobStorageKey = "clob"

type clobExecutor struct {
}

func NewClobExecutor() *clobExecutor {
	return &clobExecutor{}
}

// Execute will have the core dispatch logic. It will parse the function name from input.Function,
// instantiate the clobSC with the provided storage handler, call the appropriate endpoint on clobSC
// based on the function name and construct and return a vmcommon.VMOutput with the results and a vmcommon.Ok return code.
func (ce *clobExecutor) Execute(input *vmcommon.ContractCallInput, storage vmcommon.AccountDataHandler) (*vmcommon.VMOutput, error) {
	funcName := string(input.Function)

	// Load CLOB from storage
	clobInstance := clob.NewCLOB()
	data, _, err := storage.RetrieveValue([]byte(clobStorageKey))
	if err != nil && err != storageCommon.ErrKeyNotFound {
		return nil, err
	}
	// if err is nil or ErrKeyNotFound, proceed. data will be nil if not found.
	err = clobInstance.LoadState(data)
	if err != nil {
		return nil, fmt.Errorf("could not load clob state: %w", err)
	}

	sc := clob.NewClobSC(clobInstance)
	var returnData [][]byte
	var returnErr error

	switch funcName {
	case clob.ProcessOrderEndpoint:
		if len(input.Arguments) < 8 {
			returnErr = fmt.Errorf("invalid number of arguments for %s", funcName)
			break
		}
		orderID := string(input.Arguments[0])
		side := clob.Side(input.Arguments[1][0])
		orderType := clob.OrderType(input.Arguments[2])
		quantity, err := fpdecimal.NewFromString(string(input.Arguments[3]))
		if err != nil {
			returnErr = fmt.Errorf("invalid quantity: %w", err)
			break
		}
		price, err := fpdecimal.NewFromString(string(input.Arguments[4]))
		if err != nil {
			returnErr = fmt.Errorf("invalid price: %w", err)
			break
		}
		stop, err := fpdecimal.NewFromString(string(input.Arguments[5]))
		if err != nil {
			returnErr = fmt.Errorf("invalid stop price: %w", err)
			break
		}
		tif := clob.TIF(input.Arguments[6])
		oco := string(input.Arguments[7])

		ret, err := sc.ProcessOrder(orderID, side, orderType, quantity, price, stop, tif, oco)
		if err != nil {
			returnErr = err
		} else {
			returnData = append(returnData, ret)
		}

	case clob.CancelOrderEndpoint:
		if len(input.Arguments) < 1 {
			returnErr = fmt.Errorf("invalid number of arguments for %s", funcName)
			break
		}
		orderID := string(input.Arguments[0])
		ret, err := sc.CancelOrder(orderID)
		if err != nil {
			returnErr = err
		} else {
			returnData = append(returnData, ret)
		}

	case clob.GetOrderEndpoint:
		if len(input.Arguments) < 1 {
			returnErr = fmt.Errorf("invalid number of arguments for %s", funcName)
			break
		}
		orderID := string(input.Arguments[0])
		ret, err := sc.GetOrder(orderID)
		if err != nil {
			returnErr = err
		} else {
			returnData = append(returnData, ret)
		}

	case clob.GetDepthEndpoint:
		ret, err := sc.GetDepth()
		if err != nil {
			returnErr = err
		} else {
			returnData = append(returnData, ret)
		}

	default:
		returnErr = fmt.Errorf("invalid function name: %s", funcName)
	}

	if returnErr != nil {
		return &vmcommon.VMOutput{
			ReturnCode: vmcommon.InternalError,
			ReturnMessage: []byte(returnErr.Error()),
		}, nil
	}

	// Save the new state, only for endpoints that modify it
	switch funcName {
	case clob.ProcessOrderEndpoint, clob.CancelOrderEndpoint:
		newState, err := clobInstance.SaveState()
		if err != nil {
			return nil, fmt.Errorf("could not save clob state: %w", err)
		}
		err = storage.SaveKeyValue([]byte(clobStorageKey), newState)
		if err != nil {
			return nil, fmt.Errorf("could not save key-value: %w", err)
		}
	}


	return &vmcommon.VMOutput{
		ReturnData: returnData,
		ReturnCode: vmcommon.Ok,
	}, nil
}

// matchOrders takes no input, but starts to iterate the orders and resolve those.
func (ce *clobExecutor) matchOrders(storage vmcommon.AccountDataHandler) error {
	// Load CLOB from storage
	clobInstance := clob.NewCLOB()
	data, _, err := storage.RetrieveValue([]byte(clobStorageKey))
	if err != nil {
		if err == storageCommon.ErrKeyNotFound {
			return nil // Nothing to match
		}
		return err
	}
	err = clobInstance.LoadState(data)
	if err != nil {
		return fmt.Errorf("could not load clob state: %w", err)
	}

    // The current implementation matches orders upon insertion. A dedicated `matchOrders` function
    // could be used for end-of-block processing. The main utility would be activating stop orders
    // that might have been triggered by price movements within the block.
    // However, the current `OrderBook` logic triggers stop order activation when a trade occurs at a certain price.
    // A standalone matching function would require a more sophisticated implementation that
    // takes a price feed.
    // For now, this function is a placeholder.

	// Save the state back in case any latent matching logic is added in the future.
	newState, err := clobInstance.SaveState()
	if err != nil {
		return fmt.Errorf("could not save clob state: %w", err)
	}
	err = storage.SaveKeyValue([]byte(clobStorageKey), newState)
	if err != nil {
		return fmt.Errorf("could not save key-value: %w", err)
	}

	return nil
}