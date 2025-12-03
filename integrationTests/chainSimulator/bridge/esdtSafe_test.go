package bridge

import (
	"encoding/hex"
	"fmt"
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
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
)

const (
	helloWasmPath = "testdata/hello.wasm"
)

type wallet struct {
	addrBech32 string
	nonce      uint64
}

// This test will:
// - Generate a new wallet
// - Generate one ESDT token for each type (sovereign token with prefix)
// - For each token:
//   - Deposit 0.05 EGLD-000000 on esdt-safe in main chain (simulate deposit main to sovereign)
//   - Call registerToken to register the token in esdt-safe (simulate register sovereign to main)
//   - Generate a receiver in a random shard and execute a token transfer in esdt-safe (simulate deposit sovereign to main)
//   - Deposit one token back to sovereign (simulate deposit main to sovereign)
//
// - For each token:
//   - execute with SC call the previously transferred token in esdt-safe (simulate deposit sovereign to main with transfer data)
//
// NOTES:
// - tokens are originated from sovereign chain and have prefix
// - registerToken will issue a new token in main chain with same ticker
// - esdt-safe contract in main chain will mint tokens on execute and burn when tokens are deposited back
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
	bridgeData := deploySovereignBridgeOnMainChain(t, cs, initialAddress)
	esdtSafeAddr, _ := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Encode(bridgeData.ESDTSafeAddress)
	esdtSafeAddrShard := chainSim.GetShardForAddress(cs, esdtSafeAddr)

	wallet, err := cs.GenerateAndMintWalletAddress(0, chainSim.InitialAmount)
	require.Nil(t, err)
	nonce := uint64(0)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	receiverShardId := uint32(0)

	tokens := generateSovereignTokens()
	tokensMapper := make(map[string]string)

	// transfer sovereign chain -> main chain -> sovereign chain
	// token originated from sovereign chain
	for _, token := range tokens {
		// register sovereign token identifier
		// this will issue a new token on main chain and create a mapper between the identifiers
		registerSovereignToken(t, cs, wallet, &nonce, bridgeData.ESDTSafeAddress, token)
		tokensMapper[token.Identifier] = chainSim.GetIssuedEsdtIdentifier(t, cs, getTokenTicker(token.Identifier), token.Type.String())

		// create random receiver addresses in a shard
		receiver, _ := cs.GenerateAndMintWalletAddress(receiverShardId, chainSim.InitialAmount)
		receiverNonce := uint64(0)

		// execute operations received from sovereign chain
		// expecting the token to be minted in esdt-safe contract with the same properties and transferred to receiver address
		// -------------
		// for (dynamic) SFT/MetaESDT the contract will create one more token and keep it forever
		// because if the same token is received 2nd time, the contract will just add quantity, not create different token
		txResult := executeOperation(t, cs, bridgeData.OwnerAccount.Wallet, receiver.Bytes, &bridgeData.OwnerAccount.Nonce, bridgeData.ESDTSafeAddress, []chainSim.ArgsDepositToken{token}, wallet.Bytes, nil)
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

		_ = cs.GenerateBlocks(1)
		tokenSupply, err := cs.GetNodeHandler(esdtSafeAddrShard).GetFacadeHandler().GetTokenSupply(getTokenIdentifier(depositToken))
		require.Nil(t, err)
		require.NotNil(t, tokenSupply)
		require.Equal(t, depositAmount.String(), tokenSupply.Burned)

		nextShardId(&receiverShardId)
	}

	// generate hello contracts in each shard
	// hello contract has one endpoint "hello" which receives a number as argument
	// if number is 0 then will throw error, otherwise will do nothing
	receiverContracts := deployReceiverContractInAllShards(t, cs)

	// transfer sovereign chain -> main chain with transfer data
	// execute 2nd time the same operations
	// for (dynamic) NFT, expect creation with nonce+1 the second time because that's the next nonce
	// for (dynamic) SFT/MetaESDT, the contract should just add quantity for existing nonce
	for _, token := range tokens {
		// get contract from next shard
		receiver := receiverContracts[receiverShardId]

		// execute operations received from sovereign chain
		// expecting the token to be minted in esdt-safe contract with the same properties and transferred to receiver contract address
		trnsData := &transferData{
			GasLimit: uint64(10000000),
			Function: []byte("hello"),
			Args:     [][]byte{{0x01}},
		}
		txResult := executeOperation(t, cs, bridgeData.OwnerAccount.Wallet, receiver.Bytes, &bridgeData.OwnerAccount.Nonce, bridgeData.ESDTSafeAddress, []chainSim.ArgsDepositToken{token}, wallet.Bytes, trnsData)
		chainSim.RequireSuccessfulTransaction(t, txResult)
		receivedToken := chainSim.ArgsDepositToken{
			Identifier: tokensMapper[token.Identifier],
			Nonce:      token.Nonce,
			Amount:     token.Amount,
			Type:       token.Type,
		}
		if isNft(token.Type) {
			receivedToken.Nonce++
		}

		waitIfCrossShardProcessing(cs, esdtSafeAddrShard, receiverShardId)
		_ = cs.GenerateBlocks(1) // one more block required for ESDTTransfer esdt-safe -> hello

		chainSim.RequireAccountHasToken(t, cs, getTokenIdentifier(receivedToken), receiver.Bech32, receivedToken.Amount)
		if isSftOrMeta(receivedToken.Type) { // expect the contract to have 1 token
			chainSim.RequireAccountHasToken(t, cs, getTokenIdentifier(receivedToken), esdtSafeAddr, big.NewInt(1))
		} else {
			chainSim.RequireAccountHasToken(t, cs, getTokenIdentifier(receivedToken), esdtSafeAddr, big.NewInt(0))
		}

		nextShardId(&receiverShardId)
	}
}

