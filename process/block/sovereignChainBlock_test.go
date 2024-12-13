package block_test

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	atomicCore "github.com/multiversx/mx-chain-core-go/core/atomic"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	sovereignCore "github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/dataRetriever/requestHandlers"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process"
	blproc "github.com/multiversx/mx-chain-go/process/block"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/process/track"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	"github.com/multiversx/mx-chain-go/testscommon/epochNotifier"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/multiversx/mx-chain-go/testscommon/sovereign"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/multiversx/mx-chain-go/testscommon/storage"
)

func createSovereignChainShardTrackerMockArguments() track.ArgShardTracker {
	argsHeaderValidator := blproc.ArgsHeaderValidator{
		Hasher:      &hashingMocks.HasherMock{},
		Marshalizer: &marshallerMock.MarshalizerMock{},
	}
	headerValidator, _ := blproc.NewHeaderValidator(argsHeaderValidator)

	arguments := track.ArgShardTracker{
		ArgBaseTracker: track.ArgBaseTracker{
			Hasher:           &hashingMocks.HasherMock{},
			HeaderValidator:  headerValidator,
			Marshalizer:      &marshallerMock.MarshalizerStub{},
			RequestHandler:   &testscommon.ExtendedShardHeaderRequestHandlerStub{},
			RoundHandler:     &testscommon.RoundHandlerMock{},
			ShardCoordinator: &testscommon.ShardsCoordinatorMock{},
			Store:            &storage.ChainStorerStub{},
			StartHeaders:     createGenesisBlocks(&testscommon.ShardsCoordinatorMock{NoShards: 1}),
			PoolsHolder:      dataRetrieverMock.NewPoolsHolderMock(),
			WhitelistHandler: &testscommon.WhiteListHandlerStub{},
			FeeHandler:       &economicsmocks.EconomicsHandlerStub{},
		},
	}

	return arguments
}

func createSovereignMockArguments(
	coreComp *mock.CoreComponentsMock,
	dataComp *mock.DataComponentsMock,
	bootstrapComp *mock.BootstrapComponentsMock,
	statusComp *mock.StatusComponentsMock,
) blproc.ArgShardProcessor {
	shardArguments := createSovereignChainShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	rrh, _ := requestHandlers.NewResolverRequestHandler(
		&dataRetrieverMock.RequestersFinderStub{},
		&mock.RequestedItemsHandlerStub{},
		&testscommon.WhiteListHandlerStub{},
		1,
		0,
		time.Second,
	)

	arguments := CreateMockArguments(coreComp, dataComp, bootstrapComp, statusComp)
	arguments.BootstrapComponents = &mock.BootstrapComponentsMock{
		Coordinator:          mock.NewOneShardCoordinatorMock(),
		HdrIntegrityVerifier: &mock.HeaderIntegrityVerifierStub{},
		VersionedHdrFactory: &testscommon.VersionedHeaderFactoryStub{
			CreateCalled: func(epoch uint32) data.HeaderHandler {
				return &block.SovereignChainHeader{Header: &block.Header{}}
			},
		},
	}

	arguments.BlockTracker, _ = track.NewSovereignChainShardBlockTrack(sbt)
	arguments.RequestHandler, _ = requestHandlers.NewSovereignResolverRequestHandler(rrh)

	return arguments
}

func createArgsSovereignChainBlockProcessor(baseArgs blproc.ArgShardProcessor) blproc.ArgsSovereignChainBlockProcessor {
	sp, _ := blproc.NewShardProcessor(baseArgs)
	return blproc.ArgsSovereignChainBlockProcessor{
		ShardProcessor:                  sp,
		ValidatorStatisticsProcessor:    &testscommon.ValidatorStatisticsProcessorStub{},
		OutgoingOperationsFormatter:     &sovereign.OutgoingOperationsFormatterMock{},
		OutGoingOperationsPool:          &sovereign.OutGoingOperationsPoolMock{},
		OperationsHasher:                &testscommon.HasherStub{},
		EpochStartDataCreator:           &mock.EpochStartDataCreatorStub{},
		EpochRewardsCreator:             &testscommon.RewardsCreatorStub{},
		ValidatorInfoCreator:            &testscommon.EpochValidatorInfoCreatorStub{},
		EpochSystemSCProcessor:          &testscommon.EpochStartSystemSCStub{},
		EpochEconomics:                  &mock.EpochEconomicsStub{},
		SCToProtocol:                    &mock.SCToProtocolStub{},
		MainChainNotarizationStartRound: 11,
	}
}

