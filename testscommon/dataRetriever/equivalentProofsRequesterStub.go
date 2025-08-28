package dataRetriever

// EquivalentProofsRequesterStub -
type EquivalentProofsRequesterStub struct {
	RequesterStub
	RequestDataFromNonceCalled func(nonceShardKey []byte, epoch uint32) error
}

// RequestDataFromNonce -
func (stub *EquivalentProofsRequesterStub) RequestDataFromNonce(nonceShardKey []byte, epoch uint32) error {
	if stub.RequestDataFromHashCalled != nil {
		return stub.RequestDataFromHashCalled(nonceShardKey, epoch)
	}

	return nil
}
