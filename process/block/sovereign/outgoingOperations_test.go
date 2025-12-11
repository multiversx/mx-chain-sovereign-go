package sovereign

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
	dtoCore "github.com/multiversx/mx-chain-core-go/data/sovereign/dto"
	transactionData "github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/process/block/sovereign/incomingHeader/dto"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/multiversx/mx-chain-go/errors"
	sovTests "github.com/multiversx/mx-chain-go/testscommon/sovereign"
	"github.com/multiversx/mx-chain-go/testscommon/state"
)

func createEvents() []SubscribedEvent {
	return []SubscribedEvent{
		{
			Identifier: []byte(topicIDDeposit),
			Addresses: map[string]string{
				"decodedAddr": "encodedAddr",
			},
		},
	}
}

func createArgs() ArgsOutgoingOperations {
	return ArgsOutgoingOperations{
		SubscribedEvents:  createEvents(),
		DataCodec:         &sovTests.DataCodecMock{},
		TopicsChecker:     &sovTests.TopicsCheckerMock{},
		PeerAccountsDB:    &state.AccountsStub{},
		ChainNonceHandler: &sovTests.OutGoingChainNonceMock{},
	}
}

func TestNewOutgoingOperationsFormatter(t *testing.T) {
	t.Parallel()

	t.Run("no subscribed events, should return error", func(t *testing.T) {
		args := createArgs()
		args.SubscribedEvents = []SubscribedEvent{}
		creator, err := NewOutgoingOperationsFormatter(args)
		require.Nil(t, creator)
		require.Equal(t, errNoSubscribedEvent, err)
	})

	t.Run("invalid subscribed event, should return error", func(t *testing.T) {
		args := createArgs()
		args.SubscribedEvents = []SubscribedEvent{
			{
				Identifier: []byte("invalid"),
				Addresses: map[string]string{
					"decodedAddr": "encodedAddr",
				},
			},
		}
		creator, err := NewOutgoingOperationsFormatter(args)
		require.Nil(t, creator)
		require.ErrorIs(t, err, errUnsupportedEventType)
	})

	t.Run("nil data codec, should return error", func(t *testing.T) {
		args := createArgs()
		args.DataCodec = nil
		creator, err := NewOutgoingOperationsFormatter(args)
		require.Nil(t, creator)
		require.Equal(t, errors.ErrNilDataCodec, err)
	})

	t.Run("nil outgoing op nonce handler, should return error", func(t *testing.T) {
		args := createArgs()
		args.ChainNonceHandler = nil
		creator, err := NewOutgoingOperationsFormatter(args)
		require.Nil(t, creator)
		require.Equal(t, errors.ErrNilOutGoingOpNonceChainHandler, err)
	})

	t.Run("nil topics checker, should return error", func(t *testing.T) {
		args := createArgs()
		args.TopicsChecker = nil
		creator, err := NewOutgoingOperationsFormatter(args)
		require.Nil(t, creator)
		require.Equal(t, errors.ErrNilTopicsChecker, err)
	})

	t.Run("should work with deposit tokens formatter", func(t *testing.T) {
		args := createArgs()
		creator, err := NewOutgoingOperationsFormatter(args)
		require.Nil(t, err)
		require.False(t, creator.IsInterfaceNil())
		require.Len(t, creator.opFormatters, 1)
		require.Contains(t, creator.opFormatters, topicIDDeposit)
	})

	t.Run("should work with deposit tokens and register token formatters", func(t *testing.T) {
		args := createArgs()
		args.SubscribedEvents = append(args.SubscribedEvents, SubscribedEvent{
			Identifier: []byte("registerToken"),
			Addresses: map[string]string{
				"decodedAddr": "encodedAddr",
			},
		})
		creator, err := NewOutgoingOperationsFormatter(args)
		require.Nil(t, err)
		require.False(t, creator.IsInterfaceNil())
		require.Len(t, creator.opFormatters, 2)
		require.Contains(t, creator.opFormatters, topicIDDeposit)
		require.Contains(t, creator.opFormatters, topicIDRegisterToken)
	})
}

