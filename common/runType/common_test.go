package runType

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUIntToBytes(t *testing.T) {
	t.Parallel()

	require.Equal(t, []byte{0x0}, UIntToBytes(0))
	require.Equal(t, []byte{0x0a}, UIntToBytes(10))
}
