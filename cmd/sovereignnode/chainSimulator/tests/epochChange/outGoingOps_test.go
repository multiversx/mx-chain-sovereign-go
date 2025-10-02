package epochChange

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	sovereignChainSimulator "github.com/multiversx/mx-chain-go/cmd/sovereignnode/chainSimulator"
	"github.com/multiversx/mx-chain-go/config"
	chainSimulatorIntegrationTests "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/integrationTests/chainSimulator/staking"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/process"
	"github.com/multiversx/mx-chain-go/vm"
	"github.com/multiversx/mx-sdk-abi-go/abi"
	"github.com/stretchr/testify/require"
)

var serializer, _ = abi.NewSerializer(abi.ArgsNewSerializer{PartsSeparator: "@"})

func getBridgeDataFromPrevBlock(
	t *testing.T,
	nodeHandler process.NodeHandler,
	mbType block.OutGoingMBType,
) (uint64, *sovereign.BridgeOutGoingData) {
	prevHdrHash := nodeHandler.GetDataComponents().Blockchain().GetCurrentBlockHeader().GetPrevHash()
	prevHdr, err := nodeHandler.GetDataComponents().Datapool().Headers().GetHeaderByHash(prevHdrHash)
	require.Nil(t, err)

	outGoingMBHdr := prevHdr.(data.SovereignChainHeaderHandler).GetOutGoingMiniBlockHeaderHandler(int32(mbType))
	require.NotNil(t, outGoingMBHdr)

	return prevHdr.GetNonce(), nodeHandler.GetRunTypeComponents().OutGoingOperationsPoolHandler().Get(outGoingMBHdr.GetOutGoingOperationsHash())
}

func checkOutGoingMiniBlockRegisterValidator(
	t *testing.T,
	nodeHandler process.NodeHandler,
	numOperations int,
	latestMainChainID int,
) {
	nonce, bridgeData := getBridgeDataFromPrevBlock(t, nodeHandler, block.OutGoingMBRegisterBlsKey)
	require.Equal(t, int32(block.OutGoingMBRegisterBlsKey), bridgeData.Type)
	require.Len(t, bridgeData.OutGoingOperations, numOperations)

	blsKeys := make([][]byte, 0)
	assignedMainChainIDs := make([][]byte, 0)

	expectedMainChainIDs := make([][]byte, 0)
	for _, op := range bridgeData.OutGoingOperations {
		registeredData := deserializeRegisteredBlsKeyData(t, nodeHandler, serializer, op.Data)
		require.Equal(t, nonce, registeredData.Nonce)

		blsKeys = append(blsKeys, registeredData.Key)
		assignedMainChainIDs = append(assignedMainChainIDs, registeredData.ID)

		latestMainChainID++
		expectedMainChainIDs = append(expectedMainChainIDs, big.NewInt(int64(latestMainChainID)).Bytes())
	}

	auctionNodes := getAuctionListKeys(t, nodeHandler)
	require.ElementsMatch(t, expectedMainChainIDs, assignedMainChainIDs)
	require.Subset(t, auctionNodes, blsKeys)
}

func checkOutGoingMiniBlockUnRegisterValidator(
	t *testing.T,
	nodeHandler process.NodeHandler,
	expectedBlsKeys [][]byte,
	expectedMainChainIDs [][]byte,
) {
	// StakeNodes func from staking/common.go generates one extra block after staking tx, so we need to get
	// data from previous block
	nonce, bridgeData := getBridgeDataFromPrevBlock(t, nodeHandler, block.OutGoingMBUnRegisterBlsKey)
	require.Equal(t, int32(block.OutGoingMBUnRegisterBlsKey), bridgeData.Type)
	require.Len(t, bridgeData.OutGoingOperations, len(expectedBlsKeys))

	blsKeys := make([][]byte, 0)
	assignedMainChainIDs := make([][]byte, 0)

	for _, op := range bridgeData.OutGoingOperations {
		registeredData := deserializeRegisteredBlsKeyData(t, nodeHandler, serializer, op.Data)
		require.Equal(t, nonce, registeredData.Nonce)

		blsKeys = append(blsKeys, registeredData.Key)
		assignedMainChainIDs = append(assignedMainChainIDs, registeredData.ID)
	}

	require.ElementsMatch(t, expectedMainChainIDs, assignedMainChainIDs)
	require.ElementsMatch(t, expectedBlsKeys, blsKeys)
}

func getBlsKeyBytes(t *testing.T, keys []string) [][]byte {
	blsKeys := make([][]byte, len(keys))
	for i, k := range keys {
		key, err := hex.DecodeString(k)
		require.Nil(t, err)
		blsKeys[i] = key
	}

	return blsKeys
}

