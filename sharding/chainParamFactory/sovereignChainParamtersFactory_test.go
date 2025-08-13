package chainParamFactory

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/testscommon/commonmocks"
	"github.com/stretchr/testify/require"
)

func createSovereignArgs() sharding.ArgsChainParametersHolder {
	return sharding.ArgsChainParametersHolder{
		EpochStartEventNotifier: &mock.EpochStartNotifierStub{},
		ChainParameters: []config.ChainParametersByEpochConfig{
			{
				EnableEpoch:                 0,
				ShardMinNumNodes:            5,
				ShardConsensusGroupSize:     5,
				MetachainMinNumNodes:        0,
				MetachainConsensusGroupSize: 0,
				RoundDuration:               4000,
				Hysteresis:                  0.2,
				Adaptivity:                  false,
			},
		},
		ChainParametersNotifier: &commonmocks.ChainParametersNotifierStub{},
	}
}

func TestSovereignChainMessengerFactory_CreateChainParametersHolder(t *testing.T) {
	t.Parallel()

	f := NewSovereignChainParametersHolderFactory()
	require.False(t, f.IsInterfaceNil())

	args := createSovereignArgs()
	holder, err := f.CreateChainParametersHolder(args)
	require.Nil(t, err)
	require.NotNil(t, holder)
	require.Equal(t, "*sharding.sovereignChainParametersHolder", fmt.Sprintf("%T", holder))
}
