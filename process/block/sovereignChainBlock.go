package block

import (
	"bytes"
	"fmt"
	"math/big"
	"sort"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	sovCore "github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/hashing"
	logger "github.com/multiversx/mx-chain-logger-go"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/holders"
	"github.com/multiversx/mx-chain-go/common/logging"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	sovereignBlock "github.com/multiversx/mx-chain-go/dataRetriever/dataPool/sovereign"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/processedMb"
	"github.com/multiversx/mx-chain-go/process/block/sovereign"
	"github.com/multiversx/mx-chain-go/state"
)

var rootHash = "uncomputed root hash"

type extendedShardHeaderTrackHandler interface {
	ComputeLongestExtendedShardChainFromLastNotarized() ([]data.HeaderHandler, [][]byte, error)
	IsGenesisLastCrossNotarizedHeader() bool
	RemoveLastCrossNotarizedHeaders()
	RemoveLastSelfNotarizedHeaders()
}

type extendedShardHeaderRequestHandler interface {
	RequestExtendedShardHeaderByNonce(nonce uint64)
	RequestExtendedShardHeader(hash []byte)
}

type sovereignChainBlockProcessor struct {
	*shardProcessor
	validatorStatisticsProcessor process.ValidatorStatisticsProcessor
	uncomputedRootHash           []byte
	extendedShardHeaderTracker   extendedShardHeaderTrackHandler
	extendedShardHeaderRequester extendedShardHeaderRequestHandler
	chRcvAllExtendedShardHdrs    chan bool
	outgoingOperationsFormatter  sovereign.OutgoingOperationsFormatter
	outGoingOperationsPool       sovereignBlock.OutGoingOperationsPool
	operationsHasher             hashing.Hasher

	epochStartDataCreator process.EpochStartDataCreator
	epochRewardsCreator   process.RewardsCreator
	validatorInfoCreator  process.EpochStartValidatorInfoCreator
	scToProtocol          process.SmartContractToProtocolHandler
	epochEconomics        process.EndOfEpochEconomics

	mainChainNotarizationStartRound uint64
}

// ArgsSovereignChainBlockProcessor is a struct placeholder for args needed to create a new sovereign chain block processor
type ArgsSovereignChainBlockProcessor struct {
	ShardProcessor                  *shardProcessor
	ValidatorStatisticsProcessor    process.ValidatorStatisticsProcessor
	OutgoingOperationsFormatter     sovereign.OutgoingOperationsFormatter
	OutGoingOperationsPool          sovereignBlock.OutGoingOperationsPool
	OperationsHasher                hashing.Hasher
	EpochStartDataCreator           process.EpochStartDataCreator
	EpochRewardsCreator             process.RewardsCreator
	ValidatorInfoCreator            process.EpochStartValidatorInfoCreator
	EpochSystemSCProcessor          process.EpochStartSystemSCProcessor
	SCToProtocol                    process.SmartContractToProtocolHandler
	EpochEconomics                  process.EndOfEpochEconomics
	MainChainNotarizationStartRound uint64
}

// NewSovereignChainBlockProcessor creates a new sovereign chain block processor
func NewSovereignChainBlockProcessor(args ArgsSovereignChainBlockProcessor) (*sovereignChainBlockProcessor, error) {
	if check.IfNil(args.ShardProcessor) {
		return nil, process.ErrNilBlockProcessor
	}
	if check.IfNil(args.ValidatorStatisticsProcessor) {
		return nil, process.ErrNilValidatorStatistics
	}
	if check.IfNil(args.OutgoingOperationsFormatter) {
		return nil, errors.ErrNilOutgoingOperationsFormatter
	}
	if check.IfNil(args.OutGoingOperationsPool) {
		return nil, errors.ErrNilOutGoingOperationsPool
	}
	if check.IfNil(args.OperationsHasher) {
		return nil, errors.ErrNilOperationsHasher
	}
	if check.IfNil(args.EpochStartDataCreator) {
		return nil, process.ErrNilEpochStartDataCreator
	}
	if check.IfNil(args.EpochRewardsCreator) {
		return nil, process.ErrNilRewardsCreator
	}
	if check.IfNil(args.ValidatorInfoCreator) {
		return nil, process.ErrNilEpochStartValidatorInfoCreator
	}
	if check.IfNil(args.EpochSystemSCProcessor) {
		return nil, process.ErrNilEpochStartSystemSCProcessor
	}
	if check.IfNil(args.EpochEconomics) {
		return nil, process.ErrNilEpochEconomics
	}

	scbp := &sovereignChainBlockProcessor{
		shardProcessor:                  args.ShardProcessor,
		validatorStatisticsProcessor:    args.ValidatorStatisticsProcessor,
		outgoingOperationsFormatter:     args.OutgoingOperationsFormatter,
		outGoingOperationsPool:          args.OutGoingOperationsPool,
		operationsHasher:                args.OperationsHasher,
		epochStartDataCreator:           args.EpochStartDataCreator,
		epochRewardsCreator:             args.EpochRewardsCreator,
		validatorInfoCreator:            args.ValidatorInfoCreator,
		scToProtocol:                    args.SCToProtocol,
		epochEconomics:                  args.EpochEconomics,
		mainChainNotarizationStartRound: args.MainChainNotarizationStartRound,
	}

	scbp.baseProcessor.epochSystemSCProcessor = args.EpochSystemSCProcessor

	scbp.uncomputedRootHash = scbp.hasher.Compute(rootHash)

	extendedShardHeaderTracker, ok := scbp.blockTracker.(extendedShardHeaderTrackHandler)
	if !ok {
		return nil, fmt.Errorf("%w in NewSovereignBlockProcessor for extendedShardHeaderTracker", process.ErrWrongTypeAssertion)
	}

	scbp.extendedShardHeaderTracker = extendedShardHeaderTracker

	extendedShardHeaderRequester, ok := scbp.requestHandler.(extendedShardHeaderRequestHandler)
	if !ok {
		return nil, fmt.Errorf("%w in NewSovereignChainBlockProcessor for extendedShardHeaderRequester", process.ErrWrongTypeAssertion)
	}

	scbp.extendedShardHeaderRequester = extendedShardHeaderRequester

	scbp.chRcvAllExtendedShardHdrs = make(chan bool)

	headersPool := scbp.dataPool.Headers()
	headersPool.RegisterHandler(scbp.receivedExtendedShardHeader)

	scbp.requestMissingHeadersFunc = scbp.requestMissingHeaders
	scbp.cleanupPoolsForCrossShardFunc = scbp.cleanupPoolsForCrossShard
	scbp.cleanupBlockTrackerPoolsForShardFunc = scbp.cleanupBlockTrackerPoolsForShard
	scbp.getExtraMissingNoncesToRequestFunc = scbp.getExtraMissingNoncesToRequest
	scbp.blockProcessor = scbp

	scbp.crossNotarizer = &sovereignShardCrossNotarizer{
		baseBlockNotarizer: &baseBlockNotarizer{
			blockTracker: scbp.blockTracker,
		},
	}

	return scbp, nil
}

// CreateNewHeader creates a new header
func (scbp *sovereignChainBlockProcessor) CreateNewHeader(round uint64, nonce uint64) (data.HeaderHandler, error) {
	scbp.epochStartTrigger.Update(round, nonce)
	scbp.enableRoundsHandler.RoundConfirmed(round, 0)
	epoch := scbp.epochStartTrigger.MetaEpoch()

	header := scbp.versionedHeaderFactory.Create(epoch)

	err := scbp.setRoundNonceInitFees(round, nonce, header)
	if err != nil {
		return nil, err
	}

	err = scbp.setSovHeaderInitFields(header)
	if err != nil {
		return nil, err
	}

	return header, nil
}

func (scbp *sovereignChainBlockProcessor) setSovHeaderInitFields(header data.HeaderHandler) error {
	sovHdr, castOk := header.(data.SovereignChainHeaderHandler)
	if !castOk {
		return fmt.Errorf("%w in sovereignChainBlockProcessor.setSovHeaderInitFields", process.ErrWrongTypeAssertion)
	}

	err := sovHdr.SetValidatorStatsRootHash(scbp.uncomputedRootHash)
	if err != nil {
		return err
	}

	err = sovHdr.SetRootHash(scbp.uncomputedRootHash)
	if err != nil {
		return err
	}

	err = sovHdr.SetDeveloperFees(big.NewInt(0))
	if err != nil {
		return err
	}

	return sovHdr.SetAccumulatedFeesInEpoch(big.NewInt(0))
}

// CreateBlock selects and puts transaction into the temporary block body
func (scbp *sovereignChainBlockProcessor) CreateBlock(initialHdr data.HeaderHandler, haveTime func() bool) (data.HeaderHandler, data.BodyHandler, error) {
	if check.IfNil(initialHdr) {
		return nil, nil, process.ErrNilBlockHeader
	}

	sovereignChainHeaderHandler, ok := initialHdr.(data.SovereignChainHeaderHandler)
	if !ok {
		return nil, nil, fmt.Errorf("%w in sovereignChainBlockProcessor.CreateBlock", process.ErrWrongTypeAssertion)
	}

	scbp.processStatusHandler.SetBusy("sovereignChainBlockProcessor.CreateBlock")
	defer scbp.processStatusHandler.SetIdle()

	for _, accounts := range scbp.accountsDB {
		if accounts.JournalLen() != 0 {
			log.Error("sovereignChainBlockProcessor.CreateBlock first entry", "stack", accounts.GetStackDebugFirstEntry())
			return nil, nil, process.ErrAccountStateDirty
		}
	}

	err := scbp.createBlockStarted()
	if err != nil {
		return nil, nil, err
	}

	if scbp.epochStartTrigger.IsEpochStart() {
		epoch := scbp.epochStartTrigger.MetaEpoch()
		log.Debug("sovereignChainBlockProcessor.CreateBlock", "isEpochStart", true, "epoch from epoch start trigger", epoch)
		err = initialHdr.SetEpoch(epoch)
		if err != nil {
			return nil, nil, err
		}

		err = sovereignChainHeaderHandler.SetStartOfEpochHeader()
		if err != nil {
			return nil, nil, err
		}

		err = scbp.createEpochStartDataCrossChain(sovereignChainHeaderHandler)
		if err != nil {
			return nil, nil, err
		}

		scbp.blockChainHook.SetCurrentHeader(initialHdr)
		scbp.requestHandler.SetEpoch(initialHdr.GetEpoch())
		return initialHdr, &block.Body{}, nil
	}

	err = initialHdr.SetEpoch(scbp.epochStartTrigger.Epoch())
	if err != nil {
		return nil, nil, err
	}

	scbp.blockChainHook.SetCurrentHeader(initialHdr)

	crossMiniblocks, miniBlocks, err := scbp.createAllMiniBlocks(haveTime, initialHdr)
	if err != nil {
		return nil, nil, err
	}

	extendedShardHeaderHashes := scbp.sortExtendedShardHeaderHashesForCurrentBlockByNonce()
	_, crossMiniblockHeaders, err := scbp.createMiniBlockHeaderHandlers(crossMiniblocks)
	if err != nil {
		return nil, nil, err
	}

	err = sovereignChainHeaderHandler.SetExtendedShardHeaderHashes(extendedShardHeaderHashes)
	if err != nil {
		return nil, nil, err
	}

	err = sovereignChainHeaderHandler.SetMiniBlockHeaderHandlers(crossMiniblockHeaders)
	if err != nil {
		return nil, nil, err
	}

	scbp.requestHandler.SetEpoch(initialHdr.GetEpoch())

	return initialHdr, &block.Body{MiniBlocks: miniBlocks}, nil
}

