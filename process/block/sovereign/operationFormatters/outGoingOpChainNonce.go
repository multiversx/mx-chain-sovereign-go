package operationFormatters

import (
	"fmt"
	"sync"

	dtoCore "github.com/multiversx/mx-chain-core-go/data/sovereign/dto"
)

type outGoingOpChainNonce struct {
	mu   sync.Mutex
	data map[dtoCore.ChainID]uint64
}

// NewOutGoingOpChainNonce creates a chain nonce handler for outgoing operations
func NewOutGoingOpChainNonce() *outGoingOpChainNonce {
	return &outGoingOpChainNonce{
		// TODO: Marius C, MX-17260 get order chain ids here
		data: map[dtoCore.ChainID]uint64{
			dtoCore.MVX: 0,
		},
	}
}

// GetAndIncrementNonce will return the current nonce for the specified chain and increment it internally
func (op *outGoingOpChainNonce) GetAndIncrementNonce(chainID dtoCore.ChainID) (uint64, error) {
	op.mu.Lock()
	defer op.mu.Unlock()

	nonce, found := op.data[chainID]
	if !found {
		return 0, fmt.Errorf("%w for chain: %s in outGoingOpChainNonce.GetAndIncrementNonce",
			errChainIDNotSupported, chainID.String())
	}
	op.data[chainID]++

	return nonce, nil
}

// IsInterfaceNil checks if the underlying pointer is nil
func (op *outGoingOpChainNonce) IsInterfaceNil() bool {
	return op == nil
}
