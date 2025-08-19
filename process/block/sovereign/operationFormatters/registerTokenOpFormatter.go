package operationFormatters

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	sovData "github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/multiversx/mx-chain-go/common"
)

const (
	topicIdxTokenID     = 1
	topicIdxTokenType   = 2
	topicIdxName        = 3
	topicIdxTicker      = 4
	topicIdxNumDecimals = 5
)

type registerTokenOpFormatter struct {
	dataCodec DataCodecHandler
}

type TokenProperties struct {
	TokenIdentifier []byte
	TokenType       core.ESDTType
	Name            []byte
	Ticker          []byte
	NumDecimals     uint64
	EventData       *sovData.EventData
}

func (op *registerTokenOpFormatter) CreateOperationData(event data.EventHandler, evData *sovData.EventData) ([]byte, error) {
	tokenProperties, err := op.createTokenProperties(event.GetTopics(), evData)
	if err != nil {
		return nil, err
	}

	_ = tokenProperties
	return nil, nil
	//return op.dataCodec.Se(nil)
}

func (op *registerTokenOpFormatter) createTokenProperties(topics [][]byte, eventData *sovData.EventData) (*TokenProperties, error) {
	tokenType, err := common.ByteSliceToUint64(topics[topicIdxTokenType])
	if err != nil {
		return nil, err
	}

	numDecimals, err := common.ByteSliceToUint64(topics[topicIdxNumDecimals])
	if err != nil {
		return nil, err
	}

	return &TokenProperties{
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
