package sovereign

import (
	"context"

	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
)

type SubRoundBlockHandler interface {
	bls.SubRoundHandler
	SetBlockJob(doBlockJob func(ctx context.Context) bool)
	DoBlockComputation() (*bls.SubRoundBlockProcessArgs, func())
	ProcessReceivedBlock(ctx context.Context, cnsDta *consensus.Message) bool
}

type subroundBlockV2 struct {
	SubRoundBlockHandler
}

// NewSubroundBlockV2 creates a subroundBlockV2 object
func NewSubroundBlockV2(subroundBlock SubRoundBlockHandler) (*subroundBlockV2, error) {
	if subroundBlock == nil {
		return nil, spos.ErrNilSubround
	}

	sr := &subroundBlockV2{
		subroundBlock,
	}

	sr.SetBlockJob(sr.doBlockJob)

	return sr, nil
}

// doBlockJob method does the job of the subround Block
func (sr *subroundBlockV2) doBlockJob(ctx context.Context) bool {
	args, deferFunc := sr.DoBlockComputation()
	defer deferFunc()

	if args == nil {
		return false
	}

	cnsDta := &consensus.Message{
		PubKey:     []byte(args.Leader),
		RoundIndex: sr.RoundHandler().Index(),
	}

	return sr.ProcessReceivedBlock(ctx, cnsDta)
}
