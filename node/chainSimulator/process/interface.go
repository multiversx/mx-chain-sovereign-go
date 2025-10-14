package process

import (
	"github.com/multiversx/mx-chain-core-go/core"
	chainData "github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/api/shared"
	"github.com/multiversx/mx-chain-go/consensus"
	sovereignBlock "github.com/multiversx/mx-chain-go/dataRetriever/dataPool/sovereign"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/heartbeat/data"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
)

// NodeHandler defines what a node handler should be able to do
type NodeHandler interface {
	GetProcessComponents() factory.ProcessComponentsHolder
	GetChainHandler() chainData.ChainHandler
	GetBroadcastMessenger() consensus.BroadcastMessenger
	GetShardCoordinator() sharding.Coordinator
	GetCryptoComponents() factory.CryptoComponentsHolder
	GetCoreComponents() factory.CoreComponentsHolder
	GetDataComponents() factory.DataComponentsHolder
	GetStateComponents() factory.StateComponentsHolder
	GetFacadeHandler() shared.FacadeHandler
	GetStatusCoreComponents() factory.StatusCoreComponentsHolder
	GetNetworkComponents() factory.NetworkComponentsHolder
	GetRunTypeComponents() factory.RunTypeComponentsHolder
	GetIncomingHeaderSubscriber() process.IncomingHeaderSubscriber
	SetKeyValueForAddress(addressBytes []byte, state map[string]string) error
	SetStateForAddress(address []byte, state *dtos.AddressState) error
	RemoveAccount(address []byte) error
	ForceChangeOfEpoch() error
	GetBasePeers() map[uint32]core.PeerID
	SetBasePeers(basePeers map[uint32]core.PeerID)
	Close() error
	IsInterfaceNil() bool
}

// HeartbeatMonitorWithSet defines what a heartbeat monitor with set should be able to do
type HeartbeatMonitorWithSet interface {
	SetHeartbeats(heartbeats []data.PubKeyHeartbeat)
	IsInterfaceNil() bool
}

// BlocksProcessor defines what the block processor should be able to do
type BlocksProcessor interface {
	ProcessBlock(processor process.BlockProcessor, header chainData.HeaderHandler) (chainData.HeaderHandler, chainData.BodyHandler, error)
	ProcessHeaderProof(
		header chainData.HeaderHandler,
		proof chainData.HeaderProofHandler,
		outGoingOperationsPool sovereignBlock.ShardedOutGoingOperationPool,
	) error
	IsInterfaceNil() bool
}
