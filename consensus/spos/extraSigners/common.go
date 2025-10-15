package extraSigners

import (
	"bytes"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
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

type dataLabel struct {
	data []byte
	name string
}

type getOutGoingMBData func(outGoingMB data.OutGoingMiniBlockHeaderHandler) dataLabel

var getLeaderSigData getOutGoingMBData = func(outGoingMB data.OutGoingMiniBlockHeaderHandler) dataLabel {
	return dataLabel{
		data: outGoingMB.GetLeaderSignatureOutGoingOperations(),
		name: "getLeaderSigData",
	}
}

var getAggSigData getOutGoingMBData = func(outGoingMB data.OutGoingMiniBlockHeaderHandler) dataLabel {
	return dataLabel{
		data: outGoingMB.GetAggregatedSignatureOutGoingOperations(),
		name: "getAggSigData",
	}
}

var getOpHashData getOutGoingMBData = func(outGoingMB data.OutGoingMiniBlockHeaderHandler) dataLabel {
	return dataLabel{
		data: outGoingMB.GetOutGoingOperationsHash(),
		name: "getOpHashData",
	}
}

func checkAllMBsHaveSameData(outGoingMBs []data.OutGoingMiniBlockHeaderHandler, getDataFuncs ...getOutGoingMBData) error {
	for _, getData := range getDataFuncs {
		first := getData(outGoingMBs[0])

		for idx := 1; idx < len(outGoingMBs); idx++ {
			current := getData(outGoingMBs[idx])
			if !bytes.Equal(first.data, current.data) {
				log.Error("checkAllMBsHaveSameData mismatch",
					"error", errDataMismatchOutGoingMB,
					"mb type", block.OutGoingMBType(outGoingMBs[idx].GetOutGoingMBTypeInt32()).String(),
					"first data", first.data,
					"current data", current.data,
					"first chain id", outGoingMBs[0].GetChainID().String(),
					"current chain id", outGoingMBs[idx].GetChainID().String(),
					"data type", first.name,
				)
				return errDataMismatchOutGoingMB
			}
		}
	}

	return nil
}