func TestSovereignBlockProcessor_NewSovereignChainBlockProcessorShouldWork(t *testing.T) {
	t.Parallel()

	t.Run("should error when shard processor is nil", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := createSovereignMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		sovArgs := createArgsSovereignChainBlockProcessor(arguments)
		sovArgs.ShardProcessor = nil
		scbp, err := blproc.NewSovereignChainBlockProcessor(sovArgs)

		require.Nil(t, scbp)
		require.ErrorIs(t, err, process.ErrNilBlockProcessor)
	})

	t.Run("should error when validator statistics is nil", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := createSovereignMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		sovArgs := createArgsSovereignChainBlockProcessor(arguments)
		sovArgs.ValidatorStatisticsProcessor = nil
		scbp, err := blproc.NewSovereignChainBlockProcessor(sovArgs)

		require.Nil(t, scbp)
		require.ErrorIs(t, err, process.ErrNilValidatorStatistics)
	})

	t.Run("should error when outgoing operations formatter is nil", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := createSovereignMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		sovArgs := createArgsSovereignChainBlockProcessor(arguments)
		sovArgs.OutgoingOperationsFormatter = nil
		scbp, err := blproc.NewSovereignChainBlockProcessor(sovArgs)

		require.Nil(t, scbp)
		require.ErrorIs(t, err, errors.ErrNilOutgoingOperationsFormatter)
	})

	t.Run("should error when outgoing operation pool is nil", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := createSovereignMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		sovArgs := createArgsSovereignChainBlockProcessor(arguments)
		sovArgs.OutGoingOperationsPool = nil
		scbp, err := blproc.NewSovereignChainBlockProcessor(sovArgs)

		require.Nil(t, scbp)
		require.Equal(t, errors.ErrNilOutGoingOperationsPool, err)
	})

	t.Run("should error when operations hasher is nil", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := createSovereignMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		sovArgs := createArgsSovereignChainBlockProcessor(arguments)
		sovArgs.OperationsHasher = nil
		scbp, err := blproc.NewSovereignChainBlockProcessor(sovArgs)

		require.Nil(t, scbp)
		require.Equal(t, errors.ErrNilOperationsHasher, err)
	})

	t.Run("should error when epoch start data creator is nil", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := createSovereignMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		sovArgs := createArgsSovereignChainBlockProcessor(arguments)
		sovArgs.EpochStartDataCreator = nil
		scbp, err := blproc.NewSovereignChainBlockProcessor(sovArgs)

		require.Nil(t, scbp)
		require.Equal(t, process.ErrNilEpochStartDataCreator, err)
	})

	t.Run("should error when epoch rewards creator is nil", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := createSovereignMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		sovArgs := createArgsSovereignChainBlockProcessor(arguments)
		sovArgs.EpochRewardsCreator = nil
		scbp, err := blproc.NewSovereignChainBlockProcessor(sovArgs)

		require.Nil(t, scbp)
		require.Equal(t, process.ErrNilRewardsCreator, err)
	})

	t.Run("should error when epoch validator infor creator is nil", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := createSovereignMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		sovArgs := createArgsSovereignChainBlockProcessor(arguments)
		sovArgs.ValidatorInfoCreator = nil
		scbp, err := blproc.NewSovereignChainBlockProcessor(sovArgs)

		require.Nil(t, scbp)
		require.Equal(t, process.ErrNilEpochStartValidatorInfoCreator, err)
	})

	t.Run("should error when epoch start system sc processor is nil", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := createSovereignMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		sovArgs := createArgsSovereignChainBlockProcessor(arguments)
		sovArgs.EpochSystemSCProcessor = nil
		scbp, err := blproc.NewSovereignChainBlockProcessor(sovArgs)

		require.Nil(t, scbp)
		require.Equal(t, process.ErrNilEpochStartSystemSCProcessor, err)
	})

	t.Run("should error when epoch economics is nil", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := createSovereignMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		sovArgs := createArgsSovereignChainBlockProcessor(arguments)
		sovArgs.EpochEconomics = nil
		scbp, err := blproc.NewSovereignChainBlockProcessor(sovArgs)

		require.Nil(t, scbp)
		require.Equal(t, process.ErrNilEpochEconomics, err)
	})

	t.Run("should error when type assertion to extendedShardHeaderTrackHandler fails", func(t *testing.T) {
		t.Parallel()

		args := CreateMockArguments(createComponentHolderMocks())
		sp, _ := blproc.NewShardProcessor(args)

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := createSovereignMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		sovArgs := createArgsSovereignChainBlockProcessor(arguments)
		sovArgs.ShardProcessor = sp
		scbp, err := blproc.NewSovereignChainBlockProcessor(sovArgs)

		require.Nil(t, scbp)
		require.ErrorIs(t, err, process.ErrWrongTypeAssertion)
	})

	t.Run("should error when type assertion to extendedShardHeaderRequestHandler fails", func(t *testing.T) {
		t.Parallel()

		shardArguments := createSovereignChainShardTrackerMockArguments()
		sbt, _ := track.NewShardBlockTrack(shardArguments)

		args := CreateMockArguments(createComponentHolderMocks())
		args.BlockTracker, _ = track.NewSovereignChainShardBlockTrack(sbt)
		sp, _ := blproc.NewShardProcessor(args)

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := createSovereignMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		sovArgs := createArgsSovereignChainBlockProcessor(arguments)
		sovArgs.ShardProcessor = sp
		scbp, err := blproc.NewSovereignChainBlockProcessor(sovArgs)

		require.Nil(t, scbp)
		require.ErrorIs(t, err, process.ErrWrongTypeAssertion)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := createSovereignMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		sovArgs := createArgsSovereignChainBlockProcessor(arguments)
		scbp, err := blproc.NewSovereignChainBlockProcessor(sovArgs)

		require.NotNil(t, scbp)
		require.Nil(t, err)
	})
}

