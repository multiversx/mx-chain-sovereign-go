package incomingHeader

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"testing"

	errorsMx "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/sovereign/incomingHeader/dto"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	sovTests "github.com/multiversx/mx-chain-go/testscommon/sovereign"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/stretchr/testify/require"
)

func createArgs() ArgsIncomingHeaderProcessor {
	return ArgsIncomingHeaderProcessor{
		HeadersPool:            &mock.HeadersCacherStub{},
		TxPool:                 &testscommon.ShardedDataStub{},
		Marshaller:             &marshallerMock.MarshalizerMock{},
		Hasher:                 &hashingMocks.HasherMock{},
		OutGoingOperationsPool: &sovTests.OutGoingOperationsPoolMock{},
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

func requireErrorIsInvalidNumTopics(t *testing.T, err error, idx int, numTopics int) {
	require.True(t, strings.Contains(err.Error(), dto.ErrInvalidNumTopicsInEvent.Error()))
	require.True(t, strings.Contains(err.Error(), fmt.Sprintf("%d", idx)))
	require.True(t, strings.Contains(err.Error(), fmt.Sprintf("%d", numTopics)))
}

func createIncomingHeadersWithIncrementalRound(numRounds uint64) []sovereign.IncomingHeaderHandler {
	ret := make([]sovereign.IncomingHeaderHandler, numRounds+1)

	for i := uint64(0); i <= numRounds; i++ {
		ret[i] = &sovereign.IncomingHeader{
			Header: &block.HeaderV2{
				Header: &block.Header{
					Round: i,
				},
			},
			IncomingEvents: []*transaction.Event{
				{
					Topics:     [][]byte{[]byte(dto.TopicIDDepositIncomingTransfer), []byte("addr"), []byte("tokenID1"), []byte("nonce1"), []byte("tokenData1")},
					Data:       []byte("eventData"),
					Identifier: []byte(dto.EventIDDepositIncomingTransfer),
				},
			},
		}
	}

	return ret
}

func TestNewIncomingHeaderHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil headers pool, should return error", func(t *testing.T) {
		args := createArgs()
		args.HeadersPool = nil

		handler, err := NewIncomingHeaderProcessor(args)
		require.Equal(t, errNilHeadersPool, err)
		require.Nil(t, handler)
	})

	t.Run("nil tx pool, should return error", func(t *testing.T) {
		args := createArgs()
		args.TxPool = nil

		handler, err := NewIncomingHeaderProcessor(args)
		require.Equal(t, errNilTxPool, err)
		require.Nil(t, handler)
	})

	t.Run("nil marshaller, should return error", func(t *testing.T) {
		args := createArgs()
		args.Marshaller = nil

		handler, err := NewIncomingHeaderProcessor(args)
		require.Equal(t, core.ErrNilMarshalizer, err)
		require.Nil(t, handler)
	})

	t.Run("nil hasher, should return error", func(t *testing.T) {
		args := createArgs()
		args.Hasher = nil

		handler, err := NewIncomingHeaderProcessor(args)
		require.Equal(t, core.ErrNilHasher, err)
		require.Nil(t, handler)
	})

	t.Run("nil outgoing operations pool, should return error", func(t *testing.T) {
		args := createArgs()
		args.OutGoingOperationsPool = nil

		handler, err := NewIncomingHeaderProcessor(args)
		require.Equal(t, errorsMx.ErrNilOutGoingOperationsPool, err)
		require.Nil(t, handler)
	})

	t.Run("nil data codec, should return error", func(t *testing.T) {
		args := createArgs()
		args.DataCodec = nil

		handler, err := NewIncomingHeaderProcessor(args)
		require.Equal(t, errorsMx.ErrNilDataCodec, err)
		require.Nil(t, handler)
	})

	t.Run("nil topics checker, should return error", func(t *testing.T) {
		args := createArgs()
		args.TopicsChecker = nil

		handler, err := NewIncomingHeaderProcessor(args)
		require.Equal(t, errorsMx.ErrNilTopicsChecker, err)
		require.Nil(t, handler)
	})

	t.Run("should work", func(t *testing.T) {
		args := createArgs()
		handler, err := NewIncomingHeaderProcessor(args)
		require.Nil(t, err)
		require.False(t, check.IfNil(handler))
	})
}

