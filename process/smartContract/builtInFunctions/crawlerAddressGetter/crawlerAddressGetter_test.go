package crawlerAddressGetter

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/assert"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/sharding"
)

func TestCreateBuiltInFunctionContainerGetAllowedAddress_Errors(t *testing.T) {
	t.Parallel()

	crawlerGetter := NewCrawlerAddressGetter()
	t.Run("nil shardCoordinator", func(t *testing.T) {
		t.Parallel()

		_, addresses := GetMockShardCoordinatorAndAddresses(1)
		allowedAddressForShard, err := crawlerGetter.GetAllowedAddress(nil, addresses)
		assert.Nil(t, allowedAddressForShard)
		assert.Equal(t, process.ErrNilShardCoordinator, err)
	})
	t.Run("nil addresses", func(t *testing.T) {
		t.Parallel()

		shardCoordinator, _ := GetMockShardCoordinatorAndAddresses(1)
		allowedAddressForShard, err := crawlerGetter.GetAllowedAddress(shardCoordinator, nil)
		assert.Nil(t, allowedAddressForShard)
		assert.True(t, errors.Is(err, process.ErrNilCrawlerAllowedAddress))
		assert.True(t, strings.Contains(err.Error(), "provided count is 0"))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		shardCoordinator, addresses := GetMockShardCoordinatorAndAddresses(1)
		allowedAddressForShard, err := crawlerGetter.GetAllowedAddress(shardCoordinator, addresses)
		assert.NotNil(t, allowedAddressForShard)
		assert.Nil(t, err)
	})
	t.Run("existing address for shard 1", func(t *testing.T) {
		t.Parallel()

		shardCoordinator, _ := GetMockShardCoordinatorAndAddresses(1)
		addresses := [][]byte{
			bytes.Repeat([]byte{1}, 32), // shard 1
		}

		allowedAddressForShard, err := crawlerGetter.GetAllowedAddress(shardCoordinator, addresses)
		assert.NotNil(t, allowedAddressForShard)
		assert.Nil(t, err)
	})
	t.Run("no address for shard 1", func(t *testing.T) {
		t.Parallel()

		shardCoordinator, _ := GetMockShardCoordinatorAndAddresses(1)
		addresses := [][]byte{
			bytes.Repeat([]byte{2}, 32), // shard 2
			bytes.Repeat([]byte{3}, 32)} // shard 0

		allowedAddressForShard, err := crawlerGetter.GetAllowedAddress(shardCoordinator, addresses)
		assert.Nil(t, allowedAddressForShard)
		assert.True(t, errors.Is(err, process.ErrNilCrawlerAllowedAddress))
		expectedMessage := fmt.Sprintf("for shard %d, provided count is %d", shardCoordinator.SelfId(), len(addresses))
		assert.True(t, strings.Contains(err.Error(), expectedMessage))
	})
	t.Run("metachain takes core.SystemAccountAddress", func(t *testing.T) {
		t.Parallel()

		shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
		shardCoordinator.CurrentShard = common.MetachainShardId
		addresses := [][]byte{
			bytes.Repeat([]byte{20}, 32), // bigger addresss
			bytes.Repeat([]byte{3}, 32)}  // smaller address

		allowedAddressForShard, err := crawlerGetter.GetAllowedAddress(shardCoordinator, addresses)
		assert.Nil(t, err)
		assert.Equal(t, core.SystemAccountAddress, allowedAddressForShard)
	})
	t.Run("every shard gets an allowedCrawlerAddress", func(t *testing.T) {
		t.Parallel()

		nrShards := uint32(3)
		shardCoordinator := mock.NewMultiShardsCoordinatorMock(nrShards)
		addresses := make([][]byte, nrShards)
		for i := byte(0); i < byte(nrShards); i++ {
			addresses[i] = bytes.Repeat([]byte{i + byte(nrShards)}, 32)
		}
		shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
			lastByte := address[len(address)-1]
			return uint32(lastByte) % nrShards
		}

		currentShardId := uint32(0)
		shardCoordinator.CurrentShard = currentShardId
		allowedAddressForShard, _ := crawlerGetter.GetAllowedAddress(shardCoordinator, addresses)
		assert.Equal(t, addresses[currentShardId], allowedAddressForShard)

		currentShardId = uint32(1)
		shardCoordinator.CurrentShard = currentShardId
		allowedAddressForShard, _ = crawlerGetter.GetAllowedAddress(shardCoordinator, addresses)
		assert.Equal(t, addresses[currentShardId], allowedAddressForShard)

		currentShardId = uint32(2)
		shardCoordinator.CurrentShard = currentShardId
		allowedAddressForShard, _ = crawlerGetter.GetAllowedAddress(shardCoordinator, addresses)
		assert.Equal(t, addresses[currentShardId], allowedAddressForShard)

		currentShardId = common.MetachainShardId
		shardCoordinator.CurrentShard = currentShardId
		allowedAddressForShard, _ = crawlerGetter.GetAllowedAddress(shardCoordinator, addresses)
		assert.Equal(t, core.SystemAccountAddress, allowedAddressForShard)
	})

}

func GetMockShardCoordinatorAndAddresses(currentShardId uint32) (sharding.Coordinator, [][]byte) {
	nrShards := uint32(3)
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(nrShards)
	addresses := make([][]byte, nrShards)
	for i := byte(0); i < byte(nrShards); i++ {
		addresses[i] = bytes.Repeat([]byte{i + byte(nrShards)}, 32)
	}
	shardCoordinator.CurrentShard = currentShardId
	shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
		lastByte := address[len(address)-1]
		return uint32(lastByte) % nrShards
	}

	return shardCoordinator, addresses
}
