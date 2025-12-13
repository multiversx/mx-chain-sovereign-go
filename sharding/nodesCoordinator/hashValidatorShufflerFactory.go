package nodesCoordinator

type hashValidatorShufflerFactory struct{}

// NewHashValidatorShufflerFactory defines a new hash validator shuffler factory
func NewHashValidatorShufflerFactory() *hashValidatorShufflerFactory {
	return &hashValidatorShufflerFactory{}
}

// CreateHashValidatorShuffler creates a new hash validator shuffler
func (f *hashValidatorShufflerFactory) CreateHashValidatorShuffler(args *NodesShufflerArgs) (NodesShuffler, error) {
	return NewHashValidatorsShuffler(args)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (f *hashValidatorShufflerFactory) IsInterfaceNil() bool {
	return f == nil
}
