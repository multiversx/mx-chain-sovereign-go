package chaos

import (
	"github.com/multiversx/mx-chain-core-go/data"
	mainFactory "github.com/multiversx/mx-chain-go/factory"
)

type NodeHandler interface {
	GetCoreComponents() mainFactory.CoreComponentsHolder
}

type MyConsensusStateHandler interface {
	SelfPubKey() string
	ConsensusGroupSize() int
	ConsensusGroupIndex(pubKey string) (int, error)
	GetHeader() data.HeaderHandler
	GetLeader() (string, error)
	IsInterfaceNil() bool
}
