package clob

import (
	"encoding/json"
	"fmt"
	"math/big"

	storageCommon "github.com/multiversx/mx-chain-storage-go/common"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
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
	var arguments [][]byte
	if input.VMInput != nil {
		arguments = input.VMInput.Arguments
	}

	// Load CLOB from storage
	clobInstance := NewCLOB()
	data, _, err := storage.RetrieveValue([]byte(clobStorageKey))
	if err != nil && err != storageCommon.ErrKeyNotFound {
		return nil, fmt.Errorf("failed to retrieve clob data: %w", err)
	}
	// if err is nil or ErrKeyNotFound, proceed. data will be nil if not found.
	if len(data) > 0 {
		err = clobInstance.LoadState(data)
		if err != nil {
			return nil, fmt.Errorf("could not load clob state: %w", err)
		}
	}

	sc := NewClobSC(clobInstance)
	var returnData [][]byte
	var returnErr error

	switch funcName {
	case ProcessOrderEndpoint:
		if len(arguments) < 8 {
			returnErr = fmt.Errorf("invalid number of arguments for %s", funcName)
			break
		}
		orderID := string(arguments[0])
		side := Side(arguments[1][0])
		orderType := OrderType(arguments[2])
		quantity, _, err := big.ParseFloat(string(arguments[3]), 10, 0, big.ToNearestEven)
		if err != nil {
			returnErr = fmt.Errorf("invalid quantity: %w", err)
			break
		}
		price, _, err := big.ParseFloat(string(arguments[4]), 10, 0, big.ToNearestEven)
		if err != nil {
			returnErr = fmt.Errorf("invalid price: %w", err)
			break
		}
		stop, _, err := big.ParseFloat(string(arguments[5]), 10, 0, big.ToNearestEven)
		if err != nil {
			returnErr = fmt.Errorf("invalid stop price: %w", err)
			break
		}
		tif := TIF(arguments[6])
		oco := string(arguments[7])

		ret, err := sc.ProcessOrder(orderID, side, orderType, quantity, price, stop, tif, oco)
		if err != nil {
			returnErr = err
		} else {
			// Marshal ret to JSON
			retBytes, err := json.Marshal(ret)
			if err != nil {
				returnErr = fmt.Errorf("could not marshal return data: %w", err)
			} else {
				returnData = append(returnData, retBytes)
			}
		}

	case CancelOrderEndpoint:
		if len(arguments) < 1 {
			returnErr = fmt.Errorf("invalid number of arguments for %s", funcName)
			break
		}
		orderID := string(arguments[0])
		ret, err := sc.CancelOrder(orderID)
		if err != nil {
			returnErr = err
		} else {
			retBytes, err := json.Marshal(ret)
			if err != nil {
				returnErr = fmt.Errorf("could not marshal return data: %w", err)
			} else {
				returnData = append(returnData, retBytes)
			}
		}

	case GetOrderEndpoint:
		if len(arguments) < 1 {
			returnErr = fmt.Errorf("invalid number of arguments for %s", funcName)
			break
		}
		orderID := string(arguments[0])
		ret, err := sc.GetOrder(orderID)
		if err != nil {
			returnErr = err
		} else {
			retBytes, err := json.Marshal(ret)
			if err != nil {
				returnErr = fmt.Errorf("could not marshal return data: %w", err)
			} else {
				returnData = append(returnData, retBytes)
			}
		}

	case GetDepthEndpoint:
		ret := sc.GetDepth()
		retBytes, err := json.Marshal(ret)
		if err != nil {
			returnErr = fmt.Errorf("could not marshal return data: %w", err)
		} else {
			returnData = append(returnData, retBytes)
		}

	case MatchOrdersEndpoint:
		ret, err := sc.MatchOrders()
		if err != nil {
			returnErr = err
		} else {
			retBytes, err := json.Marshal(ret)
			if err != nil {
				returnErr = fmt.Errorf("could not marshal return data: %w", err)
			} else {
				returnData = append(returnData, retBytes)
			}
		}

	default:
		returnErr = fmt.Errorf("invalid function name: %s", funcName)
	}

	if returnErr != nil {
		return &vmcommon.VMOutput{
			ReturnCode:    vmcommon.UserError,
			ReturnMessage: []byte(returnErr.Error()),
		}, nil
	}

	// Save the new state, only for endpoints that modify it
	switch funcName {
	case ProcessOrderEndpoint, CancelOrderEndpoint, MatchOrdersEndpoint:
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