package operationFormatters

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	sovData "github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/multiversx/mx-chain-go/common"
	errMx "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process/block/sovereign/dto"
)

const (
	numExpectedTopicsInRegisterToken = 6

	topicIdxTokenID     = 1
	topicIdxTokenType   = 2
	topicIdxName        = 3
	topicIdxTicker      = 4
	topicIdxNumDecimals = 5
)

type registerTokenOpFormatter struct {
	dataCodec DataCodecHandler
}

// NewRegisterTokenOpFormatter will create a register token op formatter
func NewRegisterTokenOpFormatter(dataCodec DataCodecHandler, topicsChecker TopicsCheckerHandler) (*registerTokenOpFormatter, error) {
	if check.IfNil(dataCodec) {
		return nil, errMx.ErrNilDataCodec
	}
	if check.IfNil(topicsChecker) {
		return nil, errMx.ErrNilTopicsChecker
	}

	return &registerTokenOpFormatter{
		dataCodec: dataCodec,
	}, nil
}

// CreateOperationData will create register token operation data
func (op *registerTokenOpFormatter) CreateOperationData(event data.EventHandler) ([]byte, error) {
	evData, err := op.dataCodec.DeserializeEventData(event.GetData())
	if err != nil {
		return nil, err
	}

	tokenProperties, err := op.createTokenProperties(event.GetTopics(), evData)
	if err != nil {
		return nil, err
	}

	return op.dataCodec.SerializeTokenProperties(*tokenProperties)
}

func (op *registerTokenOpFormatter) createTokenProperties(topics [][]byte, eventData *sovData.EventData) (*dto.TokenProperties, error) {
	numTopics := len(topics)
	if numTopics != numExpectedTopicsInRegisterToken {
		return nil, fmt.Errorf("%w, expected: %d, received: %d", errInvalidNumTopicsInRegisterToken, numExpectedTopicsInRegisterToken, numTopics)
	}

	tokenType, err := common.ByteSliceToUint64(topics[topicIdxTokenType])
	if err != nil {
		return nil, err
	}

	numDecimals, err := common.ByteSliceToUint64(topics[topicIdxNumDecimals])
	if err != nil {
		return nil, err
	}

	return &dto.TokenProperties{
		TokenIdentifier: topics[topicIdxTokenID],
		TokenType:       core.ESDTType(tokenType),
		Name:            topics[topicIdxName],
		Ticker:          topics[topicIdxTicker],
		NumDecimals:     numDecimals,
		EventData:       eventData,
	}, nil
}

// IsInterfaceNil checks if the underlying pointer is nil
func (op *registerTokenOpFormatter) IsInterfaceNil() bool {
	return op == nil
}