// This test will:
// - Generate wallets in different shards and issue one ESDT token
// - For each wallet:
//   - call Deposit with one main chain token
//   - call ExecuteBridgeOp with the main chain token
//
// NOTES:
// - tokens are originated from main chain
// - esdt-safe contract in main chain will save the tokens at deposit, then send from balance at executeOperation
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
	bridgeData := deploySovereignBridgeOnMainChain(t, cs, initialAddress)
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
		txResult := executeOperation(t, cs, bridgeData.OwnerAccount.Wallet, accountAddrBytes, &bridgeData.OwnerAccount.Nonce, bridgeData.ESDTSafeAddress, []chainSim.ArgsDepositToken{executeToken}, accountAddrBytes, nil)
		chainSim.RequireSuccessfulTransaction(t, txResult)
		chainSim.RequireAccountHasToken(t, cs, getTokenIdentifier(executeToken), esdtSafeAddr, big.NewInt(0))
		waitIfCrossShardProcessing(cs, esdtSafeAddrShard, chainSim.GetShardForAddress(cs, account.addrBech32))
		chainSim.RequireAccountHasToken(t, cs, getTokenIdentifier(executeToken), account.addrBech32, token.Amount) // token original amount
	}
}

// transfer from sovereign chain to main chain with transfer data
// tokens are originated from sovereign chain
// the execution is always expected to fail because of transfer data arguments
// we also check that tokens are burned if the execution fails
func TestChainSimulator_ExecuteSovereignTokenWithTransferDataFails(t *testing.T) {
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
	esdtSafeAddrShard := chainSim.GetShardForAddress(cs, esdtSafeAddr)

	wallet, err := cs.GenerateAndMintWalletAddress(0, chainSim.InitialAmount)
	require.Nil(t, err)
	nonce := uint64(0)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	receiverShardId := uint32(0)

	// generate hello contracts in each shard
	// hello contract has one endpoint "hello" which receives a number as argument
	// if number is 0 then will throw error, otherwise will do nothing
	receiverContracts := deployReceiverContractInAllShards(t, cs)

	tokens := generateSovereignTokens()
	tokensMapper := make(map[string]string)

	// transfer sovereign chain -> main chain -> sovereign chain
	// token originated from sovereign chain
	for _, token := range tokens {
		// register sovereign token identifier
		// this will issue a new token on main chain and create a mapper between the identifiers
		registerSovereignToken(t, cs, wallet, &nonce, bridgeData.ESDTSafeAddress, token)
		tokensMapper[token.Identifier] = chainSim.GetIssuedEsdtIdentifier(t, cs, getTokenTicker(token.Identifier), token.Type.String())

		// get contract from next shard
		receiver := receiverContracts[receiverShardId]

		// execute operations received from sovereign chain
		// expecting the token to be minted in esdt-safe contract with the same properties and transferred with SC call to hello contract
		// for (dynamic) SFT/MetaESDT the contract will create one more token and keep it forever
		trnsData := &transferData{
			GasLimit: uint64(10000000),
			Function: []byte("hello"),
			Args:     [][]byte{{0x00}},
		}
		// the executed operation in hello contract is expected to fail, tokens will be minted and then burned
		txResult := executeOperation(t, cs, bridgeData.OwnerAccount.Wallet, receiver.Bytes, &bridgeData.OwnerAccount.Nonce, bridgeData.ESDTSafeAddress, []chainSim.ArgsDepositToken{token}, wallet.Bytes, trnsData)
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

		// no tokens should be in hello contract
		waitIfCrossShardProcessing(cs, esdtSafeAddrShard, receiverShardId)
		chainSim.RequireAccountHasToken(t, cs, getTokenIdentifier(receivedToken), receiver.Bech32, big.NewInt(0))

		// wait for tokens to be sent back if cross shard
		waitIfCrossShardProcessing(cs, esdtSafeAddrShard, receiverShardId)
		// still no tokens in esdt-safe, tokens should be burned
		if isSftOrMeta(receivedToken.Type) {
			chainSim.RequireAccountHasToken(t, cs, getTokenIdentifier(receivedToken), esdtSafeAddr, big.NewInt(1))
		} else {
			chainSim.RequireAccountHasToken(t, cs, getTokenIdentifier(receivedToken), esdtSafeAddr, big.NewInt(0))
		}
		tokenSupply, err := cs.GetNodeHandler(esdtSafeAddrShard).GetFacadeHandler().GetTokenSupply(getTokenIdentifier(receivedToken))
		require.Nil(t, err)
		require.NotNil(t, tokenSupply)
		require.Equal(t, receivedToken.Amount.String(), tokenSupply.Burned)

		nextShardId(&receiverShardId)
	}
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
	bridgeData := deploySovereignBridgeOnMainChain(t, cs, initialAddress)

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
		txResult = executeOperation(t, cs, bridgeData.OwnerAccount.Wallet, receiver.Bytes, &bridgeData.OwnerAccount.Nonce, bridgeData.ESDTSafeAddress, make([]chainSim.ArgsDepositToken, 0), wallet.Bytes, trnsData)
		chainSim.RequireSuccessfulTransaction(t, txResult)
	}
}

