package bootstrap

import (
	"context"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/epochStart/bootstrap/disabled"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/process/interceptors"
	interceptorsFactory "github.com/multiversx/mx-chain-go/process/interceptors/factory"
	"github.com/multiversx/mx-chain-go/sharding"
)

var _ epochStart.StartOfEpochMetaSyncer = (*epochStartMetaSyncer)(nil)

type epochStartMetaSyncer struct {
	requestHandler                 RequestHandler
	messenger                      Messenger
	marshalizer                    marshal.Marshalizer
	hasher                         hashing.Hasher
	singleDataInterceptor          process.Interceptor
	metaBlockProcessor             EpochStartMetaBlockInterceptorProcessor
	epochStartTopicProviderHandler epochStartTopicProviderHandler
}

// ArgsNewEpochStartMetaSyncer -
type ArgsNewEpochStartMetaSyncer struct {
	CoreComponentsHolder    process.CoreComponentsHolder
	CryptoComponentsHolder  process.CryptoComponentsHolder
	RequestHandler          RequestHandler
	Messenger               Messenger
	ShardCoordinator        sharding.Coordinator
	EconomicsData           process.EconomicsDataHandler
	WhitelistHandler        process.WhiteListHandler
	StartInEpochConfig      config.EpochStartConfig
	ArgsParser              process.ArgumentsParser
	HeaderIntegrityVerifier process.HeaderIntegrityVerifier
	MetaBlockProcessor      EpochStartMetaBlockInterceptorProcessor
}

// NewEpochStartMetaSyncer will return a new instance of epochStartMetaSyncer
func NewEpochStartMetaSyncer(args ArgsNewEpochStartMetaSyncer) (*epochStartMetaSyncer, error) {
	e, err := newEpochStartMetaSyncer(args)
	if err != nil {
		return nil, err
	}

	e.singleDataInterceptor, err = createMetaSingleDataInterceptor(args)
	if err != nil {
		return nil, err
	}

	e.epochStartTopicProviderHandler = e
	return e, nil
}

func newEpochStartMetaSyncer(args ArgsNewEpochStartMetaSyncer) (*epochStartMetaSyncer, error) {
	if check.IfNil(args.CoreComponentsHolder) {
		return nil, epochStart.ErrNilCoreComponentsHolder
	}
	if check.IfNil(args.CryptoComponentsHolder) {
		return nil, epochStart.ErrNilCryptoComponentsHolder
	}
	if check.IfNil(args.CoreComponentsHolder.AddressPubKeyConverter()) {
		return nil, epochStart.ErrNilPubkeyConverter
	}
	if check.IfNil(args.HeaderIntegrityVerifier) {
		return nil, epochStart.ErrNilHeaderIntegrityVerifier
	}
	if check.IfNil(args.MetaBlockProcessor) {
		return nil, epochStart.ErrNilMetablockProcessor
	}

	return &epochStartMetaSyncer{
		requestHandler:     args.RequestHandler,
		messenger:          args.Messenger,
		marshalizer:        args.CoreComponentsHolder.InternalMarshalizer(),
		hasher:             args.CoreComponentsHolder.Hasher(),
		metaBlockProcessor: args.MetaBlockProcessor,
	}, nil
}

func createMetaSingleDataInterceptor(args ArgsNewEpochStartMetaSyncer) (process.Interceptor, error) {
	argsInterceptedDataFactory := createArgsInterceptedDataFactory(args)
	interceptedMetaHdrDataFactory, err := interceptorsFactory.NewInterceptedMetaHeaderDataFactory(&argsInterceptedDataFactory)
	if err != nil {
		return nil, err
	}

	return interceptors.NewSingleDataInterceptor(
		interceptors.ArgSingleDataInterceptor{
			Topic:                factory.MetachainBlocksTopic,
			DataFactory:          interceptedMetaHdrDataFactory,
			Processor:            args.MetaBlockProcessor,
			Throttler:            disabled.NewThrottler(),
			AntifloodHandler:     disabled.NewAntiFloodHandler(),
			WhiteListRequest:     args.WhitelistHandler,
			CurrentPeerId:        args.Messenger.ID(),
			PreferredPeersHolder: disabled.NewPreferredPeersHolder(),
		},
	)
}

func createArgsInterceptedDataFactory(args ArgsNewEpochStartMetaSyncer) interceptorsFactory.ArgInterceptedDataFactory {
	return interceptorsFactory.ArgInterceptedDataFactory{
		CoreComponents:          args.CoreComponentsHolder,
		CryptoComponents:        args.CryptoComponentsHolder,
		ShardCoordinator:        args.ShardCoordinator,
		NodesCoordinator:        disabled.NewNodesCoordinator(),
		FeeHandler:              args.EconomicsData,
		HeaderSigVerifier:       disabled.NewHeaderSigVerifier(),
		HeaderIntegrityVerifier: args.HeaderIntegrityVerifier,
		ValidityAttester:        disabled.NewValidityAttester(),
		EpochStartTrigger:       disabled.NewEpochStartTrigger(),
		ArgsParser:              args.ArgsParser,
	}
}

// SyncEpochStartMeta syncs the latest epoch start metablock
func (e *epochStartMetaSyncer) SyncEpochStartMeta(timeToWait time.Duration) (data.MetaHeaderHandler, error) {
	err := e.initTopicForEpochStartMetaBlockInterceptor()
	if err != nil {
		return nil, err
	}
	defer func() {
		e.resetTopicsAndInterceptors()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), timeToWait)
	mb, errConsensusNotReached := e.metaBlockProcessor.GetEpochStartMetaBlock(ctx)
	cancel()

	if errConsensusNotReached != nil {
		return nil, errConsensusNotReached
	}

	return mb, nil
}

func (e *epochStartMetaSyncer) resetTopicsAndInterceptors() {
	err := e.messenger.UnregisterMessageProcessor(e.epochStartTopicProviderHandler.getTopic(), common.EpochStartInterceptorsIdentifier)
	if err != nil {
		log.Trace("error unregistering message processors", "error", err)
	}
}

func (e *epochStartMetaSyncer) initTopicForEpochStartMetaBlockInterceptor() error {
	err := e.messenger.CreateTopic(e.epochStartTopicProviderHandler.getTopic(), true)
	if err != nil {
		log.Warn("error messenger create topic", "error", err)
		return err
	}

	e.resetTopicsAndInterceptors()
	err = e.messenger.RegisterMessageProcessor(e.epochStartTopicProviderHandler.getTopic(), common.EpochStartInterceptorsIdentifier, e.singleDataInterceptor)
	if err != nil {
		return err
	}

	return nil
}

func (e *epochStartMetaSyncer) getTopic() string {
	return factory.MetachainBlocksTopic
}

// IsInterfaceNil returns true if underlying object is nil
func (e *epochStartMetaSyncer) IsInterfaceNil() bool {
	return e == nil
}
