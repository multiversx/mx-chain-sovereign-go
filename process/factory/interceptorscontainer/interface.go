package interceptorscontainer

import "github.com/multiversx/mx-chain-sovereign-go/process"

// InterceptorsContainerFactoryCreator defines an interceptor container factory creator
type InterceptorsContainerFactoryCreator interface {
	CreateInterceptorsContainerFactory(args CommonInterceptorsContainerFactoryArgs) (process.InterceptorsContainerFactory, error)
	IsInterfaceNil() bool
}
