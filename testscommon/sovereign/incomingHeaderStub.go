package sovereign

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
)

type IncomingHeaderStub struct {
	GetProofCalled   func() []byte
	GetNonceCalled   func() *big.Int
	GetChainIDCalled func() sovereign.ChainID
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

func (ihs *IncomingHeaderStub) GetNonce() *big.Int {
	if ihs.GetNonceCalled != nil {
		return ihs.GetNonceCalled()
	}

	return big.NewInt(0)
}

func (ihs *IncomingHeaderStub) GetChainID() sovereign.ChainID {
	if ihs.GetChainIDCalled != nil {
		return ihs.GetChainIDCalled()
	}

	return sovereign.MVX
}

func (ihs *IncomingHeaderStub) IsInterfaceNil() bool {
	return ihs == nil
}
