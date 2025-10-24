package operationFormatters

import "github.com/multiversx/mx-chain-core-go/data/sovereign/dto"

func addDataToChains(data []byte, chains []dto.ChainID) map[dto.ChainID][]byte {
	ret := map[dto.ChainID][]byte{}
	for _, chain := range chains {
		ret[chain] = data
	}
	return ret
}
