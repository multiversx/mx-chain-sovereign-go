package sharding

import (
	"fmt"

	"github.com/multiversx/mx-chain-go/config"
)

type sovereignChainParametersHolder struct {
	*chainParametersHolder
}

func NewSovereignChainParametersHolder(args ArgsChainParametersHolder) (*sovereignChainParametersHolder, error) {
	err := validateSovereignArgs(args)
	if err != nil {
		return nil, err
	}

	baseChainParamHolder, err := baseCreateChainParametersHolder(args)
	if err != nil {
		return nil, err
	}

	return &sovereignChainParametersHolder{
		chainParametersHolder: baseChainParamHolder,
	}, nil
}

func validateSovereignArgs(args ArgsChainParametersHolder) error {
	err := checkNilArgs(args)
	if err != nil {
		return err
	}
	return validateSovereignChainParameters(args.ChainParameters)
}

func validateSovereignChainParameters(chainParametersConfig []config.ChainParametersByEpochConfig) error {
	for idx, chainParameters := range chainParametersConfig {
		if chainParameters.ShardConsensusGroupSize < 1 {
			return fmt.Errorf("%w for chain parameters with index %d", ErrNegativeOrZeroConsensusGroupSize, idx)
		}
		if chainParameters.ShardMinNumNodes < chainParameters.ShardConsensusGroupSize {
			return fmt.Errorf("%w for chain parameters with index %d", ErrMinNodesPerShardSmallerThanConsensusSize, idx)
		}

		if chainParameters.MetachainConsensusGroupSize != 0 {
			return fmt.Errorf("%w for chain parameters with index %d", errSovereignInvalidMetaConsensusSize, idx)
		}
		if chainParameters.MetachainMinNumNodes != 0 {
			return fmt.Errorf("%w for chain parameters with index %d", errSovereignInvalidMetaNumNodes, idx)
		}
	}

	return nil
}
