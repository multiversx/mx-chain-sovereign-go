package epochProviders

import (
	"time"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/resolvers/epochproviders"
	"github.com/multiversx/mx-chain-go/dataRetriever/resolvers/epochproviders/disabled"
)

type currEpochProviderFactory struct {
}

func NewCurrentEpochProviderFactory() *currEpochProviderFactory {
	return &currEpochProviderFactory{}
}

// CreateCurrentEpochProvider will create an instance of dataRetriever.CurrentNetworkEpochProviderHandler using
// unix seconds time handlers
func (f *currEpochProviderFactory) CreateCurrentEpochProvider(
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
		StartTime:               startTime.Unix(),
		GetUnixHandler: func() int64 {
			return time.Now().Unix()
		},
	}

	return epochproviders.NewArithmeticEpochProvider(arg)
}

// IsInterfaceNil checks if the underlying interface is nil
func (f *currEpochProviderFactory) IsInterfaceNil() bool {
	return f == nil
}
