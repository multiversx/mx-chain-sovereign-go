package track

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/sovereign/dto"

	"github.com/multiversx/mx-chain-go/common/runType"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
)

type sovereignBlockTrackerFactory struct {
	blockTrackerCreator BlockTrackerCreator
	orderedChainIDs     []dto.ChainID
}

// NewSovereignBlockTrackerFactory creates a new shard block tracker factory for sovereign chain
func NewSovereignBlockTrackerFactory(btc BlockTrackerCreator, mainChainNotarization map[string]config.MainChainNotarization) (*sovereignBlockTrackerFactory, error) {
	if check.IfNil(btc) {
		return nil, process.ErrNilBlockTrackerCreator
	}

	orderedChainIDs, err := runType.GetOrderedCrossChainIDs(mainChainNotarization)
	if err != nil {
		return nil, err
	}

	return &sovereignBlockTrackerFactory{
		blockTrackerCreator: btc,
		orderedChainIDs:     orderedChainIDs,
	}, nil
}

// CreateBlockTracker creates a new block tracker for sovereign chain
func (sbtcf *sovereignBlockTrackerFactory) CreateBlockTracker(argBaseTracker ArgShardTracker) (process.BlockTracker, error) {
	blockTracker, err := sbtcf.blockTrackerCreator.CreateBlockTracker(argBaseTracker)
	if err != nil {
		return nil, err
	}
	shardBlockTracker, ok := blockTracker.(*shardBlockTrack)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	return NewSovereignChainShardBlockTrack(shardBlockTracker, sbtcf.orderedChainIDs)
}

// IsInterfaceNil returns true if there is no value under the interface
func (sbtcf *sovereignBlockTrackerFactory) IsInterfaceNil() bool {
	return sbtcf == nil
}
