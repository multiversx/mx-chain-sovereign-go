package storageBootstrap

import (
	"github.com/multiversx/mx-chain-core-go/data"
	dtoSov "github.com/multiversx/mx-chain-core-go/data/sovereign/dto"

	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/bootstrapStorage"
)

type sovereignChainShardStorageBootstrapper struct {
	*shardStorageBootstrapper
}

// NewSovereignChainShardStorageBootstrapper creates a new instance of sovereignChainShardStorageBootstrapper
func NewSovereignChainShardStorageBootstrapper(shardStorageBootstrapper *shardStorageBootstrapper) (*sovereignChainShardStorageBootstrapper, error) {
	if shardStorageBootstrapper == nil {
		return nil, process.ErrNilShardStorageBootstrapper
	}

	scssb := &sovereignChainShardStorageBootstrapper{
		shardStorageBootstrapper,
	}

	scssb.getScheduledRootHashMethod = scssb.sovereignChainGetScheduledRootHash
	scssb.setScheduledInfoMethod = scssb.sovereignChainSetScheduledInfo

	scssb.bootstrapper = scssb
	return scssb, nil
}

func (ssb *sovereignChainShardStorageBootstrapper) applyCrossNotarizedHeaders(crossNotarizedHeaders []bootstrapStorage.BootstrapHeaderInfo) error {
	for _, crossNotarizedHeader := range crossNotarizedHeaders {
		crossChainID := dtoSov.ChainID(crossNotarizedHeader.ShardId)
		if !dtoSov.IsValidCrossChainID(crossChainID) {
			continue
		}

		extendedHeader, err := process.GetExtendedShardHeaderFromStorage(crossNotarizedHeader.Hash, ssb.marshalizer, ssb.store)
		if err != nil {
			return err
		}

		log.Debug("added cross notarized header in block tracker",
			"shard", crossChainID.String(),
			"round", extendedHeader.GetRound(),
			"nonce", extendedHeader.GetNonce(),
			"hash", crossNotarizedHeader.Hash)

		ssb.blockTracker.AddCrossNotarizedHeader(uint32(crossChainID), extendedHeader, crossNotarizedHeader.Hash)
		ssb.blockTracker.AddTrackedHeader(extendedHeader, crossNotarizedHeader.Hash)
	}

	return nil
}

func (ssb *sovereignChainShardStorageBootstrapper) cleanupNotarizedStorage(shardHeaderHash []byte) {
	log.Debug("sovereign cleanup notarized storage")

	shardHeader, err := process.GetShardHeaderFromStorage(shardHeaderHash, ssb.marshalizer, ssb.store)
	if err != nil {
		log.Debug("sovereign shard header is not found in BlockHeaderUnit storage",
			"hash", shardHeaderHash)
		return
	}

	sovereignHeader, castOk := shardHeader.(data.SovereignChainHeaderHandler)
	if !castOk {
		log.Warn("sovereignChainShardStorageBootstrapper.cleanupNotarizedStorage",
			"error", process.ErrWrongTypeAssertion,
			"expected shard header of type", "SovereignChainHeaderHandler",
		)
		return
	}

	for _, chainData := range sovereignHeader.GetChainDataHandlers() {
		ssb.cleanupNotarizedStorageForChain(chainData)
	}
}

func (ssb *sovereignChainShardStorageBootstrapper) cleanupNotarizedStorageForChain(chainData data.ChainDataHandler) {
	for _, extendedHeaderHash := range chainData.GetExtendedShardHeaderHashes() {
		extendedHeader, err := process.GetExtendedShardHeaderFromStorage(extendedHeaderHash, ssb.marshalizer, ssb.store)
		if err != nil {
			log.Debug("extended block is not found in ExtendedShardHeadersUnit storage",
				"hash", extendedHeaderHash)
			continue
		}

		log.Debug("removing extended header from storage",
			"shardId", extendedHeader.GetShardID(),
			"nonce", extendedHeader.GetNonce(),
			"hash", extendedHeaderHash)

		ssb.removeHdrFromHeaderNonceToHashUnit(extendedHeader, extendedHeaderHash, dataRetriever.ExtendedShardHeadersNonceHashDataUnit)
		ssb.removeBlockFromBlockUnit(extendedHeader, extendedHeaderHash, dataRetriever.ExtendedShardHeadersUnit)
	}
}

func (ssb *sovereignChainShardStorageBootstrapper) cleanupNotarizedStorageForHigherNoncesIfExist(
	crossNotarizedHeaders []bootstrapStorage.BootstrapHeaderInfo,
) {
	for supportedChainID := range dtoSov.ValidChains {
		lastCrossNotarizedNonce, err := getLastCrossNotarizedHeaderNonce(crossNotarizedHeaders, uint32(supportedChainID))
		if err != nil {
			log.Warn("cleanupNotarizedStorageForHigherNoncesIfExist", "chainID", supportedChainID.String(), "error", err.Error())
			continue
		}

		ssb.cleanupCrossChainNotarizedStorage(lastCrossNotarizedNonce)
	}
}

func (ssb *sovereignChainShardStorageBootstrapper) cleanupCrossChainNotarizedStorage(lastCrossNotarizedNonce uint64) {
	log.Debug("cleanup notarized storage has been started", "from nonce", lastCrossNotarizedNonce+1)
	nonce := lastCrossNotarizedNonce

	var numConsecutiveNoncesNotFound int
	for {
		nonce++

		extendedBlock, extendedBlockHash, err := process.GetExtendedHeaderFromStorageWithNonce(
			nonce,
			ssb.store,
			ssb.uint64Converter,
			ssb.marshalizer,
		)
		if err != nil {
			log.Debug("sovereignChainShardStorageBootstrapper.cleanupNotarizedStorageForHigherNoncesIfExist:"+
				"trying to cleanup an extended header from storage that is not found",
				"nonce", nonce, "error", err.Error())

			numConsecutiveNoncesNotFound++
			if numConsecutiveNoncesNotFound > maxNumOfConsecutiveNoncesNotFoundAccepted {
				log.Debug("cleanup notarized storage has been finished",
					"from nonce", lastCrossNotarizedNonce+1,
					"to nonce", nonce)
				break
			}

			continue
		}

		numConsecutiveNoncesNotFound = 0

		log.Debug("removing extended block from storage",
			"shardId", extendedBlock.GetShardID(),
			"nonce", extendedBlock.GetNonce(),
			"hash", extendedBlockHash)

		ssb.removeHdrFromHeaderNonceToHashUnit(extendedBlock, extendedBlockHash, dataRetriever.ExtendedShardHeadersNonceHashDataUnit)
		ssb.removeBlockFromBlockUnit(extendedBlock, extendedBlockHash, dataRetriever.ExtendedShardHeadersUnit)
	}
}
