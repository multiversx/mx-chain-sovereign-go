package sovereign

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/sovereign/dto"
)

type IncomingHeaderStub struct {
	GetProofCalled         func() []byte
	GetNonceBICalled       func() *big.Int
	GetSourceChainIDCalled func() dto.ChainID
}

func (ihs *IncomingHeaderStub) GetIncomingEventHandlers() []data.EventHandler {
	return nil
}

func (ihs *IncomingHeaderStub) GetProof() []byte {
	if ihs.GetProofCalled != nil {
		return ihs.GetProofCalled()
	}

	return nil
}

func (ihs *IncomingHeaderStub) GetNonceBI() *big.Int {
	if ihs.GetNonceBICalled != nil {
		return ihs.GetNonceBICalled()
	}

	return big.NewInt(0)
}

func (ihs *IncomingHeaderStub) GetSourceChainID() dto.ChainID {
	if ihs.GetSourceChainIDCalled != nil {
		return ihs.GetSourceChainIDCalled()
	}

	return dto.MVX
}

func (ihs *IncomingHeaderStub) IsInterfaceNil() bool {
	return ihs == nil
}
