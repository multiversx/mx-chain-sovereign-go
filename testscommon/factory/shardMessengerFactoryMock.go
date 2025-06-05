package factory

import (
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/broadcast"
	cnsMock "github.com/multiversx/mx-chain-go/testscommon/consensus"
)

// ShardChainMessengerFactoryMock -
type ShardChainMessengerFactoryMock struct {
	CreateShardChainMessengerCalled     func(args broadcast.ShardChainMessengerArgs) (consensus.BroadcastMessenger, error)
	CreateDelayedBlockBroadcasterCalled func(args *broadcast.ArgsDelayedBlockBroadcaster) (broadcast.DelayedBroadcaster, error)
}

// CreateShardChainMessenger -
func (mock *ShardChainMessengerFactoryMock) CreateShardChainMessenger(args broadcast.ShardChainMessengerArgs) (consensus.BroadcastMessenger, error) {
	if mock.CreateShardChainMessengerCalled != nil {
		return mock.CreateShardChainMessengerCalled(args)
	}

	return &cnsMock.BroadcastMessengerMock{}, nil
}

// CreateDelayedBlockBroadcaster -
func (mock *ShardChainMessengerFactoryMock) CreateDelayedBlockBroadcaster(args *broadcast.ArgsDelayedBlockBroadcaster) (broadcast.DelayedBroadcaster, error) {
	if mock.CreateDelayedBlockBroadcasterCalled != nil {
		return mock.CreateDelayedBlockBroadcasterCalled(args)
	}

	return &cnsMock.DelayedBroadcasterMock{}, nil
}

// IsInterfaceNil -
func (mock *ShardChainMessengerFactoryMock) IsInterfaceNil() bool {
	return mock == nil
}
