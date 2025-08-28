package bootstrap

import (
	"github.com/multiversx/mx-chain-core-go/core"
	dataCore "github.com/multiversx/mx-chain-core-go/data"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process/factory"
)

type sovereignTopicProvider struct {
}

// internal constructor
func newSovereignTopicProvider() *sovereignTopicProvider {
	return &sovereignTopicProvider{}
}

func (e *sovereignTopicProvider) getTopic() string {
	return factory.ShardBlocksTopic + core.CommunicationIdentifierBetweenShards(core.SovereignChainShardId, core.SovereignChainShardId)
}

func (e *sovereignTopicProvider) getProofsTopic(_ uint32, _ uint32) string {
	return common.EquivalentProofsTopic + core.CommunicationIdentifierBetweenShards(core.SovereignChainShardId, core.SovereignChainShardId)
}

type epochStartSovereignBlockProcessor struct {
	*sovereignTopicProvider
	*epochStartMetaBlockProcessor
}

// internal constructor
func newEpochStartSovereignBlockProcessor(epochStartMetaBlockProcessor *epochStartMetaBlockProcessor) *epochStartSovereignBlockProcessor {
	sovBlockProc := &epochStartSovereignBlockProcessor{
		epochStartMetaBlockProcessor: epochStartMetaBlockProcessor,
		sovereignTopicProvider:       &sovereignTopicProvider{},
	}

	sovBlockProc.epochStartPeerHandler = sovBlockProc
	return sovBlockProc
}

func (e *epochStartSovereignBlockProcessor) setNumPeers(
	requestHandler RequestHandler,
	intra int, _ int,
) error {
	return requestHandler.SetNumPeersToQuery(e.sovereignTopicProvider.getTopic(), intra, 0)
}

func (e *epochStartSovereignBlockProcessor) getTopic() string {
	return e.sovereignTopicProvider.getTopic()
}

func (e *epochStartSovereignBlockProcessor) requestProofForMetaBlock(metablockHash []byte) error {
	numConnectedPeers := len(e.messenger.ConnectedPeers())
	topic := common.EquivalentProofsTopic + core.CommunicationIdentifierBetweenShards(core.SovereignChainShardId, core.SovereignChainShardId)
	err := e.requestHandler.SetNumPeersToQuery(topic, numConnectedPeers, 0)
	if err != nil {
		return err
	}

	e.requestHandler.RequestEquivalentProofByHash(core.SovereignChainShardId, metablockHash)
	return nil
}

func (e *epochStartSovereignBlockProcessor) getMetaChainShardID() uint32 {
	return core.SovereignChainShardId
}

func (e *epochStartSovereignBlockProcessor) hashMatches(hash string, proof dataCore.HeaderProofHandler) bool {
	return string(proof.GetProcessedHeaderHash()) == hash
}
