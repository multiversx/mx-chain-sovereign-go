package sovereign

import (
	"context"

	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
)

type subroundBlockV2 struct {
	bls.SubRoundBlockHandler
}

// NewSubroundBlockV2 creates a subroundBlockV2 object
func NewSubroundBlockV2(subroundBlock bls.SubRoundBlockHandler) (*subroundBlockV2, error) {
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
	args, deferFunc := sr.DoBlockComputation(ctx)
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
