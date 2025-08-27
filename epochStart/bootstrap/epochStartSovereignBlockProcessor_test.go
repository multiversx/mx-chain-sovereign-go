package bootstrap

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
)

func createSovEpochStartBlockProcessor() *epochStartSovereignBlockProcessor {
	esmbp, _ := NewEpochStartMetaBlockProcessor(
		&p2pmocks.MessengerStub{
			ConnectedPeersCalled: func() []core.PeerID {
				return []core.PeerID{"peer_0", "peer_1"}
			},
		},
		&testscommon.RequestHandlerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		99,
		2,
		2,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&dataRetriever.ProofsPoolMock{},
	)
	return newEpochStartSovereignBlockProcessor(esmbp)
}

func TestEpochStartSovereignBlockProcessor_setNumPeers(t *testing.T) {
	t.Parallel()

	numIntra := 4
	numCross := 3
	wasNumPeersSet := false
	requestHandler := &testscommon.RequestHandlerStub{
		SetNumPeersToQueryCalled: func(key string, intra int, cross int) error {
			require.Equal(t, fmt.Sprintf("%s_%d", factory.ShardBlocksTopic, core.SovereignChainShardId), key)
			require.Equal(t, numIntra, intra)
			require.Zero(t, cross)

			wasNumPeersSet = true
			return nil
		},
	}
	sovProc := createSovEpochStartBlockProcessor()
	err := sovProc.setNumPeers(requestHandler, numIntra, numCross)
	require.Nil(t, err)
	require.True(t, wasNumPeersSet)
}

func TestEpochStartSovereignBlockProcessor_getRequestTopic(t *testing.T) {
	t.Parallel()

	sovProc := createSovEpochStartBlockProcessor()
	require.Equal(t, fmt.Sprintf("%s_%d", factory.ShardBlocksTopic, core.SovereignChainShardId), sovProc.getTopic())
}

func TestEpochStartSovereignBlockProcessor_getProofsTopic(t *testing.T) {
	t.Parallel()

	sovProc := createSovEpochStartBlockProcessor()
	require.Equal(t,
		fmt.Sprintf("%s_%d", common.EquivalentProofsTopic, core.SovereignChainShardId),
		sovProc.getProofsTopic(core.MetachainShardId, 4442),
	)
}

func TestEpochStartSovereignBlockProcessor_requestProofForMetaBlock(t *testing.T) {
	t.Parallel()

	wasNumPeersSet := false
	wasProofRequested := false
	metaBlockHash := []byte("blockHash")

	sovProc := createSovEpochStartBlockProcessor()

	requestHandler := &testscommon.RequestHandlerStub{
		SetNumPeersToQueryCalled: func(key string, intra int, cross int) error {
			require.Equal(t, fmt.Sprintf("%s_%d", common.EquivalentProofsTopic, core.SovereignChainShardId), key)
			require.Equal(t, len(sovProc.messenger.ConnectedPeers()), intra)
			require.Zero(t, cross)

			wasNumPeersSet = true
			return nil
		},
		RequestEquivalentProofByHashCalled: func(headerShard uint32, headerHash []byte) {
			require.Equal(t, metaBlockHash, headerHash)
			require.Equal(t, core.SovereignChainShardId, headerShard)

			wasProofRequested = true
		},
	}

	sovProc.requestHandler = requestHandler

	err := sovProc.requestProofForMetaBlock(metaBlockHash)
	require.Nil(t, err)
	require.True(t, wasNumPeersSet)
	require.True(t, wasProofRequested)
}

func TestEpochStartSovereignBlockProcessor_getMetaChainShardID(t *testing.T) {
	t.Parallel()

	sovProc := createSovEpochStartBlockProcessor()
	require.Equal(t, core.SovereignChainShardId, sovProc.getMetaChainShardID())
}

func TestEpochStartSovereignBlockProcessor_hashMatches(t *testing.T) {
	t.Parallel()

	sovProc := createSovEpochStartBlockProcessor()

	headerHash := []byte("headerHash")
	processedHeaderHash := []byte("processedHeaderHash")
	proof := &block.HeaderProof{
		HeaderHash:          headerHash,
		ProcessedHeaderHash: processedHeaderHash,
	}

	require.False(t, sovProc.hashMatches(string(headerHash), proof))
	require.True(t, sovProc.hashMatches(string(processedHeaderHash), proof))
}
