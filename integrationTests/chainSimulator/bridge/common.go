package bridge

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/stretchr/testify/require"

	chainSim "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	"github.com/multiversx/mx-chain-go/process"
)

const (
	issuePaymentCost         = "50000000000000000"
	esdtSafeWasmPath         = "testdata/mvx-esdt-safe.wasm"
	enshrineEsdtSafeWasmPath = "testdata/enshrine-esdt-safe.wasm"
	//enshrine esdt-safe contract without checks for prefix or issue cost paid for new tokens
	simpleEnshrineEsdtSafeWasmPath = "testdata/simple-enshrine-esdt-safe.wasm"
	feeMarketWasmPath              = "testdata/mvx-fee-market.wasm"
	sovereignForgeWasmPath         = "testdata/sovereign-forge.wasm"
	chainFactoryWasmPath           = "testdata/chain-factory.wasm"
	chainConfigWasmPath            = "testdata/chain-config.wasm"
	headerVerifierWasmPath         = "testdata/header-verifier.wasm"

	sovereignForgeShardID = 1
	chainConfigIndex      = 6
	esdtSafeIndex         = 3
	feeMarketIndex        = 4
	headerVerifierIndex   = 2

	sovChainID = "sov1"

	depositFunc       = "deposit"
	registerTokenFunc = "registerToken"
)

// ArgsEsdtSafe holds the arguments for esdt safe contract argument
type ArgsEsdtSafe struct {
	ChainPrefix       string
	IssuePaymentToken string
}

// ArgsBridgeSetup holds the arguments for bridge setup
type ArgsBridgeSetup struct {
	SovereignForgeAddress []byte
	ChainConfigAddress    []byte
	HeaderVerifierAddress []byte
	ESDTSafeAddress       []byte
	FeeMarketAddress      []byte
	OwnerAccount          chainSim.Account
	RegisteredBLSKeys     []string
	NativeESDT            string
}

type transferData struct {
	GasLimit uint64
	Function []byte
	Args     [][]byte
}

func initOwnerAndSysAccState(
	t *testing.T,
	cs chainSim.ChainSimulator,
	ownerAddress string,
	argsEsdtSafe ArgsEsdtSafe,
) {
	chainSim.InitAddressesAndSysAccState(t, cs, ownerAddress)

	tokenKey := hex.EncodeToString([]byte(core.ProtectedKeyPrefix + core.ESDTKeyIdentifier + argsEsdtSafe.IssuePaymentToken))
	err := cs.SetKeyValueForAddress(chainSim.ESDTSystemAccount,
		map[string]string{
			tokenKey: "0400",
		},
	)
	require.Nil(t, err)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)
}

type esdtSafeDeployFunc func(
	t *testing.T,
	cs chainSim.ChainSimulator,
	ownerAddress []byte,
	nonce *uint64,
	systemContractDeploy []byte,
	argsEsdtSafe ArgsEsdtSafe,
	contractWasmPath string,
) []byte

// This function will:
// - deploy esdt-safe contract
// - deploy fee-market contract
// - set the fee-market address inside esdt-safe contract
// - disable fee in fee-market contract
// - unpause esdt-safe contract so deposit operations can start
func deployBridgeSetup(
	t *testing.T,
	cs chainSim.ChainSimulator,
	ownerAddress string,
	argsEsdtSafe ArgsEsdtSafe,
	deployEsdtSafeContract esdtSafeDeployFunc,
	contractWasmPath string,
) *ArgsBridgeSetup {
	nodeHandler := cs.GetNodeHandler(0)

	systemContractDeploy := chainSim.GetSysContactDeployAddressBytes(t, nodeHandler)
	ownerAddrBytes, err := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode(ownerAddress)
	require.Nil(t, err)
	nonce := uint64(0)

	esdtSafeAddress := deployEsdtSafeContract(t, cs, ownerAddrBytes, &nonce, systemContractDeploy, argsEsdtSafe, contractWasmPath)

	feeMarketArgs := "@" + hex.EncodeToString(esdtSafeAddress) + // esdt_safe_address
		"@00" // no fee
	feeMarketAddress := chainSim.DeployContract(t, cs, ownerAddrBytes, &nonce, systemContractDeploy, feeMarketArgs, feeMarketWasmPath)

	setFeeMarketAddressData := "setFeeMarketAddress" +
		"@" + hex.EncodeToString(feeMarketAddress)
	chainSim.SendTransactionWithSuccess(t, cs, ownerAddrBytes, &nonce, esdtSafeAddress, chainSim.ZeroValue, setFeeMarketAddressData, uint64(10000000))

	chainSim.SendTransactionWithSuccess(t, cs, ownerAddrBytes, &nonce, esdtSafeAddress, chainSim.ZeroValue, "unpause", uint64(10000000))

	return &ArgsBridgeSetup{
		ESDTSafeAddress:  esdtSafeAddress,
		FeeMarketAddress: feeMarketAddress,
		OwnerAccount: chainSim.Account{
			Wallet: dtos.WalletAddress{Bech32: ownerAddress, Bytes: ownerAddrBytes},
			Nonce:  nonce,
		},
	}
}

