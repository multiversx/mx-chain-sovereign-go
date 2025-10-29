package operationFormatters

import (
	"fmt"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/sovereign/dto"
	"github.com/stretchr/testify/require"
)

func TestOutGoingOpChainNonce_GetAndIncrementNonce(t *testing.T) {
	t.Parallel()

	op := NewOutGoingOpChainNonce()

	numOperations := 10
	wg := sync.WaitGroup{}
	wg.Add(numOperations)

	nonces := make(chan uint64, numOperations)
	for i := 0; i < numOperations; i++ {
		go func() {
			defer wg.Done()

			if i%2 == 0 {
				nonce, err := op.GetAndIncrementNonce(dto.SUI)
				require.ErrorIs(t, err, errChainIDNotSupported)
				require.Zero(t, nonce)
				return
			}

			nonce, err := op.GetAndIncrementNonce(dto.MVX)
			require.Nil(t, err)
			nonces <- nonce
		}()
	}

	wg.Wait()
	close(nonces)

	vals := make(map[uint64]bool)
	for n := range nonces {
		require.False(t, vals[n], fmt.Sprintf("duplicate nonce: %d", n))
		vals[n] = true
	}
	require.Equal(t, uint64(5), op.data[dto.MVX])
}
