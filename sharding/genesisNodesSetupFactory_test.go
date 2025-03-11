package sharding

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-sovereign-go/sharding/mock"
)

func TestGenesisNodesSetupFactory_CreateNodesSetup(t *testing.T) {
	t.Parallel()

	factory := NewGenesisNodesSetupFactory()
	require.False(t, factory.IsInterfaceNil())

	nodesHandler, err := factory.CreateNodesSetup(&NodesSetupArgs{
		NodesFilePath:            "mock/testdata/nodesSetupMock.json",
		AddressPubKeyConverter:   mock.NewPubkeyConverterMock(32),
		ValidatorPubKeyConverter: mock.NewPubkeyConverterMock(96),
		GenesisMaxNumShards:      100,
	})
	require.Nil(t, err)
	require.NotNil(t, nodesHandler)
	require.IsType(t, &NodesSetup{}, nodesHandler)
}
