package api

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-sovereign-go/facade"
)

func TestNewNoApiInterface(t *testing.T) {
	t.Parallel()

	instance := NewNoApiInterface()
	require.NotNil(t, instance)

	interf := instance.RestApiInterface(0)
	require.Equal(t, facade.DefaultRestPortOff, interf)
}
