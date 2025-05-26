package runType

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common/factory"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/state"
)

var pubKeyConfig = config.PubkeyConfig{
	Hrp:    "erd",
	Type:   "bech32",
	Length: 32,
}
var pubKeyConverter, _ = factory.NewPubkeyConverter(pubKeyConfig)
var address = "erd1qyu5wthldzr8wx5c9ucg8kjagg0jfs53s8nr3zpz3hypefsdd8ssycr6th"
var addressBytes, _ = pubKeyConverter.Decode(address)
var whiteListedAddrs = []string{address}

func TestSetWhiteListedAddresses(t *testing.T) {
	t.Run("invalid pubKey config, should error", func(t *testing.T) {
		invalidPubKeyConfig := config.PubkeyConfig{
			Hrp:    "erd",
			Type:   "invalid",
			Length: 32,
		}
		err := SetWhiteListedAddresses(whiteListedAddrs, invalidPubKeyConfig)
		require.Error(t, err, state.ErrInvalidPubkeyConverterType)
	})
	t.Run("invalid address, should error", func(t *testing.T) {
		addressList := []string{
			"erd1qyu5wthldzr8wx5c9ucg8kjagg0jfs53s8nr3zpz3hypefsdd8ssycr6te",
		}
		err := SetWhiteListedAddresses(addressList, pubKeyConfig)
		require.Error(t, err)
	})
	t.Run("set whitelisted addresses, should work", func(t *testing.T) {
		err := SetWhiteListedAddresses(whiteListedAddrs, pubKeyConfig)
		require.NoError(t, err)
	})
}

func TestIsWhiteListedAddress(t *testing.T) {
	t.Run("address not found", func(t *testing.T) {
		addressList := []string{
			"erd1wc3uh22g2aved3qeehkz9kzgrjwxhg9mkkxp2ee7jj7ph34p2csq0n2y5x",
		}
		err := SetWhiteListedAddresses(addressList, pubKeyConfig)
		require.NoError(t, err)

		isFound := IsAddressWhiteListed(addressBytes)
		require.False(t, isFound)
	})

	t.Run("address is found", func(t *testing.T) {
		err := SetWhiteListedAddresses(whiteListedAddrs, pubKeyConfig)
		require.NoError(t, err)

		isFound := IsAddressWhiteListed(addressBytes)
		require.True(t, isFound)
	})
}
