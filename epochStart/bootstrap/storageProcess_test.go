package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/epochStart/mock"
	errorsMx "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process"
	processMocks "github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
	epochStartMocks "github.com/multiversx/mx-chain-go/testscommon/bootstrapMocks/epochStart"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	"github.com/multiversx/mx-chain-go/testscommon/genesisMocks"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockStorageEpochStartBootstrapArgs(
	coreMock *mock.CoreComponentsMock,
	cryptoMock *mock.CryptoComponentsMock,
) ArgsStorageEpochStartBootstrap {
	args := createMockEpochStartBootstrapArgs(coreMock, cryptoMock)
	esbc := NewEpochStartBootstrapperFactory()
	esb, _ := esbc.CreateEpochStartBootstrapper(args)

	return ArgsStorageEpochStartBootstrap{
		ArgsEpochStartBootstrap:    args,
		ImportDbConfig:             config.ImportDbConfig{},
		ChanGracefullyClose:        make(chan endProcess.ArgEndProcess, 1),
		TimeToWaitForRequestedData: time.Second,
		EpochStartBootStrap:        esb.(*epochStartBootstrap),
	}
}

func TestNewStorageEpochStartBootstrap_InvalidArgumentsShouldErr(t *testing.T) {
	t.Parallel()

	t.Run("nil hash should error", func(t *testing.T) {
		t.Parallel()

		coreComp, cryptoComp := createComponentsForEpochStart()
		coreComp.Hash = nil
		args := createMockStorageEpochStartBootstrapArgs(coreComp, cryptoComp)
		sesb, err := NewStorageEpochStartBootstrap(args)
		assert.True(t, check.IfNil(sesb))
		assert.True(t, errors.Is(err, epochStart.ErrNilHasher))
	})
	t.Run("nil RunTypeComponents should err", func(t *testing.T) {
		t.Parallel()

		coreComp, cryptoComp := createComponentsForEpochStart()
		args := createMockStorageEpochStartBootstrapArgs(coreComp, cryptoComp)
		args.RunTypeComponents = nil
		sesb, err := NewStorageEpochStartBootstrap(args)
		assert.True(t, check.IfNil(sesb))
		require.True(t, errors.Is(err, errorsMx.ErrNilRunTypeComponents))
	})
	t.Run("nil ChanGracefullyClose should err", func(t *testing.T) {
		t.Parallel()

		coreComp, cryptoComp := createComponentsForEpochStart()
		args := createMockStorageEpochStartBootstrapArgs(coreComp, cryptoComp)
		args.ChanGracefullyClose = nil
		sesb, err := NewStorageEpochStartBootstrap(args)
		assert.True(t, check.IfNil(sesb))
		assert.True(t, errors.Is(err, dataRetriever.ErrNilGracefullyCloseChannel))
	})
	t.Run("nil EpochStartBootStrap should err", func(t *testing.T) {
		t.Parallel()

		coreComp, cryptoComp := createComponentsForEpochStart()
		args := createMockStorageEpochStartBootstrapArgs(coreComp, cryptoComp)
		args.EpochStartBootStrap = nil
		sesb, err := NewStorageEpochStartBootstrap(args)
		assert.True(t, check.IfNil(sesb))
		assert.Equal(t, errorsMx.ErrNilEpochStartBootstrapper, err)
	})
}

func TestNewStorageEpochStartBootstrap_ShouldWork(t *testing.T) {
	t.Parallel()

	t.Run("should work for normal", func(t *testing.T) {
		t.Parallel()

		coreComp, cryptoComp := createComponentsForEpochStart()
		args := createMockStorageEpochStartBootstrapArgs(coreComp, cryptoComp)
		esbc := NewEpochStartBootstrapperFactory()
		esb, _ := esbc.CreateEpochStartBootstrapper(args.ArgsEpochStartBootstrap)
		args.EpochStartBootStrap = esb.(*epochStartBootstrap)
		sesb, err := NewStorageEpochStartBootstrap(args)
		assert.False(t, check.IfNil(sesb))
		assert.Nil(t, err)
	})
}

func TestCreateEpochStartBootstrapper_ShouldWork(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createComponentsForEpochStart()

	args := createMockStorageEpochStartBootstrapArgs(coreComp, cryptoComp)
	esbc := NewEpochStartBootstrapperFactory()
	esb, err := esbc.CreateEpochStartBootstrapper(args.ArgsEpochStartBootstrap)

	require.NotNil(t, esb)
	assert.Nil(t, err)
}

