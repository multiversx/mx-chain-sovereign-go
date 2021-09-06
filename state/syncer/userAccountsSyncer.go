package syncer

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/trie"
	"github.com/ElrondNetwork/elrond-go/trie/statistics"
)

var _ epochStart.AccountsDBSyncer = (*userAccountsSyncer)(nil)

var log = logger.GetOrCreate("syncer")

const timeBetweenRetries = 100 * time.Millisecond

type userAccountsSyncer struct {
	*baseAccountsSyncer
	throttler   data.GoRoutineThrottler
	syncerMutex sync.Mutex
}

// ArgsNewUserAccountsSyncer defines the arguments needed for the new account syncer
type ArgsNewUserAccountsSyncer struct {
	ArgsNewBaseAccountsSyncer
	ShardId   uint32
	Throttler data.GoRoutineThrottler
}

// NewUserAccountsSyncer creates a user account syncer
func NewUserAccountsSyncer(args ArgsNewUserAccountsSyncer) (*userAccountsSyncer, error) {
	err := checkArgs(args.ArgsNewBaseAccountsSyncer)
	if err != nil {
		return nil, err
	}

	if check.IfNil(args.Throttler) {
		return nil, data.ErrNilThrottler
	}

	timeoutHandler, err := common.NewTimeoutHandler(args.Timeout)
	if err != nil {
		return nil, err
	}

	b := &baseAccountsSyncer{
		hasher:                    args.Hasher,
		marshalizer:               args.Marshalizer,
		dataTries:                 make(map[string]struct{}),
		trieStorageManager:        args.TrieStorageManager,
		requestHandler:            args.RequestHandler,
		timeoutHandler:            timeoutHandler,
		shardId:                   args.ShardId,
		cacher:                    args.Cacher,
		rootHash:                  nil,
		maxTrieLevelInMemory:      args.MaxTrieLevelInMemory,
		name:                      fmt.Sprintf("user accounts for shard %s", core.GetShardIDString(args.ShardId)),
		maxHardCapForMissingNodes: args.MaxHardCapForMissingNodes,
		trieSyncerVersion:         args.TrieSyncerVersion,
	}

	u := &userAccountsSyncer{
		baseAccountsSyncer: b,
		throttler:          args.Throttler,
	}

	return u, nil
}

// SyncAccounts will launch the syncing method to gather all the data needed for userAccounts - it is a blocking method
func (u *userAccountsSyncer) SyncAccounts(rootHash []byte) error {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	u.timeoutHandler.ResetWatchdog()

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		u.cacher.Clear()
		cancel()
	}()

	tss := statistics.NewTrieSyncStatistics()
	go u.printStatistics(tss, ctx)

	mainTrie, err := u.syncMainTrie(rootHash, factory.AccountTrieNodesTopic, tss, ctx)
	if err != nil {
		return err
	}

	defer func() {
		_ = mainTrie.Close()
	}()

	log.Debug("main trie synced, starting to sync data tries", "num data tries", len(u.dataTries))

	rootHashes, err := u.findAllAccountRootHashes(mainTrie)
	if err != nil {
		return err
	}

	err = u.syncAccountDataTries(rootHashes, tss, ctx)
	if err != nil {
		return err
	}

	return nil
}

func (u *userAccountsSyncer) syncAccountDataTries(rootHashes [][]byte, ssh data.SyncStatisticsHandler, ctx context.Context) error {
	var errFound error
	errMutex := sync.Mutex{}

	wg := sync.WaitGroup{}
	wg.Add(len(rootHashes))

	for _, rootHash := range rootHashes {
		for {
			if u.throttler.CanProcess() {
				break
			}

			select {
			case <-time.After(timeBetweenRetries):
				continue
			case <-ctx.Done():
				return data.ErrTimeIsOut
			}
		}

		//throttler.StartProcessing is required to be here as it prevent the following edge-case:
		//loop does 100k iterations because throttler.CanProcess allows it and starts 100k go routines that can
		//not execute their first instructions that could tell the throttler they have started.
		//Telling the throttler the processing has started here will prevent OOM exception because of too many go
		//routines started.
		u.throttler.StartProcessing()

		go func(trieRootHash []byte) {
			defer u.throttler.EndProcessing()

			log.Trace("sync data trie", "roothash", trieRootHash)
			newErr := u.syncDataTrie(trieRootHash, ssh, ctx)
			if newErr != nil {
				errMutex.Lock()
				errFound = newErr
				errMutex.Unlock()
			}
			atomic.AddInt32(&u.numTriesSynced, 1)
			log.Trace("finished sync data trie", "roothash", trieRootHash)
			wg.Done()
		}(rootHash)
	}

	wg.Wait()

	errMutex.Lock()
	defer errMutex.Unlock()

	return errFound
}

func (u *userAccountsSyncer) syncDataTrie(rootHash []byte, ssh data.SyncStatisticsHandler, ctx context.Context) error {
	u.syncerMutex.Lock()
	_, ok := u.dataTries[string(rootHash)]
	if ok {
		u.syncerMutex.Unlock()
		return nil
	}

	u.dataTries[string(rootHash)] = struct{}{}
	u.syncerMutex.Unlock()

	arg := trie.ArgTrieSyncer{
		RequestHandler:            u.requestHandler,
		InterceptedNodes:          u.cacher,
		DB:                        u.trieStorageManager.Database(),
		Marshalizer:               u.marshalizer,
		Hasher:                    u.hasher,
		ShardId:                   u.shardId,
		Topic:                     factory.AccountTrieNodesTopic,
		TrieSyncStatistics:        ssh,
		TimeoutHandler:            u.timeoutHandler,
		MaxHardCapForMissingNodes: u.maxHardCapForMissingNodes,
	}
	trieSyncer, err := trie.CreateTrieSyncer(arg, u.trieSyncerVersion)
	if err != nil {

		return err
	}

	err = trieSyncer.StartSyncing(rootHash, ctx)
	if err != nil {
		return err
	}

	return nil
}

func (u *userAccountsSyncer) findAllAccountRootHashes(mainTrie common.Trie) ([][]byte, error) {
	mainRootHash, err := mainTrie.RootHash()
	if err != nil {
		return nil, err
	}

	leavesChannel, err := mainTrie.GetAllLeavesOnChannel(mainRootHash)
	if err != nil {
		return nil, err
	}

	rootHashes := make([][]byte, 0)
	for leaf := range leavesChannel {
		u.resetTimeoutHandlerWatchdog()

		account := state.NewEmptyUserAccount()
		err = u.marshalizer.Unmarshal(account, leaf.Value())
		if err != nil {
			log.Trace("this must be a leaf with code", "err", err)
			continue
		}

		if len(account.RootHash) > 0 {
			rootHashes = append(rootHashes, account.RootHash)
			atomic.AddInt32(&u.numMaxTries, 1)
		}
	}

	return rootHashes, nil
}

// resetTimeoutHandlerWatchdog this method should be called whenever the syncer is doing something other than
// requesting trie nodes as to prevent the sync process being terminated prematurely.
func (u *userAccountsSyncer) resetTimeoutHandlerWatchdog() {
	u.timeoutHandler.ResetWatchdog()
}
