package storageBootstrap

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/sovereign/dto"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/bootstrapStorage"
	sovereignMocks "github.com/multiversx/mx-chain-go/testscommon/sovereign"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createShardStorageBootstrapper() *shardStorageBootstrapper {
	baseArgs := createMockShardStorageBootstrapperArgs()
	args := ArgsShardStorageBootstrapper{
		ArgsBaseStorageBootstrapper: baseArgs,
	}

	ssb, _ := NewShardStorageBootstrapper(args)
	return ssb
}

func TestNewSovereignChainBaseStorageBootstrapper(t *testing.T) {
	t.Parallel()

	t.Run("nil shard storage bootstrapper", func(t *testing.T) {
		scssb, err := NewSovereignChainShardStorageBootstrapper(nil, &sovereignMocks.OutGoingChainNonceMock{})
		require.Nil(t, scssb)
		require.Equal(t, process.ErrNilShardStorageBootstrapper, err)
	})
	t.Run("nil outgoing op nonce chain handler", func(t *testing.T) {
		scssb, err := NewSovereignChainShardStorageBootstrapper(createShardStorageBootstrapper(), nil)
		require.Nil(t, scssb)
		require.Equal(t, errors.ErrNilOutGoingOpNonceChainHandler, err)
	})
	t.Run("should work", func(t *testing.T) {
		scssb, err := NewSovereignChainShardStorageBootstrapper(createShardStorageBootstrapper(), &sovereignMocks.OutGoingChainNonceMock{})
		require.Nil(t, err)
		require.False(t, scssb.IsInterfaceNil())
	})
}

func TestBaseStorageBootstrapper_SovereignChainGetScheduledRootHash(t *testing.T) {
	t.Parallel()

	ssb := createShardStorageBootstrapper()
	scssb, _ := NewSovereignChainShardStorageBootstrapper(ssb, &sovereignMocks.OutGoingChainNonceMock{})

	expectedRootHash := []byte("rootHash")
	hdr := &block.Header{
		RootHash: expectedRootHash,
	}
	rootHash := scssb.sovereignChainGetScheduledRootHash(hdr, nil)

	assert.Equal(t, expectedRootHash, rootHash)
}

func TestSovereignChainShardStorageBootstrapper_applyCrossChainOutGoingData(t *testing.T) {
	t.Parallel()

	chains := []dto.ChainID{dto.MVX, dto.ETH}
	nonces := []uint64{4, 9}
	bootStrapData := []bootstrapStorage.BootstrapOutGoingData{
		{
			ChainID:       int32(chains[0]),
			OutGoingNonce: nonces[0],
		},
		{
			ChainID:       int32(chains[1]),
			OutGoingNonce: nonces[1],
		},
	}

	setNonceCt := 0
	outGoingOpNonceChainHandler := &sovereignMocks.OutGoingChainNonceMock{
		SetNonceCalled: func(chainID dto.ChainID, nonce uint64) {
			require.Equal(t, chainID, chains[setNonceCt])
			require.Equal(t, nonce, nonces[setNonceCt])
			setNonceCt++
		},
	}

	ssb := createShardStorageBootstrapper()
	scssb, _ := NewSovereignChainShardStorageBootstrapper(ssb, outGoingOpNonceChainHandler)

	scssb.applyCrossChainOutGoingData(bootStrapData)
	require.Equal(t, 2, setNonceCt)
}
