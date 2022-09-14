package broadcast_test

import (
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/consensus/broadcast"
	"github.com/ElrondNetwork/elrond-go/consensus/mock"
	"github.com/ElrondNetwork/elrond-go/consensus/spos"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/hashingMocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createDefaultMetaChainArgs() broadcast.MetaChainMessengerArgs {
	marshalizerMock := &mock.MarshalizerMock{}
	messengerMock := &mock.MessengerStub{}
	privateKeyMock := &mock.PrivateKeyMock{}
	shardCoordinatorMock := &mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{}
	hasher := &hashingMocks.HasherMock{}
	headersSubscriber := &mock.HeadersCacherStub{}
	interceptorsContainer := createInterceptorContainer()
	peerSigHandler := &mock.PeerSignatureHandler{Signer: singleSignerMock}
	alarmScheduler := &mock.AlarmSchedulerStub{}

	return broadcast.MetaChainMessengerArgs{
		CommonMessengerArgs: broadcast.CommonMessengerArgs{
			Marshalizer:                marshalizerMock,
			Hasher:                     hasher,
			Messenger:                  messengerMock,
			PrivateKey:                 privateKeyMock,
			ShardCoordinator:           shardCoordinatorMock,
			PeerSignatureHandler:       peerSigHandler,
			HeadersSubscriber:          headersSubscriber,
			InterceptorsContainer:      interceptorsContainer,
			MaxValidatorDelayCacheSize: 2,
			MaxDelayCacheSize:          2,
			AlarmScheduler:             alarmScheduler,
			KeysHolder:                 &testscommon.KeysHolderStub{},
		},
	}
}

func TestMetaChainMessenger_NewMetaChainMessengerNilMarshalizerShouldFail(t *testing.T) {
	args := createDefaultMetaChainArgs()
	args.Marshalizer = nil
	mcm, err := broadcast.NewMetaChainMessenger(args)

	assert.Nil(t, mcm)
	assert.Equal(t, spos.ErrNilMarshalizer, err)
}

func TestMetaChainMessenger_NewMetaChainMessengerNilMessengerShouldFail(t *testing.T) {
	args := createDefaultMetaChainArgs()
	args.Messenger = nil
	mcm, err := broadcast.NewMetaChainMessenger(args)

	assert.Nil(t, mcm)
	assert.Equal(t, spos.ErrNilMessenger, err)
}

func TestMetaChainMessenger_NewMetaChainMessengerNilPrivateKeyShouldFail(t *testing.T) {
	args := createDefaultMetaChainArgs()
	args.PrivateKey = nil
	mcm, err := broadcast.NewMetaChainMessenger(args)

	assert.Nil(t, mcm)
	assert.Equal(t, spos.ErrNilPrivateKey, err)
}

func TestMetaChainMessenger_NewMetaChainMessengerNilShardCoordinatorShouldFail(t *testing.T) {
	args := createDefaultMetaChainArgs()
	args.ShardCoordinator = nil
	mcm, err := broadcast.NewMetaChainMessenger(args)

	assert.Nil(t, mcm)
	assert.Equal(t, spos.ErrNilShardCoordinator, err)
}

func TestMetaChainMessenger_NewMetaChainMessengerNilPeerSignatureHandlerShouldFail(t *testing.T) {
	args := createDefaultMetaChainArgs()
	args.PeerSignatureHandler = nil
	mcm, err := broadcast.NewMetaChainMessenger(args)

	assert.Nil(t, mcm)
	assert.Equal(t, spos.ErrNilPeerSignatureHandler, err)
}

func TestMetaChainMessenger_NilKeysHolderShouldError(t *testing.T) {
	args := createDefaultMetaChainArgs()
	args.KeysHolder = nil
	mcm, err := broadcast.NewMetaChainMessenger(args)

	assert.Nil(t, mcm)
	assert.Equal(t, spos.ErrNilKeysHolder, err)
}

func TestMetaChainMessenger_NewMetaChainMessengerShouldWork(t *testing.T) {
	args := createDefaultMetaChainArgs()
	mcm, err := broadcast.NewMetaChainMessenger(args)

	assert.NotNil(t, mcm)
	assert.Equal(t, nil, err)
	assert.False(t, mcm.IsInterfaceNil())
}

func TestMetaChainMessenger_BroadcastBlockShouldErrNilMetaHeader(t *testing.T) {
	args := createDefaultMetaChainArgs()
	mcm, _ := broadcast.NewMetaChainMessenger(args)

	err := mcm.BroadcastBlock(newTestBlockBody(), nil)
	assert.Equal(t, spos.ErrNilMetaHeader, err)
}

func TestMetaChainMessenger_BroadcastBlockShouldErrMockMarshalizer(t *testing.T) {
	marshalizer := &mock.MarshalizerMock{
		Fail: true,
	}
	args := createDefaultMetaChainArgs()
	args.Marshalizer = marshalizer
	mcm, _ := broadcast.NewMetaChainMessenger(args)

	err := mcm.BroadcastBlock(newTestBlockBody(), &block.MetaBlock{})
	assert.Equal(t, mock.ErrMockMarshalizer, err)
}

func TestMetaChainMessenger_BroadcastBlockShouldWork(t *testing.T) {
	messenger := &mock.MessengerStub{
		BroadcastCalled: func(topic string, buff []byte) {
		},
	}
	args := createDefaultMetaChainArgs()
	args.Messenger = messenger
	mcm, _ := broadcast.NewMetaChainMessenger(args)

	err := mcm.BroadcastBlock(newTestBlockBody(), &block.MetaBlock{})
	assert.Nil(t, err)
}

func TestMetaChainMessenger_BroadcastMiniBlocksShouldWork(t *testing.T) {
	args := createDefaultMetaChainArgs()
	mcm, _ := broadcast.NewMetaChainMessenger(args)

	err := mcm.BroadcastMiniBlocks(nil, []byte("pk bytes"))
	assert.Nil(t, err)
}

func TestMetaChainMessenger_BroadcastTransactionsShouldWork(t *testing.T) {
	args := createDefaultMetaChainArgs()
	mcm, _ := broadcast.NewMetaChainMessenger(args)

	err := mcm.BroadcastTransactions(nil, []byte("pk bytes"))
	assert.Nil(t, err)
}

func TestMetaChainMessenger_BroadcastHeaderNilHeaderShouldErr(t *testing.T) {
	args := createDefaultMetaChainArgs()
	mcm, _ := broadcast.NewMetaChainMessenger(args)

	err := mcm.BroadcastHeader(nil, []byte("pk bytes"))
	assert.Equal(t, spos.ErrNilHeader, err)
}

func TestMetaChainMessenger_BroadcastHeaderOkHeaderShouldWork(t *testing.T) {
	channelBroadcastCalled := make(chan bool, 1)
	channelBroadcastUsingPrivateKeyCalled := make(chan bool, 1)

	messenger := &mock.MessengerStub{
		BroadcastCalled: func(topic string, buff []byte) {
			channelBroadcastCalled <- true
		},
		BroadcastUsingPrivateKeyCalled: func(topic string, buff []byte, pid core.PeerID, skBytes []byte) {
			channelBroadcastUsingPrivateKeyCalled <- true
		},
	}
	args := createDefaultMetaChainArgs()
	args.Messenger = messenger
	mcm, _ := broadcast.NewMetaChainMessenger(args)

	hdr := block.Header{
		Nonce: 10,
	}

	t.Run("original public key of the node", func(t *testing.T) {
		pkBytes, _ := args.PrivateKey.GeneratePublic().ToByteArray()
		err := mcm.BroadcastHeader(&hdr, pkBytes)
		assert.Nil(t, err)

		wasCalled := false
		select {
		case <-channelBroadcastCalled:
			wasCalled = true
		case <-time.After(time.Millisecond * 100):
		}

		assert.Nil(t, err)
		assert.True(t, wasCalled)
	})
	t.Run("managed key", func(t *testing.T) {
		err := mcm.BroadcastHeader(&hdr, []byte("managed key"))
		assert.Nil(t, err)

		wasCalled := false
		select {
		case <-channelBroadcastUsingPrivateKeyCalled:
			wasCalled = true
		case <-time.After(time.Millisecond * 100):
		}

		assert.Nil(t, err)
		assert.True(t, wasCalled)
	})

}

func TestMetaChainMessenger_BroadcastBlockDataLeader(t *testing.T) {
	countersBroadcast := make(map[string]int)
	mutCounters := &sync.Mutex{}

	messengerMock := &mock.MessengerStub{
		BroadcastCalled: func(topic string, buff []byte) {
			mutCounters.Lock()
			countersBroadcast[broadcastMethodPrefix+topic]++
			mutCounters.Unlock()
		},
		BroadcastUsingPrivateKeyCalled: func(topic string, buff []byte, pid core.PeerID, skBytes []byte) {
			mutCounters.Lock()
			countersBroadcast[broadcastUsingPrivateKeyCalledMethodPrefix+topic]++
			mutCounters.Unlock()
		},
	}

	args := createDefaultMetaChainArgs()
	args.Messenger = messengerMock
	mcm, _ := broadcast.NewMetaChainMessenger(args)

	miniBlocks := map[uint32][]byte{0: []byte("mbs data1"), 1: []byte("mbs data2")}
	transactions := map[string][][]byte{"topic1": {[]byte("txdata1"), []byte("txdata2")}, "topic2": {[]byte("txdata3")}}

	t.Run("original public key of the node", func(t *testing.T) {
		mutCounters.Lock()
		countersBroadcast = make(map[string]int)
		mutCounters.Unlock()

		pkBytes, _ := args.PrivateKey.GeneratePublic().ToByteArray()
		err := mcm.BroadcastBlockDataLeader(nil, miniBlocks, transactions, pkBytes)
		require.Nil(t, err)
		sleepTime := common.ExtraDelayBetweenBroadcastMbsAndTxs +
			common.ExtraDelayForBroadcastBlockInfo +
			time.Millisecond*100
		time.Sleep(sleepTime)

		mutCounters.Lock()
		defer mutCounters.Unlock()

		numBroadcast := countersBroadcast[broadcastMethodPrefix+"txBlockBodies_0"]
		numBroadcast += countersBroadcast[broadcastMethodPrefix+"txBlockBodies_0_1"]
		assert.Equal(t, len(miniBlocks), numBroadcast)

		numBroadcast = countersBroadcast[broadcastMethodPrefix+"topic1"]
		numBroadcast += countersBroadcast[broadcastMethodPrefix+"topic2"]
		assert.Equal(t, len(transactions), numBroadcast)
	})
	t.Run("managed key", func(t *testing.T) {
		mutCounters.Lock()
		countersBroadcast = make(map[string]int)
		mutCounters.Unlock()

		err := mcm.BroadcastBlockDataLeader(nil, miniBlocks, transactions, []byte("pk bytes"))
		require.Nil(t, err)
		sleepTime := common.ExtraDelayBetweenBroadcastMbsAndTxs +
			common.ExtraDelayForBroadcastBlockInfo +
			time.Millisecond*100
		time.Sleep(sleepTime)

		mutCounters.Lock()
		defer mutCounters.Unlock()

		numBroadcast := countersBroadcast[broadcastUsingPrivateKeyCalledMethodPrefix+"txBlockBodies_0"]
		numBroadcast += countersBroadcast[broadcastUsingPrivateKeyCalledMethodPrefix+"txBlockBodies_0_1"]
		assert.Equal(t, len(miniBlocks), numBroadcast)

		numBroadcast = countersBroadcast[broadcastUsingPrivateKeyCalledMethodPrefix+"topic1"]
		numBroadcast += countersBroadcast[broadcastUsingPrivateKeyCalledMethodPrefix+"topic2"]
		assert.Equal(t, len(transactions), numBroadcast)
	})
}
