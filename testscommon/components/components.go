package components

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	logger "github.com/multiversx/mx-chain-logger-go"
	wasmConfig "github.com/multiversx/mx-chain-vm-go/config"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	commonFactory "github.com/multiversx/mx-chain-go/common/factory"
	"github.com/multiversx/mx-chain-go/config"
	mockConsensus "github.com/multiversx/mx-chain-go/consensus/mock"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/epochStart/bootstrap/disabled"
	"github.com/multiversx/mx-chain-go/factory"
	bootstrapComp "github.com/multiversx/mx-chain-go/factory/bootstrap"
	consensusComp "github.com/multiversx/mx-chain-go/factory/consensus"
	coreComp "github.com/multiversx/mx-chain-go/factory/core"
	cryptoComp "github.com/multiversx/mx-chain-go/factory/crypto"
	dataComp "github.com/multiversx/mx-chain-go/factory/data"
	"github.com/multiversx/mx-chain-go/factory/mock"
	networkComp "github.com/multiversx/mx-chain-go/factory/network"
	processComp "github.com/multiversx/mx-chain-go/factory/processing"
	"github.com/multiversx/mx-chain-go/factory/runType"
	stateComp "github.com/multiversx/mx-chain-go/factory/state"
	statusComp "github.com/multiversx/mx-chain-go/factory/status"
	"github.com/multiversx/mx-chain-go/factory/statusCore"
	"github.com/multiversx/mx-chain-go/genesis"
	"github.com/multiversx/mx-chain-go/genesis/data"
	mockCoreComp "github.com/multiversx/mx-chain-go/integrationTests/mock"
	"github.com/multiversx/mx-chain-go/p2p"
	p2pConfig "github.com/multiversx/mx-chain-go/p2p/config"
	p2pFactory "github.com/multiversx/mx-chain-go/p2p/factory"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon"
	commonMocks "github.com/multiversx/mx-chain-go/testscommon/common"
	"github.com/multiversx/mx-chain-go/testscommon/dblookupext"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/mainFactoryMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/sovereign"
	statusHandlerMock "github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	"github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/multiversx/mx-chain-go/testscommon/subRoundsHolder"
	"github.com/multiversx/mx-chain-go/trie"
)

var log = logger.GetOrCreate("componentsMock")

// TestHasher -
const TestHasher = "blake2b"

// TestMarshalizer -
const TestMarshalizer = "json"

// SignedBlocksThreshold -
const SignedBlocksThreshold = 0.025

// ConsecutiveMissedBlocksPenalty -
const ConsecutiveMissedBlocksPenalty = 1.1

// DummyPk -
const DummyPk = "629e1245577afb7717ccb46b6ff3649bdd6a1311514ad4a7695da13f801cc277ee24e730a7fa8aa6c612159b4328db17" +
	"35692d0bded3a2264ba621d6bda47a981d60e17dd306d608e0875a0ba19639fb0844661f519472a175ca9ed2f33fbe16"

// DummySk -
const DummySk = "cea01c0bf060187d90394802ff223078e47527dc8aa33a922744fb1d06029c4b"

// LoadKeysFunc -
type LoadKeysFunc func(string, int) ([]byte, string, error)

// GetRunTypeCoreComponents -
func GetRunTypeCoreComponents() factory.RunTypeCoreComponentsHolder {
	runTypeCoreComponentsFactory := runType.NewRunTypeCoreComponentsFactory()
	managedRunTypeCoreComponents, err := runType.NewManagedRunTypeCoreComponents(runTypeCoreComponentsFactory)
	if err != nil {
		log.Error("getRunTypeCoreComponents NewManagedRunTypeCoreComponents", "error", err.Error())
		return nil
	}
	err = managedRunTypeCoreComponents.Create()
	if err != nil {
		log.Error("getRunTypeCoreComponents Create", "error", err.Error())
		return nil
	}
	return managedRunTypeCoreComponents
}

// GetSovereignRunTypeCoreComponents -
func GetSovereignRunTypeCoreComponents() factory.RunTypeCoreComponentsHolder {
	sovRunTypeCoreComponentsFactory := runType.NewSovereignRunTypeCoreComponentsFactory(config.SovereignEpochConfig{})
	managedRunTypeCoreComponents, err := runType.NewManagedRunTypeCoreComponents(sovRunTypeCoreComponentsFactory)
	if err != nil {
		log.Error("GetSovereignRunTypeCoreComponents.NewManagedRunTypeCoreComponents", "error", err.Error())
		return nil
	}
	err = managedRunTypeCoreComponents.Create()
	if err != nil {
		log.Error("GetSovereignRunTypeCoreComponents.Create", "error", err.Error())
		return nil
	}
	return managedRunTypeCoreComponents
}

// GetCoreArgs -
func GetCoreArgs() coreComp.CoreComponentsFactoryArgs {
	runTypeCoreComponents := GetRunTypeCoreComponents()
	return coreComp.CoreComponentsFactoryArgs{
		Config: GetGeneralConfig(),
		ConfigPathsHolder: config.ConfigurationPathsHolder{
			GasScheduleDirectoryName: "../../cmd/node/config/gasSchedules",
		},
		RatingsConfig:       CreateDummyRatingsConfig(),
		EconomicsConfig:     CreateDummyEconomicsConfig(),
		NodesFilename:       "../mock/testdata/nodesSetupMock.json",
		WorkingDirectory:    "home",
		ChanStopNodeProcess: make(chan endProcess.ArgEndProcess),
		EpochConfig: config.EpochConfig{
			GasSchedule: config.GasScheduleConfig{
				GasScheduleByEpochs: []config.GasScheduleByEpochs{
					{
						StartEpoch: 0,
						FileName:   "gasScheduleV1.toml",
					},
				},
			},
		},
		RoundConfig: config.RoundConfig{
			RoundActivations: map[string]config.ActivationRoundByName{
				"DisableAsyncCallV1": {
					Round: "18446744073709551615",
				},
			},
		},
		RunTypeCoreComponents: runTypeCoreComponents,
	}
}

// GetStatusCoreArgs -
func GetStatusCoreArgs(coreComponents factory.CoreComponentsHolder) statusCore.StatusCoreComponentsFactoryArgs {
	return statusCore.StatusCoreComponentsFactoryArgs{
		Config: GetGeneralConfig(),
		EpochConfig: config.EpochConfig{
			GasSchedule: config.GasScheduleConfig{
				GasScheduleByEpochs: []config.GasScheduleByEpochs{
					{
						StartEpoch: 0,
						FileName:   "gasScheduleV1.toml",
					},
				},
			},
		},
		RoundConfig: config.RoundConfig{
			RoundActivations: map[string]config.ActivationRoundByName{
				"Example": {
					Round: "18446744073709551615",
				},
			},
		},
		RatingsConfig:   CreateDummyRatingsConfig(),
		EconomicsConfig: CreateDummyEconomicsConfig(),
		CoreComp:        coreComponents,
	}
}

