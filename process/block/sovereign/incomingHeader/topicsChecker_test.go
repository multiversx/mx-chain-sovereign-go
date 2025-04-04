package incomingHeader

import (
	"strconv"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/process/block/sovereign/incomingHeader/dto"
)

func TestNewTopicsChecker(t *testing.T) {
	t.Parallel()

	tc := NewTopicsChecker()
	require.False(t, tc.IsInterfaceNil())
}

func TestTopicsChecker_CheckValidity(t *testing.T) {
	t.Parallel()

	tc := NewTopicsChecker()

	createTopics := func(n int) [][]byte {
		topics := make([][]byte, n)
		for i := range topics {
			topics[i] = []byte("topic" + strconv.Itoa(i))
		}
		return topics
	}

	tests := []struct {
		name         string
		topicsCount  int
		transferData *sovereign.TransferData
		expectError  bool
	}{
		{"One topic, should fail", 1, &sovereign.TransferData{}, true},
		{"One topic, should fail (no transfer data)", 1, nil, true},

		{"Two topics, should pass (has transfer data)", 2, &sovereign.TransferData{}, false},
		{"Two topics, should fail (no transfer data)", 2, nil, true},

		{"Three topics, should fail", 3, &sovereign.TransferData{}, true},
		{"Three topics, should fail (no transfer data)", 3, nil, true},

		{"Five topics, should pass", 5, &sovereign.TransferData{}, false},
		{"Five topics, should pass (no transfer data)", 5, nil, false},

		{"Six topics, should fail", 6, &sovereign.TransferData{}, true},
		{"Six topics, should fail (no transfer data)", 6, nil, true},

		{"Eight topics, should pass", 8, &sovereign.TransferData{}, false},
		{"Eight topics, should pass (no transfer data)", 8, nil, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := tc.CheckValidity(createTopics(test.topicsCount), test.transferData)
			if test.expectError {
				require.ErrorContains(t, err, dto.ErrInvalidNumTopicsInEvent.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
