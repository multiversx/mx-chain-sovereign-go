package track

import (
	"errors"
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/sovereign/dto"

	"github.com/multiversx/mx-chain-go/process"
)

type sovereignChainShardBlockTrack struct {
	*shardBlockTrack
}

// NewSovereignChainShardBlockTrack creates an object for tracking the received shard blocks
func NewSovereignChainShardBlockTrack(shardBlockTrack *shardBlockTrack) (*sovereignChainShardBlockTrack, error) {
	if shardBlockTrack == nil {
		return nil, process.ErrNilBlockTracker
	}

	scsbt := &sovereignChainShardBlockTrack{
		shardBlockTrack,
	}

	bp, ok := scsbt.blockProcessor.(*blockProcessor)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	scbp, err := NewSovereignChainBlockProcessor(bp)
	if err != nil {
		return nil, err
	}

	scsbt.blockProcessor = scbp
	scsbt.doReceivedHeaderJobFunc = scsbt.doReceivedHeaderJob
	scsbt.getFinalHeaderFunc = scsbt.getFinalHeader

	err = scsbt.initCrossNotarizedStartHeaders()
	if err != nil {
		return nil, err
	}

	return scsbt, nil
}

func (scsbt *sovereignChainShardBlockTrack) initCrossNotarizedStartHeaders() error {
	// TODO MX-15667:
	// 1. Versioning header for shard extended?
	// 2. Perhaps add dummy header nonce from config cross main chain notarized

	for chainID := range dto.ValidChains {
		extendedShardHeader := &block.ShardHeaderExtended{
			Header: &block.HeaderV2{
				Header: &block.Header{},
			},
		}

		extendedShardHeaderHash, err := core.CalculateHash(scsbt.marshalizer, scsbt.hasher, extendedShardHeader)
		if err != nil {
			return fmt.Errorf("%w in sovereignChainShardBlockTrack.initCrossNotarizedStartHeaders", err)
		}

		scsbt.AddCrossNotarizedHeader(uint32(chainID), extendedShardHeader, extendedShardHeaderHash)
	}

	return nil
}

// ComputeLongestSelfChain computes the longest chain from self shard
func (scsbt *sovereignChainShardBlockTrack) ComputeLongestSelfChain() (data.HeaderHandler, []byte, []data.HeaderHandler, [][]byte) {
	lastSelfNotarizedHeader, lastSelfNotarizedHeaderHash, err := scsbt.selfNotarizer.GetLastNotarizedHeader(scsbt.shardCoordinator.SelfId())
	if err != nil {
		log.Warn("ComputeLongestSelfChain.GetLastNotarizedHeader", "error", err.Error())
		return nil, nil, nil, nil
	}

	headers, hashes := scsbt.ComputeLongestChain(scsbt.shardCoordinator.SelfId(), lastSelfNotarizedHeader)
	return lastSelfNotarizedHeader, lastSelfNotarizedHeaderHash, headers, hashes
}

// GetSelfNotarizedHeader returns a self notarized header for self shard with a given offset, behind last self notarized header
func (scsbt *sovereignChainShardBlockTrack) GetSelfNotarizedHeader(_ uint32, offset uint64) (data.HeaderHandler, []byte, error) {
	return scsbt.selfNotarizer.GetNotarizedHeader(scsbt.shardCoordinator.SelfId(), offset)
}

func (scsbt *sovereignChainShardBlockTrack) doReceivedHeaderJob(headerHandler data.HeaderHandler, headerHash []byte) {
	extendedShardHeader, isExtendedShardHeaderReceived := headerHandler.(*block.ShardHeaderExtended)
	if isExtendedShardHeaderReceived {
		scsbt.receivedExtendedShardHeader(extendedShardHeader, headerHash)
		return
	}

	scsbt.receivedShardHeader(headerHandler, headerHash)
}

