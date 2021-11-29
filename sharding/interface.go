package sharding

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/sharding/nodesCoordinator"
)

type Validator = nodesCoordinator.Validator

// Coordinator defines what a shard state coordinator should hold
type Coordinator interface {
	NumberOfShards() uint32
	ComputeId(address []byte) uint32
	SelfId() uint32
	SameShard(firstAddress, secondAddress []byte) bool
	CommunicationIdentifier(destShardID uint32) string
	IsInterfaceNil() bool
}

// NodesCoordinator defines the behaviour of a struct able to do validator group selection
type NodesCoordinator interface {
	nodesCoordinator.NodesCoordinatorLite
	ShuffleOutForEpoch(_ uint32)
	LoadState(key []byte) error
	GetConsensusWhitelistedNodes(epoch uint32) (map[string]struct{}, error)
	GetValidatorWithPublicKey(publicKey []byte) (validator Validator, shardId uint32, err error)
	GetSavedStateKey() []byte
}

// EpochHandler defines what a component which handles current epoch should be able to do
type EpochHandler interface {
	MetaEpoch() uint32
	IsInterfaceNil() bool
}

// ArgsUpdateNodes holds the parameters required by the shuffler to generate a new nodes configuration
type ArgsUpdateNodes struct {
	Eligible          map[uint32][]Validator
	Waiting           map[uint32][]Validator
	NewNodes          []Validator
	UnStakeLeaving    []Validator
	AdditionalLeaving []Validator
	Rand              []byte
	NbShards          uint32
	Epoch             uint32
}

// ResUpdateNodes holds the result of the UpdateNodes method
type ResUpdateNodes struct {
	Eligible       map[uint32][]Validator
	Waiting        map[uint32][]Validator
	Leaving        []Validator
	StillRemaining []Validator
}

// NodesShuffler provides shuffling functionality for nodes
type NodesShuffler interface {
	UpdateParams(numNodesShard uint32, numNodesMeta uint32, hysteresis float32, adaptivity bool)
	UpdateNodeLists(args ArgsUpdateNodes) (*ResUpdateNodes, error)
	IsInterfaceNil() bool
}

//PeerAccountListAndRatingHandler provides Rating Computation Capabilites for the Nodes Coordinator and ValidatorStatistics
type PeerAccountListAndRatingHandler interface {
	//GetChance returns the chances for the the rating
	GetChance(uint32) uint32
	//GetStartRating gets the start rating values
	GetStartRating() uint32
	//GetSignedBlocksThreshold gets the threshold for the minimum signed blocks
	GetSignedBlocksThreshold() float32
	//ComputeIncreaseProposer computes the new rating for the increaseLeader
	ComputeIncreaseProposer(shardId uint32, currentRating uint32) uint32
	//ComputeDecreaseProposer computes the new rating for the decreaseLeader
	ComputeDecreaseProposer(shardId uint32, currentRating uint32, consecutiveMisses uint32) uint32
	//RevertIncreaseValidator computes the new rating if a revert for increaseProposer should be done
	RevertIncreaseValidator(shardId uint32, currentRating uint32, nrReverts uint32) uint32
	//ComputeIncreaseValidator computes the new rating for the increaseValidator
	ComputeIncreaseValidator(shardId uint32, currentRating uint32) uint32
	//ComputeDecreaseValidator computes the new rating for the decreaseValidator
	ComputeDecreaseValidator(shardId uint32, currentRating uint32) uint32
	//IsInterfaceNil verifies if the interface is nil
	IsInterfaceNil() bool
}

//ChanceComputer provides chance computation capabilities based on a rating
type ChanceComputer interface {
	//GetChance returns the chances for the the rating
	GetChance(uint32) uint32
	//IsInterfaceNil verifies if the interface is nil
	IsInterfaceNil() bool
}

//Cacher provides the capabilities needed to store and retrieve information needed in the NodesCoordinator
type Cacher interface {
	// Clear is used to completely clear the cache.
	Clear()
	// Put adds a value to the cache.  Returns true if an eviction occurred.
	Put(key []byte, value interface{}, sizeInBytes int) (evicted bool)
	// Get looks up a key's value from the cache.
	Get(key []byte) (value interface{}, ok bool)
}

// ShuffledOutHandler defines the methods needed for the computation of a shuffled out event
type ShuffledOutHandler interface {
	Process(newShardID uint32) error
	RegisterHandler(handler func(newShardID uint32))
	CurrentShardID() uint32
	IsInterfaceNil() bool
}

// RandomSelector selects randomly a subset of elements from a set of data
type RandomSelector interface {
	Select(randSeed []byte, sampleSize uint32) ([]uint32, error)
	IsInterfaceNil() bool
}

// EpochStartActionHandler defines the action taken on epoch start event
type EpochStartActionHandler interface {
	EpochStartAction(hdr data.HeaderHandler)
	EpochStartPrepare(metaHdr data.HeaderHandler, body data.BodyHandler)
	NotifyOrder() uint32
}

// GenesisNodesSetupHandler returns the genesis nodes info
type GenesisNodesSetupHandler interface {
	AllInitialNodes() []GenesisNodeInfoHandler
	InitialNodesPubKeys() map[uint32][]string
	GetShardIDForPubKey(pubkey []byte) (uint32, error)
	InitialEligibleNodesPubKeysForShard(shardId uint32) ([]string, error)
	InitialNodesInfoForShard(shardId uint32) ([]GenesisNodeInfoHandler, []GenesisNodeInfoHandler, error)
	InitialNodesInfo() (map[uint32][]GenesisNodeInfoHandler, map[uint32][]GenesisNodeInfoHandler)
	GetStartTime() int64
	GetRoundDuration() uint64
	GetShardConsensusGroupSize() uint32
	GetMetaConsensusGroupSize() uint32
	NumberOfShards() uint32
	MinNumberOfNodes() uint32
	MinNumberOfShardNodes() uint32
	MinNumberOfMetaNodes() uint32
	GetHysteresis() float32
	GetAdaptivity() bool
	MinNumberOfNodesWithHysteresis() uint32
	IsInterfaceNil() bool
}

// GenesisNodeInfoHandler defines the public methods for the genesis nodes info
type GenesisNodeInfoHandler interface {
	AssignedShard() uint32
	AddressBytes() []byte
	PubKeyBytes() []byte
	GetInitialRating() uint32
	IsInterfaceNil() bool
}

// ValidatorsDistributor distributes validators across shards
type ValidatorsDistributor interface {
	DistributeValidators(destination map[uint32][]Validator, source map[uint32][]Validator, rand []byte, balanced bool) error
	IsInterfaceNil() bool
}
