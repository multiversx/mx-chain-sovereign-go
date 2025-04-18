package track

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
)

// shardBlockTrack

func (sbt *shardBlockTrack) SetNumPendingMiniBlocks(shardID uint32, numPendingMiniBlocks uint32) {
	sbt.blockBalancer.SetNumPendingMiniBlocks(shardID, numPendingMiniBlocks)
}

func (sbt *shardBlockTrack) GetNumPendingMiniBlocks(shardID uint32) uint32 {
	return sbt.blockBalancer.GetNumPendingMiniBlocks(shardID)
}

func (sbt *shardBlockTrack) SetLastShardProcessedMetaNonce(shardID uint32, nonce uint64) {
	sbt.blockBalancer.SetLastShardProcessedMetaNonce(shardID, nonce)
}

func (sbt *shardBlockTrack) GetLastShardProcessedMetaNonce(shardID uint32) uint64 {
	return sbt.blockBalancer.GetLastShardProcessedMetaNonce(shardID)
}

func (sbt *shardBlockTrack) GetTrackedShardHeaderWithNonceAndHash(shardID uint32, nonce uint64, hash []byte) (data.HeaderHandler, error) {
	return sbt.getTrackedShardHeaderWithNonceAndHash(shardID, nonce, hash)
}

func (sbt *shardBlockTrack) SetBlockProcessor(blockProcessor blockProcessorHandler) {
	sbt.blockProcessor = blockProcessor
}

func (sbt *shardBlockTrack) SetSelfNotarizer(selfNotarizer blockNotarizerHandler) {
	sbt.selfNotarizer = selfNotarizer
}

// sovereignChainShardBlockTrack

func (scsbt *sovereignChainShardBlockTrack) ReceivedExtendedShardHeader(
	extendedShardHeaderHandler data.ShardHeaderExtendedHandler,
	extendedShardHeaderHash []byte,
) {
	scsbt.receivedExtendedShardHeader(extendedShardHeaderHandler, extendedShardHeaderHash)
}

func (scsbt *sovereignChainShardBlockTrack) ShouldAddExtendedShardHeader(extendedShardHeaderHandler data.ShardHeaderExtendedHandler) bool {
	return scsbt.shouldAddExtendedShardHeader(extendedShardHeaderHandler)
}

func (scsbt *sovereignChainShardBlockTrack) DoWhitelistWithExtendedShardHeaderIfNeeded(extendedShardHeaderHandler data.ShardHeaderExtendedHandler) {
	scsbt.doWhitelistWithExtendedShardHeaderIfNeeded(extendedShardHeaderHandler)
}

func (scsbt *sovereignChainShardBlockTrack) IsExtendedShardHeaderOutOfRange(extendedShardHeaderHandler data.ShardHeaderExtendedHandler) bool {
	return scsbt.isExtendedShardHeaderOutOfRange(extendedShardHeaderHandler)
}

func (scsbt *sovereignChainShardBlockTrack) GetFinalHeader(headerHandler data.HeaderHandler) (data.HeaderHandler, error) {
	return scsbt.getFinalHeader(headerHandler)
}

func (scsbt *sovereignChainShardBlockTrack) InitCrossNotarizedStartHeaders() error {
	return scsbt.initCrossNotarizedStartHeaders()
}

// metaBlockTrack

func (mbt *metaBlockTrack) GetTrackedMetaBlockWithHash(hash []byte) (*block.MetaBlock, error) {
	return mbt.getTrackedMetaBlockWithHash(hash)
}

// baseBlockTrack

func (bbt *baseBlockTrack) ReceivedHeader(headerHandler data.HeaderHandler, headerHash []byte) {
	bbt.receivedHeader(headerHandler, headerHash)
}

func CheckTrackerNilParameters(arguments ArgBaseTracker) error {
	return checkTrackerNilParameters(arguments)
}

func (bbt *baseBlockTrack) InitNotarizedHeaders(startHeaders map[uint32]data.HeaderHandler) error {
	return bbt.initNotarizedHeaders(startHeaders)
}

func (bbt *baseBlockTrack) ReceivedShardHeader(headerHandler data.HeaderHandler, shardHeaderHash []byte) {
	bbt.receivedShardHeader(headerHandler, shardHeaderHash)
}

func (bbt *baseBlockTrack) ReceivedMetaBlock(headerHandler data.HeaderHandler, metaBlockHash []byte) {
	bbt.receivedMetaBlock(headerHandler, metaBlockHash)
}

func (bbt *baseBlockTrack) GetMaxNumHeadersToKeepPerShard() int {
	return bbt.maxNumHeadersToKeepPerShard
}

