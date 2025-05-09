package testscommon

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/sovereign/dto"
)

// ExtendedShardHeaderTrackerStub -
type ExtendedShardHeaderTrackerStub struct {
	BlockTrackerStub
	ComputeLongestExtendedShardChainsFromLastNotarizedCalled func(chainID []dto.ChainID) ([]data.HeaderHandler, [][]byte, map[uint32][]data.HeaderHandler, error)
	RemoveLastCrossNotarizedHeadersCalled                    func(chainID dto.ChainID)
	RemoveLastSelfNotarizedHeadersCalled                     func()
}

// ComputeLongestExtendedShardChainsFromLastNotarized -
func (eshts *ExtendedShardHeaderTrackerStub) ComputeLongestExtendedShardChainsFromLastNotarized(chainID []dto.ChainID) ([]data.HeaderHandler, [][]byte, map[uint32][]data.HeaderHandler, error) {
	if eshts.ComputeLongestExtendedShardChainsFromLastNotarizedCalled != nil {
		return eshts.ComputeLongestExtendedShardChainsFromLastNotarizedCalled(chainID)
	}
	return nil, nil, nil, nil
}

// RemoveLastCrossNotarizedHeader -
func (eshts *ExtendedShardHeaderTrackerStub) RemoveLastCrossNotarizedHeader(chainID dto.ChainID) {
	if eshts.RemoveLastCrossNotarizedHeadersCalled != nil {
		eshts.RemoveLastCrossNotarizedHeadersCalled(chainID)
	}
}

// RemoveLastSelfNotarizedHeaders -
func (eshts *ExtendedShardHeaderTrackerStub) RemoveLastSelfNotarizedHeaders() {
	if eshts.RemoveLastSelfNotarizedHeadersCalled != nil {
		eshts.RemoveLastSelfNotarizedHeadersCalled()
	}
}

// IsGenesisLastCrossNotarizedHeader -
func (eshts *ExtendedShardHeaderTrackerStub) IsGenesisLastCrossNotarizedHeader(_ dto.ChainID) bool {
	return false
}
