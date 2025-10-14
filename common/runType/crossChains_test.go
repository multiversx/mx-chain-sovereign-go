package runType

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/sovereign/dto"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
)

func TestGetOrderedCrossChainIDs(t *testing.T) {
	cfg := map[string]config.MainChainNotarization{
		"INVALID_CHAIN":  {},
		dto.MVX.String(): {},
	}

	chainIDs, err := GetOrderedCrossChainIDs(cfg)
	require.ErrorIs(t, err, process.ErrInvalidChainID)
	require.Nil(t, chainIDs)

	cfg = map[string]config.MainChainNotarization{
		dto.MVX.String():         {},
		dto.UNSPECIFIED.String(): {},
	}
	chainIDs, err = GetOrderedCrossChainIDs(cfg)
	require.ErrorIs(t, err, process.ErrInvalidChainID)
	require.Nil(t, chainIDs)

	cfg = map[string]config.MainChainNotarization{
		dto.MVX.String(): {},
		dto.ETH.String(): {},
	}
	chainIDs, err = GetOrderedCrossChainIDs(cfg)
	require.Nil(t, err)
	require.Equal(t, []dto.ChainID{dto.MVX, dto.ETH}, chainIDs)
}