func createArgsOutGoingOpsFormatterWithEvents() ArgsOutgoingOperations {
	events := []SubscribedEvent{
		{
			Identifier: []byte(topicIDDeposit),
			Addresses: map[string]string{
				"addr1": "addr1",
				"addr2": "addr2",
			},
		},
	}

	return ArgsOutgoingOperations{
		SubscribedEvents:  events,
		DataCodec:         &sovTests.DataCodecMock{},
		TopicsChecker:     &sovTests.TopicsCheckerMock{},
		PeerAccountsDB:    &state.AccountsStub{},
		ChainNonceHandler: &sovTests.OutGoingChainNonceMock{},
	}
}

func createOutgoingOpsFormatter() *outgoingOperations {
	args := createArgsOutGoingOpsFormatterWithEvents()
	opFormatter, _ := NewOutgoingOperationsFormatter(args)
	return opFormatter
}

func TestOutgoingOperations_CheckEvent(t *testing.T) {
	t.Parallel()

	t.Run("invalid identifier", func(t *testing.T) {
		t.Parallel()

		args := createArgs()
		args.SubscribedEvents = []SubscribedEvent{
			{
				Identifier: []byte(""),
				Addresses: map[string]string{
					"addr1": "addr1",
				},
			},
		}

		opFormatter, err := NewOutgoingOperationsFormatter(args)
		require.Nil(t, opFormatter)
		require.ErrorContains(t, err, "no subscribed identifier")
	})
	t.Run("no addresses", func(t *testing.T) {
		t.Parallel()

		args := createArgs()
		args.SubscribedEvents = []SubscribedEvent{
			{
				Identifier: []byte("identifier"),
				Addresses:  map[string]string{},
			},
		}

		opFormatter, err := NewOutgoingOperationsFormatter(args)
		require.Nil(t, opFormatter)
		require.ErrorContains(t, err, errNoSubscribedAddresses.Error())
	})
	t.Run("invalid address", func(t *testing.T) {
		t.Parallel()

		args := createArgs()
		args.SubscribedEvents = []SubscribedEvent{
			{
				Identifier: []byte("identifier"),
				Addresses: map[string]string{
					"addr1": "",
				},
			},
		}

		opFormatter, err := NewOutgoingOperationsFormatter(args)
		require.Nil(t, opFormatter)
		require.ErrorContains(t, err, errNoSubscribedAddresses.Error())
	})
}

func TestOutgoingOperations_CreateOutgoingTxsDataErrorCases(t *testing.T) {
	t.Parallel()

	logs := []*data.LogData{
		{
			LogHandler: &transactionData.Log{
				Address: nil,
				Events: []*transactionData.Event{
					{
						Address:    []byte("addr1"),
						Identifier: []byte(topicIDDeposit),
						Topics:     [][]byte{[]byte("topic1"), []byte("topic1"), []byte("topic1"), []byte("topic1"), []byte("topic1")},
						Data:       []byte("data"),
					},
				},
			},
			TxHash: "",
		},
	}

	t.Run("nil logs", func(t *testing.T) {
		t.Parallel()

		outgoingOpsFormatter := createOutgoingOpsFormatter()
		outgoingTxData, err := outgoingOpsFormatter.CreateOutgoingTxsData(nil)
		require.NoError(t, err)
		require.Equal(t, 0, len(outgoingTxData))
	})
	t.Run("deserialize token error", func(t *testing.T) {
		t.Parallel()

		args := createArgsOutGoingOpsFormatterWithEvents()

		errDeserializeTokenData := fmt.Errorf("deserialize token data error")
		args.DataCodec = &sovTests.DataCodecMock{
			DeserializeTokenDataCalled: func(_ []byte) (*sovereign.EsdtTokenData, error) {
				return nil, errDeserializeTokenData
			},
		}
		outgoingOpsFormatter, _ := NewOutgoingOperationsFormatter(args)

		outgoingTxData, err := outgoingOpsFormatter.CreateOutgoingTxsData(logs)
		require.Nil(t, outgoingTxData)
		require.Equal(t, errDeserializeTokenData, err)
	})
	t.Run("deserialize event error", func(t *testing.T) {
		t.Parallel()

		args := createArgsOutGoingOpsFormatterWithEvents()
		errDeserializeEventData := fmt.Errorf("deserialize event data error")
		args.DataCodec = &sovTests.DataCodecMock{
			DeserializeEventDataCalled: func(data []byte) (*sovereign.EventData, error) {
				return nil, errDeserializeEventData
			},
		}

		outgoingOpsFormatter, _ := NewOutgoingOperationsFormatter(args)
		outgoingTxData, err := outgoingOpsFormatter.CreateOutgoingTxsData(logs)
		require.Nil(t, outgoingTxData)
		require.Equal(t, errDeserializeEventData, err)
	})
	t.Run("serialize operation error", func(t *testing.T) {
		t.Parallel()

		args := createArgsOutGoingOpsFormatterWithEvents()

		errSerializeOperation := fmt.Errorf("serialize operation error")
		args.DataCodec = &sovTests.DataCodecMock{
			SerializeOperationCalled: func(operation sovereign.Operation) ([]byte, error) {
				return nil, errSerializeOperation
			},
		}
		outgoingOpsFormatter, _ := NewOutgoingOperationsFormatter(args)

		outgoingTxData, err := outgoingOpsFormatter.CreateOutgoingTxsData(logs)
		require.Nil(t, outgoingTxData)
		require.Equal(t, errSerializeOperation, err)
	})
	t.Run("check validity error", func(t *testing.T) {
		t.Parallel()

		args := createArgsOutGoingOpsFormatterWithEvents()
		errInvalidTopics := fmt.Errorf("check topics error")
		args.TopicsChecker = &sovTests.TopicsCheckerMock{
			CheckValidityCalled: func(_ [][]byte, _ *sovereign.TransferData) error {
				return errInvalidTopics
			},
		}

		outgoingOpsFormatter, _ := NewOutgoingOperationsFormatter(args)
		outgoingTxData, err := outgoingOpsFormatter.CreateOutgoingTxsData(logs)
		require.Nil(t, outgoingTxData)
		require.Equal(t, errInvalidTopics, err)
	})
}

