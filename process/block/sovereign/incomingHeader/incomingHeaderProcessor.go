package incomingHeader

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	logger "github.com/multiversx/mx-chain-logger-go"

	sovereignBlock "github.com/multiversx/mx-chain-go/dataRetriever/dataPool/sovereign"
	"github.com/multiversx/mx-chain-go/errors"
	sovBlock "github.com/multiversx/mx-chain-go/process/block/sovereign"
	"github.com/multiversx/mx-chain-go/process/block/sovereign/incomingHeader/dto"
	"github.com/multiversx/mx-chain-go/process/block/sovereign/incomingHeader/incomingEventsProc"
)

var log = logger.GetOrCreate("headerSubscriber")

// ArgsIncomingHeaderProcessor is a struct placeholder for args needed to create a new incoming header processor
type ArgsIncomingHeaderProcessor struct {
	HeadersPool            HeadersPool
	OutGoingOperationsPool sovereignBlock.OutGoingOperationsPool
	TxPool                 TransactionPool
	Marshaller             marshal.Marshalizer
	Hasher                 hashing.Hasher
	// TODO: Here we need to load string from config and convert to big Int
	MainChainNotarizationStartRound uint64
	DataCodec                       sovBlock.DataCodecHandler
	TopicsChecker                   sovBlock.TopicsCheckerHandler
}

type incomingHeaderProcessor struct {
	eventsProc         IncomingEventsProcessor
	extendedHeaderProc *extendedHeaderProcessor

	outGoingPool                    sovereignBlock.OutGoingOperationsPool
	mainChainNotarizationStartRound *big.Int
	preGenesisMainChainRound        *big.Int
}

// NewIncomingHeaderProcessor creates an incoming header processor which should be able to receive incoming headers and events
// from a chain to local sovereign chain. This handler will validate the events(using proofs in the future) and create
// incoming miniblocks and transaction(which will be added in pool) to be executed in sovereign shard.
func NewIncomingHeaderProcessor(args ArgsIncomingHeaderProcessor) (*incomingHeaderProcessor, error) {
	if check.IfNil(args.HeadersPool) {
		return nil, errNilHeadersPool
	}
	if check.IfNil(args.TxPool) {
		return nil, errNilTxPool
	}
	if check.IfNil(args.Marshaller) {
		return nil, core.ErrNilMarshalizer
	}
	if check.IfNil(args.Hasher) {
		return nil, core.ErrNilHasher
	}
	if check.IfNil(args.OutGoingOperationsPool) {
		return nil, errors.ErrNilOutGoingOperationsPool
	}
	if check.IfNil(args.DataCodec) {
		return nil, errors.ErrNilDataCodec
	}
	if check.IfNil(args.TopicsChecker) {
		return nil, errors.ErrNilTopicsChecker
	}

	incomingDepositOpArgs := incomingEventsProc.EventProcDepositOperationArgs{
		Marshaller:    args.Marshaller,
		Hasher:        args.Hasher,
		DataCodec:     args.DataCodec,
		TopicsChecker: args.TopicsChecker,
	}
	depositTokensProc, err := incomingEventsProc.NewEventProcDepositTokens(incomingDepositOpArgs)
	if err != nil {
		return nil, err
	}
	scCallProc, err := incomingEventsProc.NewEventProcSCCall(incomingDepositOpArgs)
	if err != nil {
		return nil, err
	}
	depositOpProc, err := incomingEventsProc.NewEventProcDepositOperation(depositTokensProc, scCallProc)
	if err != nil {
		return nil, err
	}

	confirmExecutedOperationProc := incomingEventsProc.NewEventProcConfirmExecutedOperation()
	executedOpProc, err := incomingEventsProc.NewEventProcExecutedDepositOperation(
		depositTokensProc,
		confirmExecutedOperationProc,
	)
	if err != nil {
		return nil, err
	}

	eventsProc := incomingEventsProc.NewIncomingEventsProcessor()
	err = eventsProc.RegisterProcessor(dto.EventIDDepositIncomingTransfer, depositOpProc)
	if err != nil {
		return nil, err
	}
	err = eventsProc.RegisterProcessor(dto.EventIDExecutedOutGoingBridgeOp, executedOpProc)
	if err != nil {
		return nil, err
	}
	err = eventsProc.RegisterProcessor(dto.EventIDChangeValidatorSet, confirmExecutedOperationProc)
	if err != nil {
		return nil, err
	}

	extendedHearProc, err := newExtendedHeaderProcessor(
		args.HeadersPool,
		args.TxPool,
		args.Marshaller,
		args.Hasher,
	)
	if err != nil {
		return nil, nil
	}

	log.Debug("NewIncomingHeaderProcessor", "starting round to notarize main chain headers", args.MainChainNotarizationStartRound)

	return &incomingHeaderProcessor{
		eventsProc:                      eventsProc,
		extendedHeaderProc:              extendedHearProc,
		outGoingPool:                    args.OutGoingOperationsPool,
		mainChainNotarizationStartRound: big.NewInt(int64(args.MainChainNotarizationStartRound)),
		preGenesisMainChainRound:        big.NewInt(int64(args.MainChainNotarizationStartRound) - 1),
	}, nil
}