func waitIfCrossShardProcessing(cs chainSim.ChainSimulator, senderShard uint32, receivedShard uint32) {
	if senderShard != receivedShard {
		_ = cs.GenerateBlocks(3)
	}
}

func generateSovereignTokens() []chainSim.ArgsDepositToken {
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

	return tokens
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

func nextShardId(shardId *uint32) {
	*shardId++
	if *shardId > 2 {
		*shardId = 0
	}
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

func registerSovereignToken(
	t *testing.T,
	cs chainSim.ChainSimulator,
	wallet dtos.WalletAddress,
	nonce *uint64,
	esdtSafeAddress []byte,
	token chainSim.ArgsDepositToken,
) {
	esdtSafeAddr, _ := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Encode(esdtSafeAddress)
	esdtSafeAddrShard := chainSim.GetShardForAddress(cs, esdtSafeAddr)

	// deposit main chain -> sovereign chain
	// 0.05 EGLD-000000 which will be used for registerToken
	issueCost, _ := big.NewInt(0).SetString(issuePaymentCost, 10)
	egldPaymentToken := chainSim.ArgsDepositToken{
		Identifier: vmcommon.EGLDIdentifier,
		Nonce:      uint64(0),
		Amount:     issueCost,
	}
	txResult := deposit(t, cs, wallet.Bytes, nonce, esdtSafeAddress, []chainSim.ArgsDepositToken{egldPaymentToken}, wallet.Bytes, nil)
	chainSim.RequireSuccessfulTransaction(t, txResult)
	waitIfCrossShardProcessing(cs, esdtSafeAddrShard, chainSim.GetShardForAddress(cs, wallet.Bech32))
	chainSim.RequireBalance(t, cs, esdtSafeAddr, issueCost)

	registerToken(t, cs, wallet.Bytes, nonce, esdtSafeAddress, token)
}

func deployReceiverContractInAllShards(t *testing.T, cs chainSim.ChainSimulator) map[uint32]dtos.WalletAddress {
	nodeHandler := cs.GetNodeHandler(0)
	systemContractDeploy := chainSim.GetSysContactDeployAddressBytes(t, nodeHandler)

	receiverContracts := make(map[uint32]dtos.WalletAddress)
	for shardId := uint32(0); shardId <= nodeHandler.GetProcessComponents().ShardCoordinator().NumberOfShards(); shardId++ {
		wallet, _ := cs.GenerateAndMintWalletAddress(shardId, chainSim.InitialAmount)
		nonce := uint64(0)
		_ = cs.GenerateBlocks(1)

		contractAddress := chainSim.DeployContract(t, cs, wallet.Bytes, &nonce, systemContractDeploy, "", helloWasmPath)
		contractAddressBech32, _ := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Encode(contractAddress)
		receiverContracts[shardId] = dtos.WalletAddress{Bytes: contractAddress, Bech32: contractAddressBech32}
	}

	return receiverContracts
}

func isNft(esdtType core.ESDTType) bool {
	return esdtType == core.NonFungibleV2 ||
		esdtType == core.DynamicNFT
}

func isMeta(esdtType core.ESDTType) bool {
	return esdtType == core.MetaFungible ||
		esdtType == core.DynamicMeta
}

func isSftOrMeta(esdtType core.ESDTType) bool {
	return esdtType == core.SemiFungible ||
		esdtType == core.DynamicSFT ||
		isMeta(esdtType)
}
