package sovereign

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/sovereign/dto"
)

// IncomingHeaderStub -
type IncomingHeaderStub struct {
	GetProofCalled         func() []byte
	GetNonceBICalled       func() *big.Int
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

// GetNonceBI -
func (ihs *IncomingHeaderStub) GetNonceBI() *big.Int {
	if ihs.GetNonceBICalled != nil {
		return ihs.GetNonceBICalled()
	}

	return big.NewInt(0)
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
