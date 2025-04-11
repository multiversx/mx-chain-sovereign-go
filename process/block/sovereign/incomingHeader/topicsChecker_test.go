package incomingHeader

import (
	"strconv"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/process/block/sovereign/incomingHeader/dto"
)

func createTopics(topicID string, n int) [][]byte {
	topics := make([][]byte, n)
	topics[0] = []byte(topicID)
	for i := range topics[1:] {
		topics[i] = []byte("topic" + strconv.Itoa(i))
	}
	return topics
}

func TestNewTopicsChecker(t *testing.T) {
	t.Parallel()

	tc := NewTopicsChecker()
	require.False(t, tc.IsInterfaceNil())
}

func TestTopicsChecker_checkDepositTokensValidity(t *testing.T) {
	t.Parallel()

	tc := NewTopicsChecker()

	tests := []struct {
		name        string
		topicsCount int
		expectError bool
	}{
		{"One topic, should fail", 1, true},
		{"Two topics, should fail", 2, true},
		{"Five topics, should pass", 5, false},
		{"Six topics, should fail", 6, true},
		{"Eight topics, should pass", 8, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := tc.checkDepositTokensValidity(createTopics(dto.TopicIDDepositIncomingTransfer, test.topicsCount))
			if test.expectError {
				require.ErrorContains(t, err, dto.ErrInvalidNumTopicsInEvent.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestTopicsChecker_checkScCallValidity(t *testing.T) {
	t.Parallel()

	tc := NewTopicsChecker()

	tests := []struct {
		name         string
		topicsCount  int
		transferData *sovereign.TransferData
		expectError  bool
	}{
		{"One topic, should fail", 1, &sovereign.TransferData{}, true},
		{"One topic, should fail (no transfer data)", 1, nil, true},

		{"Two topics, should pass", 2, &sovereign.TransferData{}, false},
		{"Two topics, should fail (no transfer data)", 2, nil, true},

		{"Three topics, should fail", 3, &sovereign.TransferData{}, true},
		{"Three topics, should fail (no transfer data)", 3, nil, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := tc.checkScCallValidity(createTopics(dto.TopicIDDepositIncomingTransfer, test.topicsCount), test.transferData)
			if test.expectError {
				require.ErrorContains(t, err, dto.ErrInvalidNumTopicsInEvent.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestTopicsChecker_CheckValidity(t *testing.T) {
	t.Parallel()

	tc := NewTopicsChecker()

	topics := make([][]byte, 0)
	topics = append(topics, []byte("topic"))

	err := tc.CheckValidity(topics, nil)
	require.Error(t, err, dto.ErrInvalidIncomingTopicIdentifier)
}
