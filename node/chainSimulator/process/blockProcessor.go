package process

import (
	sovereignBlock "github.com/multiversx/mx-chain-go/dataRetriever/dataPool/sovereign"
	"github.com/multiversx/mx-chain-go/process"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
)

type blockProcessor struct {
}

// NewBlockProcessor creates a new block processor for normal chain simulator
func NewBlockProcessor() BlocksProcessor {
	return &blockProcessor{}
}

// ProcessBlock will create the block in normal chain simulator
func (bpf *blockProcessor) ProcessBlock(blockProcessor process.BlockProcessor, header data.HeaderHandler) (data.HeaderHandler, data.BodyHandler, error) {
	if check.IfNil(blockProcessor) {
		return nil, nil, process.ErrNilBlockProcessor
	}
	if check.IfNil(header) {
		return nil, nil, process.ErrNilHeaderHandler
	}

	return blockProcessor.CreateBlock(header, func() bool {
		return true
	})
}

// ProcessHeaderProof does nothing for normal run type
func (bpf *blockProcessor) ProcessHeaderProof(
	_ data.HeaderHandler,
	_ data.HeaderProofHandler,
	_ sovereignBlock.ShardedOutGoingOperationPool,
) error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (bpf *blockProcessor) IsInterfaceNil() bool {
	return bpf == nil
}
