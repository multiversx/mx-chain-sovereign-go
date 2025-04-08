package extendedHeader

import (
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/sovereign/dto"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
)

func TestNewEmptyMVXShardExtendedCreator(t *testing.T) {
	t.Parallel()

	t.Run("nil marshaller", func(t *testing.T) {
		creator, err := NewEmptyMVXShardExtendedCreator(nil)
		require.Nil(t, creator)
		require.Equal(t, data.ErrNilMarshalizer, err)
	})

	t.Run("should work", func(t *testing.T) {
		creator, err := NewEmptyMVXShardExtendedCreator(&testscommon.MarshallerStub{})
		require.Nil(t, err)
		require.False(t, creator.IsInterfaceNil())
	})
}

func TestEmptyMVXShardExtendedCreator_CreateNewExtendedHeader(t *testing.T) {
	t.Parallel()

	headerV2 := &block.HeaderV2{
		Header: &block.Header{
			Round: 11,
		},
	}

	marshaller := &marshallerMock.MarshalizerMock{}
	headerBytesProof, err := marshaller.Marshal(headerV2)
	require.Nil(t, err)

	creator, _ := NewEmptyMVXShardExtendedCreator(marshaller)
	extendedHeader, err := creator.CreateNewExtendedHeader(headerBytesProof)
	require.Nil(t, err)

	require.Equal(t, &block.ShardHeaderExtended{
		Header:        headerV2,
		Proof:         headerBytesProof,
		NonceBI:       big.NewInt(11),
		SourceChainID: dto.MVX,
	}, extendedHeader)
}

func TestEmptyMVXShardExtendedCreator_CreateNewExtendedHeaderErrorCases(t *testing.T) {
	t.Parallel()

	t.Run("invalid proof bytes", func(t *testing.T) {
		creator, _ := NewEmptyMVXShardExtendedCreator(&marshallerMock.MarshalizerMock{})

		extendedHeader, err := creator.CreateNewExtendedHeader([]byte("invalid proof"))
		require.Nil(t, extendedHeader)
		require.NotNil(t, err)
	})

	t.Run("invalid header type", func(t *testing.T) {
		marshaller := &marshallerMock.MarshalizerMock{}
		creator, _ := NewEmptyMVXShardExtendedCreator(marshaller)
		creator.headerV2BlockCreator = block.NewEmptyMetaBlockCreator()

		metaHeader := &block.MetaBlock{}
		headerBytesProof, err := marshaller.Marshal(metaHeader)
		require.Nil(t, err)

		extendedHeader, err := creator.CreateNewExtendedHeader(headerBytesProof)
		require.Nil(t, extendedHeader)
		require.ErrorIs(t, err, errors.ErrWrongTypeAssertion)
	})

}
