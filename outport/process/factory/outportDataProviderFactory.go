package factory

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"

	"github.com/multiversx/mx-chain-sovereign-go/common"
	"github.com/multiversx/mx-chain-sovereign-go/outport"
	"github.com/multiversx/mx-chain-sovereign-go/outport/process"
	"github.com/multiversx/mx-chain-sovereign-go/outport/process/alteredaccounts"
	"github.com/multiversx/mx-chain-sovereign-go/outport/process/disabled"
	"github.com/multiversx/mx-chain-sovereign-go/outport/process/transactionsfee"
	processTxs "github.com/multiversx/mx-chain-sovereign-go/process"
	"github.com/multiversx/mx-chain-sovereign-go/process/smartContract"
	"github.com/multiversx/mx-chain-sovereign-go/sharding"
	"github.com/multiversx/mx-chain-sovereign-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-sovereign-go/state"
	"github.com/multiversx/mx-chain-sovereign-go/storage"
)

// ArgOutportDataProviderFactory holds the arguments needed for creating a new instance of outport.DataProviderOutport
type ArgOutportDataProviderFactory struct {
	IsImportDBMode         bool
	HasDrivers             bool
	AddressConverter       core.PubkeyConverter
	AccountsDB             state.AccountsAdapter
	Marshaller             marshal.Marshalizer
	EsdtDataStorageHandler vmcommon.ESDTNFTStorageHandler
	TransactionsStorer     storage.Storer
	ShardCoordinator       sharding.Coordinator
	TxCoordinator          processTxs.TransactionCoordinator
	NodesCoordinator       nodesCoordinator.NodesCoordinator
	GasConsumedProvider    process.GasConsumedProvider
	EconomicsData          process.EconomicsDataHandler
	Hasher                 hashing.Hasher
	MbsStorer              storage.Storer
	EnableEpochsHandler    common.EnableEpochsHandler
	ExecutionOrderGetter   common.ExecutionOrderGetter
}

type outportDataProviderFactory struct {
}

// NewOutportDataProviderFactory creates a new outport data provider factory
func NewOutportDataProviderFactory() *outportDataProviderFactory {
	return &outportDataProviderFactory{}
}

// CreateOutportDataProvider will create a new instance of outport.DataProviderOutport
func (f *outportDataProviderFactory) CreateOutportDataProvider(arg ArgOutportDataProviderFactory) (outport.DataProviderOutport, error) {
	if !arg.HasDrivers {
		return disabled.NewDisabledOutportDataProvider(), nil
	}

	argsOutport, err := createArgs(arg)
	if err != nil {
		return nil, err
	}

	return process.NewOutportDataProvider(*argsOutport)
}

func createArgs(arg ArgOutportDataProviderFactory) (*process.ArgOutportDataProvider, error) {
	err := checkArgOutportDataProviderFactory(arg)
	if err != nil {
		return nil, err
	}

	alteredAccountsProvider, err := alteredaccounts.NewAlteredAccountsProvider(alteredaccounts.ArgsAlteredAccountsProvider{
		ShardCoordinator:       arg.ShardCoordinator,
		AddressConverter:       arg.AddressConverter,
		AccountsDB:             arg.AccountsDB,
		EsdtDataStorageHandler: arg.EsdtDataStorageHandler,
	})
	if err != nil {
		return nil, err
	}

	transactionsFeeProc, err := transactionsfee.NewTransactionsFeeProcessor(transactionsfee.ArgTransactionsFeeProcessor{
		Marshaller:          arg.Marshaller,
		TransactionsStorer:  arg.TransactionsStorer,
		ShardCoordinator:    arg.ShardCoordinator,
		TxFeeCalculator:     arg.EconomicsData,
		PubKeyConverter:     arg.AddressConverter,
		ArgsParser:          smartContract.NewArgumentParser(),
		EnableEpochsHandler: arg.EnableEpochsHandler,
	})
	if err != nil {
		return nil, err
	}

	return &process.ArgOutportDataProvider{
		IsImportDBMode:           arg.IsImportDBMode,
		ShardCoordinator:         arg.ShardCoordinator,
		AlteredAccountsProvider:  alteredAccountsProvider,
		TransactionsFeeProcessor: transactionsFeeProc,
		TxCoordinator:            arg.TxCoordinator,
		NodesCoordinator:         arg.NodesCoordinator,
		GasConsumedProvider:      arg.GasConsumedProvider,
		EconomicsData:            arg.EconomicsData,
		ExecutionOrderHandler:    arg.ExecutionOrderGetter,
		Hasher:                   arg.Hasher,
		Marshaller:               arg.Marshaller,
	}, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (f *outportDataProviderFactory) IsInterfaceNil() bool {
	return f == nil
}
