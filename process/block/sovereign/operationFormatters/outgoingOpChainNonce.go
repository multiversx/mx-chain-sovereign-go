package operationFormatters

import dtoCore "github.com/multiversx/mx-chain-core-go/data/sovereign/dto"

type outGoingOpChainNonce struct {
	data map[dtoCore.ChainID]uint64
}

func NewOutGoingOpChainNonce() *outGoingOpChainNonce {
	return &outGoingOpChainNonce{
		data: make(map[dtoCore.ChainID]uint64),
	}
}

func (op *outGoingOpChainNonce) GetAndIncrementNonce(chainID dtoCore.ChainID) (uint64, error) {
	nonce := op.data[chainID]
	op.data[chainID]++

	return nonce, nil
}

func (op *outGoingOpChainNonce) IncrementNonce(chainID dtoCore.ChainID, delta uint64) {
	op.data[chainID] += delta
	//return nil
}
