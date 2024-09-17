package state

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/vm"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

// SaveAliasAddress saves the given alias address
func (adb *AccountsDB) SaveAliasAddress(request *vmcommon.AliasSaveRequest) error {
	err := vmcommon.ValidateAliasSaveRequest(request)
	if err != nil {
		return err
	}

	return adb.saveAliasAddress(request)
}

// RequestAddress returns the requested address
func (adb *AccountsDB) RequestAddress(request *vmcommon.AddressRequest) (*vmcommon.AddressResponse, error) {
	err := vmcommon.ValidateAddressRequest(request)
	if err != nil {
		return nil, err
	}

	err = vmcommon.EnhanceAddressRequest(request)
	if err != nil {
		return nil, err
	}

	return adb.requestAddress(request)
}

func (adb *AccountsDB) saveAccounts(account UserAccountHandler, aliasSCAccount UserAccountHandler) error {
	err := adb.SaveAccount(account)
	if err != nil {
		return err
	}
	return adb.SaveAccount(aliasSCAccount)
}

func (adb *AccountsDB) saveAccountsIfRequired(account UserAccountHandler, aliasSCAccount UserAccountHandler, request *vmcommon.AddressRequest, isGenerated bool) error {
	if isGenerated && request.SaveOnGenerate {
		return adb.saveAccounts(account, aliasSCAccount)
	}
	return nil
}

func (adb *AccountsDB) requestAddress(request *vmcommon.AddressRequest) (*vmcommon.AddressResponse, error) {
	aliasSCAccount, err := adb.loadAccountHandler(vm.AliasSCAddress)
	if err != nil {
		return nil, err
	}

	account, isAccountGenerated, err := adb.loadAccountHandlerForRequest(aliasSCAccount, request)
	if err != nil {
		return nil, err
	}

	mainAddress, mainAddressIdentifier, isMainAddressIdentifierGenerated, err := requestMainAddress(account, request)
	if err != nil {
		return nil, err
	}

	switch request.RequestedIdentifier {
	case core.MVXAddressIdentifier:
		err = adb.saveAccountsIfRequired(account, aliasSCAccount, request, isAccountGenerated || isMainAddressIdentifierGenerated)
		return &vmcommon.AddressResponse{MultiversXAddress: account.AddressBytes(), RequestedAddress: account.AddressBytes()}, err
	default:
		requestedAddress, isRequestedAddressGenerated, aliasErr := requestAliasAddress(account, aliasSCAccount, mainAddress, mainAddressIdentifier, request)
		if aliasErr != nil {
			return nil, aliasErr
		}
		err = adb.saveAccountsIfRequired(account, aliasSCAccount, request, isAccountGenerated || isMainAddressIdentifierGenerated || isRequestedAddressGenerated)
		return &vmcommon.AddressResponse{MultiversXAddress: account.AddressBytes(), RequestedAddress: requestedAddress}, err
	}
}

func (adb *AccountsDB) saveAliasAddress(request *vmcommon.AliasSaveRequest) error {
	aliasSCAccount, err := adb.loadAccountHandler(vm.AliasSCAddress)
	if err != nil {
		return err
	}

	account, err := adb.loadAccountHandler(request.MultiversXAddress)
	if err != nil {
		return err
	}

	existingAccountRequest := &vmcommon.AddressRequest{
		SaveOnGenerate:      false,
		SourceAddress:       request.AliasAddress,
		SourceIdentifier:    request.AliasIdentifier,
		RequestedIdentifier: core.MVXAddressIdentifier,
	}
	existingAccount, isGenerated, err := adb.loadAccountHandlerForAlias(aliasSCAccount, existingAccountRequest)
	if err != nil {
		return err
	}
	if !isGenerated && isAccountUsed(existingAccount) {
		return ErrAliasAddressCollision
	}

	err = saveAlias(account, aliasSCAccount, request.AliasAddress, request.AliasIdentifier)
	if err != nil {
		return err
	}

	return adb.saveAccounts(account, aliasSCAccount)
}
