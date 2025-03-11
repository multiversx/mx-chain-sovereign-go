package enablers

import (
	"github.com/multiversx/mx-chain-sovereign-go/common"
	"github.com/multiversx/mx-chain-sovereign-go/config"
	"github.com/multiversx/mx-chain-sovereign-go/process"
)

// EnableEpochsFactory defines enable epochs handler factory behavior
type EnableEpochsFactory interface {
	CreateEnableEpochsHandler(enableEpochs config.EnableEpochs, epochNotifier process.EpochNotifier) (common.EnableEpochsHandler, error)
	IsInterfaceNil() bool
}
