package api

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"

	factoryDisabled "github.com/multiversx/mx-chain-sovereign-go/factory/disabled"
)

func TestSovereignApiProcessorCompsCreator_CreateAPIComps(t *testing.T) {
	factory := NewSovereignAPIProcessorCompsCreator()
	require.False(t, factory.IsInterfaceNil())

	disabledComps := &APIProcessComps{
		StakingDataProviderAPI: factoryDisabled.NewDisabledStakingDataProvider(),
		AuctionListSelector:    factoryDisabled.NewDisabledAuctionListSelector(),
	}

	args := createArgs(core.SovereignChainShardId)
	apiComps, err := factory.CreateAPIComps(args)
	require.Nil(t, err)
	require.NotEqual(t, apiComps, disabledComps)
}