func TestStorageEpochStartBootstrap_BootstrapStartInEpochNotEnabled(t *testing.T) {
	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockStorageEpochStartBootstrapArgs(coreComp, cryptoComp)

	err := errors.New("localErr")
	args.LatestStorageDataProvider = &mock.LatestStorageDataProviderStub{
		GetCalled: func() (storage.LatestDataFromStorage, error) {
			return storage.LatestDataFromStorage{}, err
		},
	}

	sesb := initializeStorageEpochStartBootstrap(args)

	params, err := sesb.Bootstrap()
	assert.Nil(t, err)
	assert.Equal(t, uint32(0), params.Epoch)
}

func TestStorageEpochStartBootstrap_BootstrapFromGenesis(t *testing.T) {
	roundsPerEpoch := int64(100)
	roundDuration := uint64(60000)
	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockStorageEpochStartBootstrapArgs(coreComp, cryptoComp)
	args.EconomicsData = &economicsmocks.EconomicsHandlerStub{
		MinGasPriceCalled: func() uint64 {
			return 1
		},
	}
	args.GenesisNodesConfig = &genesisMocks.NodesSetupStub{
		GetRoundDurationCalled: func() uint64 {
			return roundDuration
		},
	}
	args.GeneralConfig = testscommon.GetGeneralConfig()
	args.GeneralConfig.EpochStartConfig.RoundsPerEpoch = roundsPerEpoch

	sesb := initializeStorageEpochStartBootstrap(args)

	params, err := sesb.Bootstrap()
	assert.Nil(t, err)
	assert.Equal(t, uint32(0), params.Epoch)
}

func TestStorageEpochStartBootstrap_BootstrapMetablockNotFound(t *testing.T) {
	roundsPerEpoch := int64(100)
	roundDuration := uint64(6000)
	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockStorageEpochStartBootstrapArgs(coreComp, cryptoComp)
	args.EconomicsData = &economicsmocks.EconomicsHandlerStub{
		MinGasPriceCalled: func() uint64 {
			return 1
		},
	}
	args.GenesisNodesConfig = &genesisMocks.NodesSetupStub{
		GetRoundDurationCalled: func() uint64 {
			return roundDuration
		},
	}
	args.RoundHandler = &mock.RoundHandlerStub{
		RoundIndex: 2*roundsPerEpoch + 1,
	}
	args.GeneralConfig = testscommon.GetGeneralConfig()
	args.GeneralConfig.EpochStartConfig.RoundsPerEpoch = roundsPerEpoch

	sesb := initializeStorageEpochStartBootstrap(args)

	params, err := sesb.Bootstrap()
	assert.Equal(t, process.ErrNilMetaBlockHeader, err)
	assert.Equal(t, uint32(0), params.Epoch)
}

func TestStorageEpochStartBootstrap_requestAndProcessFromStorage(t *testing.T) {
	t.Parallel()

	t.Run("request and process for shard", func(t *testing.T) {
		t.Parallel()

		testRequestAndProcessFromStorageByShardId(t, uint32(0))
	})
	t.Run("request and process for meta", func(t *testing.T) {
		t.Parallel()

		testRequestAndProcessFromStorageByShardId(t, core.MetachainShardId)
	})
}

