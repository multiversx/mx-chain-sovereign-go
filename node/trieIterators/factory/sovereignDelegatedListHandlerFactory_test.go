package factory_test

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"

	trieIteratorsFactory "github.com/multiversx/mx-chain-sovereign-go/node/trieIterators/factory"
)

func TestNewSovereignDelegatedListProcessorFactory(t *testing.T) {
	t.Parallel()

	sovereignDelegatedListHandlerFactory := trieIteratorsFactory.NewSovereignDelegatedListProcessorFactory()
	require.False(t, sovereignDelegatedListHandlerFactory.IsInterfaceNil())
}

func TestSovereignDelegatedListProcessorFactory_CreateDelegatedListProcessorHandler(t *testing.T) {
	t.Parallel()

	args := createMockArgs(core.SovereignChainShardId)

	sovereignDelegatedListHandler, err := trieIteratorsFactory.NewSovereignDelegatedListProcessorFactory().CreateDelegatedListProcessorHandler(args)
	require.Nil(t, err)
	require.Equal(t, "*trieIterators.delegatedListProcessor", fmt.Sprintf("%T", sovereignDelegatedListHandler))
}
