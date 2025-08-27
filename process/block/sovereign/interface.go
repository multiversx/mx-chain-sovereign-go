package sovereign

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
	coreDTO "github.com/multiversx/mx-chain-core-go/data/sovereign/dto"
	"github.com/multiversx/mx-chain-go/process/block/sovereign/dto"
)

// OutgoingOperationsFormatter collects relevant outgoing events for bridge from the logs and creates outgoing data
// that needs to be signed by validators to bridge tokens
type OutgoingOperationsFormatter interface {
	CreateOutgoingTxsData(logs []*data.LogData) (map[coreDTO.ChainID][][]byte, error)
	CreateOutGoingChangeValidatorData(pubKeys []string, epoch uint32) (map[coreDTO.ChainID][][]byte, error)
	IsInterfaceNil() bool
}

// DataCodecHandler is the interface for serializing/deserializing data
type DataCodecHandler interface {
	SerializeEventData(eventData sovereign.EventData) ([]byte, error)
	DeserializeEventData(data []byte) (*sovereign.EventData, error)
	SerializeTokenData(tokenData sovereign.EsdtTokenData) ([]byte, error)
	DeserializeTokenData(data []byte) (*sovereign.EsdtTokenData, error)
	SerializeOperation(operation sovereign.Operation) ([]byte, error)
	SerializeTokenProperties(properties dto.TokenProperties) ([]byte, error)
	SerializeNewlyRegisteredKey(keyData dto.RegisteredBlsKey) ([]byte, error)
	IsInterfaceNil() bool
}

// TopicsCheckerHandler should be able to check the topics validity
type TopicsCheckerHandler interface {
	CheckValidity(topics [][]byte, transferData *sovereign.TransferData) error
	IsInterfaceNil() bool
}

// OperationFormatter defines an operation formatter(like deposit tokens)
type OperationFormatter interface {
	CreateOperationData(event data.EventHandler) ([]byte, error)
	IsInterfaceNil() bool
}
