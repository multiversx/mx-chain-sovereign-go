package broadcastFactory

import (
	"fmt"
	"testing"

	commonMock "github.com/multiversx/mx-chain-go/common/mock"
	"github.com/multiversx/mx-chain-go/consensus/broadcast"
	"github.com/multiversx/mx-chain-go/consensus/mock"
	processMock "github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/consensus"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	"github.com/stretchr/testify/require"
)

func createDefaultShardChainArgs() broadcast.ShardChainMessengerArgs {
	return broadcast.ShardChainMessengerArgs{
		CommonMessengerArgs: broadcast.CommonMessengerArgs{
			Marshalizer:                &mock.MarshalizerMock{},
			Hasher:                     &hashingMocks.HasherMock{},
			Messenger:                  &p2pmocks.MessengerStub{},
			ShardCoordinator:           &mock.ShardCoordinatorMock{},
			PeerSignatureHandler:       &mock.PeerSignatureHandler{},
			HeadersSubscriber:          &processMock.HeadersCacherStub{},
			InterceptorsContainer:      &testscommon.InterceptorsContainerStub{},
			MaxDelayCacheSize:          1,
			MaxValidatorDelayCacheSize: 1,
			AlarmScheduler:             &commonMock.AlarmSchedulerStub{},
			KeysHandler:                &testscommon.KeysHandlerStub{},
			DelayedBroadcaster:         &consensus.DelayedBroadcasterMock{},
		},
	}
}

func TestSovereignChainMessengerFactory_CreateShardChainMessenger(t *testing.T) {
	t.Parallel()

	f := NewSovereignShardChainMessengerFactory()
	require.False(t, f.IsInterfaceNil())

	args := createDefaultShardChainArgs()
	msg, err := f.CreateShardChainMessenger(args)
	require.Nil(t, err)
	require.NotNil(t, msg)
	require.Equal(t, "*broadcast.sovereignChainMessenger", fmt.Sprintf("%T", msg))
}
