package interceptedBlocks

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
)

var _ process.InterceptedData = (*InterceptedMiniblock)(nil)

// InterceptedMiniblock is a wrapper over a miniblock
type InterceptedMiniblock struct {
	miniblock         *block.MiniBlock
	marshalizer       marshal.Marshalizer
	hasher            hashing.Hasher
	shardCoordinator  sharding.Coordinator
	hash              []byte
	isForCurrentShard bool
}

// NewInterceptedMiniblock creates a new instance of InterceptedMiniblock struct
func NewInterceptedMiniblock(arg *ArgInterceptedMiniblock) (*InterceptedMiniblock, error) {
	err := checkMiniblockArgument(arg)
	if err != nil {
		return nil, err
	}

	miniblock, err := createMiniblock(arg.Marshalizer, arg.MiniblockBuff)
	if err != nil {
		return nil, err
	}

	inMiniblock := &InterceptedMiniblock{
		miniblock:        miniblock,
		marshalizer:      arg.Marshalizer,
		hasher:           arg.Hasher,
		shardCoordinator: arg.ShardCoordinator,
	}
	inMiniblock.processFields(arg.MiniblockBuff)

	return inMiniblock, nil
}

func createMiniblock(marshalizer marshal.Marshalizer, miniblockBuff []byte) (*block.MiniBlock, error) {
	miniblock := &block.MiniBlock{}
	err := marshalizer.Unmarshal(miniblock, miniblockBuff)
	if err != nil {
		return nil, err
	}

	return miniblock, nil
}

func (inMb *InterceptedMiniblock) processFields(mbBuff []byte) {
	inMb.hash = inMb.hasher.Compute(string(mbBuff))

	inMb.processIsForCurrentShard()
}

func (inMb *InterceptedMiniblock) processIsForCurrentShard() {
	isForCurrentShardRecv := inMb.miniblock.ReceiverShardID == inMb.shardCoordinator.SelfId()
	isForCurrentShardSender := inMb.miniblock.SenderShardID == inMb.shardCoordinator.SelfId()
	isForAllShards := inMb.miniblock.ReceiverShardID == core.AllShardId

	inMb.isForCurrentShard = isForCurrentShardRecv || isForCurrentShardSender || isForAllShards
}

// Hash gets the hash of this transaction block body
func (inMb *InterceptedMiniblock) Hash() []byte {
	return inMb.hash
}

// Miniblock returns the miniblock held by this wrapper
func (inMb *InterceptedMiniblock) Miniblock() *block.MiniBlock {
	return inMb.miniblock
}

// CheckValidity checks if the received tx block body is valid (not nil fields)
func (inMb *InterceptedMiniblock) CheckValidity() error {
	return inMb.integrity(core.MetachainShardId)
}

// IsForCurrentShard returns true if at least one contained miniblock is for current shard
func (inMb *InterceptedMiniblock) IsForCurrentShard() bool {
	return inMb.isForCurrentShard
}

// integrity checks the integrity of the tx block body
func (inMb *InterceptedMiniblock) integrity(acceptedCrossShardId uint32) error {
	miniblock := inMb.miniblock

	receiverNotCurrentShard := miniblock.ReceiverShardID >= inMb.shardCoordinator.NumberOfShards() &&
		(miniblock.ReceiverShardID != acceptedCrossShardId && miniblock.ReceiverShardID != core.AllShardId)
	if receiverNotCurrentShard {
		return process.ErrInvalidShardId
	}

	senderNotCurrentShard := miniblock.SenderShardID >= inMb.shardCoordinator.NumberOfShards() &&
		miniblock.SenderShardID != acceptedCrossShardId
	if senderNotCurrentShard {
		return process.ErrInvalidShardId
	}

	for _, txHash := range miniblock.TxHashes {
		if txHash == nil {
			return process.ErrNilTxHash
		}
	}

	if len(miniblock.GetReserved()) > maxLenMiniBlockReservedField {
		return process.ErrReservedFieldInvalid
	}

	return nil
}

// Type returns the type of this intercepted data
func (inMb *InterceptedMiniblock) Type() string {
	return "intercepted miniblock"
}

// String returns the transactions body's most important fields as string
func (inMb *InterceptedMiniblock) String() string {
	return fmt.Sprintf("miniblock type=%s, numTxs=%d, sender shardid=%d, recv shardid=%d",
		inMb.miniblock.Type.String(),
		len(inMb.miniblock.TxHashes),
		inMb.miniblock.SenderShardID,
		inMb.miniblock.ReceiverShardID,
	)
}

// Identifiers returns the identifiers used in requests
func (inMb *InterceptedMiniblock) Identifiers() [][]byte {
	return [][]byte{inMb.hash}
}

// IsInterfaceNil returns true if there is no value under the interface
func (inMb *InterceptedMiniblock) IsInterfaceNil() bool {
	return inMb == nil
}
