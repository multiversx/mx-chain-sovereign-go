package extendedHeader

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/sovereign/dto"
	"github.com/stretchr/testify/require"
)

func TestEthExtendedHeaderCreator_CreateNewExtendedHeader(t *testing.T) {
	t.Parallel()

	ethHdrCreator := NewETHExtendedHeaderCreator()
	require.False(t, ethHdrCreator.IsInterfaceNil())

	ethHdr := &types.Header{
		Number:     big.NewInt(4),
		Difficulty: big.NewInt(2),
	}
	ethProof, _ := json.Marshal(ethHdr)
	t.Run("invalid input", func(t *testing.T) {
		extendedHdr, err := ethHdrCreator.CreateNewExtendedHeader(nil)
		require.Nil(t, extendedHdr)
		require.Error(t, err)
	})

	t.Run("should create new extended header", func(t *testing.T) {
		extendedHdr, err := ethHdrCreator.CreateNewExtendedHeader(ethProof)
		require.Nil(t, err)
		require.Equal(t, &block.ShardHeaderExtended{
			Header: &block.HeaderV2{
				Header: &block.Header{
					Nonce:   ethHdr.Number.Uint64(),
					Round:   ethHdr.Number.Uint64(),
					ShardID: uint32(dto.ETH),
				},
			},
			Proof:         ethProof,
			SourceChainID: dto.ETH,
		}, extendedHdr)
	})

}
