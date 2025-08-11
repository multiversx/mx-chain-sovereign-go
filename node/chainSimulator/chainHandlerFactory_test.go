package chainSimulator

import (
	"testing"

	heartbeatMonitor "github.com/multiversx/mx-chain-go/heartbeat"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/heartbeat"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/process"
	"github.com/multiversx/mx-chain-go/testscommon/chainSimulator"

	"github.com/stretchr/testify/require"
)

func TestNewSovereignProcessorFactory(t *testing.T) {
	t.Parallel()

	fact := NewChainHandlerFactory()

	require.False(t, fact.IsInterfaceNil())
	require.IsType(t, new(processorFactory), fact)
}

func TestNewSovereignProcessorFactory_CreateChainHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil node handler, should error", func(t *testing.T) {
		fact := NewChainHandlerFactory()

		chainHandler, err := fact.CreateChainHandler(nil, heartbeat.NewHeartbeatMonitor())
		require.Nil(t, chainHandler)
		require.ErrorIs(t, err, process.ErrNilNodeHandler)
	})

	t.Run("nil heart beat monitor, should error", func(t *testing.T) {
		fact := NewChainHandlerFactory()

		chainHandler, err := fact.CreateChainHandler(&chainSimulator.NodeHandlerMock{}, nil)
		require.Nil(t, chainHandler)
		require.ErrorIs(t, err, heartbeatMonitor.ErrNilHeartbeatMonitor)
	})

	t.Run("should work", func(t *testing.T) {
		fact := NewChainHandlerFactory()

		chainHandler, err := fact.CreateChainHandler(&chainSimulator.NodeHandlerMock{}, heartbeat.NewHeartbeatMonitor())
		require.Nil(t, err)
		require.NotNil(t, chainHandler)
	})
}
