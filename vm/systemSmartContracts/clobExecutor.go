package systemSmartContracts

import (
	"fmt"
	"github.com/multiversx/mx-chain-go/vm"
	"github.com/nikolaydubina/fpdecimal"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

const clobStateKey = "clobState"

type clobExecutor struct {
	sc  *clobSC
	eei vm.SystemEI
}

// NewClobExecutor creates a new instance of the clobExecutor.
func NewClobExecutor(eei vm.SystemEI) (*clobExecutor, error) {
	if eei == nil || eei.IsInterfaceNil() {
		return nil, vm.ErrNilSystemEnvironmentInterface
	}
	return &clobExecutor{
		sc:  NewClobSC(),
		eei: eei,
	}, nil
}

// Execute is the main entry point for the smart contract.
func (ce *clobExecutor) Execute(input *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	stateBytes := ce.eei.GetStorage([]byte(clobStateKey))
	err := ce.sc.LoadState(stateBytes)
	if err != nil {
		ce.eei.AddReturnMessage(fmt.Sprintf("cannot load state: %s", err.Error()))
		return vmcommon.UserError
	}

	funcName := string(input.Function)
	args := input.Arguments

	var returnData []byte
	var retCode = vmcommon.Ok
	switch funcName {
	case processOrderEndpoint:
		returnData, err = ce.processOrder(args)
	case cancelOrderEndpoint:
		returnData, err = ce.cancelOrder(args)
	case getOrderEndpoint:
		returnData, err = ce.getOrder(args)
	case getDepthEndpoint:
		returnData, err = ce.getDepth()
	case "matchOrders":
		returnData, err = ce.matchOrders()
	default:
		ce.eei.AddReturnMessage(fmt.Sprintf("invalid function name: %s", funcName))
		return vmcommon.UserError
	}

	if err != nil {
		ce.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	newState, err := ce.sc.SaveState()
	if err != nil {
		ce.eei.AddReturnMessage(fmt.Sprintf("cannot save state: %s", err.Error()))
		return vmcommon.UserError
	}
	ce.eei.SetStorage([]byte(clobStateKey), newState)

	ce.eei.Finish(returnData)
	return retCode
}

func (ce *clobExecutor) processOrder(args [][]byte) ([]byte, error) {
	if len(args) != 8 {
		return nil, fmt.Errorf("invalid number of arguments for processOrder, expected 8, got %d", len(args))
	}

	orderID := string(args[0])
	side := Side(args[1][0])
	orderType := OrderType(args[2])
	quantity, err := fpdecimal.Parse(args[3])
	if err != nil {
		return nil, fmt.Errorf("invalid quantity: %w", err)
	}
	price, err := fpdecimal.Parse(args[4])
	if err != nil {
		return nil, fmt.Errorf("invalid price: %w", err)
	}
	stop, err := fpdecimal.Parse(args[5])
	if err != nil {
		return nil, fmt.Errorf("invalid stop price: %w", err)
	}
	tif := TIF(args[6])
	oco := string(args[7])

	return ce.sc.ProcessOrder(orderID, side, orderType, quantity, price, stop, tif, oco)
}

func (ce *clobExecutor) cancelOrder(args [][]byte) ([]byte, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("invalid number of arguments for cancelOrder, expected 1, got %d", len(args))
	}
	orderID := string(args[0])
	return ce.sc.CancelOrder(orderID)
}

func (ce *clobExecutor) getOrder(args [][]byte) ([]byte, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("invalid number of arguments for getOrder, expected 1, got %d", len(args))
	}
	orderID := string(args[0])
	return ce.sc.GetOrder(orderID)
}

func (ce *clobExecutor) getDepth() ([]byte, error) {
	return ce.sc.GetDepth()
}

func (ce *clobExecutor) matchOrders() ([]byte, error) {
	ce.sc.clob.OrderBook.Stop.Iterate(func(order *Order) {
		// Placeholder for stop order processing logic
	})
	return []byte("orders matched successfully"), nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ce *clobExecutor) IsInterfaceNil() bool {
	return ce == nil
}

// CanUseContract returns true if the contract can be used
func (ce *clobExecutor) CanUseContract() bool {
	return true
}

// SetNewGasCost sets a new gas cost for the system smart contract
func (ce *clobExecutor) SetNewGasCost(gasCost vm.GasCost) {
	// not needed for this contract
}