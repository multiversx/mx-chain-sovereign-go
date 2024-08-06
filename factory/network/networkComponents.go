package network

import (
	"context"
	"fmt"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/marshal"
	logger "github.com/multiversx/mx-chain-logger-go"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/consensus"
	p2pDebug "github.com/multiversx/mx-chain-go/debug/p2p"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/factory/disabled"
	"github.com/multiversx/mx-chain-go/p2p"
	p2pConfig "github.com/multiversx/mx-chain-go/p2p/config"
	p2pDisabled "github.com/multiversx/mx-chain-go/p2p/disabled"
	p2pFactory "github.com/multiversx/mx-chain-go/p2p/factory"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/rating/peerHonesty"
	antifloodFactory "github.com/multiversx/mx-chain-go/process/throttle/antiflood/factory"
	"github.com/multiversx/mx-chain-go/storage/cache"
	storageFactory "github.com/multiversx/mx-chain-go/storage/factory"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
)

// NetworkComponentsFactoryArgs holds the arguments to create a network component handler instance
type NetworkComponentsFactoryArgs struct {
	MainP2pConfig         p2pConfig.P2PConfig
	FullArchiveP2pConfig  p2pConfig.P2PConfig
	LightClientP2pConfig  p2pConfig.P2PConfig
	MainConfig            config.Config
	RatingsConfig         config.RatingsConfig
	StatusHandler         core.AppStatusHandler
	Marshalizer           marshal.Marshalizer
	Syncer                p2p.SyncTimer
	PreferredPeersSlices  []string
	BootstrapWaitTime     time.Duration
	NodeOperationModes    []common.NodeOperation
	ConnectionWatcherType string
	CryptoComponents      factory.CryptoComponentsHolder
}

type networkComponentsFactory struct {
	mainP2PConfig         p2pConfig.P2PConfig
	fullArchiveP2PConfig  p2pConfig.P2PConfig
	lightClientP2PConfig  p2pConfig.P2PConfig
	mainConfig            config.Config
	ratingsConfig         config.RatingsConfig
	statusHandler         core.AppStatusHandler
	marshalizer           marshal.Marshalizer
	syncer                p2p.SyncTimer
	preferredPeersSlices  []string
	bootstrapWaitTime     time.Duration
	nodeOperationModes    []common.NodeOperation
	connectionWatcherType string
	cryptoComponents      factory.CryptoComponentsHolder
}

type networkComponentsHolder struct {
	netMessenger         p2p.Messenger
	preferredPeersHolder p2p.PreferredPeersHolderHandler
}

// networkComponents struct holds the network components
type networkComponents struct {
	mainNetworkHolder        networkComponentsHolder
	fullArchiveNetworkHolder networkComponentsHolder
	lightClientNetworkHolder networkComponentsHolder
	peersRatingHandler       p2p.PeersRatingHandler
	peersRatingMonitor       p2p.PeersRatingMonitor
	inputAntifloodHandler    factory.P2PAntifloodHandler
	outputAntifloodHandler   factory.P2PAntifloodHandler
	pubKeyTimeCacher         process.TimeCacher
	topicFloodPreventer      process.TopicFloodPreventer
	floodPreventers          []process.FloodPreventer
	peerBlackListHandler     process.PeerBlackListCacher
	antifloodConfig          config.AntifloodConfig
	peerHonestyHandler       consensus.PeerHonestyHandler
	closeFunc                context.CancelFunc
}

var log = logger.GetOrCreate("factory")

// NewNetworkComponentsFactory returns a new instance of a network components factory
func NewNetworkComponentsFactory(
	args NetworkComponentsFactoryArgs,
) (*networkComponentsFactory, error) {
	if check.IfNil(args.StatusHandler) {
		return nil, errors.ErrNilStatusHandler
	}
	if check.IfNil(args.Marshalizer) {
		return nil, fmt.Errorf("%w in NewNetworkComponentsFactory", errors.ErrNilMarshalizer)
	}
	if check.IfNil(args.Syncer) {
		return nil, errors.ErrNilSyncTimer
	}
	if check.IfNil(args.CryptoComponents) {
		return nil, errors.ErrNilCryptoComponentsHolder
	}

	err := checkNodeOperationModes(args.NodeOperationModes)
	if err != nil {
		return nil, err
	}

	return &networkComponentsFactory{
		mainP2PConfig:         args.MainP2pConfig,
		fullArchiveP2PConfig:  args.FullArchiveP2pConfig,
		lightClientP2PConfig:  args.LightClientP2pConfig,
		ratingsConfig:         args.RatingsConfig,
		marshalizer:           args.Marshalizer,
		mainConfig:            args.MainConfig,
		statusHandler:         args.StatusHandler,
		syncer:                args.Syncer,
		bootstrapWaitTime:     args.BootstrapWaitTime,
		preferredPeersSlices:  args.PreferredPeersSlices,
		nodeOperationModes:    args.NodeOperationModes,
		connectionWatcherType: args.ConnectionWatcherType,
		cryptoComponents:      args.CryptoComponents,
	}, nil
}