// We should call this func only on ProcessBlock for all participants.
// No need to call it on CreateBlock for epoch start processing.
func (scbp *sovereignChainBlockProcessor) updateEpochStartHeader(header data.SovereignChainHeaderHandler) (*block.Economics, error) {
	sovHeaderHandler, castOk := header.(*block.SovereignChainHeader)
	if !castOk {
		return nil, fmt.Errorf("%w in sovereignChainBlockProcessor.updateEpochStartHeader", process.ErrWrongTypeAssertion)
	}

	sw := core.NewStopWatch()
	sw.Start("createEpochStartForSovereignBlock")
	defer func() {
		sw.Stop("createEpochStartForSovereignBlock")
		log.Debug("epochStartHeaderDataCreation", sw.GetMeasurements()...)
	}()

	totalAccumulatedFeesInEpoch := big.NewInt(0)
	totalDevFeesInEpoch := big.NewInt(0)
	currentHeader := scbp.blockChain.GetCurrentBlockHeader()
	if !check.IfNil(currentHeader) && !currentHeader.IsStartOfEpochBlock() {
		prevSovHdr, ok := currentHeader.(data.MetaHeaderHandler)
		if !ok {
			return nil, fmt.Errorf("%w in sovereignChainBlockProcessor.updateEpochStartHeader when checking prevSovHdr", process.ErrWrongTypeAssertion)
		}
		totalAccumulatedFeesInEpoch = big.NewInt(0).Set(prevSovHdr.GetAccumulatedFeesInEpoch())
		totalDevFeesInEpoch = big.NewInt(0).Set(prevSovHdr.GetDevFeesInEpoch())
	}

	err := sovHeaderHandler.SetAccumulatedFeesInEpoch(totalAccumulatedFeesInEpoch)
	if err != nil {
		return nil, err
	}

	err = sovHeaderHandler.SetDevFeesInEpoch(totalDevFeesInEpoch)
	if err != nil {
		return nil, err
	}

	economicsData, err := scbp.epochEconomics.ComputeEndOfEpochEconomics(sovHeaderHandler)
	if err != nil {
		return nil, err
	}

	sovHeaderHandler.EpochStart.Economics = *economicsData

	// do not call saveEpochStartEconomicsMetrics here as in metachain code, it will be called later
	return economicsData, nil
}

func (scbp *sovereignChainBlockProcessor) createAllMiniBlocks(
	haveTime func() bool,
	initialHdr data.HeaderHandler,
) (block.MiniBlockSlice, block.MiniBlockSlice, error) {
	var miniBlocks block.MiniBlockSlice
	if !haveTime() {
		log.Debug("sovereignChainBlockProcessor.CreateBlock", "error", process.ErrTimeIsOut)
		return nil, nil, process.ErrTimeIsOut
	}

	startTime := time.Now()
	createIncomingMiniBlocksDestMeInfo, err := scbp.createIncomingMiniBlocksDestMe(haveTime)
	elapsedTime := time.Since(startTime)
	log.Debug("elapsed time to create mbs to me", "time", elapsedTime)
	if err != nil {
		return nil, nil, err
	}

	numCrossMbs := len(createIncomingMiniBlocksDestMeInfo.miniBlocks)
	crossMiniBlocks := make(block.MiniBlockSlice, numCrossMbs)
	if numCrossMbs > 0 {
		miniBlocks = append(miniBlocks, createIncomingMiniBlocksDestMeInfo.miniBlocks...)

		log.Debug("created mini blocks and txs with destination in self shard",
			"num mini blocks", len(createIncomingMiniBlocksDestMeInfo.miniBlocks),
			"num txs", createIncomingMiniBlocksDestMeInfo.numTxsAdded,
			"num extended shard headers", createIncomingMiniBlocksDestMeInfo.numHdrsAdded)
		copy(crossMiniBlocks, createIncomingMiniBlocksDestMeInfo.miniBlocks)
	}

	startTime = time.Now()
	mbsFromMe := scbp.txCoordinator.CreateMbsAndProcessTransactionsFromMe(haveTime, initialHdr.GetPrevRandSeed())
	elapsedTime = time.Since(startTime)
	log.Debug("elapsed time to create mbs from me", "time", elapsedTime)

	if len(mbsFromMe) > 0 {
		miniBlocks = append(miniBlocks, mbsFromMe...)

		numTxs := 0
		for _, mb := range mbsFromMe {
			numTxs += len(mb.TxHashes)
		}

		log.Debug("processed miniblocks and txs from self shard",
			"num miniblocks", len(mbsFromMe),
			"num txs", numTxs)
	}

	return crossMiniBlocks, miniBlocks, nil
}

func (scbp *sovereignChainBlockProcessor) createIncomingMiniBlocksDestMe(haveTime func() bool) (*createAndProcessMiniBlocksDestMeInfo, error) {
	log.Debug("createIncomingMiniBlocksDestMe has been started")

	sw := core.NewStopWatch()
	sw.Start("ComputeLongestExtendedShardChainFromLastNotarized")
	orderedExtendedShardHeaders, orderedExtendedShardHeadersHashes, err := scbp.extendedShardHeaderTracker.ComputeLongestExtendedShardChainFromLastNotarized()
	sw.Stop("ComputeLongestExtendedShardChainFromLastNotarized")
	log.Debug("measurements", sw.GetMeasurements()...)

	if err != nil {
		return nil, err
	}

	log.Debug("extended shard headers ordered",
		"num extended shard headers", len(orderedExtendedShardHeaders),
	)

	lastExtendedShardHdr, _, err := scbp.blockTracker.GetLastCrossNotarizedHeader(core.MainChainShardId)
	if err != nil {
		return nil, err
	}

	haveAdditionalTimeFalse := func() bool {
		return false
	}

	createAndProcessInfo := &createAndProcessMiniBlocksDestMeInfo{
		haveTime:                   haveTime,
		haveAdditionalTime:         haveAdditionalTimeFalse,
		miniBlocks:                 make(block.MiniBlockSlice, 0),
		allProcessedMiniBlocksInfo: make(map[string]*processedMb.ProcessedMiniBlockInfo),
		numTxsAdded:                uint32(0),
		numHdrsAdded:               uint32(0),
		scheduledMode:              true,
	}

	// do processing in order
	scbp.hdrsForCurrBlock.mutHdrsForBlock.Lock()
	for i := 0; i < len(orderedExtendedShardHeadersHashes); i++ {
		if !createAndProcessInfo.haveTime() && !createAndProcessInfo.haveAdditionalTime() {
			log.Debug("time is up in creating incoming mini blocks destination me",
				"scheduled mode", createAndProcessInfo.scheduledMode,
				"num txs added", createAndProcessInfo.numTxsAdded,
			)
			break
		}

		if createAndProcessInfo.numHdrsAdded >= process.MaxExtendedShardHeadersAllowedInOneSovereignBlock {
			log.Debug("maximum extended shard headers allowed to be included in one sovereign block has been reached",
				"scheduled mode", createAndProcessInfo.scheduledMode,
				"extended shard headers added", createAndProcessInfo.numHdrsAdded,
			)
			break
		}

		extendedShardHeader, ok := orderedExtendedShardHeaders[i].(data.ShardHeaderExtendedHandler)
		if !ok {
			log.Debug("wrong type assertion from data.HeaderHandler to data.ShardHeaderExtendedHandler",
				"hash", orderedExtendedShardHeadersHashes[i],
				"shard", orderedExtendedShardHeaders[i].GetShardID(),
				"round", orderedExtendedShardHeaders[i].GetRound(),
				"nonce", orderedExtendedShardHeaders[i].GetNonce())
			break
		}

		createAndProcessInfo.currentHeader = orderedExtendedShardHeaders[i]
		if createAndProcessInfo.currentHeader.GetNonce() > lastExtendedShardHdr.GetNonce()+1 {
			log.Debug("skip searching",
				"scheduled mode", createAndProcessInfo.scheduledMode,
				"last extended shard hdr nonce", lastExtendedShardHdr.GetNonce(),
				"curr extended shard hdr nonce", createAndProcessInfo.currentHeader.GetNonce())
			break
		}

		createAndProcessInfo.currentHeaderHash = orderedExtendedShardHeadersHashes[i]
		if len(extendedShardHeader.GetIncomingMiniBlockHandlers()) == 0 {
			scbp.hdrsForCurrBlock.hdrHashAndInfo[string(createAndProcessInfo.currentHeaderHash)] = &hdrInfo{hdr: createAndProcessInfo.currentHeader, usedInBlock: true}
			createAndProcessInfo.numHdrsAdded++
			lastExtendedShardHdr = createAndProcessInfo.currentHeader
			continue
		}

		createAndProcessInfo.currProcessedMiniBlocksInfo = scbp.processedMiniBlocksTracker.GetProcessedMiniBlocksInfo(createAndProcessInfo.currentHeaderHash)
		createAndProcessInfo.hdrAdded = false

		shouldContinue, errCreated := scbp.createIncomingMiniBlocksAndTransactionsDestMe(createAndProcessInfo)
		if errCreated != nil {
			return nil, errCreated
		}
		if !shouldContinue {
			break
		}

		lastExtendedShardHdr = createAndProcessInfo.currentHeader
	}
	scbp.hdrsForCurrBlock.mutHdrsForBlock.Unlock()

	go scbp.requestExtendedShardHeadersIfNeeded(createAndProcessInfo.numHdrsAdded, lastExtendedShardHdr)

	for _, miniBlock := range createAndProcessInfo.miniBlocks {
		log.Debug("mini block info",
			"type", miniBlock.Type,
			"sender shard", miniBlock.SenderShardID,
			"receiver shard", miniBlock.ReceiverShardID,
			"txs added", len(miniBlock.TxHashes))
	}

	log.Debug("createIncomingMiniBlocksDestMe has been finished",
		"num txs added", createAndProcessInfo.numTxsAdded,
		"num hdrs added", createAndProcessInfo.numHdrsAdded)

	return createAndProcessInfo, nil
}

func (scbp *sovereignChainBlockProcessor) createIncomingMiniBlocksAndTransactionsDestMe(
	createAndProcessInfo *createAndProcessMiniBlocksDestMeInfo,
) (bool, error) {
	currMiniBlocksAdded, currNumTxsAdded, hdrProcessFinished, errCreated := scbp.txCoordinator.CreateMbsAndProcessCrossShardTransactionsDstMe(
		createAndProcessInfo.currentHeader,
		createAndProcessInfo.currProcessedMiniBlocksInfo,
		createAndProcessInfo.haveTime,
		createAndProcessInfo.haveAdditionalTime,
		createAndProcessInfo.scheduledMode)
	if errCreated != nil {
		return false, errCreated
	}

	for miniBlockHash, processedMiniBlockInfo := range createAndProcessInfo.currProcessedMiniBlocksInfo {
		createAndProcessInfo.allProcessedMiniBlocksInfo[miniBlockHash] = &processedMb.ProcessedMiniBlockInfo{
			FullyProcessed:         processedMiniBlockInfo.FullyProcessed,
			IndexOfLastTxProcessed: processedMiniBlockInfo.IndexOfLastTxProcessed,
		}
	}

	// all txs processed, add to processed miniblocks
	createAndProcessInfo.miniBlocks = append(createAndProcessInfo.miniBlocks, currMiniBlocksAdded...)
	createAndProcessInfo.numTxsAdded += currNumTxsAdded

	if !createAndProcessInfo.hdrAdded && currNumTxsAdded > 0 {
		scbp.hdrsForCurrBlock.hdrHashAndInfo[string(createAndProcessInfo.currentHeaderHash)] = &hdrInfo{
			hdr:         createAndProcessInfo.currentHeader,
			usedInBlock: true,
		}
		createAndProcessInfo.numHdrsAdded++
		createAndProcessInfo.hdrAdded = true
	}

	if !hdrProcessFinished {
		log.Debug("extended shard header cannot be fully processed",
			"scheduled mode", createAndProcessInfo.scheduledMode,
			"round", createAndProcessInfo.currentHeader.GetRound(),
			"nonce", createAndProcessInfo.currentHeader.GetNonce(),
			"hash", createAndProcessInfo.currentHeaderHash,
			"num mbs added", len(currMiniBlocksAdded),
			"num txs added", currNumTxsAdded)

		return false, nil
	}

	return true, nil
}

