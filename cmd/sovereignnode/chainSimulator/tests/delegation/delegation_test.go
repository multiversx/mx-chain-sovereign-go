package delegation

import (
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	coreAPI "github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/mcl"
	mclsig "github.com/multiversx/mx-chain-crypto-go/signing/mcl/singlesig"
	"github.com/stretchr/testify/require"

	sovereignChainSimulator "github.com/multiversx/mx-chain-go/cmd/sovereignnode/chainSimulator"
	"github.com/multiversx/mx-chain-go/config"
	chainSim "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/integrationTests/chainSimulator/staking"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	"github.com/multiversx/mx-chain-go/vm"
)

const (
	defaultPathToInitialConfig = "../../../../node/config/"
	sovereignConfigPath        = "../../../config/"
)

func TestSovereignChainSimulator_NewDelegationSC(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

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
				cfg.EpochConfig.EnableEpochs.DelegationSmartContractEnableEpoch = 0
			},
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	nodeHandler := cs.GetNodeHandler(core.SovereignChainShardId)

	wallet, err := cs.GenerateAndMintWalletAddress(core.SovereignChainShardId, big.NewInt(0).Mul(chainSim.OneEGLD, big.NewInt(2500)))
	require.Nil(t, err)
	nonce := uint64(0)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	txData := "createNewDelegationContract" +
		"@" + hex.EncodeToString(big.NewInt(0).Mul(chainSim.OneEGLD, big.NewInt(100000)).Bytes()) +
		"@64"
	cost := big.NewInt(0).Mul(chainSim.OneEGLD, big.NewInt(1250))
	txResult := chainSim.SendTransaction(t, cs, wallet.Bytes, &nonce, vm.DelegationManagerSCAddress, cost, txData, uint64(60000000))
	chainSim.RequireSuccessfulTransaction(t, txResult)

	secondDelegationSCAddress := txResult.Logs.Events[1].Topics[4]
	secondDelegationSCAddressBech32, _ := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Encode(secondDelegationSCAddress)
	account, _, err := nodeHandler.GetFacadeHandler().GetAccount(secondDelegationSCAddressBech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, account)
	require.True(t, len(account.Code) > 0)
}

