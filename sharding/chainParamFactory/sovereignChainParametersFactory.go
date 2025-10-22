package chainParamFactory

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
)

type sovereignChainParametersHolderFactory struct {
}

// NewSovereignChainParametersHolderFactory creates a new sovereign chain parameters holder factory
func NewSovereignChainParametersHolderFactory() *sovereignChainParametersHolderFactory {
	return &sovereignChainParametersHolderFactory{}
}

// CreateChainParametersHolder creates a chain parameters holder component for sovereign run type
func (f *sovereignChainParametersHolderFactory) CreateChainParametersHolder(args sharding.ArgsChainParametersHolder) (process.ChainParametersHandler, error) {
	return sharding.NewSovereignChainParametersHolder(args)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (f *sovereignChainParametersHolderFactory) IsInterfaceNil() bool {
	return f == nil
}