func (scbp *sovereignChainBlockProcessor) requestExtendedShardHeadersIfNeeded(hdrsAdded uint32, lastExtendedShardHdr data.HeaderHandler) {
	log.Debug("extended shard headers added",
		"num", hdrsAdded,
		"highest nonce", lastExtendedShardHdr.GetNonce(),
	)
	//TODO: A request mechanism should be implemented if extended shard header(s) is(are) needed
}

func (scbp *sovereignChainBlockProcessor) sortExtendedShardHeaderHashesForCurrentBlockByNonce() [][]byte {
	hdrsForCurrentBlockInfo := make([]*nonceAndHashInfo, 0)

	scbp.hdrsForCurrBlock.mutHdrsForBlock.RLock()
	for headerHash, headerInfo := range scbp.hdrsForCurrBlock.hdrHashAndInfo {
		hdrsForCurrentBlockInfo = append(hdrsForCurrentBlockInfo, &nonceAndHashInfo{
			nonce: headerInfo.hdr.GetNonce(),
			hash:  []byte(headerHash),
		})
	}
	scbp.hdrsForCurrBlock.mutHdrsForBlock.RUnlock()

	if len(hdrsForCurrentBlockInfo) > 1 {
		sort.Slice(hdrsForCurrentBlockInfo, func(i, j int) bool {
			return hdrsForCurrentBlockInfo[i].nonce < hdrsForCurrentBlockInfo[j].nonce
		})
	}

	hdrsHashesForCurrentBlock := make([][]byte, len(hdrsForCurrentBlockInfo))
	for index, hdrForCurrentBlockInfo := range hdrsForCurrentBlockInfo {
		hdrsHashesForCurrentBlock[index] = hdrForCurrentBlockInfo.hash
	}

	return hdrsHashesForCurrentBlock
}

func (scbp *sovereignChainBlockProcessor) createMiniBlockHeaderHandlers(miniBlocks block.MiniBlockSlice) (int, []data.MiniBlockHeaderHandler, error) {
	miniBlockHeaders := make([]data.MiniBlockHeaderHandler, 0)

	totalTxCount := 0
	for _, mb := range miniBlocks {
		txCount := len(mb.TxHashes)
		totalTxCount += txCount

		hash, err := core.CalculateHash(scbp.marshalizer, scbp.hasher, mb)
		if err != nil {
			return 0, nil, err
		}

		mbHeader := &block.MiniBlockHeader{
			Hash:            hash,
			SenderShardID:   mb.SenderShardID,
			ReceiverShardID: mb.ReceiverShardID,
			TxCount:         uint32(txCount),
			Type:            mb.Type,
			Reserved:        mb.Reserved,
		}

		miniBlockHeaders = append(miniBlockHeaders, mbHeader)
	}

	return totalTxCount, miniBlockHeaders, nil
}

// receivedExtendedShardHeader is a callback function when a new extended shard header was received
func (scbp *sovereignChainBlockProcessor) receivedExtendedShardHeader(headerHandler data.HeaderHandler, extendedShardHeaderHash []byte) {
	extendedShardHeader, ok := headerHandler.(*block.ShardHeaderExtended)
	if !ok {
		return
	}

	log.Trace("received extended shard header from network",
		"round", extendedShardHeader.GetRound(),
		"nonce", extendedShardHeader.GetNonce(),
		"hash", extendedShardHeaderHash,
	)

	scbp.checkAndSetMissingExtendedHeaders(extendedShardHeader, extendedShardHeaderHash)
	go scbp.requestIncomingTxsIfNeeded(extendedShardHeader)
}

func (scbp *sovereignChainBlockProcessor) checkAndSetMissingExtendedHeaders(
	extendedShardHeader *block.ShardHeaderExtended,
	extendedShardHeaderHash []byte,
) {
	allMissingExtendedShardHeadersReceived := false
	scbp.hdrsForCurrBlock.mutHdrsForBlock.Lock()

	haveMissingExtendedShardHeaders := scbp.hdrsForCurrBlock.missingHdrs > 0
	if haveMissingExtendedShardHeaders {
		hdrInfoForHash := scbp.hdrsForCurrBlock.hdrHashAndInfo[string(extendedShardHeaderHash)]
		headerInfoIsNotNil := hdrInfoForHash != nil
		headerIsMissing := headerInfoIsNotNil && check.IfNil(hdrInfoForHash.hdr)
		if headerIsMissing {
			hdrInfoForHash.hdr = extendedShardHeader
			scbp.hdrsForCurrBlock.missingHdrs--
		}

		missingExtendedShardHdrs := scbp.hdrsForCurrBlock.missingHdrs
		allMissingExtendedShardHeadersReceived = missingExtendedShardHdrs == 0
	}

	scbp.hdrsForCurrBlock.mutHdrsForBlock.Unlock()

	if allMissingExtendedShardHeadersReceived {
		scbp.chRcvAllExtendedShardHdrs <- true
	}
}

func (scbp *sovereignChainBlockProcessor) requestIncomingTxsIfNeeded(extendedShardHeader *block.ShardHeaderExtended) {
	mbhs := extendedShardHeader.GetIncomingMiniBlockHandlers()
	if len(mbhs) == 0 {
		return
	}

	body := &block.Body{
		MiniBlocks: make([]*block.MiniBlock, 0),
	}

	for _, mbh := range mbhs {
		mb, isMiniBlock := mbh.(*block.MiniBlock)
		if !isMiniBlock {
			continue
		}
		body.MiniBlocks = append(body.MiniBlocks, mb)
	}

	scbp.txCoordinator.RequestBlockTransactions(body)
}

func (scbp *sovereignChainBlockProcessor) requestExtendedShardHeaders(sovereignChainHeader data.SovereignChainHeaderHandler) uint32 {
	_ = core.EmptyChannel(scbp.chRcvAllExtendedShardHdrs)

	if len(sovereignChainHeader.GetExtendedShardHeaderHashes()) == 0 {
		return scbp.computeAndRequestEpochStartExtendedHeaderIfMissing(sovereignChainHeader)
	}

	return scbp.computeExistingAndRequestMissingExtendedShardHeaders(sovereignChainHeader)
}

func (scbp *sovereignChainBlockProcessor) computeAndRequestEpochStartExtendedHeaderIfMissing(sovereignChainHeader data.SovereignChainHeaderHandler) uint32 {
	if !sovereignChainHeader.IsStartOfEpochBlock() {
		return 0
	}

	lastCrossChainData := sovereignChainHeader.GetLastFinalizedCrossChainHeaderHandler()
	shouldCheckEpochStartCrossChainHash := lastCrossChainData != nil && len(lastCrossChainData.GetHeaderHash()) != 0
	if !shouldCheckEpochStartCrossChainHash {
		return 0
	}

	lastCrossChainHash := lastCrossChainData.GetHeaderHash()
	if !scbp.shouldRequestEpochStartCrossChainHash(lastCrossChainHash) {
		return 0
	}

	scbp.hdrsForCurrBlock.mutHdrsForBlock.Lock()
	scbp.hdrsForCurrBlock.missingHdrs++
	scbp.hdrsForCurrBlock.hdrHashAndInfo[string(lastCrossChainHash)] = &hdrInfo{
		hdr:         nil,
		usedInBlock: false,
	}
	scbp.hdrsForCurrBlock.mutHdrsForBlock.Unlock()

	go scbp.extendedShardHeaderRequester.RequestExtendedShardHeader(lastCrossChainHash)

	return 1
}

func (scbp *sovereignChainBlockProcessor) shouldRequestEpochStartCrossChainHash(lastCrossChainHash []byte) bool {
	_, errMissingHdrPool := process.GetExtendedShardHeaderFromPool(
		lastCrossChainHash,
		scbp.dataPool.Headers())
	_, lastNotarizedHdrHash, _ := scbp.blockTracker.GetLastCrossNotarizedHeader(core.MainChainShardId)

	missingHeaderInTracker := !bytes.Equal(lastNotarizedHdrHash, lastCrossChainHash)
	missingHeaderInPool := errMissingHdrPool != nil
	shouldRequestLastCrossChainHeader := missingHeaderInTracker || missingHeaderInPool

	log.Debug("sovereignChainBlockProcessor.checkAndRequestIfMissingEpochStartExtendedHeader",
		"missingHeaderInTracker", missingHeaderInTracker,
		"missingHeaderInPool", missingHeaderInPool,
		"shouldRequestLastCrossChainHeader", shouldRequestLastCrossChainHeader,
	)

	return shouldRequestLastCrossChainHeader
}

func (scbp *sovereignChainBlockProcessor) computeExistingAndRequestMissingExtendedShardHeaders(sovereignChainHeader data.SovereignChainHeaderHandler) uint32 {
	scbp.hdrsForCurrBlock.mutHdrsForBlock.Lock()
	defer scbp.hdrsForCurrBlock.mutHdrsForBlock.Unlock()

	extendedShardHeaderHashes := sovereignChainHeader.GetExtendedShardHeaderHashes()
	for i := 0; i < len(extendedShardHeaderHashes); i++ {
		hdr, err := process.GetExtendedShardHeaderFromPool(
			extendedShardHeaderHashes[i],
			scbp.dataPool.Headers())

		if err != nil {
			scbp.hdrsForCurrBlock.missingHdrs++
			scbp.hdrsForCurrBlock.hdrHashAndInfo[string(extendedShardHeaderHashes[i])] = &hdrInfo{
				hdr:         nil,
				usedInBlock: true,
			}
			go scbp.extendedShardHeaderRequester.RequestExtendedShardHeader(extendedShardHeaderHashes[i])
			continue
		}

		scbp.hdrsForCurrBlock.hdrHashAndInfo[string(extendedShardHeaderHashes[i])] = &hdrInfo{
			hdr:         hdr,
			usedInBlock: true,
		}
	}

	return scbp.hdrsForCurrBlock.missingHdrs
}

