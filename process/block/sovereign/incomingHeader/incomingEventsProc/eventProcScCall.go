package incomingEventsProc

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"

	sovBlock "github.com/multiversx/mx-chain-go/process/block/sovereign"
	"github.com/multiversx/mx-chain-go/process/block/sovereign/incomingHeader/dto"
)

type eventProcSCCall struct {
	marshaller    marshal.Marshalizer
	hasher        hashing.Hasher
	dataCodec     sovBlock.DataCodecHandler
	topicsChecker sovBlock.TopicsCheckerHandler
}

// NewEventProcSCCall creates a new event processor for sc call operations
func NewEventProcSCCall(args EventProcDepositOperationArgs) (*eventProcSCCall, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	return &eventProcSCCall{
		marshaller:    args.Marshaller,
		hasher:        args.Hasher,
		dataCodec:     args.DataCodec,
		topicsChecker: args.TopicsChecker,
	}, nil
}

// ProcessEvent handles incoming SC call events and returns the corresponding incoming SCR info.
// Each SC call event is identified by dto.EventIDDepositIncomingTransfer.
//
// Expected event data:
// - Data []byte – Serialized event details (nonce, gas, function call with arguments).
// - Topics [][]byte – A list of two topics, where:
//   - topic[0] = dto.TopicIDSCCall.
//   - topic[1] = Receiver address.
func (ep *eventProcSCCall) ProcessEvent(event data.EventHandler) (*dto.EventResult, error) {
	evData, err := ep.dataCodec.DeserializeEventData(event.GetData())
	if err != nil {
		return nil, err
	}

	topics := event.GetTopics()
	err = ep.topicsChecker.CheckValidity(topics, evData.TransferData)
	if err != nil {
		return nil, err
	}

	scrData, gasLimit := ep.createSCRData(evData)

	scr := &smartContractResult.SmartContractResult{
		Nonce:          evData.Nonce,
		OriginalTxHash: nil, // TODO:  Implement this in MX-14321 task
		RcvAddr:        topics[1],
		SndAddr:        core.ESDTSCAddress,
		Data:           scrData,
		Value:          big.NewInt(0),
		GasLimit:       gasLimit,
	}

	hash, err := core.CalculateHash(ep.marshaller, ep.hasher, scr)
	if err != nil {
		return nil, err
	}

	return &dto.EventResult{
		SCR: &dto.SCRInfo{
			SCR:  scr,
			Hash: hash,
		},
	}, nil
}

func (ep *eventProcSCCall) createSCRData(eventData *sovereign.EventData) ([]byte, uint64) {
	scrData := eventData.TransferData.Function
	scrData = append(scrData, extractArguments(eventData.TransferData.Args)...)
	return scrData, eventData.TransferData.GasLimit
}

// IsInterfaceNil checks if the underlying pointer is nil
func (ep *eventProcSCCall) IsInterfaceNil() bool {
	return ep == nil
}
