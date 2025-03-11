package runType

import (
	"github.com/multiversx/mx-chain-sovereign-go/common/enablers"
	"github.com/multiversx/mx-chain-sovereign-go/process/rating"
	"github.com/multiversx/mx-chain-sovereign-go/sharding"
)

type runTypeCoreComponents struct {
	genesisNodesSetupFactory sharding.GenesisNodesSetupFactory
	ratingsDataFactory       rating.RatingsDataFactory
	enableEpochsFactory      enablers.EnableEpochsFactory
}

// Close does nothing
func (rcc *runTypeCoreComponents) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (rcc *runTypeCoreComponents) IsInterfaceNil() bool {
	return rcc == nil
}
