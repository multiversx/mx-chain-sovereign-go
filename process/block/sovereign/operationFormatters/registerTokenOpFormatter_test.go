package operationFormatters

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	sovData "github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	errMx "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process/block/sovereign/dto"
	"github.com/multiversx/mx-chain-go/testscommon/sovereign"
	"github.com/stretchr/testify/require"
)

func TestNewRegisterTokenOpFormatter(t *testing.T) {
	t.Parallel()

	t.Run("nil data codec, should return error", func(t *testing.T) {
		opFormatter, err := NewRegisterTokenOpFormatter(nil, &sovereign.TopicsCheckerMock{})
		require.Nil(t, opFormatter)
		require.Equal(t, errMx.ErrNilDataCodec, err)
	})
	t.Run("nil topics checker, should return error", func(t *testing.T) {
		opFormatter, err := NewRegisterTokenOpFormatter(&sovereign.DataCodecMock{}, nil)
		require.Nil(t, opFormatter)
		require.Equal(t, errMx.ErrNilTopicsChecker, err)
	})
	t.Run("should work", func(t *testing.T) {
		opFormatter, err := NewRegisterTokenOpFormatter(&sovereign.DataCodecMock{}, &sovereign.TopicsCheckerMock{})
		require.Nil(t, err)
		require.False(t, opFormatter.IsInterfaceNil())
	})
}

func TestRegisterTokenOpFormatter_CreateOperationData(t *testing.T) {
	t.Parallel()

	eventData := &sovData.EventData{
		Nonce: 4,
	}
	topics := [][]byte{
		[]byte("registerToken"),
		[]byte("tokenID"),
		{byte(core.NonFungible)},
		[]byte("name"),
		[]byte("ticker"),
		{18},
	}

	txEvent := &transaction.Event{
		Topics: topics,
		Data:   []byte("dataEvent"),
	}

	serializedData := []byte("serialized token data")
	dataCodec := &sovereign.DataCodecMock{
		SerializeTokenPropertiesCalled: func(properties dto.TokenProperties) ([]byte, error) {
			require.Equal(t, dto.TokenProperties{
				TokenIdentifier: topics[topicIdxTokenID],
				TokenType:       core.NonFungible,
				Name:            topics[topicIdxName],
				Ticker:          topics[topicIdxTicker],
				NumDecimals:     18,
				EventData:       eventData,
			}, properties)

			return serializedData, nil
		},
		DeserializeEventDataCalled: func(data []byte) (*sovData.EventData, error) {
			require.Equal(t, txEvent.Data, data)
			return eventData, nil
		},
	}

	opFormatter, _ := NewRegisterTokenOpFormatter(dataCodec, &sovereign.TopicsCheckerMock{})
	formattedData, err := opFormatter.CreateOperationData(txEvent)
	require.Nil(t, err)
	require.Equal(t, formattedData, serializedData)
}

func TestRegisterTokenOpFormatter_CreateOperationDataErrorCases(t *testing.T) {
	t.Parallel()

	opFormatter, _ := NewRegisterTokenOpFormatter(&sovereign.DataCodecMock{}, &sovereign.TopicsCheckerMock{})
	topics := [][]byte{
		[]byte("registerToken"),
		[]byte("tokenID"),
		{byte(core.NonFungible)},
		[]byte("name"),
		[]byte("ticker"),
		{18},
	}

	t.Run("invalid num topics", func(t *testing.T) {
		formattedData, err := opFormatter.CreateOperationData(&transaction.Event{Topics: topics[1:]})
		require.Nil(t, formattedData)
		require.ErrorIs(t, err, errInvalidNumTopicsInRegisterToken)
	})
	t.Run("invalid token type", func(t *testing.T) {
		txEvent := &transaction.Event{Topics: topics}
		txEvent.Topics[topicIdxTokenType] = []byte("invalid number of bytes for a number")
		formattedData, err := opFormatter.CreateOperationData(txEvent)
		require.Nil(t, formattedData)
		require.NotNil(t, err)
	})
	t.Run("invalid num decimals", func(t *testing.T) {
		txEvent := &transaction.Event{Topics: topics}
		txEvent.Topics[topicIdxNumDecimals] = []byte("invalid number of bytes for a number")
		formattedData, err := opFormatter.CreateOperationData(txEvent)
		require.Nil(t, formattedData)
		require.NotNil(t, err)
	})
}
