package epochProviders

import (
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever/resolvers/epochproviders"
	"github.com/multiversx/mx-chain-go/dataRetriever/resolvers/epochproviders/disabled"
)

func TestCreateCurrentEpochProvider_NilCurrentEpochProvider(t *testing.T) {
	t.Parallel()

	f := NewCurrentEpochProviderFactory()
	require.False(t, f.IsInterfaceNil())

	cnep, err := f.CreateCurrentEpochProvider(
		config.Config{},
		0,
		time.Unix(0, 0),
		false,
	)

	assert.Nil(t, err)
	assert.IsType(t, disabled.NewEpochProvider(), cnep)
}

func TestCreateCurrentEpochProvider_ArithmeticEpochProvider(t *testing.T) {
	t.Parallel()

	f := NewCurrentEpochProviderFactory()
	cnep, err := f.CreateCurrentEpochProvider(
		config.Config{
			EpochStartConfig: config.EpochStartConfig{
				RoundsPerEpoch: 1,
			},
		},
		1,
		time.Unix(1, 0),
		true,
	)
	require.Nil(t, err)

	aep, _ := epochproviders.NewArithmeticEpochProvider(
		epochproviders.ArgArithmeticEpochProvider{
			RoundsPerEpoch:          1,
			RoundTimeInMilliseconds: 1,
			StartTime:               1,
			GetUnixHandler: func() int64 {
				return time.Now().Unix()
			},
		},
	)
	require.False(t, check.IfNil(aep))
	assert.IsType(t, aep, cnep)
}
