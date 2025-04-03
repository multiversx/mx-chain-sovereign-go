package incomingEventsProc

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/stretchr/testify/require"

	errorsMx "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	sovTests "github.com/multiversx/mx-chain-go/testscommon/sovereign"
)

func createArgs() EventProcDepositTokensArgs {
	return EventProcDepositTokensArgs{
		Marshaller: &marshallerMock.MarshalizerMock{},
		Hasher:     &hashingMocks.HasherMock{},
		DataCodec: &sovTests.DataCodecMock{
			DeserializeTokenDataCalled: func(_ []byte) (*sovereign.EsdtTokenData, error) {
				return &sovereign.EsdtTokenData{
					Amount: big.NewInt(0),
				}, nil
			},
		},
		TopicsChecker: &sovTests.TopicsCheckerMock{},
	}
}

func TestNewEventProcDepositTokens(t *testing.T) {
	t.Parallel()

	t.Run("nil marshaller, should return error", func(t *testing.T) {
		args := createArgs()
		args.Marshaller = nil

		handler, err := NewEventProcDepositTokens(args)
		require.Equal(t, core.ErrNilMarshalizer, err)
		require.Nil(t, handler)
	})

	t.Run("nil hasher, should return error", func(t *testing.T) {
		args := createArgs()
		args.Hasher = nil

		handler, err := NewEventProcDepositTokens(args)
		require.Equal(t, core.ErrNilHasher, err)
		require.Nil(t, handler)
	})

	t.Run("nil data codec, should return error", func(t *testing.T) {
		args := createArgs()
		args.DataCodec = nil

		handler, err := NewEventProcDepositTokens(args)
		require.Equal(t, errorsMx.ErrNilDataCodec, err)
		require.Nil(t, handler)
	})

	t.Run("nil topics checker, should return error", func(t *testing.T) {
		args := createArgs()
		args.TopicsChecker = nil

		handler, err := NewEventProcDepositTokens(args)
		require.Equal(t, errorsMx.ErrNilTopicsChecker, err)
		require.Nil(t, handler)
	})

	t.Run("should work", func(t *testing.T) {
		args := createArgs()
		handler, err := NewEventProcDepositTokens(args)
		require.NotNil(t, handler)
		require.Nil(t, err)
	})
}

func TestDepositEventProc_extractSCTransferInfo(t *testing.T) {
	t.Parallel()

	t.Run("nil transfer data", func(t *testing.T) {
		t.Parallel()

		gasLimit, functionCallWithArgs := extractSCTransferInfo(nil)
		require.Zero(t, gasLimit)
		require.Empty(t, functionCallWithArgs)
	})

	t.Run("transfer data with function no args", func(t *testing.T) {
		t.Parallel()

		func1 := []byte("func1")
		expectedGasLimit := uint64(1)
		expectedArgs := append([]byte("@"), hex.EncodeToString(func1)...)

		gasLimit, functionCallWithArgs := extractSCTransferInfo(&sovereign.TransferData{
			GasLimit: expectedGasLimit,
			Function: func1,
		})
		require.Equal(t, expectedGasLimit, gasLimit)
		require.Equal(t, expectedArgs, functionCallWithArgs)
	})

	t.Run("transfer data with function and args", func(t *testing.T) {
		t.Parallel()

		func1 := []byte("func1")
		arg1 := []byte("arg1")
		arg2 := []byte("arg2")
		expectedGasLimit := uint64(2)
		expectedArgs := append([]byte("@"), hex.EncodeToString(func1)...)
		expectedArgs = append(expectedArgs, "@"+hex.EncodeToString(arg1)...)
		expectedArgs = append(expectedArgs, "@"+hex.EncodeToString(arg2)...)

		gasLimit, functionCallWithArgs := extractSCTransferInfo(&sovereign.TransferData{
			GasLimit: expectedGasLimit,
			Function: func1,
			Args:     [][]byte{arg1, arg2},
		})
		require.Equal(t, expectedGasLimit, gasLimit)
		require.Equal(t, expectedArgs, functionCallWithArgs)
	})
}

