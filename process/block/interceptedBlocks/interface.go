package interceptedBlocks

import (
	"github.com/multiversx/mx-chain-core-go/data"

	"github.com/multiversx/mx-chain-sovereign-go/sharding"
)

type mbHeadersChecker interface {
	checkMiniBlocksHeaders(mbHeaders []data.MiniBlockHeaderHandler, coordinator sharding.Coordinator) error
}
