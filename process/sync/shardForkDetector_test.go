package sync_test

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/block"

	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/process/sync"
	"github.com/multiversx/mx-chain-go/testscommon"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"

	"github.com/stretchr/testify/assert"
)

func TestNewShardForkDetector_NilRoundHandlerShouldErr(t *testing.T) {
	t.Parallel()

	sfd, err := sync.NewShardForkDetector(
		&testscommon.LoggerStub{},
		nil,
		&testscommon.TimeCacheStub{},
		&mock.BlockTrackerMock{},
		0,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&dataRetrieverMock.ProofsPoolMock{},
	)
	assert.True(t, check.IfNil(sfd))
	assert.Equal(t, process.ErrNilRoundHandler, err)
}

func TestNewShardForkDetector_NilBlackListShouldErr(t *testing.T) {
	t.Parallel()

	sfd, err := sync.NewShardForkDetector(
		&testscommon.LoggerStub{},
		&mock.RoundHandlerMock{},
		nil,
		&mock.BlockTrackerMock{},
		0,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&dataRetrieverMock.ProofsPoolMock{},
	)
	assert.True(t, check.IfNil(sfd))
	assert.Equal(t, process.ErrNilBlackListCacher, err)
}

func TestNewShardForkDetector_NilBlockTrackerShouldErr(t *testing.T) {
	t.Parallel()

	sfd, err := sync.NewShardForkDetector(
		&testscommon.LoggerStub{},
		&mock.RoundHandlerMock{},
		&testscommon.TimeCacheStub{},
		nil,
		0,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&dataRetrieverMock.ProofsPoolMock{},
	)
	assert.True(t, check.IfNil(sfd))
	assert.Equal(t, process.ErrNilBlockTracker, err)
}

func TestNewShardForkDetector_NilEnableEpochsHandlerShouldErr(t *testing.T) {
	t.Parallel()

	sfd, err := sync.NewShardForkDetector(
		&testscommon.LoggerStub{},
		&mock.RoundHandlerMock{},
		&testscommon.TimeCacheStub{},
		&mock.BlockTrackerMock{},
		0,
		nil,
		&dataRetrieverMock.ProofsPoolMock{},
	)
	assert.True(t, check.IfNil(sfd))
	assert.Equal(t, process.ErrNilEnableEpochsHandler, err)
}

func TestNewShardForkDetector_NilProofsPoolShouldErr(t *testing.T) {
	t.Parallel()

	sfd, err := sync.NewShardForkDetector(
		&testscommon.LoggerStub{},
		&mock.RoundHandlerMock{},
		&testscommon.TimeCacheStub{},
		&mock.BlockTrackerMock{},
		0,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		nil,
	)
	assert.True(t, check.IfNil(sfd))
	assert.Equal(t, process.ErrNilProofsPool, err)
}

func TestNewShardForkDetector_OkParamsShouldWork(t *testing.T) {
	t.Parallel()

	sfd, err := sync.NewShardForkDetector(
		&testscommon.LoggerStub{},
		&mock.RoundHandlerMock{},
		&testscommon.TimeCacheStub{},
		&mock.BlockTrackerMock{},
		0,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&dataRetrieverMock.ProofsPoolMock{},
	)
	assert.Nil(t, err)
	assert.False(t, check.IfNil(sfd))

	assert.Equal(t, uint64(0), sfd.LastCheckpointNonce())
	assert.Equal(t, uint64(0), sfd.LastCheckpointRound())
	assert.Equal(t, uint64(0), sfd.FinalCheckpointNonce())
	assert.Equal(t, uint64(0), sfd.FinalCheckpointRound())
}

func TestShardForkDetector_AddHeaderNilHeaderShouldErr(t *testing.T) {
	t.Parallel()

	roundHandlerMock := &mock.RoundHandlerMock{RoundIndex: 100}
	bfd, _ := sync.NewShardForkDetector(
		&testscommon.LoggerStub{},
		roundHandlerMock,
		&testscommon.TimeCacheStub{},
		&mock.BlockTrackerMock{},
		0,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&dataRetrieverMock.ProofsPoolMock{},
	)
	err := bfd.AddHeader(nil, make([]byte, 0), process.BHProcessed, nil, nil)
	assert.Equal(t, sync.ErrNilHeader, err)
}

func TestShardForkDetector_AddHeaderNilHashShouldErr(t *testing.T) {
	t.Parallel()

	roundHandlerMock := &mock.RoundHandlerMock{RoundIndex: 100}
	bfd, _ := sync.NewShardForkDetector(
		&testscommon.LoggerStub{},
		roundHandlerMock,
		&testscommon.TimeCacheStub{},
		&mock.BlockTrackerMock{},
		0,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&dataRetrieverMock.ProofsPoolMock{},
	)
	err := bfd.AddHeader(&block.Header{}, nil, process.BHProcessed, nil, nil)
	assert.Equal(t, sync.ErrNilHash, err)
}

