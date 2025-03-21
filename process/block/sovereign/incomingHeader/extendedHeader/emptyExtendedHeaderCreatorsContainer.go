package extendedHeader

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
)

type emptyBlockCreatorsContainer struct {
	mut           sync.RWMutex
	blockCreators map[sovereign.ChainID]EmptyExtendedHeaderCreator
}

// NewEmptyBlockCreatorsContainer creates a new extended block creators container
func NewEmptyBlockCreatorsContainer() *emptyBlockCreatorsContainer {
	return &emptyBlockCreatorsContainer{
		blockCreators: make(map[sovereign.ChainID]EmptyExtendedHeaderCreator),
	}
}

// Add will add a new empty extended block creator
func (container *emptyBlockCreatorsContainer) Add(headerType sovereign.ChainID, creator EmptyExtendedHeaderCreator) error {
	if check.IfNil(creator) {
		return data.ErrNilEmptyBlockCreator
	}

	container.mut.Lock()
	container.blockCreators[headerType] = creator
	container.mut.Unlock()

	return nil
}

// Get will try to get an existing empty extended block creator. Errors if the type is not found
func (container *emptyBlockCreatorsContainer) Get(headerType sovereign.ChainID) (EmptyExtendedHeaderCreator, error) {
	container.mut.RLock()
	creator, ok := container.blockCreators[headerType]
	container.mut.RUnlock()

	if !ok {
		return nil, data.ErrInvalidHeaderType
	}

	return creator, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (container *emptyBlockCreatorsContainer) IsInterfaceNil() bool {
	return container == nil
}
