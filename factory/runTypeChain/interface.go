package runTypeChain

type runTypeChainComponentsCreator interface {
	Create() *runTypeChainComponents
	IsInterfaceNil() bool
}
