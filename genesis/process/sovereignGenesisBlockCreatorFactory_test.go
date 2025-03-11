//go:build !race

package process

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-sovereign-go/genesis/mock"
)

func TestNewSovereignGenesisBlockCreatorFactory(t *testing.T) {
	factory := NewSovereignGenesisBlockCreatorFactory()
	require.False(t, factory.IsInterfaceNil())
	require.Implements(t, new(GenesisBlockCreatorFactory), factory)
}

func TestSovereignGenesisBlockCreatorFactory_CreateGenesisBlockCreator(t *testing.T) {
	factory := NewSovereignGenesisBlockCreatorFactory()

	args := createMockArgument(t, "testdata/genesisTest1.json", &mock.InitialNodesHandlerStub{}, big.NewInt(22000))
	blockCreator, err := factory.CreateGenesisBlockCreator(args)
	require.Nil(t, err)
	require.IsType(t, &sovereignGenesisBlockCreator{}, blockCreator)
}
