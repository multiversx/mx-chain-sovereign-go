package nodesCoordinator

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/sharding/mock"
)

func TestSovereignHashValidatorShufflerFactory_CreateHashValidatorShuffler(t *testing.T) {
	t.Parallel()

	factory := NewSovereignHashValidatorShufflerFactory()
	require.False(t, factory.IsInterfaceNil())

	t.Run("nil args, should error", func(t *testing.T) {
		shuffler, err := factory.CreateHashValidatorShuffler(nil)
		require.Nil(t, shuffler)
		require.ErrorIs(t, err, ErrNilNodeShufflerArguments)
	})
	t.Run("should work", func(t *testing.T) {
		shufflerArgs := &NodesShufflerArgs{
			ShuffleBetweenShards: true,
			EnableEpochsHandler:  &mock.EnableEpochsHandlerMock{},
			EnableEpochs: config.EnableEpochs{
				StakingV4Step2EnableEpoch: 443,
				StakingV4Step3EnableEpoch: 444,
			},
		}

		shuffler, err := factory.CreateHashValidatorShuffler(shufflerArgs)
		require.NoError(t, err)
		require.NotNil(t, shuffler)
		require.IsType(t, &sovereignHashValidatorShuffler{}, shuffler)
	})
}