func TestSovereignChainSimulator_DelegateAndClaimRewards(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    50,
	}
	cs, err := sovereignChainSimulator.NewSovereignChainSimulator(sovereignChainSimulator.ArgsSovereignChainSimulator{
		SovereignConfigPath: sovereignConfigPath,
		ArgsChainSimulator: &chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:   true,
			TempDir:                  t.TempDir(),
			PathToInitialConfig:      defaultPathToInitialConfig,
			GenesisTimestamp:         time.Now().Unix(),
			RoundDurationInMillis:    uint64(6000),
			RoundsPerEpoch:           roundsPerEpoch,
			ApiInterface:             api.NewNoApiInterface(),
			MinNodesPerShard:         10,
			NumNodesWaitingListShard: 6,
			AlterConfigsFunction: func(cfg *config.Configs) {
				cfg.EpochConfig.EnableEpochs.DelegationSmartContractEnableEpoch = 0
			},
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	delegationWallet1, err := cs.GenerateAndMintWalletAddress(core.SovereignChainShardId, big.NewInt(0).Mul(chainSim.OneEGLD, big.NewInt(15000)))
	require.Nil(t, err)
	delegationWallet1Nonce := uint64(0)

	delegationWallet2, err := cs.GenerateAndMintWalletAddress(core.SovereignChainShardId, big.NewInt(0).Mul(chainSim.OneEGLD, big.NewInt(15000)))
	require.Nil(t, err)
	delegationWallet2Nonce := uint64(0)

	delegator1, err := cs.GenerateAndMintWalletAddress(core.SovereignChainShardId, big.NewInt(0).Mul(chainSim.OneEGLD, big.NewInt(2500)))
	require.Nil(t, err)
	delegator1Nonce := uint64(0)

	delegator2, err := cs.GenerateAndMintWalletAddress(core.SovereignChainShardId, big.NewInt(0).Mul(chainSim.OneEGLD, big.NewInt(3500)))
	require.Nil(t, err)
	delegator2Nonce := uint64(0)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	// create delegation contract 1 ------------------------------------------------------------------------------------------
	// generate 2 keys, add 2 nodes, stake 2 nodes
	amount := big.NewInt(0).Mul(chainSim.OneEGLD, big.NewInt(5000))
	delegationSCAddress1 := createNewDelegationContract(t, cs, delegationWallet1, &delegationWallet1Nonce, amount)

	validatorSecretKeysBytes, blsKeys, err := chainSimulator.GenerateBlsPrivateKeys(2)
	require.Nil(t, err)
	err = cs.AddValidatorKeys(validatorSecretKeysBytes)
	require.Nil(t, err)

	signatures := getSignatures(delegationSCAddress1.Bytes, validatorSecretKeysBytes)
	txResult := chainSim.SendTransaction(t, cs, delegationWallet1.Bytes, &delegationWallet1Nonce, delegationSCAddress1.Bytes, chainSim.ZeroValue, addNodesTxData(blsKeys, signatures), 500_000_000)
	chainSim.RequireSuccessfulTransaction(t, txResult)

	txData := "stakeNodes"
	for _, key := range blsKeys {
		txData = txData + "@" + key
	}
	txResult = chainSim.SendTransaction(t, cs, delegationWallet1.Bytes, &delegationWallet1Nonce, delegationSCAddress1.Bytes, chainSim.ZeroValue, txData, 500_000_000)
	chainSim.RequireSuccessfulTransaction(t, txResult)

	// create delegation contract 2 ------------------------------------------------------------------------------------------
	// generate 2 keys, add 1 node, stake 1 node
	// stake 2nd key and mergeValidatorToDelegationSameOwner
	amount = big.NewInt(0).Mul(chainSim.OneEGLD, big.NewInt(2500))
	delegationSCAddress2 := createNewDelegationContract(t, cs, delegationWallet2, &delegationWallet2Nonce, amount)

	validatorSecretKeysBytes, blsKeys, err = chainSimulator.GenerateBlsPrivateKeys(2)
	require.Nil(t, err)
	err = cs.AddValidatorKeys(validatorSecretKeysBytes)
	require.Nil(t, err)

	signatures = getSignatures(delegationSCAddress2.Bytes, validatorSecretKeysBytes)
	txResult = chainSim.SendTransaction(t, cs, delegationWallet2.Bytes, &delegationWallet2Nonce, delegationSCAddress2.Bytes, chainSim.ZeroValue, addNodesTxData(blsKeys, signatures), 500_000_000)
	chainSim.RequireSuccessfulTransaction(t, txResult)

	txData = "stakeNodes@" + blsKeys[0]
	txResult = chainSim.SendTransaction(t, cs, delegationWallet2.Bytes, &delegationWallet2Nonce, delegationSCAddress2.Bytes, chainSim.ZeroValue, txData, 200_000_000)
	chainSim.RequireSuccessfulTransaction(t, txResult)

	txData = "stake@01@" + blsKeys[1] + "@" + staking.MockBLSSignature
	txResult = chainSim.SendTransaction(t, cs, delegationWallet2.Bytes, &delegationWallet2Nonce, vm.ValidatorSCAddress, chainSim.MinimumStakeValue, txData, 500_000_000)
	chainSim.RequireSuccessfulTransaction(t, txResult)

	txData = "mergeValidatorToDelegationSameOwner@" + hex.EncodeToString(delegationSCAddress2.Bytes)
	txResult = chainSim.SendTransaction(t, cs, delegationWallet2.Bytes, &delegationWallet2Nonce, vm.DelegationManagerSCAddress, chainSim.ZeroValue, txData, 590_000_000)
	chainSim.RequireSuccessfulTransaction(t, txResult)

	// delegate -----------------------------------------------------
	// each delegator will delegate 1 EGLD to each delegation contract
	txResult = chainSim.SendTransaction(t, cs, delegator1.Bytes, &delegator1Nonce, delegationSCAddress1.Bytes, chainSim.OneEGLD, "delegate", 12_000_000)
	chainSim.RequireSuccessfulTransaction(t, txResult)
	txResult = chainSim.SendTransaction(t, cs, delegator1.Bytes, &delegator1Nonce, delegationSCAddress2.Bytes, chainSim.OneEGLD, "delegate", 12_000_000)
	chainSim.RequireSuccessfulTransaction(t, txResult)

	txResult = chainSim.SendTransaction(t, cs, delegator2.Bytes, &delegator2Nonce, delegationSCAddress1.Bytes, chainSim.OneEGLD, "delegate", 12_000_000)
	chainSim.RequireSuccessfulTransaction(t, txResult)
	txResult = chainSim.SendTransaction(t, cs, delegator2.Bytes, &delegator2Nonce, delegationSCAddress2.Bytes, chainSim.OneEGLD, "delegate", 12_000_000)
	chainSim.RequireSuccessfulTransaction(t, txResult)

	// wait some epochs to generate rewards
	err = cs.GenerateBlocksUntilEpochIsReached(6)
	require.Nil(t, err)

	claimRewardsAndCheckBalance(t, cs, delegator1, &delegator1Nonce, delegationSCAddress1.Bytes)
	claimRewardsAndCheckBalance(t, cs, delegator1, &delegator1Nonce, delegationSCAddress2.Bytes)
	claimRewardsAndCheckBalance(t, cs, delegator2, &delegator2Nonce, delegationSCAddress1.Bytes)
	claimRewardsAndCheckBalance(t, cs, delegator2, &delegator2Nonce, delegationSCAddress2.Bytes)
}

func createNewDelegationContract(
	t *testing.T,
	cs chainSim.ChainSimulator,
	wallet dtos.WalletAddress,
	nonce *uint64,
	amount *big.Int,
) dtos.WalletAddress {
	txData := "createNewDelegationContract" +
		"@021e19e0c9bab2400000" +
		"@03e8"
	txResult := chainSim.SendTransaction(t, cs, wallet.Bytes, nonce, vm.DelegationManagerSCAddress, amount, txData, uint64(60_000_000))
	chainSim.RequireSuccessfulTransaction(t, txResult)

	nodeHandler := cs.GetNodeHandler(core.SovereignChainShardId)

	delegationSCAddress := txResult.Logs.Events[1].Topics[4]
	delegationSCAddressBech32, _ := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Encode(delegationSCAddress)
	account, _, err := nodeHandler.GetFacadeHandler().GetAccount(delegationSCAddressBech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, account)
	require.True(t, len(account.Code) > 0)

	return dtos.WalletAddress{
		Bytes:  delegationSCAddress,
		Bech32: delegationSCAddressBech32,
	}
}

func claimRewardsAndCheckBalance(
	t *testing.T,
	cs chainSim.ChainSimulator,
	delegator dtos.WalletAddress,
	nonce *uint64,
	delegationSC []byte,
) {
	nodeHandler := cs.GetNodeHandler(core.SovereignChainShardId)
	account, _, err := nodeHandler.GetFacadeHandler().GetAccount(delegator.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, account)
	accountBalanceBeforeClaim, _ := big.NewInt(0).SetString(account.Balance, 10)

	txResult := chainSim.SendTransaction(t, cs, delegator.Bytes, nonce, delegationSC, chainSim.ZeroValue, "claimRewards", 6_000_000)
	chainSim.RequireSuccessfulTransaction(t, txResult)

	account, _, err = nodeHandler.GetFacadeHandler().GetAccount(delegator.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, account)
	accBalanceAfterClaim, _ := big.NewInt(0).SetString(account.Balance, 10)

	require.Greater(t, big.NewInt(0).Sub(accBalanceAfterClaim, accountBalanceBeforeClaim).Int64(), int64(0))
}

func getSignatures(msg []byte, blsKeys [][]byte) [][]byte {
	signer := mclsig.NewBlsSigner()

	signatures := make([][]byte, len(blsKeys))
	for i, blsKey := range blsKeys {
		sk, _ := signing.NewKeyGenerator(mcl.NewSuiteBLS12()).PrivateKeyFromByteArray(blsKey)
		signatures[i], _ = signer.Sign(sk, msg)
	}

	return signatures
}

func addNodesTxData(blsKeys []string, sigs [][]byte) string {
	txData := "addNodes"

	for i := range blsKeys {
		txData = txData + "@" + blsKeys[i] + "@" + hex.EncodeToString(sigs[i])
	}

	return txData
}
