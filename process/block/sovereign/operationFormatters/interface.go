package operationFormatters

import "github.com/multiversx/mx-chain-core-go/data/sovereign"

// DataCodecHandler is the interface for serializing/deserializing data
type DataCodecHandler interface {
	DeserializeTokenData(data []byte) (*sovereign.EsdtTokenData, error)
	SerializeOperation(operation sovereign.Operation) ([]byte, error)
	IsInterfaceNil() bool
}