// ProcessBlock actually processes the selected transaction and will create the final block body
func (scbp *sovereignChainBlockProcessor) ProcessBlock(headerHandler data.HeaderHandler, bodyHandler data.BodyHandler, haveTime func() time.Duration) (data.HeaderHandler, data.BodyHandler, error) {
	if haveTime == nil {
		return nil, nil, process.ErrNilHaveTimeHandler
	}

	scbp.processStatusHandler.SetBusy("sovereignChainBlockProcessor.ProcessBlock")
	defer scbp.processStatusHandler.SetIdle()

	err := scbp.checkBlockValidity(headerHandler, bodyHandler)
	if err != nil {
		if err == process.ErrBlockHashDoesNotMatch {
			log.Debug("requested missing sovereign header",
				"hash", headerHandler.GetPrevHash(),
				"for shard", headerHandler.GetShardID(),
			)

			go scbp.requestHandler.RequestShardHeader(headerHandler.GetShardID(), headerHandler.GetPrevHash())
		}

		return nil, nil, err
	}

	scbp.roundNotifier.CheckRound(headerHandler)
	scbp.epochNotifier.CheckEpoch(headerHandler)
	scbp.requestHandler.SetEpoch(headerHandler.GetEpoch())

	log.Debug("started processing block",
		"epoch", headerHandler.GetEpoch(),
		"shard", headerHandler.GetShardID(),
		"round", headerHandler.GetRound(),
		"nonce", headerHandler.GetNonce(),
	)

	sovChainHeader, ok := headerHandler.(data.SovereignChainHeaderHandler)
	if !ok {
		return nil, nil, process.ErrWrongTypeAssertion
	}

	body, ok := bodyHandler.(*block.Body)
	if !ok {
		return nil, nil, process.ErrWrongTypeAssertion
	}

	getMetricsFromBlockBody(body, scbp.marshalizer, scbp.appStatusHandler)

	txCounts, rewardCounts, unsignedCounts := scbp.txCounter.getPoolCounts(scbp.dataPool)
	log.Debug("total txs in pool", "counts", txCounts.String())
	log.Debug("total txs in rewards pool", "counts", rewardCounts.String())
	log.Debug("total txs in unsigned pool", "counts", unsignedCounts.String())

	getMetricsFromHeader(headerHandler.ShallowClone(), uint64(txCounts.GetTotal()), scbp.marshalizer, scbp.appStatusHandler)

	err = scbp.createBlockStarted()
	if err != nil {
		return nil, nil, err
	}

	scbp.blockChainHook.SetCurrentHeader(headerHandler)

	scbp.txCoordinator.RequestBlockTransactions(body)
	requestedExtendedShardHdrs := scbp.requestExtendedShardHeaders(sovChainHeader)

	if haveTime() < 0 {
		return nil, nil, process.ErrTimeIsOut
	}

	err = scbp.txCoordinator.IsDataPreparedForProcessing(haveTime)
	if err != nil {
		return nil, nil, err
	}

	err = scbp.waitForExtendedHeadersIfMissing(requestedExtendedShardHdrs, haveTime)
	if err != nil {
		return nil, nil, err
	}

	for _, accounts := range scbp.accountsDB {
		if accounts.JournalLen() != 0 {
			log.Error("sovereignChainBlockProcessor.ProcessBlock first entry", "stack", string(accounts.GetStackDebugFirstEntry()))
			return nil, nil, process.ErrAccountStateDirty
		}
	}

	defer func() {
		go scbp.checkAndRequestIfExtendedShardHeadersMissing()
	}()

	err = scbp.checkExtendedShardHeadersValidity(sovChainHeader)
	if err != nil {
		return nil, nil, err
	}

	shardHeader, castOK := headerHandler.(data.ShardHeaderHandler)
	if !castOK {
		return nil, nil, errors.ErrWrongTypeAssertion
	}

	scbp.epochStartTrigger.Update(shardHeader.GetRound(), shardHeader.GetNonce())
	err = scbp.baseProcessor.checkEpochCorrectness(shardHeader)
	if err != nil {
		return nil, nil, err
	}

	err = scbp.processIfFirstBlockAfterEpochStart()
	if err != nil {
		return nil, nil, err
	}

	if shardHeader.IsStartOfEpochBlock() {
		err = scbp.processEpochStartMetaBlock(shardHeader, body)
		if err != nil {
			return nil, nil, err
		}

		return shardHeader, body, nil
	}

	defer func() {
		if err != nil {
			scbp.RevertCurrentBlock()
		}
	}()

	newBody, err := scbp.processSovereignBlockTransactions(headerHandler, body, haveTime)
	if err != nil {
		return nil, nil, err
	}

	err = scbp.verifySovereignPostProcessBlock(headerHandler, newBody, sovChainHeader)
	if err != nil {
		return nil, nil, err
	}

	return headerHandler, newBody, nil
}

// checkExtendedShardHeadersValidity checks if used extended shard headers are valid as construction
func (scbp *sovereignChainBlockProcessor) checkExtendedShardHeadersValidity(
	sovChainHeader data.SovereignChainHeaderHandler,
) error {
	lastCrossNotarizedHeader, _, err := scbp.blockTracker.GetLastCrossNotarizedHeader(core.MainChainShardId)
	if err != nil {
		return err
	}

	log.Trace("checkExtendedShardHeadersValidity",
		"lastCrossNotarizedHeader nonce", lastCrossNotarizedHeader.GetNonce(),
		"lastCrossNotarizedHeader round", lastCrossNotarizedHeader.GetRound(),
	)

	extendedShardHdrs, err := scbp.sortExtendedShardHeadersForCurrentBlockByNonce(sovChainHeader)
	if err != nil {
		return err
	}

	if len(extendedShardHdrs) == 0 {
		return nil
	}

	// we should not have an epoch start block with main chain headers to be processed
	if sovChainHeader.IsStartOfEpochBlock() {
		return errors.ErrReceivedSovereignEpochStartBlockWithExtendedHeaders
	}

	if scbp.isGenesisHeaderWithNoPreviousTracking(extendedShardHdrs[0]) {
		// we are missing pre-genesis header, so we can't link it to previous header
		if len(extendedShardHdrs) == 1 {
			return nil
		}

		lastCrossNotarizedHeader = extendedShardHdrs[0]
		extendedShardHdrs = extendedShardHdrs[1:]

		log.Debug("checkExtendedShardHeadersValidity missing pre genesis, updating lastCrossNotarizedHeader",
			"lastCrossNotarizedHeader nonce", lastCrossNotarizedHeader.GetNonce(),
			"lastCrossNotarizedHeader round", lastCrossNotarizedHeader.GetRound(),
		)
	}

	for _, extendedShardHdr := range extendedShardHdrs {
		log.Trace("checkExtendedShardHeadersValidity",
			"extendedShardHeader nonce", extendedShardHdr.GetNonce(),
			"extendedShardHeader round", extendedShardHdr.GetRound(),
		)

		err = scbp.headerValidator.IsHeaderConstructionValid(extendedShardHdr, lastCrossNotarizedHeader)
		if err != nil {
			log.Error("sovereignChainBlockProcessor.checkExtendedShardHeadersValidity", "error", err,
				"extendedShardHdr.Nonce", extendedShardHdr.GetNonce(),
				"lastCrossNotarizedHeader.Nonce", lastCrossNotarizedHeader.GetNonce(),
				"extendedShardHdr.Round", extendedShardHdr.GetRound(),
				"lastCrossNotarizedHeader.Round", lastCrossNotarizedHeader.GetRound(),
			)
			return fmt.Errorf("%w : checkExtendedShardHeadersValidity -> isHdrConstructionValid", err)
		}

		lastCrossNotarizedHeader = extendedShardHdr
	}

	return nil
}

// this will return true if we receive the genesis main chain header in following cases:
//   - no notifier is attached => we did not track main chain and don't have pre-genesis header
//   - node is in re-sync/start in the exact epoch when we start to notarize main chain => no previous
//     main chain tracking(notifier is also disabled)
func (scbp *sovereignChainBlockProcessor) isGenesisHeaderWithNoPreviousTracking(incomingHeader data.HeaderHandler) bool {
	return scbp.extendedShardHeaderTracker.IsGenesisLastCrossNotarizedHeader() && incomingHeader.GetRound() == scbp.mainChainNotarizationStartRound
}

func (scbp *sovereignChainBlockProcessor) processEpochStartMetaBlock(
	header data.HeaderHandler,
	body *block.Body,
) error {

	currentRootHash, err := scbp.validatorStatisticsProcessor.RootHash()
	if err != nil {
		return err
	}

	allValidatorsInfo, err := scbp.validatorStatisticsProcessor.GetValidatorInfoForRootHash(currentRootHash)
	if err != nil {
		return err
	}

	err = scbp.validatorStatisticsProcessor.ProcessRatingsEndOfEpoch(allValidatorsInfo, header.GetEpoch())
	if err != nil {
		return err
	}

	sovHdr, castOk := header.(*block.SovereignChainHeader)
	if !castOk {
		return fmt.Errorf("%w in sovereignChainBlockProcessor.processEpochStartMetaBlock", process.ErrWrongTypeAssertion)
	}

	computedEconomics, err := scbp.updateEpochStartHeader(sovHdr)
	if err != nil {
		return err
	}

	err = scbp.createEpochStartDataCrossChain(sovHdr)
	if err != nil {
		return err
	}

	err = scbp.epochSystemSCProcessor.ProcessSystemSmartContract(allValidatorsInfo, header)
	if err != nil {
		return err
	}

	rewardMiniBlocks, err := scbp.epochRewardsCreator.CreateRewardsMiniBlocks(sovHdr, allValidatorsInfo, &sovHdr.EpochStart.Economics)
	if err != nil {
		return err
	}

	sovHdr.EpochStart.Economics.RewardsForProtocolSustainability.Set(scbp.epochRewardsCreator.GetProtocolSustainabilityRewards())

	err = scbp.epochSystemSCProcessor.ProcessDelegationRewards(rewardMiniBlocks, scbp.epochRewardsCreator.GetLocalTxCache())
	if err != nil {
		return err
	}

	err = scbp.epochEconomics.VerifyRewardsPerBlock(sovHdr, scbp.epochRewardsCreator.GetProtocolSustainabilityRewards(), computedEconomics)
	if err != nil {
		return err
	}

	err = scbp.verifyFees(header)
	if err != nil {
		return err
	}

	saveEpochStartEconomicsMetrics(scbp.appStatusHandler, sovHdr)

	validatorMiniBlocks, err := scbp.validatorInfoCreator.CreateValidatorInfoMiniBlocks(allValidatorsInfo)
	if err != nil {
		return err
	}

	scbp.prepareBlockHeaderInternalMapForValidatorProcessor()
	_, err = scbp.validatorStatisticsProcessor.UpdatePeerState(header, makeCommonHeaderHandlerHashMap(scbp.hdrsForCurrBlock.getHdrHashMap()))
	if err != nil {
		return err
	}

	err = scbp.validatorStatisticsProcessor.ResetValidatorStatisticsAtNewEpoch(allValidatorsInfo)
	if err != nil {
		return err
	}

	finalMiniBlocks := make([]*block.MiniBlock, 0)
	finalMiniBlocks = append(finalMiniBlocks, rewardMiniBlocks...)
	finalMiniBlocks = append(finalMiniBlocks, validatorMiniBlocks...)
	body.MiniBlocks = finalMiniBlocks

	return scbp.applyBodyToHeaderForEpochChange(header, body)
}

func (scbp *sovereignChainBlockProcessor) createEpochStartDataCrossChain(sovHdr data.SovereignChainHeaderHandler) error {
	lastCrossNotarizedHeader, lastCrossNotarizedHeaderHash, err := scbp.blockTracker.GetLastCrossNotarizedHeader(core.MainChainShardId)
	if err != nil {
		return err
	}

	if lastCrossNotarizedHeader.GetNonce() == 0 {
		log.Debug("sovereignChainBlockProcessor.createEpochStartDataCrossChain: no cross chain header notarized yet")
		return nil
	}

	log.Debug("sovereignChainBlockProcessor.createEpochStartDataCrossChain",
		"lastCrossNotarizedHeaderHash", lastCrossNotarizedHeaderHash,
		"lastCrossNotarizedHeaderRound", lastCrossNotarizedHeader.GetRound(),
		"lastCrossNotarizedHeaderNonce", lastCrossNotarizedHeader.GetNonce(),
	)

	return sovHdr.SetLastFinalizedCrossChainHeaderHandler(&block.EpochStartCrossChainData{
		ShardID:    core.MainChainShardId,
		Epoch:      lastCrossNotarizedHeader.GetEpoch(),
		Round:      lastCrossNotarizedHeader.GetRound(),
		Nonce:      lastCrossNotarizedHeader.GetNonce(),
		HeaderHash: lastCrossNotarizedHeaderHash,
	})
}

