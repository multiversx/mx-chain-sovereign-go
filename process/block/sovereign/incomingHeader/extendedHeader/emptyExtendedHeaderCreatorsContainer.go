package extendedHeader

import (
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/sovereign/dto"
)

type emptyBlockCreatorsContainer struct {
	mut           sync.RWMutex
	blockCreators map[dto.ChainID]EmptyExtendedHeaderCreator
}

// NewEmptyBlockCreatorsContainer creates a new extended block creators container
func NewEmptyBlockCreatorsContainer() *emptyBlockCreatorsContainer {
	return &emptyBlockCreatorsContainer{
		blockCreators: make(map[dto.ChainID]EmptyExtendedHeaderCreator),
	}
}

// Add will add a new empty extended block creator for a specific chain
func (container *emptyBlockCreatorsContainer) Add(chainID dto.ChainID, creator EmptyExtendedHeaderCreator) error {
	if check.IfNil(creator) {
		return fmt.Errorf("%w for chain %s", errNilEmptyExtendedHeaderCreator, chainID.String())
	}

	container.mut.Lock()
	container.blockCreators[chainID] = creator
	container.mut.Unlock()

	return nil
}

// Get will try to get an existing empty extended block creator. Errors if the type is not found
func (container *emptyBlockCreatorsContainer) Get(chainID dto.ChainID) (EmptyExtendedHeaderCreator, error) {
	container.mut.RLock()
	creator, ok := container.blockCreators[chainID]
	container.mut.RUnlock()

	if !ok {
		return nil, errChainIDNotFound
	}

	return creator, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (container *emptyBlockCreatorsContainer) IsInterfaceNil() bool {
	return container == nil
}
