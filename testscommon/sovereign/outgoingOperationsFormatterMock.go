package sovereign

import (
	"github.com/multiversx/mx-chain-core-go/data"
	dtoCore "github.com/multiversx/mx-chain-core-go/data/sovereign/dto"
	"github.com/multiversx/mx-chain-go/process/block/sovereign/incomingHeader/dto"
)

// OutgoingOperationsFormatterMock -
type OutgoingOperationsFormatterMock struct {
	CreateOutgoingTxDataCalled              func(logs []*data.LogData) (map[dtoCore.ChainID][]*dto.OutGoingOperation, error)
	CreateOutGoingChangeValidatorDataCalled func(pubKeys []string, epoch uint32) ([]byte, error)
}

// CreateOutgoingTxsData -
func (stub *OutgoingOperationsFormatterMock) CreateOutgoingTxsData(logs []*data.LogData) (map[dtoCore.ChainID][]*dto.OutGoingOperation, error) {
	if stub.CreateOutgoingTxDataCalled != nil {
		return stub.CreateOutgoingTxDataCalled(logs)
	}

	return make(map[dtoCore.ChainID][]*dto.OutGoingOperation), nil
}

// CreateOutGoingChangeValidatorData -
func (stub *OutgoingOperationsFormatterMock) CreateOutGoingChangeValidatorData(pubKeys []string, epoch uint32) ([]byte, error) {
	if stub.CreateOutGoingChangeValidatorDataCalled != nil {
		return stub.CreateOutGoingChangeValidatorDataCalled(pubKeys, epoch)
	}

	return make([]byte, 0), nil
}

// IsInterfaceNil -
func (stub *OutgoingOperationsFormatterMock) IsInterfaceNil() bool {
	return stub == nil
}
