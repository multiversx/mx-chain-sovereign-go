package sovereign

import (
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
)

type SubRoundsFactoryHandler interface {
	GenerateStartRoundSubround() (bls.SubRoundStartHandler, error)
	GenerateBlockSubround() (bls.SubRoundBlockHandler, error)
	GenerateSignatureSubround() (bls.SubRoundSignatureHandler, error)
	GenerateEndRoundSubround() (bls.SubRoundEndHandler, error)
}
