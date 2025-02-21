package chaos

import (
	"github.com/multiversx/mx-chain-core-go/data"
)

type PointInput struct {
	Name           string
	ConsensusState MyConsensusStateHandler
	NodePublicKey  string
	Header         data.HeaderHandler
	Signature      []byte
}