// AddHeader will receive the incoming header, validate it, create incoming mbs and transactions and add them to pool
func (ihp *incomingHeaderProcessor) AddHeader(headerHash []byte, header sovereign.IncomingHeaderHandler) error {
	err := checkNilInputs(header)
	if err != nil {
		return err
	}

	incomingHeaderNonce := header.GetNonceBI()
	log.Info("received incoming header",
		"hash", hex.EncodeToString(headerHash),
		"nonce", incomingHeaderNonce.String(),
	)

	// pre-genesis header, needed to track/link genesis header on top of this one. Every node with an enabled notifier
	// will validate that the next genesis header with round == mainChainNotarizationStartRound is on top of pre-genesis header.
	// just save internal header to tracker, no need to process anything from it
	if incomingHeaderNonce.Cmp(ihp.preGenesisMainChainRound) == 0 {
		log.Debug("received pre-genesis header", "round", incomingHeaderNonce)
		return ihp.extendedHeaderProc.addPreGenesisExtendedHeaderToPool(header)
	}

	if incomingHeaderNonce.Cmp(ihp.mainChainNotarizationStartRound) < 0 {
		log.Debug("do not notarize incoming header, round lower than main chain notarization start round",
			"round", incomingHeaderNonce.String(),
			"start round", ihp.mainChainNotarizationStartRound)
		return nil
	}

	res, err := ihp.eventsProc.ProcessIncomingEvents(header.GetIncomingEventHandlers())
	if err != nil {
		return err
	}

	extendedHeader, err := ihp.extendedHeaderProc.createExtendedHeader(header, res.Scrs)
	if err != nil {
		return err
	}

	err = ihp.extendedHeaderProc.addExtendedHeaderAndSCRsToPool(extendedHeader, res.Scrs)
	if err != nil {
		return err
	}

	ihp.addConfirmedBridgeOpsToPool(res.ConfirmedBridgeOps)
	return nil
}

func checkNilInputs(header sovereign.IncomingHeaderHandler) error {
	if check.IfNil(header) {
		return data.ErrNilHeader
	}
	if header.GetProof() == nil {
		return errNilProof
	}
	if header.GetNonceBI() == nil {
		return fmt.Errorf("%w for nonce in incoming header", data.ErrNilValue)
	}

	return nil

}

func (ihp *incomingHeaderProcessor) addConfirmedBridgeOpsToPool(ops []*dto.ConfirmedBridgeOp) {
	for _, op := range ops {
		// This is not a critical error. This might just happen when a leader tries to re-send unconfirmed confirmation
		// that have been already executed, but the confirmation from notifier comes too late, and we receive a double
		// confirmation.
		err := ihp.outGoingPool.ConfirmOperation(op.HashOfHashes, op.Hash)
		if err != nil {
			log.Debug("incomingHeaderProcessor.AddHeader.addConfirmedBridgeOpsToPool",
				"error", err,
				"hashOfHashes", hex.EncodeToString(op.HashOfHashes),
				"hash", hex.EncodeToString(op.Hash),
			)
		}
	}
}

// CreateExtendedHeader will create an extended shard header with incoming scrs and mbs from the events of the received header
func (ihp *incomingHeaderProcessor) CreateExtendedHeader(header sovereign.IncomingHeaderHandler) (data.ShardHeaderExtendedHandler, error) {
	res, err := ihp.eventsProc.ProcessIncomingEvents(header.GetIncomingEventHandlers())
	if err != nil {
		return nil, err
	}

	return ihp.extendedHeaderProc.createExtendedHeader(header, res.Scrs)
}

// RegisterEventHandler will register an extra incoming event processor. For the registered processor, a subscription
// should be added to NotifierConfig.SubscribedEvents from sovereignConfig.toml
func (ihp *incomingHeaderProcessor) RegisterEventHandler(event string, proc incomingEventsProc.IncomingEventHandler) error {
	return ihp.eventsProc.RegisterProcessor(event, proc)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (ihp *incomingHeaderProcessor) IsInterfaceNil() bool {
	return ihp == nil
}