func enshrineEsdtSafeContract(
	t *testing.T,
	cs chainSim.ChainSimulator,
	ownerAddress []byte,
	nonce *uint64,
	systemContractDeploy []byte,
	argsEsdtSafe ArgsEsdtSafe,
	contractWasmPath string,
) []byte {
	esdtSafeArgs := "@" + // is_sovereign_chain
		"@01" + lengthOn4Bytes(len(argsEsdtSafe.IssuePaymentToken)) + hex.EncodeToString([]byte(argsEsdtSafe.IssuePaymentToken)) + // token identifier
		"@01" + lengthOn4Bytes(len(argsEsdtSafe.ChainPrefix)) + hex.EncodeToString([]byte(argsEsdtSafe.ChainPrefix)) // prefix
	return chainSim.DeployContract(t, cs, ownerAddress, nonce, systemContractDeploy, esdtSafeArgs, contractWasmPath)
}

// This function will:
// - deploy the full sovereign bridge contracts setup
// - unpause esdt-safe contract so deposit operations can start
func deploySovereignBridgeSetup(
	t *testing.T,
	cs chainSim.ChainSimulator,
	ownerAddress string,
) *ArgsBridgeSetup {
	nodeHandler := cs.GetNodeHandler(0)

	systemContractDeploy := chainSim.GetSysContactDeployAddressBytes(t, nodeHandler)

	sovereignForgeAddress := deploySovereignSCSetup(t, cs, systemContractDeploy)

	ownerAddrBytes, err := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode(ownerAddress)
	require.Nil(t, err)
	nonce := uint64(0)

	chainConfigAddress := deployPhaseOne(t, cs, sovereignForgeAddress, ownerAddrBytes, &nonce, sovChainID)

	esdtSafeAddress, nativeESDT := deployPhaseTwo(t, cs, sovereignForgeAddress, ownerAddrBytes, &nonce, sovChainID)

	feeMarketAddress := deployPhaseThree(t, cs, sovereignForgeAddress, ownerAddrBytes, &nonce, sovChainID)

	headerVerifierAddress := deployPhaseFour(t, cs, sovereignForgeAddress, ownerAddrBytes, &nonce, sovChainID)

	_, blsKeys, err := chainSimulator.GenerateBlsPrivateKeys(2)
	require.Nil(t, err)
	for _, key := range blsKeys {
		registerArgs := "register" +
			"@" + key
		chainSim.SendTransactionWithSuccess(t, cs, ownerAddrBytes, &nonce, chainConfigAddress, chainSim.ZeroValue, registerArgs, uint64(10_000_000))
	}

	chainSim.SendTransactionWithSuccess(t, cs, ownerAddrBytes, &nonce, sovereignForgeAddress, chainSim.ZeroValue, "completeSetupPhase", uint64(70_000_000))

	return &ArgsBridgeSetup{
		SovereignForgeAddress: sovereignForgeAddress,
		ChainConfigAddress:    chainConfigAddress,
		HeaderVerifierAddress: headerVerifierAddress,
		ESDTSafeAddress:       esdtSafeAddress,
		FeeMarketAddress:      feeMarketAddress,
		OwnerAccount: chainSim.Account{
			Wallet: dtos.WalletAddress{Bech32: ownerAddress, Bytes: ownerAddrBytes},
			Nonce:  nonce,
		},
		RegisteredBLSKeys: blsKeys,
		NativeESDT:        nativeESDT,
	}
}

