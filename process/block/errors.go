package block

import (
	"errors"
)

var errOutGoingBlockHashMismatch = errors.New("outgoing miniblock hash in sovereign header mismatch")
