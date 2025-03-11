package bootstrap

import "github.com/multiversx/mx-chain-sovereign-go/factory"

func (bc *bootstrapComponents) EpochStartBootstrapper() factory.EpochStartBootstrapper {
	return bc.epochStartBootstrapper
}
