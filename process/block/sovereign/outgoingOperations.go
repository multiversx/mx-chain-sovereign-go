package sovereign

import (
	"bytes"
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
	dtoSov "github.com/multiversx/mx-chain-core-go/data/sovereign/dto"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/sovereign/incomingHeader/dto"
	"github.com/multiversx/mx-chain-go/process/block/sovereign/operationFormatters"
	logger "github.com/multiversx/mx-chain-logger-go"
	"google.golang.org/protobuf/proto"

	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/state"
)

type outGoingOpsPerChain struct {
	opType block.OutGoingOpType
	data   map[dtoSov.ChainID][]byte
}

type opFormatterData struct {
	handler OperationFormatter
	opType  block.OutGoingOpType
}
type createOpFormatterHandler func(args ArgsOutgoingOperations) (OperationFormatter, block.OutGoingOpType, error)

const (
	topicIDDeposit          = "deposit"
	topicIDRegisterToken    = "registerToken"
	topicIDRegisterBlsKey   = "registerBlsKey"
	topicIDUnRegisterBlsKey = "unRegisterBlsKey"
)

var log = logger.GetOrCreate("outgoing-operations")

// SubscribedEvent contains a subscribed event from the sovereign chain needed to be transferred to the main chain
type SubscribedEvent struct {
	Identifier []byte
	Addresses  map[string]string
}

type ArgsOutgoingOperations struct {
	SubscribedEvents  []SubscribedEvent
	DataCodec         DataCodecHandler
	TopicsChecker     TopicsCheckerHandler
	PeerAccountsDB    state.AccountsAdapter
	ChainNonceHandler dto.OutGoingOpNonceChainHandler
}

type outgoingOperations struct {
	subscribedEvents []SubscribedEvent
	dataCodec        DataCodecHandler
	topicsChecker    TopicsCheckerHandler
	peerAccountsDB   state.AccountsAdapter

	opFormatters map[string]opFormatterData
}

// TODO: We should create a common base functionality from this component. Similar behavior is also found in
// mx-chain-sovereign-notifier-go in the sovereignNotifier.go file. This applies for the factory as well
// Task: MX-14721

// NewOutgoingOperationsFormatter creates an outgoing operations formatter
func NewOutgoingOperationsFormatter(args ArgsOutgoingOperations) (*outgoingOperations, error) {
	subscribedEvents, err := checkEvents(args.SubscribedEvents)
	if err != nil {
		return nil, err
	}
	err = checkNilArgs(args)
	if err != nil {
		return nil, err
	}

	opFormatters, err := createOpFormatterHandlers(subscribedEvents, args)
	if err != nil {
		return nil, err
	}

	return &outgoingOperations{
		subscribedEvents: args.SubscribedEvents,
		dataCodec:        args.DataCodec,
		topicsChecker:    args.TopicsChecker,
		peerAccountsDB:   args.PeerAccountsDB,
		opFormatters:     opFormatters,
	}, nil
}

func checkNilArgs(args ArgsOutgoingOperations) error {
	if check.IfNil(args.DataCodec) {
		return errors.ErrNilDataCodec
	}
	if check.IfNil(args.TopicsChecker) {
		return errors.ErrNilTopicsChecker
	}
	if check.IfNil(args.PeerAccountsDB) {
		return errors.ErrNilPeerAccounts
	}
	if check.IfNil(args.ChainNonceHandler) {
		return errors.ErrNilOutGoingOpNonceChainHandler
	}

	return nil
}

func checkEvents(events []SubscribedEvent) (map[string]struct{}, error) {
	if len(events) == 0 {
		return nil, errNoSubscribedEvent
	}

	subscribedEvents := make(map[string]struct{})

	log.Debug("sovereign outgoing operations creator: received config", "num subscribed events", len(events))
	for idx, event := range events {
		if len(event.Identifier) == 0 {
			return nil, fmt.Errorf("%w at event index = %d", errNoSubscribedIdentifier, idx)
		}

		log.Debug("sovereign outgoing operations creator", "subscribed event identifier", string(event.Identifier))

		err := checkEmptyAddresses(event.Addresses)
		if err != nil {
			return nil, fmt.Errorf("%w at event index = %d", err, idx)
		}

		subscribedEvents[string(event.Identifier)] = struct{}{}
	}

	return subscribedEvents, nil
}

func checkEmptyAddresses(addresses map[string]string) error {
	if len(addresses) == 0 {
		return errNoSubscribedAddresses
	}

	for decodedAddr, encodedAddr := range addresses {
		if len(decodedAddr) == 0 || len(encodedAddr) == 0 {
			return errNoSubscribedAddresses
		}

		log.Debug("sovereign outgoing operations creator", "subscribed address", encodedAddr)
	}

	return nil
}

