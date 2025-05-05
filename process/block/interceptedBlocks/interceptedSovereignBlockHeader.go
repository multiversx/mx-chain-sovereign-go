package interceptedBlocks

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/sovereign/dto"

	"github.com/multiversx/mx-chain-go/sharding"
)

type interceptedSovereignBlockHeader struct {
	*InterceptedHeader
}

// NewSovereignInterceptedBlockHeader creates a new intercepted sovereign block header
func NewSovereignInterceptedBlockHeader(arg *ArgInterceptedBlockHeader) (*interceptedSovereignBlockHeader, error) {
	interceptedHdr, err := NewInterceptedHeader(arg)
	if err != nil {
		return nil, err
	}

	crossChainIDs := make(map[uint32]struct{})
	for chainID := range dto.ValidChains {
		crossChainIDs[uint32(chainID)] = struct{}{}
	}

	interceptedHdr.acceptedCrossShardIDs = crossChainIDs
	sovInterceptedBlock := &interceptedSovereignBlockHeader{
		interceptedHdr,
	}

	sovInterceptedBlock.mbHeadersChecker = sovInterceptedBlock
	return sovInterceptedBlock, nil
}

// TODO: Here, we can even get rid of this func now
func (isbh *interceptedSovereignBlockHeader) checkMiniBlocksHeaders(mbHeaders []data.MiniBlockHeaderHandler, coordinator sharding.Coordinator) error {
	return checkMiniBlocksHeaders(mbHeaders, coordinator, isbh.acceptedCrossShardIDs)
}

// IsInterfaceNil returns true if there is no value under the interface
func (isbh *interceptedSovereignBlockHeader) IsInterfaceNil() bool {
	return isbh == nil
}