func TestSovereignChainBlockProcessor_createAndSetOutGoingMiniBlock(t *testing.T) {
	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	arguments := createSovereignMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

	expectedLogs := []*data.LogData{
		{
			TxHash: "txHash1",
		},
	}
	arguments.TxCoordinator = &testscommon.TransactionCoordinatorMock{
		GetAllCurrentLogsCalled: func() []*data.LogData {
			return expectedLogs
		},
	}
	bridgeOp1 := []byte("bridgeOp@123@rcv1@token1@val1")
	bridgeOp2 := []byte("bridgeOp@124@rcv2@token2@val2")

	outgoingOpsHasher := &mock.HasherStub{}
	bridgeOp1Hash := outgoingOpsHasher.Compute(string(bridgeOp1))
	bridgeOp2Hash := outgoingOpsHasher.Compute(string(bridgeOp2))
	bridgeOpsHash := outgoingOpsHasher.Compute(string(append(bridgeOp1Hash, bridgeOp2Hash...)))

	outgoingOperationsFormatter := &sovereign.OutgoingOperationsFormatterMock{
		CreateOutgoingTxDataCalled: func(logs []*data.LogData) ([][]byte, error) {
			require.Equal(t, expectedLogs, logs)
			return [][]byte{bridgeOp1, bridgeOp2}, nil
		},
	}

	poolAddCt := 0
	outGoingOperationsPool := &sovereign.OutGoingOperationsPoolMock{
		AddCalled: func(data *sovereignCore.BridgeOutGoingData) {
			defer func() {
				poolAddCt++
			}()

			switch poolAddCt {
			case 0:
				require.Equal(t, &sovereignCore.BridgeOutGoingData{
					Hash: bridgeOpsHash,
					OutGoingOperations: []*sovereignCore.OutGoingOperation{
						{
							Hash: bridgeOp1Hash,
							Data: bridgeOp1,
						},
						{
							Hash: bridgeOp2Hash,
							Data: bridgeOp2,
						},
					},
				}, data)
			default:
				require.Fail(t, "should not add in pool any other operation")
			}
		},
	}

	sp, _ := blproc.NewShardProcessor(arguments)
	scbp, _ := blproc.NewSovereignChainBlockProcessor(blproc.ArgsSovereignChainBlockProcessor{
		ShardProcessor:               sp,
		ValidatorStatisticsProcessor: &testscommon.ValidatorStatisticsProcessorStub{},
		OutgoingOperationsFormatter:  outgoingOperationsFormatter,
		OutGoingOperationsPool:       outGoingOperationsPool,
		OperationsHasher:             outgoingOpsHasher,
		EpochStartDataCreator:        &mock.EpochStartDataCreatorStub{},
		EpochRewardsCreator:          &testscommon.RewardsCreatorStub{},
		ValidatorInfoCreator:         &testscommon.EpochValidatorInfoCreatorStub{},
		EpochSystemSCProcessor:       &testscommon.EpochStartSystemSCStub{},
		EpochEconomics:               &mock.EpochEconomicsStub{},
		SCToProtocol:                 &mock.SCToProtocolStub{},
	})

	sovChainHdr := &block.SovereignChainHeader{}
	processedMb := &block.MiniBlock{
		ReceiverShardID: core.SovereignChainShardId,
		SenderShardID:   core.MainChainShardId,
	}
	blockBody := &block.Body{
		MiniBlocks: []*block.MiniBlock{processedMb},
	}

	err := scbp.CreateAndSetOutGoingMiniBlock(sovChainHdr, blockBody)
	require.Nil(t, err)
	require.Equal(t, 1, poolAddCt)

	expectedOutGoingMb := &block.MiniBlock{
		TxHashes:        [][]byte{bridgeOp1Hash, bridgeOp2Hash},
		ReceiverShardID: core.MainChainShardId,
		SenderShardID:   arguments.BootstrapComponents.ShardCoordinator().SelfId(),
	}
	expectedBlockBody := &block.Body{
		MiniBlocks: []*block.MiniBlock{processedMb, expectedOutGoingMb},
	}
	require.Equal(t, expectedBlockBody, blockBody)

	expectedOutGoingMbHash, err := core.CalculateHash(arguments.CoreComponents.InternalMarshalizer(), arguments.CoreComponents.Hasher(), expectedOutGoingMb)
	require.Nil(t, err)

	expectedSovChainHeader := &block.SovereignChainHeader{
		OutGoingMiniBlockHeader: &block.OutGoingMiniBlockHeader{
			Hash:                   expectedOutGoingMbHash,
			OutGoingOperationsHash: bridgeOpsHash,
		},
	}
	require.Equal(t, expectedSovChainHeader, sovChainHdr)
}

