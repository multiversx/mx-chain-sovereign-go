package factory

import (
	"github.com/multiversx/mx-chain-sovereign-go/config"
	"github.com/multiversx/mx-chain-sovereign-go/debug/handler"
)

// NewInterceptorDebuggerFactory will instantiate an InterceptorDebugHandler based on the provided config
func NewInterceptorDebuggerFactory(config config.InterceptorResolverDebugConfig) (InterceptorDebugHandler, error) {
	if !config.Enabled {
		return handler.NewDisabledInterceptorDebugHandler(), nil
	}

	return handler.NewInterceptorDebugHandler(config)
}
