package block

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/sovereign/dto"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/process/block/bootstrapStorage"
	"github.com/multiversx/mx-chain-go/testscommon"
)

func TestSovereignShardCrossNotarizer_getLastCrossNotarizedHeaders(t *testing.T) {
	hash := []byte("hash")
	header := &block.ShardHeaderExtended{
		SourceChainID: dto.MVX,
		Header: &block.HeaderV2{
			Header: &block.Header{
				Nonce: 4,
			},
		},
	}
	sovereignNotarzier := &sovereignShardCrossNotarizer{
		&baseBlockNotarizer{
			blockTracker: &testscommon.BlockTrackerStub{
				GetLastCrossNotarizedHeaderCalled: func(shardID uint32) (data.HeaderHandler, []byte, error) {
					switch shardID {
					case uint32(dto.MVX):
						return header, hash, nil
					}
					return nil, nil, errors.New("not found")
				},
			},
		},
	}

	headers := sovereignNotarzier.getLastCrossNotarizedHeaders()
	expectedHeaders := []bootstrapStorage.BootstrapHeaderInfo{
		{
			ShardId: uint32(dto.MVX),
			Nonce:   header.GetNonce(),
			Hash:    hash,
		},
	}
	require.Equal(t, expectedHeaders, headers)
}
