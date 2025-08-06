package nodesCoordinator

import (
	"sort"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/sharding/mock"
)

func TestNewSovereignHashValidatorShuffler(t *testing.T) {
	t.Parallel()

	t.Run("nil shuffler, should error", func(t *testing.T) {
		sovShuffler, err := NewSovereignHashValidatorsShuffler(nil)
		require.Nil(t, sovShuffler)
		require.ErrorIs(t, err, ErrNilShuffler)
	})
	t.Run("should work", func(t *testing.T) {
		shuffler, err := createHashShufflerInter()
		require.NoError(t, err)
		require.False(t, shuffler.IsInterfaceNil())

		sovShuffler, err := NewSovereignHashValidatorsShuffler(shuffler)
		require.NotNil(t, sovShuffler)
		require.NoError(t, err)
	})
}

func generateSovValidatorMap(numOfNodes int) map[uint32][]Validator {
	validatorsMap := make(map[uint32][]Validator)
	validatorsMap[core.SovereignChainShardId] = generateValidatorList(numOfNodes)

	return validatorsMap
}

func TestSovereignHashValidatorsShuffler_Shuffle(t *testing.T) {
	t.Parallel()

	eligiblePerShard := 4
	waitingPerShard := 2

	eligibleMap := generateSovValidatorMap(eligiblePerShard)
	waitingMap := generateSovValidatorMap(waitingPerShard)

	unstakeLeaving := make(map[uint32][]Validator)
	additionalLeaving := make(map[uint32][]Validator)

	// no duplicates on shard 0
	firstRemovedShard0 := eligibleMap[0][3]
	secondRemovedShard0 := waitingMap[0][0]
	unstakeLeaving[0] = []Validator{
		firstRemovedShard0,
	}
	additionalLeaving[0] = []Validator{
		secondRemovedShard0,
	}

	unstakeLeavingList, additionalLeavingList := prepareListsFromMaps(unstakeLeaving, additionalLeaving)

	shufflerArgs := &NodesShufflerArgs{
		ShuffleBetweenShards: shuffleBetweenShards,
		EnableEpochs: config.EnableEpochs{
			StakingV4Step2EnableEpoch: 443,
			StakingV4Step3EnableEpoch: 444,
		},
		EnableEpochsHandler: &mock.EnableEpochsHandlerMock{},
	}
	shf, err := NewHashValidatorsShuffler(shufflerArgs)
	require.Nil(t, err)
	shuffler, _ := NewSovereignHashValidatorsShuffler(shf)
	require.NotNil(t, shuffler)

	arg := ArgsUpdateNodes{
		Eligible:          eligibleMap,
		Waiting:           waitingMap,
		NewNodes:          make([]Validator, 0),
		UnStakeLeaving:    unstakeLeavingList,
		AdditionalLeaving: additionalLeavingList,
		Rand:              generateRandomByteArray(32),
		NbShards:          1,
	}
	arg.ChainParameters = testChainParametersCreator{
		numNodesShards: uint32(eligiblePerShard),
		numNodesMeta:   uint32(0),
		hysteresis:     hysteresis,
		adaptivity:     adaptivity,
	}.build().CurrentChainParameters()

	result, err := shuffler.UpdateNodeLists(arg)
	require.Nil(t, err)

	leavingPerShardMap, stillRemainingPerShardMap := createActuallyLeavingPerShards(unstakeLeaving, additionalLeaving, result.Leaving)

	shardId := core.SovereignChainShardId
	verifyResultsIntraShardShuffling(t,
		eligibleMap[shardId],
		waitingMap[shardId],
		additionalLeaving[shardId],
		unstakeLeaving[shardId],
		result.Eligible[shardId],
		result.Waiting[shardId],
		leavingPerShardMap[shardId],
		stillRemainingPerShardMap[shardId],
	)

	removedFromShard0 := []Validator{firstRemovedShard0, secondRemovedShard0}
	sort.Sort(validatorList(removedFromShard0))
	sort.Sort(validatorList(leavingPerShardMap[0]))
	require.Equal(t, removedFromShard0, leavingPerShardMap[0])
	found, _ := searchInMap(result.Eligible, firstRemovedShard0.PubKey())
	require.False(t, found)
	found, _ = searchInMap(result.Waiting, secondRemovedShard0.PubKey())
	require.False(t, found)
}
