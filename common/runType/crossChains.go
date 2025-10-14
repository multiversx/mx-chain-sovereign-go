package runType

import (
	"fmt"
	"slices"

	"github.com/multiversx/mx-chain-core-go/data/sovereign/dto"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
)

// GetOrderedCrossChainIDs returns ordered chain ids from config
func GetOrderedCrossChainIDs(mainChainNotarizationStartRound map[string]config.MainChainNotarization) ([]dto.ChainID, error) {
	orderedChainIDs := make([]dto.ChainID, 0, len(mainChainNotarizationStartRound))
	for chainIDStr := range mainChainNotarizationStartRound {
		if !dto.IsValidCrossChainIDString(chainIDStr) && chainIDStr != "SUI" {
			return nil, fmt.Errorf("%w for chain:%s in GetOrderedCrossChainIDs", process.ErrInvalidChainID, chainIDStr)
		}

		orderedChainIDs = append(orderedChainIDs, dto.ChainID(dto.ChainID_value[chainIDStr]))
	}

	slices.Sort(orderedChainIDs)

	return orderedChainIDs, nil
}
