package process

import (
	"github.com/multiversx/mx-chain-core-go/data/block"

	"github.com/multiversx/mx-chain-sovereign-go/genesis"
	"github.com/multiversx/mx-chain-sovereign-go/update"
)

type headerCreatorArgs struct {
	mapArgsGenesisBlockCreator map[uint32]ArgsGenesisBlockCreator
	mapHardForkBlockProcessor  map[uint32]update.HardForkBlockProcessor
	mapBodies                  map[uint32]*block.Body
	shardIDs                   []uint32
	nodesListSplitter          genesis.NodesListSplitter
}
