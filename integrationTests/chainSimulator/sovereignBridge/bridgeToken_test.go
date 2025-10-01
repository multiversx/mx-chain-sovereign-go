package sovereignBridge

import (
	"encoding/hex"
	"fmt"
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

type wallet struct {
	addrBech32 string
	nonce      uint64
}

// This test will:
// - Generate wallets in different shards and issue one ESDT token
// - For each wallet:
// - Deposit 1 token to self
// - ExecuteBridgeOp 1 token to self
// NOTES:
// - tokens are originated from main chain
// - esdt-safe contract in main chain will save the tokens at deposit, then send from balance at executeOperation
// - registerBridgeOp is skipped in contract execution
func TestChainSimulator_DepositAndExecuteMainChainToken(t *testing.T) {
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

	wallets := generateAccountsAndTokens(t, cs)
	amountToTransfer := big.NewInt(1)

	// transfer main chain -> sovereign chain
	// token originated from main chain
	// expecting that tokens to be saved in esdt-safe contract after deposit
	for account, token := range wallets {
		accountAddrBytes, _ := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Decode(account.addrBech32)

		// deposit 1 token, expect to be saved in contract and supply burned is 0
		depositToken := chainSim.ArgsDepositToken{
			Identifier: token.Identifier,
			Nonce:      token.Nonce,
			Amount:     amountToTransfer,
			Type:       token.Type,
		}
		txResult := deposit(t, cs, accountAddrBytes, &account.nonce, bridgeData.ESDTSafeAddress, []chainSim.ArgsDepositToken{depositToken}, accountAddrBytes, nil)
		chainSim.RequireSuccessfulTransaction(t, txResult)
		chainSim.RequireAccountHasToken(t, cs, getTokenIdentifier(depositToken), account.addrBech32, big.NewInt(0).Sub(token.Amount, amountToTransfer))
		waitIfCrossShardProcessing(cs, chainSim.GetShardForAddress(cs, account.addrBech32), esdtSafeAddrShard)
		chainSim.RequireAccountHasToken(t, cs, getTokenIdentifier(depositToken), esdtSafeAddr, amountToTransfer)

		_ = cs.GenerateBlocks(1)
		tokenSupply, err := cs.GetNodeHandler(esdtSafeAddrShard).GetFacadeHandler().GetTokenSupply(getTokenIdentifier(depositToken))
		require.Nil(t, err)
		require.NotNil(t, tokenSupply)
		require.Equal(t, "0", tokenSupply.Burned) // nothing burned
	}

	// transfer sovereign chain -> main chain
	// token originated from main chain
	// expecting that tokens will be transferred from contract balance after executeOperation
	for account, token := range wallets {
		accountAddrBytes, _ := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Decode(account.addrBech32)

		executeToken := chainSim.ArgsDepositToken{
			Identifier: token.Identifier,
			Nonce:      token.Nonce,
			Amount:     amountToTransfer,
			Type:       token.Type,
		}
		// execute operations received from sovereign chain
		// expecting the token to be transferred from esdt-safe contract to wallet address
		txResult := executeOperation(t, cs, bridgeData, accountAddrBytes, []chainSim.ArgsDepositToken{executeToken}, accountAddrBytes, nil)
		chainSim.RequireSuccessfulTransaction(t, txResult)
		chainSim.RequireAccountHasToken(t, cs, getTokenIdentifier(executeToken), esdtSafeAddr, big.NewInt(0))
		waitIfCrossShardProcessing(cs, esdtSafeAddrShard, chainSim.GetShardForAddress(cs, account.addrBech32))
		chainSim.RequireAccountHasToken(t, cs, getTokenIdentifier(executeToken), account.addrBech32, token.Amount) // token original amount
	}
}

func generateAccountsAndTokens(
	t *testing.T,
	cs chainSim.ChainSimulator,
) map[wallet]chainSim.ArgsDepositToken {
	accountShardId := uint32(0)
	issueCost, _ := big.NewInt(0).SetString(issuePaymentCost, 10)
	wallets := make(map[wallet]chainSim.ArgsDepositToken)

	account, accountAddrBytes := createNewAccount(cs, &accountShardId)
	supply := big.NewInt(14556666767)
	tokenId := chainSim.IssueFungible(t, cs, accountAddrBytes, &account.nonce, issueCost, "TKN", "TKN", 18, supply)
	token := chainSim.ArgsDepositToken{
		Identifier: tokenId,
		Nonce:      uint64(0),
		Amount:     supply,
		Type:       core.Fungible,
	}
	wallets[account] = token

	account, accountAddrBytes = createNewAccount(cs, &accountShardId)
	collectionId := chainSim.RegisterAndSetAllRoles(t, cs, accountAddrBytes, &account.nonce, issueCost, "NFTV2", "NFTV2", core.NonFungibleESDTv2, 0)
	supply = big.NewInt(1)
	createArgs := createNftArgs(collectionId, supply, "NFTV2 #1")
	chainSim.SendTransactionWithSuccess(t, cs, accountAddrBytes, &account.nonce, accountAddrBytes, chainSim.ZeroValue, createArgs, uint64(60000000))
	token = chainSim.ArgsDepositToken{
		Identifier: collectionId,
		Nonce:      uint64(1),
		Amount:     supply,
		Type:       core.NonFungibleV2,
	}
	wallets[account] = token

	account, accountAddrBytes = createNewAccount(cs, &accountShardId)
	collectionId = chainSim.RegisterAndSetAllRolesDynamic(t, cs, accountAddrBytes, &account.nonce, issueCost, "DNFT", "DNFT", core.DynamicNFTESDT, 0)
	supply = big.NewInt(1)
	createArgs = createNftArgs(collectionId, supply, "DNFT #1")
	chainSim.SendTransactionWithSuccess(t, cs, accountAddrBytes, &account.nonce, accountAddrBytes, chainSim.ZeroValue, createArgs, uint64(60000000))
	token = chainSim.ArgsDepositToken{
		Identifier: collectionId,
		Nonce:      uint64(1),
		Amount:     supply,
		Type:       core.DynamicNFT,
	}
	wallets[account] = token

	account, accountAddrBytes = createNewAccount(cs, &accountShardId)
	collectionId = chainSim.RegisterAndSetAllRoles(t, cs, accountAddrBytes, &account.nonce, issueCost, "SFT", "SFT", core.SemiFungibleESDT, 0)
	supply = big.NewInt(125)
	createArgs = createNftArgs(collectionId, supply, "SFT #1")
	chainSim.SendTransactionWithSuccess(t, cs, accountAddrBytes, &account.nonce, accountAddrBytes, chainSim.ZeroValue, createArgs, uint64(60000000))
	token = chainSim.ArgsDepositToken{
		Identifier: collectionId,
		Nonce:      uint64(1),
		Amount:     supply,
		Type:       core.SemiFungible,
	}
	wallets[account] = token

	account, accountAddrBytes = createNewAccount(cs, &accountShardId)
	collectionId = chainSim.RegisterAndSetAllRolesDynamic(t, cs, accountAddrBytes, &account.nonce, issueCost, "DSFT", "DSFT", core.DynamicSFTESDT, 0)
	supply = big.NewInt(125)
	createArgs = createNftArgs(collectionId, supply, "DSFT #1")
	chainSim.SendTransactionWithSuccess(t, cs, accountAddrBytes, &account.nonce, accountAddrBytes, chainSim.ZeroValue, createArgs, uint64(60000000))
	token = chainSim.ArgsDepositToken{
		Identifier: collectionId,
		Nonce:      uint64(1),
		Amount:     supply,
		Type:       core.DynamicSFT,
	}
	wallets[account] = token

	account, accountAddrBytes = createNewAccount(cs, &accountShardId)
	collectionId = chainSim.RegisterAndSetAllRoles(t, cs, accountAddrBytes, &account.nonce, issueCost, "META", "META", core.MetaESDT, 15)
	supply = big.NewInt(234673347)
	createArgs = createNftArgs(collectionId, supply, "META #1")
	chainSim.SendTransactionWithSuccess(t, cs, accountAddrBytes, &account.nonce, accountAddrBytes, chainSim.ZeroValue, createArgs, uint64(60000000))
	token = chainSim.ArgsDepositToken{
		Identifier: collectionId,
		Nonce:      uint64(1),
		Amount:     supply,
		Type:       core.MetaFungible,
	}
	wallets[account] = token

	account, accountAddrBytes = createNewAccount(cs, &accountShardId)
	collectionId = chainSim.RegisterAndSetAllRolesDynamic(t, cs, accountAddrBytes, &account.nonce, issueCost, "DMETA", "DMETA", core.DynamicMetaESDT, 10)
	supply = big.NewInt(64865382)
	createArgs = createNftArgs(collectionId, supply, "DMETA #1")
	chainSim.SendTransactionWithSuccess(t, cs, accountAddrBytes, &account.nonce, accountAddrBytes, chainSim.ZeroValue, createArgs, uint64(60000000))
	token = chainSim.ArgsDepositToken{
		Identifier: collectionId,
		Nonce:      uint64(1),
		Amount:     supply,
		Type:       core.DynamicMeta,
	}
	wallets[account] = token

	return wallets
}

func createNewAccount(cs chainSim.ChainSimulator, shardId *uint32) (wallet, []byte) {
	walletAddress, _ := cs.GenerateAndMintWalletAddress(*shardId, chainSim.InitialAmount)
	nextShardId(shardId)
	_ = cs.GenerateBlocks(1)

	return wallet{
		addrBech32: walletAddress.Bech32,
		nonce:      uint64(0),
	}, walletAddress.Bytes
}

func createNftArgs(tokenIdentifier string, initialSupply *big.Int, name string) string {
	return "ESDTNFTCreate" +
		"@" + hex.EncodeToString([]byte(tokenIdentifier)) +
		"@" + hex.EncodeToString(initialSupply.Bytes()) +
		"@" + hex.EncodeToString([]byte(name)) +
		"@" + fmt.Sprintf("%04X", 2500) + // royalties 25%
		"@" + // hash
		"@" + // attributes
		"@" // uri
}

// main chain deposit and execute with no payment, only transfer data
// 1. call deposit endpoint with no payment and no transfer data - should fail
// 2. call deposit endpoint with no payment and with transfer data - should work
// 3. call execute endpoint with no payment and with transfer data - should work
func TestChainSimulator_DepositAndExecuteNoPaymentWithTransferData(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck: true,
		TempDir:                t.TempDir(),
		PathToInitialConfig:    defaultPathToInitialConfig,
		NumOfShards:            3,
		GenesisTimestamp:       time.Now().Unix(),
		RoundDurationInMillis:  uint64(6000),
		RoundsPerEpoch: core.OptionalUint64{
			HasValue: true,
			Value:    20,
		},
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

	wallet, err := cs.GenerateAndMintWalletAddress(0, chainSim.InitialAmount)
	require.Nil(t, err)
	nonce := uint64(0)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	trnsData := &transferData{
		GasLimit: uint64(10000000),
		Function: []byte("hello"),
		Args:     [][]byte{{0x01}},
	}

	// call deposit endpoint with no payment and no transfer data
	txResult := deposit(t, cs, wallet.Bytes, &nonce, bridgeData.ESDTSafeAddress, make([]chainSim.ArgsDepositToken, 0), wallet.Bytes, nil)
	chainSim.RequireSignalError(t, txResult, "Nothing to transfer")

	// call deposit endpoint with no payment and with transfer data
	txResult = deposit(t, cs, wallet.Bytes, &nonce, bridgeData.ESDTSafeAddress, make([]chainSim.ArgsDepositToken, 0), wallet.Bytes, trnsData)
	chainSim.RequireSuccessfulTransaction(t, txResult)

	// call execute endpoint with no payment and with transfer data
	receiverContracts := deployReceiverContractInAllShards(t, cs) // generate hello contracts in each shard
	for shardId := uint32(0); shardId < cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().NumberOfShards(); shardId++ {
		// get contract from a specific shard
		receiver := receiverContracts[shardId]

		// the executed operation in hello contract should work
		txResult = executeOperation(t, cs, bridgeData, receiver.Bytes, make([]chainSim.ArgsDepositToken, 0), wallet.Bytes, trnsData)
		chainSim.RequireSuccessfulTransaction(t, txResult)
	}
}
