package nodesCoordinator

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/hashing/factory"
	"github.com/stretchr/testify/require"
)

func TestNewSovereignIndexHashedNodesCoordinator(t *testing.T) {
	t.Parallel()

	t.Run("invalid consensus size, should return error", func(t *testing.T) {
		args := createSovereignArgs()
		args.ShardConsensusGroupSize = 0

		ihnc, err := NewSovereignIndexHashedNodesCoordinator(args)
		require.Nil(t, ihnc)
		require.Equal(t, ErrInvalidConsensusGroupSize, err)
	})

	t.Run("invalid number of shards, should return error", func(t *testing.T) {
		args := createSovereignArgs()
		args.NbShards = 0

		ihnc, err := NewSovereignIndexHashedNodesCoordinator(args)
		require.Nil(t, ihnc)
		require.Equal(t, ErrInvalidNumberOfShards, err)
	})

	t.Run("small eligible list, should return error", func(t *testing.T) {
		args := createSovereignArgs()
		args.ShardConsensusGroupSize = 9999

		ihnc, err := NewSovereignIndexHashedNodesCoordinator(args)
		require.Nil(t, ihnc)
		require.Equal(t, ErrSmallShardEligibleListSize, err)
	})

	t.Run("should work", func(t *testing.T) {
		args := createSovereignArgs()
		ihnc, err := NewSovereignIndexHashedNodesCoordinator(args)
		require.Nil(t, err)
		require.False(t, ihnc.IsInterfaceNil())

		bls, _ := hex.DecodeString("00634d502a2c6fd7a68f436b17b791caf9654654cf814bef672a8a6d2f9f9e2a56526ce2b6bf25a741025c5af280ae0b799be93492d46932102d27c781680b969bde66d3d47df593d1b7bdb65fc0b8d09e3374bccaee5042fc942af78a748294")

		hasher, _ := factory.NewHasher("blake2b")
		log.Info("dsa", "Dsa", hasher.Compute(string(bls)))
	})

}

func TestSovereignIndexHashedNodesCoordinator_ComputeValidatorsGroup(t *testing.T) {
	t.Parallel()

	list := []Validator{
		newValidatorMock([]byte("pk0"), 1, defaultSelectionChances),
	}
	nodesMap := map[uint32][]Validator{
		core.SovereignChainShardId: list,
	}
	arguments := createArguments()
	arguments.EligibleNodes = nodesMap
	ihnc, _ := NewSovereignIndexHashedNodesCoordinator(arguments)

	t.Run("nil randomness, should return error", func(t *testing.T) {
		t.Parallel()

		list2, err := ihnc.ComputeConsensusGroup(nil, 0, core.SovereignChainShardId, 0)
		require.Empty(t, list2)
		require.Equal(t, ErrNilRandomness, err)
	})

	t.Run("invalid shard id, should return error", func(t *testing.T) {
		t.Parallel()

		list2, err := ihnc.ComputeConsensusGroup([]byte("randomness"), 0, core.MetachainShardId, 0)
		require.Empty(t, list2)
		require.Equal(t, ErrInvalidShardId, err)
	})

	t.Run("config not found for requested epoch, should return error", func(t *testing.T) {
		t.Parallel()

		list2, err := ihnc.ComputeConsensusGroup([]byte("randomness"), 0, core.SovereignChainShardId, 99999)
		require.Empty(t, list2)
		require.True(t, strings.Contains(err.Error(), ErrEpochNodesConfigDoesNotExist.Error()))
		require.True(t, strings.Contains(err.Error(), "99999"))
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		list2, err := ihnc.ComputeConsensusGroup([]byte("randomness"), 0, core.SovereignChainShardId, 0)
		require.Equal(t, list, list2)
		require.Nil(t, err)
	})
}

func TestSovereignIndexHashedNodesCoordinator_GetConsensusValidatorsPublicKeys(t *testing.T) {
	t.Parallel()

	list := []Validator{
		newValidatorMock([]byte("pk0"), 1, defaultSelectionChances),
	}
	nodesMap := map[uint32][]Validator{
		core.SovereignChainShardId: list,
	}
	arguments := createArguments()
	arguments.EligibleNodes = nodesMap
	ihnc, _ := NewSovereignIndexHashedNodesCoordinator(arguments)

	t.Run("nil randomness, cannot compute consensus group, should return error", func(t *testing.T) {
		t.Parallel()

		list2, err := ihnc.GetConsensusValidatorsPublicKeys(nil, 0, core.SovereignChainShardId, 0)
		require.Empty(t, list2)
		require.Equal(t, ErrNilRandomness, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		pubKeys, err := ihnc.GetConsensusValidatorsPublicKeys([]byte("randomness"), 0, core.SovereignChainShardId, 0)
		require.Equal(t, []string{string(list[0].PubKey())}, pubKeys)
		require.Nil(t, err)
	})
}
