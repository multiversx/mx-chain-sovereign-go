package operationFormatters

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	sovData "github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/multiversx/mx-chain-go/common"
	errMx "github.com/multiversx/mx-chain-go/errors"
)

const (
	numTransferTopics = 3
	tokensIndex       = 2
	receiverIndex     = 1
)

type depositOpFormatter struct {
	*opFormatterHelper
	dataCodec DataCodecHandler
}

// NewDepositOpFormatter creates a new deposit token operation formatter
func NewDepositOpFormatter(dataCodec DataCodecHandler, topicsChecker TopicsCheckerHandler) (*depositOpFormatter, error) {
	if check.IfNil(dataCodec) {
		return nil, errMx.ErrNilDataCodec
	}
	if check.IfNil(topicsChecker) {
		return nil, errMx.ErrNilTopicsChecker
	}

	return &depositOpFormatter{
		opFormatterHelper: &opFormatterHelper{
			dataCodec:     dataCodec,
			topicsChecker: topicsChecker,
		},
		dataCodec: dataCodec,
	}, nil
}

// CreateOperationData creates a deposit token operation data bytes
func (op *depositOpFormatter) CreateOperationData(event data.EventHandler) ([]byte, error) {
	evData, err := op.checkAndGetEventData(event)
	if err != nil {
		return nil, err
	}

	operation, err := op.createOperationData(event.GetTopics(), evData)
	if err != nil {
		return nil, err
	}

	operationBytes, err := op.dataCodec.SerializeOperation(*operation)
	if err != nil {
		return nil, err
	}

	return operationBytes, nil
}

func (op *depositOpFormatter) createOperationData(topics [][]byte, eventData *sovData.EventData) (*sovData.Operation, error) {
	tokens := make([]sovData.EsdtToken, 0)
	for i := tokensIndex; i < len(topics); i += numTransferTopics {
		tokenIdentifier := topics[i]
		tokenNonce, err := common.ByteSliceToUint64(topics[i+1])
		if err != nil {
			return nil, err
		}
		tokenData, err := op.dataCodec.DeserializeTokenData(topics[i+2])
		if err != nil {
			return nil, err
		}

		payment := sovData.EsdtToken{
			Identifier: tokenIdentifier,
			Nonce:      tokenNonce,
			Data:       *tokenData,
		}
		tokens = append(tokens, payment)
	}

	return &sovData.Operation{
		Address: topics[receiverIndex],
		Tokens:  tokens,
		Data:    eventData,
	}, nil
}

// IsInterfaceNil checks if the underlying pointer is nil
func (op *depositOpFormatter) IsInterfaceNil() bool {
	return op == nil
}
