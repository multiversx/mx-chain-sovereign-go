package sovereignBridge

import (
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/config"
	chainSim "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
)

// This test will:
// - Generate a new wallet
// - Generate one fungible token
// - Deposit 0.05 EGLD main chain -> sovereign chain
// - Call registerToken in esdt-safe
// - executeBridgeOp few times then set bad execution nonce for the operation
// - expecting the tx to fail for wrong nonce
func TestChainSimulator_ExecuteOperationBadNonce(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}
	cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck:   true,
		TempDir:                  t.TempDir(),
		PathToInitialConfig:      defaultPathToInitialConfig,
		NumOfShards:              3,
		GenesisTimestamp:         time.Now().Unix(),
		RoundDurationInMillis:    uint64(6000),
		RoundsPerEpoch:           roundsPerEpoch,
		ApiInterface:             api.NewNoApiInterface(),
		MinNodesPerShard:         3,
		MetaChainMinNodes:        3,
		NumNodesWaitingListMeta:  0,
		NumNodesWaitingListShard: 0,
		AlterConfigsFunction: func(cfg *config.Configs) {
			cfg.SystemSCConfig.ESDTSystemSCConfig.BaseIssuingCost = issuePaymentCost
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	err = cs.GenerateBlocksUntilEpochIsReached(4)
	require.Nil(t, err)

	// deploy bridge setup
	initialAddress := "erd1l6xt0rqlyzw56a3k8xwwshq2dcjwy3q9cppucvqsmdyw8r98dz3sae0kxl"
	chainSim.InitAddressesAndSysAccState(t, cs, initialAddress)
	bridgeData := deploySovereignBridgeOnMainChain(t, cs, initialAddress)
	esdtSafeAddr, _ := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Encode(bridgeData.ESDTSafeAddress)

	wallet, err := cs.GenerateAndMintWalletAddress(1, chainSim.InitialAmount)
	require.Nil(t, err)
	nonce := uint64(0)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	sovereignToken := chainSim.ArgsDepositToken{
		Identifier: sovChainID + "-TKN-123456",
		Nonce:      uint64(0),
		Amount:     big.NewInt(1),
		Type:       core.Fungible,
	}
	registerSovereignToken(t, cs, bridgeData, wallet, &nonce, sovereignToken)
	mappedToken := chainSim.GetIssuedEsdtIdentifier(t, cs, getTokenTicker(sovereignToken.Identifier), sovereignToken.Type.String())

	numOfTransfers := 4
	numOfSuccessfulTransfers := 2
	var lastNonce uint64
	for i := 0; i < numOfTransfers; i++ {
		if i == numOfSuccessfulTransfers {
			lastNonce = bridgeData.ExecutionNonce
			bridgeData.ExecutionNonce = 100 // set bad nonce
		}

		txResult := executeOperation(t, cs, bridgeData, wallet.Bytes, []chainSim.ArgsDepositToken{sovereignToken}, wallet.Bytes, nil)

		if i < numOfSuccessfulTransfers {
			chainSim.RequireSuccessfulTransaction(t, txResult)
		} else {
			chainSim.RequireInternalVMError(t, txResult, "The operation nonce is incorrect")
		}
	}
	chainSim.RequireAccountHasToken(t, cs, getTokenIdentifier(sovereignToken), esdtSafeAddr, big.NewInt(0))
	chainSim.RequireAccountHasToken(t, cs, mappedToken, wallet.Bech32, big.NewInt(int64(numOfSuccessfulTransfers)))

	bridgeData.ExecutionNonce = lastNonce
	txResult := executeOperation(t, cs, bridgeData, wallet.Bytes, []chainSim.ArgsDepositToken{sovereignToken}, wallet.Bytes, nil)
	chainSim.RequireSuccessfulTransaction(t, txResult)
	chainSim.RequireAccountHasToken(t, cs, getTokenIdentifier(sovereignToken), esdtSafeAddr, big.NewInt(0))
	chainSim.RequireAccountHasToken(t, cs, mappedToken, wallet.Bech32, big.NewInt(int64(numOfSuccessfulTransfers)+1))
}