func (scbp *sovereignChainBlockProcessor) applyBodyToHeaderForEpochChange(header data.HeaderHandler, body *block.Body) error {
	err := header.SetMiniBlockHeaderHandlers(nil)
	if err != nil {
		return err
	}

	totalTxCount, miniBlockHeaderHandlers, err := scbp.createMiniBlockHeaderHandlers(body.MiniBlocks)
	if err != nil {
		return err
	}

	scbp.saveMiniBlocksToPool(miniBlockHeaderHandlers, body.MiniBlocks)

	err = header.SetMiniBlockHeaderHandlers(miniBlockHeaderHandlers)
	if err != nil {
		return err
	}

	err = header.SetTxCount(uint32(totalTxCount))
	if err != nil {
		return err
	}

	userAccountsRootHash, err := scbp.accountsDB[state.UserAccountsState].RootHash()
	if err != nil {
		return err
	}
	err = header.SetRootHash(userAccountsRootHash)
	if err != nil {
		return err
	}

	validatorStatsRootHash, err := scbp.accountsDB[state.PeerAccountsState].RootHash()
	if err != nil {
		return err
	}
	err = header.SetValidatorStatsRootHash(validatorStatsRootHash)
	if err != nil {
		return err
	}

	receiptsHash, err := scbp.txCoordinator.CreateReceiptsHash()
	if err != nil {
		return err
	}

	err = header.SetReceiptsHash(receiptsHash)
	if err != nil {
		return err
	}

	scbp.appStatusHandler.SetUInt64Value(common.MetricNumTxInBlock, uint64(totalTxCount))
	scbp.appStatusHandler.SetUInt64Value(common.MetricNumMiniBlocks, uint64(len(body.MiniBlocks)))

	marshaledBody, err := scbp.marshalizer.Marshal(body)
	if err != nil {
		return err
	}

	scbp.blockSizeThrottler.Add(header.GetRound(), uint32(len(marshaledBody)))

	scbp.createEpochStartData(body)

	return nil
}

func (scbp *sovereignChainBlockProcessor) saveMiniBlocksToPool(
	miniBlockHeaderHandlers []data.MiniBlockHeaderHandler,
	miniBlocks []*block.MiniBlock,
) {
	for idx, mbHeaderHandler := range miniBlockHeaderHandlers {
		mb := miniBlocks[idx]
		hash := mbHeaderHandler.GetHash()
		scbp.dataPool.MiniBlocks().Put(hash, mb, mb.Size())
	}
}

func (scbp *sovereignChainBlockProcessor) checkAndRequestIfExtendedShardHeadersMissing() {
	orderedExtendedShardHeaders, _ := scbp.blockTracker.GetTrackedHeaders(core.MainChainShardId)

	err := scbp.requestHeadersIfMissing(orderedExtendedShardHeaders, core.MainChainShardId)
	if err != nil {
		log.Debug("checkAndRequestIfExtendedShardHeadersMissing", "error", err.Error())
	}
}

func (scbp *sovereignChainBlockProcessor) requestMissingHeaders(missingNonces []uint64, shardId uint32) {
	for _, nonce := range missingNonces {
		scbp.addHeaderIntoTrackerPool(nonce, shardId)
		go scbp.extendedShardHeaderRequester.RequestExtendedShardHeaderByNonce(nonce)
	}
}

func (scbp *sovereignChainBlockProcessor) getExtraMissingNoncesToRequest(_ data.HeaderHandler, _ uint64) []uint64 {
	return make([]uint64, 0)
}

func (scbp *sovereignChainBlockProcessor) sortExtendedShardHeadersForCurrentBlockByNonce(
	sovChainHeader data.SovereignChainHeaderHandler,
) ([]data.HeaderHandler, error) {
	hdrsForCurrentBlock := make([]data.HeaderHandler, 0)

	scbp.hdrsForCurrBlock.mutHdrsForBlock.RLock()
	for _, extendedShardHeaderHash := range sovChainHeader.GetExtendedShardHeaderHashes() {
		headerInfo, found := scbp.hdrsForCurrBlock.hdrHashAndInfo[string(extendedShardHeaderHash)]
		if !found {
			scbp.hdrsForCurrBlock.mutHdrsForBlock.RUnlock()
			return nil, fmt.Errorf("%w : in sovereignChainBlockProcessor.sortExtendedShardHeadersForCurrentBlockByNonce = %s",
				process.ErrMissingHeader, extendedShardHeaderHash)
		}

		log.Trace("sovereignChainBlockProcessor.sortExtendedShardHeadersForCurrentBlockByNonce",
			"headerInfo.hdr.GetNonce()", headerInfo.hdr.GetNonce(),
			"headerInfo.usedInBlock", headerInfo.usedInBlock,
		)

		if headerInfo.usedInBlock {
			hdrsForCurrentBlock = append(hdrsForCurrentBlock, headerInfo.hdr)

		}
	}
	scbp.hdrsForCurrBlock.mutHdrsForBlock.RUnlock()

	process.SortHeadersByNonce(hdrsForCurrentBlock)

	return hdrsForCurrentBlock, nil
}

func (scbp *sovereignChainBlockProcessor) verifyCrossShardMiniBlockDstMe(sovereignChainHeader data.SovereignChainHeaderHandler) error {
	miniBlockExtendedShardHeaderHashes, err := scbp.getAllMiniBlockDstMeFromExtendedShardHeaders(sovereignChainHeader)
	if err != nil {
		return err
	}

	crossMiniBlockHashes := sovereignChainHeader.GetMiniBlockHeadersWithDst(core.SovereignChainShardId)
	for hash := range crossMiniBlockHashes {
		_, found := miniBlockExtendedShardHeaderHashes[hash]
		if !found {
			return process.ErrCrossShardMBWithoutConfirmationFromNotifier
		}
	}

	return nil
}

func (scbp *sovereignChainBlockProcessor) getAllMiniBlockDstMeFromExtendedShardHeaders(sovereignChainHeader data.SovereignChainHeaderHandler) (map[string][]byte, error) {
	lastCrossNotarizedHeader, _, err := scbp.blockTracker.GetLastCrossNotarizedHeader(core.MainChainShardId)
	if err != nil {
		return nil, err
	}

	miniBlockExtendedShardHeaderHashes := make(map[string][]byte)

	scbp.hdrsForCurrBlock.mutHdrsForBlock.RLock()
	for _, extendedShardHeaderHash := range sovereignChainHeader.GetExtendedShardHeaderHashes() {
		headerInfo, ok := scbp.hdrsForCurrBlock.hdrHashAndInfo[string(extendedShardHeaderHash)]
		if !ok {
			continue
		}
		shardHeaderExtended, ok := headerInfo.hdr.(*block.ShardHeaderExtended)
		if !ok {
			continue
		}
		if shardHeaderExtended.GetRound() <= lastCrossNotarizedHeader.GetRound() {
			continue
		}
		if shardHeaderExtended.GetNonce() <= lastCrossNotarizedHeader.GetNonce() {
			continue
		}

		incomingMiniBlocks := shardHeaderExtended.GetIncomingMiniBlocks()
		for _, mb := range incomingMiniBlocks {
			mbHash, errCalculateHash := core.CalculateHash(scbp.marshalizer, scbp.hasher, mb)
			if errCalculateHash != nil {
				return nil, errCalculateHash
			}

			miniBlockExtendedShardHeaderHashes[string(mbHash)] = extendedShardHeaderHash
		}
	}
	scbp.hdrsForCurrBlock.mutHdrsForBlock.RUnlock()

	return miniBlockExtendedShardHeaderHashes, nil
}

// applyBodyToHeader creates a miniblock header list given a block body
func (scbp *sovereignChainBlockProcessor) applyBodyToHeader(
	headerHandler data.HeaderHandler,
	body *block.Body,
) (*block.Body, error) {
	sw := core.NewStopWatch()
	sw.Start("applyBodyToHeader")
	defer func() {
		sw.Stop("applyBodyToHeader")
		log.Debug("measurements", sw.GetMeasurements()...)
	}()

	var err error
	err = headerHandler.SetMiniBlockHeaderHandlers(nil)
	if err != nil {
		return nil, err
	}

	userAccountsRootHash, err := scbp.accountsDB[state.UserAccountsState].RootHash()
	if err != nil {
		return nil, err
	}
	err = headerHandler.SetRootHash(userAccountsRootHash)
	if err != nil {
		return nil, err
	}

	validatorStatsRootHash, err := scbp.accountsDB[state.PeerAccountsState].RootHash()
	if err != nil {
		return nil, err
	}
	err = headerHandler.SetValidatorStatsRootHash(validatorStatsRootHash)
	if err != nil {
		return nil, err
	}

	newBody := deleteSelfReceiptsMiniBlocks(body)
	//TODO: This map should be passed from the caller side
	processedMiniBlocksDestMeInfo := make(map[string]*processedMb.ProcessedMiniBlockInfo)
	err = scbp.applyBodyInfoOnCommonHeader(headerHandler, newBody, processedMiniBlocksDestMeInfo)
	if err != nil {
		return nil, err
	}

	sovHeader := headerHandler.(data.SovereignChainHeaderHandler)
	accumulatedFees, devFees, err := scbp.computeAccumulatedFeesInEpoch(sovHeader)
	if err != nil {
		return nil, err
	}

	err = sovHeader.SetAccumulatedFeesInEpoch(accumulatedFees)
	if err != nil {
		return nil, err
	}

	err = sovHeader.SetDevFeesInEpoch(devFees)
	if err != nil {
		return nil, err
	}

	return newBody, nil
}

func (scbp *sovereignChainBlockProcessor) computeAccumulatedFeesInEpoch(metaHdr data.SovereignChainHeaderHandler) (*big.Int, *big.Int, error) {
	currentlyAccumulatedFeesInEpoch := big.NewInt(0)
	currentDevFeesInEpoch := big.NewInt(0)

	lastHdr := scbp.blockChain.GetCurrentBlockHeader()
	if !check.IfNil(lastHdr) {
		lastSovHeader, ok := lastHdr.(data.SovereignChainHeaderHandler)
		if !ok {
			return nil, nil, process.ErrWrongTypeAssertion
		}

		if !lastHdr.IsStartOfEpochBlock() {
			currentlyAccumulatedFeesInEpoch = big.NewInt(0).Set(lastSovHeader.GetAccumulatedFeesInEpoch())
			currentDevFeesInEpoch = big.NewInt(0).Set(lastSovHeader.GetDevFeesInEpoch())
		}
	}

	currentlyAccumulatedFeesInEpoch.Add(currentlyAccumulatedFeesInEpoch, metaHdr.GetAccumulatedFees())
	currentDevFeesInEpoch.Add(currentDevFeesInEpoch, metaHdr.GetDeveloperFees())
	log.Debug("computeAccumulatedFeesInEpoch - sovereign block fees",
		"nonce", metaHdr.GetNonce(),
		"accumulatedFees", metaHdr.GetAccumulatedFees().String(),
		"devFees", metaHdr.GetDeveloperFees().String(),
		"leader fees", core.GetIntTrimmedPercentageOfValue(big.NewInt(0).Sub(metaHdr.GetAccumulatedFees(), metaHdr.GetDeveloperFees()), scbp.economicsData.LeaderPercentage()).String())

	log.Debug("computeAccumulatedFeesInEpoch - fees in epoch",
		"accumulatedFeesInEpoch", currentlyAccumulatedFeesInEpoch.String(),
		"devFeesInEpoch", currentDevFeesInEpoch.String())

	return currentlyAccumulatedFeesInEpoch, currentDevFeesInEpoch, nil
}

func (scbp *sovereignChainBlockProcessor) processSovereignBlockTransactions(
	headerHandler data.HeaderHandler,
	body *block.Body,
	haveTime func() time.Duration,
) (*block.Body, error) {
	startTime := time.Now()
	miniblocks, err := scbp.txCoordinator.ProcessBlockTransaction(headerHandler, body, haveTime)
	elapsedTime := time.Since(startTime)
	log.Debug("elapsed time to process block transaction",
		"time [s]", elapsedTime,
	)
	if err != nil {
		return nil, err
	}

	postProcessMBs := scbp.txCoordinator.CreatePostProcessMiniBlocks()

	receiptsHash, err := scbp.txCoordinator.CreateReceiptsHash()
	if err != nil {
		return nil, err
	}

	err = headerHandler.SetReceiptsHash(receiptsHash)
	if err != nil {
		return nil, err
	}

	err = scbp.scToProtocol.UpdateProtocol(&block.Body{MiniBlocks: postProcessMBs}, headerHandler.GetNonce())
	if err != nil {
		return nil, err
	}

	scbp.prepareBlockHeaderInternalMapForValidatorProcessor()
	_, err = scbp.validatorStatisticsProcessor.UpdatePeerState(headerHandler, makeCommonHeaderHandlerHashMap(scbp.hdrsForCurrBlock.getHdrHashMap()))
	if err != nil {
		return nil, err
	}

	createdBlockBody := &block.Body{MiniBlocks: miniblocks}
	createdBlockBody.MiniBlocks = append(createdBlockBody.MiniBlocks, postProcessMBs...)

	err = scbp.createAndSetOutGoingMiniBlock(headerHandler, createdBlockBody)
	if err != nil {
		return nil, err
	}

	return scbp.applyBodyToHeader(headerHandler, createdBlockBody)
}

