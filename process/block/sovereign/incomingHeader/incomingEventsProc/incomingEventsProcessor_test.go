package incomingEventsProc

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/process/block/sovereign/incomingHeader/dto"
)

func TestIncomingEventsProcessor_RegisterProcessor(t *testing.T) {
	eventsProc := NewIncomingEventsProcessor()
	err := eventsProc.RegisterProcessor(dto.EventIDDepositIncomingTransfer, nil)
	require.Error(t, errNilIncomingEventHandler, err)
}
