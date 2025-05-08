package interceptedBlocks

type interceptedSovereignBlockHeader struct {
	*InterceptedHeader
}

// NewSovereignInterceptedBlockHeader creates a new intercepted sovereign block header
func NewSovereignInterceptedBlockHeader(arg *ArgInterceptedBlockHeader) (*interceptedSovereignBlockHeader, error) {
	interceptedHdr, err := NewInterceptedHeader(arg)
	if err != nil {
		return nil, err
	}

	interceptedHdr.acceptedCrossShardIDs = getSovereignAcceptedCrossShardIDs()
	sovInterceptedBlock := &interceptedSovereignBlockHeader{
		interceptedHdr,
	}

	return sovInterceptedBlock, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (isbh *interceptedSovereignBlockHeader) IsInterfaceNil() bool {
	return isbh == nil
}