func (scbp *sovereignChainBlockProcessor) createAndSetOutGoingMiniBlock(headerHandler data.HeaderHandler, createdBlockBody *block.Body) error {
	logs := scbp.txCoordinator.GetAllCurrentLogs()
	outGoingOperations, err := scbp.outgoingOperationsFormatter.CreateOutgoingTxsData(logs)
	if err != nil {
		return err
	}

	if len(outGoingOperations) == 0 {
		return nil
	}

	outGoingMb, outGoingOperationsHash := scbp.createOutGoingMiniBlockData(outGoingOperations)
	return scbp.setOutGoingMiniBlock(headerHandler, createdBlockBody, outGoingMb, outGoingOperationsHash)
}

func (scbp *sovereignChainBlockProcessor) createOutGoingMiniBlockData(outGoingOperations [][]byte) (*block.MiniBlock, []byte) {
	outGoingOpHashes := make([][]byte, len(outGoingOperations))
	aggregatedOutGoingOperations := make([]byte, 0)
	outGoingOperationsData := make([]*sovCore.OutGoingOperation, 0)

	for idx, outGoingOp := range outGoingOperations {
		outGoingOpHash := scbp.operationsHasher.Compute(string(outGoingOp))
		aggregatedOutGoingOperations = append(aggregatedOutGoingOperations, outGoingOpHash...)

		outGoingOpData := &sovCore.OutGoingOperation{
			Hash: outGoingOpHash,
			Data: outGoingOp,
		}
		outGoingOpHashes[idx] = outGoingOpHash
		outGoingOperationsData = append(outGoingOperationsData, outGoingOpData)

		scbp.addOutGoingTxToPool(outGoingOpData)
	}

	outGoingOperationsHash := scbp.operationsHasher.Compute(string(aggregatedOutGoingOperations))
	scbp.outGoingOperationsPool.Add(&sovCore.BridgeOutGoingData{
		Hash:               outGoingOperationsHash,
		OutGoingOperations: outGoingOperationsData,
	})

	return &block.MiniBlock{
		TxHashes:        outGoingOpHashes,
		ReceiverShardID: core.MainChainShardId,
		SenderShardID:   scbp.shardCoordinator.SelfId(),
	}, outGoingOperationsHash
}

func (scbp *sovereignChainBlockProcessor) addOutGoingTxToPool(outGoingOp *sovCore.OutGoingOperation) {
	tx := &transaction.Transaction{
		GasLimit: scbp.economicsData.ComputeGasLimit(
			&transaction.Transaction{
				Data: outGoingOp.Data,
			}),
		GasPrice: scbp.economicsData.MinGasPrice(),
		Data:     outGoingOp.Data,
	}

	cacheID := fmt.Sprintf("%d_%d", core.SovereignChainShardId, core.MainChainShardId)
	scbp.dataPool.Transactions().AddData(
		outGoingOp.Hash,
		tx,
		tx.Size(),
		cacheID,
	)
}

func (scbp *sovereignChainBlockProcessor) setOutGoingMiniBlock(
	headerHandler data.HeaderHandler,
	createdBlockBody *block.Body,
	outGoingMb *block.MiniBlock,
	outGoingOperationsHash []byte,
) error {
	outGoingMbHash, err := core.CalculateHash(scbp.marshalizer, scbp.hasher, outGoingMb)
	if err != nil {
		return err
	}

	sovereignChainHdr, ok := headerHandler.(data.SovereignChainHeaderHandler)
	if !ok {
		return fmt.Errorf("%w in sovereignChainBlockProcessor.setOutGoingOperation", process.ErrWrongTypeAssertion)
	}

	outGoingMbHeader := &block.OutGoingMiniBlockHeader{
		Hash:                   outGoingMbHash,
		OutGoingOperationsHash: outGoingOperationsHash,
	}

	err = sovereignChainHdr.SetOutGoingMiniBlockHeaderHandler(outGoingMbHeader)
	if err != nil {
		return err
	}

	createdBlockBody.MiniBlocks = append(createdBlockBody.MiniBlocks, outGoingMb)
	scbp.txCoordinator.AddTxsFromMiniBlocks([]*block.MiniBlock{outGoingMb})
	return nil
}

func (scbp *sovereignChainBlockProcessor) waitForExtendedHeadersIfMissing(requestedExtendedShardHdrs uint32, haveTime func() time.Duration) error {
	haveMissingExtendedShardHeaders := requestedExtendedShardHdrs > 0
	if haveMissingExtendedShardHeaders {
		log.Debug("requested missing extended shard headers",
			"num headers", requestedExtendedShardHdrs,
		)

		err := waitForHeaderHashes(haveTime(), scbp.chRcvAllExtendedShardHdrs)

		scbp.hdrsForCurrBlock.mutHdrsForBlock.RLock()
		missingExtendedShardHdrs := scbp.hdrsForCurrBlock.missingHdrs
		scbp.hdrsForCurrBlock.mutHdrsForBlock.RUnlock()

		scbp.hdrsForCurrBlock.resetMissingHdrs()

		log.Debug("received missing extended shard headers",
			"num headers", requestedExtendedShardHdrs-missingExtendedShardHdrs,
		)

		if err != nil {
			return err
		}
	}

	return nil
}

func (scbp *sovereignChainBlockProcessor) verifySovereignPostProcessBlock(
	headerHandler data.HeaderHandler,
	newBody *block.Body,
	sovereignChainHeader data.SovereignChainHeaderHandler,
) error {
	//TODO: This check could be removed in sovereign implementation
	err := scbp.txCoordinator.VerifyCreatedBlockTransactions(headerHandler, newBody)
	if err != nil {
		return err
	}

	//TODO: This check could be removed in sovereign implementation
	err = scbp.checkHeaderBodyCorrelation(headerHandler.GetMiniBlockHeaderHandlers(), newBody)
	if err != nil {
		return err
	}

	err = scbp.verifyCrossShardMiniBlockDstMe(sovereignChainHeader)
	if err != nil {
		return err
	}

	err = scbp.verifyFees(headerHandler)
	if err != nil {
		return err
	}

	if !scbp.verifyStateRoot(headerHandler.GetRootHash()) {
		err = process.ErrRootStateDoesNotMatch
		return err
	}

	err = scbp.verifyValidatorStatisticsRootHash(headerHandler)
	if err != nil {
		return err
	}

	return nil
}

// TODO: verify if block created from processblock is the same one as received from leader - without signature - no need for another set of checks
// actually sign check should resolve this - as you signed something generated by you

// CommitBlock - will do a lot of verification
func (scbp *sovereignChainBlockProcessor) CommitBlock(headerHandler data.HeaderHandler, bodyHandler data.BodyHandler) error {
	scbp.processStatusHandler.SetBusy("sovereignChainBlockProcessor.CommitBlock")
	var err error
	defer func() {
		if err != nil {
			scbp.RevertCurrentBlock()
		}
		scbp.processStatusHandler.SetIdle()
	}()

	err = checkForNils(headerHandler, bodyHandler)
	if err != nil {
		return err
	}

	log.Debug("started committing block",
		"epoch", headerHandler.GetEpoch(),
		"shard", headerHandler.GetShardID(),
		"round", headerHandler.GetRound(),
		"nonce", headerHandler.GetNonce(),
	)

	err = scbp.checkBlockValidity(headerHandler, bodyHandler)
	if err != nil {
		return err
	}

	scbp.store.SetEpochForPutOperation(headerHandler.GetEpoch())

	marshalizedHeader, err := scbp.marshalizer.Marshal(headerHandler)
	if err != nil {
		return err
	}

	body, ok := bodyHandler.(*block.Body)
	if !ok {
		err = process.ErrWrongTypeAssertion
		return err
	}

	sovMetaHdr, castOk := headerHandler.(data.MetaHeaderHandler)
	if !castOk {
		log.Error("sovereignChainBlockProcessor.CommitBlock: before commitEpochStart", "error", process.ErrWrongTypeAssertion)
		return process.ErrWrongTypeAssertion
	}

	// must be called before commitEpochStart
	rewardsTxs := scbp.getRewardsTxs(scbp.epochRewardsCreator, sovMetaHdr, body)

	scbp.commitEpochStart(sovMetaHdr, body, scbp.epochRewardsCreator, scbp.validatorInfoCreator)

	headerHash := scbp.hasher.Compute(string(marshalizedHeader))
	scbp.saveShardHeader(headerHandler, headerHash, marshalizedHeader)
	scbp.saveBody(body, headerHandler, headerHash)

	processedExtendedShardHdrs, err := scbp.getOrderedProcessedExtendedShardHeadersFromHeader(headerHandler)
	if err != nil {
		return err
	}

	err = scbp.addProcessedCrossMiniBlocksFromExtendedShardHeader(headerHandler)
	if err != nil {
		return err
	}

	if len(processedExtendedShardHdrs) > 0 {
		err = scbp.saveLastNotarizedHeader(core.MainChainShardId, processedExtendedShardHdrs)
		if err != nil {
			return err
		}
	}

	err = scbp.commitAll(headerHandler)
	if err != nil {
		return err
	}

	log.Info("shard block has been committed successfully",
		"epoch", headerHandler.GetEpoch(),
		"shard", headerHandler.GetShardID(),
		"round", headerHandler.GetRound(),
		"nonce", headerHandler.GetNonce(),
		"hash", headerHash,
	)

	scbp.validatorStatisticsProcessor.DisplayRatings(headerHandler.GetEpoch())

	scbp.setNonceOfFirstCommittedBlock(headerHandler.GetNonce())

	scbp.updateLastCommittedInDebugger(headerHandler.GetRound())

	errNotCritical := scbp.updateCrossShardInfo(processedExtendedShardHdrs)
	if errNotCritical != nil {
		log.Debug("updateCrossShardInfo", "error", errNotCritical.Error())
	}

	err = scbp.forkDetector.AddHeader(headerHandler, headerHash, process.BHProcessed, nil, nil)
	if err != nil {
		log.Debug("forkDetector.AddHeader", "error", err.Error())
		return err
	}

	lastSelfNotarizedHeader, lastSelfNotarizedHeaderHash := getLastSelfNotarizedHeaderByItself(scbp.blockChain)
	scbp.blockTracker.AddSelfNotarizedHeader(scbp.shardCoordinator.SelfId(), lastSelfNotarizedHeader, lastSelfNotarizedHeaderHash)

	if scbp.lastRestartNonce == 0 {
		scbp.lastRestartNonce = headerHandler.GetNonce()
	}

	scbp.updateState(lastSelfNotarizedHeader, lastSelfNotarizedHeaderHash)

	highestFinalBlockNonce := scbp.forkDetector.GetHighestFinalBlockNonce()
	log.Debug("highest final shard block",
		"shard", scbp.shardCoordinator.SelfId(),
		"nonce", highestFinalBlockNonce,
	)

	err = scbp.saveSovereignMetricsForCommittedBlock(
		logger.DisplayByteSlice(headerHash),
		highestFinalBlockNonce,
		headerHandler)
	if err != nil {
		return err
	}

	err = scbp.commonHeaderAndBodyCommit(
		headerHandler,
		body,
		headerHash,
		[]data.HeaderHandler{lastSelfNotarizedHeader},
		[][]byte{lastSelfNotarizedHeaderHash},
		rewardsTxs,
	)
	if err != nil {
		return err
	}

	scbp.indexValidatorsRatingIfNeeded(headerHandler)

	return nil
}

