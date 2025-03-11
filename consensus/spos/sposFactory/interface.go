package sposFactory

import (
	"github.com/multiversx/mx-chain-sovereign-go/consensus"
	"github.com/multiversx/mx-chain-sovereign-go/consensus/broadcast"
)

// BroadCastShardMessengerFactoryHandler defines a shard messenger factory handler
type BroadCastShardMessengerFactoryHandler interface {
	CreateShardChainMessenger(args broadcast.ShardChainMessengerArgs) (consensus.BroadcastMessenger, error)
	IsInterfaceNil() bool
}
