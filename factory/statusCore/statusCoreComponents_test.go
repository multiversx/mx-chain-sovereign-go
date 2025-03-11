package statusCore_test

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-sovereign-go/common/statistics"
	"github.com/multiversx/mx-chain-sovereign-go/config"
	errorsMx "github.com/multiversx/mx-chain-sovereign-go/errors"
	"github.com/multiversx/mx-chain-sovereign-go/factory/statusCore"
	"github.com/multiversx/mx-chain-sovereign-go/integrationTests/mock"
	"github.com/multiversx/mx-chain-sovereign-go/process"
	componentsMock "github.com/multiversx/mx-chain-sovereign-go/testscommon/components"
	"github.com/multiversx/mx-chain-sovereign-go/testscommon/economicsmocks"
	"github.com/multiversx/mx-chain-sovereign-go/testscommon/factory"
)

func TestNewStatusCoreComponentsFactory(t *testing.T) {
	t.Parallel()

	t.Run("nil core components should error", func(t *testing.T) {
		t.Parallel()

		args := componentsMock.GetStatusCoreArgs(nil)
		sccf, err := statusCore.NewStatusCoreComponentsFactory(args)
		assert.Equal(t, errorsMx.ErrNilCoreComponents, err)
		require.Nil(t, sccf)
	})
	t.Run("nil economics data should error", func(t *testing.T) {
		t.Parallel()

		coreComp := &mock.CoreComponentsStub{
			EconomicsDataField: nil,
		}

		args := componentsMock.GetStatusCoreArgs(coreComp)
		sccf, err := statusCore.NewStatusCoreComponentsFactory(args)
		assert.Equal(t, errorsMx.ErrNilEconomicsData, err)
		require.Nil(t, sccf)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := componentsMock.GetStatusCoreArgs(componentsMock.GetCoreComponents())
		sccf, err := statusCore.NewStatusCoreComponentsFactory(args)
		assert.Nil(t, err)
		require.NotNil(t, sccf)
	})
}

func TestStatusCoreComponentsFactory_Create(t *testing.T) {
	t.Parallel()

	t.Run("NewResourceMonitor fails should error", func(t *testing.T) {
		t.Parallel()

		args := componentsMock.GetStatusCoreArgs(componentsMock.GetCoreComponents())
		args.Config = config.Config{
			ResourceStats: config.ResourceStatsConfig{
				RefreshIntervalInSec: 0,
			},
		}
		sccf, err := statusCore.NewStatusCoreComponentsFactory(args)
		require.Nil(t, err)

		cc, err := sccf.Create()
		require.Nil(t, cc)
		require.True(t, errors.Is(err, statistics.ErrInvalidRefreshIntervalValue))
	})
	t.Run("NewPersistentStatusHandler fails should error", func(t *testing.T) {
		t.Parallel()

		coreCompStub := factory.NewCoreComponentsHolderStubFromRealComponent(componentsMock.GetCoreComponents())
		coreCompStub.InternalMarshalizerCalled = func() marshal.Marshalizer {
			return nil
		}
		args := componentsMock.GetStatusCoreArgs(coreCompStub)
		sccf, err := statusCore.NewStatusCoreComponentsFactory(args)
		require.Nil(t, err)

		cc, err := sccf.Create()
		require.Error(t, err)
		require.Nil(t, cc)
	})
	t.Run("SetStatusHandler fails should error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		coreCompStub := factory.NewCoreComponentsHolderStubFromRealComponent(componentsMock.GetCoreComponents())
		coreCompStub.EconomicsDataCalled = func() process.EconomicsDataHandler {
			return &economicsmocks.EconomicsHandlerStub{
				SetStatusHandlerCalled: func(statusHandler core.AppStatusHandler) error {
					return expectedErr
				},
			}
		}
		args := componentsMock.GetStatusCoreArgs(coreCompStub)
		sccf, err := statusCore.NewStatusCoreComponentsFactory(args)
		require.Nil(t, err)

		cc, err := sccf.Create()
		require.Equal(t, expectedErr, err)
		require.Nil(t, cc)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := componentsMock.GetStatusCoreArgs(componentsMock.GetCoreComponents())
		args.Config.ResourceStats.Enabled = true // coverage
		sccf, err := statusCore.NewStatusCoreComponentsFactory(args)
		require.Nil(t, err)

		cc, err := sccf.Create()
		require.NoError(t, err)
		require.NotNil(t, cc)
		require.NoError(t, cc.Close())
	})
}

// ------------ Test CoreComponents --------------------
func TestStatusCoreComponents_CloseShouldWork(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetStatusCoreArgs(componentsMock.GetCoreComponents())
	sccf, err := statusCore.NewStatusCoreComponentsFactory(args)
	require.Nil(t, err)
	cc, err := sccf.Create()
	require.NoError(t, err)

	err = cc.Close()
	require.NoError(t, err)
}
