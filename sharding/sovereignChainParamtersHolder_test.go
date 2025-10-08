package sharding

import (
	"testing"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/testscommon/commonmocks"
	mock "github.com/multiversx/mx-chain-go/testscommon/epochstartmock"
	"github.com/stretchr/testify/require"
)

func createSovereignArgs() ArgsChainParametersHolder {
	return ArgsChainParametersHolder{
		EpochStartEventNotifier: &mock.EpochStartNotifierStub{},
		ChainParameters: []config.ChainParametersByEpochConfig{
			{
				EnableEpoch:                 0,
				ShardMinNumNodes:            5,
				ShardConsensusGroupSize:     5,
				MetachainMinNumNodes:        0,
				MetachainConsensusGroupSize: 0,
				RoundDuration:               4000,
				Hysteresis:                  0.2,
				Adaptivity:                  false,
			},
		},
		ChainParametersNotifier: &commonmocks.ChainParametersNotifierStub{},
	}
}

func TestNewSovereignChainParametersHolder(t *testing.T) {
	t.Parallel()

	t.Run("invalid shard consensus size", func(t *testing.T) {
		args := createSovereignArgs()
		args.ChainParameters[0].ShardConsensusGroupSize = 0
		scp, err := NewSovereignChainParametersHolder(args)
		require.Nil(t, scp)
		require.ErrorIs(t, err, ErrNegativeOrZeroConsensusGroupSize)
	})
	t.Run("invalid shard min num nodes", func(t *testing.T) {
		args := createSovereignArgs()
		args.ChainParameters[0].ShardConsensusGroupSize = 2
		args.ChainParameters[0].ShardMinNumNodes = 1
		scp, err := NewSovereignChainParametersHolder(args)
		require.Nil(t, scp)
		require.ErrorIs(t, err, ErrMinNodesPerShardSmallerThanConsensusSize)
	})
	t.Run("invalid meta consensus size", func(t *testing.T) {
		args := createSovereignArgs()
		args.ChainParameters[0].MetachainConsensusGroupSize = 1
		scp, err := NewSovereignChainParametersHolder(args)
		require.Nil(t, scp)
		require.ErrorIs(t, err, errSovereignInvalidMetaConsensusSize)
	})
	t.Run("invalid meta min num nodes", func(t *testing.T) {
		args := createSovereignArgs()
		args.ChainParameters[0].MetachainMinNumNodes = 1
		scp, err := NewSovereignChainParametersHolder(args)
		require.Nil(t, scp)
		require.ErrorIs(t, err, errSovereignInvalidMetaNumNodes)
	})
	t.Run("nil input", func(t *testing.T) {
		args := createSovereignArgs()
		args.ChainParametersNotifier = nil
		scp, err := NewSovereignChainParametersHolder(args)
		require.Nil(t, scp)
		require.ErrorIs(t, err, ErrNilChainParametersNotifier)
	})
	t.Run("should work", func(t *testing.T) {
		args := createSovereignArgs()
		scp, err := NewSovereignChainParametersHolder(args)
		require.Nil(t, err)
		require.NotNil(t, scp)
		require.False(t, scp.IsInterfaceNil())
	})
}
