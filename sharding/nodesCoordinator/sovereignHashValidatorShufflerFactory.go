package nodesCoordinator

type sovereignHashValidatorShufflerFactory struct{}

// NewSovereignHashValidatorShufflerFactory defines a new sovereign hash validator shuffler factory
func NewSovereignHashValidatorShufflerFactory() *sovereignHashValidatorShufflerFactory {
	return &sovereignHashValidatorShufflerFactory{}
}

// CreateHashValidatorShuffler creates a new hash validator shuffler
func (f *sovereignHashValidatorShufflerFactory) CreateHashValidatorShuffler(args *NodesShufflerArgs) (NodesShuffler, error) {
	shuffler, err := NewHashValidatorsShuffler(args)
	if err != nil {
		return nil, err
	}

	return NewSovereignHashValidatorsShuffler(shuffler)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (f *sovereignHashValidatorShufflerFactory) IsInterfaceNil() bool {
	return f == nil
}
