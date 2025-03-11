package block_test

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	block2 "github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/assert"

	"github.com/multiversx/mx-chain-sovereign-go/process"
	"github.com/multiversx/mx-chain-sovereign-go/process/block"
	"github.com/multiversx/mx-chain-sovereign-go/process/mock"
	"github.com/multiversx/mx-chain-sovereign-go/testscommon/hashingMocks"
)

func TestNewSovereignChainHeaderValidator_ShouldErrNilHeaderValidator(t *testing.T) {
	t.Parallel()

	schv, err := block.NewSovereignChainHeaderValidator(nil)
	assert.Nil(t, schv)
	assert.Equal(t, process.ErrNilHeaderValidator, err)
}

func TestNewSovereignChainHeaderValidator_ShouldWork(t *testing.T) {
	t.Parallel()

	argsHeaderValidator := block.ArgsHeaderValidator{
		Hasher:      &mock.HasherStub{},
		Marshalizer: &mock.MarshalizerMock{},
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
			Hasher:      &mock.HasherStub{},
			Marshalizer: &mock.MarshalizerMock{},
		}
		hv, _ := block.NewHeaderValidator(argsHeaderValidator)
		schv, _ := block.NewSovereignChainHeaderValidator(hv)

		shardHeaderExtended := &block2.ShardHeaderExtended{}
		hash, err := schv.CalculateHeaderHash(shardHeaderExtended)
		assert.Nil(t, hash)
		assert.Equal(t, process.ErrNilHeaderHandler, err)
	})

	t.Run("should work for shard header extended handler", func(t *testing.T) {
		t.Parallel()

		argsHeaderValidator := block.ArgsHeaderValidator{
			Hasher:      &hashingMocks.HasherMock{},
			Marshalizer: &mock.MarshalizerMock{},
		}
		hv, _ := block.NewHeaderValidator(argsHeaderValidator)
		schv, _ := block.NewSovereignChainHeaderValidator(hv)

		shardHeaderExtended := &block2.ShardHeaderExtended{
			Header: &block2.HeaderV2{
				Header: &block2.Header{},
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
			Hasher:      &hashingMocks.HasherMock{},
			Marshalizer: &mock.MarshalizerMock{},
		}
		hv, _ := block.NewHeaderValidator(argsHeaderValidator)
		schv, _ := block.NewSovereignChainHeaderValidator(hv)

		header := &block2.Header{}

		expectedHash, _ := core.CalculateHash(argsHeaderValidator.Marshalizer, argsHeaderValidator.Hasher, header)
		hash, err := schv.CalculateHeaderHash(header)
		assert.Nil(t, err)
		assert.Equal(t, expectedHash, hash)
	})
}
