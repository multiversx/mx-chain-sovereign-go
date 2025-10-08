package chainSimulator

import (
	"math/big"

	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/process"
)

// ChainHandlerFactory defines what the chain factory should be able to do
type ChainHandlerFactory interface {
	CreateChainHandler(nodeHandler process.NodeHandler, monitor process.HeartbeatMonitorWithSet) (ChainHandler, error)
	IsInterfaceNil() bool
}

// ChainHandler defines what a chain handler should be able to do
type ChainHandler interface {
	IncrementRound()
	CreateNewBlock() error
	IsInterfaceNil() bool
}

// ChainSimulator defines what a chain simulator should be able to do
type ChainSimulator interface {
	GenerateBlocks(numOfBlocks int) error
	GetNodeHandler(shardID uint32) process.NodeHandler
	GenerateAddressInShard(providedShardID uint32) dtos.WalletAddress
	GenerateAndMintWalletAddress(targetShardID uint32, value *big.Int) (dtos.WalletAddress, error)
	IsInterfaceNil() bool
}
