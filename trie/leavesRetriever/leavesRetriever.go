package leavesRetriever

import (
	"context"
	"sync"

	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/trie/leavesRetriever/dfsTrieIterator"
)

type leavesRetriever struct {
	iterators      map[string]common.DfsIterator
	lruIteratorIDs [][]byte
	db             common.TrieStorageInteractor
	marshaller     marshal.Marshalizer
	hasher         hashing.Hasher
	size           uint64
	maxSize        uint64
	mutex          sync.RWMutex
}

// NewLeavesRetriever creates a new leaves retriever
func NewLeavesRetriever(db common.TrieStorageInteractor, marshaller marshal.Marshalizer, hasher hashing.Hasher, maxSize uint64) *leavesRetriever {
	return &leavesRetriever{
		iterators:      make(map[string]common.DfsIterator),
		lruIteratorIDs: make([][]byte, 0),
		db:             db,
		marshaller:     marshaller,
		hasher:         hasher,
		size:           0,
		maxSize:        maxSize,
	}
}

// GetLeaves retrieves the leaves from the trie. If there is a saved checkpoint for the iterator id, it will continue to iterate from the checkpoint.
func (lr *leavesRetriever) GetLeaves(numLeaves int, rootHash []byte, iteratorID []byte, ctx context.Context) (map[string]string, []byte, error) {
	lr.mutex.RLock()
	iterator, ok := lr.iterators[string(iteratorID)]
	lr.mutex.RUnlock()
	if !ok {
		return lr.getLeavesFromNewInstance(numLeaves, rootHash, ctx)
	}

	return lr.getLeavesFromCheckpoint(numLeaves, iterator, iteratorID, ctx)
}

func (lr *leavesRetriever) getLeavesFromNewInstance(numLeaves int, rootHash []byte, ctx context.Context) (map[string]string, []byte, error) {
	iterator, err := dfsTrieIterator.NewIterator(rootHash, lr.db, lr.marshaller, lr.hasher)
	if err != nil {
		return nil, nil, err
	}

	leaves, err := iterator.GetLeaves(numLeaves, ctx)
	if err != nil {
		return nil, nil, err
	}

	if iterator.FinishedIteration() {
		return leaves, nil, nil
	}

	iteratorId := iterator.GetIteratorId()
	if len(iteratorId) == 0 {
		return leaves, nil, nil
	}

	lr.manageIterators(iteratorId, iterator)
	return leaves, iteratorId, nil
}

func (lr *leavesRetriever) getLeavesFromCheckpoint(numLeaves int, iterator common.DfsIterator, iteratorID []byte, ctx context.Context) (map[string]string, []byte, error) {
	lr.markIteratorAsRecentlyUsed(iteratorID)
	clonedIterator := iterator.Clone()
	leaves, err := clonedIterator.GetLeaves(numLeaves, ctx)
	if err != nil {
		return nil, nil, err
	}

	if clonedIterator.FinishedIteration() {
		return leaves, nil, nil
	}

	newIteratorId := clonedIterator.GetIteratorId()
	if len(newIteratorId) == 0 {
		return leaves, nil, nil
	}

	lr.manageIterators(newIteratorId, clonedIterator)
	return leaves, newIteratorId, nil
}

func (lr *leavesRetriever) manageIterators(iteratorId []byte, iterator common.DfsIterator) {
	lr.mutex.Lock()
	defer lr.mutex.Unlock()

	lr.saveIterator(iteratorId, iterator)
	lr.removeIteratorsIfMaxSizeIsExceeded()
}

func (lr *leavesRetriever) saveIterator(iteratorId []byte, iterator common.DfsIterator) {
	lr.lruIteratorIDs = append(lr.lruIteratorIDs, iteratorId)
	lr.iterators[string(iteratorId)] = iterator
	lr.size += iterator.Size() + uint64(len(iteratorId))
}

func (lr *leavesRetriever) markIteratorAsRecentlyUsed(iteratorId []byte) {
	lr.mutex.Lock()
	defer lr.mutex.Unlock()

	for i, id := range lr.lruIteratorIDs {
		if string(id) == string(iteratorId) {
			lr.lruIteratorIDs = append(lr.lruIteratorIDs[:i], lr.lruIteratorIDs[i+1:]...)
			lr.lruIteratorIDs = append(lr.lruIteratorIDs, id)
			return
		}
	}
}

func (lr *leavesRetriever) removeIteratorsIfMaxSizeIsExceeded() {
	if lr.size <= lr.maxSize {
		return
	}

	for i := 0; i < len(lr.lruIteratorIDs); i++ {
		id := lr.lruIteratorIDs[i]
		iterator := lr.iterators[string(id)]
		lr.size -= iterator.Size() + uint64(len(id))
		delete(lr.iterators, string(id))
		lr.lruIteratorIDs = append(lr.lruIteratorIDs[:i], lr.lruIteratorIDs[i+1:]...)

		if lr.size <= lr.maxSize {
			break
		}
	}
}
