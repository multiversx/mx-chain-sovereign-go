package incomingEventsProc

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"

	"github.com/multiversx/mx-chain-go/process/block/sovereign/incomingHeader/dto"
)

type eventProcDepositOperation struct {
	eventProcDepositTokens IncomingEventHandler
	eventProcScCall        IncomingEventHandler
}

// NewEventProcDepositOperation creates a new event processor for deposit operations
func NewEventProcDepositOperation(
	eventProcDepositTokens IncomingEventHandler,
	eventProcScCall IncomingEventHandler,
) (*eventProcDepositOperation, error) {
	if check.IfNil(eventProcDepositTokens) {
		return nil, errNilEventProcDepositTokens
	}
	if check.IfNil(eventProcScCall) {
		return nil, errNilEventProcScCall
	}

	return &eventProcDepositOperation{
		eventProcDepositTokens: eventProcDepositTokens,
		eventProcScCall:        eventProcScCall,
	}, nil
}

// ProcessEvent handles deposit events
// Each event is identified by dto.EventIDDepositIncomingTransfer.
//
// # An incoming deposit operation can be a deposit or a SC call
//
// Expected event topics ([][]byte):
// - topic[0] = dto.TopicIDDepositIncomingTransfer → Indicates a deposit operation.
//   - The event is treated as an **incoming deposit event**, and the tokens will be added to balance.
//   - The remaining topic fields follow the format defined in `eventProcDepositTokens.go`.
//
// - topic[0] = dto.TopicIDScCall → Indicates a SC call.
//   - The event is treated as an **incoming sc call event**, and no tokens are transferred, only transfer data.
func (ep *eventProcDepositOperation) ProcessEvent(event data.EventHandler) (*dto.EventResult, error) {
	topics := event.GetTopics()
	if len(topics) == 0 {
		return nil, fmt.Errorf("%w for event id: %s", dto.ErrInvalidNumTopicsInEvent, dto.EventIDDepositIncomingTransfer)
	}

	switch string(topics[0]) {
	case dto.TopicIDDepositIncomingTransfer:
		return ep.eventProcDepositTokens.ProcessEvent(event)
	case dto.TopicIDScCall:
		return ep.eventProcScCall.ProcessEvent(event)
	default:
		return nil, dto.ErrInvalidIncomingTopicIdentifier
	}
}

// IsInterfaceNil checks if the underlying pointer is nil
func (ep *eventProcDepositOperation) IsInterfaceNil() bool {
	return ep == nil
}