// GetConsensusArgs -
func GetConsensusArgs(shardCoordinator sharding.Coordinator) consensusComp.ConsensusComponentsFactoryArgs {
	coreComponents := GetCoreComponents()
	cryptoComponents := GetCryptoComponents(coreComponents)
	networkComponents := GetNetworkComponents(cryptoComponents)
	stateComponents := GetStateComponents(coreComponents, GetStatusCoreComponents())
	dataComponents := GetDataComponents(coreComponents, shardCoordinator)
	processComponents := GetProcessComponents(
		shardCoordinator,
		coreComponents,
		networkComponents,
		dataComponents,
		cryptoComponents,
		stateComponents,
	)
	statusComponents := GetStatusComponents(
		coreComponents,
		networkComponents,
		stateComponents,
		shardCoordinator,
		processComponents.NodesCoordinator(),
	)

	args := spos.ScheduledProcessorWrapperArgs{
		SyncTimer:                coreComponents.SyncTimer(),
		Processor:                processComponents.BlockProcessor(),
		RoundTimeDurationHandler: coreComponents.RoundHandler(),
	}
	scheduledProcessor, _ := spos.NewScheduledProcessorWrapper(args)

	return consensusComp.ConsensusComponentsFactoryArgs{
		Config:               testscommon.GetGeneralConfig(),
		FlagsConfig:          config.ContextFlagsConfig{},
		BootstrapRoundIndex:  0,
		CoreComponents:       coreComponents,
		NetworkComponents:    networkComponents,
		CryptoComponents:     cryptoComponents,
		DataComponents:       dataComponents,
		ProcessComponents:    processComponents,
		StateComponents:      stateComponents,
		StatusComponents:     statusComponents,
		StatusCoreComponents: GetStatusCoreComponents(),
		ScheduledProcessor:   scheduledProcessor,
		RunTypeComponents:    GetRunTypeComponents(),
		ExtraSignersHolder:   &subRoundsHolder.ExtraSignersHolderMock{},
		SubRoundEndV2Creator: bls.NewSubRoundEndV2Creator(),
	}
}

// GetSovereignConsensusArgs -
func GetSovereignConsensusArgs(shardCoordinator sharding.Coordinator) consensusComp.ConsensusComponentsFactoryArgs {
	coreComponents := GetSovereignCoreComponents()
	cryptoComponents := GetCryptoComponents(coreComponents)
	networkComponents := GetNetworkComponents(cryptoComponents)
	stateComponents := GetSovereignStateComponents(coreComponents, GetStatusCoreComponents())
	dataComponents := GetSovereignDataComponents(coreComponents, shardCoordinator)
	processComponents := GetSovereignProcessComponents(
		shardCoordinator,
		coreComponents,
		networkComponents,
		dataComponents,
		cryptoComponents,
		stateComponents,
	)
	statusComponents := GetStatusComponents(
		coreComponents,
		networkComponents,
		stateComponents,
		shardCoordinator,
		processComponents.NodesCoordinator(),
	)

	args := spos.ScheduledProcessorWrapperArgs{
		SyncTimer:                coreComponents.SyncTimer(),
		Processor:                processComponents.BlockProcessor(),
		RoundTimeDurationHandler: coreComponents.RoundHandler(),
	}
	scheduledProcessor, _ := spos.NewScheduledProcessorWrapper(args)

	return consensusComp.ConsensusComponentsFactoryArgs{
		Config:               testscommon.GetGeneralConfig(),
		FlagsConfig:          config.ContextFlagsConfig{},
		BootstrapRoundIndex:  0,
		CoreComponents:       coreComponents,
		NetworkComponents:    networkComponents,
		CryptoComponents:     cryptoComponents,
		DataComponents:       dataComponents,
		ProcessComponents:    processComponents,
		StateComponents:      stateComponents,
		StatusComponents:     statusComponents,
		StatusCoreComponents: GetStatusCoreComponents(),
		ScheduledProcessor:   scheduledProcessor,
		RunTypeComponents:    GetSovereignRunTypeComponents(),
		ExtraSignersHolder:   &subRoundsHolder.ExtraSignersHolderMock{},
		SubRoundEndV2Creator: bls.NewSubRoundEndV2Creator(),
	}
}

// GetCryptoArgs -
func GetCryptoArgs(coreComponents factory.CoreComponentsHolder) cryptoComp.CryptoComponentsFactoryArgs {
	args := cryptoComp.CryptoComponentsFactoryArgs{
		Config: config.Config{
			GeneralSettings: config.GeneralSettingsConfig{ChainID: "undefined"},
			Consensus: config.ConsensusConfig{
				Type: "bls",
			},
			MultisigHasher: config.TypeConfig{Type: "blake2b"},
			PublicKeyPIDSignature: config.CacheConfig{
				Capacity: 1000,
				Type:     "LRU",
			},
			Hasher: config.TypeConfig{Type: "blake2b"},
		},
		SkIndex:                              0,
		ValidatorKeyPemFileName:              "validatorKey.pem",
		CoreComponentsHolder:                 coreComponents,
		ActivateBLSPubKeyMessageVerification: false,
		KeyLoader: &mock.KeyLoaderStub{
			LoadKeyCalled: DummyLoadSkPkFromPemFile([]byte(DummySk), DummyPk, nil),
		},
		EnableEpochs: config.EnableEpochs{
			BLSMultiSignerEnableEpoch: []config.MultiSignerConfig{{EnableEpoch: 0, Type: "no-KOSK"}},
			MaxNodesChangeEnableEpoch: []config.MaxNodesChangeConfig{
				{
					EpochEnable:            0,
					MaxNumNodes:            100,
					NodesToShufflePerShard: 2,
				},
			},
		},
	}

	return args
}

// GetDataArgs -
func GetDataArgs(coreComponents factory.CoreComponentsHolder, shardCoordinator sharding.Coordinator) dataComp.DataComponentsFactoryArgs {
	runTypeComponents := GetRunTypeComponents()

	return dataComp.DataComponentsFactoryArgs{
		Config: testscommon.GetGeneralConfig(),
		PrefsConfig: config.PreferencesConfig{
			FullArchive: false,
		},
		ShardCoordinator:                shardCoordinator,
		Core:                            coreComponents,
		StatusCore:                      GetStatusCoreComponents(),
		Crypto:                          GetCryptoComponents(coreComponents),
		CurrentEpoch:                    0,
		CreateTrieEpochRootHashStorer:   false,
		NodeProcessingMode:              common.Normal,
		FlagsConfigs:                    config.ContextFlagsConfig{},
		AdditionalStorageServiceCreator: runTypeComponents.AdditionalStorageServiceCreator(),
	}
}

// GetCoreComponents -
func GetCoreComponents() factory.CoreComponentsHolder {
	coreArgs := GetCoreArgs()
	return createCoreComponents(coreArgs)
}

// GetSovereignCoreComponents -
func GetSovereignCoreComponents() factory.CoreComponentsHolder {
	sovRunTypeCoreComponents := GetSovereignRunTypeCoreComponents()
	coreArgs := GetCoreArgs()
	coreArgs.NodesFilename = "../mock/testdata/sovereignNodesSetupMock.json"
	coreArgs.RunTypeCoreComponents = sovRunTypeCoreComponents
	return createCoreComponents(coreArgs)
}

func createCoreComponents(coreArgs coreComp.CoreComponentsFactoryArgs) factory.CoreComponentsHolder {
	coreComponentsFactory, _ := coreComp.NewCoreComponentsFactory(coreArgs)
	coreComponents, err := coreComp.NewManagedCoreComponents(coreComponentsFactory)
	if err != nil {
		fmt.Println("getCoreComponents NewManagedCoreComponents", "error", err.Error())
		return nil
	}
	err = coreComponents.Create()
	if err != nil {
		fmt.Println("getCoreComponents Create", "error", err.Error())
	}
	return coreComponents
}