// Create creates and returns the network components
func (ncf *networkComponentsFactory) Create() (*networkComponents, error) {
	peersRatingHandler, peersRatingMonitor, err := ncf.createPeersRatingComponents()
	if err != nil {
		return nil, err
	}

	mainNetworkComp, err := ncf.createMainNetworkHolder(peersRatingHandler)
	if err != nil {
		return nil, fmt.Errorf("%w for the main network holder", err)
	}

	fullArchiveNetworkComp, err := ncf.createFullArchiveNetworkHolder(peersRatingHandler)
	if err != nil {
		return nil, fmt.Errorf("%w for the full archive network holder", err)
	}

	lightClientNetworkComp, err := ncf.createLightClientNetworkHolder(peersRatingHandler)
	if err != nil {
		return nil, fmt.Errorf("%w for the light client network holder", err)
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer func() {
		if err != nil {
			cancelFunc()
		}
	}()

	antiFloodComponents, inputAntifloodHandler, outputAntifloodHandler, peerHonestyHandler, err := ncf.createAntifloodComponents(ctx, mainNetworkComp.netMessenger.ID())
	if err != nil {
		return nil, err
	}

	err = mainNetworkComp.netMessenger.Bootstrap()
	if err != nil {
		return nil, err
	}

	mainNetworkComp.netMessenger.WaitForConnections(ncf.bootstrapWaitTime, ncf.mainP2PConfig.Node.MinNumPeersToWaitForOnBootstrap)

	err = fullArchiveNetworkComp.netMessenger.Bootstrap()
	if err != nil {
		return nil, err
	}

	err = lightClientNetworkComp.netMessenger.Bootstrap()
	if err != nil {
		return nil, err
	}

	return &networkComponents{
		mainNetworkHolder:        mainNetworkComp,
		fullArchiveNetworkHolder: fullArchiveNetworkComp,
		lightClientNetworkHolder: lightClientNetworkComp,
		peersRatingHandler:       peersRatingHandler,
		peersRatingMonitor:       peersRatingMonitor,
		inputAntifloodHandler:    inputAntifloodHandler,
		outputAntifloodHandler:   outputAntifloodHandler,
		pubKeyTimeCacher:         antiFloodComponents.PubKeysCacher,
		topicFloodPreventer:      antiFloodComponents.TopicPreventer,
		floodPreventers:          antiFloodComponents.FloodPreventers,
		peerBlackListHandler:     antiFloodComponents.BlacklistHandler,
		antifloodConfig:          ncf.mainConfig.Antiflood,
		peerHonestyHandler:       peerHonestyHandler,
		closeFunc:                cancelFunc,
	}, nil
}

func (ncf *networkComponentsFactory) createAntifloodComponents(
	ctx context.Context,
	currentPid core.PeerID,
) (*antifloodFactory.AntiFloodComponents, factory.P2PAntifloodHandler, factory.P2PAntifloodHandler, consensus.PeerHonestyHandler, error) {
	var antiFloodComponents *antifloodFactory.AntiFloodComponents
	antiFloodComponents, err := antifloodFactory.NewP2PAntiFloodComponents(ctx, ncf.mainConfig, ncf.statusHandler, currentPid)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	inputAntifloodHandler, ok := antiFloodComponents.AntiFloodHandler.(factory.P2PAntifloodHandler)
	if !ok {
		err = errors.ErrWrongTypeAssertion
		return nil, nil, nil, nil, fmt.Errorf("%w when casting input antiflood handler to P2PAntifloodHandler", err)
	}

	var outAntifloodHandler process.P2PAntifloodHandler
	outAntifloodHandler, err = antifloodFactory.NewP2POutputAntiFlood(ctx, ncf.mainConfig)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	outputAntifloodHandler, ok := outAntifloodHandler.(factory.P2PAntifloodHandler)
	if !ok {
		err = errors.ErrWrongTypeAssertion
		return nil, nil, nil, nil, fmt.Errorf("%w when casting output antiflood handler to P2PAntifloodHandler", err)
	}

	var peerHonestyHandler consensus.PeerHonestyHandler
	peerHonestyHandler, err = ncf.createPeerHonestyHandler(
		&ncf.mainConfig,
		ncf.ratingsConfig,
		antiFloodComponents.PubKeysCacher,
	)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	return antiFloodComponents, inputAntifloodHandler, outputAntifloodHandler, peerHonestyHandler, nil
}

func (ncf *networkComponentsFactory) createPeerHonestyHandler(
	config *config.Config,
	ratingConfig config.RatingsConfig,
	pkTimeCache process.TimeCacher,
) (consensus.PeerHonestyHandler, error) {

	suCache, err := storageunit.NewCache(storageFactory.GetCacherFromConfig(config.PeerHonesty))
	if err != nil {
		return nil, err
	}

	return peerHonesty.NewP2pPeerHonesty(ratingConfig.PeerHonesty, pkTimeCache, suCache)
}

func (ncf *networkComponentsFactory) createNetworkHolder(
	p2pConfig p2pConfig.P2PConfig,
	logger p2p.Logger,
	peersRatingHandler p2p.PeersRatingHandler,
	networkType p2p.NetworkType,
) (networkComponentsHolder, error) {

	peersHolder, err := p2pFactory.NewPeersHolder(ncf.preferredPeersSlices)
	if err != nil {
		return networkComponentsHolder{}, err
	}

	argsMessenger := p2pFactory.ArgsNetworkMessenger{
		Marshaller:            ncf.marshalizer,
		P2pConfig:             p2pConfig,
		SyncTimer:             ncf.syncer,
		PreferredPeersHolder:  peersHolder,
		PeersRatingHandler:    peersRatingHandler,
		ConnectionWatcherType: ncf.connectionWatcherType,
		P2pPrivateKey:         ncf.cryptoComponents.P2pPrivateKey(),
		P2pSingleSigner:       ncf.cryptoComponents.P2pSingleSigner(),
		P2pKeyGenerator:       ncf.cryptoComponents.P2pKeyGen(),
		NetworkType:           networkType,
		Logger:                logger,
	}
	networkMessenger, err := p2pFactory.NewNetworkMessenger(argsMessenger)
	if err != nil {
		return networkComponentsHolder{}, err
	}

	err = networkMessenger.SetDebugger(p2pDebug.NewP2PDebugger(networkMessenger.ID()))
	if err != nil {
		return networkComponentsHolder{}, err
	}

	return networkComponentsHolder{
		netMessenger:         networkMessenger,
		preferredPeersHolder: peersHolder,
	}, nil
}

func (ncf *networkComponentsFactory) createMainNetworkHolder(peersRatingHandler p2p.PeersRatingHandler) (networkComponentsHolder, error) {
	loggerInstance := logger.GetOrCreate("main/p2p")
	return ncf.createNetworkHolder(ncf.mainP2PConfig, loggerInstance, peersRatingHandler, p2p.MainNetwork)
}

func (ncf *networkComponentsFactory) createFullArchiveNetworkHolder(peersRatingHandler p2p.PeersRatingHandler) (networkComponentsHolder, error) {
	if !common.Contains(ncf.nodeOperationModes, common.FullArchiveMode) {
		return networkComponentsHolder{
			netMessenger:         p2pDisabled.NewNetworkMessenger(),
			preferredPeersHolder: disabled.NewPreferredPeersHolder(),
		}, nil
	}

	loggerInstance := logger.GetOrCreate("full-archive/p2p")

	return ncf.createNetworkHolder(ncf.fullArchiveP2PConfig, loggerInstance, peersRatingHandler, p2p.FullArchiveNetwork)
}

func (ncf *networkComponentsFactory) createLightClientNetworkHolder(peersRatingHandler p2p.PeersRatingHandler) (networkComponentsHolder, error) {
	if !common.Contains(ncf.nodeOperationModes, common.LightClientMode) &&
		!common.Contains(ncf.nodeOperationModes, common.LightClientSupplierMode) {
		return networkComponentsHolder{
			netMessenger:         p2pDisabled.NewNetworkMessenger(),
			preferredPeersHolder: disabled.NewPreferredPeersHolder(),
		}, nil
	}

	loggerInstance := logger.GetOrCreate("light-client/p2p")

	return ncf.createNetworkHolder(ncf.lightClientP2PConfig, loggerInstance, peersRatingHandler, p2p.LightClientNetwork)
}

func (ncf *networkComponentsFactory) createPeersRatingComponents() (p2p.PeersRatingHandler, p2p.PeersRatingMonitor, error) {
	peersRatingCfg := ncf.mainConfig.PeersRatingConfig
	topRatedCache, err := cache.NewLRUCache(peersRatingCfg.TopRatedCacheCapacity)
	if err != nil {
		return nil, nil, err
	}
	badRatedCache, err := cache.NewLRUCache(peersRatingCfg.BadRatedCacheCapacity)
	if err != nil {
		return nil, nil, err
	}

	peersRatingLogger := logger.GetOrCreate("peersRating")
	argsPeersRatingHandler := p2pFactory.ArgPeersRatingHandler{
		TopRatedCache: topRatedCache,
		BadRatedCache: badRatedCache,
		Logger:        peersRatingLogger,
	}
	peersRatingHandler, err := p2pFactory.NewPeersRatingHandler(argsPeersRatingHandler)
	if err != nil {
		return nil, nil, err
	}

	argsPeersRatingMonitor := p2pFactory.ArgPeersRatingMonitor{
		TopRatedCache: topRatedCache,
		BadRatedCache: badRatedCache,
	}
	peersRatingMonitor, err := p2pFactory.NewPeersRatingMonitor(argsPeersRatingMonitor)
	if err != nil {
		return nil, nil, err
	}

	return peersRatingHandler, peersRatingMonitor, nil
}

// Close closes all underlying components that need closing
func (nc *networkComponents) Close() error {
	nc.closeFunc()

	if !check.IfNil(nc.inputAntifloodHandler) {
		log.LogIfError(nc.inputAntifloodHandler.Close())
	}
	if !check.IfNil(nc.outputAntifloodHandler) {
		log.LogIfError(nc.outputAntifloodHandler.Close())
	}
	if !check.IfNil(nc.peerHonestyHandler) {
		log.LogIfError(nc.peerHonestyHandler.Close())
	}

	mainNetMessenger := nc.mainNetworkHolder.netMessenger
	if !check.IfNil(mainNetMessenger) {
		log.Debug("calling close on the main network messenger instance...")
		log.LogIfError(mainNetMessenger.Close())
	}

	fullArchiveNetMessenger := nc.fullArchiveNetworkHolder.netMessenger
	if !check.IfNil(fullArchiveNetMessenger) {
		log.Debug("calling close on the full archive network messenger instance...")
		log.LogIfError(fullArchiveNetMessenger.Close())
	}

	lightClientMessenger := nc.lightClientNetworkHolder.netMessenger
	if !check.IfNil(lightClientMessenger) {
		log.Debug("calling close on the light client network messenger instance...")
		log.LogIfError(lightClientMessenger.Close())
	}

	return nil
}

func checkNodeOperationModes(nodeOperationModes []common.NodeOperation) error {
	// check if there are more than 2 simultaneous operating modes
	if len(nodeOperationModes) > 2 {
		return fmt.Errorf("cannot have more than 2 node operation modes, got %d modes instead",
			len(nodeOperationModes))
	}

	// if the node modes doesn't contain any of the valid ones, then the configuration is invalid
	if !common.Contains(nodeOperationModes, []common.NodeOperation{
		common.NormalOperation,
		common.FullArchiveMode,
		common.LightClientMode,
		common.LightClientSupplierMode}...) {
		return errors.ErrInvalidOperationMode
	}

	// the node must contain at least one of the following: common.NormalOperation or common.FullArchive
	if !common.Contains(nodeOperationModes, common.NormalOperation) &&
		!common.Contains(nodeOperationModes, common.FullArchiveMode) {
		return errors.ErrInvalidMainNodeOperationMode
	}

	// the node cannot be in both normal and full archive modes
	if common.Contains(nodeOperationModes, common.NormalOperation) &&
		common.Contains(nodeOperationModes, common.FullArchiveMode) {
		return errors.ErrInvalidNodeOperationModeCombo
	}

	// the node cannot be in both light client & light client supplier mode simultaneously
	if common.Contains(nodeOperationModes, common.LightClientMode) &&
		common.Contains(nodeOperationModes, common.LightClientSupplierMode) {
		return errors.ErrInvalidNodeOperationModeCombo
	}

	return nil
}
