package factory

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	"github.com/multiversx/mx-chain-core-go/data/typeConverters"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"

	"github.com/multiversx/mx-chain-sovereign-go/common"
	"github.com/multiversx/mx-chain-sovereign-go/consensus"
	"github.com/multiversx/mx-chain-sovereign-go/factory"
	"github.com/multiversx/mx-chain-sovereign-go/ntp"
	"github.com/multiversx/mx-chain-sovereign-go/process"
	"github.com/multiversx/mx-chain-sovereign-go/sharding"
	"github.com/multiversx/mx-chain-sovereign-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-sovereign-go/storage"
)

// CoreComponentsHolderMock -
type CoreComponentsHolderMock struct {
	InternalMarshalizerCalled           func() marshal.Marshalizer
	SetInternalMarshalizerCalled        func(marshalizer marshal.Marshalizer) error
	TxMarshalizerCalled                 func() marshal.Marshalizer
	VmMarshalizerCalled                 func() marshal.Marshalizer
	HasherCalled                        func() hashing.Hasher
	TxSignHasherCalled                  func() hashing.Hasher
	Uint64ByteSliceConverterCalled      func() typeConverters.Uint64ByteSliceConverter
	AddressPubKeyConverterCalled        func() core.PubkeyConverter
	ValidatorPubKeyConverterCalled      func() core.PubkeyConverter
	PathHandlerCalled                   func() storage.PathManagerHandler
	WatchdogCalled                      func() core.WatchdogTimer
	AlarmSchedulerCalled                func() core.TimersScheduler
	SyncTimerCalled                     func() ntp.SyncTimer
	RoundHandlerCalled                  func() consensus.RoundHandler
	EconomicsDataCalled                 func() process.EconomicsDataHandler
	APIEconomicsDataCalled              func() process.EconomicsDataHandler
	RatingsDataCalled                   func() process.RatingsInfoHandler
	RaterCalled                         func() sharding.PeerAccountListAndRatingHandler
	GenesisNodesSetupCalled             func() sharding.GenesisNodesSetupHandler
	NodesShufflerCalled                 func() nodesCoordinator.NodesShuffler
	EpochNotifierCalled                 func() process.EpochNotifier
	EnableRoundsHandlerCalled           func() process.EnableRoundsHandler
	EpochStartNotifierWithConfirmCalled func() factory.EpochStartNotifierWithConfirm
	ChanStopNodeProcessCalled           func() chan endProcess.ArgEndProcess
	GenesisTimeCalled                   func() time.Time
	ChainIDCalled                       func() string
	MinTransactionVersionCalled         func() uint32
	TxVersionCheckerCalled              func() process.TxVersionCheckerHandler
	EncodedAddressLenCalled             func() uint32
	NodeTypeProviderCalled              func() core.NodeTypeProviderHandler
	WasmVMChangeLockerCalled            func() common.Locker
	ProcessStatusHandlerCalled          func() common.ProcessStatusHandler
	HardforkTriggerPubKeyCalled         func() []byte
	EnableEpochsHandlerCalled           func() common.EnableEpochsHandler
	RoundNotifierCalled                 func() process.RoundNotifier
}

// NewCoreComponentsHolderStubFromRealComponent -
func NewCoreComponentsHolderStubFromRealComponent(coreComponents factory.CoreComponentsHolder) *CoreComponentsHolderMock {
	return &CoreComponentsHolderMock{
		InternalMarshalizerCalled:           coreComponents.InternalMarshalizer,
		SetInternalMarshalizerCalled:        coreComponents.SetInternalMarshalizer,
		TxMarshalizerCalled:                 coreComponents.TxMarshalizer,
		VmMarshalizerCalled:                 coreComponents.VmMarshalizer,
		HasherCalled:                        coreComponents.Hasher,
		TxSignHasherCalled:                  coreComponents.TxSignHasher,
		Uint64ByteSliceConverterCalled:      coreComponents.Uint64ByteSliceConverter,
		AddressPubKeyConverterCalled:        coreComponents.AddressPubKeyConverter,
		ValidatorPubKeyConverterCalled:      coreComponents.ValidatorPubKeyConverter,
		PathHandlerCalled:                   coreComponents.PathHandler,
		WatchdogCalled:                      coreComponents.Watchdog,
		AlarmSchedulerCalled:                coreComponents.AlarmScheduler,
		SyncTimerCalled:                     coreComponents.SyncTimer,
		RoundHandlerCalled:                  coreComponents.RoundHandler,
		EconomicsDataCalled:                 coreComponents.EconomicsData,
		APIEconomicsDataCalled:              coreComponents.APIEconomicsData,
		RatingsDataCalled:                   coreComponents.RatingsData,
		RaterCalled:                         coreComponents.Rater,
		GenesisNodesSetupCalled:             coreComponents.GenesisNodesSetup,
		NodesShufflerCalled:                 coreComponents.NodesShuffler,
		EpochNotifierCalled:                 coreComponents.EpochNotifier,
		EnableRoundsHandlerCalled:           coreComponents.EnableRoundsHandler,
		EpochStartNotifierWithConfirmCalled: coreComponents.EpochStartNotifierWithConfirm,
		ChanStopNodeProcessCalled:           coreComponents.ChanStopNodeProcess,
		GenesisTimeCalled:                   coreComponents.GenesisTime,
		ChainIDCalled:                       coreComponents.ChainID,
		MinTransactionVersionCalled:         coreComponents.MinTransactionVersion,
		TxVersionCheckerCalled:              coreComponents.TxVersionChecker,
		EncodedAddressLenCalled:             coreComponents.EncodedAddressLen,
		NodeTypeProviderCalled:              coreComponents.NodeTypeProvider,
		WasmVMChangeLockerCalled:            coreComponents.WasmVMChangeLocker,
		ProcessStatusHandlerCalled:          coreComponents.ProcessStatusHandler,
		HardforkTriggerPubKeyCalled:         coreComponents.HardforkTriggerPubKey,
		EnableEpochsHandlerCalled:           coreComponents.EnableEpochsHandler,
		RoundNotifierCalled:                 coreComponents.RoundNotifier,
	}
}

