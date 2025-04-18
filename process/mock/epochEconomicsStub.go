package mock

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
)

// EpochEconomicsStub -
type EpochEconomicsStub struct {
	ComputeEndOfEpochEconomicsCalled func(metaBlock data.MetaHeaderHandler) (*block.Economics, error)
	VerifyRewardsPerBlockCalled      func(
		metaBlock data.MetaHeaderHandler, correctedProtocolSustainability *big.Int, computedEconomics *block.Economics,
	) error
}

// ComputeEndOfEpochEconomics -
func (e *EpochEconomicsStub) ComputeEndOfEpochEconomics(metaBlock data.MetaHeaderHandler) (*block.Economics, error) {
	if e.ComputeEndOfEpochEconomicsCalled != nil {
		return e.ComputeEndOfEpochEconomicsCalled(metaBlock)
	}
	return &block.Economics{
		RewardsForProtocolSustainability: big.NewInt(0),
	}, nil
}

// VerifyRewardsPerBlock -
func (e *EpochEconomicsStub) VerifyRewardsPerBlock(metaBlock data.MetaHeaderHandler, correctedProtocolSustainability *big.Int, computedEconomics *block.Economics) error {
	if e.VerifyRewardsPerBlockCalled != nil {
		return e.VerifyRewardsPerBlockCalled(metaBlock, correctedProtocolSustainability, computedEconomics)
	}
	return nil
}

// IsInterfaceNil -
func (e *EpochEconomicsStub) IsInterfaceNil() bool {
	return e == nil
}
