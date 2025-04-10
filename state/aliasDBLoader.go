package state

import (
	"github.com/multiversx/mx-chain-core-go/core"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

func (adb *AccountsDB) loadAccountHandler(address []byte) (UserAccountHandler, error) {
	foundAccount, err := adb.LoadAccount(address)
	if err != nil {
		return nil, err
	}
	account, ok := foundAccount.(UserAccountHandler)
	if !ok {
		return nil, ErrWrongTypeAssertion
	}
	return account, nil
}

func (adb *AccountsDB) loadAccountHandlerForAlias(aliasSCAccount UserAccountHandler, request *vmcommon.AddressRequest) (UserAccountHandler, bool, error) {
	multiversXAddress, isGenerated, err := fetchOrGenerateMultiversXAddress(aliasSCAccount, request.SourceAddress, request.SourceIdentifier)
	if err != nil {
		return nil, false, err
	}

	account, err := adb.loadAccountHandler(multiversXAddress)
	if err != nil {
		return nil, false, err
	}

	if isGenerated && request.SaveOnGenerate {
		err = saveAlias(account, aliasSCAccount, request.SourceAddress, request.SourceIdentifier)
		if err != nil {
			return nil, false, err
		}
	}
	return account, isGenerated, nil
}

func (adb *AccountsDB) loadAccountHandlerForRequest(aliasSCAccount UserAccountHandler, request *vmcommon.AddressRequest) (UserAccountHandler, bool, error) {
	switch request.SourceIdentifier {
	case core.MVXAddressIdentifier:
		account, err := adb.loadAccountHandler(request.SourceAddress)
		return account, false, err
	default:
		return adb.loadAccountHandlerForAlias(aliasSCAccount, request)
	}
}
