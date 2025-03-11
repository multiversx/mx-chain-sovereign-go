package processor

import (
	"github.com/multiversx/mx-chain-sovereign-go/dataRetriever"
	"github.com/multiversx/mx-chain-sovereign-go/process"
)

// ArgHdrInterceptorProcessor is the argument for the interceptor processor used for headers (shard, meta and so on)
type ArgHdrInterceptorProcessor struct {
	Headers        dataRetriever.HeadersPool
	BlockBlackList process.TimeCacher
}
