package incomingHeader

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/data/sovereign"

	"github.com/multiversx/mx-chain-go/process/block/sovereign/incomingHeader/dto"
)

type topicsChecker struct{}

// NewTopicsChecker creates a topics checker which is able to validate topics
func NewTopicsChecker() *topicsChecker {
	return &topicsChecker{}
}

// CheckDepositTokensValidity will receive the deposit topics and validate them
func (tc *topicsChecker) CheckDepositTokensValidity(topics [][]byte) error {
	// TODO: Check each param validity (e.g. check that topic[0] == valid address)
	if len(topics) < dto.MinTopicsInTransferEvent || len(topics[2:])%dto.NumTransferTopics != 0 {
		log.Error("topicsChecker.CheckDepositTokensValidity",
			"error", dto.ErrInvalidNumTopicsInEvent,
			"num topics", len(topics),
			"topics", topics)

		return fmt.Errorf("%w for deposit event; num topics = %d", dto.ErrInvalidNumTopicsInEvent, len(topics))
	}

	return nil
}

// CheckScCallValidity will receive the topics and transfer data, and validate them
func (tc *topicsChecker) CheckScCallValidity(topics [][]byte, transferData *sovereign.TransferData) error {
	// TODO: Check each param validity (e.g. check that topic[0] == valid address, valid transferData)
	if len(topics) != dto.NumScCallTopics || transferData == nil {
		log.Error("topicsChecker.CheckScCallValidity",
			"error", dto.ErrInvalidNumTopicsInEvent,
			"num topics", len(topics),
			"topics", topics,
			"transferData is nil", transferData == nil)

		return fmt.Errorf("%w for sc call event; num topics = %d", dto.ErrInvalidNumTopicsInEvent, len(topics))
	}

	return nil
}

// CheckValidity will receive the topics and validate them
func (tc *topicsChecker) CheckValidity(topics [][]byte, transferData *sovereign.TransferData) error {
	err := tc.CheckDepositTokensValidity(topics)
	if err == nil {
		return nil
	}

	err = tc.CheckScCallValidity(topics, transferData)
	if err == nil {
		return nil
	}

	log.Error("topicsChecker.CheckValidity",
		"error", dto.ErrInvalidNumTopicsInEvent,
		"num topics", len(topics),
		"topics", topics,
		"transferData is nil", transferData == nil)

	return fmt.Errorf("%w for outgoing deposit operation; num topics = %d", dto.ErrInvalidNumTopicsInEvent, len(topics))
}

// IsInterfaceNil checks if the underlying pointer is nil
func (tc *topicsChecker) IsInterfaceNil() bool {
	return tc == nil
}
