package disabled

import dtoCore "github.com/multiversx/mx-chain-core-go/data/sovereign/dto"

// DisabledOutGoingChainNonce is a disabled outgoing chain nonce component
type DisabledOutGoingChainNonce struct {
}

// GetAndIncrementNonce does nothing
func (ocn *DisabledOutGoingChainNonce) GetAndIncrementNonce(_ dtoCore.ChainID) (uint64, error) {
	return 0, nil
}

// GetNonce does nothing
func (ocn *DisabledOutGoingChainNonce) GetNonce(_ dtoCore.ChainID) uint64 {
	return 0
}

// SetNonce does nothing
func (ocn *DisabledOutGoingChainNonce) SetNonce(_ dtoCore.ChainID, _ uint64) {
}

// IsInterfaceNil checks if the underlying pointer is nil
func (ocn *DisabledOutGoingChainNonce) IsInterfaceNil() bool {
	return ocn == nil
}
