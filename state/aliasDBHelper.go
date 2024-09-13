package state

import (
	"bytes"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-crypto-go/address"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

var multiversXDefaultAddress = make([]byte, len(core.SystemAccountAddress))

var ethereumDefaultAddress = make([]byte, common.AddressLength)

func enhanceRequest(request *vmcommon.AddressRequest) error {
	if len(request.SourceAddress) == 0 {
		defaultAddress, err := requestDefaultAddress(request.SourceIdentifier)
		if err != nil {
			return err
		}
		request.SourceAddress = defaultAddress
	}
	return nil
}

func isDefaultSourceAddress(sourceAddress []byte, sourceIdentifier core.AddressIdentifier) bool {
	switch sourceIdentifier {
	case core.MVXAddressIdentifier:
		return bytes.Equal(sourceAddress, multiversXDefaultAddress)
	case core.ETHAddressIdentifier:
		return bytes.Equal(sourceAddress, ethereumDefaultAddress)
	default:
		return false
	}
}

func requestDefaultAddress(requestedIdentifier core.AddressIdentifier) ([]byte, error) {
	switch requestedIdentifier {
	case core.MVXAddressIdentifier:
		return multiversXDefaultAddress, nil
	case core.ETHAddressIdentifier:
		return ethereumDefaultAddress, nil
	default:
		return nil, ErrFunctionalityNotImplemented
	}
}

func buildMainAddressIdentifierKey() []byte {
	return []byte(core.ProtectedKeyPrefix + "MainAddressIdentifier")
}

func buildAliasAddressKey(aliasIdentifier core.AddressIdentifier) []byte {
	return []byte(core.ProtectedKeyPrefix + "AliasAddress" + aliasIdentifier.String())
}

func buildMultiversXAddressKey(aliasAddress []byte, aliasIdentifier core.AddressIdentifier) []byte {
	return []byte(core.ProtectedKeyPrefix + "MultiversXAddress" + aliasIdentifier.BuildAddressIdentifier(aliasAddress))
}

func isAccountUsed(account UserAccountHandler) bool {
	return account.GetNonce() > 0 || len(account.GetCodeHash()) > 0
}

func retrieveValueIgnoreNilTrie(account UserAccountHandler, key []byte) ([]byte, error) {
	value, _, err := account.RetrieveValue(key)
	if err != nil {
		if errors.Is(err, ErrNilTrie) {
			return nil, nil
		}
		return nil, err
	}
	return value, nil
}

func saveAlias(account UserAccountHandler, aliasSCAccount UserAccountHandler, aliasAddress []byte, aliasIdentifier core.AddressIdentifier) error {
	err := aliasSCAccount.SaveKeyValue(buildMultiversXAddressKey(aliasAddress, aliasIdentifier), account.AddressBytes())
	if err != nil {
		return err
	}
	return account.SaveKeyValue(buildAliasAddressKey(aliasIdentifier), aliasAddress)
}

func saveMainAddressIdentifier(account UserAccountHandler, mainAddressIdentifier core.AddressIdentifier) error {
	return account.SaveKeyValue(buildMainAddressIdentifierKey(), mainAddressIdentifier.Spread())
}

func fetchMultiversXAddress(aliasSCAccount UserAccountHandler, aliasAddress []byte, aliasIdentifier core.AddressIdentifier) ([]byte, error) {
	return retrieveValueIgnoreNilTrie(aliasSCAccount, buildMultiversXAddressKey(aliasAddress, aliasIdentifier))
}

func fetchAliasAddress(account UserAccountHandler, aliasIdentifier core.AddressIdentifier) ([]byte, error) {
	return retrieveValueIgnoreNilTrie(account, buildAliasAddressKey(aliasIdentifier))
}

func fetchMainAddressIdentifier(account UserAccountHandler) (core.AddressIdentifier, error) {
	mainAddressIdentifier, err := retrieveValueIgnoreNilTrie(account, buildMainAddressIdentifierKey())
	if err != nil {
		return core.InvalidAddressIdentifier, err
	}
	return core.ParseAddressIdentifier(mainAddressIdentifier), nil
}

func fetchOrGenerateMultiversXAddress(aliasSCAccount UserAccountHandler, aliasAddress []byte, aliasIdentifier core.AddressIdentifier) ([]byte, bool, error) {
	multiversXAddress, err := fetchMultiversXAddress(aliasSCAccount, aliasAddress, aliasIdentifier)
	if err != nil {
		return nil, false, err
	}
	if len(multiversXAddress) > 0 {
		return multiversXAddress, false, nil
	}

	if isDefaultSourceAddress(aliasAddress, aliasIdentifier) {
		multiversXAddress, err = requestDefaultAddress(core.MVXAddressIdentifier)
		return multiversXAddress, true, err
	}
	multiversXAddress, err = address.GeneratePseudoAddress(aliasAddress, aliasIdentifier, core.MVXAddressIdentifier)
	return multiversXAddress, true, err
}

func fetchOrGenerateAliasAddress(account UserAccountHandler, mainAddress []byte, mainAddressIdentifier core.AddressIdentifier, aliasIdentifier core.AddressIdentifier) ([]byte, bool, error) {
	aliasAddress, err := fetchAliasAddress(account, aliasIdentifier)
	if err != nil {
		return nil, false, err
	}
	if len(aliasAddress) > 0 {
		return aliasAddress, false, nil
	}

	if isDefaultSourceAddress(mainAddress, mainAddressIdentifier) {
		aliasAddress, err = requestDefaultAddress(aliasIdentifier)
		return aliasAddress, true, err
	}
	aliasAddress, err = address.GeneratePseudoAddress(mainAddress, mainAddressIdentifier, aliasIdentifier)
	return aliasAddress, true, err
}

func fetchOrGenerateMainAddressIdentifier(account UserAccountHandler, sourceIdentifier core.AddressIdentifier) (core.AddressIdentifier, bool, error) {
	mainAddressIdentifier, err := fetchMainAddressIdentifier(account)
	if err != nil {
		return core.InvalidAddressIdentifier, false, err
	}

	switch mainAddressIdentifier {
	case core.InvalidAddressIdentifier:
		return sourceIdentifier, true, nil
	default:
		return mainAddressIdentifier, false, nil
	}
}

func FetchValidMultiversXAddress(aliasSCAccount UserAccountHandler, aliasAddress []byte, aliasIdentifier core.AddressIdentifier) ([]byte, error) {
	multiversXAddress, err := fetchMultiversXAddress(aliasSCAccount, aliasAddress, aliasIdentifier)
	if err != nil {
		return nil, err
	}
	if len(multiversXAddress) > 0 {
		return multiversXAddress, nil
	}
	return nil, ErrInvalidMultiversXAddress
}

func FetchValidAliasAddress(account UserAccountHandler, aliasIdentifier core.AddressIdentifier) ([]byte, error) {
	aliasAddress, err := fetchAliasAddress(account, aliasIdentifier)
	if err != nil {
		return nil, err
	}
	if len(aliasAddress) > 0 {
		return aliasAddress, nil
	}
	return nil, ErrInvalidAliasAddress
}

func FetchValidMainAddressIdentifier(account UserAccountHandler) (core.AddressIdentifier, error) {
	mainAddressIdentifier, err := fetchMainAddressIdentifier(account)
	if err != nil {
		return core.InvalidAddressIdentifier, err
	}
	if mainAddressIdentifier != core.InvalidAddressIdentifier {
		return mainAddressIdentifier, nil
	}
	return core.InvalidAddressIdentifier, ErrInvalidMainAddressIdentifier
}

func RequestSenderAndReceiver(
	accounts AccountsAdapter,
	sender []byte,
	receiver []byte,
	sourceIdentifier core.AddressIdentifier,
	requestedIdentifier core.AddressIdentifier,
	saveOnGenerate bool,
) ([]byte, []byte, error) {
	senderResponse, err := accounts.RequestAddress(&vmcommon.AddressRequest{
		SourceAddress:       sender,
		SourceIdentifier:    sourceIdentifier,
		RequestedIdentifier: requestedIdentifier,
		SaveOnGenerate:      saveOnGenerate,
	})
	if err != nil {
		return nil, nil, err
	}

	receiverResponse, err := accounts.RequestAddress(&vmcommon.AddressRequest{
		SourceAddress:       receiver,
		SourceIdentifier:    sourceIdentifier,
		RequestedIdentifier: requestedIdentifier,
		SaveOnGenerate:      saveOnGenerate,
	})
	if err != nil {
		return nil, nil, err
	}

	return senderResponse.RequestedAddress, receiverResponse.RequestedAddress, nil
}

func GenerateAddressesForIdentifier(
	accounts AccountsAdapter,
	sourceAddresses [][]byte,
	sourceIdentifier core.AddressIdentifier,
	requestedIdentifier core.AddressIdentifier,
) error {
	for _, sourceAddress := range sourceAddresses {
		_, err := accounts.RequestAddress(&vmcommon.AddressRequest{
			SourceAddress:       sourceAddress,
			SourceIdentifier:    sourceIdentifier,
			RequestedIdentifier: requestedIdentifier,
			SaveOnGenerate:      true,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func GenerateAddressesForIdentifiers(
	accounts AccountsAdapter,
	sourceAddresses [][]byte,
	sourceIdentifier core.AddressIdentifier,
	requestedIdentifiers []core.AddressIdentifier,
) error {
	for _, requestedIdentifier := range requestedIdentifiers {
		if requestedIdentifier == sourceIdentifier {
			continue
		}

		err := GenerateAddressesForIdentifier(accounts, sourceAddresses, sourceIdentifier, requestedIdentifier)
		if err != nil {
			return err
		}
	}
	return nil
}

func GenerateAddresses(
	accounts AccountsAdapter,
	sourceAddresses [][]byte,
	sourceIdentifier core.AddressIdentifier,
) error {
	requestedIdentifiers := []core.AddressIdentifier{core.MVXAddressIdentifier, core.ETHAddressIdentifier}
	return GenerateAddressesForIdentifiers(accounts, sourceAddresses, sourceIdentifier, requestedIdentifiers)
}
