package sharding

import (
	"testing"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/sharding/mock"
	"github.com/multiversx/mx-chain-go/testscommon/chainParameters"
	"github.com/stretchr/testify/require"
)

func TestGenesisNodesSetupFactory_CreateNodesSetup(t *testing.T) {
	t.Parallel()

	factory := NewGenesisNodesSetupFactory()
	require.False(t, factory.IsInterfaceNil())

	nodesHandler, err := factory.CreateNodesSetup(&NodesSetupArgs{
		GenesisMaxNumShards: 1,
		NodesConfig: config.NodesConfig{
			InitialNodes: createInitialNodes(),
		},
		AddressPubKeyConverter:   mock.NewPubkeyConverterMock(32),
		ValidatorPubKeyConverter: mock.NewPubkeyConverterMock(96),
		ChainParametersProvider:  &chainParameters.ChainParametersHolderMock{},
	})
	require.Nil(t, err)
	require.NotNil(t, nodesHandler)
	require.IsType(t, &NodesSetup{}, nodesHandler)
}
