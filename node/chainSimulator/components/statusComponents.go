package components

import (
	"context"
	"fmt"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/appStatusPolling"
	outportCfg "github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/statistics"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/integrationTests/mock"
	"github.com/multiversx/mx-chain-go/outport"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon"
)

type statusComponentsHolder struct {
	closeHandler             *closeHandler
	outportHandler           outport.OutportHandler
	softwareVersionChecker   statistics.SoftwareVersionChecker
	managedPeerMonitor       common.ManagedPeersMonitor
	appStatusHandler         core.AppStatusHandler
	forkDetector             process.ForkDetector
	statusPollingIntervalSec int
	cancelFunc               func()
}

// CreateStatusComponents will create a new instance of status components holder
func CreateStatusComponents(shardID uint32, appStatusHandler core.AppStatusHandler, statusPollingIntervalSec int) (factory.StatusComponentsHandler, error) {
	var err error
	instance := &statusComponentsHolder{
		closeHandler:             NewCloseHandler(),
		appStatusHandler:         appStatusHandler,
		statusPollingIntervalSec: statusPollingIntervalSec,
	}

	// TODO add drivers to index data
	instance.outportHandler, err = outport.NewOutport(100*time.Millisecond, outportCfg.OutportConfig{
		ShardID: shardID,
	})
	if err != nil {
		return nil, err
	}
	instance.softwareVersionChecker = &mock.SoftwareVersionCheckerMock{}
	instance.managedPeerMonitor = &testscommon.ManagedPeersMonitorStub{}

	instance.collectClosableComponents()

	return instance, nil
}

// OutportHandler will return the outport handler
func (s *statusComponentsHolder) OutportHandler() outport.OutportHandler {
	return s.outportHandler
}

// SoftwareVersionChecker will return the software version checker
func (s *statusComponentsHolder) SoftwareVersionChecker() statistics.SoftwareVersionChecker {
	return s.softwareVersionChecker
}

// ManagedPeersMonitor will return the managed peers monitor
func (s *statusComponentsHolder) ManagedPeersMonitor() common.ManagedPeersMonitor {
	return s.managedPeerMonitor
}

func (s *statusComponentsHolder) collectClosableComponents() {
	s.closeHandler.AddComponent(s.outportHandler)
	s.closeHandler.AddComponent(s.softwareVersionChecker)
}

// Close will call the Close methods on all inner components
func (s *statusComponentsHolder) Close() error {
	return s.closeHandler.Close()
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *statusComponentsHolder) IsInterfaceNil() bool {
	return s == nil
}

// Create will do nothing
func (s *statusComponentsHolder) Create() error {
	return nil
}

// CheckSubcomponents will do nothing
func (s *statusComponentsHolder) CheckSubcomponents() error {
	return nil
}

// String will do nothing
func (s *statusComponentsHolder) String() string {
	return ""
}

// SetForkDetector will do nothing
func (s *statusComponentsHolder) SetForkDetector(forkDetector process.ForkDetector) error {
	s.forkDetector = forkDetector

	return nil
}

// StartPolling starts polling for the updated status
func (s *statusComponentsHolder) StartPolling() error {
	var ctx context.Context
	ctx, s.cancelFunc = context.WithCancel(context.Background())

	appStatusPollingHandler, err := appStatusPolling.NewAppStatusPolling(
		s.appStatusHandler,
		time.Duration(s.statusPollingIntervalSec)*time.Second,
		log,
	)
	if err != nil {
		return errors.ErrStatusPollingInit
	}

	err = registerPollProbableHighestNonce(appStatusPollingHandler, s.forkDetector)
	if err != nil {
		return err
	}

	appStatusPollingHandler.Poll(ctx)

	return nil
}

func registerPollProbableHighestNonce(
	appStatusPollingHandler *appStatusPolling.AppStatusPolling,
	forkDetector process.ForkDetector,
) error {

	probableHighestNonceHandlerFunc := func(appStatusHandler core.AppStatusHandler) {
		probableHigherNonce := forkDetector.ProbableHighestNonce()
		appStatusHandler.SetUInt64Value(common.MetricProbableHighestNonce, probableHigherNonce)
	}

	err := appStatusPollingHandler.RegisterPollingFunc(probableHighestNonceHandlerFunc)
	if err != nil {
		return fmt.Errorf("%w, cannot register handler func for forkdetector's probable higher nonce", err)
	}

	return nil
}
