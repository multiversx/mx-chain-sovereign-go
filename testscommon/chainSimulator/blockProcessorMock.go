package chainSimulator

import (
	sovereignBlock "github.com/multiversx/mx-chain-go/dataRetriever/dataPool/sovereign"
	"github.com/multiversx/mx-chain-go/process"

	"github.com/multiversx/mx-chain-core-go/data"
)

// BlockProcessorMock -
type BlockProcessorMock struct {
	ProcessBlockCalled       func(blockProcessor process.BlockProcessor, header data.HeaderHandler) (data.HeaderHandler, data.BodyHandler, error)
	ProcessHeaderProofCalled func(
		header data.HeaderHandler,
		proof data.HeaderProofHandler,
		outGoingOperationsPool sovereignBlock.ShardedOutGoingOperationPool,
	) error
}

// ProcessBlock -
func (mock *BlockProcessorMock) ProcessBlock(blockProcessor process.BlockProcessor, header data.HeaderHandler) (data.HeaderHandler, data.BodyHandler, error) {
	if mock.ProcessBlockCalled != nil {
		return mock.ProcessBlockCalled(blockProcessor, header)
	}
	return nil, nil, nil
}

// ProcessHeaderProof -
func (mock *BlockProcessorMock) ProcessHeaderProof(
	header data.HeaderHandler,
	proof data.HeaderProofHandler,
	outGoingOperationsPool sovereignBlock.ShardedOutGoingOperationPool,
) error {
	if mock.ProcessHeaderProofCalled != nil {
		return mock.ProcessHeaderProofCalled(header, proof, outGoingOperationsPool)
	}
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (mock *BlockProcessorMock) IsInterfaceNil() bool {
	return mock == nil
}
