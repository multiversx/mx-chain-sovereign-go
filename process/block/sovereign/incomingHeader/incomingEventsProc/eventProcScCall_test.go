package incomingEventsProc

import (
	"encoding/hex"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/stretchr/testify/require"

	errorsMx "github.com/multiversx/mx-chain-go/errors"
)

func TestNewEventProcScCall(t *testing.T) {
	t.Parallel()

	t.Run("nil marshaller, should return error", func(t *testing.T) {
		args := createArgs()
		args.Marshaller = nil

		handler, err := NewEventProcScCall(args)
		require.Equal(t, core.ErrNilMarshalizer, err)
		require.Nil(t, handler)
	})

	t.Run("nil hasher, should return error", func(t *testing.T) {
		args := createArgs()
		args.Hasher = nil

		handler, err := NewEventProcScCall(args)
		require.Equal(t, core.ErrNilHasher, err)
		require.Nil(t, handler)
	})

	t.Run("nil data codec, should return error", func(t *testing.T) {
		args := createArgs()
		args.DataCodec = nil

		handler, err := NewEventProcScCall(args)
		require.Equal(t, errorsMx.ErrNilDataCodec, err)
		require.Nil(t, handler)
	})

	t.Run("nil topics checker, should return error", func(t *testing.T) {
		args := createArgs()
		args.TopicsChecker = nil

		handler, err := NewEventProcScCall(args)
		require.Equal(t, errorsMx.ErrNilTopicsChecker, err)
		require.Nil(t, handler)
	})

	t.Run("should work", func(t *testing.T) {
		args := createArgs()
		handler, err := NewEventProcScCall(args)
		require.NotNil(t, handler)
		require.Nil(t, err)
	})
}

func TestScCallEventProc_createSCRData(t *testing.T) {
	t.Parallel()

	transferGas := uint64(1)
	func1 := []byte("func1")
	arg1 := []byte("arg1")
	arg2 := []byte("arg2")

	args := createArgs()
	handler, _ := NewEventProcScCall(args)

	eventData := &sovereign.EventData{
		TransferData: &sovereign.TransferData{
			GasLimit: transferGas,
			Function: func1,
			Args:     [][]byte{arg1, arg2},
		},
	}

	scrData, gasLimit := handler.createSCRData(eventData)
	require.Equal(t, transferGas, gasLimit)

	expectedSCR := func1
	expectedSCR = append(expectedSCR, "@"+hex.EncodeToString(arg1)...)
	expectedSCR = append(expectedSCR, "@"+hex.EncodeToString(arg2)...)
	require.Equal(t, expectedSCR, scrData)
}
