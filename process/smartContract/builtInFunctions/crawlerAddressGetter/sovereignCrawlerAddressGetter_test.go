package crawlerAddressGetter

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"
)

func TestSovereignCrawlerAddressGetter_GetAllowedAddress(t *testing.T) {
	t.Parallel()

	addrGetter := NewSovereignCrawlerAddressGetter()
	require.False(t, addrGetter.IsInterfaceNil())

	address, err := addrGetter.GetAllowedAddress(nil, nil)
	require.Nil(t, err)
	require.Equal(t, core.SystemAccountAddress, address)
}