func TestShardForkDetector_AddHeaderNotPresentShouldWork(t *testing.T) {
	t.Parallel()

	hdr := &block.Header{Nonce: 1, Round: 1, PubKeysBitmap: []byte("X")}
	hash := make([]byte, 0)
	roundHandlerMock := &mock.RoundHandlerMock{RoundIndex: 1}
	bfd, _ := sync.NewShardForkDetector(
		&testscommon.LoggerStub{},
		roundHandlerMock,
		&testscommon.TimeCacheStub{},
		&mock.BlockTrackerMock{},
		0,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&dataRetrieverMock.ProofsPoolMock{},
	)
	err := bfd.AddHeader(hdr, hash, process.BHProcessed, nil, nil)
	assert.Nil(t, err)

	hInfos := bfd.GetHeaders(1)
	assert.Equal(t, 1, len(hInfos))
	assert.Equal(t, hash, hInfos[0].Hash())
}

func TestShardForkDetector_AddHeaderPresentShouldAppend(t *testing.T) {
	t.Parallel()

	hdr1 := &block.Header{Nonce: 1, Round: 1, PubKeysBitmap: []byte("X")}
	hash1 := []byte("hash1")
	hdr2 := &block.Header{Nonce: 1, Round: 1, PubKeysBitmap: []byte("X")}
	hash2 := []byte("hash2")
	roundHandlerMock := &mock.RoundHandlerMock{RoundIndex: 1}
	bfd, _ := sync.NewShardForkDetector(
		&testscommon.LoggerStub{},
		roundHandlerMock,
		&testscommon.TimeCacheStub{},
		&mock.BlockTrackerMock{},
		0,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&dataRetrieverMock.ProofsPoolMock{},
	)
	_ = bfd.AddHeader(hdr1, hash1, process.BHProcessed, nil, nil)
	err := bfd.AddHeader(hdr2, hash2, process.BHProcessed, nil, nil)
	assert.Nil(t, err)

	hInfos := bfd.GetHeaders(1)
	assert.Equal(t, 2, len(hInfos))
	assert.Equal(t, hash1, hInfos[0].Hash())
	assert.Equal(t, hash2, hInfos[1].Hash())
}

func TestShardForkDetector_AddHeaderWithProcessedBlockShouldSetCheckpoint(t *testing.T) {
	t.Parallel()

	hdr1 := &block.Header{Nonce: 69, Round: 72, PubKeysBitmap: []byte("X")}
	hash1 := []byte("hash1")
	roundHandlerMock := &mock.RoundHandlerMock{RoundIndex: 73}
	bfd, _ := sync.NewShardForkDetector(
		&testscommon.LoggerStub{},
		roundHandlerMock,
		&testscommon.TimeCacheStub{},
		&mock.BlockTrackerMock{},
		0,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&dataRetrieverMock.ProofsPoolMock{},
	)
	_ = bfd.AddHeader(hdr1, hash1, process.BHProcessed, nil, nil)
	assert.Equal(t, hdr1.Nonce, bfd.LastCheckpointNonce())
}

func TestShardForkDetector_AddHeaderPresentShouldNotRewriteState(t *testing.T) {
	t.Parallel()

	hdr1 := &block.Header{Nonce: 1, Round: 1, PubKeysBitmap: []byte("X")}
	hash := []byte("hash1")
	hdr2 := &block.Header{Nonce: 1, Round: 1, PubKeysBitmap: []byte("X")}
	roundHandlerMock := &mock.RoundHandlerMock{RoundIndex: 1}
	bfd, _ := sync.NewShardForkDetector(
		&testscommon.LoggerStub{},
		roundHandlerMock,
		&testscommon.TimeCacheStub{},
		&mock.BlockTrackerMock{},
		0,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&dataRetrieverMock.ProofsPoolMock{},
	)
	_ = bfd.AddHeader(hdr1, hash, process.BHReceived, nil, nil)
	err := bfd.AddHeader(hdr2, hash, process.BHProcessed, nil, nil)
	assert.Nil(t, err)

	hInfos := bfd.GetHeaders(1)
	assert.Equal(t, 2, len(hInfos))
	assert.Equal(t, hash, hInfos[0].Hash())
	assert.Equal(t, process.BHReceived, hInfos[0].GetBlockHeaderState())
	assert.Equal(t, process.BHProcessed, hInfos[1].GetBlockHeaderState())
}

func TestShardForkDetector_AddHeaderHigherNonceThanRoundShouldErr(t *testing.T) {
	t.Parallel()

	roundHandlerMock := &mock.RoundHandlerMock{RoundIndex: 100}
	bfd, _ := sync.NewShardForkDetector(
		&testscommon.LoggerStub{},
		roundHandlerMock,
		&testscommon.TimeCacheStub{},
		&mock.BlockTrackerMock{},
		0,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&dataRetrieverMock.ProofsPoolMock{},
	)
	err := bfd.AddHeader(
		&block.Header{Nonce: 1, Round: 0, PubKeysBitmap: []byte("X")}, []byte("hash1"), process.BHProcessed, nil, nil)
	assert.Equal(t, sync.ErrHigherNonceInBlock, err)
}
