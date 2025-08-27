package sharding

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/config"
)

// NodesSetupArgs defines arguments needed to create a genesis nodes setup handler
type NodesSetupArgs struct {
	GenesisMaxNumShards      uint32
	NodesConfig              config.NodesConfig
	AddressPubKeyConverter   core.PubkeyConverter
	ValidatorPubKeyConverter core.PubkeyConverter
	ChainParametersProvider  ChainParametersHandler
}
