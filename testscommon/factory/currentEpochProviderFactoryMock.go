package factory

import (
	"time"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	retrieverMock "github.com/multiversx/mx-chain-go/dataRetriever/mock"
)

// CurrEpochProviderFactoryMock -
type CurrEpochProviderFactoryMock struct {
	CreateCurrentEpochProviderCalled func(
		generalConfigs config.Config,
		roundTimeInMilliseconds uint64,
		startTime time.Time,
		isFullArchive bool,
	) (dataRetriever.CurrentNetworkEpochProviderHandler, error)
}

// CreateCurrentEpochProvider -
func (mock *CurrEpochProviderFactoryMock) CreateCurrentEpochProvider(
	generalConfigs config.Config,
	roundTimeInMilliseconds uint64,
	startTime time.Time,
	isFullArchive bool,
) (dataRetriever.CurrentNetworkEpochProviderHandler, error) {
	if mock.CreateCurrentEpochProviderCalled != nil {
		return mock.CreateCurrentEpochProviderCalled(generalConfigs, roundTimeInMilliseconds, startTime, isFullArchive)
	}

	return &retrieverMock.CurrentNetworkEpochProviderStub{}, nil
}

// IsInterfaceNil -
func (f *CurrEpochProviderFactoryMock) IsInterfaceNil() bool {
	return f == nil
}