func deploySovereignSCSetup(
	t *testing.T,
	cs chainSim.ChainSimulator,
	systemContractDeploy []byte,
) []byte {
	owner, _ := cs.GenerateAndMintWalletAddress(sovereignForgeShardID, chainSim.InitialAmount)
	ownerNonce := uint64(0)
	_ = cs.GenerateBlocks(1)
	sovereignForgeAddress := chainSim.DeployContract(t, cs, owner.Bytes, &ownerNonce, systemContractDeploy, "", sovereignForgeWasmPath)

	for shardId := uint32(0); shardId < cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().NumberOfShards(); shardId++ {
		wallet, _ := cs.GenerateAndMintWalletAddress(shardId, chainSim.InitialAmount)
		nonce := uint64(0)
		_ = cs.GenerateBlocks(1)

		chainConfigTemplateAddress := chainSim.DeployContract(t, cs, wallet.Bytes, &nonce, systemContractDeploy, "", chainConfigWasmPath)
		esdtSafeTemplateAddress := chainSim.DeployContract(t, cs, wallet.Bytes, &nonce, systemContractDeploy, "@"+hex.EncodeToString(wallet.Bytes)+"@31", esdtSafeWasmPath)              // random prefix
		feeMarketTemplateAddress := chainSim.DeployContract(t, cs, wallet.Bytes, &nonce, systemContractDeploy, "@"+hex.EncodeToString(esdtSafeTemplateAddress)+"@00", feeMarketWasmPath) // no fee
		headerVerifierTemplateAddress := chainSim.DeployContract(t, cs, wallet.Bytes, &nonce, systemContractDeploy, "", headerVerifierWasmPath)

		chainFactoryArgs := "@" + hex.EncodeToString(sovereignForgeAddress) +
			"@" + hex.EncodeToString(chainConfigTemplateAddress) +
			"@" + hex.EncodeToString(headerVerifierTemplateAddress) +
			"@" + hex.EncodeToString(esdtSafeTemplateAddress) +
			"@" + hex.EncodeToString(feeMarketTemplateAddress)
		chainFactoryAddress := chainSim.DeployContract(t, cs, wallet.Bytes, &nonce, systemContractDeploy, chainFactoryArgs, chainFactoryWasmPath)

		registerChainFactoryArgs := "registerChainFactory" +
			"@" + hex.EncodeToString(big.NewInt(int64(shardId)).Bytes()) +
			"@" + hex.EncodeToString(chainFactoryAddress)
		chainSim.SendTransactionWithSuccess(t, cs, owner.Bytes, &ownerNonce, sovereignForgeAddress, chainSim.ZeroValue, registerChainFactoryArgs, uint64(30_000_000))
	}

	return sovereignForgeAddress
}

func deployPhaseOne(
	t *testing.T,
	cs chainSim.ChainSimulator,
	contractAddress []byte,
	wallet []byte,
	nonce *uint64,
	sovChainID string,
) []byte {
	phaseOneArgs := "deployPhaseOne" +
		"@01" + lengthOn4Bytes(len(sovChainID)) + hex.EncodeToString([]byte(sovChainID))
	chainSim.SendTransactionWithSuccess(t, cs, wallet, nonce, contractAddress, chainSim.ZeroValue, phaseOneArgs, uint64(25_000_000))
	return readContractAddress(t, cs, contractAddress, sovChainID, chainConfigIndex)
}

func deployPhaseTwo(
	t *testing.T,
	cs chainSim.ChainSimulator,
	contractAddress []byte,
	wallet []byte,
	nonce *uint64,
	sovChainID string,
) ([]byte, string) {
	phaseTwoArgs := "deployPhaseTwo"
	chainSim.SendTransactionWithSuccess(t, cs, wallet, nonce, contractAddress, chainSim.ZeroValue, phaseTwoArgs, uint64(30_000_000))
	esdtSafeAddress := readContractAddress(t, cs, contractAddress, sovChainID, esdtSafeIndex)

	nativeTokenTicker := "SOV"
	nativeTokenName := "SovToken"
	issueCost, _ := big.NewInt(0).SetString(issuePaymentCost, 10)
	registerNativeTokenArgs := "registerNativeToken" +
		"@" + hex.EncodeToString([]byte(nativeTokenTicker)) +
		"@" + hex.EncodeToString([]byte(nativeTokenName))
	chainSim.SendTransactionWithSuccess(t, cs, wallet, nonce, esdtSafeAddress, issueCost, registerNativeTokenArgs, uint64(80_000_000))
	_ = cs.GenerateBlocks(2)
	nativeESDT := readNativeESDT(t, cs, esdtSafeAddress)

	return esdtSafeAddress, nativeESDT
}