func TestSovereignShardProcessor_CreateNewBlockExpectCheckRoundCalled(t *testing.T) {
	t.Parallel()

	round := uint64(4)
	checkRoundCt := atomicCore.Counter{}

	roundsNotifier := &epochNotifier.RoundNotifierStub{
		CheckRoundCalled: func(header data.HeaderHandler) {
			checkRoundCt.Increment()
			require.Equal(t, round, header.GetRound())
		},
	}

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	coreComponents.RoundNotifierField = roundsNotifier
	arguments := createSovereignMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	sovArgs := createArgsSovereignChainBlockProcessor(arguments)
	scbp, err := blproc.NewSovereignChainBlockProcessor(sovArgs)
	require.Nil(t, err)

	headerHandler, err := scbp.CreateNewHeader(round, 1)
	require.Nil(t, err)
	require.NotNil(t, headerHandler)
	require.Equal(t, int64(1), checkRoundCt.Get())
}

func TestSovereignShardProcessor_CreateNewHeaderValsOK(t *testing.T) {
	t.Parallel()

	round := uint64(7)
	nonce := uint64(5)

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	arguments := createSovereignMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	sovArgs := createArgsSovereignChainBlockProcessor(arguments)
	scbp, err := blproc.NewSovereignChainBlockProcessor(sovArgs)
	require.Nil(t, err)

	h, err := scbp.CreateNewHeader(round, nonce)
	require.Nil(t, err)
	require.IsType(t, &block.SovereignChainHeader{}, h)
	require.Equal(t, round, h.GetRound())
	require.Equal(t, nonce, h.GetNonce())

	zeroInt := big.NewInt(0)
	require.Nil(t, h.GetValidatorStatsRootHash())
	require.Nil(t, h.GetRootHash())
	require.Equal(t, zeroInt, h.GetDeveloperFees())
	require.Equal(t, zeroInt, h.GetAccumulatedFees())
}

