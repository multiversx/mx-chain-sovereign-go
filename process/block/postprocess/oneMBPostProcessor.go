package postprocess

import (
	"bytes"
	"sort"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"

	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
)

var _ process.IntermediateTransactionHandler = (*oneMBPostProcessor)(nil)

type oneMBPostProcessor struct {
	blockType block.Type
	*basePostProcessor
}

// NewOneMiniBlockPostProcessor creates a new intermediate results processor
func NewOneMiniBlockPostProcessor(
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	coordinator sharding.Coordinator,
	store dataRetriever.StorageService,
	blockType block.Type,
	storageType dataRetriever.UnitType,
	economicsFee process.FeeHandler,
) (*oneMBPostProcessor, error) {
	if check.IfNil(hasher) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(marshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(coordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	if check.IfNil(store) {
		return nil, process.ErrNilStorage
	}
	if check.IfNil(economicsFee) {
		return nil, process.ErrNilEconomicsFeeHandler
	}

	base := &basePostProcessor{
		hasher:             hasher,
		marshalizer:        marshalizer,
		shardCoordinator:   coordinator,
		store:              store,
		storageType:        storageType,
		mapProcessedResult: make(map[string]*processedResult),
		economicsFee:       economicsFee,
	}

	opp := &oneMBPostProcessor{
		basePostProcessor: base,
		blockType:         blockType,
	}

	opp.interResultsForBlock = make(map[string]*txInfo)

	return opp, nil
}

// GetNumOfCrossInterMbsAndTxs returns the number of cross shard miniblocks and transactions for the current round,
// created from the receipts/bad transactions
func (opp *oneMBPostProcessor) GetNumOfCrossInterMbsAndTxs() (int, int) {
	return 0, 0
}

// CreateAllInterMiniBlocks returns the miniblock for the current round created from the receipts/bad transactions
func (opp *oneMBPostProcessor) CreateAllInterMiniBlocks() []*block.MiniBlock {
	selfId := opp.shardCoordinator.SelfId()

	miniBlocks := make([]*block.MiniBlock, 0)
	opp.mutInterResultsForBlock.Lock()
	defer opp.mutInterResultsForBlock.Unlock()

	if len(opp.interResultsForBlock) == 0 {
		return miniBlocks
	}

	miniBlock := &block.MiniBlock{
		Type:            opp.blockType,
		ReceiverShardID: selfId,
		SenderShardID:   selfId,
	}

	for key := range opp.interResultsForBlock {
		miniBlock.TxHashes = append(miniBlock.TxHashes, []byte(key))
	}

	sort.Slice(miniBlock.TxHashes, func(a, b int) bool {
		return bytes.Compare(miniBlock.TxHashes[a], miniBlock.TxHashes[b]) < 0
	})

	log.Debug("oneMBPostProcessor.CreateAllInterMiniBlocks",
		"type", miniBlock.Type,
		"senderShardID", miniBlock.SenderShardID,
		"receiverShardID", miniBlock.ReceiverShardID,
	)

	for _, hash := range miniBlock.TxHashes {
		log.Trace("tx", "hash", hash)
	}
	miniBlocks = append(miniBlocks, miniBlock)
	opp.intraShardMiniBlock = miniBlock.DeepClone()

	return miniBlocks
}

// VerifyInterMiniBlocks verifies if the receipts/bad transactions added to the block are valid
func (opp *oneMBPostProcessor) VerifyInterMiniBlocks(body *block.Body) error {
	scrMbs := opp.CreateAllInterMiniBlocks()
	createdMapMbs := createMiniBlocksMap(scrMbs)

	verifiedOne := false
	for i := 0; i < len(body.MiniBlocks); i++ {
		mb := body.MiniBlocks[i]
		if mb.Type != opp.blockType {
			continue
		}

		if verifiedOne {
			return process.ErrTooManyReceiptsMiniBlocks
		}

		err := opp.verifyMiniBlock(createdMapMbs, mb)
		if err != nil {
			return err
		}

		verifiedOne = true
	}

	return nil
}

// AddIntermediateTransactions adds receipts/bad transactions resulting from transaction processor
func (opp *oneMBPostProcessor) AddIntermediateTransactions(txs []data.TransactionHandler, key []byte) error {
	opp.mutInterResultsForBlock.Lock()
	defer opp.mutInterResultsForBlock.Unlock()

	selfId := opp.shardCoordinator.SelfId()

	for i := 0; i < len(txs); i++ {
		txHash, err := core.CalculateHash(opp.marshalizer, opp.hasher, txs[i])
		if err != nil {
			return err
		}

		opp.addIntermediateTxToResultsForBlock(txs[i], txHash, selfId, selfId, key)
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (opp *oneMBPostProcessor) IsInterfaceNil() bool {
	return opp == nil
}
