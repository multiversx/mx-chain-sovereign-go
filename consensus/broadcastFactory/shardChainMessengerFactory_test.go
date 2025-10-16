package broadcastFactory

import (
	"fmt"
	"testing"

	commonMock "github.com/multiversx/mx-chain-go/common/mock"
	"github.com/multiversx/mx-chain-go/consensus/broadcast"
	"github.com/multiversx/mx-chain-go/consensus/mock"
	processMock "github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/require"
)

func createDefaultDelayedBlockBroadcasterArgs() *broadcast.ArgsDelayedBlockBroadcaster {
	return &broadcast.ArgsDelayedBlockBroadcaster{
		InterceptorsContainer: &testscommon.InterceptorsContainerStub{},
		HeadersSubscriber:     &processMock.HeadersCacherStub{},
		ShardCoordinator:      &mock.ShardCoordinatorMock{},
		LeaderCacheSize:       10,
		ValidatorCacheSize:    10,
		AlarmScheduler:        &commonMock.AlarmSchedulerStub{},
	}
}

func TestChainMessengerFactory_CreateShardChainMessenger(t *testing.T) {
	t.Parallel()

	f := NewShardChainMessengerFactory()
	require.False(t, f.IsInterfaceNil())

	args := createDefaultShardChainArgs()
	msg, err := f.CreateShardChainMessenger(args)
	require.Nil(t, err)
	require.NotNil(t, msg)
	require.Equal(t, "*broadcast.shardChainMessenger", fmt.Sprintf("%T", msg))
}

func TestChainMessengerFactory_CreateDelayedBlockBroadcaster(t *testing.T) {
	t.Parallel()

	f := NewShardChainMessengerFactory()
	args := createDefaultDelayedBlockBroadcasterArgs()
	dbb, err := f.CreateDelayedBlockBroadcaster(args)
	require.Nil(t, err)
	require.NotNil(t, dbb)
	require.Equal(t, "*broadcast.delayedBlockBroadcaster", fmt.Sprintf("%T", dbb))
}
