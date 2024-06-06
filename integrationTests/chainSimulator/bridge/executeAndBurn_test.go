package bridge

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-go/config"
	chainSim "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"
)

const (
	issuePrice = "5000000000000000000"
)

func TestChainSimulator_ExecuteWithMintAndBurnFungibleWithDeposit(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	token := "sov1-SOVTKN-1a2b3c"
	tokenNonce := uint64(0)

	mintTokens := make([]chainSim.ArgsDepositToken, 0)
	mintTokens = append(mintTokens, chainSim.ArgsDepositToken{
		Identifier: token,
		Nonce:      tokenNonce,
		Amount:     big.NewInt(123),
		Type:       core.Fungible,
	})
	mintTokens = append(mintTokens, chainSim.ArgsDepositToken{
		Identifier: token,
		Nonce:      tokenNonce,
		Amount:     big.NewInt(100),
		Type:       core.Fungible,
	})

	depositTokens := make([]chainSim.ArgsDepositToken, 0)
	depositTokens = append(depositTokens, chainSim.ArgsDepositToken{
		Identifier: token,
		Nonce:      tokenNonce,
		Amount:     big.NewInt(12),
		Type:       core.Fungible,
	})
	depositTokens = append(depositTokens, chainSim.ArgsDepositToken{
		Identifier: token,
		Nonce:      tokenNonce,
		Amount:     big.NewInt(10),
		Type:       core.Fungible,
	})

	simulateExecutionAndDeposit(t, mintTokens, depositTokens)
}

func TestChainSimulator_ExecuteWithMintAndBurnNftWithDeposit(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	nft := "sov2-SOVNFT-123456"
	nftNonce := uint64(1)
	token := "sov3-TKN-1q2w3e"

	mintTokens := make([]chainSim.ArgsDepositToken, 0)
	mintTokens = append(mintTokens, chainSim.ArgsDepositToken{
		Identifier: nft,
		Nonce:      nftNonce,
		Amount:     big.NewInt(1),
		Type:       core.NonFungible,
	})
	mintTokens = append(mintTokens, chainSim.ArgsDepositToken{
		Identifier: token,
		Nonce:      0,
		Amount:     big.NewInt(1),
		Type:       core.Fungible,
	})

	depositTokens := make([]chainSim.ArgsDepositToken, 0)
	depositTokens = append(depositTokens, chainSim.ArgsDepositToken{
		Identifier: nft,
		Nonce:      nftNonce,
		Amount:     big.NewInt(1),
		Type:       core.NonFungible,
	})

	simulateExecutionAndDeposit(t, mintTokens, depositTokens)
}

func TestChainSimulator_ExecuteWithMintAndBurnSftWithDeposit(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	sft := "sov3-SOVSFT-654321"
	sftNonce := uint64(1)

	mintTokens := make([]chainSim.ArgsDepositToken, 0)
	mintTokens = append(mintTokens, chainSim.ArgsDepositToken{
		Identifier: sft,
		Nonce:      sftNonce,
		Amount:     big.NewInt(50),
		Type:       core.SemiFungible,
	})

	depositTokens := make([]chainSim.ArgsDepositToken, 0)
	depositTokens = append(depositTokens, chainSim.ArgsDepositToken{
		Identifier: sft,
		Nonce:      sftNonce,
		Amount:     big.NewInt(20),
		Type:       core.SemiFungible,
	})

	simulateExecutionAndDeposit(t, mintTokens, depositTokens)
}

