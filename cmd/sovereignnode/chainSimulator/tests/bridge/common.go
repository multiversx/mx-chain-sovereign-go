package bridge

import (
	"encoding/hex"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	coreAPI "github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/stretchr/testify/require"

	chainSim "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/process"
)

// ArgsBridgeSetup holds the arguments for bridge setup
type ArgsBridgeSetup struct {
	ESDTSafeAddress  []byte
	FeeMarketAddress []byte
}

// deploySovereignBridgeSetup will deploy all bridge contracts
// This function will:
// - deploy esdt-safe contract
// - deploy fee-market contract
// - set the fee-market address inside esdt-safe contract
// - disable fee in fee-market contract
// - unpause esdt-safe contract so deposit operations can start
func deploySovereignBridgeSetup(
	t *testing.T,
	cs chainSim.ChainSimulator,
	wallet dtos.WalletAddress,
	esdtSafeWasmPath string,
	feeMarketWasmPath string,
) ArgsBridgeSetup {
	nodeHandler := cs.GetNodeHandler(core.SovereignChainShardId)
	systemScAddress := chainSim.GetSysAccBytesAddress(t, nodeHandler)
	nonce := GetNonce(t, nodeHandler, wallet.Bech32)

	esdtSafeArgs := "@000000000000000005002412c4ab184562d62a3eddaa7227730e9f17c53268a3" // pre-computed fee_market_address
	esdtSafeAddress := chainSim.DeployContract(t, cs, wallet.Bytes, &nonce, systemScAddress, esdtSafeArgs, esdtSafeWasmPath)

	feeMarketArgs := "@" + hex.EncodeToString(esdtSafeAddress) + // esdt_safe_address
		"@00" // no fee
	feeMarketAddress := chainSim.DeployContract(t, cs, wallet.Bytes, &nonce, systemScAddress, feeMarketArgs, feeMarketWasmPath)

	chainSim.SendTransactionWithSuccess(t, cs, wallet.Bytes, &nonce, esdtSafeAddress, chainSim.ZeroValue, "unpause", uint64(10000000))

	return ArgsBridgeSetup{
		ESDTSafeAddress:  esdtSafeAddress,
		FeeMarketAddress: feeMarketAddress,
	}
}

// GetNonce returns account's nonce
func GetNonce(t *testing.T, nodeHandler process.NodeHandler, address string) uint64 {
	acc, _, err := nodeHandler.GetFacadeHandler().GetAccount(address, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)

	return acc.Nonce
}
