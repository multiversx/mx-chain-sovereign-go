package factory

import (
	"github.com/multiversx/mx-chain-sovereign-go/config"
	"github.com/multiversx/mx-chain-sovereign-go/storage"
	"github.com/multiversx/mx-chain-sovereign-go/storage/databaseremover"
	"github.com/multiversx/mx-chain-sovereign-go/storage/databaseremover/disabled"
)

// CreateCustomDatabaseRemover will handle the creation of a custom database remover based on the configuration
func CreateCustomDatabaseRemover(storagePruningConfig config.StoragePruningConfig) (storage.CustomDatabaseRemoverHandler, error) {
	if storagePruningConfig.AccountsTrieCleanOldEpochsData {
		return databaseremover.NewCustomDatabaseRemover(storagePruningConfig)
	}

	return disabled.NewDisabledCustomDatabaseRemover(), nil
}
