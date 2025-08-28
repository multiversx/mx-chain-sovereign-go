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
	numExpectedTopicsInRegisterNewValidator = 1
	topicIdxBlsKey                          = 0
)

type registerNewValidatorOpFormatter struct {
	peerAccountsDB state.AccountsAdapter
	dataCodec      DataCodecHandler
}

func NewRegisterValidatorOpFormatter(
	peerAccountsDB state.AccountsAdapter,
	dataCodec DataCodecHandler,
) (*registerNewValidatorOpFormatter, error) {
	if check.IfNil(peerAccountsDB) {
		return nil, errMx.ErrNilPeerAccounts
	}
	if check.IfNil(dataCodec) {
		return nil, errMx.ErrNilDataCodec
	}

	return &registerNewValidatorOpFormatter{
		peerAccountsDB: peerAccountsDB,
		dataCodec:      dataCodec,
	}, nil
}

// CreateOperationData creates a register new validator operation data
func (op *registerNewValidatorOpFormatter) CreateOperationData(event data.EventHandler) ([]byte, error) {
	numTopics := len(event.GetTopics())
	if numTopics != numExpectedTopicsInRegisterNewValidator {
		return nil, fmt.Errorf("%w, expected: %d, received: %d", errInvalidNumTopicsInRegisterValidator, numExpectedTopicsInRegisterNewValidator, numTopics)
	}

	if !bytes.Equal(event.GetAddress(), vm.StakingSCAddress) {
		return nil, fmt.Errorf("%w in registerNewValidatorOpFormatter, expected StakingSCAddress", vm.ErrInvalidAddress)
	}

	peerAcc, err := process.GetPeerAccount(event.GetTopics()[topicIdxBlsKey], op.peerAccountsDB)
	if err != nil {
		return nil, err
	}

	return op.dataCodec.SerializeNewlyRegisteredKey(dto.RegisteredBlsKey{
		ID:  peerAcc.GetMainChainID(),
		Key: peerAcc.GetBLSPublicKey(),
	})
}

// IsInterfaceNil checks if the underlying pointer is nil
func (op *registerNewValidatorOpFormatter) IsInterfaceNil() bool {
	return op == nil
}
