package bls

import (
	"context"

	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/outport"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
)

// SubRoundStartExtraSignersHolder manages extra signers during start subround in a consensus process
type SubRoundStartExtraSignersHolder interface {
	Reset(pubKeys []string) error
	RegisterExtraSigningHandler(extraSigner consensus.SubRoundStartExtraSignatureHandler) error
	IsInterfaceNil() bool
}

// SubRoundSignatureExtraSignersHolder manages extra signers during signing subround in a consensus process
type SubRoundSignatureExtraSignersHolder interface {
	CreateExtraSignatureShares(header data.HeaderHandler, selfIndex uint16, selfPubKey []byte) (map[string][]byte, error)
	AddExtraSigSharesToConsensusMessage(extraSigShares map[string][]byte, cnsMsg *consensus.Message) error
	StoreExtraSignatureShare(index uint16, cnsMsg *consensus.Message) error
	RegisterExtraSigningHandler(extraSigner consensus.SubRoundSignatureExtraSignatureHandler) error
	IsInterfaceNil() bool
}

// SubRoundEndExtraSignersHolder manages extra signers during end subround in a consensus process
type SubRoundEndExtraSignersHolder interface {
	AggregateSignatures(bitmap []byte, header data.HeaderHandler) (map[string][]byte, error)
	AddLeaderAndAggregatedSignatures(header data.HeaderHandler, cnsMsg *consensus.Message) error
	SignAndSetLeaderSignature(header data.HeaderHandler, leaderPubKey []byte) error
	SetAggregatedSignatureInHeader(header data.HeaderHandler, aggregatedSigs map[string][]byte) error
	VerifyAggregatedSignatures(header data.HeaderHandler, bitmap []byte) error
	HaveConsensusHeaderWithFullInfo(header data.HeaderHandler, cnsMsg *consensus.Message) error
	RegisterExtraSigningHandler(extraSigner consensus.SubRoundEndExtraSignatureHandler) error
	IsInterfaceNil() bool
}

// ExtraSignersHolder manages all extra signer holders
type ExtraSignersHolder interface {
	GetSubRoundStartExtraSignersHolder() SubRoundStartExtraSignersHolder
	GetSubRoundSignatureExtraSignersHolder() SubRoundSignatureExtraSignersHolder
	GetSubRoundEndExtraSignersHolder() SubRoundEndExtraSignersHolder
	IsInterfaceNil() bool
}

// BridgeOperationsHandler handles sending outgoing txs from sovereign to main chain
type BridgeOperationsHandler interface {
	Send(ctx context.Context, data *sovereign.BridgeOperations) (*sovereign.BridgeOperationsResponse, error)
	IsInterfaceNil() bool
}

// OutGoingOperationsPool defines the behavior of a timed cache for outgoing operations
type OutGoingOperationsPool interface {
	Add(data *sovereign.BridgeOutGoingData)
	Get(hash []byte) *sovereign.BridgeOutGoingData
	Delete(hash []byte)
	GetUnconfirmedOperations() []*sovereign.BridgeOutGoingData
	ResetTimer(hashes [][]byte)
	IsInterfaceNil() bool
}

// SubRoundHandler defines a sub round handler (e.g.: v1,v2)
type SubRoundHandler interface {
	spos.ConsensusCoreHandler
	spos.ConsensusStateHandler
	consensus.SubroundHandler
}

// SubRoundStartHandler defines a sub round start handler
type SubRoundStartHandler interface {
	SubRoundHandler
	SetOutportHandler(outportHandler outport.OutportHandler) error
}

// SubRoundBlockHandler defines a sub round block handler
type SubRoundBlockHandler interface {
	SubRoundHandler
	SetBlockJob(doBlockJob func(ctx context.Context) bool)
	DoBlockComputation() (*SubRoundBlockProcessRes, func())
	ProcessReceivedBlock(ctx context.Context, cnsDta *consensus.Message) bool
}

// SubRoundEndHandler defines a sub round end handler
type SubRoundEndHandler interface {
	SubRoundHandler
	SetMessageToVerifySigFunc(verifyMsgFunc func() []byte)
	SetBlockJob(doBlockJob func(ctx context.Context) bool)
	ReceivedBlockHeaderFinalInfo(ctx context.Context, cnsDta *consensus.Message) bool
	DoEndRoundJob(ctx context.Context) bool
	IsSelfLeaderInCurrentRound() bool
}

// SubRoundSignatureHandler defines a sub round signature handler
type SubRoundSignatureHandler interface {
	SubRoundHandler
	SetMessageToSignFunc(verifyMsgFunc func() []byte)
}
