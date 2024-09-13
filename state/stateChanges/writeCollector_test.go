package stateChanges

import (
	"fmt"
	"strconv"
	"testing"

	data "github.com/multiversx/mx-chain-core-go/data/stateChange"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

func getDefaultStateChange() *data.StateChange {
	return &data.StateChange{
		Type: "write",
	}
}

func TestNewStateChangesCollector(t *testing.T) {
	t.Parallel()

	stateChangesCollector := NewStateChangesCollector()
	require.False(t, stateChangesCollector.IsInterfaceNil())
}

func TestStateChangesCollector_AddStateChange(t *testing.T) {
	t.Parallel()

	scc := NewStateChangesCollector()
	assert.Equal(t, 0, len(scc.stateChanges))

	numStateChanges := 10
	for i := 0; i < numStateChanges; i++ {
		scc.AddStateChange(getDefaultStateChange())
	}
	assert.Equal(t, numStateChanges, len(scc.stateChanges))
}

func TestStateChangesCollector_GetStateChanges(t *testing.T) {
	t.Parallel()

	t.Run("getStateChanges with tx hash", func(t *testing.T) {
		t.Parallel()

		scc := NewStateChangesCollector()
		assert.Equal(t, 0, len(scc.stateChanges))
		assert.Equal(t, 0, len(scc.GetStateChanges()))

		numStateChanges := 10
		for i := 0; i < numStateChanges; i++ {
			scc.AddStateChange(&data.StateChange{
				Type:        "write",
				MainTrieKey: []byte(strconv.Itoa(i)),
			})
		}
		assert.Equal(t, numStateChanges, len(scc.stateChanges))
		assert.Equal(t, 0, len(scc.GetStateChanges()))
		scc.AddTxHashToCollectedStateChanges([]byte("txHash"), &transaction.Transaction{})
		assert.Equal(t, numStateChanges, len(scc.stateChanges))
		assert.Equal(t, 1, len(scc.GetStateChanges()))
		assert.Equal(t, []byte("txHash"), scc.GetStateChanges()[0].TxHash)
		assert.Equal(t, numStateChanges, len(scc.GetStateChanges()[0].StateChanges))

		stateChangesForTx := scc.GetStateChanges()
		assert.Equal(t, 1, len(stateChangesForTx))
		assert.Equal(t, []byte("txHash"), stateChangesForTx[0].TxHash)
		for i := 0; i < len(stateChangesForTx[0].StateChanges); i++ {
			sc, ok := stateChangesForTx[0].StateChanges[i].(*data.StateChange)
			require.True(t, ok)

			assert.Equal(t, []byte(strconv.Itoa(i)), sc.MainTrieKey)
		}
	})

	t.Run("getStateChanges without tx hash", func(t *testing.T) {
		t.Parallel()

		scc := NewStateChangesCollector()
		assert.Equal(t, 0, len(scc.stateChanges))
		assert.Equal(t, 0, len(scc.GetStateChanges()))

		numStateChanges := 10
		for i := 0; i < numStateChanges; i++ {
			scc.AddStateChange(&data.StateChange{
				Type:        "write",
				MainTrieKey: []byte(strconv.Itoa(i)),
			})
		}
		assert.Equal(t, numStateChanges, len(scc.stateChanges))
		assert.Equal(t, 0, len(scc.GetStateChanges()))

		stateChangesForTx := scc.GetStateChanges()
		assert.Equal(t, 0, len(stateChangesForTx))
	})
}

func TestStateChangesCollector_AddTxHashToCollectedStateChanges(t *testing.T) {
	t.Parallel()

	scc := NewStateChangesCollector()
	assert.Equal(t, 0, len(scc.stateChanges))
	assert.Equal(t, 0, len(scc.GetStateChanges()))

	scc.AddTxHashToCollectedStateChanges([]byte("txHash0"), &transaction.Transaction{})

	stateChange := &data.StateChange{
		Type:            "write",
		MainTrieKey:     []byte("mainTrieKey"),
		MainTrieVal:     []byte("mainTrieVal"),
		DataTrieChanges: []*data.DataTrieChange{{Key: []byte("dataTrieKey"), Val: []byte("dataTrieVal")}},
	}
	scc.AddStateChange(stateChange)

	assert.Equal(t, 1, len(scc.stateChanges))
	assert.Equal(t, 0, len(scc.GetStateChanges()))
	scc.AddTxHashToCollectedStateChanges([]byte("txHash"), &transaction.Transaction{})
	assert.Equal(t, 1, len(scc.stateChanges))
	assert.Equal(t, 1, len(scc.GetStateChanges()))

	stateChangesForTx := scc.GetStateChanges()
	assert.Equal(t, 1, len(stateChangesForTx))
	assert.Equal(t, []byte("txHash"), stateChangesForTx[0].TxHash)
	assert.Equal(t, 1, len(stateChangesForTx[0].StateChanges))

	sc, ok := stateChangesForTx[0].StateChanges[0].(*data.StateChange)
	require.True(t, ok)

	assert.Equal(t, []byte("mainTrieKey"), sc.MainTrieKey)
	assert.Equal(t, []byte("mainTrieVal"), sc.MainTrieVal)
	assert.Equal(t, 1, len(sc.DataTrieChanges))
}

func TestStateChangesCollector_RevertToIndex_FailIfWrongIndex(t *testing.T) {
	t.Parallel()

	scc := NewStateChangesCollector()
	numStateChanges := len(scc.stateChanges)

	err := scc.RevertToIndex(-1)
	require.Equal(t, ErrStateChangesIndexOutOfBounds, err)

	err = scc.RevertToIndex(numStateChanges + 1)
	require.Equal(t, ErrStateChangesIndexOutOfBounds, err)
}

func TestStateChangesCollector_RevertToIndex(t *testing.T) {
	t.Parallel()

	scc := NewStateChangesCollector()

	numStateChanges := 10
	for i := 0; i < numStateChanges; i++ {
		scc.AddStateChange(getDefaultStateChange())
		err := scc.SetIndexToLastStateChange(i)
		require.Nil(t, err)
	}
	scc.AddTxHashToCollectedStateChanges([]byte("txHash1"), &transaction.Transaction{})

	for i := numStateChanges; i < numStateChanges*2; i++ {
		scc.AddStateChange(getDefaultStateChange())
		scc.AddTxHashToCollectedStateChanges([]byte("txHash"+fmt.Sprintf("%d", i)), &transaction.Transaction{})
	}
	err := scc.SetIndexToLastStateChange(numStateChanges)
	require.Nil(t, err)

	assert.Equal(t, numStateChanges*2, len(scc.stateChanges))

	err = scc.RevertToIndex(numStateChanges)
	require.Nil(t, err)
	assert.Equal(t, numStateChanges*2-1, len(scc.stateChanges))

	err = scc.RevertToIndex(numStateChanges - 1)
	require.Nil(t, err)
	assert.Equal(t, numStateChanges-1, len(scc.stateChanges))

	err = scc.RevertToIndex(numStateChanges / 2)
	require.Nil(t, err)
	assert.Equal(t, numStateChanges/2, len(scc.stateChanges))

	err = scc.RevertToIndex(1)
	require.Nil(t, err)
	assert.Equal(t, 1, len(scc.stateChanges))

	err = scc.RevertToIndex(0)
	require.Nil(t, err)
	assert.Equal(t, 0, len(scc.stateChanges))
}

func TestStateChangesCollector_SetIndexToLastStateChange(t *testing.T) {
	t.Parallel()

	t.Run("should fail if valid index", func(t *testing.T) {
		t.Parallel()

		scc := NewStateChangesCollector()

		err := scc.SetIndexToLastStateChange(-1)
		require.Equal(t, ErrStateChangesIndexOutOfBounds, err)

		numStateChanges := len(scc.stateChanges)
		err = scc.SetIndexToLastStateChange(numStateChanges + 1)
		require.Equal(t, ErrStateChangesIndexOutOfBounds, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		scc := NewStateChangesCollector()

		numStateChanges := 10
		for i := 0; i < numStateChanges; i++ {
			scc.AddStateChange(getDefaultStateChange())
			err := scc.SetIndexToLastStateChange(i)
			require.Nil(t, err)
		}
		scc.AddTxHashToCollectedStateChanges([]byte("txHash1"), &transaction.Transaction{})

		for i := numStateChanges; i < numStateChanges*2; i++ {
			scc.AddStateChange(getDefaultStateChange())
			scc.AddTxHashToCollectedStateChanges([]byte("txHash"+fmt.Sprintf("%d", i)), &transaction.Transaction{})
		}
		err := scc.SetIndexToLastStateChange(numStateChanges)
		require.Nil(t, err)

		assert.Equal(t, numStateChanges*2, len(scc.stateChanges))
	})
}

func TestStateChangesCollector_Reset(t *testing.T) {
	t.Parallel()

	scc := NewStateChangesCollector()
	assert.Equal(t, 0, len(scc.stateChanges))

	numStateChanges := 10
	for i := 0; i < numStateChanges; i++ {
		scc.AddStateChange(getDefaultStateChange())
	}
	scc.AddTxHashToCollectedStateChanges([]byte("txHash"), &transaction.Transaction{})
	for i := numStateChanges; i < numStateChanges*2; i++ {
		scc.AddStateChange(getDefaultStateChange())
	}
	assert.Equal(t, numStateChanges*2, len(scc.stateChanges))

	assert.Equal(t, 1, len(scc.GetStateChanges()))

	scc.Reset()
	assert.Equal(t, 0, len(scc.stateChanges))

	assert.Equal(t, 0, len(scc.GetStateChanges()))
}

func TestStateChangesCollector_GetStateChangesForTx(t *testing.T) {
	t.Parallel()

	scc := NewStateChangesCollector()
	assert.Equal(t, 0, len(scc.stateChanges))

	numStateChanges := 10
	for i := 0; i < numStateChanges; i++ {
		scc.AddStateChange(&data.StateChange{
			Type: "write",
			// distribute evenly based on parity of the index
			TxHash: []byte(fmt.Sprintf("hash%d", i%2)),
		})
	}

	stateChangesForTx := scc.GetStateChangesForTxs()

	require.Len(t, stateChangesForTx, 2)
	require.Len(t, stateChangesForTx["hash0"].StateChanges, 5)
	require.Len(t, stateChangesForTx["hash1"].StateChanges, 5)

	require.Equal(t, stateChangesForTx, map[string]*data.StateChanges{
		"hash0" : {
			[]*data.StateChange{
				{Type: "write", TxHash: []byte("hash0")},
				{Type: "write", TxHash: []byte("hash0")},
				{Type: "write", TxHash: []byte("hash0")},
				{Type: "write", TxHash: []byte("hash0")},
				{Type: "write", TxHash: []byte("hash0")},
			},
		},
		"hash1" : {
			[]*data.StateChange{
				{Type: "write", TxHash: []byte("hash1")},
				{Type: "write", TxHash: []byte("hash1")},
				{Type: "write", TxHash: []byte("hash1")},
				{Type: "write", TxHash: []byte("hash1")},
				{Type: "write", TxHash: []byte("hash1")},
			},
		},
	})
}
