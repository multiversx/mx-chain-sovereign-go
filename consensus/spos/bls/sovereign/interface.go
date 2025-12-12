package sovereign

import (
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
)

// SubRoundsFactoryHandler defines what a base sub round factory should be able to do
type SubRoundsFactoryHandler interface {
	GenerateStartRoundSubround() (bls.SubRoundStartHandler, error)
	GenerateBlockSubround() (bls.SubRoundBlockHandler, error)
	GenerateSignatureSubround() (bls.SubRoundSignatureHandler, error)
	GenerateEndRoundSubround() (bls.SubRoundEndHandler, error)
}