// GetNetworkFactoryArgs -
func GetNetworkFactoryArgs() networkComp.NetworkComponentsFactoryArgs {
	p2pCfg := p2pConfig.P2PConfig{
		Node: p2pConfig.NodeConfig{
			Port: "0",
			Transports: p2pConfig.P2PTransportConfig{
				TCP: p2pConfig.P2PTCPTransport{
					ListenAddress: p2p.LocalHostListenAddrWithIp4AndTcp,
				},
			},
			ResourceLimiter: p2pConfig.P2PResourceLimiterConfig{
				Type: p2p.DefaultWithScaleResourceLimiter,
			},
		},
		KadDhtPeerDiscovery: p2pConfig.KadDhtPeerDiscoveryConfig{
			Enabled:                          false,
			Type:                             "optimized",
			RefreshIntervalInSec:             10,
			ProtocolIDs:                      []string{"erd/kad/1.0.0"},
			InitialPeerList:                  []string{"peer0", "peer1"},
			BucketSize:                       10,
			RoutingTableRefreshIntervalInSec: 5,
		},
		Sharding: p2pConfig.ShardingConfig{
			TargetPeerCount:         10,
			MaxIntraShardValidators: 10,
			MaxCrossShardValidators: 10,
			MaxIntraShardObservers:  10,
			MaxCrossShardObservers:  10,
			MaxSeeders:              2,
			Type:                    "NilListSharder",
		},
	}

	mainConfig := config.Config{
		PeerHonesty: config.CacheConfig{
			Type:     "LRU",
			Capacity: 5000,
			Shards:   16,
		},
		Debug: config.DebugConfig{
			Antiflood: config.AntifloodDebugConfig{
				Enabled:                    true,
				CacheSize:                  100,
				IntervalAutoPrintInSeconds: 1,
			},
		},
		PeersRatingConfig: config.PeersRatingConfig{
			TopRatedCacheCapacity: 1000,
			BadRatedCacheCapacity: 1000,
		},
		PoolsCleanersConfig: config.PoolsCleanersConfig{
			MaxRoundsToKeepUnprocessedMiniBlocks:   50,
			MaxRoundsToKeepUnprocessedTransactions: 50,
		},
	}

	appStatusHandler := statusHandlerMock.NewAppStatusHandlerMock()

	cryptoCompMock := GetDefaultCryptoComponents()

	return networkComp.NetworkComponentsFactoryArgs{
		MainP2pConfig:     p2pCfg,
		NodeOperationMode: common.NormalOperation,
		MainConfig:        mainConfig,
		StatusHandler:     appStatusHandler,
		Marshalizer:       &mock.MarshalizerMock{},
		RatingsConfig: config.RatingsConfig{
			General:    config.General{},
			ShardChain: config.ShardChain{},
			MetaChain:  config.MetaChain{},
			PeerHonesty: config.PeerHonestyConfig{
				DecayCoefficient:             0.9779,
				DecayUpdateIntervalInSeconds: 10,
				MaxScore:                     100,
				MinScore:                     -100,
				BadPeerThreshold:             -80,
				UnitValue:                    1.0,
			},
		},
		Syncer:           &p2pFactory.LocalSyncTimer{},
		CryptoComponents: cryptoCompMock,
	}
}

// GetStateFactoryArgs -
func GetStateFactoryArgs(coreComponents factory.CoreComponentsHolder, statusCoreComp factory.StatusCoreComponentsHolder) stateComp.StateComponentsFactoryArgs {
	tsm, _ := trie.NewTrieStorageManager(storage.GetStorageManagerArgs())
	storageManagerUser, _ := trie.NewTrieStorageManagerWithoutPruning(tsm)
	tsm, _ = trie.NewTrieStorageManager(storage.GetStorageManagerArgs())
	storageManagerPeer, _ := trie.NewTrieStorageManagerWithoutPruning(tsm)

	trieStorageManagers := make(map[string]common.StorageManager)
	trieStorageManagers[dataRetriever.UserAccountsUnit.String()] = storageManagerUser
	trieStorageManagers[dataRetriever.PeerAccountsUnit.String()] = storageManagerPeer

	triesHolder := state.NewDataTriesHolder()
	trieUsers, _ := trie.NewTrie(storageManagerUser, coreComponents.InternalMarshalizer(), coreComponents.Hasher(), coreComponents.EnableEpochsHandler(), 5)
	triePeers, _ := trie.NewTrie(storageManagerPeer, coreComponents.InternalMarshalizer(), coreComponents.Hasher(), coreComponents.EnableEpochsHandler(), 5)
	triesHolder.Put([]byte(dataRetriever.UserAccountsUnit.String()), trieUsers)
	triesHolder.Put([]byte(dataRetriever.PeerAccountsUnit.String()), triePeers)

	stateComponentsFactoryArgs := stateComp.StateComponentsFactoryArgs{
		Config:          GetGeneralConfig(),
		Core:            coreComponents,
		StatusCore:      statusCoreComp,
		StorageService:  disabled.NewChainStorer(),
		ProcessingMode:  common.Normal,
		ChainHandler:    &testscommon.ChainHandlerStub{},
		AccountsCreator: GetRunTypeComponents().AccountsCreator(),
	}

	return stateComponentsFactoryArgs
}

// GetProcessComponentsFactoryArgs -
func GetProcessComponentsFactoryArgs(shardCoordinator sharding.Coordinator) processComp.ProcessComponentsFactoryArgs {
	coreComponents := GetCoreComponents()
	cryptoComponents := GetCryptoComponents(coreComponents)
	networkComponents := GetNetworkComponents(cryptoComponents)
	dataComponents := GetDataComponents(coreComponents, shardCoordinator)
	stateComponents := GetStateComponents(coreComponents, GetStatusCoreComponents())
	processArgs := GetProcessArgs(
		shardCoordinator,
		coreComponents,
		dataComponents,
		cryptoComponents,
		stateComponents,
		networkComponents,
	)
	return processArgs
}

// GetSovereignProcessComponentsFactoryArgs -
func GetSovereignProcessComponentsFactoryArgs(shardCoordinator sharding.Coordinator) processComp.ProcessComponentsFactoryArgs {
	coreComponents := GetSovereignCoreComponents()
	cryptoComponents := GetCryptoComponents(coreComponents)
	networkComponents := GetNetworkComponents(cryptoComponents)
	dataComponents := GetSovereignDataComponents(coreComponents, shardCoordinator)
	stateComponents := GetSovereignStateComponents(coreComponents, GetSovereignStatusCoreComponents())
	processArgs := GetSovereignProcessArgs(
		shardCoordinator,
		coreComponents,
		dataComponents,
		cryptoComponents,
		stateComponents,
		networkComponents,
	)

	return processArgs
}

// GetBootStrapFactoryArgs -
func GetBootStrapFactoryArgs() bootstrapComp.BootstrapComponentsFactoryArgs {
	coreComponents := GetCoreComponents()
	cryptoComponents := GetCryptoComponents(coreComponents)
	networkComponents := GetNetworkComponents(cryptoComponents)
	statusCoreComponents := GetStatusCoreComponents()
	return bootstrapComp.BootstrapComponentsFactoryArgs{
		Config:               testscommon.GetGeneralConfig(),
		WorkingDir:           "home",
		CoreComponents:       coreComponents,
		CryptoComponents:     cryptoComponents,
		NetworkComponents:    networkComponents,
		StatusCoreComponents: statusCoreComponents,
		PrefConfig: config.Preferences{
			Preferences: config.PreferencesConfig{
				DestinationShardAsObserver: "0",
				ConnectionWatcherType:      "print",
			},
		},
		ImportDbConfig: config.ImportDbConfig{
			IsImportDBMode: false,
		},
		RoundConfig: config.RoundConfig{},
		FlagsConfig: config.ContextFlagsConfig{
			ForceStartFromNetwork: false,
		},
		RunTypeComponents: GetRunTypeComponents(),
	}
}

