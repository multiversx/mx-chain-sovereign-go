package blockchain

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"

	"github.com/multiversx/mx-chain-sovereign-go/testscommon"
)

func TestNewBootstrapBlockchain(t *testing.T) {
	t.Parallel()

	blockchain := NewBootstrapBlockchain()
	assert.False(t, check.IfNil(blockchain))
	providedHeaderHandler := &testscommon.HeaderHandlerStub{}
	assert.Nil(t, blockchain.SetCurrentBlockHeaderAndRootHash(providedHeaderHandler, nil))
	assert.Equal(t, providedHeaderHandler, blockchain.GetCurrentBlockHeader())
}
