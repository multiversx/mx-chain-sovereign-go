package sharding

import "github.com/multiversx/mx-chain-core-go/core"

type sovereignShardCoordinator struct {
	*multiShardCoordinator
}

// TODO: Here (MX-16867), analyse if it is useful to have another constructor here with ordered chain ids
// and check if it would be useful in the future to treat other shards or numShards as chain ids

// NewSovereignShardCoordinator creates a new sovereign shard coordinator
func NewSovereignShardCoordinator() *sovereignShardCoordinator {
	sr := &sovereignShardCoordinator{
		multiShardCoordinator: &multiShardCoordinator{
			selfId:         core.SovereignChainShardId,
			numberOfShards: 1,
		},
	}
	sr.maskHigh, sr.maskLow = sr.calculateMasks()
	return sr
}

// ComputeId returns the self sovereign shard id
func (ssc *sovereignShardCoordinator) ComputeId(_ []byte) uint32 {
	return ssc.selfId
}

// SameShard returns true if the addresses are in the same shard
func (msc *sovereignShardCoordinator) SameShard(_, _ []byte) bool {
	return true
}

// TotalNumberOfShards returns one for sovereign chain
func (msc *sovereignShardCoordinator) TotalNumberOfShards() uint32 {
	return 1
}

// IsInterfaceNil returns true if there is no value under the interface
func (msc *sovereignShardCoordinator) IsInterfaceNil() bool {
	return msc == nil
}
