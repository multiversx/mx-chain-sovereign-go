package sovereign

import (
	"context"

	"github.com/multiversx/mx-chain-go/consensus"
)

// GetMessageToSign gets the message that should be signed
func (sr *subroundSignatureV2) GetMessageToSign() []byte {
	return sr.getMessageToSign()
}

// ReceivedBlockHeaderFinalInfo calls the unexported receivedBlockHeaderFinalInfo function
func (sr *sovereignSubRoundEnd) ReceivedBlockHeaderFinalInfo(cnsDta *consensus.Message) bool {
	return sr.receivedBlockHeaderFinalInfo(context.Background(), cnsDta)
}

// GetMessageToVerifySig gets the message on which the signature should be verified
func (sr *subroundEndRoundV2) GetMessageToVerifySig() []byte {
	return sr.getMessageToVerifySig()
}

// DoSovereignEndRoundJob -
func (sr *sovereignSubRoundEnd) DoSovereignEndRoundJob(ctx context.Context) bool {
	return sr.doSovereignEndRoundJob(ctx)
}

// DoBlockJob method does the job of the subround Block
func (sr *subroundBlockV2) DoBlockJob() bool {
	return sr.doBlockJob(context.Background())
}
