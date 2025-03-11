package consensus

import "github.com/multiversx/mx-chain-sovereign-go/process"

func (cc *consensusComponents) BootStrapper() process.Bootstrapper {
	return cc.bootstrapper
}
