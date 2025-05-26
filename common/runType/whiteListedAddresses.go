package runType

import (
	"github.com/multiversx/mx-chain-go/common/factory"
	"github.com/multiversx/mx-chain-go/config"
)

var whiteListedAddresses map[string]struct{}

func init() {
	whiteListedAddresses = make(map[string]struct{})
}

// SetWhiteListedAddresses will set the whitelisted whiteListedAddresses bytes
func SetWhiteListedAddresses(bech32Addresses []string, pubKeyConverter config.PubkeyConfig) error {
	addressPubKeyConverter, err := factory.NewPubkeyConverter(pubKeyConverter)
	if err != nil {
		return err
	}

	whiteListedAddresses = make(map[string]struct{}, len(bech32Addresses))

	for _, addr := range bech32Addresses {
		addrBytes, errDecode := addressPubKeyConverter.Decode(addr)
		if errDecode != nil {
			return errDecode
		}
		whiteListedAddresses[string(addrBytes)] = struct{}{}
	}

	return nil
}

// IsAddressWhiteListed checks if the address bytes is whitelisted
func IsAddressWhiteListed(addressBytes []byte) bool {
	_, exists := whiteListedAddresses[string(addressBytes)]
	return exists
}
