package runTypeChain_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/factory/runTypeChain"
)

func createChainComponents() (factory.RunTypeChainComponentsHandler, error) {
	rtccf := runTypeChain.NewRunTypeChainComponentsFactory()
	return runTypeChain.NewManagedRunTypeChainComponents(rtccf)
}

func TestNewManagedRunTypeChainComponents(t *testing.T) {
	t.Parallel()

	t.Run("should error", func(t *testing.T) {
		managedRunTypeChainComponents, err := runTypeChain.NewManagedRunTypeChainComponents(nil)
		require.ErrorIs(t, err, errors.ErrNilRunTypeChainComponentsFactory)
		require.True(t, managedRunTypeChainComponents.IsInterfaceNil())
	})
	t.Run("should work", func(t *testing.T) {
		managedRunTypeChainComponents, err := createChainComponents()
		require.NoError(t, err)
		require.False(t, managedRunTypeChainComponents.IsInterfaceNil())
	})
}

func TestManagedRunTypeChainComponents_Create(t *testing.T) {
	t.Parallel()

	managedRunTypeChainComponents, err := createChainComponents()
	require.NoError(t, err)

	// require.Nil(t, managedRunTypeChainComponents.SomeFactoryCreator())

	err = managedRunTypeChainComponents.Create()
	require.NoError(t, err)

	// require.NotNil(t, managedRunTypeChainComponents.SomeFactoryCreator())

	require.Equal(t, factory.RunTypeChainComponentsName, managedRunTypeChainComponents.String())
	require.NoError(t, managedRunTypeChainComponents.Close())
}

func TestManagedRunTypeChainComponents_Close(t *testing.T) {
	t.Parallel()

	managedRunTypeChainComponents, _ := createChainComponents()
	require.NoError(t, managedRunTypeChainComponents.Close())

	err := managedRunTypeChainComponents.Create()
	require.NoError(t, err)

	require.NoError(t, managedRunTypeChainComponents.Close())
	// require.Nil(t, managedRunTypeChainComponents.SomeFactoryCreator())
}

func TestManagedRunTypeChainComponents_CheckSubcomponents(t *testing.T) {
	t.Parallel()

	managedRunTypeChainComponents, _ := createChainComponents()
	err := managedRunTypeChainComponents.CheckSubcomponents()
	require.Equal(t, errors.ErrNilRunTypeChainComponents, err)

	err = managedRunTypeChainComponents.Create()
	require.NoError(t, err)

	//TODO check for nil each subcomponent - MX-15371
	err = managedRunTypeChainComponents.CheckSubcomponents()
	require.NoError(t, err)

	require.NoError(t, managedRunTypeChainComponents.Close())
}
