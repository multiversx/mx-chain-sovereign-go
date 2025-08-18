package operationFormatters

import (
	"github.com/multiversx/mx-chain-core-go/data"
	sovereign2 "github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/multiversx/mx-chain-go/common"
)

const (
	numTransferTopics = 3
	tokensIndex       = 2
	receiverIndex     = 1
)

type depositOpFormatter struct {
	dataCodec DataCodecHandler
}

func NewDepositOpFormatter(dataCodec DataCodecHandler) (*depositOpFormatter, error) {
	return &depositOpFormatter{
		dataCodec: dataCodec,
	}, nil
}

func (op *depositOpFormatter) CreateOperationData(event data.EventHandler, evData *sovereign2.EventData) ([]byte, error) {
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

func (op *depositOpFormatter) createOperationData(topics [][]byte, eventData *sovereign2.EventData) (*sovereign2.Operation, error) {
	tokens := make([]sovereign2.EsdtToken, 0)
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

		payment := sovereign2.EsdtToken{
			Identifier: tokenIdentifier,
			Nonce:      tokenNonce,
			Data:       *tokenData,
		}
		tokens = append(tokens, payment)
	}

	return &sovereign2.Operation{
		Address: topics[receiverIndex],
		Tokens:  tokens,
		Data:    eventData,
	}, nil
}
