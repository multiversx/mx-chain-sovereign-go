package incomingEventsProc

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"

	"github.com/multiversx/mx-chain-go/process/block/sovereign/incomingHeader/dto"
)

type eventProcScCall struct {
	*eventProcDepositTokens
}

// NewEventProcScCall creates a new event processor for sc call operations
func NewEventProcScCall(args EventProcDepositOperationArgs) (*eventProcScCall, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	return &eventProcScCall{
		&eventProcDepositTokens{
			marshaller:    args.Marshaller,
			hasher:        args.Hasher,
			dataCodec:     args.DataCodec,
			topicsChecker: args.TopicsChecker,
		},
	}, nil
}

func (ep *eventProcScCall) ProcessEvent(event data.EventHandler) (*dto.EventResult, error) {
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

func (ep *eventProcScCall) createSCRData(eventData *sovereign.EventData) ([]byte, uint64) {
	scrData := eventData.TransferData.Function
	scrData = append(scrData, extractArguments(eventData.TransferData.Args)...)
	return scrData, eventData.TransferData.GasLimit
}
