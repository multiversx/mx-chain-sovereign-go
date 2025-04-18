package bootstrap

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	dataCore "github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/process"
)

var _ process.InterceptorProcessor = (*epochStartMetaBlockProcessor)(nil)

type storageEpochStartMetaBlockProcessor struct {
	messenger      Messenger
	requestHandler RequestHandler
	marshalizer    marshal.Marshalizer
	hasher         hashing.Hasher
	chanReceived   chan struct{}
	mutMetablock   sync.Mutex
	metaBlock      dataCore.MetaHeaderHandler
}

// NewStorageEpochStartMetaBlockProcessor will return an interceptor processor for epoch start meta block when importing
// data from storage
func NewStorageEpochStartMetaBlockProcessor(
	messenger Messenger,
	handler RequestHandler,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
) (*storageEpochStartMetaBlockProcessor, error) {
	if check.IfNil(messenger) {
		return nil, epochStart.ErrNilMessenger
	}
	if check.IfNil(handler) {
		return nil, epochStart.ErrNilRequestHandler
	}
	if check.IfNil(marshalizer) {
		return nil, epochStart.ErrNilMarshalizer
	}
	if check.IfNil(hasher) {
		return nil, epochStart.ErrNilHasher
	}

	processor := &storageEpochStartMetaBlockProcessor{
		messenger:      messenger,
		requestHandler: handler,
		marshalizer:    marshalizer,
		hasher:         hasher,
		chanReceived:   make(chan struct{}, 1),
	}

	return processor, nil
}

// Validate will return nil as there is no need for validation
func (ses *storageEpochStartMetaBlockProcessor) Validate(_ process.InterceptedData, _ core.PeerID) error {
	return nil
}

// Save will handle the consensus mechanism for the fetched metablocks
// All errors are just logged because if this function returns an error, the processing is finished. This way, we ignore
// wrong received data and wait for relevant intercepted data
func (ses *storageEpochStartMetaBlockProcessor) Save(data process.InterceptedData, _ core.PeerID, _ string) error {
	if check.IfNil(data) {
		log.Debug("epoch bootstrapper: nil intercepted data")
		return nil
	}

	log.Debug("received header", "type", data.Type(), "hash", data.Hash())
	interceptedHdr, ok := data.(process.HdrValidatorHandler)
	if !ok {
		log.Warn("saving epoch start meta block error", "error", epochStart.ErrWrongTypeAssertion)
		return nil
	}

	metaBlock, ok := interceptedHdr.HeaderHandler().(dataCore.MetaHeaderHandler)
	if !ok {
		log.Warn("saving epoch start meta block error", "error", epochStart.ErrWrongTypeAssertion,
			"header", interceptedHdr.HeaderHandler())
		return nil
	}

	if !metaBlock.IsStartOfEpochBlock() {
		log.Warn("received metablock is not of type epoch start", "error", epochStart.ErrNotEpochStartBlock)
		return nil
	}

	log.Debug("received epoch start meta", "epoch", metaBlock.GetEpoch(), "from peer", "self")
	ses.mutMetablock.Lock()
	ses.metaBlock = metaBlock
	ses.mutMetablock.Unlock()

	select {
	case ses.chanReceived <- struct{}{}:
	default:
	}

	return nil
}

// GetEpochStartMetaBlock will return the metablock after it is confirmed or an error if the number of tries was exceeded
// This is a blocking method which will end after the consensus for the meta block is obtained or the context is done
func (ses *storageEpochStartMetaBlockProcessor) GetEpochStartMetaBlock(ctx context.Context) (dataCore.MetaHeaderHandler, error) {
	ses.requestMetaBlock()

	chanRequests := time.After(durationBetweenReRequests)
	for {
		select {
		case <-ses.chanReceived:
			return ses.getMetablock()
		case <-ctx.Done():
			return ses.getMetablock()
		case <-chanRequests:
			ses.requestMetaBlock()
			chanRequests = time.After(durationBetweenReRequests)
		}
	}
}

func (ses *storageEpochStartMetaBlockProcessor) getMetablock() (dataCore.MetaHeaderHandler, error) {
	ses.mutMetablock.Lock()
	defer ses.mutMetablock.Unlock()

	if check.IfNil(ses.metaBlock) {
		return nil, process.ErrNilMetaBlockHeader
	}

	return ses.metaBlock, nil
}

func (ses *storageEpochStartMetaBlockProcessor) requestMetaBlock() {
	unknownEpoch := uint32(math.MaxUint32)
	ses.requestHandler.RequestStartOfEpochMetaBlock(unknownEpoch)
}

// RegisterHandler registers a callback function to be notified of incoming epoch start metablocks
func (ses *storageEpochStartMetaBlockProcessor) RegisterHandler(_ func(topic string, hash []byte, data interface{})) {
	log.Error("storageEpochStartMetaBlockProcessor.RegisterHandler not implemented")
}

// IsInterfaceNil returns true if there is no value under the interface
func (ses *storageEpochStartMetaBlockProcessor) IsInterfaceNil() bool {
	return ses == nil
}
