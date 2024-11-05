package external

import "github.com/multiversx/mx-chain-core-go/data/transaction"

// ArgsCreateTransaction defines arguments for creating a transaction
type ArgsCreateTransaction struct {
	Nonce                uint64
	Value                string
	Receiver             string
	ReceiverUsername     []byte
	ReceiverAliasAddress []byte
	Sender               string
	SenderUsername       []byte
	SenderAliasAddress   []byte
	GasPrice             uint64
	GasLimit             uint64
	DataField            []byte
	OriginalDataField    []byte
	SignatureHex         string
	ChainID              string
	Version              uint32
	Options              uint32
	Guardian             string
	GuardianSigHex       string
	Relayer              string
	InnerTransactions    []*transaction.Transaction
}
