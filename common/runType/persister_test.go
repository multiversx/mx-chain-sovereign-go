package runType

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreatePersister(t *testing.T) {
	SetShouldCreatePersisterForNextEpoch(false)
	require.False(t, ShouldCreatePersister())

	SetShouldCreatePersisterForNextEpoch(true)
	require.True(t, ShouldCreatePersister())
}
