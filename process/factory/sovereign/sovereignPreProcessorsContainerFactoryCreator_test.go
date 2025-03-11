package sovereign

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-sovereign-go/process/factory/shard/data"
	"github.com/multiversx/mx-chain-sovereign-go/process/mock"
	"github.com/multiversx/mx-chain-sovereign-go/testscommon"
	mockCommon "github.com/multiversx/mx-chain-sovereign-go/testscommon/common"
	dataRetrieverMock "github.com/multiversx/mx-chain-sovereign-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-sovereign-go/testscommon/economicsmocks"
	"github.com/multiversx/mx-chain-sovereign-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-sovereign-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-sovereign-go/testscommon/processMocks"
	stateMock "github.com/multiversx/mx-chain-sovereign-go/testscommon/state"
	storageStubs "github.com/multiversx/mx-chain-sovereign-go/testscommon/storage"
)

func createMockPreProcessorsContainerFactoryArguments() data.ArgPreProcessorsContainerFactory {
	return data.ArgPreProcessorsContainerFactory{
		ShardCoordinator:             mock.NewMultiShardsCoordinatorMock(3),
		Store:                        &storageStubs.ChainStorerStub{},
		Marshaller:                   &mock.MarshalizerMock{},
		Hasher:                       &hashingMocks.HasherMock{},
		DataPool:                     dataRetrieverMock.NewPoolsHolderMock(),
		PubkeyConverter:              testscommon.NewPubkeyConverterMock(32),
		Accounts:                     &stateMock.AccountsStub{},
		RequestHandler:               &testscommon.RequestHandlerStub{},
		TxProcessor:                  &testscommon.TxProcessorMock{},
		ScProcessor:                  &testscommon.SCProcessorMock{},
		ScResultProcessor:            &testscommon.SmartContractResultsProcessorMock{},
		RewardsTxProcessor:           &testscommon.RewardTxProcessorMock{},
		EconomicsFee:                 &economicsmocks.EconomicsHandlerStub{},
		GasHandler:                   &testscommon.GasHandlerStub{},
		BlockTracker:                 &mock.BlockTrackerMock{},
		BlockSizeComputation:         &testscommon.BlockSizeComputationStub{},
		BalanceComputation:           &testscommon.BalanceComputationStub{},
		EnableEpochsHandler:          &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		TxTypeHandler:                &testscommon.TxTypeHandlerMock{},
		ScheduledTxsExecutionHandler: &testscommon.ScheduledTxsExecutionStub{},
		ProcessedMiniBlocksTracker:   &testscommon.ProcessedMiniBlocksTrackerStub{},
		TxExecutionOrderHandler:      &mockCommon.TxExecutionOrderHandlerStub{},
		RunTypeComponents:            processMocks.NewRunTypeComponentsStub(),
	}
}

func TestSovereignPreProcessorContainerFactoryCreator_CreatePreProcessorContainerFactory(t *testing.T) {
	t.Parallel()

	f := NewSovereignPreProcessorContainerFactoryCreator()
	require.False(t, f.IsInterfaceNil())

	args := createMockPreProcessorsContainerFactoryArguments()
	containerFactory, err := f.CreatePreProcessorContainerFactory(args)
	require.Nil(t, err)
	require.Equal(t, "*sovereign.sovereignPreProcessorsContainerFactory", fmt.Sprintf("%T", containerFactory))
}
