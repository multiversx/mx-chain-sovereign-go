package hooks

import (
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"

	"github.com/multiversx/mx-chain-sovereign-go/process"
)

// BlockChainHookCounter defines the operations of a blockchain hook counter handler
type BlockChainHookCounter interface {
	ProcessCrtNumberOfTrieReadsCounter() error
	ProcessMaxBuiltInCounters(input *vmcommon.ContractCallInput) error
	ResetCounters()
	SetMaximumValues(mapsOfValues map[string]uint64)
	GetCounterValues() map[string]uint64
	IsInterfaceNil() bool
}

// BlockChainHookHandlerCreator defines the blockchain hook factory handler
type BlockChainHookHandlerCreator interface {
	CreateBlockChainHookHandler(args ArgBlockChainHook) (process.BlockChainHookWithAccountsAdapter, error)
	IsInterfaceNil() bool
}
