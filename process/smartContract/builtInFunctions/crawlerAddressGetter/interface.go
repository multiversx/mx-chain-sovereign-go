package crawlerAddressGetter

import (
	"github.com/multiversx/mx-chain-go/sharding"
)

// CrawlerAddressGetterHandler defines a handler to get crawler addresses
type CrawlerAddressGetterHandler interface {
	GetAllowedAddress(coordinator sharding.Coordinator, addresses [][]byte) ([]byte, error)
	IsInterfaceNil() bool
}
