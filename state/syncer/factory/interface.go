package factory

import (
	"github.com/multiversx/mx-chain-sovereign-go/process"
	"github.com/multiversx/mx-chain-sovereign-go/state/syncer"
)

// ValidatorAccountsSyncerFactoryHandler defines a factory able to create a validator accounts db syncer
type ValidatorAccountsSyncerFactoryHandler interface {
	CreateValidatorAccountsSyncer(args syncer.ArgsNewValidatorAccountsSyncer) (process.AccountsDBSyncer, error)
	IsInterfaceNil() bool
}
