package incomingHeader

import "errors"

var errNilHeadersPool = errors.New("nil headers pool provided")

var errNilTxPool = errors.New("nil tx pool provided")

var errInvalidEventType = errors.New("incoming event is not of type transaction event")

var errNilProof = errors.New("nil proof in incoming header")

var errSourceChainNotSupported = errors.New("source chain id from configuration is not supported")
