package operationFormatters

import "errors"

var errInvalidNumTopicsInRegisterToken = errors.New("invalid num topics in event for register token")

var errInvalidNumTopicsInRegisterValidator = errors.New("invalid num topics in event for register validator")

var errNilNonceChainHandler = errors.New("nil nonce chain handler")

var errChainIDNotSupported = errors.New("chain ID not supported")
