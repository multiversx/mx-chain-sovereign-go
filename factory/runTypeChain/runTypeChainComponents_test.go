package runTypeChain_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/factory/runTypeChain"
)

func TestNewRunTypeChainComponentsFactory(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		rtccf := runTypeChain.NewRunTypeChainComponentsFactory()
		require.False(t, rtccf.IsInterfaceNil())
	})
}

func TestRunTypeChainComponentsFactory_CreateAndClose(t *testing.T) {
	t.Parallel()

	rtccf := runTypeChain.NewRunTypeChainComponentsFactory()
	require.False(t, rtccf.IsInterfaceNil())

	rtcc := rtccf.Create()
	require.False(t, rtcc.IsInterfaceNil())

	require.NoError(t, rtcc.Close())
}
