package disabled

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/sovereign/dto"
	"github.com/stretchr/testify/require"
)

func TestDisabledOutGoingOperationsPool(t *testing.T) {
	t.Parallel()

	disabledComp := &DisabledOutGoingChainNonce{}

	require.NotPanics(t, func() {
		disabledComp.SetNonce(dto.MVX, 1)

		nonce, err := disabledComp.GetAndIncrementNonce(dto.MVX)
		require.NoError(t, err)
		require.Zero(t, nonce)

		nonce = disabledComp.GetNonce(dto.MVX)
		require.Zero(t, nonce)

		require.False(t, disabledComp.IsInterfaceNil())
	})
}