// GetProcessArgs -
func GetProcessArgs(
	shardCoordinator sharding.Coordinator,
	coreComponents factory.CoreComponentsHolder,
	dataComponents factory.DataComponentsHolder,
	cryptoComponents factory.CryptoComponentsHolder,
	stateComponents factory.StateComponentsHolder,
	networkComponents factory.NetworkComponentsHolder,
) processComp.ProcessComponentsFactoryArgs {
	gasSchedule := wasmConfig.MakeGasMapForTests()
	// TODO: check if these could be initialized by MakeGasMapForTests()
	gasSchedule["BuiltInCost"]["SaveUserName"] = 1
	gasSchedule["BuiltInCost"]["SaveKeyValue"] = 1
	gasSchedule["BuiltInCost"]["ESDTTransfer"] = 1
	gasSchedule["BuiltInCost"]["ESDTBurn"] = 1
	gasSchedule[common.MetaChainSystemSCsCost] = FillGasMapMetaChainSystemSCsCosts(1)

	gasScheduleNotifier := &testscommon.GasScheduleNotifierMock{
		GasSchedule: gasSchedule,
	}

	nc := &shardingMocks.NodesCoordinatorMock{}
	statusComponents := GetStatusComponents(
		coreComponents,
		networkComponents,
		stateComponents,
		shardCoordinator,
		nc,
	)

	bootstrapComponentsFactoryArgs := GetBootStrapFactoryArgs()
	bootstrapComponentsFactory, err := bootstrapComp.NewBootstrapComponentsFactory(bootstrapComponentsFactoryArgs)
	if err != nil {
		panic(err)
	}

	bootstrapComponents, err := bootstrapComp.NewTestManagedBootstrapComponents(bootstrapComponentsFactory)
	if err != nil {
		panic(err)
	}

	err = bootstrapComponents.Create()
	if err != nil {
		panic(err)
	}

	err = bootstrapComponents.SetShardCoordinator(shardCoordinator)
	if err != nil {
		panic(err)
	}

	args := processComp.ProcessComponentsFactoryArgs{
		Config:                 testscommon.GetGeneralConfig(),
		SmartContractParser:    &mock.SmartContractParserStub{},
		GasSchedule:            gasScheduleNotifier,
		NodesCoordinator:       nc,
		Data:                   dataComponents,
		CoreData:               coreComponents,
		Crypto:                 cryptoComponents,
		State:                  stateComponents,
		Network:                networkComponents,
		StatusComponents:       statusComponents,
		BootstrapComponents:    bootstrapComponents,
		StatusCoreComponents:   GetStatusCoreComponents(),
		RequestedItemsHandler:  &testscommon.RequestedItemsHandlerStub{},
		WhiteListHandler:       &testscommon.WhiteListHandlerStub{},
		WhiteListerVerifiedTxs: &testscommon.WhiteListHandlerStub{},
		MaxRating:              100,
		ImportStartHandler:     &testscommon.ImportStartHandlerStub{},
		SystemSCConfig: &config.SystemSmartContractsConfig{
			ESDTSystemSCConfig: config.ESDTSystemSCConfig{
				BaseIssuingCost: "1000",
				OwnerAddress:    "erd1fpkcgel4gcmh8zqqdt043yfcn5tyx8373kg6q2qmkxzu4dqamc0swts65c",
			},
			GovernanceSystemSCConfig: config.GovernanceSystemSCConfig{
				V1: config.GovernanceSystemSCConfigV1{
					ProposalCost:     "500",
					NumNodes:         100,
					MinQuorum:        50,
					MinPassThreshold: 50,
					MinVetoThreshold: 50,
				},
				Active: config.GovernanceSystemSCConfigActive{
					ProposalCost:     "500",
					MinQuorum:        0.5,
					MinPassThreshold: 0.5,
					MinVetoThreshold: 0.5,
					LostProposalFee:  "1",
				},
				OwnerAddress: "erd1vxy22x0fj4zv6hktmydg8vpfh6euv02cz4yg0aaws6rrad5a5awqgqky80",
			},
			StakingSystemSCConfig: config.StakingSystemSCConfig{
				GenesisNodePrice:                     "2500000000000000000000",
				MinStakeValue:                        "1",
				UnJailValue:                          "1",
				MinStepValue:                         "1",
				UnBondPeriod:                         0,
				NumRoundsWithoutBleed:                0,
				MaximumPercentageToBleed:             0,
				BleedPercentagePerRound:              0,
				MaxNumberOfNodesForStake:             10,
				ActivateBLSPubKeyMessageVerification: false,
				MinUnstakeTokensValue:                "1",
				StakeLimitPercentage:                 100,
				NodeLimitPercentage:                  100,
			},
			DelegationManagerSystemSCConfig: config.DelegationManagerSystemSCConfig{
				MinCreationDeposit:  "100",
				MinStakeAmount:      "100",
				ConfigChangeAddress: "erd1vxy22x0fj4zv6hktmydg8vpfh6euv02cz4yg0aaws6rrad5a5awqgqky80",
			},
			DelegationSystemSCConfig: config.DelegationSystemSCConfig{
				MinServiceFee: 0,
				MaxServiceFee: 100,
			},
			SoftAuctionConfig: config.SoftAuctionConfig{
				TopUpStep:             "10",
				MinTopUp:              "1",
				MaxTopUp:              "32000000",
				MaxNumberOfIterations: 100000,
			},
		},
		HistoryRepo: &dblookupext.HistoryRepositoryStub{},
		FlagsConfig: config.ContextFlagsConfig{
			Version: "v1.0.0",
		},
		RoundConfig:             testscommon.GetDefaultRoundsConfig(),
		TxExecutionOrderHandler: &commonMocks.TxExecutionOrderHandlerStub{},
		EpochConfig: config.EpochConfig{
			EnableEpochs: config.EnableEpochs{
				MaxNodesChangeEnableEpoch: []config.MaxNodesChangeConfig{
					{
						EpochEnable:            0,
						MaxNumNodes:            100,
						NodesToShufflePerShard: 2,
					},
				},
			},
		},
		IncomingHeaderSubscriber: &sovereign.IncomingHeaderSubscriberStub{},
	}

	initialAccounts := createAccounts()
	runTypeComponents := GetRunTypeComponentsStub(GetRunTypeComponents())
	runTypeComponents.AccountParser = &mock.AccountsParserStub{
		InitialAccountsCalled: func() []genesis.InitialAccountHandler {
			return initialAccounts
		},
		GenerateInitialTransactionsCalled: func(shardCoordinator sharding.Coordinator, initialIndexingData map[uint32]*genesis.IndexingData) ([]*block.MiniBlock, map[uint32]*outport.TransactionPool, error) {
			txsPool := make(map[uint32]*outport.TransactionPool)
			for i := uint32(0); i < shardCoordinator.NumberOfShards(); i++ {
				txsPool[i] = &outport.TransactionPool{}
			}

			return make([]*block.MiniBlock, 4), txsPool, nil
		},
		InitialAccountsSplitOnAddressesShardsCalled: func(shardCoordinator sharding.Coordinator) (map[uint32][]genesis.InitialAccountHandler, error) {
			return map[uint32][]genesis.InitialAccountHandler{
				0: initialAccounts,
			}, nil
		},
	}
	args.RunTypeComponents = runTypeComponents
	args.EnableEpochsFactory = GetRunTypeCoreComponents().EnableEpochsFactoryCreator()
	return args
}