func createOpFormatterHandlers(subscribedEvents map[string]struct{}, args ArgsOutgoingOperations) (map[string]opFormatterData, error) {
	handlers := make(map[string]opFormatterData)

	blsKeyOpFormatter, err := operationFormatters.NewRegisterValidatorOpFormatter(args.PeerAccountsDB, args.DataCodec, args.ChainNonceHandler)
	if err != nil {
		return nil, err
	}

	availableHandlers := map[string]createOpFormatterHandler{
		topicIDDeposit: func(args ArgsOutgoingOperations) (OperationFormatter, block.OutGoingOpType, error) {
			opFormatter, err := operationFormatters.NewDepositOpFormatter(args.DataCodec, args.TopicsChecker, args.ChainNonceHandler)
			return opFormatter, block.OutGoingOpDeposit, err
		},
		topicIDRegisterToken: func(args ArgsOutgoingOperations) (OperationFormatter, block.OutGoingOpType, error) {
			opFormatter, err := operationFormatters.NewRegisterTokenOpFormatter(args.DataCodec, args.ChainNonceHandler)
			return opFormatter, block.OutGoingOpRegisterToken, err
		},
		topicIDRegisterBlsKey: func(args ArgsOutgoingOperations) (OperationFormatter, block.OutGoingOpType, error) {
			return blsKeyOpFormatter, block.OutGoingOpRegisterBlsKey, nil
		},
		topicIDUnRegisterBlsKey: func(args ArgsOutgoingOperations) (OperationFormatter, block.OutGoingOpType, error) {
			return blsKeyOpFormatter, block.OutGoingOpUnRegisterBlsKey, nil
		},
	}

	for handlerID, handlerCreator := range availableHandlers {
		err := addHandlerIfSubscribed(
			handlerID,
			subscribedEvents,
			handlers,
			handlerCreator,
			args,
		)
		if err != nil {
			return nil, err
		}
	}

	if len(subscribedEvents) != 0 {
		return nil, fmt.Errorf("%w, event ids: %v", errUnsupportedEventType, subscribedEvents)
	}

	return handlers, nil
}

func addHandlerIfSubscribed(
	id string,
	subscribedEvents map[string]struct{},
	allHandlers map[string]opFormatterData,
	createOpFormatterHandlerFunc createOpFormatterHandler,
	args ArgsOutgoingOperations,
) error {
	_, found := subscribedEvents[id]
	if !found {
		return nil
	}

	opHandler, opType, err := createOpFormatterHandlerFunc(args)
	if err != nil {
		return err
	}

	allHandlers[id] = opFormatterData{
		handler: opHandler,
		opType:  opType,
	}

	delete(subscribedEvents, id)
	return nil
}

// CreateOutgoingTxsData collects relevant outgoing events(based on subscribed addresses and topics) for bridge from the
// logs and creates outgoing data that needs to be signed by validators to bridge tokens
func (op *outgoingOperations) CreateOutgoingTxsData(logs []*data.LogData) (map[dtoSov.ChainID][]*dto.OutGoingOperation, error) {
	operations := make(map[dtoSov.ChainID][]*dto.OutGoingOperation)

	outgoingEvents := op.collectOutGoingEvents(logs)
	if len(outgoingEvents) == 0 {
		return operations, nil
	}

	for i, event := range outgoingEvents {
		operation, err := op.getOperationData(event)
		if err != nil {
			log.Error("outgoingOperations.CreateOutgoingTxsData error",
				"tx hash", logs[i].TxHash,
				"event", string(event.GetIdentifier()),
				"error", err)

			return nil, err
		}

		for chainID, opData := range operation.data {
			operations[chainID] = append(operations[chainID], &dto.OutGoingOperation{
				Type: operation.opType,
				Data: opData,
			})
		}
	}

	// TODO: Check gas limit here and split tx data in multiple batches if required
	// Task: MX-14720
	return operations, nil
}

func (op *outgoingOperations) collectOutGoingEvents(logs []*data.LogData) []data.EventHandler {
	events := make([]data.EventHandler, 0)

	for _, logData := range logs {
		eventsFromLog := op.createOutgoingEvent(logData)
		events = append(events, eventsFromLog...)
	}

	return events
}

func (op *outgoingOperations) createOutgoingEvent(logData *data.LogData) []data.EventHandler {
	events := make([]data.EventHandler, 0)

	for _, event := range logData.GetLogEvents() {
		if !op.isSubscribed(event, logData.TxHash) {
			continue
		}

		events = append(events, event)
	}

	return events
}

func (op *outgoingOperations) isSubscribed(event data.EventHandler, txHash string) bool {
	for _, subEvent := range op.subscribedEvents {
		if !bytes.Equal(event.GetIdentifier(), subEvent.Identifier) {
			continue
		}

		receiver := event.GetAddress()
		encodedAddr, found := subEvent.Addresses[string(receiver)]
		if !found {
			continue
		}

		log.Trace("found outgoing event", "original tx hash", txHash, "receiver", encodedAddr)
		return true
	}

	return false
}

func (op *outgoingOperations) getOperationData(event data.EventHandler) (*outGoingOpsPerChain, error) {
	opFormatter, found := op.opFormatters[string(event.GetIdentifier())]
	if !found {
		log.Error("outgoingOperations.getOperationData: event not found", "event", string(event.GetIdentifier()))
		return nil, errEventIDNotFound
	}

	opData, err := opFormatter.handler.CreateOperationData(event)
	return &outGoingOpsPerChain{
		opType: opFormatter.opType,
		data:   opData,
	}, err
}

// CreateOutGoingChangeValidatorData will create the necessary outgoing data for validator set change
func (op *outgoingOperations) CreateOutGoingChangeValidatorData(pubKeys []string, epoch uint32) ([]byte, error) {
	validatorsID := make([][]byte, len(pubKeys))

	for idx, pubKey := range pubKeys {
		peerAcc, err := process.GetPeerAccount([]byte(pubKey), op.peerAccountsDB)
		if err != nil {
			return nil, err
		}

		validatorsID[idx] = peerAcc.GetMainChainID()
	}

	return proto.Marshal(&sovereign.BridgeOutGoingDataValidatorSetChange{
		Epoch:     epoch,
		PubKeyIDs: validatorsID,
	})
}

// IsInterfaceNil checks if the underlying pointer is nil
func (op *outgoingOperations) IsInterfaceNil() bool {
	return op == nil
}