func deployPhaseThree(
	t *testing.T,
	cs chainSim.ChainSimulator,
	contractAddress []byte,
	wallet []byte,
	nonce *uint64,
	sovChainID string,
) []byte {
	phaseThreeArgs := "deployPhaseThree" +
		"@00"
	chainSim.SendTransactionWithSuccess(t, cs, wallet, nonce, contractAddress, chainSim.ZeroValue, phaseThreeArgs, uint64(30_000_000))
	return readContractAddress(t, cs, contractAddress, sovChainID, feeMarketIndex)
}

func deployPhaseFour(
	t *testing.T,
	cs chainSim.ChainSimulator,
	contractAddress []byte,
	wallet []byte,
	nonce *uint64,
	sovChainID string,
) []byte {
	phaseFourArgs := "deployPhaseFour"
	chainSim.SendTransactionWithSuccess(t, cs, wallet, nonce, contractAddress, chainSim.ZeroValue, phaseFourArgs, uint64(25_000_000))
	return readContractAddress(t, cs, contractAddress, sovChainID, headerVerifierIndex)
}

func readContractAddress(
	t *testing.T,
	cs chainSim.ChainSimulator,
	sovereignForgeAddress []byte,
	sovChainID string,
	index int,
) []byte {
	res, _, err := cs.GetNodeHandler(sovereignForgeShardID).GetFacadeHandler().ExecuteSCQuery(&process.SCQuery{
		ScAddress: sovereignForgeAddress,
		FuncName:  "getDeployedSovereignContracts",
		Arguments: [][]byte{[]byte(sovChainID)},
	})
	require.NoError(t, err)
	require.Equal(t, chainSim.OkReturnCode, res.ReturnCode)

	return getAddressAtIndex(t, index, res.ReturnData)
}

func getAddressAtIndex(
	t *testing.T,
	index int,
	data [][]byte,
) []byte {
	for _, dt := range data {
		if int(dt[0]) == index {
			addr := make([]byte, 32)
			copy(addr, dt[1:33])
			return addr
		}
	}

	require.Fail(t, "Address not found at index", index)
	return nil
}

func readNativeESDT(
	t *testing.T,
	cs chainSim.ChainSimulator,
	esdtSafeAddress []byte,
) string {
	res, _, err := cs.GetNodeHandler(1).GetFacadeHandler().ExecuteSCQuery(&process.SCQuery{
		ScAddress: esdtSafeAddress,
		FuncName:  "getNativeToken",
	})
	require.NoError(t, err)
	require.Equal(t, chainSim.OkReturnCode, res.ReturnCode)

	return string(res.ReturnData[0])
}

// deposit will deposit tokens in the bridge sc safe contract
func deposit(
	t *testing.T,
	cs chainSim.ChainSimulator,
	sender []byte,
	nonce *uint64,
	contract []byte,
	tokens []chainSim.ArgsDepositToken,
	receiver []byte,
	transferData *transferData,
) *transaction.ApiTransactionResult {
	if len(tokens) == 0 {
		return depositScCall(t, cs, sender, nonce, contract, receiver, transferData)
	}

	depositArgs := core.BuiltInFunctionMultiESDTNFTTransfer +
		"@" + hex.EncodeToString(contract) +
		"@" + fmt.Sprintf("%02X", len(tokens))

	for _, token := range tokens {
		depositArgs = depositArgs +
			"@" + hex.EncodeToString([]byte(token.Identifier)) +
			"@" + getTokenNonce(token.Nonce) +
			"@" + hex.EncodeToString(token.Amount.Bytes())
	}

	depositArgs = depositArgs +
		"@" + hex.EncodeToString([]byte(depositFunc)) +
		"@" + hex.EncodeToString(receiver) +
		getDepositTransferDataArgs(transferData)

	return chainSim.SendTransaction(t, cs, sender, nonce, sender, chainSim.ZeroValue, depositArgs, uint64(20000000))
}

