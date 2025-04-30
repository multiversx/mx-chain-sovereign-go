package runTypeChain

type sovereignRunTypeChainComponentsFactory struct{}

// NewSovereignRunTypeChainComponentsFactory will return a new instance of sovereign runType chain components factory
func NewSovereignRunTypeChainComponentsFactory() *sovereignRunTypeChainComponentsFactory {
	return &sovereignRunTypeChainComponentsFactory{}
}

// Create will create the runType chain components
func (srtccf *sovereignRunTypeChainComponentsFactory) Create() *runTypeChainComponents {
	return &runTypeChainComponents{}
}

// IsInterfaceNil returns true if there is no value under the interface
func (srtccf *sovereignRunTypeChainComponentsFactory) IsInterfaceNil() bool {
	return srtccf == nil
}
