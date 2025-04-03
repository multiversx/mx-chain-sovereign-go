package dto

import (
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
)

const (
	// MinTopicsInTransferEvent represents the minimum number of topics required in a transfer event.
	MinTopicsInTransferEvent = 2

	// NumTransferTopics defines the expected number of topics related to a token transfer.
	NumTransferTopics = 3

	// TokensIndex is the index position in the topics array where token-related data is stored.
	TokensIndex = 2
)

const (
	// EventIDExecutedOutGoingBridgeOp identifies an event related to the execution of an outgoing bridge operation.
	EventIDExecutedOutGoingBridgeOp = "execute"

	// EventIDDepositIncomingTransfer identifies an event related to an incoming token deposit.
	EventIDDepositIncomingTransfer = "deposit"

	// EventIDChangeValidatorSet identifies an event related to validator set changes.
	EventIDChangeValidatorSet = "changeValidatorSet"

	// TopicIDConfirmedOutGoingOperation is used as a topic identifier for confirmed outgoing bridge operations.
	TopicIDConfirmedOutGoingOperation = "executedBridgeOp"

	// TopicIDDepositIncomingTransfer is used as a topic identifier for incoming token deposit events.
	TopicIDDepositIncomingTransfer = "deposit"
)

// SCRInfo holds an incoming scr that is created based on an incoming cross chain event and its hash
type SCRInfo struct {
	SCR  *smartContractResult.SmartContractResult
	Hash []byte
}

// ConfirmedBridgeOp holds the hashes for a bridge operations that are confirmed from the main chain
type ConfirmedBridgeOp struct {
	HashOfHashes []byte
	Hash         []byte
}

// EventResult holds the result of processing an incoming cross chain event
type EventResult struct {
	SCR               *SCRInfo
	ConfirmedBridgeOp *ConfirmedBridgeOp
}

// EventsResult holds the results of processing incoming cross chain events
type EventsResult struct {
	Scrs               []*SCRInfo
	ConfirmedBridgeOps []*ConfirmedBridgeOp
}
