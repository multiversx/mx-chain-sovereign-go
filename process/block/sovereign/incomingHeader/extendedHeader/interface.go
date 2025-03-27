package extendedHeader

import (
	"github.com/multiversx/mx-chain-core-go/data"
)

// EmptyExtendedHeaderCreator is able to create empty extended header instances
type EmptyExtendedHeaderCreator interface {
	CreateNewExtendedHeader(proof []byte) (data.ShardHeaderExtendedHandler, error)
	IsInterfaceNil() bool
}
