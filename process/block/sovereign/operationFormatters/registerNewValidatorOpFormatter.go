package operationFormatters

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/multiversx/mx-chain-core-go/data"
	sovData "github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/vm"
)

const (
	numExpectedTopicsInRegisterNewValidator = 1
	topicIdxBlsKey                          = 0
)

type registerNewValidatorOpFormatter struct {
	peerAccountsDB state.AccountsAdapter
}

// CreateOperationData creates a register new validator operation data
func (op *registerNewValidatorOpFormatter) CreateOperationData(event data.EventHandler, _ *sovData.EventData) ([]byte, error) {
	numTopics := len(event.GetTopics())
	if numTopics != numExpectedTopicsInRegisterNewValidator {
		return nil, fmt.Errorf("%w, expected: %d, received: %d", errInvalidNumTopicsInRegisterValidator, numExpectedTopicsInRegisterNewValidator, numTopics)
	}

	if !bytes.Equal(event.GetAddress(), vm.StakingSCAddress) {
		return nil, fmt.Errorf("%w in registerNewValidatorOpFormatter, expected StakingSCAddress", vm.ErrInvalidAddress)
	}

	peerAcc, err := op.getPeerAccount(event.GetTopics()[topicIdxBlsKey])
	if err != nil {
		return nil, err
	}

	return "@" + hex.EncodeToString(peerAcc.GetBLSPublicKey()) + "@" + hex.EncodeToString(peerAcc.GetMainChainID())
}

// todo: Here do not duplicate
func (op *registerNewValidatorOpFormatter) getPeerAccount(key []byte) (state.PeerAccountHandler, error) {
	account, err := op.peerAccountsDB.LoadAccount(key)
	if err != nil {
		return nil, err
	}

	peerAcc, ok := account.(state.PeerAccountHandler)
	if !ok {
		return nil, epochStart.ErrWrongTypeAssertion
	}

	return peerAcc, nil
}

// IsInterfaceNil checks if the underlying pointer is nil
func (op *registerNewValidatorOpFormatter) IsInterfaceNil() bool {
	return op == nil
}
