package runTypeChain

type runTypeChainComponents struct{}

// Close does nothing
func (rtcc *runTypeChainComponents) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (rtcc *runTypeChainComponents) IsInterfaceNil() bool {
	return rtcc == nil
}