func (bbt *baseBlockTrack) ShouldAddHeaderForCrossShard(headerHandler data.HeaderHandler) bool {
	return bbt.shouldAddHeaderForShard(headerHandler, bbt.crossNotarizer, headerHandler.GetShardID())
}

func (bbt *baseBlockTrack) ShouldAddHeaderForSelfShard(headerHandler data.HeaderHandler) bool {
	return bbt.shouldAddHeaderForShard(headerHandler, bbt.selfNotarizer, core.MetachainShardId)
}

func (bbt *baseBlockTrack) AddHeader(header data.HeaderHandler, hash []byte, shardID uint32) bool {
	return bbt.addHeader(header, hash, shardID)
}

func (bbt *baseBlockTrack) AppendTrackedHeader(headerHandler data.HeaderHandler) {
	bbt.mutHeaders.Lock()
	headersForShard, ok := bbt.headers[headerHandler.GetShardID()]
	if !ok {
		headersForShard = make(map[uint64][]*HeaderInfo)
		bbt.headers[headerHandler.GetShardID()] = headersForShard
	}

	headersForShard[headerHandler.GetNonce()] = append(headersForShard[headerHandler.GetNonce()], &HeaderInfo{Header: headerHandler})
	bbt.mutHeaders.Unlock()
}

func (bbt *baseBlockTrack) CleanupTrackedHeadersBehindNonce(shardID uint32, nonce uint64) {
	bbt.cleanupTrackedHeadersBehindNonce(shardID, nonce)
}

func (bbt *baseBlockTrack) DisplayTrackedHeadersForShard(shardID uint32, message string) {
	bbt.displayTrackedHeadersForShard(shardID, message)
}

func (bbt *baseBlockTrack) SetRoundHandler(roundHandler process.RoundHandler) {
	bbt.roundHandler = roundHandler
}

func (bbt *baseBlockTrack) SetCrossNotarizer(notarizer blockNotarizerHandler) {
	bbt.crossNotarizer = notarizer
}

func (bbt *baseBlockTrack) SetSelfNotarizer(notarizer blockNotarizerHandler) {
	bbt.selfNotarizer = notarizer
}

func (bbt *baseBlockTrack) SetShardCoordinator(coordinator sharding.Coordinator) {
	bbt.shardCoordinator = coordinator
}

func NewBaseBlockTrack() *baseBlockTrack {
	return &baseBlockTrack{}
}

func (bbt *baseBlockTrack) DoWhitelistWithMetaBlockIfNeeded(metaBlock *block.MetaBlock) {
	bbt.doWhitelistWithMetaBlockIfNeeded(metaBlock)
}

func (bbt *baseBlockTrack) DoWhitelistWithShardHeaderIfNeeded(shardHeader *block.Header) {
	bbt.doWhitelistWithShardHeaderIfNeeded(shardHeader)
}

func (bbt *baseBlockTrack) IsHeaderOutOfRange(headerHandler data.HeaderHandler) bool {
	return bbt.isHeaderOutOfRange(headerHandler)
}

func (bbt *baseBlockTrack) SetGetFinalHeaderFunc() {
	bbt.getFinalHeaderFunc = bbt.getFinalHeader
}

func (bbt *baseBlockTrack) ClearStartHeaders() {
	bbt.mutStartHeaders.Lock()
	bbt.startHeaders = make(map[uint32]data.HeaderHandler)
	bbt.mutStartHeaders.Unlock()
}

func (bbt *baseBlockTrack) SetStartHeaders(startHeaders map[uint32]data.HeaderHandler) {
	bbt.mutStartHeaders.Lock()
	bbt.startHeaders = startHeaders
	bbt.mutStartHeaders.Unlock()
}

// blockNotifier

func (bn *blockNotifier) GetNotarizedHeadersHandlers() []func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte) {
	bn.mutNotarizedHeadersHandlers.RLock()
	notarizedHeadersHandlers := bn.notarizedHeadersHandlers
	bn.mutNotarizedHeadersHandlers.RUnlock()

	return notarizedHeadersHandlers
}

// blockNotarizer

func (bn *blockNotarizer) AppendNotarizedHeader(headerHandler data.HeaderHandler) {
	bn.mutNotarizedHeaders.Lock()
	bn.notarizedHeaders[headerHandler.GetShardID()] = append(bn.notarizedHeaders[headerHandler.GetShardID()], &HeaderInfo{Header: headerHandler})
	bn.mutNotarizedHeaders.Unlock()
}

func (bn *blockNotarizer) GetNotarizedHeaders() map[uint32][]*HeaderInfo {
	bn.mutNotarizedHeaders.RLock()
	notarizedHeaders := bn.notarizedHeaders
	bn.mutNotarizedHeaders.RUnlock()

	return notarizedHeaders
}

