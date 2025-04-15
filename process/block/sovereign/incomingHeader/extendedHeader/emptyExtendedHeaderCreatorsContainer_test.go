package extendedHeader

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/sovereign/dto"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestEmptyBlockCreatorsContainer_Add_Get(t *testing.T) {
	t.Parallel()

	container := NewEmptyBlockCreatorsContainer()
	require.False(t, container.IsInterfaceNil())

	err := container.Add(dto.MVX, nil)
	require.ErrorIs(t, err, errNilEmptyExtendedHeaderCreator)

	retrievedCreator, err := container.Get(dto.MVX)
	require.Nil(t, retrievedCreator)
	require.Equal(t, err, errChainIDNotFound)

	mvxChainExtendedHdrCreator, _ := NewEmptyMVXShardExtendedCreator(&testscommon.MarshallerMock{})
	err = container.Add(dto.MVX, mvxChainExtendedHdrCreator)
	require.Nil(t, err)

	retrievedCreator, err = container.Get(dto.MVX)
	require.Equal(t, mvxChainExtendedHdrCreator, retrievedCreator)
	require.Nil(t, err)
}
