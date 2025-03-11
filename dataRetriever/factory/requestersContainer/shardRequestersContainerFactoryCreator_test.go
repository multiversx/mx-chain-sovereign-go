package requesterscontainer_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-sovereign-go/dataRetriever/factory/requestersContainer"
)

func TestNewShardRequestersContainerFactoryCreator(t *testing.T) {
	t.Parallel()

	factory := requesterscontainer.NewShardRequestersContainerFactoryCreator()
	require.False(t, factory.IsInterfaceNil())
	require.Implements(t, new(requesterscontainer.RequesterContainerFactoryCreator), factory)
}

func TestShardRequestersContainerFactoryCreator_CreateRequesterContainerFactory(t *testing.T) {
	t.Parallel()

	factory := requesterscontainer.NewShardRequestersContainerFactoryCreator()

	args := getArguments()
	container, err := factory.CreateRequesterContainerFactory(args)
	require.Nil(t, err)
	require.IsType(t, container, requesterscontainer.ShardRequestersContainerFactory)
}
