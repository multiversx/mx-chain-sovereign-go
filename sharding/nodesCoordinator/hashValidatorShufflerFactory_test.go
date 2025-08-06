package nodesCoordinator

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/sharding/mock"
)

func TestHashValidatorShufflerFactory_CreateHashValidatorShuffler(t *testing.T) {
	t.Parallel()

	factory := NewHashValidatorShufflerFactory()
	require.False(t, factory.IsInterfaceNil())

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
	require.IsType(t, &randHashShuffler{}, shuffler)
}
