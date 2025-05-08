package esdt

import (
	"encoding/hex"
	"math/big"
	"strconv"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	dataApi "github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/stretchr/testify/require"

	sovereignChainSimulator "github.com/multiversx/mx-chain-go/cmd/sovereignnode/chainSimulator"
	"github.com/multiversx/mx-chain-go/config"
	chainSim "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	"github.com/multiversx/mx-chain-go/vm"
)

const (
	proofTypeMicro = "uPFT"
	proofTypeDAG   = "dPFT"
)

var proofRoles = []string{
	core.ESDTRoleNFTCreate,
	core.ESDTRoleNFTBurn,
	core.ESDTRoleNFTAddQuantity,
	core.ESDTRoleNFTUpdateAttributes,
	core.ESDTRoleNFTAddURI,
	core.ESDTRoleNFTRecreate,
	core.ESDTRoleModifyCreator,
	core.ESDTRoleModifyRoyalties,
	core.ESDTRoleSetNewURI,
	core.ESDTRoleNFTUpdate,
}

func TestSovereignChainSimulator_RegisterTokenProof(t *testing.T) {
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
				cfg.SystemSCConfig.ESDTSystemSCConfig.BaseIssuingCost = issuePrice
			},
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	time.Sleep(time.Second) // wait for VM to be ready for processing queries

	nonce := uint64(0)
	wallet, err := cs.GenerateAndMintWalletAddress(core.SovereignChainShardId, chainSim.InitialAmount)
	require.Nil(t, err)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	// Register Proof Ticker
	proofIdentifier := registerProofTicker(t, cs, wallet, &nonce)

	proofs := make([]string, 0)

	// Create MicroProof FT
	createMicroProofFT(t, cs, wallet, &nonce, proofIdentifier)
	proof := validateProof(t, cs, proofIdentifier, 1, proofTypeMicro)
	proofs = append(proofs, proof)

	// Create secondMicroProof FT
	createMicroProofFT(t, cs, wallet, &nonce, proofIdentifier)
	proof = validateProof(t, cs, proofIdentifier, 2, proofTypeMicro)
	proofs = append(proofs, proof)

	// Create DAGProof FT
	createDAGProofFT(t, cs, wallet, &nonce, proofIdentifier, proofs)
	validateProof(t, cs, proofIdentifier, 3, proofTypeDAG)
}

func registerProofTicker(
	t *testing.T,
	cs chainSim.ChainSimulator,
	wallet dtos.WalletAddress,
	nonce *uint64,
) string {
	nodeHandler := cs.GetNodeHandler(core.SovereignChainShardId)

	issueCost, _ := big.NewInt(0).SetString(issuePrice, 10)
	proofName := "ProofName"
	proofTicker := "PROOF"
	args := "registerProofTicker" +
		"@" + hex.EncodeToString([]byte(proofName)) +
		"@" + hex.EncodeToString([]byte(proofTicker))
	chainSim.SendTransactionWithSuccess(t, cs, wallet.Bytes, nonce, vm.ESDTSCAddress, issueCost, args, uint64(60000000))

	issuedESDTs, err := nodeHandler.GetFacadeHandler().GetAllIssuedESDTs("proof")
	require.Nil(t, err)
	require.NotNil(t, issuedESDTs)
	require.Equal(t, 1, len(issuedESDTs))

	proofIdentifier := issuedESDTs[0]
	checkAllRoles(t, nodeHandler, wallet.Bech32, proofIdentifier, proofRoles)

	return proofIdentifier
}

func createMicroProofFT(
	t *testing.T,
	cs chainSim.ChainSimulator,
	wallet dtos.WalletAddress,
	nonce *uint64,
	proofIdentifier string,
) {
	args := "createProof" +
		"@" + hex.EncodeToString([]byte(proofIdentifier)) +
		"@" + hex.EncodeToString([]byte(proofTypeMicro)) +
		"@" + hex.EncodeToString([]byte("proofData"))
	chainSim.SendTransactionWithSuccess(t, cs, wallet.Bytes, nonce, vm.ESDTSCAddress, chainSim.ZeroValue, args, uint64(60000000))
}

func createDAGProofFT(
	t *testing.T,
	cs chainSim.ChainSimulator,
	wallet dtos.WalletAddress,
	nonce *uint64,
	proofIdentifier string,
	parentKeys []string,
) {
	args := "createProof" +
		"@" + hex.EncodeToString([]byte(proofIdentifier)) +
		"@" + hex.EncodeToString([]byte(proofTypeDAG)) +
		"@" + hex.EncodeToString([]byte("proofData"))
	for _, parentKey := range parentKeys {
		args = args +
			"@" + hex.EncodeToString([]byte(parentKey))
	}

	chainSim.SendTransactionWithSuccess(t, cs, wallet.Bytes, nonce, vm.ESDTSCAddress, chainSim.ZeroValue, args, uint64(60000000))
}

func validateProof(
	t *testing.T,
	cs chainSim.ChainSimulator,
	proofIdentifier string,
	proofNonce uint64,
	proofType string,
) string {
	nodeHandler := cs.GetNodeHandler(core.SovereignChainShardId)
	systemAccount, _ := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Encode(core.SystemAccountAddress)
	systemAccountKeys, _, err := nodeHandler.GetFacadeHandler().GetKeyValuePairs(systemAccount, dataApi.AccountQueryOptions{})
	require.Nil(t, err)

	nonceStr := strconv.FormatUint(proofNonce, 10)
	proofKey := hex.EncodeToString([]byte("PFT:" + proofIdentifier + "-" + nonceStr))
	proofValue := systemAccountKeys[proofKey]
	require.NotNil(t, proofValue)
	proofValueBytes, _ := hex.DecodeString(proofValue)
	validateProofValueLength(t, len(proofValueBytes), proofType)

	proofNonceKey := hex.EncodeToString([]byte(core.ESDTNFTLatestNonceIdentifier + proofIdentifier))
	require.Equal(t, hex.EncodeToString(big.NewInt(0).SetUint64(proofNonce).Bytes()), systemAccountKeys[proofNonceKey])

	return proofIdentifier + "-" + nonceStr
}

func validateProofValueLength(
	t *testing.T,
	proofLen int,
	proofType string,
) {
	switch proofType {
	case proofTypeMicro:
		require.Equal(t, proofLen, 44)
	case proofTypeDAG:
		require.Equal(t, proofLen, 80)
	default:
		require.Fail(t, "invalid proofType")
	}
}
