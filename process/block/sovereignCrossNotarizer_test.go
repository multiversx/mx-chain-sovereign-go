package block

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-sovereign-go/process/block/bootstrapStorage"
	"github.com/multiversx/mx-chain-sovereign-go/testscommon"
)

func TestSovereignShardCrossNotarizer_getLastCrossNotarizedHeaders(t *testing.T) {
	hash := []byte("hash")
	header := &block.SovereignChainHeader{
		Header: &block.Header{
			ShardID: core.SovereignChainShardId,
			Nonce:   4,
		},
	}
	sovereignNotarzier := &sovereignShardCrossNotarizer{
		&baseBlockNotarizer{
			blockTracker: &testscommon.BlockTrackerStub{
				GetLastCrossNotarizedHeaderCalled: func(shardID uint32) (data.HeaderHandler, []byte, error) {
					require.Equal(t, core.MainChainShardId, shardID)
					return header, hash, nil
				},
			},
		},
	}

	headers := sovereignNotarzier.getLastCrossNotarizedHeaders()
	expectedHeaders := []bootstrapStorage.BootstrapHeaderInfo{
		{
			ShardId: core.MainChainShardId,
			Nonce:   header.GetNonce(),
			Hash:    hash,
		},
	}
	require.Equal(t, expectedHeaders, headers)
}
