package factory

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/interceptedBlocks"
)

type interceptedSovereignShardHeaderDataFactory struct {
	*interceptedShardHeaderDataFactory
}

// NewInterceptedSovereignShardHeaderDataFactory creates a sovereign shard header data factory
func NewInterceptedSovereignShardHeaderDataFactory(argument *ArgInterceptedDataFactory) (*interceptedSovereignShardHeaderDataFactory, error) {
	baseInterceptedShardHeaderFactory, err := NewInterceptedShardHeaderDataFactory(argument)
	if err != nil {
		return nil, err
	}

	return &interceptedSovereignShardHeaderDataFactory{
		baseInterceptedShardHeaderFactory,
	}, nil
}

// Create creates instances of InterceptedData by unmarshalling provided buffer
func (ishdf *interceptedSovereignShardHeaderDataFactory) Create(buff []byte, _ core.PeerID) (process.InterceptedData, error) {
	arg := ishdf.createArgsInterceptedBlockHeader(buff)
	return interceptedBlocks.NewSovereignInterceptedBlockHeader(arg)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ishdf *interceptedSovereignShardHeaderDataFactory) IsInterfaceNil() bool {
	return ishdf == nil
}
