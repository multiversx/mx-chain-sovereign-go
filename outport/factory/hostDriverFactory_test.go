package factory

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-communication-go/websocket/data"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-sovereign-go/config"
	"github.com/multiversx/mx-chain-sovereign-go/testscommon/marshallerMock"
)

func TestCreateHostDriver(t *testing.T) {
	t.Parallel()

	args := ArgsHostDriverFactory{
		HostConfig: config.HostDriversConfig{
			URL:                "ws://localhost",
			RetryDurationInSec: 1,
			MarshallerType:     "json",
			Mode:               data.ModeClient,
		},
		Marshaller: &marshallerMock.MarshalizerStub{},
	}

	driver, err := CreateHostDriver(args)
	require.Nil(t, err)
	require.NotNil(t, driver)
	require.Equal(t, "*host.hostDriver", fmt.Sprintf("%T", driver))
}