func simulateExecutionAndDeposit(
	t *testing.T,
	mintTokens []chainSim.ArgsDepositToken,
	depositTokens []chainSim.ArgsDepositToken,
) {
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}

	whiteListedAddress := "erd1qqqqqqqqqqqqqpgqmzzm05jeav6d5qvna0q2pmcllelkz8xddz3syjszx5"
	cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck:      false,
		TempDir:                     t.TempDir(),
		PathToInitialConfig:         defaultPathToInitialConfig,
		NumOfShards:                 1,
		GenesisTimestamp:            time.Now().Unix(),
		RoundDurationInMillis:       uint64(6000),
		RoundsPerEpoch:              roundsPerEpoch,
		ApiInterface:                api.NewNoApiInterface(),
		MinNodesPerShard:            3,
		MetaChainMinNodes:           3,
		NumNodesWaitingListMeta:     0,
		NumNodesWaitingListShard:    0,
		ConsensusGroupSize:          1,
		MetaChainConsensusGroupSize: 1,
		AlterConfigsFunction: func(cfg *config.Configs) {
			cfg.GeneralConfig.VirtualMachine.Execution.TransferAndExecuteByUserAddresses = []string{whiteListedAddress}
			cfg.SystemSCConfig.ESDTSystemSCConfig.BaseIssuingCost = issuePrice
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	nodeHandler := cs.GetNodeHandler(0)

	initialAddress := "erd1l6xt0rqlyzw56a3k8xwwshq2dcjwy3q9cppucvqsmdyw8r98dz3sae0kxl"
	initialAddrBytes, err := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode(initialAddress)
	require.Nil(t, err)
	err = cs.SetStateMultiple([]*dtos.AddressState{
		{
			Address: initialAddress,
			Balance: "10000000000000000000000",
		},
		{
			Address: "erd1lllllllllllllllllllllllllllllllllllllllllllllllllllsckry7t", // init sys account
		},
	})
	require.Nil(t, err)

	err = cs.GenerateBlocksUntilEpochIsReached(3)
	require.Nil(t, err)

	wallet := dtos.WalletAddress{Bech32: initialAddress, Bytes: initialAddrBytes}
	nonce := uint64(0)

	bridgeData := DeployBridgeSetup(t, cs, wallet.Bytes, &nonce, esdtSafeWasmPath, feeMarketWasmPath)

	esdtSafeEncoded, _ := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Encode(bridgeData.ESDTSafeAddress)
	require.Equal(t, esdtSafeEncoded, whiteListedAddress)

	// We will deposit an array of prefixed tokens from a sovereign chain to the main chain,
	// expecting these tokens to be minted by the whitelisted ESDT safe sc and transferred to our wallet address.
	executeBridgeOperation(t, cs, wallet, &nonce, bridgeData.ESDTSafeAddress, mintTokens)

	// Deposit an array of tokens from main chain to sovereign chain,
	// expecting these tokens to be burned by the whitelisted ESDT safe sc
	Deposit(t, cs, wallet.Bytes, &nonce, bridgeData.ESDTSafeAddress, depositTokens, wallet.Bytes)
	mintedTokens := groupTokens(mintTokens)
	for _, token := range groupTokens(depositTokens) {
		mintedValue, err := getMintedValue(mintedTokens, token.Identifier)
		require.Nil(t, err)

		fullTokenIdentifier := getTokenIdentifier(token)
		chainSim.RequireAccountHasToken(t, cs, fullTokenIdentifier, wallet.Bech32, big.NewInt(0).Sub(mintedValue, token.Amount))
		chainSim.RequireAccountHasToken(t, cs, fullTokenIdentifier, esdtSafeEncoded, big.NewInt(0))

		tokenSupply, err := nodeHandler.GetFacadeHandler().GetTokenSupply(fullTokenIdentifier)
		require.Nil(t, err)
		require.NotNil(t, tokenSupply)
		require.Equal(t, token.Amount.String(), tokenSupply.Burned)
	}
}

func getMintedValue(mintTokens []chainSim.ArgsDepositToken, token string) (*big.Int, error) {
	for _, tkn := range mintTokens {
		if tkn.Identifier == token {
			return tkn.Amount, nil
		}
	}
	return nil, fmt.Errorf("token not found")
}

func executeBridgeOperation(
	t *testing.T,
	cs chainSim.ChainSimulator,
	wallet dtos.WalletAddress,
	nonce *uint64,
	esdtSafeAddress []byte,
	mintTokens []chainSim.ArgsDepositToken,
) {
	executeBridgeOpsData := "executeBridgeOps" +
		"@de96b8d3842668aad676f915f545403b3e706f8f724cefb0c15b728e83864ce7" + //dummy hash
		"@" + // operation
		hex.EncodeToString(wallet.Bytes) + // receiver address
		lengthOn4Bytes(len(mintTokens)) + // nr of tokens
		getTokenDataArgs(mintTokens) + // tokens encoded arg
		"0000000000000000" + // event nonce
		hex.EncodeToString(wallet.Bytes) + // sender address from other chain
		"00" // no transfer data
	chainSim.SendTransaction(t, cs, wallet.Bytes, nonce, esdtSafeAddress, chainSim.ZeroValue, executeBridgeOpsData, uint64(50000000))
	for _, token := range groupTokens(mintTokens) {
		chainSim.RequireAccountHasToken(t, cs, getTokenIdentifier(token), wallet.Bech32, token.Amount)
	}
}

func getTokenDataArgs(tokens []chainSim.ArgsDepositToken) string {
	var arg string
	for _, token := range tokens {
		arg = arg +
			lengthOn4Bytes(len(token.Identifier)) + // length of token identifier
			hex.EncodeToString([]byte(token.Identifier)) + //token identifier
			getNonceHex(token.Nonce) + // nonce
			fmt.Sprintf("%02x", token.Type) + // type
			lengthOn4Bytes(len(token.Amount.Bytes())) + // length of amount
			hex.EncodeToString(token.Amount.Bytes()) + // amount
			"00" + // not frozen
			lengthOn4Bytes(0) + // length of hash
			lengthOn4Bytes(0) + // length of name
			lengthOn4Bytes(0) + // length of attributes
			hex.EncodeToString(bytes.Repeat([]byte{0x00}, 32)) + // creator
			lengthOn4Bytes(0) + // length of royalties
			lengthOn4Bytes(0) // length of uris
	}
	return arg
}

func getTokenIdentifier(token chainSim.ArgsDepositToken) string {
	if token.Nonce == 0 {
		return token.Identifier
	}
	return token.Identifier + "-" + fmt.Sprintf("%02x", token.Nonce)
}

func getNonceHex(nonce uint64) string {
	nonceBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(nonceBytes, nonce)
	return hex.EncodeToString(nonceBytes)
}

func lengthOn4Bytes(number int) string {
	numberBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(numberBytes, uint32(number))
	return hex.EncodeToString(numberBytes)
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
