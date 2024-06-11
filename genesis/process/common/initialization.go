package common

import (
	"math/big"

	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/vm"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

var zero = big.NewInt(0)

// UpdateSystemSCContractsCode will update system accounts contract codes, code, and owner, except for the sys account for each shard
func UpdateSystemSCContractsCode(contractMetadata []byte, userAccountsDB state.AccountsAdapter) error {
	contractsToUpdate := make([][]byte, 0)
	contractsToUpdate = append(contractsToUpdate, vm.StakingSCAddress)
	contractsToUpdate = append(contractsToUpdate, vm.ValidatorSCAddress)
	contractsToUpdate = append(contractsToUpdate, vm.GovernanceSCAddress)
	contractsToUpdate = append(contractsToUpdate, vm.ESDTSCAddress)
	contractsToUpdate = append(contractsToUpdate, vm.DelegationManagerSCAddress)
	contractsToUpdate = append(contractsToUpdate, vm.FirstDelegationSCAddress)

	for _, address := range contractsToUpdate {
		userAcc, err := getUserAccount(address, userAccountsDB)
		if err != nil {
			return err
		}

		userAcc.SetOwnerAddress(address)
		userAcc.SetCodeMetadata(contractMetadata)
		userAcc.SetCode(address)

		err = userAccountsDB.SaveAccount(userAcc)
		if err != nil {
			return err
		}
	}

	return nil
}

func getUserAccount(address []byte, userAccountsDB state.AccountsAdapter) (state.UserAccountHandler, error) {
	acc, err := userAccountsDB.LoadAccount(address)
	if err != nil {
		return nil, err
	}

	userAcc, ok := acc.(state.UserAccountHandler)
	if !ok {
		return nil, errors.ErrWrongTypeAssertion
	}

	return userAcc, nil
}

func InitDelegationSystemSC(systemVM vmcommon.VMExecutionHandler, userAccountsDB state.AccountsAdapter) error {
	codeMetaData := &vmcommon.CodeMetadata{
		Upgradeable: false,
		Payable:     false,
		Readable:    true,
	}

	vmInput := &vmcommon.ContractCreateInput{
		VMInput: vmcommon.VMInput{
			CallerAddr: vm.DelegationManagerSCAddress,
			Arguments:  [][]byte{},
			CallValue:  big.NewInt(0),
		},
		ContractCode:         vm.DelegationManagerSCAddress,
		ContractCodeMetadata: codeMetaData.ToBytes(),
	}

	vmOutput, err := systemVM.RunSmartContractCreate(vmInput)
	if err != nil {
		return err
	}
	if vmOutput.ReturnCode != vmcommon.Ok {
		return epochStart.ErrCouldNotInitDelegationSystemSC
	}

	err = ProcessSCOutputAccounts(vmOutput, userAccountsDB)
	if err != nil {
		return err
	}

	err = UpdateSystemSCContractsCode(vmInput.ContractCodeMetadata, userAccountsDB)
	if err != nil {
		return err
	}

	return nil
}

func ProcessSCOutputAccounts(
	vmOutput *vmcommon.VMOutput,
	userAccountsDB state.AccountsAdapter,
) error {

	outputAccounts := process.SortVMOutputInsideData(vmOutput)
	for _, outAcc := range outputAccounts {
		acc, err := getUserAccount(outAcc.Address, userAccountsDB)
		if err != nil {
			return err
		}

		storageUpdates := process.GetSortedStorageUpdates(outAcc)
		for _, storeUpdate := range storageUpdates {
			err = acc.SaveKeyValue(storeUpdate.Offset, storeUpdate.Data)
			if err != nil {
				return err
			}
		}

		if outAcc.BalanceDelta != nil && outAcc.BalanceDelta.Cmp(zero) != 0 {
			err = acc.AddToBalance(outAcc.BalanceDelta)
			if err != nil {
				return err
			}
		}

		err = userAccountsDB.SaveAccount(acc)
		if err != nil {
			return err
		}
	}

	return nil
}
