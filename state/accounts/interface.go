package accounts

import (
	"github.com/multiversx/mx-chain-sovereign-go/state"
)

type dataTrieInteractor interface {
	state.DataTrieTracker
}
