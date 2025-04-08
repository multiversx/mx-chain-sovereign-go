package incomingEventsProc

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewEventProcDepositOperation(t *testing.T) {
	t.Parallel()

	t.Run("nil deposit event proc, should return error", func(t *testing.T) {
		handler, err := NewEventProcDepositOperation(nil, &eventProcScCall{})
		require.Equal(t, errNilEventProcDepositTokens, err)
		require.Nil(t, handler)
	})

	t.Run("nil sc call event proc, should return error", func(t *testing.T) {
		handler, err := NewEventProcDepositOperation(&eventProcDepositTokens{}, nil)
		require.Equal(t, errNilEventProcScCall, err)
		require.Nil(t, handler)
	})

	t.Run("should work", func(t *testing.T) {
		handler, err := NewEventProcDepositOperation(&eventProcDepositTokens{}, &eventProcScCall{&eventProcDepositTokens{}})
		require.NotNil(t, handler)
		require.Nil(t, err)
	})
}
