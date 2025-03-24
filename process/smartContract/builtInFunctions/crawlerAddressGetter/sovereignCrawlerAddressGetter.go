package crawlerAddressGetter

import (
	"github.com/multiversx/mx-chain-core-go/core"

	"github.com/multiversx/mx-chain-go/sharding"
)

type sovereignCrawlerAddressGetter struct {
}

// NewSovereignCrawlerAddressGetter creates a crawler address getter for sovereign run type
func NewSovereignCrawlerAddressGetter() *sovereignCrawlerAddressGetter {
	return &sovereignCrawlerAddressGetter{}
}

// GetAllowedAddress returns the allowed crawler address on the current shard
func (scag *sovereignCrawlerAddressGetter) GetAllowedAddress(_ sharding.Coordinator, addresses [][]byte) ([]byte, error) {

	if len(addresses) != 0 {
		log.Debug("found automatic crawler addresses set in sovereign config, these will not be used")
	}

	return core.SystemAccountAddress, nil
}

// IsInterfaceNil checks if the underlying pointer is nil
func (scag *sovereignCrawlerAddressGetter) IsInterfaceNil() bool {
	return scag == nil
}
