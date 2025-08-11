package consensus

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/integrationTests"

	"github.com/multiversx/mx-chain-core-go/core/check"
)

func initNodesWithTestSigner(
	numMetaNodes,
	numNodes,
	consensusSize,
	numInvalid uint32,
	roundTime uint64,
	consensusType string,
	// TODO: Marius C: MX-16953 Inject run type comps here
	consensusModel consensus.ConsensusModel,
) (map[uint32][]*integrationTests.TestFullNode, map[string]struct{}) {

	fmt.Println("Step 1. Setup nodes...")

	equivalentProofsActivationEpoch := uint32(0)

	enableEpochsConfig := integrationTests.CreateEnableEpochsConfig()
	enableEpochsConfig.AndromedaEnableEpoch = equivalentProofsActivationEpoch
	isSovereign := false
	if consensusModel == consensus.ConsensusModelV2 {
		// TODO: MARIUS C: MX-16954 Here, have this enabled when we integrate consensus v2 into sovereign consensus
		enableEpochsConfig.AndromedaEnableEpoch = 99999
		enableEpochsConfig.ConsensusModelV2EnableEpoch = 0
		isSovereign = true
	} else {
		enableEpochsConfig.ConsensusModelV2EnableEpoch = 999999
	}

	nodes := integrationTests.CreateNodesWithTestFullNode(
		int(numMetaNodes),
		int(numNodes),
		int(consensusSize),
		roundTime,
		consensusType,
		1,
		enableEpochsConfig,
		false,
		isSovereign,
	)

	time.Sleep(p2pBootstrapDelay)

	invalidNodesAddresses := make(map[string]struct{})

	for shardID := range nodes {
		if numInvalid < numNodes {
			for i := uint32(0); i < numInvalid; i++ {
				ii := numNodes - i - 1
				nodes[shardID][ii].MultiSigner.CreateSignatureShareCalled = func(privateKeyBytes, message []byte) ([]byte, error) {
					var invalidSigShare []byte
					if i%2 == 0 {
						// invalid sig share but with valid format
						invalidSigShare, _ = hex.DecodeString("2ee350b9a821e20df97ba487a80b0d0ffffca7da663185cf6a562edc7c2c71e3ca46ed71b31bccaf53c626b87f2b6e08")
					} else {
						// sig share with invalid size
						invalidSigShare = bytes.Repeat([]byte("a"), 3)
					}
					log.Warn("invalid sig share from ", "pk", nodes[shardID][ii].NodeKeys.MainKey.Pk, "sig", invalidSigShare)

					return invalidSigShare, nil
				}

				invalidNodesAddresses[string(nodes[shardID][ii].OwnAccount.Address)] = struct{}{}
			}
		}
	}

	return nodes, invalidNodesAddresses
}

func TestConsensusWithInvalidSignersConsensusModelV1(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	runConsensusWithInvalidSigners(t, consensus.ConsensusModelV1)
}

// TODO: MARIUS C: MX-16953 Fix this test once we have run type comps integrated in this node
/*
func TestConsensusWithInvalidSignersConsensusModelV2(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	runConsensusWithInvalidSigners(t, consensus.ConsensusModelV2)
}
*/

func runConsensusWithInvalidSigners(t *testing.T, consensusModel consensus.ConsensusModel) {
	numMetaNodes := uint32(4)
	numNodes := uint32(4)
	consensusSize := uint32(4)
	numInvalid := uint32(1)
	roundTime := uint64(1000)

	nodes, invalidNodesAddresses := initNodesWithTestSigner(numMetaNodes, numNodes, consensusSize, numInvalid, roundTime, blsConsensusType, consensusModel)

	defer func() {
		for shardID := range nodes {
			for _, n := range nodes[shardID] {
				_ = n.MainMessenger.Close()
				_ = n.FullArchiveMessenger.Close()
			}
		}
	}()

	// delay for bootstrapping and topic announcement
	fmt.Println("Start consensus...")
	time.Sleep(time.Second)

	for _, nodesList := range nodes {
		for _, n := range nodesList {
			err := startFullConsensusNode(n)
			require.Nil(t, err)
		}
	}

	fmt.Println("Wait for several rounds...")

	time.Sleep(15 * time.Second)

	fmt.Println("Checking shards...")

	expectedNonce := uint64(10)
	for _, nodesList := range nodes {
		for _, n := range nodesList {
			for i := 1; i < len(nodes); i++ {
				_, ok := invalidNodesAddresses[string(n.OwnAccount.Address)]
				if ok {
					continue
				}

				if check.IfNil(n.Node.GetDataComponents().Blockchain().GetCurrentBlockHeader()) {
					assert.Fail(t, fmt.Sprintf("Node with idx %d does not have a current block", i))
				} else {
					assert.GreaterOrEqual(t, n.Node.GetDataComponents().Blockchain().GetCurrentBlockHeader().GetNonce(), expectedNonce)
				}
			}
		}
	}
}
