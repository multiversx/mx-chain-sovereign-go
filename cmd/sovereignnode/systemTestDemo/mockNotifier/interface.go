package main

// GRPCServer holds the mock server actions
type GRPCServer interface {
	ExtractRandomBridgeTopicsForConfirmation() ([]*ConfirmedBridgeOp, error)
}

// GRPCServerConnection holds the grpc server actions
type GRPCServerConnection interface {
	Stop()
}
