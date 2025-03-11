package api

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-sovereign-go/config"
	"github.com/multiversx/mx-chain-sovereign-go/epochStart/metachain"
	factoryDisabled "github.com/multiversx/mx-chain-sovereign-go/factory/disabled"
	"github.com/multiversx/mx-chain-sovereign-go/process/mock"
	"github.com/multiversx/mx-chain-sovereign-go/testscommon/enableEpochsHandlerMock"
	epochNotifierMock "github.com/multiversx/mx-chain-sovereign-go/testscommon/epochNotifier"
)

func createArgs(shardID uint32) ArgsCreateAPIProcessComps {
	return ArgsCreateAPIProcessComps{
		ArgsStakingDataProvider: metachain.StakingDataProviderArgs{
			EnableEpochsHandler: enableEpochsHandlerMock.NewEnableEpochsHandlerStub(),
			SystemVM:            &mock.VMExecutionHandlerStub{},
			MinNodePrice:        "100",
		},
		ShardCoordinator: mock.NewMultipleShardsCoordinatorMockWithSelfShard(shardID),
		EpochNotifier:    &epochNotifierMock.EpochNotifierStub{},
		SoftAuctionConfig: config.SoftAuctionConfig{
			TopUpStep:             "10",
			MinTopUp:              "10",
			MaxTopUp:              "100",
			MaxNumberOfIterations: 10,
		},
		EnableEpochs: config.EnableEpochs{
			MaxNodesChangeEnableEpoch: []config.MaxNodesChangeConfig{
				{
					EpochEnable:            0,
					MaxNumNodes:            8,
					NodesToShufflePerShard: 2,
				},
			},
		},
		Denomination: 1,
	}
}

func TestApiProcessorCompsCreator_CreateAPIComps(t *testing.T) {
	factory := NewAPIProcessorCompsCreator()
	require.False(t, factory.IsInterfaceNil())

	disabledComps := &APIProcessComps{
		StakingDataProviderAPI: factoryDisabled.NewDisabledStakingDataProvider(),
		AuctionListSelector:    factoryDisabled.NewDisabledAuctionListSelector(),
	}

	// shard components, should be disabled
	args := createArgs(2)
	apiComps, err := factory.CreateAPIComps(args)
	require.Nil(t, err)
	require.Equal(t, apiComps, disabledComps)

	// meta components
	args = createArgs(core.MetachainShardId)
	apiComps, err = factory.CreateAPIComps(args)
	require.Nil(t, err)
	require.NotEqual(t, apiComps, disabledComps)
}
