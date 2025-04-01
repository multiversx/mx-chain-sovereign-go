package factory

import (
	"github.com/multiversx/mx-chain-core-go/core"

	"github.com/multiversx/mx-chain-go/sharding"
)

// CrawlerAddressGetterMock -
type CrawlerAddressGetterMock struct {
	GetAllowedAddressCalled func(coordinator sharding.Coordinator, addresses [][]byte) ([]byte, error)
}

// GetAllowedAddress -
func (mock *CrawlerAddressGetterMock) GetAllowedAddress(coordinator sharding.Coordinator, addresses [][]byte) ([]byte, error) {
	if mock.GetAllowedAddressCalled != nil {
		return mock.GetAllowedAddressCalled(coordinator, addresses)
	}

	return core.SystemAccountAddress, nil
}

// IsInterfaceNil -
func (mock *CrawlerAddressGetterMock) IsInterfaceNil() bool {
	return mock == nil
}
