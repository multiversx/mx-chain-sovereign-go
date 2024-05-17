package vm_test

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory/vm"
	"github.com/multiversx/mx-chain-go/process/factory/metachain"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks"
	"github.com/multiversx/mx-chain-go/testscommon"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/epochNotifier"
	"github.com/multiversx/mx-chain-go/testscommon/factory"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/multiversx/mx-chain-go/testscommon/vmContext"
	"github.com/multiversx/mx-chain-go/vm/systemSmartContracts/defaults"
	vmcommonBuiltInFunctions "github.com/multiversx/mx-chain-vm-common-go/builtInFunctions"
	wasmConfig "github.com/multiversx/mx-chain-vm-go/config"
	"github.com/stretchr/testify/require"
)

func createMockBlockChainHookArgs() hooks.ArgBlockChainHook {
	return hooks.ArgBlockChainHook{
		Accounts:              &stateMock.AccountsStub{},
		PubkeyConv:            &testscommon.PubkeyConverterMock{},
		StorageService:        &storageStubs.ChainStorerStub{},
		BlockChain:            &testscommon.ChainHandlerStub{},
		ShardCoordinator:      &testscommon.ShardsCoordinatorMock{},
		Marshalizer:           &testscommon.ProtoMarshalizerMock{},
		Uint64Converter:       &testscommon.Uint64ByteSliceConverterStub{},
		BuiltInFunctions:      vmcommonBuiltInFunctions.NewBuiltInFunctionContainer(),
		NFTStorageHandler:     &testscommon.SimpleNFTStorageHandlerStub{},
		GlobalSettingsHandler: &testscommon.ESDTGlobalSettingsHandlerStub{},
		DataPool:              &dataRetrieverMock.PoolsHolderMock{},
		CompiledSCPool:        &testscommon.CacherStub{},
		EpochNotifier:         &epochNotifier.EpochNotifierStub{},
		EnableEpochsHandler:   &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		NilCompiledSCStore:    true,
		GasSchedule: &testscommon.GasScheduleNotifierMock{
			LatestGasScheduleCalled: func() map[string]map[string]uint64 {
				return make(map[string]map[string]uint64)
			},
		},
		Counter:                  &testscommon.BlockChainHookCounterStub{},
		MissingTrieNodesNotifier: &testscommon.MissingTrieNodesNotifierStub{},
	}
}

func makeGasSchedule() core.GasScheduleNotifier {
	gasSchedule := wasmConfig.MakeGasMapForTests()
	defaults.FillGasMapInternal(gasSchedule, 1)
	return testscommon.NewGasScheduleNotifierMock(gasSchedule)
}

func createVmContainerMockArgument(gasSchedule core.GasScheduleNotifier) metachain.ArgsNewVMContainerFactory {
	return metachain.ArgsNewVMContainerFactory{
		BlockChainHook:      &testscommon.BlockChainHookStub{},
		PubkeyConv:          testscommon.NewPubkeyConverterMock(32),
		Economics:           &economicsmocks.EconomicsHandlerStub{},
		MessageSignVerifier: &mock.MessageSignVerifierMock{},
		GasSchedule:         gasSchedule,
		NodesConfigProvider: &mock.NodesConfigProviderStub{},
		Hasher:              &hashingMocks.HasherMock{},
		Marshalizer:         &mock.MarshalizerMock{},
		SystemSCConfig: &config.SystemSmartContractsConfig{
			ESDTSystemSCConfig: config.ESDTSystemSCConfig{
				BaseIssuingCost: "100000000",
				OwnerAddress:    "aaaaaa",
			},
			GovernanceSystemSCConfig: config.GovernanceSystemSCConfig{
				V1: config.GovernanceSystemSCConfigV1{
					ProposalCost: "500",
				},
				Active: config.GovernanceSystemSCConfigActive{
					ProposalCost:     "500",
					MinQuorum:        0.5,
					MinPassThreshold: 0.5,
					MinVetoThreshold: 0.5,
					LostProposalFee:  "1",
				},
				OwnerAddress: "3132333435363738393031323334353637383930313233343536373839303234",
			},
			StakingSystemSCConfig: config.StakingSystemSCConfig{
				GenesisNodePrice:                     "1000",
				UnJailValue:                          "100",
				MinStepValue:                         "100",
				MinStakeValue:                        "1",
				UnBondPeriod:                         1,
				NumRoundsWithoutBleed:                1,
				MaximumPercentageToBleed:             1,
				BleedPercentagePerRound:              1,
				MaxNumberOfNodesForStake:             100,
				ActivateBLSPubKeyMessageVerification: false,
				MinUnstakeTokensValue:                "1",
				StakeLimitPercentage:                 100.0,
				NodeLimitPercentage:                  100.0,
			},
			DelegationManagerSystemSCConfig: config.DelegationManagerSystemSCConfig{
				MinCreationDeposit:  "100",
				MinStakeAmount:      "100",
				ConfigChangeAddress: "3132333435363738393031323334353637383930313233343536373839303234",
			},
			DelegationSystemSCConfig: config.DelegationSystemSCConfig{
				MinServiceFee: 0,
				MaxServiceFee: 100,
			},
		},
		ValidatorAccountsDB: &stateMock.AccountsStub{},
		UserAccountsDB:      &stateMock.AccountsStub{},
		ChanceComputer:      &mock.RaterMock{},
		ShardCoordinator:    &mock.ShardCoordinatorStub{},
		EnableEpochsHandler: enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.StakeFlag),
		NodesCoordinator:    &shardingMocks.NodesCoordinatorMock{},
	}
}

