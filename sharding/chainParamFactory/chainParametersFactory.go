package chainParamFactory

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
)

type chainParametersHolderFactory struct {
}

// NewChainParametersHolderFactory creates a new chain paramters holder factory
func NewChainParametersHolderFactory() *chainParametersHolderFactory {
	return &chainParametersHolderFactory{}
}

// CreateChainParametersHolder creates a chain parameters holder component for normal run type
func (f *chainParametersHolderFactory) CreateChainParametersHolder(args sharding.ArgsChainParametersHolder) (process.ChainParametersHandler, error) {
	return sharding.NewChainParametersHolder(args)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (f *chainParametersHolderFactory) IsInterfaceNil() bool {
	return f == nil
}
