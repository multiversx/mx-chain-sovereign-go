package crawlerAddressGetter

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	logger "github.com/multiversx/mx-chain-logger-go"

	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
)

var log = logger.GetOrCreate("crawlerAddressGetter")

type crawlerAddressGetter struct {
}

// NewCrawlerAddressGetter creates a crawler address getter for normal run type
func NewCrawlerAddressGetter() *crawlerAddressGetter {
	return &crawlerAddressGetter{}
}

// GetAllowedAddress returns system account address
func (cag *crawlerAddressGetter) GetAllowedAddress(coordinator sharding.Coordinator, addresses [][]byte) ([]byte, error) {
	if check.IfNil(coordinator) {
		return nil, process.ErrNilShardCoordinator
	}

	if len(addresses) == 0 {
		return nil, fmt.Errorf("%w for shard %d, provided count is %d", process.ErrNilCrawlerAllowedAddress, coordinator.SelfId(), len(addresses))
	}

	if coordinator.SelfId() == core.MetachainShardId {
		return core.SystemAccountAddress, nil
	}

	for _, address := range addresses {
		allowedAddressShardId := coordinator.ComputeId(address)
		if allowedAddressShardId == coordinator.SelfId() {
			return address, nil
		}
	}

	return nil, fmt.Errorf("%w for shard %d, provided count is %d", process.ErrNilCrawlerAllowedAddress, coordinator.SelfId(), len(addresses))
}

// IsInterfaceNil checks if the underlying pointer is nil
func (cag *crawlerAddressGetter) IsInterfaceNil() bool {
	return cag == nil
}