func TestOutgoingOperations_CreateOutgoingTxData(t *testing.T) {
	t.Parallel()

	addr1 := []byte("addr1")
	addr2 := []byte("addr2")

	identifier1 := []byte(topicIDDeposit)
	identifier2 := []byte("send")

	tokenData1 := []byte("tokenData1")
	topic1 := [][]byte{
		[]byte(topicIDDeposit),
		[]byte("rcv1"),
		[]byte("token1"),
		[]byte("nonce1"),
		tokenData1,
	}
	data1 := []byte("data1")

	evData := &sovereign.EventData{
		Nonce: 1,
		TransferData: &sovereign.TransferData{
			GasLimit: 20000000,
			Function: []byte("add"),
			Args:     [][]byte{big.NewInt(20000000).Bytes()},
		},
	}

	amount := new(big.Int)
	amount.SetString("123000000000000000000", 10)
	tokenData := sovereign.EsdtTokenData{
		TokenType: core.Fungible,
		Amount:    amount,
	}

	operationBytes := []byte("operationBytes")

	dataCodec := &sovTests.DataCodecMock{
		DeserializeEventDataCalled: func(data []byte) (*sovereign.EventData, error) {
			require.Equal(t, data1, data)

			return evData, nil
		},
		DeserializeTokenDataCalled: func(data []byte) (*sovereign.EsdtTokenData, error) {
			require.Equal(t, tokenData1, data)

			return &tokenData, nil
		},
		SerializeOperationCalled: func(operation sovereign.Operation) ([]byte, error) {
			require.Equal(t, evData, operation.Data)
			require.Equal(t, tokenData, operation.Tokens[0].Data)

			return operationBytes, nil
		},
	}

	events := []SubscribedEvent{
		{
			Identifier: identifier1,
			Addresses: map[string]string{
				string(addr1): string(addr1),
				string(addr2): string(addr2),
			},
		},
	}

	args := ArgsOutgoingOperations{
		SubscribedEvents:  events,
		DataCodec:         dataCodec,
		TopicsChecker:     &sovTests.TopicsCheckerMock{},
		PeerAccountsDB:    &state.AccountsStub{},
		ChainNonceHandler: &sovTests.OutGoingChainNonceMock{},
	}
	opFormatter, _ := NewOutgoingOperationsFormatter(args)

	logs := []*data.LogData{
		{
			LogHandler: &transactionData.Log{
				Address: nil,
				Events: []*transactionData.Event{
					{
						Address:    addr1,
						Identifier: identifier1,
						Topics:     topic1,
						Data:       data1,
					},
					{
						Address:    []byte("addr4"),
						Identifier: identifier2,
						Topics:     topic1,
						Data:       data1,
					},
				},
			},
			TxHash: "",
		},
	}

	outgoingTxData, err := opFormatter.CreateOutgoingTxsData(logs)
	require.Nil(t, err)
	require.Equal(t, map[dtoCore.ChainID][]*dto.OutGoingOperation{
		dtoCore.MVX: {
			{
				Type: block.OutGoingOpDeposit,
				Data: operationBytes,
			},
		},
	}, outgoingTxData)
}

