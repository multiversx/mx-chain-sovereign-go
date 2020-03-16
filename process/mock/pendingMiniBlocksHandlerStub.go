package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

// PendingMiniBlocksHandlerStub -
type PendingMiniBlocksHandlerStub struct {
	AddProcessedHeaderCalled      func(handler data.HeaderHandler) error
	RevertHeaderCalled            func(handler data.HeaderHandler) error
	GetNumPendingMiniBlocksCalled func(shardID uint32) uint32
	SetNumPendingMiniBlocksCalled func(shardID uint32, numPendingMiniBlocks uint32)
}

// AddProcessedHeader -
func (p *PendingMiniBlocksHandlerStub) AddProcessedHeader(handler data.HeaderHandler) error {
	if p.AddProcessedHeaderCalled != nil {
		return p.AddProcessedHeaderCalled(handler)
	}
	return nil
}

// RevertHeader -
func (p *PendingMiniBlocksHandlerStub) RevertHeader(handler data.HeaderHandler) error {
	if p.RevertHeaderCalled != nil {
		return p.RevertHeaderCalled(handler)
	}
	return nil
}

// GetNumPendingMiniBlocks -
func (p *PendingMiniBlocksHandlerStub) GetNumPendingMiniBlocks(shardID uint32) uint32 {
	if p.GetNumPendingMiniBlocksCalled != nil {
		return p.GetNumPendingMiniBlocksCalled(shardID)
	}
	return 0
}

// SetNumPendingMiniBlocks -
func (p *PendingMiniBlocksHandlerStub) SetNumPendingMiniBlocks(shardID uint32, numPendingMiniBlocks uint32) {
	if p.SetNumPendingMiniBlocksCalled != nil {
		p.SetNumPendingMiniBlocksCalled(shardID, numPendingMiniBlocks)
	}
}

// IsInterfaceNil -
func (p *PendingMiniBlocksHandlerStub) IsInterfaceNil() bool {
	return p == nil
}
