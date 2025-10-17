package state

import (
	"github.com/multiversx/mx-chain-core-go/core"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

func requestMainAddressIdentifier(account UserAccountHandler, request *vmcommon.AddressRequest) (core.AddressIdentifier, bool, error) {
	mainAddressIdentifier, isGenerated, err := fetchOrGenerateMainAddressIdentifier(account, request.SourceIdentifier)
	if err != nil {
		return core.InvalidAddressIdentifier, false, err
	}

	if isGenerated && request.SaveOnGenerate {
		err = saveMainAddressIdentifier(account, mainAddressIdentifier)
		if err != nil {
			return core.InvalidAddressIdentifier, false, err
		}
	}
	return mainAddressIdentifier, isGenerated, nil
}

func requestMainAddress(account UserAccountHandler, request *vmcommon.AddressRequest) ([]byte, core.AddressIdentifier, bool, error) {
	mainAddressIdentifier, isGenerated, err := requestMainAddressIdentifier(account, request)
	if err != nil {
		return nil, core.InvalidAddressIdentifier, false, err
	}

	switch mainAddressIdentifier {
	case request.SourceIdentifier:
		return request.SourceAddress, mainAddressIdentifier, isGenerated, nil
	case core.MVXAddressIdentifier:
		return account.AddressBytes(), mainAddressIdentifier, isGenerated, nil
	default:
		aliasAddress, aliasErr := FetchValidAliasAddress(account, mainAddressIdentifier)
		if aliasErr != nil {
			return nil, core.InvalidAddressIdentifier, false, aliasErr
		}
		return aliasAddress, mainAddressIdentifier, isGenerated, nil
	}
}

func requestAliasAddress(account UserAccountHandler, aliasSCAccount UserAccountHandler, mainAddress []byte, mainAddressIdentifier core.AddressIdentifier, request *vmcommon.AddressRequest) ([]byte, bool, error) {
	aliasAddress, isGenerated, err := fetchOrGenerateAliasAddress(account, mainAddress, mainAddressIdentifier, request.RequestedIdentifier)
	if err != nil {
		return nil, false, err
	}

	if isGenerated && request.SaveOnGenerate {
		err = saveAlias(account, aliasSCAccount, aliasAddress, request.RequestedIdentifier)
		if err != nil {
			return nil, false, err
		}
	}
	return aliasAddress, isGenerated, nil
}