func (scsbt *sovereignChainShardBlockTrack) receivedExtendedShardHeader(
	extendedShardHeaderHandler data.ShardHeaderExtendedHandler,
	extendedShardHeaderHash []byte,
) {

	log.Debug("received extended shard header from network in block tracker",
		"shard", extendedShardHeaderHandler.GetShardID(),
		"epoch", extendedShardHeaderHandler.GetEpoch(),
		"round", extendedShardHeaderHandler.GetRound(),
		"nonce", extendedShardHeaderHandler.GetNonce(),
		"hash", extendedShardHeaderHash,
	)

	// TODO: This condition will permit to the sovereign chain to follow the main chain headers starting with a header
	// having a nonce higher than nonce 1 (the first block after genesis)
	if scsbt.IsGenesisLastCrossNotarizedHeader(extendedShardHeaderHandler.GetSourceChainID()) {
		scsbt.crossNotarizer.AddNotarizedHeader(uint32(extendedShardHeaderHandler.GetSourceChainID()), extendedShardHeaderHandler, extendedShardHeaderHash)
	}

	if !scsbt.shouldAddExtendedShardHeader(extendedShardHeaderHandler) {
		log.Trace("received extended shard header is out of range", "nonce", extendedShardHeaderHandler.GetNonce())
		return
	}

	if !scsbt.addHeader(extendedShardHeaderHandler, extendedShardHeaderHash, uint32(extendedShardHeaderHandler.GetSourceChainID())) {
		log.Trace("received extended shard header was not added", "nonce", extendedShardHeaderHandler.GetNonce())
		return
	}

	scsbt.doWhitelistWithExtendedShardHeaderIfNeeded(extendedShardHeaderHandler)
	scsbt.blockProcessor.ProcessReceivedHeader(extendedShardHeaderHandler)
}

// IsGenesisLastCrossNotarizedHeader returns true if the last cross chain notarized header is the dummy genesis header
func (scsbt *sovereignChainShardBlockTrack) IsGenesisLastCrossNotarizedHeader(chainID dto.ChainID) bool {
	lastNotarizedHeader, _, err := scsbt.crossNotarizer.GetLastNotarizedHeader(uint32(chainID))

	isGenesisLastCrossNotarizedHeader := err != nil && errors.Is(err, process.ErrNotarizedHeadersSliceForShardIsNil) ||
		lastNotarizedHeader != nil && lastNotarizedHeader.GetNonce() == 0

	return isGenesisLastCrossNotarizedHeader
}

func (scsbt *sovereignChainShardBlockTrack) shouldAddExtendedShardHeader(extendedShardHeaderHandler data.ShardHeaderExtendedHandler) bool {
	lastNotarizedHeader, _, err := scsbt.crossNotarizer.GetLastNotarizedHeader(uint32(extendedShardHeaderHandler.GetSourceChainID()))
	if err != nil {
		log.Debug("shouldAddExtendedShardHeader.GetLastNotarizedHeader",
			"shard", extendedShardHeaderHandler.GetShardID(),
			"error", err.Error())
		return false
	}

	lastNotarizedHeaderNonce := lastNotarizedHeader.GetNonce()

	isHeaderOutOfRange := extendedShardHeaderHandler.GetNonce() > lastNotarizedHeaderNonce+uint64(scsbt.maxNumHeadersToKeepPerShard)
	return !isHeaderOutOfRange
}

func (scsbt *sovereignChainShardBlockTrack) doWhitelistWithExtendedShardHeaderIfNeeded(extendedShardHeaderHandler data.ShardHeaderExtendedHandler) {
	if check.IfNil(extendedShardHeaderHandler) {
		return
	}
	if scsbt.isExtendedShardHeaderOutOfRange(extendedShardHeaderHandler) {
		return
	}

	miniBlockHandlers := extendedShardHeaderHandler.GetIncomingMiniBlockHandlers()

	keys := make([][]byte, 0)
	for _, miniBlockHandler := range miniBlockHandlers {
		miniBlock, ok := miniBlockHandler.(*block.MiniBlock)
		if !ok {
			log.Warn("doWhitelistWithExtendedShardHeaderIfNeeded", "error", process.ErrWrongTypeAssertion)
			continue
		}

		keys = append(keys, miniBlock.TxHashes...)
	}

	scsbt.whitelistHandler.Add(keys)
}

func (scsbt *sovereignChainShardBlockTrack) isExtendedShardHeaderOutOfRange(extendedShardHeaderHandler data.ShardHeaderExtendedHandler) bool {
	lastCrossNotarizedHeader, _, err := scsbt.GetLastCrossNotarizedHeader(uint32(extendedShardHeaderHandler.GetSourceChainID()))
	if err != nil {
		log.Debug("isExtendedShardHeaderOutOfRange.GetLastCrossNotarizedHeader",
			"shard", extendedShardHeaderHandler.GetShardID(),
			"error", err.Error())
		return true
	}

	isExtendedShardHeaderOutOfRange := extendedShardHeaderHandler.GetNonce() > lastCrossNotarizedHeader.GetNonce()+process.MaxHeadersToWhitelistInAdvance
	return isExtendedShardHeaderOutOfRange
}

