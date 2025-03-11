package node_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-sovereign-go/config"
	"github.com/multiversx/mx-chain-sovereign-go/errors"
	"github.com/multiversx/mx-chain-sovereign-go/factory/mock"
	"github.com/multiversx/mx-chain-sovereign-go/node"
	componentsMock "github.com/multiversx/mx-chain-sovereign-go/testscommon/components"
	"github.com/multiversx/mx-chain-sovereign-go/testscommon/consensus/factoryMocks"
	"github.com/multiversx/mx-chain-sovereign-go/testscommon/factory"
	"github.com/multiversx/mx-chain-sovereign-go/testscommon/mainFactoryMocks"
)

func TestCreateNode(t *testing.T) {
	t.Parallel()

	t.Run("nil node factory should not work", func(t *testing.T) {
		t.Parallel()

		nodeHandler, err := node.CreateNode(
			&config.Config{},
			componentsMock.GetRunTypeComponents(),
			&factory.StatusCoreComponentsStub{},
			getDefaultBootstrapComponents(),
			getDefaultCoreComponents(),
			getDefaultCryptoComponents(),
			getDefaultDataComponents(),
			getDefaultNetworkComponents(),
			getDefaultProcessComponents(),
			getDefaultStateComponents(),
			&mainFactoryMocks.StatusComponentsStub{},
			&mock.HeartbeatV2ComponentsStub{},
			&factoryMocks.ConsensusComponentsStub{
				GroupSize: 1,
			},
			0,
			false,
			nil)

		require.NotNil(t, err)
		require.Equal(t, errors.ErrNilNode, err)
		require.Nil(t, nodeHandler)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		nodeHandler, err := node.CreateNode(
			&config.Config{},
			componentsMock.GetRunTypeComponents(),
			&factory.StatusCoreComponentsStub{},
			getDefaultBootstrapComponents(),
			getDefaultCoreComponents(),
			getDefaultCryptoComponents(),
			getDefaultDataComponents(),
			getDefaultNetworkComponents(),
			getDefaultProcessComponents(),
			getDefaultStateComponents(),
			&mainFactoryMocks.StatusComponentsStub{},
			&mock.HeartbeatV2ComponentsStub{},
			&factoryMocks.ConsensusComponentsStub{
				GroupSize: 1,
			},
			0,
			false,
			node.NewSovereignNodeFactory(nativeESDT))

		require.Nil(t, err)
		require.NotNil(t, nodeHandler)
	})
}