func TestIncomingHeaderHandler_AddHeaderErrorCases(t *testing.T) {
	t.Parallel()

	t.Run("nil header, should return error", func(t *testing.T) {
		t.Parallel()

		args := createArgs()
		handler, _ := NewIncomingHeaderProcessor(args)

		err := handler.AddHeader([]byte("hash"), nil)
		require.Equal(t, data.ErrNilHeader, err)

		incomingHeader := &sovTests.IncomingHeaderStub{
			GetHeaderHandlerCalled: func() data.HeaderHandler {
				return nil
			},
		}
		err = handler.AddHeader([]byte("hash"), incomingHeader)
		require.Equal(t, data.ErrNilHeader, err)
	})

	t.Run("should not add header before start round", func(t *testing.T) {
		t.Parallel()

		startRound := uint64(11)

		args := createArgs()
		args.MainChainNotarizationStartRound = startRound
		wasHeaderAddedCt := 0
		args.HeadersPool = &mock.HeadersCacherStub{
			AddHeaderInShardCalled: func(headerHash []byte, header data.HeaderHandler, shardID uint32) {
				wasHeaderAddedCt++
				switch wasHeaderAddedCt {
				case 1:
					// pre-genesis header, just track internal header
					require.Empty(t, header.(data.ShardHeaderExtendedHandler).GetIncomingEventHandlers())
					require.Equal(t, header.GetRound(), startRound-1)
				case 2:
					require.NotEmpty(t, header.(data.ShardHeaderExtendedHandler).GetIncomingEventHandlers())
					require.Equal(t, header.GetRound(), startRound)
				}
			},
		}
		handler, _ := NewIncomingHeaderProcessor(args)
		headers := createIncomingHeadersWithIncrementalRound(startRound + 1)

		for i := 0; i <= int(startRound-2); i++ {
			err := handler.AddHeader([]byte("hash"), headers[i])
			require.Nil(t, err)
			require.Zero(t, wasHeaderAddedCt)
		}

		err := handler.AddHeader([]byte("hash"), headers[startRound-1])
		require.Nil(t, err)
		require.Equal(t, 1, wasHeaderAddedCt)

		err = handler.AddHeader([]byte("hash"), headers[startRound])
		require.Nil(t, err)
		require.Equal(t, 2, wasHeaderAddedCt)
	})

	t.Run("invalid header type, should return error", func(t *testing.T) {
		t.Parallel()

		args := createArgs()
		handler, _ := NewIncomingHeaderProcessor(args)

		incomingHeader := &sovTests.IncomingHeaderStub{
			GetHeaderHandlerCalled: func() data.HeaderHandler {
				return &block.MetaBlock{}
			},
		}
		err := handler.AddHeader([]byte("hash"), incomingHeader)
		require.Equal(t, errInvalidHeaderType, err)
	})

	t.Run("cannot compute extended header hash, should return error", func(t *testing.T) {
		t.Parallel()

		args := createArgs()

		errMarshaller := errors.New("cannot marshal")
		args.Marshaller = &marshallerMock.MarshalizerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				return nil, errMarshaller
			},
		}
		handler, _ := NewIncomingHeaderProcessor(args)

		err := handler.AddHeader([]byte("hash"), &sovereign.IncomingHeader{Header: &block.HeaderV2{}})
		require.Equal(t, errMarshaller, err)
	})

	t.Run("invalid num topics in deposit event, should return error", func(t *testing.T) {
		t.Parallel()

		errNumTopics := fmt.Errorf("invalid number of topics")

		args := createArgs()
		args.TopicsChecker = &sovTests.TopicsCheckerMock{
			CheckDepositTokensValidityCalled: func(_ [][]byte) error {
				return errNumTopics
			},
			CheckScCallValidityCalled: func(_ [][]byte, _ *sovereign.TransferData) error {
				return errNumTopics
			},
		}

		incomingHeader := &sovereign.IncomingHeader{
			Header: &block.HeaderV2{},
			IncomingEvents: []*transaction.Event{
				{
					Identifier: []byte(dto.EventIDDepositIncomingTransfer),
					Topics:     [][]byte{},
				},
			},
		}

		handler, _ := NewIncomingHeaderProcessor(args)
		err := handler.AddHeader([]byte("hash"), incomingHeader)
		require.ErrorContains(t, err, errNumTopics.Error())
	})

	t.Run("invalid num topics in executed ops event, should return error", func(t *testing.T) {
		t.Parallel()

		errNumTopics := fmt.Errorf("invalid num topics")

		args := createArgs()
		args.TopicsChecker = &sovTests.TopicsCheckerMock{
			CheckDepositTokensValidityCalled: func(_ [][]byte) error {
				return errNumTopics
			},
			CheckScCallValidityCalled: func(_ [][]byte, _ *sovereign.TransferData) error {
				return errNumTopics
			},
		}

		numConfirmedOperations := 0
		args.OutGoingOperationsPool = &sovTests.OutGoingOperationsPoolMock{
			ConfirmOperationCalled: func(hashOfHashes []byte, hash []byte) error {
				numConfirmedOperations++
				return nil
			},
		}

		incomingHeader := &sovereign.IncomingHeader{
			Header: &block.HeaderV2{},
			IncomingEvents: []*transaction.Event{
				{
					Topics:     [][]byte{},
					Identifier: []byte(dto.EventIDExecutedOutGoingBridgeOp),
				},
			},
		}

		handler, _ := NewIncomingHeaderProcessor(args)

		err := handler.AddHeader([]byte("hash"), incomingHeader)
		require.ErrorContains(t, err, dto.ErrInvalidNumTopicsInEvent.Error())

		incomingHeader.IncomingEvents[0] = &transaction.Event{Topics: [][]byte{[]byte(dto.TopicIDDepositIncomingTransfer)}, Identifier: []byte(dto.EventIDExecutedOutGoingBridgeOp)}
		err = handler.AddHeader([]byte("hash"), incomingHeader)
		require.ErrorContains(t, err, errNumTopics.Error())

		incomingHeader.IncomingEvents[0] = &transaction.Event{Topics: [][]byte{[]byte(dto.TopicIDDepositIncomingTransfer), []byte("receiver"), []byte("tokenID")}, Identifier: []byte(dto.EventIDExecutedOutGoingBridgeOp)}
		err = handler.AddHeader([]byte("hash"), incomingHeader)
		require.ErrorContains(t, err, errNumTopics.Error())

		incomingHeader.IncomingEvents[0] = &transaction.Event{Topics: [][]byte{[]byte(dto.TopicIDDepositIncomingTransfer), []byte("receiver"), []byte("token1ID"), []byte("token1Nonce"), []byte("token1Data"), []byte("token2ID")}, Identifier: []byte(dto.EventIDExecutedOutGoingBridgeOp)}
		err = handler.AddHeader([]byte("hash"), incomingHeader)
		require.ErrorContains(t, err, errNumTopics.Error())

		incomingHeader.IncomingEvents[0] = &transaction.Event{Topics: [][]byte{[]byte(dto.TopicIDConfirmedOutGoingOperation)}, Identifier: []byte(dto.EventIDExecutedOutGoingBridgeOp)}
		err = handler.AddHeader([]byte("hash"), incomingHeader)
		requireErrorIsInvalidNumTopics(t, err, 0, 1)

		incomingHeader.IncomingEvents[0] = &transaction.Event{Topics: [][]byte{[]byte(dto.TopicIDConfirmedOutGoingOperation), []byte("hash")}, Identifier: []byte(dto.EventIDExecutedOutGoingBridgeOp)}
		err = handler.AddHeader([]byte("hash"), incomingHeader)
		requireErrorIsInvalidNumTopics(t, err, 0, 2)

		incomingHeader.IncomingEvents[0] = &transaction.Event{Topics: [][]byte{[]byte(dto.TopicIDConfirmedOutGoingOperation), []byte("hash"), []byte("hash1"), []byte("hash2")}, Identifier: []byte(dto.EventIDExecutedOutGoingBridgeOp)}
		err = handler.AddHeader([]byte("hash"), incomingHeader)
		requireErrorIsInvalidNumTopics(t, err, 0, 4)

		require.Equal(t, 0, numConfirmedOperations)

		incomingHeader.IncomingEvents[0] = &transaction.Event{Topics: [][]byte{[]byte("topicID")}, Identifier: []byte(dto.EventIDExecutedOutGoingBridgeOp)}
		err = handler.AddHeader([]byte("hash"), incomingHeader)
		require.True(t, strings.Contains(err.Error(), dto.ErrInvalidIncomingTopicIdentifier.Error()))
	})

	t.Run("invalid event id, should return error", func(t *testing.T) {
		t.Parallel()

		args := createArgs()

		incomingHeader := &sovereign.IncomingHeader{
			Header: &block.HeaderV2{},
			IncomingEvents: []*transaction.Event{
				{
					Identifier: []byte("eventID"),
					Topics:     [][]byte{},
				},
			},
		}

		handler, _ := NewIncomingHeaderProcessor(args)
		err := handler.AddHeader([]byte("hash"), incomingHeader)
		require.Equal(t, dto.ErrInvalidIncomingEventIdentifier, err)
	})

	t.Run("cannot compute scr hash, should return error", func(t *testing.T) {
		t.Parallel()

		args := createArgs()

		numSCRsAdded := 0
		args.TxPool = &testscommon.ShardedDataStub{
			AddDataCalled: func(key []byte, data interface{}, sizeInBytes int, cacheID string) {
				numSCRsAdded++
			},
		}

		errMarshaller := errors.New("cannot marshal")
		args.Marshaller = &marshallerMock.MarshalizerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				_, isSCR := obj.(*smartContractResult.SmartContractResult)
				if isSCR {
					return nil, errMarshaller
				}

				return json.Marshal(obj)
			},
		}

		incomingHeader := &sovereign.IncomingHeader{
			Header: &block.HeaderV2{},
			IncomingEvents: []*transaction.Event{
				{
					Identifier: []byte(dto.EventIDDepositIncomingTransfer),
					Topics:     [][]byte{[]byte(dto.TopicIDDepositIncomingTransfer), []byte("addr"), []byte("tokenID1"), []byte("nonce1"), []byte("tokenData1")},
					Data:       []byte("eventData"),
				},
			},
		}

		handler, _ := NewIncomingHeaderProcessor(args)
		err := handler.AddHeader([]byte("hash"), incomingHeader)
		require.ErrorIs(t, err, errMarshaller)
		require.Equal(t, 0, numSCRsAdded)
	})

	t.Run("cannot create event data, should return error", func(t *testing.T) {
		t.Parallel()

		errCannotDeserializeEventData := fmt.Errorf("cannot deserialize event data")

		args := createArgs()
		args.DataCodec = &sovTests.DataCodecMock{
			DeserializeEventDataCalled: func(_ []byte) (*sovereign.EventData, error) {
				return nil, errCannotDeserializeEventData
			},
		}

		incomingHeader := &sovereign.IncomingHeader{
			Header: &block.HeaderV2{},
			IncomingEvents: []*transaction.Event{
				{
					Identifier: []byte(dto.EventIDDepositIncomingTransfer),
					Topics:     [][]byte{[]byte(dto.TopicIDDepositIncomingTransfer), []byte("addr"), []byte("tokenID1"), []byte("nonce1"), []byte("tokenData1")},
					Data:       []byte("data"),
				},
			},
		}

		handler, _ := NewIncomingHeaderProcessor(args)
		err := handler.AddHeader([]byte("hash"), incomingHeader)
		require.ErrorContains(t, err, errCannotDeserializeEventData.Error())
	})

	t.Run("cannot create token data, should return error", func(t *testing.T) {
		t.Parallel()

		errCannotDeserializeTokenData := fmt.Errorf("cannot deserialize token data")

		args := createArgs()
		args.DataCodec = &sovTests.DataCodecMock{
			DeserializeTokenDataCalled: func(_ []byte) (*sovereign.EsdtTokenData, error) {
				return nil, errCannotDeserializeTokenData
			},
		}

		incomingHeader := &sovereign.IncomingHeader{
			Header: &block.HeaderV2{},
			IncomingEvents: []*transaction.Event{
				{
					Identifier: []byte(dto.EventIDDepositIncomingTransfer),
					Topics:     [][]byte{[]byte(dto.TopicIDDepositIncomingTransfer), []byte("addr"), []byte("tokenID1"), []byte("nonce1"), []byte("tokenData1")},
					Data:       []byte("data"),
				},
			},
		}

		handler, _ := NewIncomingHeaderProcessor(args)
		err := handler.AddHeader([]byte("hash"), incomingHeader)
		require.ErrorContains(t, err, errCannotDeserializeTokenData.Error())
	})
}

