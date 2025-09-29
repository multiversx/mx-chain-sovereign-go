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
	if input != nil {
		arguments = input.VMInput.Arguments
	}

	clobInstance, err := ce.loadClob(storage)
	if err != nil {
		return nil, err
	}

	sc := NewClobSC(clobInstance)
	var returnData [][]byte
	var returnErr error

	switch funcName {
	case ProcessOrderEndpoint:
		returnData, returnErr = ce.processOrder(sc, arguments)
	case CancelOrderEndpoint:
		returnData, returnErr = ce.cancelOrder(sc, arguments)
	case GetOrderEndpoint:
		returnData, returnErr = ce.getOrder(sc, arguments)
	case GetDepthEndpoint:
		returnData, returnErr = ce.getDepth(sc)
	case MatchOrdersEndpoint:
		returnData, returnErr = ce.matchOrders(sc)
	default:
		returnErr = fmt.Errorf("invalid function name: %s", funcName)
	}

	if returnErr != nil {
		return &vmcommon.VMOutput{
			ReturnCode:    vmcommon.UserError,
			ReturnMessage: returnErr.Error(),
		}, nil
	}

	err = ce.saveClob(storage, clobInstance, funcName)
	if err != nil {
		return nil, err
	}

	return &vmcommon.VMOutput{
		ReturnData: returnData,
		ReturnCode: vmcommon.Ok,
	}, nil
}

func (ce *clobExecutor) loadClob(storage vmcommon.AccountDataHandler) (*CLOB, error) {
	clobInstance := NewCLOB()
	data, _, err := storage.RetrieveValue([]byte(clobStorageKey))
	if err != nil && err != storageCommon.ErrKeyNotFound {
		return nil, fmt.Errorf("failed to retrieve clob data: %w", err)
	}
	if len(data) > 0 {
		err = clobInstance.LoadState(data)
		if err != nil {
			return nil, fmt.Errorf("could not load clob state: %w", err)
		}
	}
	return clobInstance, nil
}

func (ce *clobExecutor) saveClob(storage vmcommon.AccountDataHandler, clobInstance *CLOB, funcName string) error {
	switch funcName {
	case ProcessOrderEndpoint, CancelOrderEndpoint, MatchOrdersEndpoint:
		newState, err := clobInstance.SaveState()
		if err != nil {
			return fmt.Errorf("could not save clob state: %w", err)
		}
		err = storage.SaveKeyValue([]byte(clobStorageKey), newState)
		if err != nil {
			return fmt.Errorf("could not save key-value: %w", err)
		}
	}
	return nil
}

func (ce *clobExecutor) processOrder(sc *clobSC, arguments [][]byte) ([][]byte, error) {
	if len(arguments) < 8 {
		return nil, fmt.Errorf("invalid number of arguments for %s", ProcessOrderEndpoint)
	}
	orderID := string(arguments[0])
	side := Side(arguments[1][0])
	orderType := OrderType(arguments[2])
	quantity, _, err := big.ParseFloat(string(arguments[3]), 10, 0, big.ToNearestEven)
	if err != nil {
		return nil, fmt.Errorf("invalid quantity: %w", err)
	}
	price, _, err := big.ParseFloat(string(arguments[4]), 10, 0, big.ToNearestEven)
	if err != nil {
		return nil, fmt.Errorf("invalid price: %w", err)
	}
	stop, _, err := big.ParseFloat(string(arguments[5]), 10, 0, big.ToNearestEven)
	if err != nil {
		return nil, fmt.Errorf("invalid stop price: %w", err)
	}
	tif := TIF(arguments[6])
	oco := string(arguments[7])

	ret, err := sc.ProcessOrder(orderID, side, orderType, quantity, price, stop, tif, oco)
	if err != nil {
		return nil, err
	}

	retBytes, err := json.Marshal(ret)
	if err != nil {
		return nil, fmt.Errorf("could not marshal return data: %w", err)
	}
	return [][]byte{retBytes}, nil
}

func (ce *clobExecutor) cancelOrder(sc *clobSC, arguments [][]byte) ([][]byte, error) {
	if len(arguments) < 1 {
		return nil, fmt.Errorf("invalid number of arguments for %s", CancelOrderEndpoint)
	}
	orderID := string(arguments[0])
	ret, err := sc.CancelOrder(orderID)
	if err != nil {
		return nil, err
	}

	retBytes, err := json.Marshal(ret)
	if err != nil {
		return nil, fmt.Errorf("could not marshal return data: %w", err)
	}
	return [][]byte{retBytes}, nil
}

func (ce *clobExecutor) getOrder(sc *clobSC, arguments [][]byte) ([][]byte, error) {
	if len(arguments) < 1 {
		return nil, fmt.Errorf("invalid number of arguments for %s", GetOrderEndpoint)
	}
	orderID := string(arguments[0])
	ret, err := sc.GetOrder(orderID)
	if err != nil {
		return nil, err
	}

	retBytes, err := json.Marshal(ret)
	if err != nil {
		return nil, fmt.Errorf("could not marshal return data: %w", err)
	}
	return [][]byte{retBytes}, nil
}

func (ce *clobExecutor) getDepth(sc *clobSC) ([][]byte, error) {
	ret := sc.GetDepth()
	retBytes, err := json.Marshal(ret)
	if err != nil {
		return nil, fmt.Errorf("could not marshal return data: %w", err)
	}
	return [][]byte{retBytes}, nil
}

func (ce *clobExecutor) matchOrders(sc *clobSC) ([][]byte, error) {
	ret, err := sc.MatchOrders()
	if err != nil {
		return nil, err
	}

	retBytes, err := json.Marshal(ret)
	if err != nil {
		return nil, fmt.Errorf("could not marshal return data: %w", err)
	}
	return [][]byte{retBytes}, nil
}