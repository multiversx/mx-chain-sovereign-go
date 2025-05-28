package bls

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/consensus/spos"
)

// SubRoundEndHandler defines a sub round end handler
type SubRoundEndHandler interface {
	SubRoundHandler
	SetMessageToVerifySigFunc(verifyMsgFunc func() []byte)
}

type subroundEndRoundV2 struct {
	SubRoundEndHandler
}

// NewSubroundEndRoundV2 creates a subroundEndRoundV2 object
func NewSubroundEndRoundV2(subroundEndRound SubRoundEndHandler) (*subroundEndRoundV2, error) {
	if subroundEndRound == nil {
		return nil, spos.ErrNilSubround
	}

	sr := &subroundEndRoundV2{
		subroundEndRound,
	}

	sr.SetMessageToVerifySigFunc(sr.getMessageToVerifySig)

	return sr, nil
}

func (sr *subroundEndRoundV2) getMessageToVerifySig() []byte {
	headerHash, err := core.CalculateHash(sr.Marshalizer(), sr.Hasher(), sr.GetHeader())
	if err != nil {
		log.Error("subroundEndRoundV2.getMessageToVerifySig", "error", err.Error())
		return nil
	}

	return headerHash
}

// IsInterfaceNil checks if the underlying interface is nil
func (sr *subroundEndRoundV2) IsInterfaceNil() bool {
	return sr == nil
}