// AddHeader with 5 incoming operations
// 1. multi-transfer with 2 tokens and SC call
// 2. multi-transfer with 1 token and SC call
// 3. executed operation confirmation
// 4. sc call, no tokens
// 5. multi-transfer with 1 token, no SC call
func TestIncomingHeaderHandler_AddHeader(t *testing.T) {
	t.Parallel()

	args := createArgs()

	addr1 := []byte("addr1")
	addr2 := []byte("addr2")
	addr3 := []byte("addr3")

	token1 := []byte("token1")
	token2 := []byte("token2")

	token1Nonce := make([]byte, 0)
	token2Nonce := []byte{0x01}

	amount1 := big.NewInt(100)
	token1Data := amount1.Bytes()
	amount2 := big.NewInt(50)
	token2Data := amount2.Bytes()

	scr1Nonce := uint64(0)
	scr2Nonce := uint64(1)
	scr3Nonce := uint64(2)
	scr4Nonce := uint64(3)
	gasLimit1 := uint64(45100)
	gasLimit2 := uint64(54100)
	gasLimit3 := uint64(63100)
	func1 := []byte("func1")
	func2 := []byte("func2")
	func3 := "func3"
	arg1 := []byte("arg1")
	arg2 := []byte("arg2")

	scr1 := &smartContractResult.SmartContractResult{
		Nonce:   scr1Nonce,
		Value:   big.NewInt(0),
		RcvAddr: addr1,
		SndAddr: core.ESDTSCAddress,
		Data: []byte(core.BuiltInFunctionMultiESDTNFTTransfer + "@02" +
			"@" + hex.EncodeToString(token1) +
			"@" + hex.EncodeToString(token1Nonce) +
			"@" + hex.EncodeToString(token1Data) +
			"@" + hex.EncodeToString(token2) +
			"@" + hex.EncodeToString(token2Nonce) +
			"@" + hex.EncodeToString(token2Data) +
			"@" + hex.EncodeToString(func1) +
			"@" + hex.EncodeToString(arg1) +
			"@" + hex.EncodeToString(arg2)),
		GasLimit: gasLimit1,
	}
	scr2 := &smartContractResult.SmartContractResult{
		Nonce:   scr2Nonce,
		Value:   big.NewInt(0),
		RcvAddr: addr2,
		SndAddr: core.ESDTSCAddress,
		Data: []byte(core.BuiltInFunctionMultiESDTNFTTransfer + "@01" +
			"@" + hex.EncodeToString(token1) +
			"@" + hex.EncodeToString(token1Nonce) +
			"@" + hex.EncodeToString(token2Data) +
			"@" + hex.EncodeToString(func2) +
			"@" + hex.EncodeToString(arg1)),
		GasLimit: gasLimit2,
	}
	scr3 := &smartContractResult.SmartContractResult{
		Nonce:   scr3Nonce,
		Value:   big.NewInt(0),
		RcvAddr: addr3,
		SndAddr: core.ESDTSCAddress,
		Data: []byte(func3 +
			"@" + hex.EncodeToString(arg1) +
			"@" + hex.EncodeToString(arg2)),
		GasLimit: gasLimit3,
	}
	scr4 := &smartContractResult.SmartContractResult{
		Nonce:   scr4Nonce,
		Value:   big.NewInt(0),
		RcvAddr: addr3,
		SndAddr: core.ESDTSCAddress,
		Data: []byte(core.BuiltInFunctionMultiESDTNFTTransfer + "@01" +
			"@" + hex.EncodeToString(token2) +
			"@" + hex.EncodeToString(token2Nonce) +
			"@" + hex.EncodeToString(token2Data)),
		GasLimit: 0,
	}

	scrHash1, err := core.CalculateHash(args.Marshaller, args.Hasher, scr1)
	require.Nil(t, err)
	scrHash2, err := core.CalculateHash(args.Marshaller, args.Hasher, scr2)
	require.Nil(t, err)
	scrHash3, err := core.CalculateHash(args.Marshaller, args.Hasher, scr3)
	require.Nil(t, err)
	scrHash4, err := core.CalculateHash(args.Marshaller, args.Hasher, scr4)
	require.Nil(t, err)

	cacheID := process.ShardCacherIdentifier(core.MainChainShardId, core.SovereignChainShardId)

	type scrInPool struct {
		data        *smartContractResult.SmartContractResult
		sizeInBytes int
		cacheID     string
	}
	expectedSCRsInPool := map[string]*scrInPool{
		string(scrHash1): {
			data:        scr1,
			sizeInBytes: scr1.Size(),
			cacheID:     cacheID,
		},
		string(scrHash2): {
			data:        scr2,
			sizeInBytes: scr2.Size(),
			cacheID:     cacheID,
		},
		string(scrHash3): {
			data:        scr3,
			sizeInBytes: scr3.Size(),
			cacheID:     cacheID,
		},
		string(scrHash4): {
			data:        scr4,
			sizeInBytes: scr4.Size(),
			cacheID:     cacheID,
		},
	}

	headerV2 := &block.HeaderV2{ScheduledRootHash: []byte("root hash")}

	transfer1 := [][]byte{
		token1,
		token1Nonce,
		token1Data,
	}
	transfer2 := [][]byte{
		token2,
		token2Nonce,
		token2Data,
	}
	topic1 := [][]byte{
		[]byte(dto.TopicIDDepositIncomingTransfer),
		addr1,
	}
	topic1 = append(topic1, transfer1...)
	topic1 = append(topic1, transfer2...)
	eventData1 := []byte("eventData1")

	transfer3 := [][]byte{
		token1,
		token1Nonce,
		token2Data,
	}
	topic2 := [][]byte{
		[]byte(dto.TopicIDDepositIncomingTransfer),
		addr2,
	}
	topic2 = append(topic2, transfer3...)
	eventData2 := []byte("eventData2")

	topic3 := [][]byte{
		[]byte(dto.TopicIDConfirmedOutGoingOperation),
		[]byte("hashOfHashes"),
		[]byte("hashOfBridgeOp"),
	}

	topic4 := [][]byte{
		[]byte(dto.TopicIDSCCall),
		addr3,
	}
	eventData3 := []byte("eventData3")

	transfer4 := [][]byte{
		token2,
		token2Nonce,
		token2Data,
	}
	topic5 := [][]byte{
		[]byte(dto.TopicIDDepositIncomingTransfer),
		addr3,
	}
	topic5 = append(topic5, transfer4...)
	eventData4 := []byte("eventData4")

	incomingEvents := []*transaction.Event{
		{
			Identifier: []byte(dto.EventIDDepositIncomingTransfer),
			Topics:     topic1,
			Data:       eventData1,
		},
		{
			Identifier: []byte(dto.EventIDDepositIncomingTransfer),
			Topics:     topic2,
			Data:       eventData2,
		},
		{
			Identifier: []byte(dto.EventIDExecutedOutGoingBridgeOp),
			Topics:     topic3,
		},
		{
			Identifier: []byte(dto.EventIDDepositIncomingTransfer),
			Topics:     topic4,
			Data:       eventData3,
		},
		{
			Identifier: []byte(dto.EventIDDepositIncomingTransfer),
			Topics:     topic5,
			Data:       eventData4,
		},
	}

	extendedHeader := &block.ShardHeaderExtended{
		Header: headerV2,
		IncomingMiniBlocks: []*block.MiniBlock{
			{
				TxHashes:        [][]byte{scrHash1, scrHash2, scrHash3, scrHash4},
				ReceiverShardID: core.SovereignChainShardId,
				SenderShardID:   core.MainChainShardId,
				Type:            block.SmartContractResultBlock,
			},
		},
		IncomingEvents: incomingEvents,
	}
	extendedHeaderHash, err := core.CalculateHash(args.Marshaller, args.Hasher, extendedHeader)
	require.Nil(t, err)

	wasAddedInHeaderPool := false
	args.HeadersPool = &mock.HeadersCacherStub{
		AddHeaderInShardCalled: func(headerHash []byte, header data.HeaderHandler, shardID uint32) {
			require.Equal(t, extendedHeaderHash, headerHash)
			require.Equal(t, extendedHeader, header)
			require.Equal(t, core.MainChainShardId, shardID)

			wasAddedInHeaderPool = true
		},
	}

	wasAddedInTxPool := false
	args.TxPool = &testscommon.ShardedDataStub{
		AddDataCalled: func(key []byte, data interface{}, sizeInBytes int, cacheID string) {
			expectedSCR, found := expectedSCRsInPool[string(key)]
			require.True(t, found)

			require.Equal(t, expectedSCR.data, data)
			require.Equal(t, expectedSCR.sizeInBytes, sizeInBytes)
			require.Equal(t, expectedSCR.cacheID, cacheID)

			wasAddedInTxPool = true
		},
	}

	wasOutGoingOpConfirmed := false
	args.OutGoingOperationsPool = &sovTests.OutGoingOperationsPoolMock{
		ConfirmOperationCalled: func(hashOfHashes []byte, hash []byte) error {
			require.Equal(t, topic3[0], []byte(dto.TopicIDConfirmedOutGoingOperation))
			require.Equal(t, topic3[1], hashOfHashes)
			require.Equal(t, topic3[2], hash)

			wasOutGoingOpConfirmed = true
			return nil
		},
	}

	checkDepositValidityCt := 0
	checkScCallValidityCt := 0
	args.TopicsChecker = &sovTests.TopicsCheckerMock{
		CheckDepositTokensValidityCalled: func(topics [][]byte) error {
			checkDepositValidityCt++

			switch checkDepositValidityCt {
			case 1:
				require.Equal(t, topic1, topics)
			case 2:
				require.Equal(t, topic2, topics)
			case 3:
				require.Equal(t, topic5, topics)
			default:
				require.Fail(t, "check deposit tokens validity called more than 3 times")
			}

			return nil
		},
		CheckScCallValidityCalled: func(topics [][]byte, transferData *sovereign.TransferData) error {
			checkScCallValidityCt++

			switch checkScCallValidityCt {
			case 1:
				require.Equal(t, topic4, topics)
				require.NotNil(t, transferData)
			default:
				require.Fail(t, "check sc call validity called more than 1 time")
			}

			return nil
		},
	}

	deserializeEventDataCt := 0
	deserializeTokenDataCt := 0
	args.DataCodec = &sovTests.DataCodecMock{
		DeserializeEventDataCalled: func(data []byte) (*sovereign.EventData, error) {
			deserializeEventDataCt++

			switch deserializeEventDataCt {
			case 1:
				require.Equal(t, eventData1, data)

				return &sovereign.EventData{
					Nonce: scr1Nonce,
					TransferData: &sovereign.TransferData{
						Function: func1,
						Args:     [][]byte{arg1, arg2},
						GasLimit: gasLimit1,
					},
				}, nil

			case 2:
				require.Equal(t, eventData2, data)

				return &sovereign.EventData{
					Nonce: scr2Nonce,
					TransferData: &sovereign.TransferData{
						Function: func2,
						Args:     [][]byte{arg1},
						GasLimit: gasLimit2,
					},
				}, nil

			case 3:
				require.Equal(t, eventData3, data)

				return &sovereign.EventData{
					Nonce: scr3Nonce,
					TransferData: &sovereign.TransferData{
						Function: []byte(func3),
						Args:     [][]byte{arg1, arg2},
						GasLimit: gasLimit3,
					},
				}, nil

			case 4:
				require.Equal(t, eventData4, data)

				return &sovereign.EventData{
					Nonce:        scr4Nonce,
					TransferData: nil,
				}, nil

			default:
				require.Fail(t, "deserialize event data called more than 4 times")
				return nil, nil
			}
		},
		DeserializeTokenDataCalled: func(data []byte) (*sovereign.EsdtTokenData, error) {
			deserializeTokenDataCt++

			switch deserializeTokenDataCt {
			case 1:
				require.Equal(t, token1Data, data)
				return &sovereign.EsdtTokenData{
					TokenType: core.Fungible,
					Amount:    amount1,
				}, nil

			case 2, 3, 4:
				require.Equal(t, token2Data, data)
				return &sovereign.EsdtTokenData{
					TokenType: core.Fungible,
					Amount:    amount2,
				}, nil

			default:
				require.Fail(t, "deserialize token data called more than 4 times")
				return nil, nil
			}
		},
	}

	handler, _ := NewIncomingHeaderProcessor(args)
	incomingHeader := &sovereign.IncomingHeader{
		Header:         headerV2,
		IncomingEvents: incomingEvents,
	}

	err = handler.AddHeader([]byte("hash"), incomingHeader)
	require.Nil(t, err)
	require.Equal(t, 3, checkDepositValidityCt)
	require.Equal(t, 1, checkScCallValidityCt)
	require.Equal(t, 4, deserializeEventDataCt)
	require.Equal(t, 4, deserializeTokenDataCt)
	require.True(t, wasAddedInHeaderPool)
	require.True(t, wasAddedInTxPool)
	require.True(t, wasOutGoingOpConfirmed)
}
