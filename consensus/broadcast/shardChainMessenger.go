package broadcast

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/process/factory"
)

const validatorDelayPerOrder = time.Second

var _ consensus.BroadcastMessenger = (*shardChainMessenger)(nil)

type dataToBroadcast struct {
	marshalledHeader []byte
	marshalledBody   []byte
}

type shardChainMessenger struct {
	*commonMessenger
}

// ShardChainMessengerArgs holds the arguments for creating a shardChainMessenger instance
type ShardChainMessengerArgs struct {
	CommonMessengerArgs
}

// NewShardChainMessenger creates a new shardChainMessenger object
func NewShardChainMessenger(
	args ShardChainMessengerArgs,
) (*shardChainMessenger, error) {

	err := checkShardChainNilParameters(args)
	if err != nil {
		return nil, err
	}

	cm := &commonMessenger{
		marshalizer:          args.Marshalizer,
		hasher:               args.Hasher,
		messenger:            args.Messenger,
		shardCoordinator:     args.ShardCoordinator,
		peerSignatureHandler: args.PeerSignatureHandler,
		keysHandler:          args.KeysHandler,
	}

	cm.broadcasterFilterHandler = cm

	dbbArgs := &ArgsDelayedBlockBroadcaster{
		InterceptorsContainer: args.InterceptorsContainer,
		HeadersSubscriber:     args.HeadersSubscriber,
		LeaderCacheSize:       args.MaxDelayCacheSize,
		ValidatorCacheSize:    args.MaxValidatorDelayCacheSize,
		ShardCoordinator:      args.ShardCoordinator,
		AlarmScheduler:        args.AlarmScheduler,
	}

	dbb, err := NewDelayedBlockBroadcaster(dbbArgs)
	if err != nil {
		return nil, err
	}

	cm.delayedBlockBroadcaster = dbb

	scm := &shardChainMessenger{
		commonMessenger: cm,
	}

	err = dbb.SetBroadcastHandlers(scm.BroadcastMiniBlocks, scm.BroadcastTransactions, scm.BroadcastHeader)
	if err != nil {
		return nil, err
	}

	return scm, nil
}

func checkShardChainNilParameters(
	args ShardChainMessengerArgs,
) error {
	err := checkCommonMessengerNilParameters(args.CommonMessengerArgs)
	if err != nil {
		return err
	}

	return nil
}

// BroadcastBlock will send on in-shard headers topic and on in-shard miniblocks topic the header and block body
func (scm *shardChainMessenger) BroadcastBlock(blockBody data.BodyHandler, header data.HeaderHandler) error {
	broadCastData, err := scm.getBroadCastBlockData(blockBody, header)
	if err != nil {
		return err
	}

	headerIdentifier := scm.shardCoordinator.CommunicationIdentifier(core.MetachainShardId)
	selfIdentifier := scm.shardCoordinator.CommunicationIdentifier(scm.shardCoordinator.SelfId())

	scm.messenger.Broadcast(factory.ShardBlocksTopic+headerIdentifier, broadCastData.marshalledHeader)
	scm.messenger.Broadcast(factory.MiniBlocksTopic+selfIdentifier, broadCastData.marshalledBody)

	return nil
}

func (scm *shardChainMessenger) getBroadCastBlockData(blockBody data.BodyHandler, header data.HeaderHandler) (*dataToBroadcast, error) {
	if check.IfNil(blockBody) {
		return nil, spos.ErrNilBody
	}

	err := blockBody.IntegrityAndValidity()
	if err != nil {
		return nil, err
	}

	if check.IfNil(header) {
		return nil, spos.ErrNilHeader
	}

	msgHeader, err := scm.marshalizer.Marshal(header)
	if err != nil {
		return nil, err
	}

	b := blockBody.(*block.Body)
	msgBlockBody, err := scm.marshalizer.Marshal(b)
	if err != nil {
		return nil, err
	}

	return &dataToBroadcast{
		marshalledHeader: msgHeader,
		marshalledBody:   msgBlockBody,
	}, nil
}

// BroadcastHeader will send on in-shard headers topic the header
func (scm *shardChainMessenger) BroadcastHeader(header data.HeaderHandler, pkBytes []byte) error {
	shardIdentifier := scm.shardCoordinator.CommunicationIdentifier(core.MetachainShardId)
	return scm.broadcastHeader(header, pkBytes, shardIdentifier)
}

