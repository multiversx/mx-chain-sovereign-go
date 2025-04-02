package epochProviders

import (
	"time"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/resolvers/epochproviders"
	"github.com/multiversx/mx-chain-go/dataRetriever/resolvers/epochproviders/disabled"
)

type sovCurrEpochProviderFactory struct {
}

// NewSovereignCurrentEpochProviderFactory creates a sovereign current epoch provider factory
func NewSovereignCurrentEpochProviderFactory() *sovCurrEpochProviderFactory {
	return &sovCurrEpochProviderFactory{}
}

// CreateCurrentEpochProvider will create an instance of dataRetriever.CurrentNetworkEpochProviderHandler using
// unix milliseconds time handlers
func (f *sovCurrEpochProviderFactory) CreateCurrentEpochProvider(
	generalConfigs config.Config,
	roundTimeInMilliseconds uint64,
	startTime time.Time,
	isFullArchive bool,
) (dataRetriever.CurrentNetworkEpochProviderHandler, error) {
	if !isFullArchive {
		return disabled.NewEpochProvider(), nil
	}

	arg := epochproviders.ArgArithmeticEpochProvider{
		RoundsPerEpoch:          uint32(generalConfigs.EpochStartConfig.RoundsPerEpoch),
		RoundTimeInMilliseconds: roundTimeInMilliseconds,
		StartTime:               startTime.UnixMilli(),
		GetUnixHandler: func() int64 {
			return time.Now().UnixMilli()
		},
	}

	return epochproviders.NewArithmeticEpochProvider(arg)
}

// IsInterfaceNil checks if the underlying interface is nil
func (f *sovCurrEpochProviderFactory) IsInterfaceNil() bool {
	return f == nil
}