func TestOutgoingOperations_CreateOutgoingTxScCall(t *testing.T) {
	t.Parallel()

	addr := []byte("addr")
	identifier := []byte(topicIDDeposit)
	topics := [][]byte{
		[]byte(topicIDDeposit),
		[]byte("receiver"),
	}
	eventData := []byte("eventData")

	evData := &sovereign.EventData{
		Nonce: 1,
		TransferData: &sovereign.TransferData{
			GasLimit: 20000000,
			Function: []byte("add"),
			Args:     [][]byte{big.NewInt(20000000).Bytes()},
		},
	}

	operationBytes := []byte("operationBytes")

	dataCodec := &sovTests.DataCodecMock{
		DeserializeEventDataCalled: func(data []byte) (*sovereign.EventData, error) {
			require.Equal(t, eventData, data)
			return evData, nil
		},
		DeserializeTokenDataCalled: func(data []byte) (*sovereign.EsdtTokenData, error) {
			require.Fail(t, "DeserializeTokenData should not be called")
			return &sovereign.EsdtTokenData{}, nil
		},
		SerializeOperationCalled: func(operation sovereign.Operation) ([]byte, error) {
			require.Equal(t, topics[1], operation.Address)
			require.Equal(t, 0, len(operation.Tokens))
			require.Equal(t, evData, operation.Data)

			return operationBytes, nil
		},
	}

	events := []SubscribedEvent{
		{
			Identifier: identifier,
			Addresses: map[string]string{
				string(addr): string(addr),
			},
		},
	}

	args := ArgsOutgoingOperations{
		SubscribedEvents:  events,
		DataCodec:         dataCodec,
		TopicsChecker:     &sovTests.TopicsCheckerMock{},
		PeerAccountsDB:    &state.AccountsStub{},
		ChainNonceHandler: &sovTests.OutGoingChainNonceMock{},
	}
	opFormatter, _ := NewOutgoingOperationsFormatter(args)

	logs := []*data.LogData{
		{
			LogHandler: &transactionData.Log{
				Address: nil,
				Events: []*transactionData.Event{
					{
						Address:    addr,
						Identifier: identifier,
						Topics:     topics,
						Data:       eventData,
					},
				},
			},
			TxHash: "",
		},
	}

	outgoingTxData, err := opFormatter.CreateOutgoingTxsData(logs)
	require.Nil(t, err)
	require.Equal(t, map[dtoCore.ChainID][]*dto.OutGoingOperation{
		dtoCore.MVX: {
			{
				Type: block.OutGoingOpDeposit,
				Data: operationBytes,
			},
		},
	}, outgoingTxData)
}

func TestOutgoingOperations_CreateOutGoingChangeValidatorData(t *testing.T) {
	t.Parallel()

	args := createArgs()
	pubKeys := []string{"pk1", "pk2"}
	acc1 := &state.PeerAccountHandlerMock{
		MainChainID: []byte("id1"),
	}
	acc2 := &state.PeerAccountHandlerMock{
		MainChainID: []byte("id2"),
	}
	args.PeerAccountsDB = &state.AccountsStub{
		LoadAccountCalled: func(container []byte) (vmcommon.AccountHandler, error) {
			switch string(container) {
			case pubKeys[0]:
				return acc1, nil
			case pubKeys[1]:
				return acc2, nil
			}

			require.Fail(t, "should not load any other account")
			return nil, nil
		},
	}

	formatter, _ := NewOutgoingOperationsFormatter(args)

	res, err := formatter.CreateOutGoingChangeValidatorData(pubKeys, 4)
	require.Nil(t, err)

	resBridgeData := sovereign.BridgeOutGoingDataValidatorSetChange{}
	err = proto.Unmarshal(res, &resBridgeData)
	require.Nil(t, err)
	require.Equal(t, uint32(4), resBridgeData.GetEpoch())
	require.Equal(t, [][]byte{[]byte("id1"), []byte("id2")}, resBridgeData.GetPubKeyIDs())
}
