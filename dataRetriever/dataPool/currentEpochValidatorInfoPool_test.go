package dataPool

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-sovereign-go/dataRetriever"
	"github.com/multiversx/mx-chain-sovereign-go/state"
)

func TestCurrentEpochValidatorInfoPool_AddGetCleanTx(t *testing.T) {
	t.Parallel()

	validatorInfoHash := []byte("hash")
	validatorInfo := &state.ShardValidatorInfo{}
	currentValidatorInfoPool := NewCurrentEpochValidatorInfoPool()
	require.False(t, currentValidatorInfoPool.IsInterfaceNil())

	currentValidatorInfoPool.AddValidatorInfo(validatorInfoHash, validatorInfo)
	currentValidatorInfoPool.AddValidatorInfo(validatorInfoHash, nil)

	validatorInfoFromPool, err := currentValidatorInfoPool.GetValidatorInfo([]byte("wrong hash"))
	require.Nil(t, validatorInfoFromPool)
	require.Equal(t, dataRetriever.ErrValidatorInfoNotFoundInEpochPool, err)

	validatorInfoFromPool, err = currentValidatorInfoPool.GetValidatorInfo(validatorInfoHash)
	require.Nil(t, err)
	require.Equal(t, validatorInfo, validatorInfoFromPool)

	currentValidatorInfoPool.Clean()
	validatorInfoFromPool, err = currentValidatorInfoPool.GetValidatorInfo(validatorInfoHash)
	require.Nil(t, validatorInfoFromPool)
	require.Equal(t, dataRetriever.ErrValidatorInfoNotFoundInEpochPool, err)
}
