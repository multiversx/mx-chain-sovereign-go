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
	sovereignChainSimulator "github.com/multiversx/mx-chain-go/cmd/sovereignnode/chainSimulator"
	chainSimulatorIntegrationTests "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/integrationTests/chainSimulator/staking"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/process"
	"github.com/multiversx/mx-chain-go/vm"
	"github.com/multiversx/mx-sdk-abi-go/abi"
	"github.com/stretchr/testify/require"
)

var serializer, _ = abi.NewSerializer(abi.ArgsNewSerializer{PartsSeparator: "@"})

func getBridgeDataFromPrevBlock(
	t *testing.T,
	nodeHandler process.NodeHandler,
) (uint64, *sovereign.BridgeOutGoingData) {
	prevHdrHash := nodeHandler.GetDataComponents().Blockchain().GetCurrentBlockHeader().GetPrevHash()
	prevHdr, err := nodeHandler.GetDataComponents().Datapool().Headers().GetHeaderByHash(prevHdrHash)
	require.Nil(t, err)

	outGoingMBHdrs := prevHdr.(data.SovereignChainHeaderHandler).GetOutGoingMiniBlockHeaderHandlers()
	require.Len(t, outGoingMBHdrs, 1)

	return prevHdr.GetNonce(), nodeHandler.GetRunTypeComponents().OutGoingOperationsPoolHandler().Get(outGoingMBHdrs[0].GetOutGoingOperationsHash())
}

func checkOutGoingMiniBlockUnRegisterValidator(
	t *testing.T,
	nodeHandler process.NodeHandler,
	numOperations int,
	expectedBlsKeys [][]byte,
	latestMainChainID int,
) {
	// StakeNodes func from staking/common.go generates one extra block after staking tx, so we need to get
	// data from previous block
	nonce, bridgeData := getBridgeDataFromPrevBlock(t, nodeHandler)
	require.Equal(t, int32(block.OutGoingMBUnRegisterBlsKey), bridgeData.Type)
	require.Len(t, bridgeData.OutGoingOperations, numOperations)

	blsKeys := make([][]byte, 0)
	assignedMainChainIDs := make([][]byte, 0)
	expectedMainChainIDs := make([][]byte, 0)

	for _, op := range bridgeData.OutGoingOperations {
		registeredData := deserializeRegisteredBlsKeyData(t, nodeHandler, serializer, op.Data)
		require.Equal(t, nonce, registeredData.Nonce)

		blsKeys = append(blsKeys, registeredData.Key)
		assignedMainChainIDs = append(assignedMainChainIDs, registeredData.ID)

		expectedMainChainIDs = append(expectedMainChainIDs, big.NewInt(int64(latestMainChainID)).Bytes())
		latestMainChainID--
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
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	privateKeys, blsKeys, err := chainSimulator.GenerateBlsPrivateKeys(1)
	require.Nil(t, err)
	err = cs.AddValidatorKeys(privateKeys)
	require.Nil(t, err)

	mintValue := big.NewInt(0).Mul(chainSimulatorIntegrationTests.OneEGLD, big.NewInt(2600))
	walletAddress, err := cs.GenerateAndMintWalletAddress(core.SovereignChainShardId, mintValue)
	require.Nil(t, err)

	err = cs.GenerateBlocksUntilEpochIsReached(1)
	require.Nil(t, err)

	txDataField := fmt.Sprintf("stake@01@%s@%s", blsKeys[0], staking.MockBLSSignature)
	txStake := chainSimulatorIntegrationTests.GenerateTransaction(walletAddress.Bytes, 0, vm.ValidatorSCAddress, chainSimulatorIntegrationTests.MinimumStakeValue, txDataField, staking.GasLimitForStakeOperation)
	stakeTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txStake, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, stakeTx)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	latestMainChainID := 8 // 8 nodes from genesis
	nodeHandler := cs.GetNodeHandler(core.SovereignChainShardId)
	checkOutGoingMiniBlockRegisterValidator(t, nodeHandler, 1, latestMainChainID)
	latestMainChainID++

	txUnStake := chainSimulatorIntegrationTests.GenerateTransaction(walletAddress.Bytes, 1, vm.ValidatorSCAddress, chainSimulatorIntegrationTests.ZeroValue, fmt.Sprintf("unStake@%s", blsKeys[0]), staking.GasLimitForStakeOperation)
	unStakeTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txUnStake, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, unStakeTx)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	blsKeysBytes := getBlsKeyBytes(t, blsKeys)
	require.Equal(t, "unStaked", staking.GetBLSKeyStatus(t, nodeHandler, blsKeysBytes[0]))

	checkOutGoingMiniBlockUnRegisterValidator(t, nodeHandler, 1, getBlsKeyBytes(t, blsKeys), latestMainChainID)
}
