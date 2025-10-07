package bridge

import (
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-sdk-abi-go/abi"
	"github.com/stretchr/testify/require"

	sovereignChainSimulator "github.com/multiversx/mx-chain-go/cmd/sovereignnode/chainSimulator"
	"github.com/multiversx/mx-chain-go/cmd/sovereignnode/dataCodec"
	"github.com/multiversx/mx-chain-go/config"
	chainSim "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
)

// This test will:
// - deploy bridge contracts setup
// - issue a new fungible token
// - deposit the token multiple times in the same block
// - verify outgoing operations order in the outgoing miniblock
// - confirm a part of the operation and let the rest expire
// - check GetUnconfirmedOperations will return the operations in expected order
// - deposit the token again multiple times in the same block
// - verify the new outgoing operations order in the outgoing miniblock
// - wait until all the operations expire
// - check GetUnconfirmedOperations will return all the remaining operations in expected order
// - confirm all remaining operation
func TestSovereignChainSimulator_ValidateOutgoingOperationsOrder(t *testing.T) {
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

	//tokenIdentifier, _ := issueAndRegisterToken(t, cs, wallet, &nonce, bridgeData.ESDTSafeAddress)
	// Issue new fungible token
	issueCost, _ := big.NewInt(0).SetString(issuePaymentCost, 10)
	supply, _ := big.NewInt(0).SetString("123000000000000000000", 10)
	tokenName := "SovToken"
	tokenTicker := "SVN"
	numDecimals := 18
	tokenIdentifier := chainSim.IssueFungible(t, cs, wallet.Bytes, &nonce, issueCost, tokenName, tokenTicker, numDecimals, supply)

	creator, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000000")
	esdtToken := sovereign.EsdtToken{
		Identifier: []byte(tokenIdentifier),
		Nonce:      0,
		Data: sovereign.EsdtTokenData{
			TokenType:  core.Fungible,
			Amount:     big.NewInt(1),
			Frozen:     false,
			Hash:       []byte(""),
			Name:       []byte(""),
			Attributes: []byte(""),
			Creator:    creator,
			Royalties:  big.NewInt(0),
			Uris:       [][]byte{},
		},
	}

	currentOpNonce := uint64(0)
	numOfTransfers := 10
	sovereignOutGoingOps := make([]sovereign.Operation, 0)

	// Deposit tokens
	operations := depositEsdtTokenMultipleTimes(t, cs, esdtToken, wallet, &nonce, bridgeData.ESDTSafeAddress, &currentOpNonce, numOfTransfers)
	sovereignOutGoingOps = append(sovereignOutGoingOps, operations...)

	serializer, _ := abi.NewSerializer(abi.ArgsNewSerializer{
		PartsSeparator: "@",
	})
	dtaCodec, _ := dataCodec.NewDataCodec(serializer)

	outGoingOpsPool := nodeHandler.GetRunTypeComponents().OutGoingOperationsPoolHandler()
	sovHeader, _ := nodeHandler.GetChainHandler().GetCurrentBlockHeader().(data.SovereignChainHeaderHandler)

	outGoingMbHeaders := sovHeader.GetOutGoingMiniBlockHeaderHandlers()
	require.Len(t, outGoingMbHeaders, 1)

	bridgeOutGoingData := outGoingOpsPool.Get(outGoingMbHeaders[0].GetOutGoingOperationsHash())
	// Verify outgoing operations order
	for i, outGoingOp := range bridgeOutGoingData.OutGoingOperations {
		serializedOperation, err := dtaCodec.SerializeOperation(sovereignOutGoingOps[i])
		require.NoError(t, err)
		require.Equal(t, outGoingOp.Data, serializedOperation)
	}

	// Generate extra blocks after outgoing sovereignOutGoingOps are created
	err = cs.GenerateBlocks(5)
	require.Nil(t, err)

	outGoingOperationsCopy := getOutgoingOpsCopy(bridgeOutGoingData.OutGoingOperations)
	// Confirm some outgoing operations
	numOpsToConfirm := 5
	for i := 0; i < numOpsToConfirm; i++ {
		outGoingOp := outGoingOperationsCopy[i]
		err = outGoingOpsPool.ConfirmOperation(bridgeOutGoingData.Hash, outGoingOp.Hash)
		require.NoError(t, err)

		sovereignOutGoingOps = sovereignOutGoingOps[1:] // delete confirmed operations
	}

	numOfRemainingTransfers := numOfTransfers - numOpsToConfirm

	err = cs.GenerateBlocks(5)
	require.Nil(t, err)

	// Wait for outgoing sovereignOutGoingOps to get unconfirmed
	time.Sleep(time.Second)

	// Check number of remaining unconfirmed operations
	unconfirmedBridgeOutGoingData := outGoingOpsPool.GetUnconfirmedOperations()
	require.Len(t, unconfirmedBridgeOutGoingData, 1)
	require.Len(t, unconfirmedBridgeOutGoingData[0].OutGoingOperations, numOfRemainingTransfers)
	// Verify unconfirmed operations order
	for i, outGoingOp := range unconfirmedBridgeOutGoingData[0].OutGoingOperations {
		serializedOperation, err := dtaCodec.SerializeOperation(sovereignOutGoingOps[i])
		require.NoError(t, err)
		require.Equal(t, outGoingOp.Data, serializedOperation)
	}

	// Deposit more tokens
	numOfExtraTransfers := 2
	operations = depositEsdtTokenMultipleTimes(t, cs, esdtToken, wallet, &nonce, bridgeData.ESDTSafeAddress, &currentOpNonce, numOfExtraTransfers)
	sovereignOutGoingOps = append(sovereignOutGoingOps, operations...)

	sovHeader, _ = nodeHandler.GetChainHandler().GetCurrentBlockHeader().(data.SovereignChainHeaderHandler)
	outGoingMbHeaders = sovHeader.GetOutGoingMiniBlockHeaderHandlers()
	require.Len(t, outGoingMbHeaders, 1)

	bridgeOutGoingData = outGoingOpsPool.Get(outGoingMbHeaders[0].GetOutGoingOperationsHash())
	// Verify outgoing operations order
	for i, outGoingOp := range bridgeOutGoingData.OutGoingOperations {
		serializedOperation, err := dtaCodec.SerializeOperation(sovereignOutGoingOps[numOfRemainingTransfers+i])
		require.NoError(t, err)
		require.Equal(t, outGoingOp.Data, serializedOperation)
	}

	// Generate extra blocks after outgoing sovereignOutGoingOps are created
	err = cs.GenerateBlocks(5)
	require.Nil(t, err)

	// Wait for outgoing sovereignOutGoingOps to get unconfirmed and check we have one, which is also saved in storage
	time.Sleep(time.Second)

	// Check number of remaining unconfirmed operations
	unconfirmedBridgeOutGoingData = outGoingOpsPool.GetUnconfirmedOperations()
	require.Len(t, unconfirmedBridgeOutGoingData, 2)
	require.Len(t, unconfirmedBridgeOutGoingData[0].OutGoingOperations, numOfRemainingTransfers)
	require.Len(t, unconfirmedBridgeOutGoingData[1].OutGoingOperations, numOfExtraTransfers)

	// Confirm all remaining outgoing operations
	for _, bridgeOutGoingData := range unconfirmedBridgeOutGoingData {
		outGoingOperationsCopy := getOutgoingOpsCopy(bridgeOutGoingData.OutGoingOperations)
		for _, outGoingOp := range outGoingOperationsCopy {
			serializedOperation, _ := dtaCodec.SerializeOperation(sovereignOutGoingOps[0])
			require.Equal(t, outGoingOp.Data, serializedOperation)

			err = outGoingOpsPool.ConfirmOperation(bridgeOutGoingData.Hash, outGoingOp.Hash)
			require.NoError(t, err)

			sovereignOutGoingOps = sovereignOutGoingOps[1:] // delete confirmed operations
		}
	}

	err = cs.GenerateBlocks(5)
	require.Nil(t, err)

	// No outgoing sovereignOutGoingOps unconfirmed
	unconfirmedBridgeOutGoingData = outGoingOpsPool.GetUnconfirmedOperations()
	require.Len(t, unconfirmedBridgeOutGoingData, 0)
	require.Len(t, sovereignOutGoingOps, 0)
}

