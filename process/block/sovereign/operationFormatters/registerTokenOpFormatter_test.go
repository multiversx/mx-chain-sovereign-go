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
		opFormatter, err := NewRegisterTokenOpFormatter(nil)
		require.Nil(t, opFormatter)
		require.Equal(t, errMx.ErrNilDataCodec, err)
	})
	t.Run("should work", func(t *testing.T) {
		opFormatter, err := NewRegisterTokenOpFormatter(&sovereign.DataCodecMock{})
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
	}

	opFormatter, _ := NewRegisterTokenOpFormatter(dataCodec)
	formattedData, err := opFormatter.CreateOperationData(&transaction.Event{Topics: topics}, eventData)
	require.Nil(t, err)
	require.Equal(t, formattedData, serializedData)
}
