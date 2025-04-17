package sovereign

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/sovereign/dto"
)

// IncomingHeaderStub -
type IncomingHeaderStub struct {
	NonceField             uint64
	GetProofCalled         func() []byte
	GetSourceChainIDCalled func() dto.ChainID
}

// GetIncomingEventHandlers -
func (ihs *IncomingHeaderStub) GetIncomingEventHandlers() []data.EventHandler {
	return nil
}

// GetProof -
func (ihs *IncomingHeaderStub) GetProof() []byte {
	if ihs.GetProofCalled != nil {
		return ihs.GetProofCalled()
	}

	return nil
}

// GetNonce -
func (ihs *IncomingHeaderStub) GetNonce() uint64 {
	return ihs.NonceField
}

// GetSourceChainID -
func (ihs *IncomingHeaderStub) GetSourceChainID() dto.ChainID {
	if ihs.GetSourceChainIDCalled != nil {
		return ihs.GetSourceChainIDCalled()
	}

	return dto.MVX
}

// IsInterfaceNil -
func (ihs *IncomingHeaderStub) IsInterfaceNil() bool {
	return ihs == nil
}