func testRequestAndProcessFromStorageByShardId(t *testing.T, shardId uint32) {
	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockStorageEpochStartBootstrapArgs(coreComp, cryptoComp)
	args.GenesisNodesConfig = getNodesConfigMock(1)
	args.DestinationShardAsObserver = shardId

	prevPrevEpochStartMetaHeaderHash := []byte("prevPrevEpochStartMetaHeaderHash")
	prevEpochStartMetaHeaderHash := []byte("prevEpochStartMetaHeaderHash")
	notarizedShardHeaderHash := []byte("notarizedShardHeaderHash")
	epochStartMetaBlockHash := []byte("epochStartMetaBlockHash")
	prevNotarizedShardHeaderHash := []byte("prevNotarizedShardHeaderHash")
	notarizedShardHeader := &block.Header{
		PrevHash: prevNotarizedShardHeaderHash,
	}
	prevNotarizedShardHeader := &block.Header{}
	notarizedMetaHeaderHash := []byte("notarizedMetaHeaderHash")
	prevMetaHeaderHash := []byte("prevMetaHeaderHash")
	notarizedMetaHeader := &block.MetaBlock{
		PrevHash: prevMetaHeaderHash,
	}

	epochStartMetaBlock := &block.MetaBlock{
		Epoch: 0,
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{
				{
					HeaderHash:            notarizedShardHeaderHash,
					ShardID:               shardId,
					FirstPendingMetaBlock: notarizedMetaHeaderHash,
				},
			},
			Economics: block.Economics{
				PrevEpochStartHash: prevEpochStartMetaHeaderHash,
			},
		},
	}
	prevEpochStartMetaBlock := &block.MetaBlock{
		Epoch: 0,
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{
				{
					HeaderHash: notarizedShardHeaderHash,
					ShardID:    shardId,
				},
			},
			Economics: block.Economics{
				PrevEpochStartHash: prevPrevEpochStartMetaHeaderHash,
			},
		},
	}

	sesb := initializeStorageEpochStartBootstrap(args)
	sesb.epochStartMeta = epochStartMetaBlock
	sesb.requestHandler = &testscommon.RequestHandlerStub{}
	sesb.dataPool = dataRetrieverMock.NewPoolsHolderMock()

	sesb.headersSyncer = &epochStartMocks.HeadersByHashSyncerStub{
		GetHeadersCalled: func() (m map[string]data.HeaderHandler, err error) {
			return map[string]data.HeaderHandler{
				string(notarizedShardHeaderHash):     notarizedShardHeader,
				string(prevEpochStartMetaHeaderHash): prevEpochStartMetaBlock,
				string(epochStartMetaBlockHash):      epochStartMetaBlock,
				string(prevNotarizedShardHeaderHash): prevNotarizedShardHeader,
				string(notarizedMetaHeaderHash):      notarizedMetaHeader,
			}, nil
		},
	}
	sesb.miniBlocksSyncer = &epochStartMocks.PendingMiniBlockSyncHandlerStub{}

	params, err := sesb.requestAndProcessFromStorage()

	pksBytes := createPkBytes(args.GenesisNodesConfig.NumberOfShards())

	requiredParameters := Parameters{
		Epoch:       0,
		SelfShardId: shardId,
		NumOfShards: 1,
		NodesConfig: &nodesCoordinator.NodesCoordinatorRegistry{
			EpochsConfig: map[string]*nodesCoordinator.EpochValidators{
				"0": {
					EligibleValidators: map[string][]*nodesCoordinator.SerializableValidator{
						"0": {
							&nodesCoordinator.SerializableValidator{
								PubKey:  pksBytes[0],
								Chances: 1,
								Index:   0,
							},
						},
						"4294967295": {
							&nodesCoordinator.SerializableValidator{
								PubKey:  pksBytes[core.MetachainShardId],
								Chances: 1,
								Index:   0,
							},
						},
					},
					WaitingValidators: map[string][]*nodesCoordinator.SerializableValidator{},
					LeavingValidators: map[string][]*nodesCoordinator.SerializableValidator{},
				},
			},
			CurrentEpoch: 0,
		},
	}

	assert.Nil(t, err)
	assert.Equal(t, requiredParameters, params)
}