func TestSovereignShardProcessor_CreateBlock(t *testing.T) {
	t.Parallel()

	t.Run("nil block should error", func(t *testing.T) {
		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := createSovereignMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		sovArgs := createArgsSovereignChainBlockProcessor(arguments)
		scbp, err := blproc.NewSovereignChainBlockProcessor(sovArgs)
		require.Nil(t, err)

		doesHaveTime := func() bool {
			return true
		}

		hdr, body, err := scbp.CreateBlock(nil, doesHaveTime)
		require.True(t, check.IfNil(body))
		require.True(t, check.IfNil(hdr))
		require.Equal(t, process.ErrNilBlockHeader, err)
	})
	t.Run("wrong header type should error", func(t *testing.T) {
		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := createSovereignMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		sovArgs := createArgsSovereignChainBlockProcessor(arguments)
		scbp, err := blproc.NewSovereignChainBlockProcessor(sovArgs)
		require.Nil(t, err)

		doesHaveTime := func() bool {
			return true
		}

		meta := &block.MetaBlock{}

		hdr, body, err := scbp.CreateBlock(meta, doesHaveTime)
		require.True(t, check.IfNil(body))
		require.True(t, check.IfNil(hdr))
		require.ErrorContains(t, err, process.ErrWrongTypeAssertion.Error())
	})
	t.Run("account state dirty should error", func(t *testing.T) {
		journalLen := func() int { return 3 }
		revToSnapshot := func(snapshot int) error { return nil }

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := createSovereignMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.AccountsDB[state.UserAccountsState] = &stateMock.AccountsStub{
			JournalLenCalled:       journalLen,
			RevertToSnapshotCalled: revToSnapshot,
		}
		processHandler := arguments.CoreComponents.ProcessStatusHandler()
		mockProcessHandler := processHandler.(*testscommon.ProcessStatusHandlerStub)
		busyIdleCalled := make([]string, 0)
		mockProcessHandler.SetIdleCalled = func() {
			busyIdleCalled = append(busyIdleCalled, idleIdentifier)
		}
		mockProcessHandler.SetBusyCalled = func(reason string) {
			busyIdleCalled = append(busyIdleCalled, busyIdentifier)
		}
		expectedBusyIdleSequencePerCall := []string{busyIdentifier, idleIdentifier}

		sovArgs := createArgsSovereignChainBlockProcessor(arguments)
		scbp, err := blproc.NewSovereignChainBlockProcessor(sovArgs)
		require.Nil(t, err)

		doesHaveTime := func() bool {
			return true
		}
		sovHeader := &block.SovereignChainHeader{
			Header: &block.Header{
				Nonce:         1,
				PubKeysBitmap: []byte("0100101"),
				PrevHash:      []byte(""),
				PrevRandSeed:  []byte("rand seed"),
				Signature:     []byte("signature"),
				RootHash:      []byte("roothash"),
			},
		}

		hdr, body, err := scbp.CreateBlock(sovHeader, doesHaveTime)
		require.True(t, check.IfNil(body))
		require.True(t, check.IfNil(hdr))
		require.Equal(t, process.ErrAccountStateDirty, err)
		require.Equal(t, expectedBusyIdleSequencePerCall, busyIdleCalled)
	})
	t.Run("create block started should error", func(t *testing.T) {
		expectedErr := fmt.Errorf("createBlockStarted error")

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := createSovereignMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.TxCoordinator = &testscommon.TransactionCoordinatorMock{
			AddIntermediateTransactionsCalled: func(_ map[block.Type][]data.TransactionHandler, _ []byte) error {
				return expectedErr
			},
		}
		sovArgs := createArgsSovereignChainBlockProcessor(arguments)
		processHandler := arguments.CoreComponents.ProcessStatusHandler()
		mockProcessHandler := processHandler.(*testscommon.ProcessStatusHandlerStub)
		busyIdleCalled := make([]string, 0)
		mockProcessHandler.SetIdleCalled = func() {
			busyIdleCalled = append(busyIdleCalled, idleIdentifier)
		}
		mockProcessHandler.SetBusyCalled = func(reason string) {
			busyIdleCalled = append(busyIdleCalled, busyIdentifier)
		}
		expectedBusyIdleSequencePerCall := []string{busyIdentifier, idleIdentifier}

		scbp, err := blproc.NewSovereignChainBlockProcessor(sovArgs)
		require.Nil(t, err)

		doesHaveTime := func() bool {
			return true
		}

		expectedSovHeader := &block.SovereignChainHeader{
			Header: &block.Header{
				Nonce: 37,
				Round: 38,
				Epoch: 1,
			},
		}

		hdr, bodyHandler, err := scbp.CreateBlock(expectedSovHeader, doesHaveTime)
		require.True(t, check.IfNil(bodyHandler))
		require.True(t, check.IfNil(hdr))
		require.Equal(t, expectedErr, err)
		require.Equal(t, expectedBusyIdleSequencePerCall, busyIdleCalled)
	})
	t.Run("should work with sovereign header and epoch start rewriting the epoch value", func(t *testing.T) {
		currentEpoch := uint32(1)
		nextEpoch := uint32(2)

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := createSovereignMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.EpochStartTrigger = &testscommon.EpochStartTriggerStub{
			IsEpochStartCalled: func() bool {
				return true
			},
			MetaEpochCalled: func() uint32 {
				return nextEpoch
			},
		}
		sovArgs := createArgsSovereignChainBlockProcessor(arguments)
		processHandler := arguments.CoreComponents.ProcessStatusHandler()
		mockProcessHandler := processHandler.(*testscommon.ProcessStatusHandlerStub)
		busyIdleCalled := make([]string, 0)
		mockProcessHandler.SetIdleCalled = func() {
			busyIdleCalled = append(busyIdleCalled, idleIdentifier)
		}
		mockProcessHandler.SetBusyCalled = func(reason string) {
			busyIdleCalled = append(busyIdleCalled, busyIdentifier)
		}
		expectedBusyIdleSequencePerCall := []string{busyIdentifier, idleIdentifier}

		scbp, err := blproc.NewSovereignChainBlockProcessor(sovArgs)
		require.Nil(t, err)

		sovHeader := &block.SovereignChainHeader{
			Header: &block.Header{
				Nonce: 37,
				Round: 38,
				Epoch: currentEpoch,
			},
		}
		expectedSovHeader := &block.SovereignChainHeader{
			Header: &block.Header{
				Nonce: 37,
				Round: 38,
				Epoch: nextEpoch,
			},
			IsStartOfEpoch: true,
		}
		doesHaveTime := func() bool {
			return true
		}

		hdr, bodyHandler, err := scbp.CreateBlock(sovHeader, doesHaveTime)
		require.False(t, check.IfNil(bodyHandler))
		body, ok := bodyHandler.(*block.Body)
		require.True(t, ok)
		require.Zero(t, len(body.MiniBlocks))
		require.False(t, check.IfNil(hdr))
		require.Equal(t, expectedSovHeader, hdr)
		require.Nil(t, err)
		require.Equal(t, expectedBusyIdleSequencePerCall, busyIdleCalled)
	})
	t.Run("should work with sovereign header", func(t *testing.T) {
		currentEpoch := uint32(1)

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := createSovereignMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.EpochStartTrigger = &testscommon.EpochStartTriggerStub{
			EpochCalled: func() uint32 {
				return currentEpoch
			}}
		sovArgs := createArgsSovereignChainBlockProcessor(arguments)
		processHandler := arguments.CoreComponents.ProcessStatusHandler()
		mockProcessHandler := processHandler.(*testscommon.ProcessStatusHandlerStub)
		busyIdleCalled := make([]string, 0)
		mockProcessHandler.SetIdleCalled = func() {
			busyIdleCalled = append(busyIdleCalled, idleIdentifier)
		}
		mockProcessHandler.SetBusyCalled = func(reason string) {
			busyIdleCalled = append(busyIdleCalled, busyIdentifier)
		}
		expectedBusyIdleSequencePerCall := []string{busyIdentifier, idleIdentifier}

		scbp, err := blproc.NewSovereignChainBlockProcessor(sovArgs)
		require.Nil(t, err)

		expectedExtendedShardHeaderHashes := [][]byte{[]byte("extHeader")}
		expectedSovHeader := &block.SovereignChainHeader{
			Header: &block.Header{
				Nonce: 37,
				Round: 38,
				Epoch: currentEpoch,
			},
			ExtendedShardHeaderHashes: expectedExtendedShardHeaderHashes,
		}
		doesHaveTime := func() bool {
			return true
		}

		hdr, bodyHandler, err := scbp.CreateBlock(expectedSovHeader, doesHaveTime)
		require.False(t, check.IfNil(bodyHandler))
		body, ok := bodyHandler.(*block.Body)
		require.True(t, ok)
		require.Zero(t, len(body.MiniBlocks))
		require.False(t, check.IfNil(hdr))
		require.Equal(t, expectedSovHeader, hdr)
		require.Nil(t, err)
		require.Equal(t, expectedBusyIdleSequencePerCall, busyIdleCalled)
		sovHdr, ok := hdr.(data.SovereignChainHeaderHandler)
		require.Equal(t, expectedExtendedShardHeaderHashes, sovHdr.GetExtendedShardHeaderHashes())
	})
}
