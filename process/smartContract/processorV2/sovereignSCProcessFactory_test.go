package processorV2_test

import (
	"testing"

	"github.com/multiversx/mx-chain-go/process/smartContract/processorV2"
	"github.com/multiversx/mx-chain-go/process/smartContract/scrCommon"
	"github.com/stretchr/testify/require"
)

func TestNewSovereignSCProcessFactory(t *testing.T) {
	t.Parallel()

	fact, err := processorV2.NewSovereignSCProcessFactory(nil)
	require.NotNil(t, err)
	require.Nil(t, fact)

	f, _ := processorV2.NewSCProcessFactory()
	fact, err = processorV2.NewSovereignSCProcessFactory(f)
	require.Nil(t, err)
	require.NotNil(t, fact)
	require.Implements(t, new(scrCommon.SCProcessorCreator), fact)
}

func TestSovereignSCProcessFactory_CreateSCProcessor(t *testing.T) {
	t.Parallel()

	t.Run("Nil EpochNotifier should not fail because it is not used", func(t *testing.T) {
		f, _ := processorV2.NewSCProcessFactory()
		fact, _ := processorV2.NewSovereignSCProcessFactory(f)

		args := processorV2.CreateMockSmartContractProcessorArguments()
		args.EpochNotifier = nil
		scProcessor, err := fact.CreateSCProcessor(args)
		require.Nil(t, err)
		require.NotNil(t, scProcessor)
	})

	t.Run("CreateSCProcessor should work", func(t *testing.T) {
		f, _ := processorV2.NewSCProcessFactory()
		fact, _ := processorV2.NewSovereignSCProcessFactory(f)

		args := processorV2.CreateMockSmartContractProcessorArguments()
		scProcessor, err := fact.CreateSCProcessor(args)
		require.Nil(t, err)
		require.NotNil(t, scProcessor)
		require.Implements(t, new(scrCommon.SCRProcessorHandler), scProcessor)
	})
}

func TestSovereignSCProcessFactory_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	f, _ := processorV2.NewSCProcessFactory()
	fact, _ := processorV2.NewSovereignSCProcessFactory(f)
	require.False(t, fact.IsInterfaceNil())
}
