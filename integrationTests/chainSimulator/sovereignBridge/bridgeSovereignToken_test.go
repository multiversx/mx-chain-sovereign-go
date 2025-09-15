package sovereignBridge

import (
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/config"
	chainSim "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
)

const (
	defaultPathToInitialConfig = "../../../cmd/node/config/"
)

// This test will:
// - Generate a new wallet
// - Generate one ESDT token for each type
// - For each token:
//   - Deposit 0.05 EGLD main chain -> sovereign chain
//   - Call registerToken in esdt-safe
//   - Generate a receiver in a random shard & executeBridgeOp
//   - Deposit 1 token to self main chain -> sovereign chain
//
// NOTES:
// - tokens are originated from sovereign chain and have prefix
// - registerToken will issue a new token in main chain with same ticker
// - esdt-safe contract in main chain will mint with executeOperation and burn when tokens are deposited back
// - registerBridgeOp is skipped in contract execution
func TestChainSimulator_DepositAndExecuteSovereignToken(t *testing.T) {
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
	bridgeData := deployBridgeSetup(t, cs, initialAddress)
	esdtSafeAddr, _ := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Encode(bridgeData.ESDTSafeAddress)
	esdtSafeAddrShard := chainSim.GetShardForAddress(cs, esdtSafeAddr)

	wallet, err := cs.GenerateAndMintWalletAddress(1, chainSim.InitialAmount)
	require.Nil(t, err)
	nonce := uint64(0)

	err = cs.GenerateBlocks(1)
	require.NoError(t, err)

	tokens := make([]chainSim.ArgsDepositToken, 0)
	tokens = append(tokens, chainSim.ArgsDepositToken{
		Identifier: sovChainID + "-TKN-123456",
		Nonce:      uint64(0),
		Amount:     big.NewInt(14556666767),
		Type:       core.Fungible,
	})
	tokens = append(tokens, chainSim.ArgsDepositToken{
		Identifier: sovChainID + "-NFTV2-1a2b3c",
		Nonce:      uint64(1),
		Amount:     big.NewInt(1),
		Type:       core.NonFungibleV2,
	})
	tokens = append(tokens, chainSim.ArgsDepositToken{
		Identifier: sovChainID + "-DNFT-ead43f",
		Nonce:      uint64(1),
		Amount:     big.NewInt(1),
		Type:       core.DynamicNFT,
	})
	tokens = append(tokens, chainSim.ArgsDepositToken{
		Identifier: sovChainID + "-SFT-cedd55",
		Nonce:      uint64(1),
		Amount:     big.NewInt(1421),
		Type:       core.SemiFungible,
	})
	tokens = append(tokens, chainSim.ArgsDepositToken{
		Identifier: sovChainID + "-DSFT-f6b4c2",
		Nonce:      uint64(1),
		Amount:     big.NewInt(1534),
		Type:       core.DynamicSFT,
	})
	tokens = append(tokens, chainSim.ArgsDepositToken{
		Identifier: sovChainID + "-META-4b543b",
		Nonce:      uint64(1),
		Amount:     big.NewInt(6231),
		Type:       core.MetaFungible,
	})
	tokens = append(tokens, chainSim.ArgsDepositToken{
		Identifier: sovChainID + "-DMETA-5ac72b",
		Nonce:      uint64(1),
		Amount:     big.NewInt(162367),
		Type:       core.DynamicMeta,
	})
	tokensMapper := make(map[string]string)

	for _, token := range tokens {
		// deposit main chain -> sovereign chain
		// 0.05 EGLD-000000 which will be used for registerToken
		issueCost, _ := big.NewInt(0).SetString(issuePaymentCost, 10)
		egldPaymentToken := chainSim.ArgsDepositToken{
			Identifier: vmcommon.EGLDIdentifier,
			Nonce:      uint64(0),
			Amount:     issueCost,
		}
		txResult := deposit(t, cs, wallet.Bytes, &nonce, bridgeData.ESDTSafeAddress, []chainSim.ArgsDepositToken{egldPaymentToken}, wallet.Bytes, nil)
		chainSim.RequireSuccessfulTransaction(t, txResult)
		waitIfCrossShardProcessing(cs, esdtSafeAddrShard, chainSim.GetShardForAddress(cs, wallet.Bech32))
		chainSim.RequireBalance(t, cs, esdtSafeAddr, issueCost)

		// registerToken from sovereign chain
		// expecting that a new token is issued by esdt-safe contract
		registerTokens(t, cs, wallet.Bytes, &nonce, bridgeData.ESDTSafeAddress, token)
		tokensMapper[token.Identifier] = chainSim.GetIssuedEsdtIdentifier(t, cs, getTokenTicker(token.Identifier), token.Type.String())

		// transfer sovereign chain -> main chain
		// token originated from sovereign chain
		// expecting that tokens will be minted after executeOperation

		// create random receiver addresses in a shard
		receiverShardId := uint32(0)
		receiver, _ := cs.GenerateAndMintWalletAddress(receiverShardId, chainSim.InitialAmount)
		receiverNonce := uint64(0)

		// execute operations received from sovereign chain
		// expecting the token to be minted in esdt-safe contract with the same properties and transferred to receiver address
		// -------------
		// for (dynamic) SFT/MetaESDT the contract will create one more token and keep it forever
		// because if the same token is received 2nd time, the contract will just add quantity, not create different token
		txResult = executeOperation(t, cs, bridgeData.OwnerAccount.Wallet, receiver.Bytes, &bridgeData.OwnerAccount.Nonce, bridgeData.ESDTSafeAddress, []chainSim.ArgsDepositToken{token}, wallet.Bytes, nil)
		chainSim.RequireSuccessfulTransaction(t, txResult)
		receivedToken := chainSim.ArgsDepositToken{
			Identifier: tokensMapper[token.Identifier],
			Nonce:      token.Nonce,
			Amount:     token.Amount,
			Type:       token.Type,
		}
		if isSftOrMeta(receivedToken.Type) {
			chainSim.RequireAccountHasToken(t, cs, getTokenIdentifier(receivedToken), esdtSafeAddr, big.NewInt(1))
		} else {
			chainSim.RequireAccountHasToken(t, cs, getTokenIdentifier(receivedToken), esdtSafeAddr, big.NewInt(0))
		}

		waitIfCrossShardProcessing(cs, esdtSafeAddrShard, receiverShardId)
		chainSim.RequireAccountHasToken(t, cs, getTokenIdentifier(receivedToken), receiver.Bech32, receivedToken.Amount)

		// --------------------------------------------
		// deposit back 1 token, expect token to be burned and supply burned is 1
		depositAmount := big.NewInt(1)
		depositToken := chainSim.ArgsDepositToken{
			Identifier: receivedToken.Identifier,
			Nonce:      receivedToken.Nonce,
			Amount:     depositAmount,
			Type:       receivedToken.Type,
		}
		txResult = deposit(t, cs, receiver.Bytes, &receiverNonce, bridgeData.ESDTSafeAddress, []chainSim.ArgsDepositToken{depositToken}, receiver.Bytes, nil)
		chainSim.RequireSuccessfulTransaction(t, txResult)
		chainSim.RequireAccountHasToken(t, cs, getTokenIdentifier(receivedToken), receiver.Bech32, big.NewInt(0).Sub(receivedToken.Amount, depositAmount))

		waitIfCrossShardProcessing(cs, esdtSafeAddrShard, receiverShardId)

		expectedAmount := big.NewInt(0)
		if isSftOrMeta(receivedToken.Type) {
			// for (dynamic) SFT/MetaESDT, the contract should always have 1 esdt
			expectedAmount = big.NewInt(1)
		}
		chainSim.RequireAccountHasToken(t, cs, getTokenIdentifier(depositToken), esdtSafeAddr, expectedAmount)
		waitIfCrossShardProcessing(cs, esdtSafeAddrShard, receiverShardId)
		chainSim.RequireAccountHasToken(t, cs, getTokenIdentifier(receivedToken), receiver.Bech32, big.NewInt(0).Sub(receivedToken.Amount, depositAmount))

		_ = cs.GenerateBlocks(1)
		tokenSupply, err := cs.GetNodeHandler(esdtSafeAddrShard).GetFacadeHandler().GetTokenSupply(getTokenIdentifier(depositToken))
		require.Nil(t, err)
		require.NotNil(t, tokenSupply)
		require.Equal(t, depositAmount.String(), tokenSupply.Burned)

		nextShardId(&receiverShardId)
	}
}
