package operationFormatters

import (
	"testing"

	errMx "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/testscommon/sovereign"
	"github.com/stretchr/testify/require"
)

func TestNewDepositOpFormatter(t *testing.T) {
	t.Parallel()

	t.Run("nil data codec, should return error", func(t *testing.T) {
		opFormatter, err := NewDepositOpFormatter(nil)
		require.Nil(t, opFormatter)
		require.Equal(t, errMx.ErrNilDataCodec, err)
	})
	t.Run("should work", func(t *testing.T) {
		opFormatter, err := NewDepositOpFormatter(&sovereign.DataCodecMock{})
		require.Nil(t, err)
		require.False(t, opFormatter.IsInterfaceNil())
	})
}
