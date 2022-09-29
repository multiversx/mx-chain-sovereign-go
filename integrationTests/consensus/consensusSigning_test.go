package consensus

import (
	"fmt"
	"sync"
	"testing"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/stretchr/testify/assert"
)

// TODO: refactor to use nodes from multiple shards
func initNodesWithTestSigner(
	numMetaNodes,
	numNodes,
	consensusSize,
	numInvalid uint32,
	roundTime uint64,
	consensusType string,
) map[uint32][]*integrationTests.TestConsensusNode {

	fmt.Println("Step 1. Setup nodes...")

	nodes := integrationTests.CreateNodesWithTestConsensusNode(
		int(numMetaNodes),
		int(numNodes),
		int(consensusSize),
		roundTime,
		consensusType,
	)

	for shardID, nodesList := range nodes {
		displayAndStartNodes(shardID, nodesList)
	}

	time.Sleep(p2pBootstrapDelay)

	for shardID := range nodes {
		if numInvalid < numNodes {
			for i := uint32(0); i < numInvalid; i++ {
				iCopy := i
				nodes[shardID][i].MultiSigner.CreateSignatureShareCalled = func(privateKeyBytes, message []byte) ([]byte, error) {
					fmt.Println("invalid sig share from ",
						getPkEncoded(nodes[shardID][iCopy].NodeKeys.Pk),
					)
					return []byte("invalid sig share"), nil
				}
			}
		}
	}

	return nodes
}

func TestConsensusWithInvalidSigners(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	_ = logger.SetLogLevel("*:DEBUG")

	numMetaNodes := uint32(4)
	numNodes := uint32(4)
	consensusSize := uint32(4)
	numInvalid := uint32(0)
	roundTime := uint64(1000)
	numCommBlock := uint64(8)

	nodes := initNodesWithTestSigner(numMetaNodes, numNodes, consensusSize, numInvalid, roundTime, blsConsensusType)

	defer func() {
		for shardID := range nodes {
			for _, n := range nodes[shardID] {
				_ = n.Messenger.Close()
			}
		}
	}()

	// delay for bootstrapping and topic announcement
	fmt.Println("Start consensus...")
	time.Sleep(time.Second)

	for shardID := range nodes {
		mutex := &sync.Mutex{}
		nonceForRoundMap := make(map[uint64]uint64)
		totalCalled := 0

		err := startNodesWithCommitBlock(nodes[shardID], mutex, nonceForRoundMap, &totalCalled)
		assert.Nil(t, err)

		chDone := make(chan bool)
		go checkBlockProposedEveryRound(numCommBlock, nonceForRoundMap, mutex, chDone, t)

		extraTime := uint64(2)
		endTime := time.Duration(roundTime) * time.Duration(numCommBlock+extraTime) * time.Millisecond
		select {
		case <-chDone:
		case <-time.After(endTime):
			mutex.Lock()
			fmt.Println("currently saved nonces for rounds: \n", nonceForRoundMap)
			assert.Fail(t, "consensus too slow, not working.")
			mutex.Unlock()
			return
		}
	}
}
