package sovereign

import (
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/multiversx/mx-chain-core-go/data/sovereign/dto"
	"github.com/stretchr/testify/require"
)

func TestShardedOutGoingOperationPool_Basic_MultiChain(t *testing.T) {
	pool := NewShardedOutGoingOperationPool(200 * time.Millisecond)

	chainIDs := []dto.ChainID{dto.MVX, dto.ETH, dto.SUI}
	dataPerChain := make(map[dto.ChainID]*sovereign.BridgeOutGoingData)

	for i, chainID := range chainIDs {
		data := &sovereign.BridgeOutGoingData{
			Hash: []byte("hash-" + string(rune('A'+i))),
			OutGoingOperations: []*sovereign.OutGoingOperation{
				{
					Hash: []byte("hash-" + string(rune('B'+i))),
				},
			},
		}
		dataPerChain[chainID] = data
		pool.Add(data, chainID)
	}

	for _, chainID := range chainIDs {
		data := dataPerChain[chainID]
		got := pool.Get(data.Hash, chainID)
		require.NotNil(t, got, "should get data for chain", chainID.String())
		require.Equal(t, data.Hash, got.Hash)
	}

	time.Sleep(200 * time.Millisecond)
	unconfirmed := pool.GetUnconfirmedOperations()
	require.Len(t, unconfirmed, len(chainIDs))

	ethData := dataPerChain[dto.ETH]
	pool.ResetTimer([][]byte{ethData.Hash}, dto.ETH)
	unconfirmed = pool.GetUnconfirmedOperations()
	require.Len(t, unconfirmed, len(chainIDs)-1)

	err := pool.ConfirmOperation(ethData.Hash, ethData.OutGoingOperations[0].Hash, dto.ETH)
	require.Nil(t, err)

	for _, chainID := range chainIDs {
		data := dataPerChain[chainID]

		got := pool.Get(data.Hash, chainID)
		if chainID == dto.ETH {
			require.Nil(t, got, "data should be deleted for chain", chainID.String())
		} else {
			require.NotNil(t, got, "data should not be deleted for chain", chainID.String())
		}

		pool.Delete(data.Hash, chainID)
		got = pool.Get(data.Hash, chainID)
		require.Nil(t, got, "data should be deleted for chain", chainID.String())
	}

	time.Sleep(200 * time.Millisecond)
	unconfirmed = pool.GetUnconfirmedOperations()
	require.Empty(t, unconfirmed)
}

func TestShardedOutGoingOperationPool_ConcurrentAccess(t *testing.T) {
	pool := NewShardedOutGoingOperationPool(time.Millisecond)
	chainIDs := []dto.ChainID{dto.MVX, dto.ETH, dto.SUI}

	var wg sync.WaitGroup
	opCount := 1000

	for i := 0; i < opCount; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			chainID := chainIDs[rand.Intn(len(chainIDs))]
			hash := []byte("hash-" + string(rune(i)))
			data := &sovereign.BridgeOutGoingData{Hash: hash}

			switch rand.Intn(6) {
			case 0:
				pool.Add(data, chainID)
			case 1:
				_ = pool.Get(hash, chainID)
			case 2:
				pool.Delete(hash, chainID)
			case 3:
				_ = pool.ConfirmOperation(hash, hash, chainID)
			case 4:
				_ = pool.GetUnconfirmedOperations()
			case 5:
				pool.ResetTimer([][]byte{hash}, chainID)
			}
		}(i)
	}

	wg.Wait()
}
