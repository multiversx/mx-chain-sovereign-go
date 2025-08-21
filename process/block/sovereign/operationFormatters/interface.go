package operationFormatters

import (
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/multiversx/mx-chain-go/process/block/sovereign/dto"
)

// DataCodecHandler is the interface for serializing/deserializing data
type DataCodecHandler interface {
	DeserializeTokenData(data []byte) (*sovereign.EsdtTokenData, error)
	SerializeOperation(operation sovereign.Operation) ([]byte, error)
	SerializeTokenProperties(properties dto.TokenProperties) ([]byte, error)
	SerializeNewlyRegisteredKey(keyData dto.RegisteredBlsKey) ([]byte, error)
	IsInterfaceNil() bool
}
