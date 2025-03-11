package storage

import (
	"github.com/multiversx/mx-chain-sovereign-go/common/statistics/disabled"
	"github.com/multiversx/mx-chain-sovereign-go/config"
	"github.com/multiversx/mx-chain-sovereign-go/dataRetriever"
	"github.com/multiversx/mx-chain-sovereign-go/genesis/mock"
	"github.com/multiversx/mx-chain-sovereign-go/testscommon"
	"github.com/multiversx/mx-chain-sovereign-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-sovereign-go/trie"
)

// GetStorageManagerArgs returns mock args for trie storage manager creation
func GetStorageManagerArgs() trie.NewTrieStorageManagerArgs {
	return trie.NewTrieStorageManagerArgs{
		MainStorer:  testscommon.NewSnapshotPruningStorerMock(),
		Marshalizer: &mock.MarshalizerMock{},
		Hasher:      &hashingMocks.HasherMock{},
		GeneralConfig: config.TrieStorageManagerConfig{
			PruningBufferLen:      1000,
			SnapshotsBufferLen:    10,
			SnapshotsGoroutineNum: 2,
		},
		IdleProvider:   &testscommon.ProcessStatusHandlerStub{},
		Identifier:     dataRetriever.UserAccountsUnit.String(),
		StatsCollector: disabled.NewStateStatistics(),
	}
}

// GetStorageManagerOptions returns default options for trie storage manager creation
func GetStorageManagerOptions() trie.StorageManagerOptions {
	return trie.StorageManagerOptions{
		PruningEnabled:   true,
		SnapshotsEnabled: true,
	}
}
