package processing_test

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"

	"github.com/multiversx/mx-chain-go/factory/processing"
	"github.com/multiversx/mx-chain-go/process/mock"
)

func TestManagedProcessComponents_createAPITransactionEvaluator(t *testing.T) {
	t.Parallel()

	shardCoordinatorForShardID2 := mock.NewMultiShardsCoordinatorMock(3)
	shardCoordinatorForShardID2.CurrentShard = 2

	shardCoordinatorForMetachain := mock.NewMultiShardsCoordinatorMock(3)
	shardCoordinatorForMetachain.CurrentShard = core.MetachainShardId

	// no further t.Parallel as these tests are quite heavy (they open netMessengers and other components that start a lot of goroutines)
	t.Run("invalid VMOutputCacher config should error", func(t *testing.T) {
		processArgs := createProcessFactoryArgs(t, shardCoordinatorForShardID2)
		processArgs.Config.VMOutputCacher.Type = "invalid"
		pcf, _ := processing.NewProcessComponentsFactory(processArgs)

		apiTransactionEvaluator, vmContainerFactory, err := pcf.CreateAPITransactionEvaluator()
		assert.NotNil(t, err)
		assert.True(t, check.IfNil(apiTransactionEvaluator))
		assert.True(t, check.IfNil(vmContainerFactory))
		assert.Contains(t, err.Error(), "not supported cache type")
	})
	t.Run("should work for shard", func(t *testing.T) {
		processArgs := createProcessFactoryArgs(t, shardCoordinatorForShardID2)
		pcf, _ := processing.NewProcessComponentsFactory(processArgs)

		apiTransactionEvaluator, vmContainerFactory, err := pcf.CreateAPITransactionEvaluator()
		assert.Nil(t, err)
		assert.False(t, check.IfNil(apiTransactionEvaluator))
		assert.False(t, check.IfNil(vmContainerFactory))
	})
	t.Run("should work for metachain", func(t *testing.T) {
		processArgs := createProcessFactoryArgs(t, shardCoordinatorForMetachain)
		pcf, _ := processing.NewProcessComponentsFactory(processArgs)

		apiTransactionEvaluator, vmContainerFactory, err := pcf.CreateAPITransactionEvaluator()
		assert.Nil(t, err)
		assert.False(t, check.IfNil(apiTransactionEvaluator))
		assert.False(t, check.IfNil(vmContainerFactory))
	})
}
