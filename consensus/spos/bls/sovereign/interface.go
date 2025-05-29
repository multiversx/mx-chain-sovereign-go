package sovereign

import (
	"context"

	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
	"github.com/multiversx/mx-chain-go/outport"
)

type SubRoundStartHandler interface {
	bls.SubRoundHandler
	SetOutportHandler(outportHandler outport.OutportHandler) error
}

type SubRoundsFactoryHandler interface {
	GenerateStartRoundSubround() (SubRoundStartHandler, error)
	GenerateBlockSubround() (SubRoundBlockHandler, error)
	GenerateSignatureSubround() (SubRoundSignatureHandler, error)
	GenerateEndRoundSubround() (SubRoundEndHandler, error)
}

type SubRoundBlockHandler interface {
	bls.SubRoundHandler
	SetBlockJob(doBlockJob func(ctx context.Context) bool)
	DoBlockComputation() (*bls.SubRoundBlockProcessArgs, func())
	ProcessReceivedBlock(ctx context.Context, cnsDta *consensus.Message) bool
}

// SubRoundEndHandler defines a sub round end handler
type SubRoundEndHandler interface {
	bls.SubRoundHandler
	SetMessageToVerifySigFunc(verifyMsgFunc func() []byte)
	SetBlockJob(doBlockJob func(ctx context.Context) bool)
	ReceivedBlockHeaderFinalInfo(ctx context.Context, cnsDta *consensus.Message) bool
	DoEndRoundJob(ctx context.Context) bool
	IsSelfLeaderInCurrentRound() bool
}

type SubRoundSignatureHandler interface {
	bls.SubRoundHandler
	SetMessageToSignFunc(verifyMsgFunc func() []byte)
}
