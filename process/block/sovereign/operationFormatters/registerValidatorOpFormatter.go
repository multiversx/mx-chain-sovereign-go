package operationFormatters

import (
	"bytes"
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	errMx "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/sovereign/dto"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/vm"
)

const (
	numExpectedTopicsInRegisterNewValidator = 2
	topicIdxBlsKey                          = 0
	topicIdxOwner                           = 1
)

type registerValidatorOpFormatter struct {
	peerAccountsDB state.AccountsAdapter
	dataCodec      DataCodecHandler
}

// NewRegisterValidatorOpFormatter will create a register/unregister validator op formatter
func NewRegisterValidatorOpFormatter(
	peerAccountsDB state.AccountsAdapter,
	dataCodec DataCodecHandler,
) (*registerValidatorOpFormatter, error) {
	if check.IfNil(peerAccountsDB) {
		return nil, errMx.ErrNilPeerAccounts
	}
	if check.IfNil(dataCodec) {
		return nil, errMx.ErrNilDataCodec
	}

	return &registerValidatorOpFormatter{
		peerAccountsDB: peerAccountsDB,
		dataCodec:      dataCodec,
	}, nil
}

// CreateOperationData creates a register/unregister new validator operation data
func (op *registerValidatorOpFormatter) CreateOperationData(event data.EventHandler) ([]byte, error) {
	numTopics := len(event.GetTopics())
	if numTopics != numExpectedTopicsInRegisterNewValidator {
		return nil, fmt.Errorf("%w, expected: %d, received: %d", errInvalidNumTopicsInRegisterValidator, numExpectedTopicsInRegisterNewValidator, numTopics)
	}

	if !bytes.Equal(event.GetAddress(), vm.StakingSCAddress) {
		return nil, fmt.Errorf("%w in registerValidatorOpFormatter, expected StakingSCAddress", vm.ErrInvalidAddress)
	}

	peerAcc, err := process.GetPeerAccount(event.GetTopics()[topicIdxBlsKey], op.peerAccountsDB)
	if err != nil {
		return nil, err
	}

	return op.dataCodec.SerializeNewlyRegisteredKey(dto.RegisteredBlsKey{
		ID:    peerAcc.GetMainChainID(),
		Key:   peerAcc.GetBLSPublicKey(),
		Owner: event.GetTopics()[topicIdxOwner],
	})
}

// IsInterfaceNil checks if the underlying pointer is nil
func (op *registerValidatorOpFormatter) IsInterfaceNil() bool {
	return op == nil
}
