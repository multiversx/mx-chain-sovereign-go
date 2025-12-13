package runType

import (
	"github.com/multiversx/mx-chain-go/common/enablers"
	"github.com/multiversx/mx-chain-go/process/rating"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/chainParamFactory"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
)

type runTypeCoreComponents struct {
	genesisNodesSetupFactory     sharding.GenesisNodesSetupFactory
	ratingsDataFactory           rating.RatingsDataFactory
	enableEpochsFactory          enablers.EnableEpochsFactory
	chainParametersFactory       chainParamFactory.ChainParametersHolderFactory
	hashValidatorShufflerFactory nodesCoordinator.HashValidatorShufflerFactory
}

// Close does nothing
func (rcc *runTypeCoreComponents) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (rcc *runTypeCoreComponents) IsInterfaceNil() bool {
	return rcc == nil
}
