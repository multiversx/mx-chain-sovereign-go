package operationFormatters

import (
	"bytes"
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	dtoCore "github.com/multiversx/mx-chain-core-go/data/sovereign/dto"
	errMx "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/sovereign/dto"
	dtoSov "github.com/multiversx/mx-chain-go/process/block/sovereign/incomingHeader/dto"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/vm"
)

const (
	numExpectedTopicsInRegisterValidator = 2
	topicIdxBlsKey                       = 0
	topicIdxOwner                        = 1
)

type registerValidatorOpFormatter struct {
	peerAccountsDB    state.AccountsAdapter
	dataCodec         DataCodecHandler
	subscribedChains  []dtoCore.ChainID
	chainNonceHandler dtoSov.OutGoingOpNonceChainHandler
}

// NewRegisterValidatorOpFormatter will create a register/unregister validator op formatter
func NewRegisterValidatorOpFormatter(
	peerAccountsDB state.AccountsAdapter,
	dataCodec DataCodecHandler,
	chainNonceHandler dtoSov.OutGoingOpNonceChainHandler,
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
		// TODO: Marius C. : MX-17260 Use ordered chains here
		subscribedChains:  []dtoCore.ChainID{dtoCore.MVX},
		chainNonceHandler: chainNonceHandler,
	}, nil
}

// CreateOperationData creates a register/unregister new validator operation data
func (op *registerValidatorOpFormatter) CreateOperationData(event data.EventHandler) (map[dtoCore.ChainID][]byte, error) {
	numTopics := len(event.GetTopics())
	if numTopics != numExpectedTopicsInRegisterValidator {
		return nil, fmt.Errorf("%w, expected: %d, received: %d", errInvalidNumTopicsInRegisterValidator, numExpectedTopicsInRegisterValidator, numTopics)
	}

	if !bytes.Equal(event.GetAddress(), vm.StakingSCAddress) {
		return nil, fmt.Errorf("%w in registerValidatorOpFormatter, expected StakingSCAddress", vm.ErrInvalidAddress)
	}

	peerAcc, err := process.GetPeerAccount(event.GetTopics()[topicIdxBlsKey], op.peerAccountsDB)
	if err != nil {
		return nil, err
	}

	ret := map[dtoCore.ChainID][]byte{}
	for _, chainID := range op.subscribedChains {
		chainNonce, err := op.chainNonceHandler.GetNonce(chainID)
		if err != nil {
			return nil, err
		}

		regKeyData, err := op.dataCodec.SerializeNewlyRegisteredKey(dto.RegisteredBlsKey{
			ID:    peerAcc.GetMainChainID(),
			Key:   peerAcc.GetBLSPublicKey(),
			Owner: event.GetTopics()[topicIdxOwner],
			Nonce: chainNonce,
		})
		if err != nil {
			return nil, err
		}

		ret[chainID] = regKeyData
	}

	return ret, nil
}

// IsInterfaceNil checks if the underlying pointer is nil
func (op *registerValidatorOpFormatter) IsInterfaceNil() bool {
	return op == nil
}
