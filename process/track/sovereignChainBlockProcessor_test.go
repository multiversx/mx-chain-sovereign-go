package track_test

import (
	"errors"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/process/track"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewSovereignChainBlockProcessor_ShouldErrNilBlockProcessor(t *testing.T) {
	t.Parallel()

	scpb, err := track.NewSovereignChainBlockProcessor(nil)
	assert.Nil(t, scpb)
	assert.Equal(t, process.ErrNilBlockProcessor, err)
}

func TestNewSovereignChainBlockProcessor_ShouldWork(t *testing.T) {
	t.Parallel()

	blockProcessorArguments := CreateBlockProcessorMockArguments()
	bp, _ := track.NewBlockProcessor(blockProcessorArguments)

	scpb, err := track.NewSovereignChainBlockProcessor(bp)
	assert.NotNil(t, scpb)
	assert.Nil(t, err)
}

func TestSovereignChainBlockProcessor_ShouldProcessReceivedHeaderShouldWork(t *testing.T) {
	t.Parallel()

	header := &block.Header{ShardID: 1}
	blockProcessorArguments := CreateBlockProcessorMockArguments()
	blockProcessorArguments.SelfNotarizer = &mock.BlockNotarizerHandlerMock{
		GetLastNotarizedHeaderCalled: func(shardID uint32) (data.HeaderHandler, []byte, error) {
			if shardID != header.GetShardID() {
				return nil, nil, errors.New("wrong shard ID")
			}
			return &block.Header{Nonce: 499}, []byte("hash"), nil
		},
	}
	bp, _ := track.NewBlockProcessor(blockProcessorArguments)
	scbp, _ := track.NewSovereignChainBlockProcessor(bp)

	header.Nonce = 499
	assert.False(t, scbp.ShouldProcessReceivedHeader(header))

	header.Nonce = 500
	assert.True(t, scbp.ShouldProcessReceivedHeader(header))
}
