package sharding

import (
	"testing"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/sharding/mock"
	"github.com/multiversx/mx-chain-go/testscommon/chainParameters"
	"github.com/stretchr/testify/require"
)

func TestSovereignGenesisNodesSetupFactory_CreateNodesSetup(t *testing.T) {
	t.Parallel()

	factory := NewSovereignGenesisNodesSetupFactory()
	require.False(t, factory.IsInterfaceNil())

	nodesHandler, err := factory.CreateNodesSetup(&NodesSetupArgs{
		GenesisMaxNumShards: 100,
		NodesConfig: config.NodesConfig{
			InitialNodes: createInitialNodes(),
		},
		AddressPubKeyConverter:   mock.NewPubkeyConverterMock(32),
		ValidatorPubKeyConverter: mock.NewPubkeyConverterMock(96),
		ChainParametersProvider: &chainParameters.ChainParametersHandlerStub{
			ChainParametersForEpochCalled: func(epoch uint32) (config.ChainParametersByEpochConfig, error) {
				return config.ChainParametersByEpochConfig{
					ShardConsensusGroupSize:     1,
					ShardMinNumNodes:            1,
					MetachainConsensusGroupSize: 0,
					MetachainMinNumNodes:        0,
				}, nil
			},
		},
	})
	require.Nil(t, err)
	require.NotNil(t, nodesHandler)
	require.IsType(t, &SovereignNodesSetup{}, nodesHandler)
}