// depositScCall will make a smart contract call through deposit endpoint
func depositScCall(t *testing.T,
	cs chainSim.ChainSimulator,
	sender []byte,
	nonce *uint64,
	contract []byte,
	receiver []byte,
	transferData *transferData,
) *transaction.ApiTransactionResult {
	depositArgs := depositFunc +
		"@" + hex.EncodeToString(receiver) +
		getDepositTransferDataArgs(transferData)

	return chainSim.SendTransaction(t, cs, sender, nonce, contract, chainSim.ZeroValue, depositArgs, uint64(20000000))
}

func getDepositTransferDataArgs(transferData *transferData) string {
	if transferData == nil {
		return ""
	}

	args := ""
	for _, arg := range transferData.Args {
		args = args + "@" +
			hex.EncodeToString(arg)
	}

	return "@" + getUint64Bytes(transferData.GasLimit) +
		"@" + hex.EncodeToString(transferData.Function) +
		args
}

func registerSovereignNewTokens(
	t *testing.T,
	cs chainSim.ChainSimulator,
	wallet dtos.WalletAddress,
	nonce *uint64,
	esdtSafeAddress []byte,
	issuePaymentToken string,
	tokens []string,
) {
	lenSovTokens := big.NewInt(int64(len(tokens)))
	issueCostOneToken, _ := big.NewInt(0).SetString(issuePaymentCost, 10)
	issueTotalCost := big.NewInt(0).Mul(issueCostOneToken, lenSovTokens)

	registerData := core.BuiltInFunctionESDTTransfer +
		"@" + hex.EncodeToString([]byte(issuePaymentToken)) +
		"@" + hex.EncodeToString(issueTotalCost.Bytes()) +
		"@" + hex.EncodeToString([]byte("registerNewTokenID"))
	for _, token := range tokens {
		registerData = registerData +
			"@" + hex.EncodeToString([]byte(token))
	}
	txResult := chainSim.SendTransaction(t, cs, wallet.Bytes, nonce, esdtSafeAddress, chainSim.ZeroValue, registerData, uint64(10000000))
	chainSim.RequireSuccessfulTransaction(t, txResult)
}

func executeOperation(
	t *testing.T,
	cs chainSim.ChainSimulator,
	wallet dtos.WalletAddress,
	receiver []byte,
	nonce *uint64,
	esdtSafeAddress []byte,
	bridgedInTokens []chainSim.ArgsDepositToken,
	originalSender []byte,
	transferData *transferData,
) *transaction.ApiTransactionResult {
	executeBridgeOpsData := "executeBridgeOps" +
		"@" + generateRandomHash() + //dummy hash
		"@" + // operation
		hex.EncodeToString(receiver) + // receiver address
		lengthOn4Bytes(len(bridgedInTokens)) + // nr of tokens
		getTokenDataArgs(wallet.Bytes, bridgedInTokens) + // tokens encoded arg
		getUint64Bytes(0) + // event nonce
		hex.EncodeToString(originalSender) + // sender address from other chain
		getTransferDataArgs(transferData)
	return chainSim.SendTransaction(t, cs, wallet.Bytes, nonce, esdtSafeAddress, chainSim.ZeroValue, executeBridgeOpsData, uint64(100000000))
}

func generateRandomHash() string {
	randomBytes := make([]byte, 32)
	_, _ = rand.Read(randomBytes)
	return hex.EncodeToString(randomBytes)
}

func getTokenDataArgs(creator []byte, tokens []chainSim.ArgsDepositToken) string {
	var arg string
	for _, token := range tokens {
		arg = arg +
			lengthOn4Bytes(len(token.Identifier)) + // length of token identifier
			hex.EncodeToString([]byte(token.Identifier)) + //token identifier
			getUint64Bytes(token.Nonce) + // nonce
			fmt.Sprintf("%02x", uint32(token.Type)) + // type
			lengthOn4Bytes(len(token.Amount.Bytes())) + // length of amount
			hex.EncodeToString(token.Amount.Bytes()) + // amount
			"00" + // not frozen
			lengthOn4Bytes(0) + // length of hash
			lengthOn4Bytes(4) + // length of name
			hex.EncodeToString([]byte("ESDT")) + // name
			lengthOn4Bytes(0) + // length of attributes
			hex.EncodeToString(creator) + // creator
			lengthOn4Bytes(0) + //length of royalties
			lengthOn4Bytes(0) // length of uris
	}
	return arg
}

