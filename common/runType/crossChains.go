package runType

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/data/sovereign/dto"
	"golang.org/x/exp/slices"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
)

// GetOrderedCrossChainIDs returns ordered chain ids from config
func GetOrderedCrossChainIDs(mainChainNotarizationStartRound map[string]config.MainChainNotarization) ([]dto.ChainID, error) {
	orderedChainIDs := make([]dto.ChainID, 0, len(mainChainNotarizationStartRound))
	for chainIDStr := range mainChainNotarizationStartRound {
		chainID, valid := dto.ChainID_value[chainIDStr]
		if !valid {
			return nil, fmt.Errorf("%w for chain:%s in GetOrderedCrossChainIDs", process.ErrInvalidChainID, chainIDStr)
		}

		orderedChainIDs = append(orderedChainIDs, dto.ChainID(chainID))
	}

	slices.SortStableFunc(orderedChainIDs, func(a, b dto.ChainID) bool {
		return a < b
	})

	return orderedChainIDs, nil
}
