package chainParamFactory

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
)

// ChainParametersHolderFactory defines the chain params holder factory handler
type ChainParametersHolderFactory interface {
	CreateChainParametersHolder(args sharding.ArgsChainParametersHolder) (process.ChainParametersHandler, error)
	IsInterfaceNil() bool
}