func GetSovereignProcessArgs(
	shardCoordinator sharding.Coordinator,
	coreComponents factory.CoreComponentsHolder,
	dataComponents factory.DataComponentsHolder,
	cryptoComponents factory.CryptoComponentsHolder,
	stateComponents factory.StateComponentsHolder,
	networkComponents factory.NetworkComponentsHolder,
) processComp.ProcessComponentsFactoryArgs {
	processArgs := GetProcessArgs(
		shardCoordinator,
		coreComponents,
		dataComponents,
		cryptoComponents,
		stateComponents,
		networkComponents,
	)

	initialAccounts := createSovereignAccounts()
	runTypeComponents := GetRunTypeComponentsStub(GetSovereignRunTypeComponents())
	runTypeComponents.AccountParser = &mock.AccountsParserStub{
		InitialAccountsCalled: func() []genesis.InitialAccountHandler {
			return initialAccounts
		},
		GenerateInitialTransactionsCalled: func(shardCoordinator sharding.Coordinator, initialIndexingData map[uint32]*genesis.IndexingData) ([]*block.MiniBlock, map[uint32]*outport.TransactionPool, error) {
			txsPool := make(map[uint32]*outport.TransactionPool)
			for i := uint32(0); i < shardCoordinator.NumberOfShards(); i++ {
				txsPool[i] = &outport.TransactionPool{}
			}

			return make([]*block.MiniBlock, 4), txsPool, nil
		},
		InitialAccountsSplitOnAddressesShardsCalled: func(shardCoordinator sharding.Coordinator) (map[uint32][]genesis.InitialAccountHandler, error) {
			return map[uint32][]genesis.InitialAccountHandler{
				0: initialAccounts,
			}, nil
		},
	}

	bootstrapComponentsFactoryArgs := GetBootStrapFactoryArgs()
	bootstrapComponentsFactoryArgs.RunTypeComponents = runTypeComponents
	bootstrapComponentsFactory, _ := bootstrapComp.NewBootstrapComponentsFactory(bootstrapComponentsFactoryArgs)
	bootstrapComponents, _ := bootstrapComp.NewTestManagedBootstrapComponents(bootstrapComponentsFactory)
	_ = bootstrapComponents.Create()
	_ = bootstrapComponents.SetShardCoordinator(shardCoordinator)

	statusCoreComponents := GetSovereignStatusCoreComponents()

	processArgs.BootstrapComponents = bootstrapComponents
	processArgs.StatusCoreComponents = statusCoreComponents
	processArgs.IncomingHeaderSubscriber = &sovereign.IncomingHeaderSubscriberStub{}
	processArgs.RunTypeComponents = runTypeComponents
	processArgs.EnableEpochsFactory = GetSovereignRunTypeCoreComponents().EnableEpochsFactoryCreator()

	return processArgs
}

// GetStatusComponents -
func GetStatusComponents(
	coreComponents factory.CoreComponentsHolder,
	networkComponents factory.NetworkComponentsHolder,
	stateComponents factory.StateComponentsHolder,
	shardCoordinator sharding.Coordinator,
	nodesCoordinator nodesCoordinator.NodesCoordinator,
) factory.StatusComponentsHandler {
	indexerURL := "url"
	elasticUsername := "user"
	elasticPassword := "pass"
	statusArgs := statusComp.StatusComponentsFactoryArgs{
		Config: testscommon.GetGeneralConfig(),
		ExternalConfig: config.ExternalConfig{
			ElasticSearchConnector: config.ElasticSearchConfig{
				Enabled:        false,
				URL:            indexerURL,
				Username:       elasticUsername,
				Password:       elasticPassword,
				EnabledIndexes: []string{"transactions", "blocks"},
			},
			EventNotifierConnector: config.EventNotifierConfig{
				Enabled:        false,
				ProxyUrl:       "https://localhost:5000",
				MarshallerType: "json",
			},
		},
		EconomicsConfig:      config.EconomicsConfig{},
		ShardCoordinator:     shardCoordinator,
		NodesCoordinator:     nodesCoordinator,
		EpochStartNotifier:   coreComponents.EpochStartNotifierWithConfirm(),
		CoreComponents:       coreComponents,
		NetworkComponents:    networkComponents,
		StateComponents:      stateComponents,
		IsInImportMode:       false,
		StatusCoreComponents: GetStatusCoreComponents(),
		CryptoComponents:     GetDefaultCryptoComponents(),
	}

	statusComponentsFactory, _ := statusComp.NewStatusComponentsFactory(statusArgs)
	managedStatusComponents, err := statusComp.NewManagedStatusComponents(statusComponentsFactory)
	if err != nil {
		log.Error("getStatusComponents NewManagedStatusComponents", "error", err.Error())
		return nil
	}
	err = managedStatusComponents.Create()
	if err != nil {
		log.Error("getStatusComponents Create", "error", err.Error())
		return nil
	}
	return managedStatusComponents
}

// GetStatusComponentsFactoryArgsAndProcessComponents -
func GetStatusComponentsFactoryArgsAndProcessComponents(shardCoordinator sharding.Coordinator) (statusComp.StatusComponentsFactoryArgs, factory.ProcessComponentsHolder) {
	coreComponents := GetCoreComponents()
	cryptoComponents := GetCryptoComponents(coreComponents)
	networkComponents := GetNetworkComponents(cryptoComponents)
	dataComponents := GetDataComponents(coreComponents, shardCoordinator)
	stateComponents := GetStateComponents(coreComponents, GetStatusCoreComponents())
	processComponents := GetProcessComponents(
		shardCoordinator,
		coreComponents,
		networkComponents,
		dataComponents,
		cryptoComponents,
		stateComponents,
	)
	statusCoreComponents := GetStatusCoreComponents()

	indexerURL := "url"
	elasticUsername := "user"
	elasticPassword := "pass"
	return statusComp.StatusComponentsFactoryArgs{
		Config: testscommon.GetGeneralConfig(),
		ExternalConfig: config.ExternalConfig{
			ElasticSearchConnector: config.ElasticSearchConfig{
				Enabled:        false,
				URL:            indexerURL,
				Username:       elasticUsername,
				Password:       elasticPassword,
				EnabledIndexes: []string{"transactions", "blocks"},
			},
			EventNotifierConnector: config.EventNotifierConfig{
				Enabled:           false,
				ProxyUrl:          "http://localhost:5000",
				RequestTimeoutSec: 30,
				MarshallerType:    "json",
			},
			HostDriversConfig: []config.HostDriversConfig{
				{
					MarshallerType:     "json",
					Mode:               "client",
					URL:                "ws://localhost:12345",
					RetryDurationInSec: 1,
				},
			},
		},
		EconomicsConfig:      config.EconomicsConfig{},
		ShardCoordinator:     mock.NewMultiShardsCoordinatorMock(2),
		NodesCoordinator:     &shardingMocks.NodesCoordinatorMock{},
		EpochStartNotifier:   &mock.EpochStartNotifierStub{},
		CoreComponents:       coreComponents,
		NetworkComponents:    networkComponents,
		StateComponents:      stateComponents,
		StatusCoreComponents: statusCoreComponents,
		IsInImportMode:       false,
		CryptoComponents:     cryptoComponents,
	}, processComponents
}

// GetNetworkComponents -
func GetNetworkComponents(cryptoComp factory.CryptoComponentsHolder) factory.NetworkComponentsHolder {
	networkArgs := GetNetworkFactoryArgs()
	networkArgs.CryptoComponents = cryptoComp
	networkComponentsFactory, _ := networkComp.NewNetworkComponentsFactory(networkArgs)
	networkComponents, _ := networkComp.NewManagedNetworkComponents(networkComponentsFactory)

	_ = networkComponents.Create()

	return networkComponents
}

// GetDataComponents -
func GetDataComponents(coreComponents factory.CoreComponentsHolder, shardCoordinator sharding.Coordinator) factory.DataComponentsHolder {
	dataArgs := GetDataArgs(coreComponents, shardCoordinator)
	return createDataComponents(dataArgs)
}

