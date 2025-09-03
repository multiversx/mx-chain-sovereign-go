package extraSigners

import (
	"bytes"
	"fmt"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/consensus"
)

func initExtraSignatureEntry(cnsMsg *consensus.Message, key string) {
	if cnsMsg.ExtraSignatures == nil {
		cnsMsg.ExtraSignatures = make(map[string]*consensus.ExtraSignatureData)
	}

	if _, found := cnsMsg.ExtraSignatures[key]; !found {
		cnsMsg.ExtraSignatures[key] = &consensus.ExtraSignatureData{}
	}
}

type getOutGoingMBData func(outGoingMB data.OutGoingMiniBlockHeaderHandler) []byte

var getLeaderSigData getOutGoingMBData = func(outGoingMB data.OutGoingMiniBlockHeaderHandler) []byte {
	return outGoingMB.GetLeaderSignatureOutGoingOperations()
}

var getAggSigData getOutGoingMBData = func(outGoingMB data.OutGoingMiniBlockHeaderHandler) []byte {
	return outGoingMB.GetAggregatedSignatureOutGoingOperations()
}

var getOpHashData getOutGoingMBData = func(outGoingMB data.OutGoingMiniBlockHeaderHandler) []byte {
	return outGoingMB.GetOutGoingOperationsHash()
}

func checkAllMBsHaveSameData(outGoingMBs []data.OutGoingMiniBlockHeaderHandler, getDataFuncs ...getOutGoingMBData) error {
	for _, getData := range getDataFuncs {
		first := getData(outGoingMBs[0])

		for idx := 1; idx < len(outGoingMBs); idx++ {
			current := getData(outGoingMBs[idx])
			if !bytes.Equal(first, current) {
				log.Error("checkAllMBsHaveSameData mismatch",
					"error", errDataMismatchOutGoingMB,
					"first", first,
					"current", current,
					"idx", idx,
					"data type", fmt.Sprintf("%T", getData),
				)
				return errDataMismatchOutGoingMB
			}
		}
	}

	return nil
}