func getTransferDataArgs(transferData *transferData) string {
	if transferData == nil {
		return "00"
	}

	transferDataArgs := "01" +
		getUint64Bytes(transferData.GasLimit) +
		lengthOn4Bytes(len(transferData.Function)) +
		hex.EncodeToString(transferData.Function) +
		lengthOn4Bytes(len(transferData.Args))
	for _, arg := range transferData.Args {
		transferDataArgs = transferDataArgs +
			lengthOn4Bytes(len(arg)) +
			hex.EncodeToString(arg)
	}
	return transferDataArgs
}

func registerTokens(
	t *testing.T,
	cs chainSim.ChainSimulator,
	wallet []byte,
	nonce *uint64,
	esdtSafeAddress []byte,
	token chainSim.ArgsDepositToken,
) {
	ticker := getTokenTicker(token.Identifier)

	registerTokenArgs := registerTokenFunc +
		"@" + generateRandomHash() +
		"@" +
		lengthOn4Bytes(len(token.Identifier)) + // length of identifier
		hex.EncodeToString([]byte(token.Identifier)) + // identifier
		fmt.Sprintf("%02x", uint32(token.Type)) + // type
		lengthOn4Bytes(len(ticker)) + // length of name
		hex.EncodeToString([]byte(ticker)) + // name
		lengthOn4Bytes(len(ticker)) + // length of ticker
		hex.EncodeToString([]byte(ticker)) + // ticker
		//getUint64Bytes(18) + // num decimals
		"00000012" + // num decimals
		getUint64Bytes(1) + // event nonce
		hex.EncodeToString(wallet) + // sender address from other chain
		"00" // no transfer data
	txResult := chainSim.SendTransaction(t, cs, wallet, nonce, esdtSafeAddress, chainSim.ZeroValue, registerTokenArgs, uint64(100_000_000))
	chainSim.RequireSuccessfulTransaction(t, txResult)

	// wait for issue processing from metachain
	err := cs.GenerateBlocks(2)
	require.Nil(t, err)
}

func getTokenTicker(tokenIdentifier string) string {
	return strings.Split(tokenIdentifier, "-")[1]
}

func getTokenIdentifier(token chainSim.ArgsDepositToken) string {
	if token.Nonce == 0 {
		return token.Identifier
	}
	return token.Identifier + "-" + fmt.Sprintf("%02x", token.Nonce)
}

func getTokenNonce(nonce uint64) string {
	if nonce == 0 {
		return ""
	}

	hexStr := fmt.Sprintf("%X", nonce)
	if len(hexStr)%2 != 0 {
		hexStr = "0" + hexStr
	}
	return hexStr
}

func getUint64Bytes(number uint64) string {
	nonceBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(nonceBytes, number)
	return hex.EncodeToString(nonceBytes)
}

func groupTokens(tokens []chainSim.ArgsDepositToken) []chainSim.ArgsDepositToken {
	groupMap := make(map[string]*chainSim.ArgsDepositToken)

	for _, token := range tokens {
		key := fmt.Sprintf("%s:%d", token.Identifier, token.Nonce)
		if existingToken, found := groupMap[key]; found {
			existingToken.Amount.Add(existingToken.Amount, token.Amount)
		} else {
			newAmount := new(big.Int).Set(token.Amount)
			groupMap[key] = &chainSim.ArgsDepositToken{
				Identifier: token.Identifier,
				Nonce:      token.Nonce,
				Amount:     newAmount,
				Type:       token.Type,
			}
		}
	}

	result := make([]chainSim.ArgsDepositToken, 0, len(groupMap))
	for _, token := range groupMap {
		result = append(result, *token)
	}

	return result
}

func lengthOn4Bytes(number int) string {
	numberBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(numberBytes, uint32(number))
	return hex.EncodeToString(numberBytes)
}
