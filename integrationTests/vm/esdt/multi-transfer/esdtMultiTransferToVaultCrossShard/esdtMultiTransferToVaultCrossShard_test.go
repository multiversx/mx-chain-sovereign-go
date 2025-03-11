package esdtMultiTransferToVaultCrossShard

import (
	"testing"

	multitransfer "github.com/multiversx/mx-chain-sovereign-go/integrationTests/vm/esdt/multi-transfer"
)

func TestESDTMultiTransferToVaultCrossShard(t *testing.T) {
	multitransfer.EsdtMultiTransferToVault(t, true, "../../testdata/vaultV2.wasm")
}
