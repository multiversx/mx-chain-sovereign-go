package relayedTx

import (
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"

	sovChainSimulator "github.com/multiversx/mx-chain-go/cmd/sovereignnode/chainSimulator"
	"github.com/multiversx/mx-chain-go/config"
	testsChainSimulator "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
)

const (
	defaultPathToSovereignConfig = "../../../cmd/sovereignnode/config/"
)

func TestRelayedV3WithSovereignChainSimulator(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	runRelayedV3TestsWithChainSimulator(t, sovereignChainSimulator)
}

func sovereignChainSimulator(
	t *testing.T,
	alterConfigsFunction func(cfg *config.Configs),
) testsChainSimulator.ChainSimulator {
	roundDurationInMillis := uint64(6000)
	roundsPerEpochOpt := core.OptionalUint64{
		HasValue: true,
		Value:    roundsPerEpoch,
	}

	cs, err := sovChainSimulator.NewSovereignChainSimulator(sovChainSimulator.ArgsSovereignChainSimulator{
		SovereignConfigPath: defaultPathToSovereignConfig,
		ArgsChainSimulator: &chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:   true,
			TempDir:                  t.TempDir(),
			PathToInitialConfig:      defaultPathToInitialConfig,
			NumOfShards:              1,
			GenesisTimestamp:         time.Now().Unix(),
			RoundDurationInMillis:    roundDurationInMillis,
			RoundsPerEpoch:           roundsPerEpochOpt,
			ApiInterface:             api.NewNoApiInterface(),
			MinNodesPerShard:         3,
			NumNodesWaitingListShard: 3,
			AlterConfigsFunction:     alterConfigsFunction,
		},
	})
	require.NoError(t, err)
	require.NotNil(t, cs)

	err = cs.GenerateBlocksUntilEpochIsReached(1)
	require.NoError(t, err)

	return cs
}