// GetSovereignDataComponents -
func GetSovereignDataComponents(coreComponents factory.CoreComponentsHolder, shardCoordinator sharding.Coordinator) factory.DataComponentsHolder {
	dataArgs := GetDataArgs(coreComponents, shardCoordinator)
	dataArgs.Crypto = GetCryptoComponents(coreComponents)
	dataArgs.StatusCore = GetSovereignStatusCoreComponents()
	dataArgs.AdditionalStorageServiceCreator = GetSovereignRunTypeComponents().AdditionalStorageServiceCreator()
	return createDataComponents(dataArgs)
}

func createDataComponents(dataArgs dataComp.DataComponentsFactoryArgs) factory.DataComponentsHolder {
	dataComponentsFactory, err := dataComp.NewDataComponentsFactory(dataArgs)
	if err != nil {
		panic(err)
	}

	dataComponents, err := dataComp.NewManagedDataComponents(dataComponentsFactory)
	if err != nil {
		panic(err)
	}

	err = dataComponents.Create()
	if err != nil {
		panic(err)
	}
	return dataComponents
}

// GetCryptoComponents -
func GetCryptoComponents(coreComponents factory.CoreComponentsHolder) factory.CryptoComponentsHolder {
	cryptoArgs := GetCryptoArgs(coreComponents)
	cryptoComponentsFactory, _ := cryptoComp.NewCryptoComponentsFactory(cryptoArgs)
	cryptoComponents, err := cryptoComp.NewManagedCryptoComponents(cryptoComponentsFactory)
	if err != nil {
		log.Error("getCryptoComponents NewManagedCryptoComponents", "error", err.Error())
		return nil
	}

	err = cryptoComponents.Create()
	if err != nil {
		log.Error("getCryptoComponents Create", "error", err.Error())
		return nil
	}
	return cryptoComponents
}

// GetStateComponents -
func GetStateComponents(coreComponents factory.CoreComponentsHolder, statusCoreComponents factory.StatusCoreComponentsHolder) factory.StateComponentsHolder {
	stateArgs := GetStateFactoryArgs(coreComponents, statusCoreComponents)
	return createStateComponents(stateArgs)
}

// GetSovereignStateComponents -
func GetSovereignStateComponents(coreComponents factory.CoreComponentsHolder, statusCoreComponents factory.StatusCoreComponentsHolder) factory.StateComponentsHolder {
	stateArgs := GetStateFactoryArgs(coreComponents, statusCoreComponents)
	stateArgs.AccountsCreator = GetSovereignRunTypeComponents().AccountsCreator()
	return createStateComponents(stateArgs)
}

func createStateComponents(stateArgs stateComp.StateComponentsFactoryArgs) factory.StateComponentsHolder {
	stateComponentsFactory, err := stateComp.NewStateComponentsFactory(stateArgs)
	if err != nil {
		log.Error("GetStateComponents NewStateComponentsFactory", "error", err.Error())
		return nil
	}

	stateComponents, err := stateComp.NewManagedStateComponents(stateComponentsFactory)
	if err != nil {
		log.Error("GetStateComponents NewManagedStateComponents", "error", err.Error())
		return nil
	}
	err = stateComponents.Create()
	if err != nil {
		log.Error("GetStateComponents Create", "error", err.Error())
		return nil
	}
	return stateComponents
}

// GetStatusCoreComponents -
func GetStatusCoreComponents() factory.StatusCoreComponentsHolder {
	args := GetStatusCoreArgs(GetCoreComponents())
	return createStatusCoreComponents(args)
}

// GetSovereignStatusCoreComponents -
func GetSovereignStatusCoreComponents() factory.StatusCoreComponentsHolder {
	args := GetStatusCoreArgs(GetSovereignCoreComponents())
	return createStatusCoreComponents(args)
}

func createStatusCoreComponents(args statusCore.StatusCoreComponentsFactoryArgs) factory.StatusCoreComponentsHolder {
	statusCoreFactory, err := statusCore.NewStatusCoreComponentsFactory(args)
	if err != nil {
		log.Error("GetStatusCoreComponents NewStatusCoreComponentsFactory", "error", err.Error())
		return nil
	}

	statusCoreComponents, err := statusCore.NewManagedStatusCoreComponents(statusCoreFactory)
	if err != nil {
		log.Error("GetStatusCoreComponents NewManagedStatusCoreComponents", "error", err.Error())
		return nil
	}

	err = statusCoreComponents.Create()
	if err != nil {
		log.Error("GetStatusCoreComponents Create", "error", err.Error())
		return nil
	}

	return statusCoreComponents
}

// GetProcessComponents -
func GetProcessComponents(
	shardCoordinator sharding.Coordinator,
	coreComponents factory.CoreComponentsHolder,
	networkComponents factory.NetworkComponentsHolder,
	dataComponents factory.DataComponentsHolder,
	cryptoComponents factory.CryptoComponentsHolder,
	stateComponents factory.StateComponentsHolder,
) factory.ProcessComponentsHolder {
	processArgs := GetProcessArgs(
		shardCoordinator,
		coreComponents,
		dataComponents,
		cryptoComponents,
		stateComponents,
		networkComponents,
	)
	processComponentsFactory, _ := processComp.NewProcessComponentsFactory(processArgs)
	managedProcessComponents, err := processComp.NewManagedProcessComponents(processComponentsFactory)
	if err != nil {
		log.Error("getProcessComponents NewManagedProcessComponents", "error", err.Error())
		return nil
	}
	err = managedProcessComponents.Create()
	if err != nil {
		log.Error("getProcessComponents Create", "error", err.Error())
		return nil
	}
	return managedProcessComponents
}

// GetSovereignProcessComponents -
func GetSovereignProcessComponents(
	shardCoordinator sharding.Coordinator,
	coreComponents factory.CoreComponentsHolder,
	networkComponents factory.NetworkComponentsHolder,
	dataComponents factory.DataComponentsHolder,
	cryptoComponents factory.CryptoComponentsHolder,
	stateComponents factory.StateComponentsHolder,
) factory.ProcessComponentsHolder {
	processArgs := GetSovereignProcessArgs(
		shardCoordinator,
		coreComponents,
		dataComponents,
		cryptoComponents,
		stateComponents,
		networkComponents,
	)
	processComponentsFactory, _ := processComp.NewProcessComponentsFactory(processArgs)
	managedProcessComponents, err := processComp.NewManagedProcessComponents(processComponentsFactory)
	if err != nil {
		log.Error("GetSovereignProcessComponents NewManagedProcessComponents", "error", err.Error())
		return nil
	}
	err = managedProcessComponents.Create()
	if err != nil {
		log.Error("GetSovereignProcessComponents Create", "error", err.Error())
		return nil
	}
	return managedProcessComponents
}

