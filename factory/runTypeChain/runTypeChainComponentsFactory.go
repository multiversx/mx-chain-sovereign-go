package runTypeChain

type runTypeChainComponentsFactory struct{}

// NewRunTypeChainComponentsFactory will return a new instance of runType chain components factory
func NewRunTypeChainComponentsFactory() *runTypeChainComponentsFactory {
	return &runTypeChainComponentsFactory{}
}

// Create will create the runType chain components
func (rtccf *runTypeChainComponentsFactory) Create() *runTypeChainComponents {
	return &runTypeChainComponents{}
}

// IsInterfaceNil returns true if there is no value under the interface
func (rtccf *runTypeChainComponentsFactory) IsInterfaceNil() bool {
	return rtccf == nil
}
