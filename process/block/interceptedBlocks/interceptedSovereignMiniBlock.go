package interceptedBlocks

type interceptedSovereignMiniBlock struct {
	*InterceptedMiniblock
}

// NewInterceptedSovereignMiniBlock creates a new instance of intercepted sovereign mini block
func NewInterceptedSovereignMiniBlock(arg *ArgInterceptedMiniblock) (*interceptedSovereignMiniBlock, error) {
	interceptedMbHandler, err := NewInterceptedMiniblock(arg)
	if err != nil {
		return nil, err
	}

	interceptedMbHandler.acceptedCrossShardIDs = getSovereignAcceptedCrossShardIDs()
	return &interceptedSovereignMiniBlock{
		interceptedMbHandler,
	}, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ismb *interceptedSovereignMiniBlock) IsInterfaceNil() bool {
	return ismb == nil
}
