package sovereign

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
)

type subroundSignatureV2 struct {
	bls.SubRoundSignatureHandler
}

// NewSubroundSignatureV2 creates a subroundSignatureV2 object
func NewSubroundSignatureV2(subroundSignature bls.SubRoundSignatureHandler) (*subroundSignatureV2, error) {
	if subroundSignature == nil {
		return nil, spos.ErrNilSubround
	}

	sr := &subroundSignatureV2{
		subroundSignature,
	}

	sr.SetMessageToSignFunc(sr.getMessageToSign)

	return sr, nil
}

func (sr *subroundSignatureV2) getMessageToSign() []byte {
	headerHash, err := core.CalculateHash(sr.Marshalizer(), sr.Hasher(), sr.GetHeader())
	if err != nil {
		log.Error("subroundSignatureV2.getMessageToSign", "error", err.Error())
		return nil
	}

	return headerHash
}