func createAccounts() []genesis.InitialAccountHandler {
	addrConverter, _ := commonFactory.NewPubkeyConverter(config.PubkeyConfig{
		Length:          32,
		Type:            "bech32",
		SignatureLength: 0,
		Hrp:             "erd",
	})
	acc1 := data.InitialAccount{
		Address:      "erd1ulhw20j7jvgfgak5p05kv667k5k9f320sgef5ayxkt9784ql0zssrzyhjp",
		Supply:       big.NewInt(0).Mul(big.NewInt(5000000), big.NewInt(1000000000000000000)),
		Balance:      big.NewInt(0).Mul(big.NewInt(4997500), big.NewInt(1000000000000000000)),
		StakingValue: big.NewInt(0).Mul(big.NewInt(2500), big.NewInt(1000000000000000000)),
		Delegation: &data.DelegationData{
			Address: "",
			Value:   big.NewInt(0),
		},
	}
	acc2 := data.InitialAccount{
		Address:      "erd17c4fs6mz2aa2hcvva2jfxdsrdknu4220496jmswer9njznt22eds0rxlr4",
		Supply:       big.NewInt(0).Mul(big.NewInt(5000000), big.NewInt(1000000000000000000)),
		Balance:      big.NewInt(0).Mul(big.NewInt(4997500), big.NewInt(1000000000000000000)),
		StakingValue: big.NewInt(0).Mul(big.NewInt(2500), big.NewInt(1000000000000000000)),
		Delegation: &data.DelegationData{
			Address: "",
			Value:   big.NewInt(0),
		},
	}
	acc3 := data.InitialAccount{
		Address:      "erd10d2gufxesrp8g409tzxljlaefhs0rsgjle3l7nq38de59txxt8csj54cd3",
		Supply:       big.NewInt(0).Mul(big.NewInt(10000000), big.NewInt(1000000000000000000)),
		Balance:      big.NewInt(0).Mul(big.NewInt(9997500), big.NewInt(1000000000000000000)),
		StakingValue: big.NewInt(0).Mul(big.NewInt(2500), big.NewInt(1000000000000000000)),
		Delegation: &data.DelegationData{
			Address: "",
			Value:   big.NewInt(0),
		},
	}

	acc1Bytes, _ := addrConverter.Decode(acc1.Address)
	acc1.SetAddressBytes(acc1Bytes)
	acc2Bytes, _ := addrConverter.Decode(acc2.Address)
	acc2.SetAddressBytes(acc2Bytes)
	acc3Bytes, _ := addrConverter.Decode(acc3.Address)
	acc3.SetAddressBytes(acc3Bytes)

	return []genesis.InitialAccountHandler{&acc1, &acc2, &acc3}
}

func createSovereignAccounts() []genesis.InitialAccountHandler {
	addrConverter, _ := commonFactory.NewPubkeyConverter(config.PubkeyConfig{
		Length:          32,
		Type:            "bech32",
		SignatureLength: 0,
		Hrp:             "erd",
	})
	acc1 := data.InitialAccount{
		Address:      "erd1whq0zspt6ktnv37gqj303da0vygyqwf5q52m7erftd0rl7laygfs6rhpct",
		Supply:       big.NewInt(0).Mul(big.NewInt(10000000), big.NewInt(1000000000000000000)),
		Balance:      big.NewInt(0).Mul(big.NewInt(9997500), big.NewInt(1000000000000000000)),
		StakingValue: big.NewInt(0).Mul(big.NewInt(2500), big.NewInt(1000000000000000000)),
		Delegation: &data.DelegationData{
			Address: "",
			Value:   big.NewInt(0),
		},
	}
	acc2 := data.InitialAccount{
		Address:      "erd129ppuuvtylghsx7muf29xnzw5lm9v2v8h4942ynymjpu2ftycgtq0rgq3h",
		Supply:       big.NewInt(0).Mul(big.NewInt(10000000), big.NewInt(1000000000000000000)),
		Balance:      big.NewInt(0).Mul(big.NewInt(9997500), big.NewInt(1000000000000000000)),
		StakingValue: big.NewInt(0).Mul(big.NewInt(2500), big.NewInt(1000000000000000000)),
		Delegation: &data.DelegationData{
			Address: "",
			Value:   big.NewInt(0),
		},
	}

	acc1Bytes, _ := addrConverter.Decode(acc1.Address)
	acc1.SetAddressBytes(acc1Bytes)

	acc2Bytes, _ := addrConverter.Decode(acc2.Address)
	acc2.SetAddressBytes(acc2Bytes)
	return []genesis.InitialAccountHandler{&acc1, &acc2}
}

func createArgsRunTypeComponents() runType.ArgsRunTypeComponents {
	generalCfg := GetGeneralConfig()
	return runType.ArgsRunTypeComponents{
		CoreComponents: &mockCoreComp.CoreComponentsStub{
			HasherField:                 &hashingMocks.HasherMock{},
			InternalMarshalizerField:    &marshallerMock.MarshalizerMock{},
			EnableEpochsHandlerField:    &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			AddressPubKeyConverterField: &testscommon.PubkeyConverterStub{},
		},
		CryptoComponents: &mockCoreComp.CryptoComponentsStub{
			TxKeyGen: &mockCoreComp.KeyGenMock{},
			BlockSig: &mockConsensus.SingleSignerMock{},
		},
		Configs: config.Configs{
			EconomicsConfig: &config.EconomicsConfig{
				GlobalSettings: config.GlobalSettings{
					GenesisTotalSupply:          "20000000000000000000000000",
					GenesisMintingSenderAddress: "erd17rc0pu8s7rc0pu8s7rc0pu8s7rc0pu8s7rc0pu8s7rc0pu8s7rcqqkhty3",
				},
			},
			GeneralConfig: &generalCfg,
		},
		InitialAccounts: createAccounts(),
	}
}

func GetRunTypeComponentsStub(rt factory.RunTypeComponentsHandler) *mainFactoryMocks.RunTypeComponentsStub {
	return &mainFactoryMocks.RunTypeComponentsStub{
		BlockChainHookHandlerFactory:                rt.BlockChainHookHandlerCreator(),
		BlockProcessorFactory:                       rt.BlockProcessorCreator(),
		BlockTrackerFactory:                         rt.BlockTrackerCreator(),
		BootstrapperFromStorageFactory:              rt.BootstrapperFromStorageCreator(),
		BootstrapperFactory:                         rt.BootstrapperCreator(),
		EpochStartBootstrapperFactory:               rt.EpochStartBootstrapperCreator(),
		ForkDetectorFactory:                         rt.ForkDetectorCreator(),
		HeaderValidatorFactory:                      rt.HeaderValidatorCreator(),
		RequestHandlerFactory:                       rt.RequestHandlerCreator(),
		ScheduledTxsExecutionFactory:                rt.ScheduledTxsExecutionCreator(),
		TransactionCoordinatorFactory:               rt.TransactionCoordinatorCreator(),
		ValidatorStatisticsProcessorFactory:         rt.ValidatorStatisticsProcessorCreator(),
		AdditionalStorageServiceFactory:             rt.AdditionalStorageServiceCreator(),
		SCResultsPreProcessorFactory:                rt.SCResultsPreProcessorCreator(),
		SCProcessorFactory:                          rt.SCProcessorCreator(),
		ConsensusModelType:                          rt.ConsensusModel(),
		VmContainerMetaFactory:                      rt.VmContainerMetaFactoryCreator(),
		VmContainerShardFactory:                     rt.VmContainerShardFactoryCreator(),
		AccountParser:                               rt.AccountsParser(),
		AccountCreator:                              rt.AccountsCreator(),
		VMContextCreatorHandler:                     rt.VMContextCreator(),
		OutGoingOperationsPool:                      rt.OutGoingOperationsPoolHandler(),
		DataCodec:                                   rt.DataCodecHandler(),
		TopicsChecker:                               rt.TopicsCheckerHandler(),
		ShardCoordinatorFactory:                     rt.ShardCoordinatorCreator(),
		NodesCoordinatorWithRaterFactory:            rt.NodesCoordinatorWithRaterCreator(),
		RequestersContainerFactory:                  rt.RequestersContainerFactoryCreator(),
		InterceptorsContainerFactory:                rt.InterceptorsContainerFactoryCreator(),
		ShardResolversContainerFactory:              rt.ShardResolversContainerFactoryCreator(),
		TxPreProcessorFactory:                       rt.TxPreProcessorCreator(),
		ExtraHeaderSigVerifier:                      rt.ExtraHeaderSigVerifierHolder(),
		GenesisBlockFactory:                         rt.GenesisBlockCreatorFactory(),
		GenesisMetaBlockChecker:                     rt.GenesisMetaBlockCheckerCreator(),
		NodesSetupCheckerFactoryField:               rt.NodesSetupCheckerFactory(),
		EpochStartTriggerFactoryField:               rt.EpochStartTriggerFactory(),
		LatestDataProviderFactoryField:              rt.LatestDataProviderFactory(),
		StakingToPeerFactoryField:                   rt.StakingToPeerFactory(),
		ValidatorInfoCreatorFactoryField:            rt.ValidatorInfoCreatorFactory(),
		APIProcessorCompsCreatorHandlerField:        rt.ApiProcessorCompsCreatorHandler(),
		EndOfEpochEconomicsFactoryHandlerField:      rt.EndOfEpochEconomicsFactoryHandler(),
		RewardsCreatorFactoryField:                  rt.RewardsCreatorFactory(),
		SystemSCProcessorFactoryField:               rt.SystemSCProcessorFactory(),
		PreProcessorsContainerFactoryCreatorField:   rt.PreProcessorsContainerFactoryCreator(),
		DataRetrieverContainersSetterField:          rt.DataRetrieverContainersSetter(),
		ShardMessengerFactoryField:                  rt.BroadCastShardMessengerFactoryHandler(),
		ExportHandlerFactoryCreatorField:            rt.ExportHandlerFactoryCreator(),
		ValidatorAccountsSyncerFactoryHandlerField:  rt.ValidatorAccountsSyncerFactoryHandler(),
		ShardRequestersContainerCreatorHandlerField: rt.ShardRequestersContainerCreatorHandler(),
		APIRewardsTxHandlerField:                    rt.APIRewardsTxHandler(),
		OutportDataProviderFactoryField:             rt.OutportDataProviderFactory(),
		DelegatedListFactoryField:                   rt.DelegatedListFactoryHandler(),
		DirectStakedListFactoryField:                rt.DirectStakedListFactoryHandler(),
		TotalStakedValueFactoryField:                rt.TotalStakedValueFactoryHandler(),
		VersionedHeaderFactoryField:                 rt.VersionedHeaderFactory(),
	}
}

