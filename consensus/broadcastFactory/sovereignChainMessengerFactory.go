package broadcastFactory

import (
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/broadcast"
)

type sovereignChainMessengerFactory struct {
}

// NewSovereignShardChainMessengerFactory creates a sovereign shard messenger factory
func NewSovereignShardChainMessengerFactory() *sovereignChainMessengerFactory {
	return &sovereignChainMessengerFactory{}
}

// CreateShardChainMessenger creates a shard messenger for sovereign chain run type
func (f *sovereignChainMessengerFactory) CreateShardChainMessenger(args broadcast.ShardChainMessengerArgs) (consensus.BroadcastMessenger, error) {
	argsDelayedBlockBroadcaster := &broadcast.ArgsDelayedBlockBroadcaster{
		InterceptorsContainer: args.InterceptorsContainer,
		HeadersSubscriber:     args.HeadersSubscriber,
		ShardCoordinator:      args.ShardCoordinator,
		LeaderCacheSize:       args.MaxDelayCacheSize,
		ValidatorCacheSize:    args.MaxValidatorDelayCacheSize,
		AlarmScheduler:        args.AlarmScheduler,
	}
	sovDelayedBroadcaster, err := broadcast.NewSovereignDelayedBlockBroadcaster(argsDelayedBlockBroadcaster)
	if err != nil {
		return nil, err
	}

	argsSovereignShardChainMessenger := broadcast.ArgsSovereignShardChainMessenger{
		DelayedBroadcaster:   sovDelayedBroadcaster,
		Marshaller:           args.Marshalizer,
		Hasher:               args.Hasher,
		ShardCoordinator:     args.ShardCoordinator,
		Messenger:            args.Messenger,
		PeerSignatureHandler: args.PeerSignatureHandler,
		KeysHandler:          args.KeysHandler,
	}
	return broadcast.NewSovereignShardChainMessenger(argsSovereignShardChainMessenger)
}

// CreateDelayedBlockBroadcaster creates a delayed block broadcaster for sovereign chain run type
func (f *sovereignChainMessengerFactory) CreateDelayedBlockBroadcaster(args *broadcast.ArgsDelayedBlockBroadcaster) (broadcast.DelayedBroadcaster, error) {
	return broadcast.NewSovereignDelayedBlockBroadcaster(args)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (f *sovereignChainMessengerFactory) IsInterfaceNil() bool {
	return f == nil
}
