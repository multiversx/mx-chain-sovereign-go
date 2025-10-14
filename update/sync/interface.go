package sync

// CrossHeaderRequester defines a cross shard/chain header requester
type CrossHeaderRequester interface {
	ShouldRequestHeader(shardId uint32) bool
	RequestHeader(hash []byte)
	ShouldSkipProofCheck(shardId uint32) bool
	IsInterfaceNil() bool
}

// ExtendedShardHeaderRequestHandler defines an extended shard header requester
type ExtendedShardHeaderRequestHandler interface {
	RequestExtendedShardHeader(hash []byte)
	IsInterfaceNil() bool
}
