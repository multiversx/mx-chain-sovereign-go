package bridge

import (
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	coreAPI "github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"

	sovereignChainSimulator "github.com/multiversx/mx-chain-go/cmd/sovereignnode/chainSimulator"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	chainSim "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
)

const (
	defaultPathToInitialConfig = "../../../../node/config/"
	sovereignConfigPath        = "../../../config/"
	esdtSafeWasmPath           = "../testdata/sov-esdt-safe.wasm"
	feeMarketWasmPath          = "../testdata/sov-fee-market.wasm"
	issuePaymentCost           = "50000000000000000"
)

// This test will:
// - deploy bridge contracts setup
// - set the native esdt token
// - deposit native token in esdt-safe contract
// - check the sender balance is correct
// - check the token burned amount is correct after deposit
func TestSovereignChainSimulator_DeployBridgeContractsAndDepositNativeESDTToken(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	nativeESDT := "SOV-1a2b3c"
	outGoingSubscribedAddress := "erd1qqqqqqqqqqqqqpgqmzzm05jeav6d5qvna0q2pmcllelkz8xddz3syjszx5"
	cs, err := sovereignChainSimulator.NewSovereignChainSimulator(sovereignChainSimulator.ArgsSovereignChainSimulator{
		SovereignConfigPath: sovereignConfigPath,
		ArgsChainSimulator: &chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck: true,
			TempDir:                t.TempDir(),
			PathToInitialConfig:    defaultPathToInitialConfig,
			GenesisTimestamp:       time.Now().Unix(),
			RoundDurationInMillis:  uint64(6000),
			RoundsPerEpoch:         core.OptionalUint64{},
			ApiInterface:           api.NewNoApiInterface(),
			MinNodesPerShard:       2,
			AlterConfigsFunction: func(cfg *config.Configs) {
				cfg.SystemSCConfig.ESDTSystemSCConfig.BaseIssuingCost = issuePaymentCost
				cfg.GeneralConfig.SovereignConfig.OutgoingSubscribedEvents.SubscribedEvents = []config.SubscribedEvent{
					{
						Identifier: "deposit",
						Addresses:  []string{outGoingSubscribedAddress},
					},
				}
				cfg.GeneralConfig.SovereignConfig.OutgoingSubscribedEvents.TimeToWaitForUnconfirmedOutGoingOperationInSeconds = 1
				cfg.GeneralConfig.VirtualMachine.Execution.TransferAndExecuteByUserAddresses = []string{outGoingSubscribedAddress}
				cfg.GeneralConfig.GeneralSettings.BaseTokenID = nativeESDT
			},
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	nodeHandler := cs.GetNodeHandler(core.SovereignChainShardId)

	initialAddress := "erd1l6xt0rqlyzw56a3k8xwwshq2dcjwy3q9cppucvqsmdyw8r98dz3sae0kxl"
	initialAddrBytes, err := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode(initialAddress)
	require.Nil(t, err)

	chainSim.InitAddressesAndSysAccState(t, cs, initialAddress)

	expectedESDTSafeAddressBytes, err := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode(outGoingSubscribedAddress)
	require.Nil(t, err)

	initialWallet := dtos.WalletAddress{Bech32: initialAddress, Bytes: initialAddrBytes}
	bridgeData := deploySovereignBridgeSetup(t, cs, initialWallet, esdtSafeWasmPath, feeMarketWasmPath)
	require.Equal(t, expectedESDTSafeAddressBytes, bridgeData.ESDTSafeAddress)

	wallet, err := cs.GenerateAndMintWalletAddress(core.SovereignChainShardId, chainSim.InitialAmount)
	require.Nil(t, err)
	nonce := uint64(0)
	_ = cs.GenerateBlocks(1)

	amountToDeposit, _ := big.NewInt(0).SetString("2000000000000000000", 10)
	depositTokens := make([]chainSim.ArgsDepositToken, 0)
	depositTokens = append(depositTokens, chainSim.ArgsDepositToken{
		Identifier: nativeESDT,
		Nonce:      0,
		Amount:     amountToDeposit,
	})

	txResult := Deposit(t, cs, wallet.Bytes, &nonce, bridgeData.ESDTSafeAddress, depositTokens, wallet.Bytes, nil)
	chainSim.RequireSuccessfulTransaction(t, txResult)

	txFee, _ := big.NewInt(0).SetString(txResult.Fee, 10)
	amountAfterFee := big.NewInt(0).Sub(chainSim.InitialAmount, txFee)

	nativeBalance, _, err := nodeHandler.GetFacadeHandler().GetBalance(wallet.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, nativeBalance)
	require.Equal(t, big.NewInt(0).Sub(amountAfterFee, amountToDeposit).String(), nativeBalance.String())

	tokenSupply, err := nodeHandler.GetFacadeHandler().GetTokenSupply(nativeESDT)
	require.Nil(t, err)
	require.NotNil(t, tokenSupply)
	require.Equal(t, amountToDeposit.String(), tokenSupply.Burned)

	// Wait for outgoing operations to get unconfirmed and check we have one, which is also saved in storage
	time.Sleep(time.Second)

	outGoingOps := nodeHandler.GetRunTypeComponents().OutGoingOperationsPoolHandler().GetUnconfirmedOperations()
	require.Len(t, outGoingOps, 1)
	require.Len(t, outGoingOps[0].OutGoingOperations, 1)
	checkOutGoingOperation(t, cs, outGoingOps[0].OutGoingOperations[0])
}

// This test will:
// - deploy bridge contracts setup
// - issue a new fungible token
// - deposit some tokens in esdt-safe contract
// - check the sender balance is correct
// - check the token burned amount is correct after deposit
func TestSovereignChainSimulator_DeployBridgeContractsThenIssueAndDeposit(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	outGoingSubscribedAddress := "erd1qqqqqqqqqqqqqpgqmzzm05jeav6d5qvna0q2pmcllelkz8xddz3syjszx5"
	cs, err := sovereignChainSimulator.NewSovereignChainSimulator(sovereignChainSimulator.ArgsSovereignChainSimulator{
		SovereignConfigPath: sovereignConfigPath,
		ArgsChainSimulator: &chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck: true,
			TempDir:                t.TempDir(),
			PathToInitialConfig:    defaultPathToInitialConfig,
			GenesisTimestamp:       time.Now().Unix(),
			RoundDurationInMillis:  uint64(6000),
			RoundsPerEpoch:         core.OptionalUint64{},
			ApiInterface:           api.NewNoApiInterface(),
			MinNodesPerShard:       2,
			AlterConfigsFunction: func(cfg *config.Configs) {
				cfg.SystemSCConfig.ESDTSystemSCConfig.BaseIssuingCost = issuePaymentCost
				cfg.GeneralConfig.SovereignConfig.OutgoingSubscribedEvents.SubscribedEvents = []config.SubscribedEvent{
					{
						Identifier: "deposit",
						Addresses:  []string{outGoingSubscribedAddress},
					},
				}
				cfg.GeneralConfig.SovereignConfig.OutgoingSubscribedEvents.TimeToWaitForUnconfirmedOutGoingOperationInSeconds = 1
			},
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	nodeHandler := cs.GetNodeHandler(core.SovereignChainShardId)

	initialAddress := "erd1l6xt0rqlyzw56a3k8xwwshq2dcjwy3q9cppucvqsmdyw8r98dz3sae0kxl"
	initialAddrBytes, err := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode(initialAddress)
	require.Nil(t, err)

	err = cs.SetStateMultiple([]*dtos.AddressState{
		{
			Address: initialAddress,
			Balance: "10000000000000000000000",
		},
	})
	require.Nil(t, err)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	wallet := dtos.WalletAddress{Bech32: initialAddress, Bytes: initialAddrBytes}

	expectedESDTSafeAddressBytes, err := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode(outGoingSubscribedAddress)
	require.Nil(t, err)

	bridgeData := deploySovereignBridgeSetup(t, cs, wallet, esdtSafeWasmPath, feeMarketWasmPath)
	require.Equal(t, expectedESDTSafeAddressBytes, bridgeData.ESDTSafeAddress)

	nonce := GetNonce(t, nodeHandler, wallet.Bech32)

	issueCost, _ := big.NewInt(0).SetString(issuePaymentCost, 10)
	supply, _ := big.NewInt(0).SetString("123000000000000000000", 10)
	tokenName := "SovToken"
	tokenTicker := "SVN"
	numDecimals := 18
	tokenIdentifier := chainSim.IssueFungible(t, cs, wallet.Bytes, &nonce, issueCost, tokenName, tokenTicker, numDecimals, supply)

	depositAndCheckTokens(t, cs, wallet, &nonce, bridgeData, tokenIdentifier, supply)

	// Wait for outgoing operations to get unconfirmed and check we have one, which is also saved in storage
	time.Sleep(time.Second)

	outGoingOps := nodeHandler.GetRunTypeComponents().OutGoingOperationsPoolHandler().GetUnconfirmedOperations()
	require.Len(t, outGoingOps, 1)
	require.Len(t, outGoingOps[0].OutGoingOperations, 1)
	checkOutGoingOperation(t, cs, outGoingOps[0].OutGoingOperations[0])
}

// This test will:
// - deploy bridge contracts setup
// - generate new wallet and set tokens without prefix (a.k.a main chain token)
// - deposit the token in esdt-safe contract
// - check the sender balance is correct
// - check the token burned amount is correct after deposit
func TestSovereignChainSimulator_DeployBridgeContractsAndDepositMainChainToken(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	outGoingSubscribedAddress := "erd1qqqqqqqqqqqqqpgqmzzm05jeav6d5qvna0q2pmcllelkz8xddz3syjszx5"
	cs, err := sovereignChainSimulator.NewSovereignChainSimulator(sovereignChainSimulator.ArgsSovereignChainSimulator{
		SovereignConfigPath: sovereignConfigPath,
		ArgsChainSimulator: &chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck: true,
			TempDir:                t.TempDir(),
			PathToInitialConfig:    defaultPathToInitialConfig,
			GenesisTimestamp:       time.Now().Unix(),
			RoundDurationInMillis:  uint64(6000),
			RoundsPerEpoch:         core.OptionalUint64{},
			ApiInterface:           api.NewNoApiInterface(),
			MinNodesPerShard:       2,
			AlterConfigsFunction: func(cfg *config.Configs) {
				cfg.GeneralConfig.SovereignConfig.OutgoingSubscribedEvents.SubscribedEvents = []config.SubscribedEvent{
					{
						Identifier: "deposit",
						Addresses:  []string{outGoingSubscribedAddress},
					},
				}
				cfg.GeneralConfig.SovereignConfig.OutgoingSubscribedEvents.TimeToWaitForUnconfirmedOutGoingOperationInSeconds = 1
				cfg.GeneralConfig.VirtualMachine.Execution.TransferAndExecuteByUserAddresses = []string{outGoingSubscribedAddress}
				cfg.GeneralConfig.GeneralSettings.BaseTokenID = "WEGLD-1a2b3c"
			},
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	nodeHandler := cs.GetNodeHandler(core.SovereignChainShardId)

	initialAddress := "erd1l6xt0rqlyzw56a3k8xwwshq2dcjwy3q9cppucvqsmdyw8r98dz3sae0kxl"
	initialAddrBytes, err := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode(initialAddress)
	require.Nil(t, err)

	chainSim.InitAddressesAndSysAccState(t, cs, initialAddress)

	expectedESDTSafeAddressBytes, err := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode(outGoingSubscribedAddress)
	require.Nil(t, err)

	initialWallet := dtos.WalletAddress{Bech32: initialAddress, Bytes: initialAddrBytes}
	bridgeData := deploySovereignBridgeSetup(t, cs, initialWallet, esdtSafeWasmPath, feeMarketWasmPath)
	require.Equal(t, expectedESDTSafeAddressBytes, bridgeData.ESDTSafeAddress)

	// Generate new wallet
	wallet, err := cs.GenerateAndMintWalletAddress(core.SovereignChainShardId, chainSim.InitialAmount)
	require.Nil(t, err)
	nonce := uint64(0)

	depositMainChainToken(t, cs, bridgeData, wallet, &nonce, "MAIN-1a2b3c")
	depositMainChainToken(t, cs, bridgeData, wallet, &nonce, vmcommon.EGLDIdentifier)
}

func depositMainChainToken(
	t *testing.T,
	cs chainSim.ChainSimulator,
	bridgeData ArgsBridgeSetup,
	wallet dtos.WalletAddress,
	nonce *uint64,
	mainChainToken string,
) {
	// Set main chain token supply in wallet
	mainChainTokenSupply, _ := big.NewInt(0).SetString("12000000000000000000", 10)
	chainSim.SetEsdtInWallet(t, cs, wallet, mainChainToken, 0, esdt.ESDigitalToken{Value: mainChainTokenSupply})

	depositAndCheckTokens(t, cs, wallet, nonce, bridgeData, mainChainToken, mainChainTokenSupply)

	// Wait for outgoing operations to get unconfirmed and check we have one, which is also saved in storage
	time.Sleep(time.Second)

	nodeHandler := cs.GetNodeHandler(core.SovereignChainShardId)
	outGoingOps := nodeHandler.GetRunTypeComponents().OutGoingOperationsPoolHandler().GetUnconfirmedOperations()
	require.Len(t, outGoingOps, 1)
	require.Len(t, outGoingOps[0].OutGoingOperations, 1)
	checkOutGoingOperation(t, cs, outGoingOps[0].OutGoingOperations[0])
	confirmOutgoingOperations(t, cs, outGoingOps)
}

func depositAndCheckTokens(
	t *testing.T,
	cs chainSim.ChainSimulator,
	wallet dtos.WalletAddress,
	nonce *uint64,
	bridgeData ArgsBridgeSetup,
	mainChainToken string,
	mainChainTokenSupply *big.Int,
) {
	nodeHandler := cs.GetNodeHandler(core.SovereignChainShardId)

	amountToDeposit, _ := big.NewInt(0).SetString("2000000000000000000", 10)
	depositTokens := make([]chainSim.ArgsDepositToken, 0)
	depositTokens = append(depositTokens, chainSim.ArgsDepositToken{
		Identifier: mainChainToken,
		Nonce:      0,
		Amount:     amountToDeposit,
	})

	txResult := Deposit(t, cs, wallet.Bytes, nonce, bridgeData.ESDTSafeAddress, depositTokens, wallet.Bytes, nil)
	chainSim.RequireSuccessfulTransaction(t, txResult)

	tokens, _, err := nodeHandler.GetFacadeHandler().GetAllESDTTokens(wallet.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, tokens)
	require.Equal(t, big.NewInt(0).Sub(mainChainTokenSupply, amountToDeposit).String(), tokens[mainChainToken].GetValue().String())

	tokenSupply, err := nodeHandler.GetFacadeHandler().GetTokenSupply(mainChainToken)
	require.Nil(t, err)
	require.NotNil(t, tokenSupply)
	require.Equal(t, amountToDeposit.String(), tokenSupply.Burned)
}

func checkOutGoingOperation(t *testing.T, cs chainSim.ChainSimulator, outGoingOp *sovereign.OutGoingOperation) {
	nodeHandler := cs.GetNodeHandler(core.SovereignChainShardId)

	savedMarshalledTx, err := nodeHandler.GetDataComponents().StorageService().Get(dataRetriever.TransactionUnit, outGoingOp.Hash)
	require.Nil(t, err)
	require.NotNil(t, savedMarshalledTx)

	savedTx := &transaction.Transaction{}
	err = nodeHandler.GetCoreComponents().InternalMarshalizer().Unmarshal(savedTx, savedMarshalledTx)
	require.Nil(t, err)

	expectedSavedTx := &transaction.Transaction{
		GasPrice: nodeHandler.GetCoreComponents().EconomicsData().MinGasPrice(),
		GasLimit: nodeHandler.GetCoreComponents().EconomicsData().ComputeGasLimit(
			&transaction.Transaction{
				Data: outGoingOp.Data,
			}),
		Data: outGoingOp.Data,
	}
	require.Equal(t, expectedSavedTx, savedTx)

	// Generate extra blocks after outgoing operations are created
	err = cs.GenerateBlocks(10)
	require.Nil(t, err)
}

func confirmOutgoingOperations(t *testing.T, cs chainSim.ChainSimulator, outGoingOps []*sovereign.BridgeOutGoingData) {
	nodeHandler := cs.GetNodeHandler(core.SovereignChainShardId)
	require.True(t, len(outGoingOps) > 0)

	for _, outGoingOp := range outGoingOps {
		for _, operation := range outGoingOp.OutGoingOperations {
			err := nodeHandler.GetRunTypeComponents().OutGoingOperationsPoolHandler().ConfirmOperation(outGoingOp.Hash, operation.Hash)
			require.NoError(t, err)
		}
	}
}

func TestSovereignChainSimulator_DepositNoPaymentWithTransferData(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	outGoingSubscribedAddress := "erd1qqqqqqqqqqqqqpgqmzzm05jeav6d5qvna0q2pmcllelkz8xddz3syjszx5"
	cs, err := sovereignChainSimulator.NewSovereignChainSimulator(sovereignChainSimulator.ArgsSovereignChainSimulator{
		SovereignConfigPath: sovereignConfigPath,
		ArgsChainSimulator: &chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck: true,
			TempDir:                t.TempDir(),
			PathToInitialConfig:    defaultPathToInitialConfig,
			GenesisTimestamp:       time.Now().Unix(),
			RoundDurationInMillis:  uint64(6000),
			RoundsPerEpoch:         core.OptionalUint64{},
			ApiInterface:           api.NewNoApiInterface(),
			MinNodesPerShard:       2,
			AlterConfigsFunction: func(cfg *config.Configs) {
				cfg.GeneralConfig.SovereignConfig.OutgoingSubscribedEvents.SubscribedEvents = []config.SubscribedEvent{
					{
						Identifier: "deposit",
						Addresses:  []string{outGoingSubscribedAddress},
					},
				}
				cfg.GeneralConfig.SovereignConfig.OutgoingSubscribedEvents.TimeToWaitForUnconfirmedOutGoingOperationInSeconds = 1
			},
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	nodeHandler := cs.GetNodeHandler(core.SovereignChainShardId)

	wallet, err := cs.GenerateAndMintWalletAddress(core.SovereignChainShardId, chainSim.InitialAmount)
	require.Nil(t, err)
	nonce := uint64(0)

	initialAddress := "erd1l6xt0rqlyzw56a3k8xwwshq2dcjwy3q9cppucvqsmdyw8r98dz3sae0kxl"
	initialAddrBytes, err := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode(initialAddress)
	require.Nil(t, err)
	chainSim.InitAddressesAndSysAccState(t, cs, initialAddress)

	expectedESDTSafeAddressBytes, err := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode(outGoingSubscribedAddress)
	require.Nil(t, err)

	initialWallet := dtos.WalletAddress{Bech32: initialAddress, Bytes: initialAddrBytes}
	bridgeData := deploySovereignBridgeSetup(t, cs, initialWallet, esdtSafeWasmPath, feeMarketWasmPath)
	require.Equal(t, expectedESDTSafeAddressBytes, bridgeData.ESDTSafeAddress)

	txResult := Deposit(t, cs, wallet.Bytes, &nonce, bridgeData.ESDTSafeAddress, make([]chainSim.ArgsDepositToken, 0), wallet.Bytes, nil)
	chainSim.RequireSignalError(t, txResult, "Nothing to transfer")

	trnsData := &sovereign.TransferData{
		GasLimit: uint64(10000000),
		Function: []byte("hello"),
		Args:     [][]byte{{0x01}},
	}
	txResult = Deposit(t, cs, wallet.Bytes, &nonce, bridgeData.ESDTSafeAddress, make([]chainSim.ArgsDepositToken, 0), wallet.Bytes, trnsData)
	chainSim.RequireSuccessfulTransaction(t, txResult)

	// Wait for outgoing operations to get unconfirmed and check we have one, which is also saved in storage
	time.Sleep(time.Second)

	outGoingOps := nodeHandler.GetRunTypeComponents().OutGoingOperationsPoolHandler().GetUnconfirmedOperations()
	require.Len(t, outGoingOps, 1)
	require.Len(t, outGoingOps[0].OutGoingOperations, 1)
	checkOutGoingOperation(t, cs, outGoingOps[0].OutGoingOperations[0])
}

// This test will:
// - deploy bridge contracts setup
// - issue and register a new fungible token
// - deposit the token in esdt-safe contract
// - check the sender balance is correct
// - check the token burned amount is correct after deposit
func TestSovereignChainSimulator_DeployBridgeContractsThenRegisterTokenAndDeposit(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	outGoingSubscribedAddress := "erd1qqqqqqqqqqqqqpgqmzzm05jeav6d5qvna0q2pmcllelkz8xddz3syjszx5"
	cs, err := sovereignChainSimulator.NewSovereignChainSimulator(sovereignChainSimulator.ArgsSovereignChainSimulator{
		SovereignConfigPath: sovereignConfigPath,
		ArgsChainSimulator: &chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck: true,
			TempDir:                t.TempDir(),
			PathToInitialConfig:    defaultPathToInitialConfig,
			GenesisTimestamp:       time.Now().Unix(),
			RoundDurationInMillis:  uint64(6000),
			RoundsPerEpoch:         core.OptionalUint64{},
			ApiInterface:           api.NewNoApiInterface(),
			MinNodesPerShard:       2,
			AlterConfigsFunction: func(cfg *config.Configs) {
				cfg.SystemSCConfig.ESDTSystemSCConfig.BaseIssuingCost = issuePaymentCost
				cfg.GeneralConfig.SovereignConfig.OutgoingSubscribedEvents.SubscribedEvents = []config.SubscribedEvent{
					{
						Identifier: "deposit",
						Addresses:  []string{outGoingSubscribedAddress},
					},
				}
				cfg.GeneralConfig.SovereignConfig.OutgoingSubscribedEvents.TimeToWaitForUnconfirmedOutGoingOperationInSeconds = 1
				cfg.GeneralConfig.VirtualMachine.Execution.TransferAndExecuteByUserAddresses = []string{outGoingSubscribedAddress}
			},
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	nodeHandler := cs.GetNodeHandler(core.SovereignChainShardId)

	initialAddress := "erd1l6xt0rqlyzw56a3k8xwwshq2dcjwy3q9cppucvqsmdyw8r98dz3sae0kxl"
	initialAddrBytes, err := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode(initialAddress)
	require.Nil(t, err)

	chainSim.InitAddressesAndSysAccState(t, cs, initialAddress)
	initialWallet := dtos.WalletAddress{Bech32: initialAddress, Bytes: initialAddrBytes}
	bridgeData := deploySovereignBridgeSetup(t, cs, initialWallet, esdtSafeWasmPath, feeMarketWasmPath)

	wallet, err := cs.GenerateAndMintWalletAddress(core.SovereignChainShardId, chainSim.InitialAmount)
	require.Nil(t, err)
	nonce := uint64(0)
	_ = cs.GenerateBlocks(1)

	tokenIdentifier, supply := issueAndRegisterToken(t, cs, wallet, &nonce, bridgeData.ESDTSafeAddress)

	depositAndCheckTokens(t, cs, wallet, &nonce, bridgeData, tokenIdentifier, supply)

	// Wait for outgoing operations to get unconfirmed and check we have one, which is also saved in storage
	time.Sleep(time.Second)

	outGoingOps := nodeHandler.GetRunTypeComponents().OutGoingOperationsPoolHandler().GetUnconfirmedOperations()
	require.Len(t, outGoingOps, 1)
	require.Len(t, outGoingOps[0].OutGoingOperations, 1)
	checkOutGoingOperation(t, cs, outGoingOps[0].OutGoingOperations[0])
}

func issueAndRegisterToken(
	t *testing.T,
	cs chainSim.ChainSimulator,
	wallet dtos.WalletAddress,
	nonce *uint64,
	esdtSafeAddress []byte,
) (string, *big.Int) {
	// Set EGLD-000000 supply in wallet
	egldSupply, _ := big.NewInt(0).SetString("12000000000000000000", 10)
	chainSim.SetEsdtInWallet(t, cs, wallet, vmcommon.EGLDIdentifier, 0, esdt.ESDigitalToken{Value: egldSupply})
	_ = cs.GenerateBlocks(1)

	// Issue new fungible token
	issueCost, _ := big.NewInt(0).SetString(issuePaymentCost, 10)
	supply, _ := big.NewInt(0).SetString("123000000000000000000", 10)
	tokenName := "SovToken"
	tokenTicker := "SVN"
	numDecimals := 18
	tokenIdentifier := chainSim.IssueFungible(t, cs, wallet.Bytes, nonce, issueCost, tokenName, tokenTicker, numDecimals, supply)

	// Register the issued token
	registerTokenArgs := [][]byte{
		[]byte("registerToken"),
		[]byte(tokenIdentifier),
		[]byte{0x00},
		[]byte(tokenName),
		[]byte(tokenTicker),
		big.NewInt(int64(numDecimals)).Bytes(),
	}
	paymentTokens := make([]chainSim.ArgsDepositToken, 0)
	paymentTokens = append(paymentTokens, chainSim.ArgsDepositToken{
		Identifier: vmcommon.EGLDIdentifier,
		Nonce:      uint64(0),
		Amount:     issueCost,
	})
	// Call registerToken with 0.05 EGLD-000000 as payment
	chainSim.TransferMultiESDTNFT(t, cs, wallet.Bytes, esdtSafeAddress, nonce, paymentTokens, registerTokenArgs...)

	return tokenIdentifier, supply
}
