package systemSmartContracts

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/vm/systemSmartContracts/clob"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

func TestNewClobExecutor(t *testing.T) {
	ce := NewClobExecutor()
	require.NotNil(t, ce)
}

func TestClobExecutor_Execute(t *testing.T) {
	ce := NewClobExecutor()
	storage := &testscommon.AccountStorageMock{}

	// Test with an invalid function name
	input := &vmcommon.ContractCallInput{
		Function: "invalidFunction",
	}
	output, err := ce.Execute(input, storage)
	require.NoError(t, err)
	require.Equal(t, vmcommon.InternalError, output.ReturnCode)
	require.Contains(t, string(output.ReturnMessage), "invalid function name")

	// Test with a valid function name but not enough arguments
	input = &vmcommon.ContractCallInput{
		Function: clob.ProcessOrderEndpoint,
	}
	output, err = ce.Execute(input, storage)
	require.NoError(t, err)
	require.Equal(t, vmcommon.InternalError, output.ReturnCode)
	require.Contains(t, string(output.ReturnMessage), "invalid number of arguments")
}

func createTestSCInput(funcName string, args ...[]byte) *vmcommon.ContractCallInput {
	return &vmcommon.ContractCallInput{
		Caller:         []byte("caller"),
		Value:          nil,
		Function:       funcName,
		Arguments:      args,
		GasLimit:       1000000,
		GasPrice:       1,
		BlockHeader:    &block.Header{},
		BlockTimestamp: 123456,
		TxHash:         []byte("txHash"),
		ShardID:        sharding.MetachainShardId,
		TxHeader:       &transaction.Header{},
	}
}