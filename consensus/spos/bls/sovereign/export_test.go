package sovereign

import (
	"context"

	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/testscommon/subRounds"
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

// GenerateBlockSubroundV2 generates the instance of subround Block V2 and added it to the chronology subrounds list
func (fct *factory) GenerateBlockSubroundV2() error {
	return fct.generateBlockSubroundV2()
}

// GenerateSignatureSubroundV2 generates the instance of subround Signature V2 and added it to the chronology subrounds list
func (fct *factory) GenerateSignatureSubroundV2() error {
	return fct.generateSignatureSubroundV2(&subRounds.SubRoundSignatureExtraSignersHolderMock{})
}

// GenerateEndRoundSubroundV2 generates the instance of subround EndRound V2 and added it to the chronology subrounds list
func (fct *factory) GenerateEndRoundSubroundV2() error {
	return fct.generateEndRoundSubroundV2(&subRounds.SubRoundEndExtraSignersHolderMock{})
}