// GetRunTypeComponents -
func GetRunTypeComponents() factory.RunTypeComponentsHolder {
	runTypeComponentsFactory, _ := runType.NewRunTypeComponentsFactory(createArgsRunTypeComponents())
	managedRunTypeComponents, err := runType.NewManagedRunTypeComponents(runTypeComponentsFactory)
	if err != nil {
		log.Error("getRunTypeComponents NewManagedRunTypeComponents", "error", err.Error())
		return nil
	}
	err = managedRunTypeComponents.Create()
	if err != nil {
		log.Error("getRunTypeComponents Create", "error", err.Error())
		return nil
	}
	return managedRunTypeComponents
}

// GetRunTypeComponentsWithCoreComp -
func GetRunTypeComponentsWithCoreComp(coreComponents factory.CoreComponentsHolder) factory.RunTypeComponentsHolder {
	args := runType.ArgsRunTypeComponents{
		CoreComponents: coreComponents,
		CryptoComponents: &mockCoreComp.CryptoComponentsStub{
			TxKeyGen: &mockCoreComp.KeyGenMock{},
		},
		Configs: config.Configs{
			EconomicsConfig: &config.EconomicsConfig{
				GlobalSettings: config.GlobalSettings{
					GenesisTotalSupply:          "20000000000000000000000000",
					GenesisMintingSenderAddress: "erd17rc0pu8s7rc0pu8s7rc0pu8s7rc0pu8s7rc0pu8s7rc0pu8s7rcqqkhty3",
				},
			},
			GeneralConfig: &config.Config{
				Versions: config.VersionsConfig{
					VersionsByEpochs: []config.VersionByEpochs{
						{
							Version: "*",
						},
					},
					Cache: config.CacheConfig{
						Type:     "LRU",
						Capacity: 10000,
						Shards:   1,
					},
					DefaultVersion: "*",
				},
			},
		},
		InitialAccounts: createAccounts(),
	}

	runTypeComponentsFactory, _ := runType.NewRunTypeComponentsFactory(args)
	managedRunTypeComponents, err := runType.NewManagedRunTypeComponents(runTypeComponentsFactory)
	if err != nil {
		log.Error("getRunTypeComponents NewManagedRunTypeComponents", "error", err.Error())
		return nil
	}
	err = managedRunTypeComponents.Create()
	if err != nil {
		log.Error("getRunTypeComponents Create", "error", err.Error())
		return nil
	}
	return managedRunTypeComponents
}

// GetSovereignRunTypeComponents -
func GetSovereignRunTypeComponents() factory.RunTypeComponentsHolder {
	sovereignComponentsFactory, _ := runType.NewSovereignRunTypeComponentsFactory(createSovRunTypeArgs())
	managedRunTypeComponents, err := runType.NewManagedRunTypeComponents(sovereignComponentsFactory)
	if err != nil {
		log.Error("getRunTypeComponents NewManagedRunTypeComponents", "error", err.Error())
		return nil
	}
	err = managedRunTypeComponents.Create()
	if err != nil {
		log.Error("getRunTypeComponents Create", "error", err.Error())
		return nil
	}
	return managedRunTypeComponents
}

func createSovRunTypeArgs() runType.ArgsSovereignRunTypeComponents {
	runTypeComponentsFactory, _ := runType.NewRunTypeComponentsFactory(createArgsRunTypeComponents())
	return runType.ArgsSovereignRunTypeComponents{
		RunTypeComponentsFactory: runTypeComponentsFactory,
		Config: config.SovereignConfig{
			GenesisConfig: config.GenesisConfig{
				NativeESDT: "WEGLD-ab47da",
			},
		},
		DataCodec:     &sovereign.DataCodecMock{},
		TopicsChecker: &sovereign.TopicsCheckerMock{},
	}
}

// DummyLoadSkPkFromPemFile -
func DummyLoadSkPkFromPemFile(sk []byte, pk string, err error) LoadKeysFunc {
	return func(_ string, _ int) ([]byte, string, error) {
		return sk, pk, err
	}
}

// FillGasMapMetaChainSystemSCsCosts -
func FillGasMapMetaChainSystemSCsCosts(value uint64) map[string]uint64 {
	gasMap := make(map[string]uint64)
	gasMap["Stake"] = value
	gasMap["UnStake"] = value
	gasMap["UnBond"] = value
	gasMap["Claim"] = value
	gasMap["Get"] = value
	gasMap["ChangeRewardAddress"] = value
	gasMap["ChangeValidatorKeys"] = value
	gasMap["UnJail"] = value
	gasMap["ESDTIssue"] = value
	gasMap["ESDTOperations"] = value
	gasMap["Proposal"] = value
	gasMap["Vote"] = value
	gasMap["DelegateVote"] = value
	gasMap["RevokeVote"] = value
	gasMap["CloseProposal"] = value
	gasMap["DelegationOps"] = value
	gasMap["UnStakeTokens"] = value
	gasMap["UnBondTokens"] = value
	gasMap["DelegationMgrOps"] = value
	gasMap["GetAllNodeStates"] = value
	gasMap["ValidatorToDelegation"] = value
	gasMap["GetActiveFund"] = value
	gasMap["FixWaitingListSize"] = value

	return gasMap
}

// SetShardCoordinator -
func SetShardCoordinator(t *testing.T, bootstrapComponents factory.BootstrapComponentsHolder, coordinator sharding.Coordinator) {
	type testBootstrapComponents interface {
		SetShardCoordinator(shardCoordinator sharding.Coordinator) error
	}

	testBootstrap, ok := bootstrapComponents.(testBootstrapComponents)
	require.True(t, ok)

	_ = testBootstrap.SetShardCoordinator(coordinator)
}