// InternalMarshalizer -
func (stub *CoreComponentsHolderMock) InternalMarshalizer() marshal.Marshalizer {
	if stub.InternalMarshalizerCalled != nil {
		return stub.InternalMarshalizerCalled()
	}
	return nil
}

// SetInternalMarshalizer -
func (stub *CoreComponentsHolderMock) SetInternalMarshalizer(marshalizer marshal.Marshalizer) error {
	if stub.SetInternalMarshalizerCalled != nil {
		return stub.SetInternalMarshalizerCalled(marshalizer)
	}
	return nil
}

// TxMarshalizer -
func (stub *CoreComponentsHolderMock) TxMarshalizer() marshal.Marshalizer {
	if stub.TxMarshalizerCalled != nil {
		return stub.TxMarshalizerCalled()
	}
	return nil
}

// VmMarshalizer -
func (stub *CoreComponentsHolderMock) VmMarshalizer() marshal.Marshalizer {
	if stub.VmMarshalizerCalled != nil {
		return stub.VmMarshalizerCalled()
	}
	return nil
}

// Hasher -
func (stub *CoreComponentsHolderMock) Hasher() hashing.Hasher {
	if stub.HasherCalled != nil {
		return stub.HasherCalled()
	}
	return nil
}

// TxSignHasher -
func (stub *CoreComponentsHolderMock) TxSignHasher() hashing.Hasher {
	if stub.TxSignHasherCalled != nil {
		return stub.TxSignHasherCalled()
	}
	return nil
}

// Uint64ByteSliceConverter -
func (stub *CoreComponentsHolderMock) Uint64ByteSliceConverter() typeConverters.Uint64ByteSliceConverter {
	if stub.Uint64ByteSliceConverterCalled != nil {
		return stub.Uint64ByteSliceConverterCalled()
	}
	return nil
}

// AddressPubKeyConverter -
func (stub *CoreComponentsHolderMock) AddressPubKeyConverter() core.PubkeyConverter {
	if stub.AddressPubKeyConverterCalled != nil {
		return stub.AddressPubKeyConverterCalled()
	}
	return nil
}

// ValidatorPubKeyConverter -
func (stub *CoreComponentsHolderMock) ValidatorPubKeyConverter() core.PubkeyConverter {
	if stub.ValidatorPubKeyConverterCalled != nil {
		return stub.ValidatorPubKeyConverterCalled()
	}
	return nil
}

// PathHandler -
func (stub *CoreComponentsHolderMock) PathHandler() storage.PathManagerHandler {
	if stub.PathHandlerCalled != nil {
		return stub.PathHandlerCalled()
	}
	return nil
}

// Watchdog -
func (stub *CoreComponentsHolderMock) Watchdog() core.WatchdogTimer {
	if stub.WatchdogCalled != nil {
		return stub.WatchdogCalled()
	}
	return nil
}

// AlarmScheduler -
func (stub *CoreComponentsHolderMock) AlarmScheduler() core.TimersScheduler {
	if stub.AlarmSchedulerCalled != nil {
		return stub.AlarmSchedulerCalled()
	}
	return nil
}

// SyncTimer -
func (stub *CoreComponentsHolderMock) SyncTimer() ntp.SyncTimer {
	if stub.SyncTimerCalled != nil {
		return stub.SyncTimerCalled()
	}
	return nil
}

// RoundHandler -
func (stub *CoreComponentsHolderMock) RoundHandler() consensus.RoundHandler {
	if stub.RoundHandlerCalled != nil {
		return stub.RoundHandlerCalled()
	}
	return nil
}

// EconomicsData -
func (stub *CoreComponentsHolderMock) EconomicsData() process.EconomicsDataHandler {
	if stub.EconomicsDataCalled != nil {
		return stub.EconomicsDataCalled()
	}
	return nil
}

// APIEconomicsData -
func (stub *CoreComponentsHolderMock) APIEconomicsData() process.EconomicsDataHandler {
	if stub.APIEconomicsDataCalled != nil {
		return stub.APIEconomicsDataCalled()
	}
	return nil
}

