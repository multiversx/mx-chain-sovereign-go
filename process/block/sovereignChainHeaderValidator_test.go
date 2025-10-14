package block_test

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	coreBlock "github.com/multiversx/mx-chain-core-go/data/block"
	coreDTO "github.com/multiversx/mx-chain-core-go/data/sovereign/dto"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createSovHdrValidator() process.HeaderConstructionValidator {
	argsHeaderValidator := block.ArgsHeaderValidator{
		Hasher:              &mock.HasherStub{},
		Marshalizer:         &mock.MarshalizerMock{},
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
	}
	hv, _ := block.NewHeaderValidator(argsHeaderValidator)
	schv, _ := block.NewSovereignChainHeaderValidator(hv)
	return schv
}

func TestNewSovereignChainHeaderValidator_ShouldErrNilHeaderValidator(t *testing.T) {
	t.Parallel()

	schv, err := block.NewSovereignChainHeaderValidator(nil)
	assert.Nil(t, schv)
	assert.Equal(t, process.ErrNilHeaderValidator, err)
}

func TestNewSovereignChainHeaderValidator_ShouldWork(t *testing.T) {
	t.Parallel()

	argsHeaderValidator := block.ArgsHeaderValidator{
		Hasher:              &mock.HasherStub{},
		Marshalizer:         &mock.MarshalizerMock{},
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
	}
	hv, _ := block.NewHeaderValidator(argsHeaderValidator)

	schv, err := block.NewSovereignChainHeaderValidator(hv)
	assert.NotNil(t, schv)
	assert.Nil(t, err)
}

func TestGetHeaderHash_ShouldWork(t *testing.T) {
	t.Parallel()

	t.Run("should error nil header handler", func(t *testing.T) {
		t.Parallel()

		argsHeaderValidator := block.ArgsHeaderValidator{
			Hasher:              &mock.HasherStub{},
			Marshalizer:         &mock.MarshalizerMock{},
			EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		}
		hv, _ := block.NewHeaderValidator(argsHeaderValidator)
		schv, _ := block.NewSovereignChainHeaderValidator(hv)

		shardHeaderExtended := &coreBlock.ShardHeaderExtended{}
		hash, err := schv.CalculateHeaderHash(shardHeaderExtended)
		assert.Nil(t, hash)
		assert.Equal(t, process.ErrNilHeaderHandler, err)
	})

	t.Run("should work for shard header extended handler", func(t *testing.T) {
		t.Parallel()

		argsHeaderValidator := block.ArgsHeaderValidator{
			Hasher:              &hashingMocks.HasherMock{},
			Marshalizer:         &mock.MarshalizerMock{},
			EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		}
		hv, _ := block.NewHeaderValidator(argsHeaderValidator)
		schv, _ := block.NewSovereignChainHeaderValidator(hv)

		shardHeaderExtended := &coreBlock.ShardHeaderExtended{
			Header: &coreBlock.HeaderV2{
				Header: &coreBlock.Header{},
			},
		}

		expectedHash, _ := core.CalculateHash(argsHeaderValidator.Marshalizer, argsHeaderValidator.Hasher, shardHeaderExtended.Header)
		hash, err := schv.CalculateHeaderHash(shardHeaderExtended)
		assert.Nil(t, err)
		assert.Equal(t, expectedHash, hash)
	})

	t.Run("should work for header handler", func(t *testing.T) {
		t.Parallel()

		argsHeaderValidator := block.ArgsHeaderValidator{
			Hasher:              &hashingMocks.HasherMock{},
			Marshalizer:         &mock.MarshalizerMock{},
			EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		}
		hv, _ := block.NewHeaderValidator(argsHeaderValidator)
		schv, _ := block.NewSovereignChainHeaderValidator(hv)

		header := &coreBlock.Header{}

		expectedHash, _ := core.CalculateHash(argsHeaderValidator.Marshalizer, argsHeaderValidator.Hasher, header)
		hash, err := schv.CalculateHeaderHash(header)
		assert.Nil(t, err)
		assert.Equal(t, expectedHash, hash)
	})
}

func TestSovereignChainHeaderValidator_IsHeaderConstructionValid(t *testing.T) {
	t.Parallel()

	t.Run("should work for ETH hdr without hashes check", func(t *testing.T) {
		sovHdrValidator := createSovHdrValidator()

		currHdr := &coreBlock.ShardHeaderExtended{
			Header: &coreBlock.HeaderV2{
				Header: &coreBlock.Header{
					Nonce: 2,
					Round: 2,
				},
			},
			SourceChainID: coreDTO.ETH,
		}
		prevHdr := &coreBlock.ShardHeaderExtended{
			Header: &coreBlock.HeaderV2{
				Header: &coreBlock.Header{
					Nonce: 1,
					Round: 1,
				},
			},
			SourceChainID: coreDTO.ETH,
		}

		require.Nil(t, sovHdrValidator.IsHeaderConstructionValid(currHdr, prevHdr))

		_ = currHdr.SetRound(1)
		require.Equal(t, process.ErrLowerRoundInBlock, sovHdrValidator.IsHeaderConstructionValid(currHdr, prevHdr))
	})

	t.Run("should work for MVX hdr with hashes check", func(t *testing.T) {
		argsHeaderValidator := block.ArgsHeaderValidator{
			Hasher:              &hashingMocks.HasherMock{},
			Marshalizer:         &mock.MarshalizerMock{},
			EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		}
		hdrValidator, _ := block.NewHeaderValidator(argsHeaderValidator)
		sovHdrValidator, _ := block.NewSovereignChainHeaderValidator(hdrValidator)

		randSeed := []byte("randSeed")
		prevHdr := &coreBlock.ShardHeaderExtended{
			Header: &coreBlock.HeaderV2{
				Header: &coreBlock.Header{
					Nonce:    1,
					Round:    1,
					PrevHash: randSeed,
				},
			},
			SourceChainID: coreDTO.MVX,
		}
		// Header hash check should be done on internal original Header, not on extended header
		prevHdrHash, _ := core.CalculateHash(argsHeaderValidator.Marshalizer, argsHeaderValidator.Hasher, prevHdr.Header)

		currHdr := &coreBlock.ShardHeaderExtended{
			Header: &coreBlock.HeaderV2{
				Header: &coreBlock.Header{
					Nonce:    2,
					Round:    2,
					PrevHash: prevHdrHash,
					RandSeed: randSeed,
				},
			},
			SourceChainID: coreDTO.MVX,
		}

		require.Nil(t, sovHdrValidator.IsHeaderConstructionValid(currHdr, prevHdr))

		// Set wrong prev hash as extended header hash
		prevHdrHash, _ = core.CalculateHash(argsHeaderValidator.Marshalizer, argsHeaderValidator.Hasher, prevHdr)
		_ = currHdr.SetPrevHash(prevHdrHash)
		require.Equal(t, process.ErrBlockHashDoesNotMatch, sovHdrValidator.IsHeaderConstructionValid(currHdr, prevHdr))
	})
}