func (scbp *sovereignChainBlockProcessor) indexValidatorsRatingIfNeeded(
	header data.HeaderHandler,
) {
	if !scbp.outportHandler.HasDrivers() {
		return
	}

	indexValidatorsRating(scbp.outportHandler, scbp.validatorStatisticsProcessor, header)
}

func (scbp *sovereignChainBlockProcessor) createEpochStartData(body *block.Body) {
	// this will create validators info data and save it to pool
	_ = scbp.validatorInfoCreator.CreateMarshalledData(body)
	_ = scbp.epochRewardsCreator.CreateMarshalledData(body)
}

// getOrderedProcessedExtendedShardHeadersFromHeader returns all the extended shard headers fully processed
func (scbp *sovereignChainBlockProcessor) getOrderedProcessedExtendedShardHeadersFromHeader(header data.HeaderHandler) ([]data.HeaderHandler, error) {
	if check.IfNil(header) {
		return nil, process.ErrNilBlockHeader
	}

	miniBlockHeaders := header.GetMiniBlockHeaderHandlers()
	miniBlockHashes := make(map[int][]byte, len(miniBlockHeaders))
	for i := 0; i < len(miniBlockHeaders); i++ {
		miniBlockHashes[i] = miniBlockHeaders[i].GetHash()
	}

	log.Trace("cross mini blocks in body",
		"num miniblocks", len(miniBlockHashes),
	)

	processedExtendedShardHeaders, err := scbp.getOrderedProcessedExtendedShardHeadersFromMiniBlockHashes(miniBlockHeaders, miniBlockHashes)
	if err != nil {
		return nil, err
	}

	return processedExtendedShardHeaders, nil
}

func (scbp *sovereignChainBlockProcessor) getOrderedProcessedExtendedShardHeadersFromMiniBlockHashes(
	miniBlockHeaders []data.MiniBlockHeaderHandler,
	miniBlockHashes map[int][]byte,
) ([]data.HeaderHandler, error) {

	processedExtendedShardHeaders := make([]data.HeaderHandler, 0, len(scbp.hdrsForCurrBlock.hdrHashAndInfo))
	processedCrossMiniBlocksHashes := make(map[string]bool, len(scbp.hdrsForCurrBlock.hdrHashAndInfo))

	scbp.hdrsForCurrBlock.mutHdrsForBlock.RLock()
	for extendedShardHeaderHash, headerInfo := range scbp.hdrsForCurrBlock.hdrHashAndInfo {
		if !headerInfo.usedInBlock {
			continue
		}

		extendedShardHeader, ok := headerInfo.hdr.(*block.ShardHeaderExtended)
		if !ok {
			scbp.hdrsForCurrBlock.mutHdrsForBlock.RUnlock()
			return nil, process.ErrWrongTypeAssertion
		}

		log.Trace("extended shard header",
			"nonce", extendedShardHeader.GetNonce(),
		)

		processedAll, err := scbp.processCrossMiniBlockHashes(
			extendedShardHeader,
			extendedShardHeaderHash,
			processedCrossMiniBlocksHashes,
			miniBlockHashes,
			miniBlockHeaders,
		)
		if err != nil {
			scbp.hdrsForCurrBlock.mutHdrsForBlock.RUnlock()
			return nil, err
		}

		if processedAll {
			processedExtendedShardHeaders = append(processedExtendedShardHeaders, extendedShardHeader)
		}
	}
	scbp.hdrsForCurrBlock.mutHdrsForBlock.RUnlock()

	process.SortHeadersByNonce(processedExtendedShardHeaders)

	return processedExtendedShardHeaders, nil
}

func (scbp *sovereignChainBlockProcessor) processCrossMiniBlockHashes(
	extendedShardHeader *block.ShardHeaderExtended,
	extendedShardHeaderHash string,
	processedCrossMiniBlocksHashes map[string]bool,
	miniBlockHashes map[int][]byte,
	miniBlockHeaders []data.MiniBlockHeaderHandler,
) (bool, error) {
	crossMiniBlockHashes := make(map[string]struct{})
	incomingMiniBlocks := extendedShardHeader.GetIncomingMiniBlocks()
	for _, mb := range incomingMiniBlocks {
		mbHash, err := core.CalculateHash(scbp.marshalizer, scbp.hasher, mb)
		if err != nil {
			scbp.hdrsForCurrBlock.mutHdrsForBlock.RUnlock()
			return false, err
		}

		crossMiniBlockHashes[string(mbHash)] = struct{}{}
	}

	for hash := range crossMiniBlockHashes {
		processedCrossMiniBlocksHashes[hash] = scbp.processedMiniBlocksTracker.IsMiniBlockFullyProcessed([]byte(extendedShardHeaderHash), []byte(hash))
	}

	for key, miniBlockHash := range miniBlockHashes {
		_, ok := crossMiniBlockHashes[string(miniBlockHash)]
		if !ok {
			continue
		}

		processedCrossMiniBlocksHashes[string(miniBlockHash)] = miniBlockHeaders[key].IsFinal()

		delete(miniBlockHashes, key)
	}

	log.Trace("cross mini blocks in extended shard header",
		"num miniblocks", len(crossMiniBlockHashes),
	)

	processedAll := true
	for hash := range crossMiniBlockHashes {
		if !processedCrossMiniBlocksHashes[hash] {
			processedAll = false
			break
		}
	}

	return processedAll, nil
}

func (scbp *sovereignChainBlockProcessor) addProcessedCrossMiniBlocksFromExtendedShardHeader(headerHandler data.HeaderHandler) error {
	if check.IfNil(headerHandler) {
		return process.ErrNilBlockHeader
	}

	sovereignChainShardHeader, ok := headerHandler.(data.SovereignChainHeaderHandler)
	if !ok {
		return process.ErrWrongTypeAssertion
	}
	miniBlockHashes := make(map[int][]byte, len(headerHandler.GetMiniBlockHeaderHandlers()))
	for i := 0; i < len(headerHandler.GetMiniBlockHeaderHandlers()); i++ {
		miniBlockHashes[i] = headerHandler.GetMiniBlockHeaderHandlers()[i].GetHash()
	}

	scbp.hdrsForCurrBlock.mutHdrsForBlock.RLock()
	for _, extendedShardHeaderHash := range sovereignChainShardHeader.GetExtendedShardHeaderHashes() {
		headerInfo, found := scbp.hdrsForCurrBlock.hdrHashAndInfo[string(extendedShardHeaderHash)]
		if !found {
			scbp.hdrsForCurrBlock.mutHdrsForBlock.RUnlock()
			return fmt.Errorf("%w : addProcessedCrossMiniBlocksFromExtendedShardHeader extendedShardHeaderHash = %s",
				process.ErrMissingHeader, logger.DisplayByteSlice(extendedShardHeaderHash))
		}

		shardHeaderExtended, isShardHeaderExtended := headerInfo.hdr.(*block.ShardHeaderExtended)
		if !isShardHeaderExtended {
			scbp.hdrsForCurrBlock.mutHdrsForBlock.RUnlock()
			return process.ErrWrongTypeAssertion
		}

		crossMiniBlockHashes, err := scbp.getMiniBlocksHashes(shardHeaderExtended.GetIncomingMiniBlocks())
		if err != nil {
			scbp.hdrsForCurrBlock.mutHdrsForBlock.RUnlock()
			return err
		}

		scbp.setProcessedMiniBlocks(miniBlockHashes, crossMiniBlockHashes, headerHandler, extendedShardHeaderHash)
	}
	scbp.hdrsForCurrBlock.mutHdrsForBlock.RUnlock()

	return nil
}

func (scbp *sovereignChainBlockProcessor) getMiniBlocksHashes(mbs []*block.MiniBlock) (map[string]struct{}, error) {
	mbHashes := make(map[string]struct{})

	for _, mb := range mbs {
		mbHash, err := core.CalculateHash(scbp.marshalizer, scbp.hasher, mb)
		if err != nil {
			return nil, err
		}

		mbHashes[string(mbHash)] = struct{}{}
	}

	return mbHashes, nil
}

func (scbp *sovereignChainBlockProcessor) setProcessedMiniBlocks(
	miniBlockHashes map[int][]byte,
	crossMiniBlockHashes map[string]struct{},
	headerHandler data.HeaderHandler,
	extendedShardHeaderHash []byte,
) {
	for key, miniBlockHash := range miniBlockHashes {
		_, ok := crossMiniBlockHashes[string(miniBlockHash)]
		if !ok {
			continue
		}

		miniBlockHeader := process.GetMiniBlockHeaderWithHash(headerHandler, miniBlockHash)
		if miniBlockHeader == nil {
			log.Warn("sovereignChainBlockProcessor.addProcessedCrossMiniBlocksFromExtendedShardHeader: GetMiniBlockHeaderWithHash", "mb hash", miniBlockHash, "error", process.ErrMissingMiniBlockHeader)
			continue
		}

		scbp.processedMiniBlocksTracker.SetProcessedMiniBlockInfo(extendedShardHeaderHash, miniBlockHash, &processedMb.ProcessedMiniBlockInfo{
			FullyProcessed:         miniBlockHeader.IsFinal(),
			IndexOfLastTxProcessed: miniBlockHeader.GetIndexOfLastTxProcessed(),
		})

		delete(miniBlockHashes, key)
	}
}

func (scbp *sovereignChainBlockProcessor) updateCrossShardInfo(processedExtendedShardHdrs []data.HeaderHandler) error {
	lastCrossNotarizedHeader, _, err := scbp.blockTracker.GetLastCrossNotarizedHeader(core.MainChainShardId)
	if err != nil {
		return err
	}

	// processedExtendedShardHdrs is also sorted
	for i := 0; i < len(processedExtendedShardHdrs); i++ {
		hdr := processedExtendedShardHdrs[i]

		// remove process finished
		if hdr.GetNonce() > lastCrossNotarizedHeader.GetNonce() {
			continue
		}

		// extended shard header was processed and finalized
		marshalledHeader, errMarshal := scbp.marshalizer.Marshal(hdr)
		if errMarshal != nil {
			log.Debug("updateCrossShardInfo.Marshal", "error", errMarshal.Error())
			continue
		}

		headerHash := scbp.hasher.Compute(string(marshalledHeader))

		scbp.saveExtendedShardHeader(hdr, headerHash, marshalledHeader)

		scbp.processedMiniBlocksTracker.RemoveHeaderHash(headerHash)
	}

	return nil
}

func (scbp *sovereignChainBlockProcessor) saveExtendedShardHeader(header data.HeaderHandler, headerHash []byte, marshalizedHeader []byte) {
	startTime := time.Now()

	nonceToByteSlice := scbp.uint64Converter.ToByteSlice(header.GetNonce())

	errNotCritical := scbp.store.Put(dataRetriever.ExtendedShardHeadersNonceHashDataUnit, nonceToByteSlice, headerHash)
	if errNotCritical != nil {
		logging.LogErrAsWarnExceptAsDebugIfClosingError(log, errNotCritical,
			"saveExtendedShardHeader.Put -> ExtendedShardHdrNonceHashDataUnit",
			"err", errNotCritical)
	}

	errNotCritical = scbp.store.Put(dataRetriever.ExtendedShardHeadersUnit, headerHash, marshalizedHeader)
	if errNotCritical != nil {
		logging.LogErrAsWarnExceptAsDebugIfClosingError(log, errNotCritical,
			"saveExtendedShardHeader.Put -> ExtendedShardHeadersUnit",
			"err", errNotCritical)
	}

	elapsedTime := time.Since(startTime)
	if elapsedTime >= common.PutInStorerMaxTime {
		log.Warn("saveExtendedShardHeader", "elapsed time", elapsedTime)
	}
}

