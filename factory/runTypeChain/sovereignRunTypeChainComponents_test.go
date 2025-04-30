package runTypeChain_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/factory/runTypeChain"
)

func TestSovereignRunTypeChainComponentsFactory_CreateAndClose(t *testing.T) {
	t.Parallel()

	srtccf := runTypeChain.NewSovereignRunTypeChainComponentsFactory()
	require.False(t, srtccf.IsInterfaceNil())

	srtcc := srtccf.Create()
	require.False(t, srtcc.IsInterfaceNil())

	require.NoError(t, srtcc.Close())
}
