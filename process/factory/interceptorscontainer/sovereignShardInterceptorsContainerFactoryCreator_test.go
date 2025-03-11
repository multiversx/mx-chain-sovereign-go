package interceptorscontainer_test

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-sovereign-go/process/factory/interceptorscontainer"
	"github.com/multiversx/mx-chain-sovereign-go/testscommon/sovereign"
)

func TestNewSovereignShardInterceptorsContainerFactoryCreator(t *testing.T) {
	t.Parallel()

	factoryCreator := interceptorscontainer.NewSovereignShardInterceptorsContainerFactoryCreator()
	require.Implements(t, new(interceptorscontainer.InterceptorsContainerFactoryCreator), factoryCreator)
	require.False(t, check.IfNil(factoryCreator))
}

func TestSovereignShardInterceptorsContainerFactoryCreator_CreateInterceptorsContainerFactory(t *testing.T) {
	t.Parallel()

	factoryCreator := interceptorscontainer.NewSovereignShardInterceptorsContainerFactoryCreator()
	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsShard(coreComp, cryptoComp)
	args.IncomingHeaderSubscriber = &sovereign.IncomingHeaderSubscriberStub{}
	factory, err := factoryCreator.CreateInterceptorsContainerFactory(args)
	require.Nil(t, err)
	require.Equal(t, "*interceptorscontainer.sovereignShardInterceptorsContainerFactory", fmt.Sprintf("%T", factory))
}
