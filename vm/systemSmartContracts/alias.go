package systemSmartContracts

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/vm"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"strconv"
	"sync"
)

const crossAddressTransfer = "crossAddressTransfer"

const crossAddressTransferArgsSize = 2
const sourceAddressIndex = 0
const sourceIdentifierIndex = 1

type AliasContract struct {
	eei          vm.SystemEI
	gasCost      vm.GasCost
	mutExecution sync.RWMutex
}

// ArgsAliasContract defines the arguments to create the alias smart contract
type ArgsAliasContract struct {
	Eei     vm.SystemEI
	GasCost vm.GasCost
}

// NewAliasContract creates a new alias smart contract
func NewAliasContract(args ArgsAliasContract) (*AliasContract, error) {
	if check.IfNil(args.Eei) {
		return nil, vm.ErrNilSystemEnvironmentInterface
	}
	return &AliasContract{
		eei:     args.Eei,
		gasCost: args.GasCost,
	}, nil
}

// Execute calls one of the functions from the alias contract and runs the code according to the input
func (ac *AliasContract) Execute(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	ac.mutExecution.RLock()
	defer ac.mutExecution.RUnlock()

	err := CheckIfNil(args)
	if err != nil {
		ac.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	switch args.Function {
	case core.SCDeployInitFunctionName:
		return vmcommon.Ok
	case crossAddressTransfer:
		return ac.doCrossAddressTransfer(args)
	default:
		return vmcommon.FunctionNotFound
	}
}

// CanUseContract returns true if contract is enabled
func (ac *AliasContract) CanUseContract() bool {
	return true
}

// SetNewGasCost is called whenever a gas cost was changed
func (ac *AliasContract) SetNewGasCost(gasCost vm.GasCost) {
	ac.mutExecution.Lock()
	ac.gasCost = gasCost
	ac.mutExecution.Unlock()
}

// IsInterfaceNil returns true if underlying object is nil
func (ac *AliasContract) IsInterfaceNil() bool {
	return ac == nil
}

func (ac *AliasContract) doCrossAddressTransfer(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if len(args.Arguments) != crossAddressTransferArgsSize {
		ac.eei.AddReturnMessage("expected " + strconv.Itoa(crossAddressTransferArgsSize) + " arguments")
		return vmcommon.FunctionWrongSignature
	}

	sourceAddress := args.Arguments[sourceAddressIndex]
	sourceIdentifier := core.ParseAddressIdentifier(args.Arguments[sourceIdentifierIndex])
	multiversXAddress, err := ac.determineMultiversXAddress(sourceAddress, sourceIdentifier)
	if err != nil {
		ac.eei.AddReturnMessage("address conversion failed")
		return vmcommon.UserError
	}

	ac.eei.Transfer(multiversXAddress, args.RecipientAddr, args.CallValue, nil, 0)
	return vmcommon.Ok
}

func (ac *AliasContract) determineMultiversXAddress(sourceAddress []byte, sourceIdentifier core.AddressIdentifier) ([]byte, error) {
	switch sourceIdentifier {
	case core.InvalidAddressIdentifier:
		return nil, vm.ErrInvalidAddress
	case core.MVXAddressIdentifier:
		return sourceAddress, nil
	default:
		requestedAddress, err := ac.eei.RequestAddress(&vmcommon.AddressRequest{
			SourceAddress:       sourceAddress,
			SourceIdentifier:    sourceIdentifier,
			RequestedIdentifier: core.MVXAddressIdentifier,
			SaveOnGenerate:      true,
		})
		if err != nil {
			return nil, err
		}
		return requestedAddress.RequestedAddress, nil
	}
}
