package facade

import (
	"github.com/multiversx/mx-chain-go/ntp"
)

// GetSyncer returns the current syncer
func (nf *NodeFacade) GetSyncer() ntp.SyncTimer {
	return nf.syncer
}
