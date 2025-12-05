package extendedHeader

import (
	"errors"
)

var errNilEmptyExtendedHeaderCreator = errors.New("nil empty extended header creator")

var errChainIDNotFound = errors.New("chain id not found for empty extended header creator")
