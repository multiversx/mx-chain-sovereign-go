package notifier

import (
	"github.com/multiversx/mx-chain-core-go/core/sovereign"
)

// SovereignNotifierBootstrapper defines a sovereign notifier bootstrapper
type SovereignNotifierBootstrapper interface {
	Start()
	Close() error
	IsInterfaceNil() bool
}

// SovereignNotifier defines what a sovereign notifier should do
type SovereignNotifier interface {
	RegisterHandler(handler sovereign.IncomingHeaderSubscriber) error
	IsInterfaceNil() bool
}
