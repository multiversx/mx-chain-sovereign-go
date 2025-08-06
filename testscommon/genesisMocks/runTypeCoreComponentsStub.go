package genesisMocks

import (
	"github.com/multiversx/mx-chain-go/common/enablers"
	"github.com/multiversx/mx-chain-go/config"
	genesisMocks "github.com/multiversx/mx-chain-go/genesis/mock"
	"github.com/multiversx/mx-chain-go/process/rating"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/chainParamFactory"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/testscommon"
)

// RunTypeCoreComponentsStub -
type RunTypeCoreComponentsStub struct {
	GenesisNodesSetupFactory     sharding.GenesisNodesSetupFactory
	RatingsDataFactory           rating.RatingsDataFactory
	EnableEpochsFactory          enablers.EnableEpochsFactory
	ChainParametersFactory       chainParamFactory.ChainParametersHolderFactory
	HashValidatorShufflerFactory nodesCoordinator.HashValidatorShufflerFactory
}

// NewRunTypeCoreComponentsStub -
func NewRunTypeCoreComponentsStub() *RunTypeCoreComponentsStub {
	return &RunTypeCoreComponentsStub{
		GenesisNodesSetupFactory:     &genesisMocks.GenesisNodesSetupFactoryMock{},
		RatingsDataFactory:           &testscommon.RatingsDataFactoryMock{},
		EnableEpochsFactory:          enablers.NewEnableEpochsFactory(),
		ChainParametersFactory:       chainParamFactory.NewChainParametersHolderFactory(),
		HashValidatorShufflerFactory: nodesCoordinator.NewHashValidatorShufflerFactory(),
	}
}

// NewSovereignRunTypeCoreComponentsStub -
func NewSovereignRunTypeCoreComponentsStub() *RunTypeCoreComponentsStub {
	return &RunTypeCoreComponentsStub{
		GenesisNodesSetupFactory:     &genesisMocks.GenesisNodesSetupFactoryMock{},
		RatingsDataFactory:           &testscommon.RatingsDataFactoryMock{},
		EnableEpochsFactory:          enablers.NewSovereignEnableEpochsFactory(config.SovereignEpochConfig{}),
		ChainParametersFactory:       chainParamFactory.NewSovereignChainParametersHolderFactory(),
		HashValidatorShufflerFactory: nodesCoordinator.NewSovereignHashValidatorShufflerFactory(),
	}
}

// Create -
func (r *RunTypeCoreComponentsStub) Create() error {
	return nil
}

// Close -
func (r *RunTypeCoreComponentsStub) Close() error {
	return nil
}

// CheckSubcomponents -
func (r *RunTypeCoreComponentsStub) CheckSubcomponents() error {
	return nil
}

// String -
func (r *RunTypeCoreComponentsStub) String() string {
	return ""
}

// GenesisNodesSetupFactoryCreator -
func (r *RunTypeCoreComponentsStub) GenesisNodesSetupFactoryCreator() sharding.GenesisNodesSetupFactory {
	return r.GenesisNodesSetupFactory
}

// RatingsDataFactoryCreator -
func (r *RunTypeCoreComponentsStub) RatingsDataFactoryCreator() rating.RatingsDataFactory {
	return r.RatingsDataFactory
}

// EnableEpochsFactoryCreator -
func (r *RunTypeCoreComponentsStub) EnableEpochsFactoryCreator() enablers.EnableEpochsFactory {
	return r.EnableEpochsFactory
}

// ChainParametersHolderFactory -
func (r *RunTypeCoreComponentsStub) ChainParametersHolderFactory() chainParamFactory.ChainParametersHolderFactory {
	return r.ChainParametersFactory
}

// HashValidatorShufflerFactoryCreator -
func (r *RunTypeCoreComponentsStub) HashValidatorShufflerFactoryCreator() nodesCoordinator.HashValidatorShufflerFactory {
	return r.HashValidatorShufflerFactory
}

// IsInterfaceNil -
func (r *RunTypeCoreComponentsStub) IsInterfaceNil() bool {
	return r == nil
}