func createStakeTxs(
	nonce *uint64,
	walletAddress dtos.WalletAddress,
	blsKeys []string,
) []*transaction.Transaction {
	txs := make([]*transaction.Transaction, len(blsKeys))
	for i := 0; i < len(blsKeys); i++ {
		txDataField := fmt.Sprintf("stake@01@%s@%s", blsKeys[i], staking.MockBLSSignature)
		txStake := chainSimulatorIntegrationTests.GenerateTransaction(
			walletAddress.Bytes,
			*nonce,
			vm.ValidatorSCAddress,
			chainSimulatorIntegrationTests.MinimumStakeValue,
			txDataField,
			staking.GasLimitForStakeOperation,
		)

		txs[i] = txStake
		*nonce++
	}

	return txs
}

func createUnStakeTxs(
	nonce *uint64,
	walletAddress dtos.WalletAddress,
	blsKeys []string,
) []*transaction.Transaction {
	txs := make([]*transaction.Transaction, len(blsKeys))

	for i := 0; i < len(blsKeys); i++ {
		txUnStake := chainSimulatorIntegrationTests.GenerateTransaction(
			walletAddress.Bytes,
			*nonce,
			vm.ValidatorSCAddress,
			chainSimulatorIntegrationTests.ZeroValue,
			fmt.Sprintf("unStake@%s", blsKeys[i]),
			staking.GasLimitForStakeOperation,
		)

		txs[i] = txUnStake
		*nonce++
	}

	return txs
}

func sendTxsAndGenerateOneBlock(t *testing.T, cs chainSimulatorIntegrationTests.ChainSimulator, txs []*transaction.Transaction) {
	txsResults, err := cs.SendTxsAndGenerateBlocksTilAreExecuted(txs, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txsResults)
	err = cs.GenerateBlocks(1)
	require.Nil(t, err)
}

func TestSovereignChainSimulator_OutgoingOpRegisterAndUnregisterNode(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    25,
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
			MinNodesPerShard:         6,
			NumNodesWaitingListShard: 2,
			AlterConfigsFunction: func(cfg *config.Configs) {
				cfg.SystemSCConfig.StakingSystemSCConfig.NodeLimitPercentage = 1.0
			},
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	privateKeys, blsKeys, err := chainSimulator.GenerateBlsPrivateKeys(4)
	require.Nil(t, err)
	err = cs.AddValidatorKeys(privateKeys)
	require.Nil(t, err)

	mintValue := big.NewInt(0).Mul(chainSimulatorIntegrationTests.OneEGLD, big.NewInt(2600*10))
	walletAddress, err := cs.GenerateAndMintWalletAddress(core.SovereignChainShardId, mintValue)
	require.Nil(t, err)

	err = cs.GenerateBlocksUntilEpochIsReached(1)
	require.Nil(t, err)

	nonce := uint64(0)
	nodeHandler := cs.GetNodeHandler(core.SovereignChainShardId)

	// Create a block with 3 stake transactions (by staking first 3/4 nodes) and check outgoing mbs
	txs := createStakeTxs(&nonce, walletAddress, blsKeys[:3])
	sendTxsAndGenerateOneBlock(t, cs, txs)
	checkOutGoingMiniBlockRegisterValidator(t, nodeHandler, 3, 8)

	// Create a block with unStake(1/4 & 3/4 nodes) and stake(4/4 node) transactions and check there are two outgoing mbs with
	// different types
	txsUnStake := createUnStakeTxs(&nonce, walletAddress, []string{blsKeys[0], blsKeys[2]})
	txStakeLastNode := createStakeTxs(&nonce, walletAddress, blsKeys[3:])
	sendTxsAndGenerateOneBlock(t, cs, append(txsUnStake, txStakeLastNode...))

	err = cs.ForceResetValidatorStatisticsCache()
	require.Nil(t, err)

	blsKeysBytes := getBlsKeyBytes(t, blsKeys)
	checkOutGoingMiniBlockUnRegisterValidator(t, nodeHandler, [][]byte{blsKeysBytes[0], blsKeysBytes[2]}, [][]byte{{0x09}, {0x0b}})
	checkOutGoingMiniBlockRegisterValidator(t, nodeHandler, 1, 8+3) // 8 from genesis + 3 staked nodes in total

	require.Equal(t, "unStaked", staking.GetBLSKeyStatus(t, nodeHandler, blsKeysBytes[0]))
	require.Equal(t, "staked", staking.GetBLSKeyStatus(t, nodeHandler, blsKeysBytes[1]))
	require.Equal(t, "unStaked", staking.GetBLSKeyStatus(t, nodeHandler, blsKeysBytes[2]))
	require.Equal(t, "staked", staking.GetBLSKeyStatus(t, nodeHandler, blsKeysBytes[3]))
}
