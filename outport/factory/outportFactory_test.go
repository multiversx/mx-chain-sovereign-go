package factory_test

import (
	"errors"
	"testing"
	"time"

	indexerFactory "github.com/ElrondNetwork/elastic-indexer-go/factory"
	"github.com/ElrondNetwork/elrond-go/outport"
	"github.com/ElrondNetwork/elrond-go/outport/factory"
	notifierFactory "github.com/ElrondNetwork/elrond-go/outport/factory"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/testscommon/hashingMocks"
	"github.com/stretchr/testify/require"
)

func createMockArgsOutportHandler(indexerEnabled, notifierEnabled bool) *factory.OutportFactoryArgs {
	mockElasticArgs := &indexerFactory.ArgsIndexerFactory{
		Enabled: indexerEnabled,
	}
	mockNotifierArgs := &notifierFactory.EventNotifierFactoryArgs{
		Enabled: notifierEnabled,
	}

	return &factory.OutportFactoryArgs{
		RetrialInterval:           time.Second,
		ElasticIndexerFactoryArgs: mockElasticArgs,
		EventNotifierFactoryArgs:  mockNotifierArgs,
	}
}

func TestNewIndexerFactory(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		argsFunc func() *factory.OutportFactoryArgs
		exError  error
	}{
		{
			name: "NilArgsOutportFactory",
			argsFunc: func() *factory.OutportFactoryArgs {
				return nil
			},
			exError: outport.ErrNilArgsOutportFactory,
		},
		{
			name: "invalid retrial duration",
			argsFunc: func() *factory.OutportFactoryArgs {
				args := createMockArgsOutportHandler(false, false)
				args.RetrialInterval = 0
				return args
			},
			exError: outport.ErrInvalidRetrialInterval,
		},
		{
			name: "AllOkShouldWork",
			argsFunc: func() *factory.OutportFactoryArgs {
				return createMockArgsOutportHandler(false, false)
			},
			exError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := factory.CreateOutport(tt.argsFunc())
			require.True(t, errors.Is(err, tt.exError))
		})
	}
}

func TestCreateOutport_EnabledDriversNilMockArgsExpectErrorSubscribingDrivers(t *testing.T) {
	t.Parallel()
	tests := []struct {
		argsFunc func() *factory.OutportFactoryArgs
	}{
		{
			argsFunc: func() *factory.OutportFactoryArgs {
				return createMockArgsOutportHandler(true, false)
			},
		},
		{
			argsFunc: func() *factory.OutportFactoryArgs {
				return createMockArgsOutportHandler(false, true)
			},
		},
	}

	for _, currTest := range tests {
		_, err := factory.CreateOutport(currTest.argsFunc())
		require.NotNil(t, err)
	}
}

func TestCreateOutport_SubscribeNotifierDriver(t *testing.T) {
	args := createMockArgsOutportHandler(false, true)

	args.EventNotifierFactoryArgs.Marshaller = &mock.MarshalizerMock{}
	args.EventNotifierFactoryArgs.Hasher = &hashingMocks.HasherMock{}
	args.EventNotifierFactoryArgs.PubKeyConverter = &mock.PubkeyConverterMock{}
	outPort, err := factory.CreateOutport(args)

	defer func(c outport.OutportHandler) {
		_ = c.Close()
	}(outPort)

	require.Nil(t, err)
	require.True(t, outPort.HasDrivers())
}