func getOutgoingOpsCopy(ops []*sovereign.OutGoingOperation) []*sovereign.OutGoingOperation {
	copied := make([]*sovereign.OutGoingOperation, len(ops))
	copy(copied, ops)
	return copied
}

func depositEsdtTokenMultipleTimes(
	t *testing.T,
	cs chainSim.ChainSimulator,
	esdtToken sovereign.EsdtToken,
	wallet dtos.WalletAddress,
	nonce *uint64,
	esdtSafeAddress []byte,
	currentOpNonce *uint64,
	numOfDeposits int,
) []sovereign.Operation {
	transactions := make([]*transaction.Transaction, 0)
	operations := make([]sovereign.Operation, 0)

	for i := 0; i < numOfDeposits; i++ {
		depositToken := chainSim.ArgsDepositToken{
			Identifier: string(esdtToken.Identifier),
			Nonce:      0,
			Amount:     big.NewInt(1),
		}
		depositArgs := createDepositArgs(t, esdtSafeAddress, []chainSim.ArgsDepositToken{depositToken}, wallet.Bytes, nil)
		tx := chainSim.GenerateTransaction(wallet.Bytes, *nonce, wallet.Bytes, chainSim.ZeroValue, depositArgs, uint64(20000000))
		*nonce++
		transactions = append(transactions, tx)

		operations = append(operations, sovereign.Operation{
			Address: wallet.Bytes,
			Tokens:  []sovereign.EsdtToken{esdtToken},
			Data: &sovereign.EventData{
				Nonce:  *currentOpNonce,
				Sender: wallet.Bytes,
			},
		})
		*currentOpNonce++
	}

	txResults, err := cs.SendTxsAndGenerateBlocksTilAreExecuted(transactions, 10)
	require.NoError(t, err)
	for _, txResult := range txResults {
		chainSim.RequireSuccessfulTransaction(t, txResult)
	}

	return operations
}