func TestNewVmContainerMetaCreatorFactory(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		vmContainerMetaFactory, err := vm.NewVmContainerMetaFactory(&factory.BlockChainHookHandlerFactoryMock{}, &vmContext.VMContextCreatorStub{})
		require.Nil(t, err)
		require.False(t, vmContainerMetaFactory.IsInterfaceNil())
	})

	t.Run("nil blockchain hook creator", func(t *testing.T) {
		vmContainerMetaFactory, err := vm.NewVmContainerMetaFactory(nil, &vmContext.VMContextCreatorStub{})
		require.ErrorIs(t, err, errors.ErrNilBlockChainHookCreator)
		require.True(t, vmContainerMetaFactory.IsInterfaceNil())
	})

	t.Run("nil vm context creator", func(t *testing.T) {
		vmContainerMetaFactory, err := vm.NewVmContainerMetaFactory(&factory.BlockChainHookHandlerFactoryMock{}, nil)
		require.ErrorIs(t, err, errors.ErrNilVMContextCreator)
		require.True(t, vmContainerMetaFactory.IsInterfaceNil())
	})
}

func TestVmContainerMetaFactory_CreateVmContainerFactoryMeta(t *testing.T) {
	t.Parallel()

	vmContainerMetaFactory, err := vm.NewVmContainerMetaFactory(&factory.BlockChainHookHandlerFactoryMock{}, &vmContext.VMContextCreatorStub{})
	require.Nil(t, err)
	require.False(t, vmContainerMetaFactory.IsInterfaceNil())

	argsBlockchain := createMockBlockChainHookArgs()
	gasSchedule := makeGasSchedule()
	argsMeta := createVmContainerMockArgument(gasSchedule)
	args := vm.ArgsVmContainerFactory{
		Economics:           argsMeta.Economics,
		MessageSignVerifier: argsMeta.MessageSignVerifier,
		GasSchedule:         argsMeta.GasSchedule,
		NodesConfigProvider: argsMeta.NodesConfigProvider,
		Hasher:              argsMeta.Hasher,
		Marshalizer:         argsMeta.Marshalizer,
		SystemSCConfig:      argsMeta.SystemSCConfig,
		ValidatorAccountsDB: argsMeta.ValidatorAccountsDB,
		UserAccountsDB:      argsMeta.UserAccountsDB,
		ChanceComputer:      argsMeta.ChanceComputer,
		ShardCoordinator:    argsMeta.ShardCoordinator,
		PubkeyConv:          argsMeta.PubkeyConv,
		EnableEpochsHandler: argsMeta.EnableEpochsHandler,
		NodesCoordinator:    argsMeta.NodesCoordinator,
	}

	vmContainer, vmFactory, err := vmContainerMetaFactory.CreateVmContainerFactory(argsBlockchain, args)
	require.Nil(t, err)
	require.Equal(t, "*containers.virtualMachinesContainer", fmt.Sprintf("%T", vmContainer))
	require.Equal(t, "*metachain.vmContainerFactory", fmt.Sprintf("%T", vmFactory))
}