func (scbp *sovereignChainBlockProcessor) saveSovereignMetricsForCommittedBlock(
	currentBlockHash string,
	highestFinalBlockNonce uint64,
	shardHeader data.HeaderHandler,
) error {
	baseSaveMetricsForCommittedShardBlock(
		scbp.nodesCoordinator,
		scbp.appStatusHandler,
		currentBlockHash,
		highestFinalBlockNonce,
		shardHeader,
		scbp.managedPeersHolder,
	)
	lastMainChainHdr, _, err := scbp.blockTracker.GetLastCrossNotarizedHeader(core.MainChainShardId)
	if err != nil {
		return err
	}

	scbp.appStatusHandler.SetStringValue(common.MetricCrossCheckBlockHeight, fmt.Sprintf("mainChain %d", lastMainChainHdr.GetNonce()))
	return nil
}

// RestoreBlockIntoPools restores block into pools
func (scbp *sovereignChainBlockProcessor) RestoreBlockIntoPools(header data.HeaderHandler, body data.BodyHandler) error {
	scbp.restoreBlockBody(header, body)

	// TODO: MX-16507: check how/if to restore incoming scrs/cross chain txs once we have a testnet setup for it
	// 1. For incoming scrs We should probably have something similar to (sp *shardProcessor) rollBackProcessedMiniBlocksInfo
	// 2. For outgoing operations, we should analyse how to revert operations or delay them before sending (wait for them
	// to be finalized and impossible to rollback maybe?)

	sovChainHdr, castOk := header.(data.SovereignChainHeaderHandler)
	if !castOk {
		return fmt.Errorf("%w in sovereignChainBlockProcessor.RestoreBlockIntoPools", errors.ErrWrongTypeAssertion)
	}

	err := scbp.restoreExtendedHeaderIntoPool(sovChainHdr.GetExtendedShardHeaderHashes())
	if err != nil {
		return err
	}

	scbp.extendedShardHeaderTracker.RemoveLastSelfNotarizedHeaders()

	numOfNotarizedExtendedHeaders := len(sovChainHdr.GetExtendedShardHeaderHashes())
	log.Debug("sovereignChainBlockProcessor.RestoreBlockIntoPools", "numOfNotarizedExtendedHeaders", numOfNotarizedExtendedHeaders)
	if numOfNotarizedExtendedHeaders == 0 {
		return nil
	}

	scbp.extendedShardHeaderTracker.RemoveLastCrossNotarizedHeaders()
	return nil
}

func (scbp *sovereignChainBlockProcessor) restoreExtendedHeaderIntoPool(extendedShardHeaderHashes [][]byte) error {
	for _, extendedHdrHash := range extendedShardHeaderHashes {
		extendedHdr, errNotCritical := process.GetExtendedShardHeaderFromStorage(extendedHdrHash, scbp.marshalizer, scbp.store)
		if errNotCritical != nil {
			log.Debug("extended header is not fully processed yet and not committed in ExtendedShardHeadersUnit",
				"hash", extendedHdrHash)
			continue
		}

		scbp.dataPool.Headers().AddHeaderInShard(extendedHdrHash, extendedHdr, core.MainChainShardId)

		extendedHdrStorer, err := scbp.store.GetStorer(dataRetriever.ExtendedShardHeadersUnit)
		if err != nil {
			log.Error("unable to get storage unit",
				"unit", dataRetriever.ExtendedShardHeadersUnit.String())
			return err
		}

		err = extendedHdrStorer.Remove(extendedHdrHash)
		if err != nil {
			log.Error("unable to remove hash from ExtendedShardHeadersUnit",
				"hash", extendedHdrHash)
			return err
		}

		nonceToByteSlice := scbp.uint64Converter.ToByteSlice(extendedHdr.GetNonce())
		extendedHdrNonceHashStorer, err := scbp.store.GetStorer(dataRetriever.ExtendedShardHeadersNonceHashDataUnit)
		if err != nil {
			log.Error("unable to get storage unit",
				"unit", dataRetriever.ExtendedShardHeadersNonceHashDataUnit.String())
			return err
		}

		errNotCritical = extendedHdrNonceHashStorer.Remove(nonceToByteSlice)
		if errNotCritical != nil {
			log.Debug("extendedHdrNonceHashStorer.Remove error not critical",
				"error", errNotCritical.Error())
		}

		log.Debug("extended block has been restored successfully",
			"round", extendedHdr.GetRound(),
			"nonce", extendedHdr.GetNonce(),
			"hash", extendedHdrHash)
	}

	return nil
}

// RevertStateToBlock reverts state in tries
func (scbp *sovereignChainBlockProcessor) RevertStateToBlock(header data.HeaderHandler, rootHash []byte) error {
	rootHashHolder := holders.NewDefaultRootHashesHolder(rootHash)
	err := scbp.accountsDB[state.UserAccountsState].RecreateTrie(rootHashHolder)
	if err != nil {
		log.Debug("recreate trie with error for header",
			"nonce", header.GetNonce(),
			"header root hash", header.GetRootHash(),
			"given root hash", rootHash,
			"error", err.Error(),
		)

		return err
	}

	err = scbp.epochStartTrigger.RevertStateToBlock(header)
	if err != nil {
		log.Debug("revert epoch start trigger for header",
			"nonce", header.GetNonce(),
			"error", err,
		)
		return err
	}

	err = scbp.validatorStatisticsProcessor.RevertPeerState(header)
	if err != nil {
		log.Debug("revert peer state with error for header",
			"nonce", header.GetNonce(),
			"validators root hash", header.GetValidatorStatsRootHash(),
			"error", err.Error(),
		)

		return err
	}

	return nil
}

// RevertCurrentBlock reverts the current block for cleanup failed process
func (scbp *sovereignChainBlockProcessor) RevertCurrentBlock() {
	scbp.revertAccountState()
}

// PruneStateOnRollback prunes states of all accounts DBs
func (scbp *sovereignChainBlockProcessor) PruneStateOnRollback(currHeader data.HeaderHandler, _ []byte, prevHeader data.HeaderHandler, _ []byte) {
	for key := range scbp.accountsDB {
		if !scbp.accountsDB[key].IsPruningEnabled() {
			continue
		}

		currentRootHash, prevRootHash := scbp.getRootHashes(currHeader, prevHeader, key)
		if bytes.Equal(currentRootHash, prevRootHash) {
			continue
		}

		scbp.accountsDB[key].CancelPrune(prevRootHash, state.OldRoot)
		scbp.accountsDB[key].PruneTrie(currentRootHash, state.NewRoot, scbp.getPruningHandler(currHeader.GetNonce()))
	}
}

// ProcessScheduledBlock does nothing - as this uses new execution model
func (scbp *sovereignChainBlockProcessor) ProcessScheduledBlock(_ data.HeaderHandler, _ data.BodyHandler, _ func() time.Duration) error {
	return nil
}

// DecodeBlockHeader decodes the current header
func (scbp *sovereignChainBlockProcessor) DecodeBlockHeader(data []byte) data.HeaderHandler {
	if data == nil {
		return nil
	}

	header, err := process.UnmarshalSovereignChainHeader(scbp.marshalizer, data)
	if err != nil {
		log.Debug("DecodeBlockHeader.UnmarshalSovereignChainHeader", "error", err.Error())
		return nil
	}

	return header
}

func (scbp *sovereignChainBlockProcessor) verifyValidatorStatisticsRootHash(headerHandler data.HeaderHandler) error {
	validatorStatsRH, err := scbp.accountsDB[state.PeerAccountsState].RootHash()
	if err != nil {
		return err
	}

	if !bytes.Equal(validatorStatsRH, headerHandler.GetValidatorStatsRootHash()) {
		log.Debug("validator stats root hash mismatch",
			"computed", validatorStatsRH,
			"received", headerHandler.GetValidatorStatsRootHash(),
		)
		return fmt.Errorf("%s, sovereign, computed: %s, received: %s, header nonce: %d",
			process.ErrValidatorStatsRootHashDoesNotMatch,
			logger.DisplayByteSlice(validatorStatsRH),
			logger.DisplayByteSlice(headerHandler.GetValidatorStatsRootHash()),
			headerHandler.GetNonce(),
		)
	}

	return nil
}

func (scbp *sovereignChainBlockProcessor) updateState(header data.HeaderHandler, headerHash []byte) {
	if check.IfNil(header) {
		log.Debug("updateState nil header")
		return
	}

	scbp.validatorStatisticsProcessor.SetLastFinalizedRootHash(header.GetValidatorStatsRootHash())

	prevHeaderHash := header.GetPrevHash()
	prevHeader, errNotCritical := process.GetSovereignChainHeader(
		prevHeaderHash,
		scbp.dataPool.Headers(),
		scbp.marshalizer,
		scbp.store,
	)
	if errNotCritical != nil {
		log.Debug("could not get header with validator stats from storage")
		return
	}

	if header.IsStartOfEpochBlock() {
		log.Debug("trie snapshot",
			"rootHash", header.GetRootHash(),
			"prevRootHash", prevHeader.GetRootHash(),
			"validatorStatsRootHash", header.GetValidatorStatsRootHash())
		scbp.accountsDB[state.UserAccountsState].SnapshotState(header.GetRootHash(), header.GetEpoch())
		scbp.accountsDB[state.PeerAccountsState].SnapshotState(header.GetValidatorStatsRootHash(), header.GetEpoch())

		scbp.markSnapshotDoneInPeerAccounts()

		go func() {
			sovHdr, ok := header.(data.MetaHeaderHandler)
			if !ok {
				log.Warn("cannot commit Trie Epoch Root Hash: last sov header is not of type data.MetaHeaderHandler")
				return
			}
			err := scbp.commitTrieEpochRootHashIfNeeded(sovHdr, header.GetRootHash())
			if err != nil {
				log.Warn("couldn't commit trie checkpoint", "epoch", sovHdr.GetEpoch(), "error", err)
			}
		}()
	}

	scbp.updateStateStorage(
		header,
		header.GetRootHash(),
		prevHeader.GetRootHash(),
		scbp.accountsDB[state.UserAccountsState],
	)

	scbp.updateStateStorage(
		header,
		header.GetValidatorStatsRootHash(),
		prevHeader.GetValidatorStatsRootHash(),
		scbp.accountsDB[state.PeerAccountsState],
	)

	scbp.setFinalizedHeaderHashInIndexer(header.GetPrevHash())
	scbp.blockChain.SetFinalBlockInfo(header.GetNonce(), headerHash, header.GetRootHash())
}

// TODO: (sovereign) remove these cleanUpFunc pointer functions once shardCoordinator will return the correct shard id from task: MX-14132
func (scbp *sovereignChainBlockProcessor) cleanupBlockTrackerPoolsForShard(shardID uint32, noncesToPrevFinal uint64) {
	actualShardID := shardID
	if shardID == core.MetachainShardId {
		actualShardID = core.MainChainShardId
	}
	scbp.baseCleanupBlockTrackerPoolsForShard(actualShardID, noncesToPrevFinal)
}

func (scbp *sovereignChainBlockProcessor) cleanupPoolsForCrossShard(_ uint32, noncesToPrevFinal uint64) {
	scbp.baseCleanupPoolsForCrossShard(core.MainChainShardId, noncesToPrevFinal)
}

func (scbp *sovereignChainBlockProcessor) removeStartOfEpochBlockDataFromPools(
	headerHandler data.HeaderHandler,
	bodyHandler data.BodyHandler,
) error {
	if !headerHandler.IsStartOfEpochBlock() {
		return nil
	}

	sovMetaBlock, ok := headerHandler.(data.MetaHeaderHandler)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	body, ok := bodyHandler.(*block.Body)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	scbp.epochRewardsCreator.RemoveBlockDataFromPools(sovMetaBlock, body)
	scbp.validatorInfoCreator.RemoveBlockDataFromPools(sovMetaBlock, body)

	return nil
}

// IsInterfaceNil returns true if underlying object is nil
func (scbp *sovereignChainBlockProcessor) IsInterfaceNil() bool {
	return scbp == nil
}