func (scm *shardChainMessenger) broadcastHeader(header data.HeaderHandler, pkBytes []byte, shardIdentifier string) error {
	if check.IfNil(header) {
		return spos.ErrNilHeader
	}

	msgHeader, err := scm.marshalizer.Marshal(header)
	if err != nil {
		return err
	}

	scm.broadcast(factory.ShardBlocksTopic+shardIdentifier, msgHeader, pkBytes)
	return nil
}

// BroadcastBlockDataLeader broadcasts the block data as consensus group leader
func (scm *shardChainMessenger) BroadcastBlockDataLeader(
	header data.HeaderHandler,
	miniBlocks map[uint32][]byte,
	transactions map[string][][]byte,
	pkBytes []byte,
) error {
	if check.IfNil(header) {
		return spos.ErrNilHeader
	}
	if len(miniBlocks) == 0 {
		return nil
	}

	headerHash, err := core.CalculateHash(scm.marshalizer, scm.hasher, header)
	if err != nil {
		return err
	}

	metaMiniBlocks, metaTransactions := scm.extractMetaMiniBlocksAndTransactions(miniBlocks, transactions)

	broadcastData := &delayedBroadcastData{
		headerHash:     headerHash,
		miniBlocksData: miniBlocks,
		transactions:   transactions,
		pkBytes:        pkBytes,
	}

	err = scm.delayedBlockBroadcaster.SetLeaderData(broadcastData)
	if err != nil {
		return err
	}

	go scm.BroadcastBlockData(metaMiniBlocks, metaTransactions, pkBytes, common.ExtraDelayForBroadcastBlockInfo)
	return nil
}

// PrepareBroadcastHeaderValidator prepares the validator header broadcast in case leader broadcast fails
func (scm *shardChainMessenger) PrepareBroadcastHeaderValidator(
	header data.HeaderHandler,
	_ map[uint32][]byte,
	_ map[string][][]byte,
	idx int,
	pkBytes []byte,
) {
	if check.IfNil(header) {
		log.Error("shardChainMessenger.PrepareBroadcastHeaderValidator", "error", spos.ErrNilHeader)
		return
	}

	headerHash, err := core.CalculateHash(scm.marshalizer, scm.hasher, header)
	if err != nil {
		log.Error("shardChainMessenger.PrepareBroadcastHeaderValidator", "error", err)
		return
	}

	vData := &validatorHeaderBroadcastData{
		headerHash: headerHash,
		header:     header,
		order:      uint32(idx),
		pkBytes:    pkBytes,
	}

	err = scm.delayedBlockBroadcaster.SetHeaderForValidator(vData)
	if err != nil {
		log.Error("shardChainMessenger.PrepareBroadcastHeaderValidator", "error", err)
		return
	}
}

// PrepareBroadcastBlockDataValidator prepares the validator block data broadcast in case leader broadcast fails
func (scm *shardChainMessenger) PrepareBroadcastBlockDataValidator(
	header data.HeaderHandler,
	miniBlocks map[uint32][]byte,
	transactions map[string][][]byte,
	idx int,
	pkBytes []byte,
) {
	if check.IfNil(header) {
		log.Error("shardChainMessenger.PrepareBroadcastBlockDataValidator", "error", spos.ErrNilHeader)
		return
	}
	if len(miniBlocks) == 0 {
		return
	}

	headerHash, err := core.CalculateHash(scm.marshalizer, scm.hasher, header)
	if err != nil {
		log.Error("shardChainMessenger.PrepareBroadcastBlockDataValidator", "error", err)
		return
	}

	broadcastData := &delayedBroadcastData{
		headerHash:     headerHash,
		header:         header,
		miniBlocksData: miniBlocks,
		transactions:   transactions,
		order:          uint32(idx),
		pkBytes:        pkBytes,
	}

	err = scm.delayedBlockBroadcaster.SetValidatorData(broadcastData)
	if err != nil {
		log.Error("shardChainMessenger.PrepareBroadcastBlockDataValidator", "error", err)
		return
	}
}

// Close closes all the started infinite looping goroutines and subcomponents
func (scm *shardChainMessenger) Close() {
	scm.delayedBlockBroadcaster.Close()
}

// IsInterfaceNil returns true if there is no value under the interface
func (scm *shardChainMessenger) IsInterfaceNil() bool {
	return scm == nil
}
