package consensus_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/multiversx/mx-chain-sovereign-go/consensus"
)

func TestConsensusMessage_NewConsensusMessageShouldWork(t *testing.T) {
	t.Parallel()

	cnsMsg := consensus.NewConsensusMessage(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		-1,
		0,
		[]byte("chain ID"),
		nil,
		nil,
		nil,
		"pid",
		nil,
		nil,
	)

	assert.NotNil(t, cnsMsg)
}
