package sharding

import (
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/sharding/mock"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/testscommon/chainParameters"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createSovereignMockNodesSetup(args argsTestNodesSetup) SovereignNodesSetup {
	nodesSetup := &NodesSetup{
		genesisChainParameters: config.ChainParametersByEpochConfig{
			EnableEpoch:                 0,
			ShardMinNumNodes:            args.shardMinNodes,
			ShardConsensusGroupSize:     args.shardConsensusSize,
			MetachainMinNumNodes:        args.metaMinNodes,
			MetachainConsensusGroupSize: args.metaConsensusSize,
		},
		addressPubkeyConverter:   mock.NewPubkeyConverterMock(32),
		validatorPubkeyConverter: mock.NewPubkeyConverterMock(96),
	}

	initNodesSetup(nodesSetup, config.NodesConfig{InitialNodes: createInitialNodes()})

	return SovereignNodesSetup{
		NodesSetup: nodesSetup,
	}
}

func createInitialNodes() []*config.InitialNodeConfig {
	return []*config.InitialNodeConfig{
		{
			PubKey:  pubKeys[0],
			Address: address[0],
		},
		{
			PubKey:  pubKeys[1],
			Address: address[1],
		},
	}
}

func createSovereignNodesSetupArgs() *SovereignNodesSetupArgs {
	return &SovereignNodesSetupArgs{
		NodesConfig: config.NodesConfig{
			InitialNodes: createInitialNodes(),
		},
		AddressPubKeyConverter:   mock.NewPubkeyConverterMock(32),
		ValidatorPubKeyConverter: mock.NewPubkeyConverterMock(32),
		ChainParametersProvider:  &chainParameters.ChainParametersHolderMock{},
	}
}

func TestNewSovereignNodesSetupErrorCases(t *testing.T) {
	t.Parallel()

	t.Run("nil address converter", func(t *testing.T) {
		t.Parallel()

		args := createSovereignNodesSetupArgs()
		args.AddressPubKeyConverter = nil

		ns, err := NewSovereignNodesSetup(args)
		require.Nil(t, ns)
		require.ErrorIs(t, err, ErrNilPubkeyConverter)
		require.True(t, strings.Contains(err.Error(), "addressPubkeyConverter"))
	})

	t.Run("nil validator converter", func(t *testing.T) {
		t.Parallel()

		args := createSovereignNodesSetupArgs()
		args.ValidatorPubKeyConverter = nil

		ns, err := NewSovereignNodesSetup(args)
		require.Nil(t, ns)
		require.ErrorIs(t, err, ErrNilPubkeyConverter)
		require.True(t, strings.Contains(err.Error(), "validatorPubkeyConverter"))
	})

	t.Run("nil chain parameters provider", func(t *testing.T) {
		t.Parallel()

		args := createSovereignNodesSetupArgs()
		args.ChainParametersProvider = nil

		ns, err := NewSovereignNodesSetup(args)
		require.Nil(t, ns)
		require.ErrorIs(t, err, ErrNilChainParametersProvider)
	})
}

func TestNewSovereignNodesSetupShouldWork(t *testing.T) {
	t.Parallel()

	addrPubKeyConverter := mock.NewPubkeyConverterMock(32)
	validatorPubKeyConverter := mock.NewPubkeyConverterMock(96)
	ns, err := NewSovereignNodesSetup(
		&SovereignNodesSetupArgs{
			NodesConfig: config.NodesConfig{
				StartTime:    1689935785,
				InitialNodes: createInitialNodes(),
			},
			AddressPubKeyConverter:   addrPubKeyConverter,
			ValidatorPubKeyConverter: validatorPubKeyConverter,
			ChainParametersProvider: &chainParameters.ChainParametersHandlerStub{
				ChainParametersForEpochCalled: func(epoch uint32) (config.ChainParametersByEpochConfig, error) {
					return config.ChainParametersByEpochConfig{
						RoundDuration:               5000,
						ShardConsensusGroupSize:     1,
						ShardMinNumNodes:            1,
						MetachainConsensusGroupSize: 0,
						MetachainMinNumNodes:        0,
					}, nil
				},
			},
		},
	)
	require.Nil(t, err)
	require.NotNil(t, ns)

	encodedPubKeys := []string{
		"41378f754e2c7b2745208c3ed21b151d297acdc84c3aca00b9e292cf28ec2d444771070157ea7760ed83c26f4fed387d0077e00b563a95825dac2cbc349fc0025ccf774e37b0a98ad9724d30e90f8c29b4091ccb738ed9ffc0573df776ee9ea30b3c038b55e532760ea4a8f152f2a52848020e5cee1cc537f2c2323399723081",
		"52f3bf5c01771f601ec2137e267319ab6716ef6ff5dfddaea48b42d955f631167f2ce19296a202bb8fd174f4e94f8c85f619df85a7f9f8de0f3768e5e6d8c48187b767deccf9829be246aa331aa86d182eb8fa28ea8a3e45d357ed1647a9be020a5569d686253a6f89e9123c7f21f302e82f67d3e3cd69cf267b9910a663ef32",
	}
	encodedAddresses := []string{
		"9e95a4e46da335a96845b4316251fc1bb197e1b8136d96ecc62bf6604eca9e49",
		"7a330039e77ca06bc127319fd707cc4911a80db489a39fcfb746283a05f61836",
	}

	decodedPubKey1, _ := addrPubKeyConverter.Decode(encodedPubKeys[0])
	decodedPubKey2, _ := addrPubKeyConverter.Decode(encodedPubKeys[1])

	decodedAddr1, _ := addrPubKeyConverter.Decode(encodedAddresses[0])
	decodedAddr2, _ := addrPubKeyConverter.Decode(encodedAddresses[1])

	nodesInfo := []nodeInfo{
		{
			assignedShard: core.SovereignChainShardId,
			eligible:      true,
			pubKey:        decodedPubKey1,
			address:       decodedAddr1,
			initialRating: defaultInitialRating,
		},
		{
			assignedShard: core.SovereignChainShardId,
			eligible:      true,
			pubKey:        decodedPubKey2,
			address:       decodedAddr2,
			initialRating: defaultInitialRating,
		},
	}

	expectedInitialNodes := []nodesCoordinator.GenesisNodeInfoHandler{
		&InitialNode{
			PubKey:        encodedPubKeys[0],
			Address:       encodedAddresses[0],
			InitialRating: 0,
			nodeInfo:      nodesInfo[0],
		},
		&InitialNode{
			PubKey:        encodedPubKeys[1],
			Address:       encodedAddresses[1],
			InitialRating: 0,
			nodeInfo:      nodesInfo[1],
		},
	}

	require.Equal(t, expectedInitialNodes, ns.AllInitialNodes())

	eligible, waiting := ns.InitialNodesInfo()
	require.Empty(t, waiting)
	require.Equal(t, []nodesCoordinator.GenesisNodeInfoHandler{&nodesInfo[0], &nodesInfo[1]}, eligible[core.SovereignChainShardId])

	shardID, err := ns.GetShardIDForPubKey(decodedPubKey1)
	require.Nil(t, err)
	require.Equal(t, core.SovereignChainShardId, shardID)

	shardID, err = ns.GetShardIDForPubKey(decodedPubKey2)
	require.Nil(t, err)
	require.Equal(t, core.SovereignChainShardId, shardID)

	require.Equal(t, int64(1689935785), ns.GetStartTime())
	require.Equal(t, uint64(5000), ns.GetRoundDuration())
	require.Equal(t, uint32(1), ns.GetShardConsensusGroupSize())
	require.Equal(t, uint32(0), ns.GetMetaConsensusGroupSize())
	require.Equal(t, uint32(1), ns.NumberOfShards())
	require.Equal(t, uint32(1), ns.MinNumberOfNodes())
	require.Equal(t, uint32(1), ns.MinNumberOfShardNodes())
	require.Equal(t, uint32(0), ns.MinNumberOfMetaNodes())
	require.Equal(t, float32(0), ns.GetHysteresis())
	require.Equal(t, uint32(1), ns.MinNumberOfNodesWithHysteresis())
	require.Equal(t, false, ns.GetAdaptivity())
}

func TestProcessSovereignConfigErrorCases(t *testing.T) {
	t.Parallel()

	t.Run("invalid consensus size", func(t *testing.T) {
		t.Parallel()

		ns := createSovereignMockNodesSetup(argsTestNodesSetup{
			shardConsensusSize: 0,
			shardMinNodes:      0,
			metaConsensusSize:  0,
			metaMinNodes:       0,
			numInitialNodes:    2,
			genesisMaxShards:   1,
		})
		err := ns.processSovereignConfig()
		assert.Equal(t, ErrNegativeOrZeroConsensusGroupSize, err)
	})

	t.Run("invalid min nodes vs consensus size", func(t *testing.T) {
		t.Parallel()

		ns := createSovereignMockNodesSetup(argsTestNodesSetup{
			shardConsensusSize: 4,
			shardMinNodes:      3,
			metaConsensusSize:  0,
			metaMinNodes:       0,
			numInitialNodes:    2,
			genesisMaxShards:   1,
		})

		err := ns.processSovereignConfig()
		assert.Equal(t, ErrMinNodesPerShardSmallerThanConsensusSize, err)
	})

	t.Run("invalid min nodes vs num of actual nodes", func(t *testing.T) {
		t.Parallel()

		ns := createSovereignMockNodesSetup(argsTestNodesSetup{
			shardConsensusSize: 4,
			shardMinNodes:      999,
			metaConsensusSize:  0,
			metaMinNodes:       0,
			numInitialNodes:    2,
			genesisMaxShards:   1,
		})

		err := ns.processSovereignConfig()
		assert.Equal(t, ErrNodesSizeSmallerThanMinNoOfNodes, err)
	})

	t.Run("invalid meta chain num nodes", func(t *testing.T) {
		t.Parallel()

		ns := createSovereignMockNodesSetup(argsTestNodesSetup{
			shardConsensusSize: 2,
			shardMinNodes:      2,
			metaConsensusSize:  1,
			metaMinNodes:       0,
			numInitialNodes:    2,
			genesisMaxShards:   1,
		})
		err := ns.processSovereignConfig()
		require.ErrorIs(t, err, errSovereignInvalidMetaConsensusSize)

		ns = createSovereignMockNodesSetup(argsTestNodesSetup{
			shardConsensusSize: 2,
			shardMinNodes:      2,
			metaConsensusSize:  0,
			metaMinNodes:       1,
			numInitialNodes:    2,
			genesisMaxShards:   1,
		})
		err = ns.processSovereignConfig()
		require.ErrorIs(t, err, errSovereignInvalidMetaConsensusSize)

		ns = createSovereignMockNodesSetup(argsTestNodesSetup{
			shardConsensusSize: 2,
			shardMinNodes:      2,
			metaConsensusSize:  1,
			metaMinNodes:       1,
			numInitialNodes:    2,
			genesisMaxShards:   1,
		})
		err = ns.processSovereignConfig()
		require.ErrorIs(t, err, errSovereignInvalidMetaConsensusSize)
	})

}