// ComputeLongestExtendedShardChainFromLastNotarized returns the longest valid chain for extended shard chain from its last cross notarized header
func (scsbt *sovereignChainShardBlockTrack) ComputeLongestExtendedShardChainFromLastNotarized(chainID dto.ChainID) ([]data.HeaderHandler, [][]byte, error) {
	lastCrossNotarizedHeader, _, err := scsbt.GetLastCrossNotarizedHeader(uint32(chainID))
	if err != nil {
		return nil, nil, err
	}

	log.Debug("sovereignChainShardBlockTrack.ComputeLongestExtendedShardChainFromLastNotarized: GetLastCrossNotarizedHeader", "nonce", lastCrossNotarizedHeader.GetNonce())

	hdrsForShard, hdrsHashesForShard := scsbt.ComputeLongestChain(uint32(chainID), lastCrossNotarizedHeader)

	log.Debug("sovereignChainShardBlockTrack.ComputeLongestExtendedShardChainFromLastNotarized: ComputeLongestChain", "num headers", len(hdrsForShard))
	for index := range hdrsForShard {
		log.Debug("sovereignChainShardBlockTrack.ComputeLongestExtendedShardChainFromLastNotarized", "round", hdrsForShard[index].GetRound(), "nonce", hdrsForShard[index].GetNonce(), "hash", hdrsHashesForShard[index])
	}

	return hdrsForShard, hdrsHashesForShard, nil
}

// CleanupHeadersBehindNonce removes from local pools old headers
func (scsbt *sovereignChainShardBlockTrack) CleanupHeadersBehindNonce(
	shardID uint32,
	selfNotarizedNonce uint64,
	crossNotarizedNonce uint64,
) {
	scsbt.cleanupHeadersBehindNonce(shardID, core.MainChainShardId, selfNotarizedNonce, crossNotarizedNonce)
}

// DisplayTrackedHeaders displays tracked headers
func (scsbt *sovereignChainShardBlockTrack) DisplayTrackedHeaders() {
	scsbt.displayTrackedHeadersForShard(scsbt.shardCoordinator.SelfId(), "tracked headers")
	scsbt.selfNotarizer.DisplayNotarizedHeaders(scsbt.shardCoordinator.SelfId(), "self notarized headers")

	for chainID := range dto.ValidChains {
		scsbt.displayTrackedHeadersForShard(uint32(chainID), fmt.Sprintf("cross tracked headers, chain: %s", dto.ChainID_name[int32(chainID)]))
		scsbt.crossNotarizer.DisplayNotarizedHeaders(uint32(chainID), fmt.Sprintf("cross notarized headers, chain: %s", dto.ChainID_name[int32(chainID)]))
	}
}

func (scsbt *sovereignChainShardBlockTrack) getFinalHeader(headerHandler data.HeaderHandler) (data.HeaderHandler, error) {
	shardID := headerHandler.GetShardID()

	extendedHeader, isExtendedShardHeaderReceived := headerHandler.(*block.ShardHeaderExtended)
	if isExtendedShardHeaderReceived {
		shardID = uint32(extendedHeader.GetSourceChainID())
	}

	finalHeader, _, err := scsbt.getFinalHeaderForShard(shardID)
	if err != nil {
		return nil, err
	}

	return finalHeader, nil
}

// ShouldSkipMiniBlocksCreationFromSelf returns false for sovereign chain
func (scsbt *sovereignChainShardBlockTrack) ShouldSkipMiniBlocksCreationFromSelf() bool {
	return false
}

// IsShardStuck returns false for sovereign chain
func (scsbt *sovereignChainShardBlockTrack) IsShardStuck(_ uint32) bool {
	return false
}

// ComputeCrossInfo does nothing for sovereign chain
func (scsbt *sovereignChainShardBlockTrack) ComputeCrossInfo(_ []data.HeaderHandler) {
}

// RemoveLastCrossNotarizedHeaders removes last cross notarized headers from tracker list
func (scsbt *sovereignChainShardBlockTrack) RemoveLastCrossNotarizedHeaders() {
	scsbt.crossNotarizer.RemoveLastNotarizedHeader()
}

// RemoveLastCrossNotarizedHeader removes last cross notarized header for a specific chain from tracker list
func (scsbt *sovereignChainShardBlockTrack) RemoveLastCrossNotarizedHeader(chainID dto.ChainID) {
	scsbt.crossNotarizer.RemoveLastNotarizedHeaderForShard(uint32(chainID))
}

// RemoveLastSelfNotarizedHeaders removes last self notarized headers from tracker list
func (scsbt *sovereignChainShardBlockTrack) RemoveLastSelfNotarizedHeaders() {
	scsbt.selfNotarizer.RemoveLastNotarizedHeader()
}
