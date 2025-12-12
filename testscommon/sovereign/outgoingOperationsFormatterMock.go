package sovereign

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
)

// OutgoingOperationsFormatterMock -
type OutgoingOperationsFormatterMock struct {
	CreateOutgoingTxDataCalled              func(logs []*data.LogData) (map[block.OutGoingMBType][][]byte, error)
	CreateOutGoingChangeValidatorDataCalled func(pubKeys []string, epoch uint32) ([]byte, error)
}

// CreateOutgoingTxsData -
func (stub *OutgoingOperationsFormatterMock) CreateOutgoingTxsData(logs []*data.LogData) (map[block.OutGoingMBType][][]byte, error) {
	if stub.CreateOutgoingTxDataCalled != nil {
		return stub.CreateOutgoingTxDataCalled(logs)
	}

	return make(map[block.OutGoingMBType][][]byte), nil
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