func (bn *blockNotarizer) GetNotarizedHeaderWithIndex(shardID uint32, index int) data.HeaderHandler {
	bn.mutNotarizedHeaders.RLock()
	notarizedHeader := bn.notarizedHeaders[shardID][index].Header
	bn.mutNotarizedHeaders.RUnlock()

	return notarizedHeader
}

func (bn *blockNotarizer) LastNotarizedHeaderInfo(shardID uint32) *HeaderInfo {
	return bn.lastNotarizedHeaderInfo(shardID)
}

// blockProcessor

func (bp *blockProcessor) DoJobOnReceivedHeader(shardID uint32) {
	bp.doJobOnReceivedHeader(shardID)
}

func (bp *blockProcessor) DoJobOnReceivedCrossNotarizedHeader(shardID uint32) {
	bp.doJobOnReceivedCrossNotarizedHeaderFunc(shardID)
}

func (bp *blockProcessor) ComputeLongestChainFromLastCrossNotarized(shardID uint32) (data.HeaderHandler, []byte, []data.HeaderHandler, [][]byte) {
	return bp.computeLongestChainFromLastCrossNotarized(shardID)
}

func (bp *blockProcessor) ComputeSelfNotarizedHeaders(headers []data.HeaderHandler) ([]data.HeaderHandler, [][]byte) {
	return bp.computeSelfNotarizedHeaders(headers)
}

func (bp *blockProcessor) GetNextHeader(longestChainHeadersIndexes *[]int, headersIndexes []int, prevHeader data.HeaderHandler, sortedHeaders []data.HeaderHandler, index int) {
	bp.getNextHeader(longestChainHeadersIndexes, headersIndexes, prevHeader, sortedHeaders, index)
}

func (bp *blockProcessor) CheckHeaderFinality(header data.HeaderHandler, sortedHeaders []data.HeaderHandler, index int) error {
	return bp.checkHeaderFinality(header, sortedHeaders, index)
}

func (bp *blockProcessor) RequestHeadersIfNeeded(lastNotarizedHeader data.HeaderHandler, sortedHeaders []data.HeaderHandler, longestChainHeaders []data.HeaderHandler, shardID uint32) {
	bp.requestHeadersIfNeeded(lastNotarizedHeader, sortedHeaders, longestChainHeaders, shardID)
}

func (bp *blockProcessor) GetLatestValidHeader(lastNotarizedHeader data.HeaderHandler, longestChainHeaders []data.HeaderHandler) data.HeaderHandler {
	return bp.getLatestValidHeader(lastNotarizedHeader, longestChainHeaders)
}

func (bp *blockProcessor) GetHighestRoundInReceivedHeaders(latestValidHeader data.HeaderHandler, sortedReceivedHeaders []data.HeaderHandler) uint64 {
	return bp.getHighestRoundInReceivedHeaders(latestValidHeader, sortedReceivedHeaders)
}

func (bp *blockProcessor) RequestHeadersIfNothingNewIsReceived(lastNotarizedHeaderNonce uint64, latestValidHeader data.HeaderHandler, highestRoundInReceivedHeaders uint64, shardID uint32) {
	bp.requestHeadersIfNothingNewIsReceived(lastNotarizedHeaderNonce, latestValidHeader, highestRoundInReceivedHeaders, shardID)
}

func (bp *blockProcessor) RequestHeaders(shardID uint32, fromNonce uint64) {
	bp.requestHeaders(shardID, fromNonce)
}

func (bp *blockProcessor) ShouldProcessReceivedHeader(headerHandler data.HeaderHandler) bool {
	return bp.shouldProcessReceivedHeaderFunc(headerHandler)
}

// sovereignChainBlockProcessor

func (scbp *sovereignChainBlockProcessor) ShouldProcessReceivedHeader(headerHandler data.HeaderHandler) bool {
	return scbp.shouldProcessReceivedHeaderFunc(headerHandler)
}

func (scbp *sovereignChainBlockProcessor) GetBlockFinality() uint64 {
	return scbp.blockFinality
}

// miniBlockTrack

func (mbt *miniBlockTrack) ReceivedMiniBlock(key []byte, value interface{}) {
	mbt.receivedMiniBlock(key, value)
}

func (mbt *miniBlockTrack) GetTransactionPool(mbType block.Type) dataRetriever.ShardedDataCacherNotifier {
	return mbt.getTransactionPool(mbType)
}

func (mbt *miniBlockTrack) SetBlockTransactionsPool(blockTransactionsPool dataRetriever.ShardedDataCacherNotifier) {
	mbt.blockTransactionsPool = blockTransactionsPool
}
