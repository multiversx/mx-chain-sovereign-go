package processor

import (
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"

	"github.com/multiversx/mx-chain-sovereign-go/process"
	"github.com/multiversx/mx-chain-sovereign-go/sharding"
	"github.com/multiversx/mx-chain-sovereign-go/storage"
)

// ArgMiniblockInterceptorProcessor is the argument for the interceptor processor used for miniblocks
type ArgMiniblockInterceptorProcessor struct {
	MiniblockCache   storage.Cacher
	Marshalizer      marshal.Marshalizer
	Hasher           hashing.Hasher
	ShardCoordinator sharding.Coordinator
	WhiteListHandler process.WhiteListHandler
}
