package disabled

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/stretchr/testify/require"
)

func TestTopicsChecker_MethodsShouldNotPanic(t *testing.T) {
	t.Parallel()

	tc := NewDisabledTopicsChecker()
	require.False(t, check.IfNil(tc))

	require.NotPanics(t, func() {
		err := tc.CheckValidity([][]byte{[]byte("topic")}, nil)
		require.NoError(t, err)

		err = tc.CheckValidity([][]byte{[]byte("topic")}, &sovereign.TransferData{})
		require.NoError(t, err)
	})
}