func TestDepositEventProc_createSCRData(t *testing.T) {
	t.Parallel()

	topicID := []byte("topicID")
	receiver := []byte("rcv")
	nft := []byte("nft")
	nonce := []byte("nonce")
	nftData := []byte("nftData")
	transferGas := uint64(1)
	func1 := []byte("func1")
	arg1 := []byte("arg1")
	arg2 := []byte("arg2")

	t.Run("create SCR data with no tokens, with transfer data", func(t *testing.T) {
		t.Parallel()

		args := createArgs()
		handler, _ := NewEventProcDepositTokens(args)

		topics := [][]byte{
			topicID,
			receiver,
		}
		eventData := &sovereign.EventData{
			TransferData: &sovereign.TransferData{
				GasLimit: transferGas,
				Function: func1,
				Args:     [][]byte{arg1, arg2},
			},
		}

		scrData, gasLimit, err := handler.createSCRData(topics, eventData)
		require.Nil(t, err)
		require.Equal(t, transferGas, gasLimit)

		expectedSCR := func1
		expectedSCR = append(expectedSCR, "@"+hex.EncodeToString(arg1)...)
		expectedSCR = append(expectedSCR, "@"+hex.EncodeToString(arg2)...)
		require.Equal(t, expectedSCR, scrData)
	})
	t.Run("create SCR data with tokens, no transfer data", func(t *testing.T) {
		t.Parallel()

		args := createArgs()
		args.DataCodec = &sovTests.DataCodecMock{
			DeserializeTokenDataCalled: func(_ []byte) (*sovereign.EsdtTokenData, error) {
				return &sovereign.EsdtTokenData{
					TokenType: core.NonFungible,
					Royalties: big.NewInt(0),
				}, nil
			},
		}
		args.Marshaller = &marshallerMock.MarshalizerStub{
			MarshalCalled: func(_ interface{}) ([]byte, error) {
				return nftData, nil
			},
		}
		handler, _ := NewEventProcDepositTokens(args)

		topics := [][]byte{
			topicID,
			receiver,
			nft,
			nonce,
			nftData,
		}
		eventData := &sovereign.EventData{
			TransferData: nil,
		}

		ret, gasLimit, err := handler.createSCRData(topics, eventData)
		require.Nil(t, err)
		require.Zero(t, gasLimit)

		expectedSCR := []byte(core.BuiltInFunctionMultiESDTNFTTransfer + "@01")
		expectedSCR = append(expectedSCR, "@"+hex.EncodeToString(nft)...)
		expectedSCR = append(expectedSCR, "@"+hex.EncodeToString(nonce)...)
		expectedSCR = append(expectedSCR, "@"+hex.EncodeToString(nftData)...)
		require.Equal(t, expectedSCR, ret)
	})
	t.Run("create SCR data with tokens, with transfer data", func(t *testing.T) {
		t.Parallel()

		args := createArgs()
		args.DataCodec = &sovTests.DataCodecMock{
			DeserializeTokenDataCalled: func(_ []byte) (*sovereign.EsdtTokenData, error) {
				return &sovereign.EsdtTokenData{
					TokenType: core.NonFungible,
					Royalties: big.NewInt(0),
				}, nil
			},
		}
		args.Marshaller = &marshallerMock.MarshalizerStub{
			MarshalCalled: func(_ interface{}) ([]byte, error) {
				return nftData, nil
			},
		}
		handler, _ := NewEventProcDepositTokens(args)

		topics := [][]byte{
			topicID,
			receiver,
			nft,
			nonce,
			nftData,
		}
		eventData := &sovereign.EventData{
			TransferData: &sovereign.TransferData{
				GasLimit: transferGas,
				Function: func1,
				Args:     [][]byte{arg1, arg2},
			},
		}

		ret, gasLimit, err := handler.createSCRData(topics, eventData)
		require.Nil(t, err)
		require.Equal(t, transferGas, gasLimit)

		expectedSCR := []byte(core.BuiltInFunctionMultiESDTNFTTransfer + "@01")
		expectedSCR = append(expectedSCR, "@"+hex.EncodeToString(nft)...)
		expectedSCR = append(expectedSCR, "@"+hex.EncodeToString(nonce)...)
		expectedSCR = append(expectedSCR, "@"+hex.EncodeToString(nftData)...)
		expectedSCR = append(expectedSCR, "@"+hex.EncodeToString(func1)...)
		expectedSCR = append(expectedSCR, "@"+hex.EncodeToString(arg1)...)
		expectedSCR = append(expectedSCR, "@"+hex.EncodeToString(arg2)...)
		require.Equal(t, expectedSCR, ret)
	})
}
