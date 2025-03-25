package incomingHeader

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
	sovDto "github.com/multiversx/mx-chain-core-go/data/sovereign/dto"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"

	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/sovereign/incomingHeader/dto"
	"github.com/multiversx/mx-chain-go/process/block/sovereign/incomingHeader/extendedHeader"
)

type extendedHeaderProcessor struct {
	headersPool HeadersPool
	txPool      TransactionPool
	marshaller  marshal.Marshalizer
	hasher      hashing.Hasher
	container   EmptyBlockCreatorsContainerHandler
}

func newExtendedHeaderProcessor(
	headersPool HeadersPool,
	txPool TransactionPool,
	marshaller marshal.Marshalizer,
	hasher hashing.Hasher,
) (*extendedHeaderProcessor, error) {
	container := extendedHeader.NewEmptyBlockCreatorsContainer()
	mvxHeaderCreator, err := extendedHeader.NewEmptyHeaderV2Creator(marshaller)
	if err != nil {
		return nil, err
	}

	err = container.Add(sovDto.MVX, mvxHeaderCreator)
	if err != nil {
		return nil, err
	}
	return &extendedHeaderProcessor{
		headersPool: headersPool,
		txPool:      txPool,
		marshaller:  marshaller,
		hasher:      hasher,
		container:   container,
	}, nil
}

func (ehp *extendedHeaderProcessor) createExtendedHeader(incomingHeader sovereign.IncomingHeaderHandler, scrs []*dto.SCRInfo) (data.ShardHeaderExtendedHandler, error) {
	extendedShardHeader, err := ehp.createChainSpecificExtendedHeader(incomingHeader)
	if err != nil {
		return nil, err
	}

	// todo: here have setters on interface
	ret := extendedShardHeader.(*block.ShardHeaderExtended)

	events, err := getEvents(incomingHeader.GetIncomingEventHandlers())
	if err != nil {
		return nil, err
	}

	ret.IncomingEvents = events
	ret.IncomingMiniBlocks = createIncomingMb(scrs)

	return ret, nil
}

func (ehp *extendedHeaderProcessor) createChainSpecificExtendedHeader(incomingHeader sovereign.IncomingHeaderHandler) (data.ShardHeaderExtendedHandler, error) {
	shardExtendedHeaderCreator, err := ehp.container.Get(incomingHeader.GetSourceChainID())
	if err != nil {
		return nil, err
	}

	return shardExtendedHeaderCreator.CreateNewExtendedHeader(incomingHeader.GetProof())
}

func getEvents(events []data.EventHandler) ([]*transaction.Event, error) {
	ret := make([]*transaction.Event, len(events))

	for idx, eventHandler := range events {
		event, castOk := eventHandler.(*transaction.Event)
		if !castOk {
			return nil, errInvalidEventType
		}

		ret[idx] = event
	}

	return ret, nil
}

func createIncomingMb(scrs []*dto.SCRInfo) []*block.MiniBlock {
	if len(scrs) == 0 {
		return make([]*block.MiniBlock, 0)
	}

	scrHashes := make([][]byte, len(scrs))
	for idx, scrData := range scrs {
		scrHashes[idx] = scrData.Hash
	}

	return []*block.MiniBlock{
		{
			TxHashes:        scrHashes,
			ReceiverShardID: core.SovereignChainShardId,
			SenderShardID:   core.MainChainShardId,
			Type:            block.SmartContractResultBlock,
		},
	}
}

func (ehp *extendedHeaderProcessor) addPreGenesisExtendedHeaderToPool(incomingHeader sovereign.IncomingHeaderHandler) error {
	extendedShardHeader, err := ehp.createChainSpecificExtendedHeader(incomingHeader)
	if err != nil {
		return err
	}

	return ehp.addExtendedHeaderAndSCRsToPool(extendedShardHeader, make([]*dto.SCRInfo, 0))
}

func (ehp *extendedHeaderProcessor) addExtendedHeaderAndSCRsToPool(extendedHeader data.ShardHeaderExtendedHandler, scrs []*dto.SCRInfo) error {
	extendedHeaderHash, err := core.CalculateHash(ehp.marshaller, ehp.hasher, extendedHeader)
	if err != nil {
		return err
	}

	ehp.addSCRsToPool(scrs)
	ehp.headersPool.AddHeaderInShard(extendedHeaderHash, extendedHeader, core.MainChainShardId)
	return nil
}

func (ehp *extendedHeaderProcessor) addSCRsToPool(scrs []*dto.SCRInfo) {
	cacheID := process.ShardCacherIdentifier(core.MainChainShardId, core.SovereignChainShardId)

	for _, scrData := range scrs {
		ehp.txPool.AddData(scrData.Hash, scrData.SCR, scrData.SCR.Size(), cacheID)
	}
}