func TestStorageEpochStartBootstrap_syncHeadersFromStorage(t *testing.T) {
	t.Parallel()

	t.Run("fail to sync missing headers", func(t *testing.T) {
		t.Parallel()

		coreComp, cryptoComp := createComponentsForEpochStart()
		args := createMockStorageEpochStartBootstrapArgs(coreComp, cryptoComp)

		hdrHash1 := []byte("hdrHash1")
		hdrHash2 := []byte("hdrHash2")

		metaBlock := &block.MetaBlock{
			Epoch: 2,
			EpochStart: block.EpochStart{
				LastFinalizedHeaders: []block.EpochStartShardData{
					{HeaderHash: hdrHash1, ShardID: 0},
				},
				Economics: block.Economics{
					PrevEpochStartHash: hdrHash2,
				},
			},
		}

		sesb, _ := NewStorageEpochStartBootstrap(args)
		expectedErr := errors.New("expected error")
		sesb.headersSyncer = &epochStartMocks.HeadersByHashSyncerStub{
			SyncMissingHeadersByHashCalled: func(shardIDs []uint32, headersHashes [][]byte, ctx context.Context) error {
				return expectedErr
			},
		}

		syncedHeaders, err := sesb.bootStrapShardProcessor.syncHeadersFromStorage(
			metaBlock,
			0,
			sesb.importDbConfig.ImportDBTargetShardID,
			sesb.timeToWaitForRequestedData,
		)
		assert.Nil(t, syncedHeaders)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("fail to get synced headers", func(t *testing.T) {
		t.Parallel()

		coreComp, cryptoComp := createComponentsForEpochStart()
		args := createMockStorageEpochStartBootstrapArgs(coreComp, cryptoComp)

		hdrHash1 := []byte("hdrHash1")
		hdrHash2 := []byte("hdrHash2")

		metaBlock := &block.MetaBlock{
			Epoch: 2,
			EpochStart: block.EpochStart{
				LastFinalizedHeaders: []block.EpochStartShardData{
					{HeaderHash: hdrHash1, ShardID: 0},
				},
				Economics: block.Economics{
					PrevEpochStartHash: hdrHash2,
				},
			},
		}

		sesb, _ := NewStorageEpochStartBootstrap(args)
		expectedErr := errors.New("expected error")
		sesb.headersSyncer = &epochStartMocks.HeadersByHashSyncerStub{
			GetHeadersCalled: func() (m map[string]data.HeaderHandler, err error) {
				return nil, expectedErr
			},
		}

		syncedHeaders, err := sesb.bootStrapShardProcessor.syncHeadersFromStorage(
			metaBlock,
			0,
			sesb.importDbConfig.ImportDBTargetShardID,
			sesb.timeToWaitForRequestedData,
		)
		assert.Nil(t, syncedHeaders)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("empty prev meta block when first epoch", func(t *testing.T) {
		t.Parallel()

		coreComp, cryptoComp := createComponentsForEpochStart()
		args := createMockStorageEpochStartBootstrapArgs(coreComp, cryptoComp)

		hdrHash1 := []byte("hdrHash1")
		hdrHash2 := []byte("hdrHash2")

		metaBlock := &block.MetaBlock{
			Epoch: 1,
			EpochStart: block.EpochStart{
				LastFinalizedHeaders: []block.EpochStartShardData{
					{HeaderHash: hdrHash1, ShardID: 0},
				},
				Economics: block.Economics{
					PrevEpochStartHash: hdrHash2,
				},
			},
		}
		prevMetaBlock := &block.MetaBlock{
			Epoch: 0,
			Nonce: 10,
			EpochStart: block.EpochStart{
				LastFinalizedHeaders: []block.EpochStartShardData{
					{HeaderHash: hdrHash1, ShardID: 0},
				},
			},
		}

		sesb := initializeStorageEpochStartBootstrap(args)
		expectedHeaders := map[string]data.HeaderHandler{
			string(hdrHash1): metaBlock,
			string(hdrHash2): prevMetaBlock,
		}
		sesb.headersSyncer = &epochStartMocks.HeadersByHashSyncerStub{
			GetHeadersCalled: func() (m map[string]data.HeaderHandler, err error) {
				return expectedHeaders, nil
			},
		}

		expectedSyncedHeader := map[string]data.HeaderHandler{
			string(hdrHash1): metaBlock,
			string(hdrHash2): &block.MetaBlock{},
		}

		syncedHeaders, err := sesb.bootStrapShardProcessor.syncHeadersFromStorage(
			metaBlock,
			0,
			sesb.importDbConfig.ImportDBTargetShardID,
			sesb.timeToWaitForRequestedData,
		)
		assert.Nil(t, err)
		assert.Equal(t, expectedSyncedHeader, syncedHeaders)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		coreComp, cryptoComp := createComponentsForEpochStart()
		args := createMockStorageEpochStartBootstrapArgs(coreComp, cryptoComp)

		hdrHash1 := []byte("hdrHash1")
		hdrHash2 := []byte("hdrHash2")

		metaBlock := &block.MetaBlock{
			Epoch: 2,
			EpochStart: block.EpochStart{
				LastFinalizedHeaders: []block.EpochStartShardData{
					{HeaderHash: hdrHash1, ShardID: 0},
				},
				Economics: block.Economics{
					PrevEpochStartHash: hdrHash2,
				},
			},
		}

		sesb, _ := NewStorageEpochStartBootstrap(args)
		expectedHeaders := map[string]data.HeaderHandler{
			string(hdrHash1): metaBlock,
		}
		sesb.headersSyncer = &epochStartMocks.HeadersByHashSyncerStub{
			GetHeadersCalled: func() (m map[string]data.HeaderHandler, err error) {
				return expectedHeaders, nil
			},
		}

		syncedHeaders, err := sesb.bootStrapShardProcessor.syncHeadersFromStorage(
			metaBlock,
			0,
			sesb.importDbConfig.ImportDBTargetShardID,
			sesb.timeToWaitForRequestedData,
		)
		assert.Nil(t, err)
		assert.Equal(t, expectedHeaders, syncedHeaders)
	})
}

func TestStorageEpochStartBootstrap_processNodesConfig(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createComponentsForEpochStart()
	hdrHash1 := []byte("hdrHash1")
	hdrHash2 := []byte("hdrHash2")
	metaBlock := &block.MetaBlock{
		Epoch: 0,
		MiniBlockHeaders: []block.MiniBlockHeader{
			{
				Hash:          hdrHash1,
				SenderShardID: 1,
			},
			{
				Hash:          hdrHash2,
				SenderShardID: 2,
			},
		},
	}

	pksBytes := createPkBytes(1)
	expectedNodesConfig := &nodesCoordinator.NodesCoordinatorRegistry{
		EpochsConfig: map[string]*nodesCoordinator.EpochValidators{
			"0": {
				EligibleValidators: map[string][]*nodesCoordinator.SerializableValidator{
					"0": {
						&nodesCoordinator.SerializableValidator{
							PubKey:  pksBytes[0],
							Chances: 1,
							Index:   0,
						},
					},
					"4294967295": {
						&nodesCoordinator.SerializableValidator{
							PubKey:  pksBytes[core.MetachainShardId],
							Chances: 1,
							Index:   0,
						},
					},
				},
				WaitingValidators: map[string][]*nodesCoordinator.SerializableValidator{},
				LeavingValidators: map[string][]*nodesCoordinator.SerializableValidator{},
			},
		},
		CurrentEpoch: 0,
	}

	args := createMockStorageEpochStartBootstrapArgs(coreComp, cryptoComp)
	args.GeneralConfig = testscommon.GetGeneralConfig()
	args.GenesisNodesConfig = getNodesConfigMock(1)

	sesb := initializeStorageEpochStartBootstrap(args)
	sesb.dataPool = dataRetrieverMock.NewPoolsHolderMock()
	sesb.requestHandler = &testscommon.RequestHandlerStub{}
	sesb.epochStartMeta = metaBlock
	sesb.prevEpochStartMeta = metaBlock

	var err error
	sesb.nodesConfig, sesb.baseData.shardId, err = sesb.bootStrapShardProcessor.processNodesConfigFromStorage([]byte("pubkey"), sesb.importDbConfig.ImportDBTargetShardID)

	assert.Nil(t, err)
	assert.Equal(t, expectedNodesConfig, sesb.nodesConfig)
	assert.Equal(t, sesb.baseData.shardId, args.DestinationShardAsObserver)
}

func TestCreateStorageRequestHandler_ShouldWork(t *testing.T) {
	t.Parallel()

	t.Run("should create a main chain instance", func(t *testing.T) {
		t.Parallel()
		coreComp, cryptoComp := createComponentsForEpochStart()

		args := createMockStorageEpochStartBootstrapArgs(coreComp, cryptoComp)
		sesb, _ := NewStorageEpochStartBootstrap(args)

		requestHandler, err := sesb.createStorageRequestHandler()

		require.NotNil(t, requestHandler)
		assert.Equal(t, "*requestHandlers.resolverRequestHandler", fmt.Sprintf("%T", requestHandler))
		assert.Nil(t, err)
	})

	t.Run("should create a sovereign chain instance", func(t *testing.T) {
		t.Parallel()
		coreComp, cryptoComp := createComponentsForEpochStart()

		args := createMockStorageEpochStartBootstrapArgs(coreComp, cryptoComp)
		args.RunTypeComponents = processMocks.NewSovereignRunTypeComponentsStub()
		sesb := initializeStorageEpochStartBootstrap(args)

		requestHandler, err := sesb.createStorageRequestHandler()

		require.NotNil(t, requestHandler)
		assert.Equal(t, "*requestHandlers.sovereignResolverRequestHandler", fmt.Sprintf("%T", requestHandler))
		assert.Nil(t, err)
	})

}

func initializeStorageEpochStartBootstrap(args ArgsStorageEpochStartBootstrap) *storageEpochStartBootstrap {
	esbc := NewEpochStartBootstrapperFactory()
	esb, _ := esbc.CreateEpochStartBootstrapper(args.ArgsEpochStartBootstrap)
	args.EpochStartBootStrap = esb.(*epochStartBootstrap)
	sesb, _ := NewStorageEpochStartBootstrap(args)
	return sesb
}