// RatingsData -
func (stub *CoreComponentsHolderMock) RatingsData() process.RatingsInfoHandler {
	if stub.RatingsDataCalled != nil {
		return stub.RatingsDataCalled()
	}
	return nil
}

// Rater -
func (stub *CoreComponentsHolderMock) Rater() sharding.PeerAccountListAndRatingHandler {
	if stub.RaterCalled != nil {
		return stub.RaterCalled()
	}
	return nil
}

// GenesisNodesSetup -
func (stub *CoreComponentsHolderMock) GenesisNodesSetup() sharding.GenesisNodesSetupHandler {
	if stub.GenesisNodesSetupCalled != nil {
		return stub.GenesisNodesSetupCalled()
	}
	return nil
}

// NodesShuffler -
func (stub *CoreComponentsHolderMock) NodesShuffler() nodesCoordinator.NodesShuffler {
	if stub.NodesShufflerCalled != nil {
		return stub.NodesShufflerCalled()
	}
	return nil
}

// EpochNotifier -
func (stub *CoreComponentsHolderMock) EpochNotifier() process.EpochNotifier {
	if stub.EpochNotifierCalled != nil {
		return stub.EpochNotifierCalled()
	}
	return nil
}

// EnableRoundsHandler -
func (stub *CoreComponentsHolderMock) EnableRoundsHandler() process.EnableRoundsHandler {
	if stub.EnableRoundsHandlerCalled != nil {
		return stub.EnableRoundsHandlerCalled()
	}
	return nil
}

// EpochStartNotifierWithConfirm -
func (stub *CoreComponentsHolderMock) EpochStartNotifierWithConfirm() factory.EpochStartNotifierWithConfirm {
	if stub.EpochStartNotifierWithConfirmCalled != nil {
		return stub.EpochStartNotifierWithConfirmCalled()
	}
	return nil
}

// ChanStopNodeProcess -
func (stub *CoreComponentsHolderMock) ChanStopNodeProcess() chan endProcess.ArgEndProcess {
	if stub.ChanStopNodeProcessCalled != nil {
		return stub.ChanStopNodeProcessCalled()
	}
	return nil
}

// GenesisTime -
func (stub *CoreComponentsHolderMock) GenesisTime() time.Time {
	if stub.GenesisTimeCalled != nil {
		return stub.GenesisTimeCalled()
	}
	return time.Unix(0, 0)
}

// ChainID -
func (stub *CoreComponentsHolderMock) ChainID() string {
	if stub.ChainIDCalled != nil {
		return stub.ChainIDCalled()
	}
	return ""
}

// MinTransactionVersion -
func (stub *CoreComponentsHolderMock) MinTransactionVersion() uint32 {
	if stub.MinTransactionVersionCalled != nil {
		return stub.MinTransactionVersionCalled()
	}
	return 0
}

// TxVersionChecker -
func (stub *CoreComponentsHolderMock) TxVersionChecker() process.TxVersionCheckerHandler {
	if stub.TxVersionCheckerCalled != nil {
		return stub.TxVersionCheckerCalled()
	}
	return nil
}

// EncodedAddressLen -
func (stub *CoreComponentsHolderMock) EncodedAddressLen() uint32 {
	if stub.EncodedAddressLenCalled != nil {
		return stub.EncodedAddressLenCalled()
	}
	return 0
}

// NodeTypeProvider -
func (stub *CoreComponentsHolderMock) NodeTypeProvider() core.NodeTypeProviderHandler {
	if stub.NodeTypeProviderCalled != nil {
		return stub.NodeTypeProviderCalled()
	}
	return nil
}

// WasmVMChangeLocker -
func (stub *CoreComponentsHolderMock) WasmVMChangeLocker() common.Locker {
	if stub.WasmVMChangeLockerCalled != nil {
		return stub.WasmVMChangeLockerCalled()
	}
	return nil
}

// ProcessStatusHandler -
func (stub *CoreComponentsHolderMock) ProcessStatusHandler() common.ProcessStatusHandler {
	if stub.ProcessStatusHandlerCalled != nil {
		return stub.ProcessStatusHandlerCalled()
	}
	return nil
}

// HardforkTriggerPubKey -
func (stub *CoreComponentsHolderMock) HardforkTriggerPubKey() []byte {
	if stub.HardforkTriggerPubKeyCalled != nil {
		return stub.HardforkTriggerPubKeyCalled()
	}
	return nil
}

// EnableEpochsHandler -
func (stub *CoreComponentsHolderMock) EnableEpochsHandler() common.EnableEpochsHandler {
	if stub.EnableEpochsHandlerCalled != nil {
		return stub.EnableEpochsHandlerCalled()
	}
	return nil
}

// RoundNotifier -
func (stub *CoreComponentsHolderMock) RoundNotifier() process.RoundNotifier {
	if stub.RoundNotifierCalled != nil {
		return stub.RoundNotifierCalled()
	}
	return nil
}

// IsInterfaceNil -
func (stub *CoreComponentsHolderMock) IsInterfaceNil() bool {
	return stub == nil
}
