package runTypeChain

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"

	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory"
)

var _ factory.ComponentHandler = (*managedRunTypeChainComponents)(nil)
var _ factory.RunTypeChainComponentsHandler = (*managedRunTypeChainComponents)(nil)
var _ factory.RunTypeChainComponentsHolder = (*managedRunTypeChainComponents)(nil)

type managedRunTypeChainComponents struct {
	*runTypeChainComponents
	factory                   runTypeChainComponentsCreator
	mutRunTypeChainComponents sync.RWMutex
}

// NewManagedRunTypeChainComponents returns a news instance of managed runType chain components
func NewManagedRunTypeChainComponents(rccf runTypeChainComponentsCreator) (*managedRunTypeChainComponents, error) {
	if rccf == nil {
		return nil, errors.ErrNilRunTypeChainComponentsFactory
	}

	return &managedRunTypeChainComponents{
		runTypeChainComponents: nil,
		factory:                rccf,
	}, nil
}

// Create will create the managed components
func (mrtcc *managedRunTypeChainComponents) Create() error {
	rtcc := mrtcc.factory.Create()

	mrtcc.mutRunTypeChainComponents.Lock()
	mrtcc.runTypeChainComponents = rtcc
	mrtcc.mutRunTypeChainComponents.Unlock()

	return nil
}

// Close will close all underlying subcomponents
func (mrtcc *managedRunTypeChainComponents) Close() error {
	mrtcc.mutRunTypeChainComponents.RLock()
	defer mrtcc.mutRunTypeChainComponents.RUnlock()

	if check.IfNil(mrtcc.runTypeChainComponents) {
		return nil
	}

	err := mrtcc.runTypeChainComponents.Close()
	if err != nil {
		return err
	}
	mrtcc.runTypeChainComponents = nil

	return nil
}

// CheckSubcomponents verifies all subcomponents
func (mrtcc *managedRunTypeChainComponents) CheckSubcomponents() error {
	mrtcc.mutRunTypeChainComponents.RLock()
	defer mrtcc.mutRunTypeChainComponents.RUnlock()

	if check.IfNil(mrtcc.runTypeChainComponents) {
		return errors.ErrNilRunTypeChainComponents
	}

	return nil
}

// String returns the name of the component
func (mrtcc *managedRunTypeChainComponents) String() string {
	return factory.RunTypeChainComponentsName
}

// IsInterfaceNil returns true if the interface is nil
func (mrtcc *managedRunTypeChainComponents) IsInterfaceNil() bool {
	return mrtcc == nil
}
