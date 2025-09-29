package clob

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestErrors(t *testing.T) {
	require.Equal(t, "invalid order type", ErrInvalidOrderType.Error())
	require.Equal(t, "order not found", ErrOrderNotFound.Error())
	require.Equal(t, "order book is nil", ErrNilOrderBook.Error())
}