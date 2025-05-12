package block

import (
	"github.com/multiversx/mx-chain-core-go/data/sovereign/dto"

	"github.com/multiversx/mx-chain-go/process/block/bootstrapStorage"
)

type sovereignShardCrossNotarizer struct {
	*baseBlockNotarizer
	orderedChainIDs []dto.ChainID
}

func (scn *sovereignShardCrossNotarizer) getLastCrossNotarizedHeaders() []bootstrapStorage.BootstrapHeaderInfo {
	lastCrossNotarizedHeaders := make([]bootstrapStorage.BootstrapHeaderInfo, 0)
	for _, chainID := range scn.orderedChainIDs {
		bootstrapHeaderInfo := scn.getLastCrossNotarizedHeadersForShard(uint32(chainID))
		if bootstrapHeaderInfo == nil {
			continue
		}

		bootstrapHeaderInfo.ShardId = uint32(chainID)
		lastCrossNotarizedHeaders = append(lastCrossNotarizedHeaders, *bootstrapHeaderInfo)
	}

	return trimSliceBootstrapHeaderInfo(lastCrossNotarizedHeaders)
}
