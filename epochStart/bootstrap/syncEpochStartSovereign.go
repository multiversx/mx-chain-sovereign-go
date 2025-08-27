package bootstrap

import (
	"github.com/multiversx/mx-chain-core-go/core"

	"github.com/multiversx/mx-chain-go/epochStart/bootstrap/disabled"
	"github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/process/interceptors"
	interceptorsFactory "github.com/multiversx/mx-chain-go/process/interceptors/factory"
)

type epochStartSovereignSyncer struct {
	*epochStartMetaSyncer
}

// internal constructor
func newEpochStartSovereignSyncer(args ArgsNewEpochStartMetaSyncer) (*epochStartSovereignSyncer, error) {
	baseSyncer, err := newEpochStartMetaSyncer(args)
	if err != nil {
		return nil, err
	}

	singleDtaInterceptors, err := createSovereignSingleDataInterceptors(args)
	if err != nil {
		return nil, err
	}

	baseSyncer.singleDataInterceptor = singleDtaInterceptors.singleDataInterceptor
	baseSyncer.proofsInterceptor = singleDtaInterceptors.proofsInterceptor
	baseSyncer.epochStartTopicProviderHandler = newSovereignTopicProvider()

	return &epochStartSovereignSyncer{
		epochStartMetaSyncer: baseSyncer,
	}, nil
}

func createSovereignSingleDataInterceptors(args ArgsNewEpochStartMetaSyncer) (*singleDataInterceptors, error) {
	argsInterceptedDataFactory := createArgsInterceptedDataFactory(args)
	interceptedMetaHdrDataFactory, err := interceptorsFactory.NewInterceptedSovereignShardHeaderDataFactory(&argsInterceptedDataFactory)
	if err != nil {
		return nil, err
	}

	interceptedDataVerifier, err := args.InterceptedDataVerifierFactory.Create(factory.ShardBlocksTopic)
	if err != nil {
		return nil, err
	}

	singleDataInterceptor, err := interceptors.NewSingleDataInterceptor(
		interceptors.ArgSingleDataInterceptor{
			Topic:                   factory.ShardBlocksTopic,
			DataFactory:             interceptedMetaHdrDataFactory,
			Processor:               args.MetaBlockProcessor,
			Throttler:               disabled.NewThrottler(),
			AntifloodHandler:        disabled.NewAntiFloodHandler(),
			WhiteListRequest:        args.WhitelistHandler,
			CurrentPeerId:           args.Messenger.ID(),
			PreferredPeersHolder:    disabled.NewPreferredPeersHolder(),
			InterceptedDataVerifier: interceptedDataVerifier,
		},
	)
	if err != nil {
		return nil, err
	}

	proofInterceptor, err := createProofInterceptor(args, argsInterceptedDataFactory, interceptedDataVerifier, core.SovereignChainShardId, core.SovereignChainShardId)
	if err != nil {
		return nil, err
	}

	return &singleDataInterceptors{
		singleDataInterceptor: singleDataInterceptor,
		proofsInterceptor:     proofInterceptor,
	}, nil
}
