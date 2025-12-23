package facade

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/throttler"
	chainData "github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/alteredAccount"
	apiData "github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/data/validator"
	"github.com/multiversx/mx-chain-core-go/data/vm"
	logger "github.com/multiversx/mx-chain-logger-go"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/debug"
	"github.com/multiversx/mx-chain-go/epochStart/bootstrap/disabled"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/heartbeat/data"
	"github.com/multiversx/mx-chain-go/node/external"
	"github.com/multiversx/mx-chain-go/ntp"
	"github.com/multiversx/mx-chain-go/process"
	txSimData "github.com/multiversx/mx-chain-go/process/transactionEvaluator/data"
	"github.com/multiversx/mx-chain-go/state"
)

// DefaultRestInterface is the default interface the rest API will start on if not specified
const DefaultRestInterface = "localhost:8080"

// DefaultRestPortOff is the default value that should be passed if it is desired
// to start the node without a REST endpoint available
const DefaultRestPortOff = "off"

var log = logger.GetOrCreate("facade")

// ArgNodeFacade represents the argument for the NodeFacade
type ArgNodeFacade struct {
	Node                   NodeHandler
	ApiResolver            ApiResolver
	RestAPIServerDebugMode bool
	WsAntifloodConfig      config.WebServerAntifloodConfig
	FacadeConfig           config.FacadeConfig
	ApiRoutesConfig        config.ApiRoutesConfig
	AccountsState          state.AccountsAdapter
	PeerState              state.AccountsAdapter
	Blockchain             chainData.ChainHandler
}

// NodeFacade represents a facade for grouping the functionality for the node
type NodeFacade struct {
	node                   NodeHandler
	apiResolver            ApiResolver
	syncer                 ntp.SyncTimer
	config                 config.FacadeConfig
	apiRoutesConfig        config.ApiRoutesConfig
	endpointsThrottlers    map[string]core.Throttler
	wsAntifloodConfig      config.WebServerAntifloodConfig
	restAPIServerDebugMode bool
	accountsState          state.AccountsAdapter
	peerState              state.AccountsAdapter
	blockchain             chainData.ChainHandler
}

// NewNodeFacade creates a new Facade with a NodeWrapper
func NewNodeFacade(arg ArgNodeFacade) (*NodeFacade, error) {
	if check.IfNil(arg.Node) {
		return nil, errors.ErrNilNode
	}
	if check.IfNil(arg.ApiResolver) {
		return nil, ErrNilApiResolver
	}
	if len(arg.ApiRoutesConfig.APIPackages) == 0 {
		return nil, ErrNoApiRoutesConfig
	}
	err := checkWebserverAntifloodConfig(arg.WsAntifloodConfig)
	if err != nil {
		return nil, err
	}
	if check.IfNil(arg.AccountsState) {
		return nil, ErrNilAccountState
	}
	if check.IfNil(arg.PeerState) {
		return nil, ErrNilPeerState
	}
	if check.IfNil(arg.Blockchain) {
		return nil, ErrNilBlockchain
	}

	throttlersMap := computeEndpointsNumGoRoutinesThrottlers(arg.WsAntifloodConfig)

	nf := &NodeFacade{
		node:                   arg.Node,
		apiResolver:            arg.ApiResolver,
		restAPIServerDebugMode: arg.RestAPIServerDebugMode,
		wsAntifloodConfig:      arg.WsAntifloodConfig,
		config:                 arg.FacadeConfig,
		apiRoutesConfig:        arg.ApiRoutesConfig,
		endpointsThrottlers:    throttlersMap,
		accountsState:          arg.AccountsState,
		peerState:              arg.PeerState,
		blockchain:             arg.Blockchain,
	}

	return nf, nil
}

func checkWebserverAntifloodConfig(cfg config.WebServerAntifloodConfig) error {
	if !cfg.WebServerAntifloodEnabled {
		return nil
	}

	if cfg.SimultaneousRequests == 0 {
		return fmt.Errorf("%w, SimultaneousRequests should not be 0", ErrInvalidValue)
	}
	if cfg.SameSourceRequests == 0 {
		return fmt.Errorf("%w, SameSourceRequests should not be 0", ErrInvalidValue)
	}
	if cfg.SameSourceResetIntervalInSec == 0 {
		return fmt.Errorf("%w, SameSourceResetIntervalInSec should not be 0", ErrInvalidValue)
	}
	if cfg.TrieOperationsDeadlineMilliseconds == 0 {
		return fmt.Errorf("%w, TrieOperationsDeadlineMilliseconds should not be 0", ErrInvalidValue)
	}

	return nil
}

func computeEndpointsNumGoRoutinesThrottlers(webServerAntiFloodConfig config.WebServerAntifloodConfig) map[string]core.Throttler {
	throttlersMap := make(map[string]core.Throttler)
	for _, endpointSetting := range webServerAntiFloodConfig.EndpointsThrottlers {
		newThrottler, err := throttler.NewNumGoRoutinesThrottler(endpointSetting.MaxNumGoRoutines)
		if err != nil {
			log.Warn("error when setting the maximum go routines throttler for endpoint",
				"endpoint", endpointSetting.Endpoint,
				"max go routines", endpointSetting.MaxNumGoRoutines,
				"error", err,
			)
			continue
		}
		throttlersMap[endpointSetting.Endpoint] = newThrottler
	}

	return throttlersMap
}

// SetSyncer sets the current syncer
func (nf *NodeFacade) SetSyncer(syncer ntp.SyncTimer) {
	nf.syncer = syncer
}

// RestAPIServerDebugMode return true is debug mode for Rest API is enabled
func (nf *NodeFacade) RestAPIServerDebugMode() bool {
	return nf.restAPIServerDebugMode
}

// RestApiInterface returns the interface on which the rest API should start on, based on the config file provided.
// The API will start on the DefaultRestInterface value unless a correct value is passed or
// //	the value is explicitly set to off, in which case it will not start at all
func (nf *NodeFacade) RestApiInterface() string {
	if nf.config.RestApiInterface == "" {
		return DefaultRestInterface
	}

	return nf.config.RestApiInterface
}

// GetBalance gets the current balance for a specified address
func (nf *NodeFacade) GetBalance(address string, options apiData.AccountQueryOptions) (*big.Int, apiData.BlockInfo, error) {
	return nf.node.GetBalance(address, options)
}

// GetUsername gets the username for a specified address
func (nf *NodeFacade) GetUsername(address string, options apiData.AccountQueryOptions) (string, apiData.BlockInfo, error) {
	return nf.node.GetUsername(address, options)
}

// GetCodeHash gets the code hash for a specified address
func (nf *NodeFacade) GetCodeHash(address string, options apiData.AccountQueryOptions) ([]byte, apiData.BlockInfo, error) {
	return nf.node.GetCodeHash(address, options)
}

// GetValueForKey gets the value for a key in a given address
func (nf *NodeFacade) GetValueForKey(address string, key string, options apiData.AccountQueryOptions) (string, apiData.BlockInfo, error) {
	return nf.node.GetValueForKey(address, key, options)
}

// GetESDTData returns the ESDT data for the given address, tokenID and nonce
func (nf *NodeFacade) GetESDTData(address string, key string, nonce uint64, options apiData.AccountQueryOptions) (*esdt.ESDigitalToken, apiData.BlockInfo, error) {
	return nf.node.GetESDTData(address, key, nonce, options)
}

// GetESDTsRoles returns all the tokens identifiers and roles for the given address
func (nf *NodeFacade) GetESDTsRoles(address string, options apiData.AccountQueryOptions) (map[string][]string, apiData.BlockInfo, error) {
	ctx, cancel := nf.getContextForApiTrieRangeOperations()
	defer cancel()

	return nf.node.GetESDTsRoles(address, options, ctx)
}

// GetNFTTokenIDsRegisteredByAddress returns all the token identifiers for semi or non fungible tokens registered by the address
func (nf *NodeFacade) GetNFTTokenIDsRegisteredByAddress(address string, options apiData.AccountQueryOptions) ([]string, apiData.BlockInfo, error) {
	ctx, cancel := nf.getContextForApiTrieRangeOperations()
	defer cancel()

	return nf.node.GetNFTTokenIDsRegisteredByAddress(address, options, ctx)
}

// GetESDTsWithRole returns all the tokens with the given role for the given address
func (nf *NodeFacade) GetESDTsWithRole(address string, role string, options apiData.AccountQueryOptions) ([]string, apiData.BlockInfo, error) {
	ctx, cancel := nf.getContextForApiTrieRangeOperations()
	defer cancel()

	return nf.node.GetESDTsWithRole(address, role, options, ctx)
}

// GetKeyValuePairs returns all the key-value pairs under the provided address
func (nf *NodeFacade) GetKeyValuePairs(address string, options apiData.AccountQueryOptions) (map[string]string, apiData.BlockInfo, error) {
	ctx, cancel := nf.getContextForApiTrieRangeOperations()
	defer cancel()

	return nf.node.GetKeyValuePairs(address, options, ctx)
}

// IterateKeys starts from the given iteratorState and returns the next key-value pairs and the new iteratorState
func (nf *NodeFacade) IterateKeys(address string, numKeys uint, iteratorState [][]byte, options apiData.AccountQueryOptions) (map[string]string, [][]byte, apiData.BlockInfo, error) {
	ctx, cancel := nf.getContextForApiTrieRangeOperations()
	defer cancel()

	return nf.node.IterateKeys(address, numKeys, iteratorState, options, ctx)
}

// GetGuardianData returns the guardian data for the provided address
func (nf *NodeFacade) GetGuardianData(address string, options apiData.AccountQueryOptions) (apiData.GuardianData, apiData.BlockInfo, error) {
	return nf.node.GetGuardianData(address, options)
}

// GetAllESDTTokens returns all the esdt tokens for a given address
func (nf *NodeFacade) GetAllESDTTokens(address string, options apiData.AccountQueryOptions) (map[string]*esdt.ESDigitalToken, apiData.BlockInfo, error) {
	ctx, cancel := nf.getContextForApiTrieRangeOperations()
	defer cancel()

	return nf.node.GetAllESDTTokens(address, options, ctx)
}

// GetTokenSupply returns the provided token supply
func (nf *NodeFacade) GetTokenSupply(token string) (*apiData.ESDTSupply, error) {
	return nf.node.GetTokenSupply(token)
}

// GetAllIssuedESDTs returns all the issued esdts from the esdt system smart contract
func (nf *NodeFacade) GetAllIssuedESDTs(tokenType string) ([]string, error) {
	ctx, cancel := nf.getContextForApiTrieRangeOperations()
	defer cancel()

	return nf.node.GetAllIssuedESDTs(tokenType, ctx)
}

func (nf *NodeFacade) getContextForApiTrieRangeOperations() (context.Context, context.CancelFunc) {
	if !nf.wsAntifloodConfig.WebServerAntifloodEnabled {
		return context.WithCancel(context.Background())
	}

	timeout := time.Duration(nf.wsAntifloodConfig.TrieOperationsDeadlineMilliseconds) * time.Millisecond
	return context.WithTimeout(context.Background(), timeout)
}

// CreateTransaction creates a transaction from all needed fields
func (nf *NodeFacade) CreateTransaction(requestTx map[string]interface{}) (chainData.TransactionHandler, []byte, error) {
	return nf.node.CreateTransaction(requestTx)
}

// ValidateTransaction will validate a transaction
func (nf *NodeFacade) ValidateTransaction(tx chainData.TransactionHandler) error {
	return nf.node.ValidateTransaction(tx)
}

// ValidateTransactionForSimulation will validate a transaction for the simulation process
func (nf *NodeFacade) ValidateTransactionForSimulation(tx chainData.TransactionHandler, checkSignature bool) error {
	return nf.node.ValidateTransactionForSimulation(tx, checkSignature)
}

// ValidatorStatisticsApi will return the statistics for all validators
func (nf *NodeFacade) ValidatorStatisticsApi() (map[string]*validator.ValidatorStatistics, error) {
	return nf.node.ValidatorStatisticsApi()
}

// AuctionListApi will return the data about the validators in the auction list
func (nf *NodeFacade) AuctionListApi() ([]*common.AuctionListValidatorAPIResponse, error) {
	return nf.node.AuctionListApi()
}

// SendBulkTransactions will send a bulk of transactions on the topic channel
func (nf *NodeFacade) SendBulkTransactions(txs []chainData.TransactionHandler) (uint64, error) {
	return nf.node.SendBulkTransactions(txs)
}

// SimulateTransactionExecution will simulate a transaction's execution and will return the results
func (nf *NodeFacade) SimulateTransactionExecution(tx chainData.TransactionHandler) (*txSimData.SimulationResultsWithVMOutput, error) {
	return nf.apiResolver.SimulateTransactionExecution(tx)
}

// GetTransaction gets the transaction with a specified hash
func (nf *NodeFacade) GetTransaction(hash string, withResults bool) (*transaction.ApiTransactionResult, error) {
	return nf.apiResolver.GetTransaction(hash, withResults)
}

// GetSCRsByTxHash will return a list of smart contract results based on a provided tx hash and smart contract result hash
func (nf *NodeFacade) GetSCRsByTxHash(txHash string, scrHash string) ([]*transaction.ApiSmartContractResult, error) {
	return nf.apiResolver.GetSCRsByTxHash(txHash, scrHash)
}

// GetTransactionsPool will return a structure containing the transactions pool that is to be returned on API calls
func (nf *NodeFacade) GetTransactionsPool(fields string) (*common.TransactionsPoolAPIResponse, error) {
	return nf.apiResolver.GetTransactionsPool(fields)
}

// GetTransactionsPoolForSender will return a structure containing the transactions for sender that is to be returned on API calls
func (nf *NodeFacade) GetTransactionsPoolForSender(sender, fields string) (*common.TransactionsPoolForSenderApiResponse, error) {
	return nf.apiResolver.GetTransactionsPoolForSender(sender, fields)
}

// GetLastPoolNonceForSender will return the last nonce from pool for sender that is to be returned on API calls
func (nf *NodeFacade) GetLastPoolNonceForSender(sender string) (uint64, error) {
	return nf.apiResolver.GetLastPoolNonceForSender(sender)
}

// GetTransactionsPoolNonceGapsForSender will return the nonce gaps from pool for sender, if exists, that is to be returned on API calls
func (nf *NodeFacade) GetTransactionsPoolNonceGapsForSender(sender string) (*common.TransactionsPoolNonceGapsForSenderApiResponse, error) {
	accountResponse, _, err := nf.node.GetAccount(sender, apiData.AccountQueryOptions{})
	if err != nil {
		return &common.TransactionsPoolNonceGapsForSenderApiResponse{}, err
	}

	return nf.apiResolver.GetTransactionsPoolNonceGapsForSender(sender, accountResponse.Nonce)
}

// ComputeTransactionGasLimit will estimate how many gas a transaction will consume
func (nf *NodeFacade) ComputeTransactionGasLimit(tx chainData.TransactionHandler) (*transaction.CostResponse, error) {
	return nf.apiResolver.ComputeTransactionGasLimit(tx)
}

// GetAccount returns a response containing information about the account correlated with provided address
func (nf *NodeFacade) GetAccount(address string, options apiData.AccountQueryOptions) (apiData.AccountResponse, apiData.BlockInfo, error) {
	var accountResponse apiData.AccountResponse
	var blockInfo apiData.BlockInfo
	var err error

	if options.WithKeys {
		ctx, cancel := nf.getContextForApiTrieRangeOperations()
		defer cancel()

		accountResponse, blockInfo, err = nf.node.GetAccountWithKeys(address, options, ctx)
	} else {
		accountResponse, blockInfo, err = nf.node.GetAccount(address, options)
	}

	if err != nil {
		return apiData.AccountResponse{}, apiData.BlockInfo{}, err
	}

	codeHash := accountResponse.CodeHash
	code, _ := nf.node.GetCode(codeHash, options)
	accountResponse.Code = hex.EncodeToString(code)
	return accountResponse, blockInfo, nil
}

// GetAccounts returns the state of the provided addresses
func (nf *NodeFacade) GetAccounts(addresses []string, options apiData.AccountQueryOptions) (map[string]*apiData.AccountResponse, apiData.BlockInfo, error) {
	numAddresses := uint32(len(addresses))
	// TODO: check if Antiflood is enabled before applying this constraint (EN-13278)
	maxBulkSize := nf.wsAntifloodConfig.GetAddressesBulkMaxSize
	if numAddresses > maxBulkSize {
		return nil, apiData.BlockInfo{}, fmt.Errorf("%w (provided: %d, maximum: %d)", ErrTooManyAddressesInBulk, numAddresses, maxBulkSize)
	}

	response := make(map[string]*apiData.AccountResponse)
	var blockInfo apiData.BlockInfo

	for i, address := range addresses {
		accountResponse, blockInfoForAccount, err := nf.node.GetAccount(address, options)
		if err != nil {
			return nil, apiData.BlockInfo{}, err
		}
		// Use the first block info as the block info for the whole bulk
		if i == 0 {
			blockInfo = blockInfoForAccount
			blockRootHash, errBlockRootHash := hex.DecodeString(blockInfoForAccount.RootHash)
			if errBlockRootHash == nil {
				options.BlockRootHash = blockRootHash
			}
		}

		codeHash := accountResponse.CodeHash
		code, _ := nf.node.GetCode(codeHash, options)
		accountResponse.Code = hex.EncodeToString(code)

		response[address] = &accountResponse
	}

	return response, blockInfo, nil
}

// GetHeartbeats returns the heartbeat status for each public key from initial list or later joined to the network
func (nf *NodeFacade) GetHeartbeats() ([]data.PubKeyHeartbeat, error) {
	hbStatus := nf.node.GetHeartbeats()
	if hbStatus == nil {
		return nil, ErrHeartbeatsNotActive
	}

	return hbStatus, nil
}

// StatusMetrics will return the node's status metrics
func (nf *NodeFacade) StatusMetrics() external.StatusMetricsHandler {
	return nf.apiResolver.StatusMetrics()
}

// GetTotalStakedValue will return total staked value
func (nf *NodeFacade) GetTotalStakedValue() (*apiData.StakeValues, error) {
	ctx, cancel := nf.getContextForApiTrieRangeOperations()
	defer cancel()

	return nf.apiResolver.GetTotalStakedValue(ctx)
}

// GetDirectStakedList will output the list for the direct staked addresses
func (nf *NodeFacade) GetDirectStakedList() ([]*apiData.DirectStakedValue, error) {
	ctx, cancel := nf.getContextForApiTrieRangeOperations()
	defer cancel()

	return nf.apiResolver.GetDirectStakedList(ctx)
}

// GetDelegatorsList will output the list for the delegators addresses
func (nf *NodeFacade) GetDelegatorsList() ([]*apiData.Delegator, error) {
	ctx, cancel := nf.getContextForApiTrieRangeOperations()
	defer cancel()

	return nf.apiResolver.GetDelegatorsList(ctx)
}

// ExecuteSCQuery retrieves data from existing SC trie
func (nf *NodeFacade) ExecuteSCQuery(query *process.SCQuery) (*vm.VMOutputApi, apiData.BlockInfo, error) {
	vmOutput, blockInfo, err := nf.apiResolver.ExecuteSCQuery(query)
	if err != nil {
		return nil, apiData.BlockInfo{}, err
	}

	return nf.convertVmOutputToApiResponse(vmOutput), queryBlockInfoToApiResource(blockInfo), nil
}

// PprofEnabled returns if profiling mode should be active or not on the application
func (nf *NodeFacade) PprofEnabled() bool {
	return nf.config.PprofEnabled
}

// Trigger will trigger a hardfork event
func (nf *NodeFacade) Trigger(epoch uint32, withEarlyEndOfEpoch bool) error {
	return nf.node.DirectTrigger(epoch, withEarlyEndOfEpoch)
}

// IsSelfTrigger returns true if the self public key is the same with the registered public key
func (nf *NodeFacade) IsSelfTrigger() bool {
	return nf.node.IsSelfTrigger()
}

// EncodeAddressPubkey will encode the provided address public key bytes to string
func (nf *NodeFacade) EncodeAddressPubkey(pk []byte) (string, error) {
	return nf.node.EncodeAddressPubkey(pk)
}

// DecodeAddressPubkey will try to decode the provided address public key string
func (nf *NodeFacade) DecodeAddressPubkey(pk string) ([]byte, error) {
	return nf.node.DecodeAddressPubkey(pk)
}

// GetQueryHandler returns the query handler if existing
func (nf *NodeFacade) GetQueryHandler(name string) (debug.QueryHandler, error) {
	return nf.node.GetQueryHandler(name)
}

// GetEpochStartDataAPI returns epoch start data of the provided epoch
func (nf *NodeFacade) GetEpochStartDataAPI(epoch uint32) (*common.EpochStartDataAPI, error) {
	return nf.node.GetEpochStartDataAPI(epoch)
}

// GetPeerInfo returns the peer info of a provided pid
func (nf *NodeFacade) GetPeerInfo(pid string) ([]core.QueryP2PPeerInfo, error) {
	return nf.node.GetPeerInfo(pid)
}

// GetConnectedPeersRatingsOnMainNetwork returns the connected peers ratings on the main network
func (nf *NodeFacade) GetConnectedPeersRatingsOnMainNetwork() (string, error) {
	return nf.node.GetConnectedPeersRatingsOnMainNetwork()
}

// GetThrottlerForEndpoint returns the throttler for a given endpoint if found
func (nf *NodeFacade) GetThrottlerForEndpoint(endpoint string) (core.Throttler, bool) {
	if !nf.wsAntifloodConfig.WebServerAntifloodEnabled {
		return disabled.NewThrottler(), true
	}

	throttlerForEndpoint, ok := nf.endpointsThrottlers[endpoint]
	isThrottlerOk := ok && throttlerForEndpoint != nil

	return throttlerForEndpoint, isThrottlerOk
}

// GetBlockByHash return the block for a given hash
func (nf *NodeFacade) GetBlockByHash(hash string, options apiData.BlockQueryOptions) (*apiData.Block, error) {
	return nf.apiResolver.GetBlockByHash(hash, options)
}

// GetBlockByNonce returns the block for a given nonce
func (nf *NodeFacade) GetBlockByNonce(nonce uint64, options apiData.BlockQueryOptions) (*apiData.Block, error) {
	return nf.apiResolver.GetBlockByNonce(nonce, options)
}

// GetBlockByRound returns the block for a given round
func (nf *NodeFacade) GetBlockByRound(round uint64, options apiData.BlockQueryOptions) (*apiData.Block, error) {
	return nf.apiResolver.GetBlockByRound(round, options)
}

// GetAlteredAccountsForBlock returns the altered accounts for a given block
func (nf *NodeFacade) GetAlteredAccountsForBlock(options apiData.GetAlteredAccountsForBlockOptions) ([]*alteredAccount.AlteredAccount, error) {
	return nf.apiResolver.GetAlteredAccountsForBlock(options)
}

// GetInternalMetaBlockByHash return the meta block for a given hash
func (nf *NodeFacade) GetInternalMetaBlockByHash(format common.ApiOutputFormat, hash string) (interface{}, error) {
	return nf.apiResolver.GetInternalMetaBlockByHash(format, hash)
}

// GetInternalMetaBlockByNonce returns the meta block for a given nonce
func (nf *NodeFacade) GetInternalMetaBlockByNonce(format common.ApiOutputFormat, nonce uint64) (interface{}, error) {
	return nf.apiResolver.GetInternalMetaBlockByNonce(format, nonce)
}

// GetInternalMetaBlockByRound returns the meta block for a given round
func (nf *NodeFacade) GetInternalMetaBlockByRound(format common.ApiOutputFormat, round uint64) (interface{}, error) {
	return nf.apiResolver.GetInternalMetaBlockByRound(format, round)
}

// GetInternalStartOfEpochMetaBlock will return start of epoch meta block
// for a specified epoch
func (nf *NodeFacade) GetInternalStartOfEpochMetaBlock(format common.ApiOutputFormat, epoch uint32) (interface{}, error) {
	return nf.apiResolver.GetInternalStartOfEpochMetaBlock(format, epoch)
}

// GetInternalStartOfEpochValidatorsInfo will return start of epoch validators info
// for a specified epoch
func (nf *NodeFacade) GetInternalStartOfEpochValidatorsInfo(epoch uint32) ([]*state.ShardValidatorInfo, error) {
	return nf.apiResolver.GetInternalStartOfEpochValidatorsInfo(epoch)
}

// GetInternalShardBlockByHash return the shard block for a given hash
func (nf *NodeFacade) GetInternalShardBlockByHash(format common.ApiOutputFormat, hash string) (interface{}, error) {
	return nf.apiResolver.GetInternalShardBlockByHash(format, hash)
}

// GetInternalShardBlockByNonce returns the shard block for a given nonce
func (nf *NodeFacade) GetInternalShardBlockByNonce(format common.ApiOutputFormat, nonce uint64) (interface{}, error) {
	return nf.apiResolver.GetInternalShardBlockByNonce(format, nonce)
}

// GetInternalShardBlockByRound returns the shard block for a given round
func (nf *NodeFacade) GetInternalShardBlockByRound(format common.ApiOutputFormat, round uint64) (interface{}, error) {
	return nf.apiResolver.GetInternalShardBlockByRound(format, round)
}

// GetInternalMiniBlockByHash return the miniblock for a given hash
func (nf *NodeFacade) GetInternalMiniBlockByHash(format common.ApiOutputFormat, txHash string, epoch uint32) (interface{}, error) {
	return nf.apiResolver.GetInternalMiniBlock(format, txHash, epoch)
}

// Close will clean up started go routines
func (nf *NodeFacade) Close() error {
	log.LogIfError(nf.apiResolver.Close())

	return nil
}

// GetProof returns the Merkle proof for the given address and root hash
func (nf *NodeFacade) GetProof(rootHash string, address string) (*common.GetProofResponse, error) {
	return nf.node.GetProof(rootHash, address)
}

// GetProofDataTrie returns the Merkle Proof for the given address, and another Merkle Proof
// for the given key, if it exists in the dataTrie
func (nf *NodeFacade) GetProofDataTrie(rootHash string, address string, key string) (*common.GetProofResponse, *common.GetProofResponse, error) {
	return nf.node.GetProofDataTrie(rootHash, address, key)
}

// GetProofCurrentRootHash returns the Merkle proof for the given address and current root hash
func (nf *NodeFacade) GetProofCurrentRootHash(address string) (*common.GetProofResponse, error) {
	rootHash := nf.blockchain.GetCurrentBlockRootHash()
	if len(rootHash) == 0 {
		return nil, ErrEmptyRootHash
	}

	hexRootHash := hex.EncodeToString(rootHash)

	return nf.node.GetProof(hexRootHash, address)
}

// VerifyProof verifies the given Merkle proof
func (nf *NodeFacade) VerifyProof(rootHash string, address string, proof [][]byte) (bool, error) {
	return nf.node.VerifyProof(rootHash, address, proof)
}

// IsDataTrieMigrated returns true if the data trie for the given address is migrated
func (nf *NodeFacade) IsDataTrieMigrated(address string, options apiData.AccountQueryOptions) (bool, error) {
	return nf.node.IsDataTrieMigrated(address, options)
}

// GetManagedKeysCount returns the number of managed keys when node is running in multikey mode
func (nf *NodeFacade) GetManagedKeysCount() int {
	return nf.apiResolver.GetManagedKeysCount()
}

// GetManagedKeys returns all keys that should act as validator(main or backup that took over) and will be managed by this node
func (nf *NodeFacade) GetManagedKeys() []string {
	return nf.apiResolver.GetManagedKeys()
}

// GetLoadedKeys returns all keys that were loaded by this node
func (nf *NodeFacade) GetLoadedKeys() []string {
	return nf.apiResolver.GetLoadedKeys()
}

// GetEligibleManagedKeys returns the eligible managed keys when node is running in multikey mode
func (nf *NodeFacade) GetEligibleManagedKeys() ([]string, error) {
	return nf.apiResolver.GetEligibleManagedKeys()
}

// GetWaitingManagedKeys returns the waiting managed keys when node is running in multikey mode
func (nf *NodeFacade) GetWaitingManagedKeys() ([]string, error) {
	return nf.apiResolver.GetWaitingManagedKeys()
}

// GetWaitingEpochsLeftForPublicKey returns the number of epochs left for the public key until it becomes eligible
func (nf *NodeFacade) GetWaitingEpochsLeftForPublicKey(publicKey string) (uint32, error) {
	return nf.apiResolver.GetWaitingEpochsLeftForPublicKey(publicKey)
}

func (nf *NodeFacade) convertVmOutputToApiResponse(input *vmcommon.VMOutput) *vm.VMOutputApi {
	outputAccounts := make(map[string]*vm.OutputAccountApi)
	for key, acc := range input.OutputAccounts {
		outputAddress, err := nf.node.EncodeAddressPubkey(acc.GetAddress())
		if err != nil {
			log.Warn("cannot encode address", "error", err)
			outputAddress = ""
		}

		storageUpdates := make(map[string]*vm.StorageUpdateApi)
		for updateKey, updateVal := range acc.GetStorageUpdates() {
			storageUpdates[hex.EncodeToString([]byte(updateKey))] = &vm.StorageUpdateApi{
				Offset: updateVal.Offset,
				Data:   updateVal.Data,
			}
		}
		outKey := hex.EncodeToString([]byte(key))
		outAcc := &vm.OutputAccountApi{
			Address:        outputAddress,
			Nonce:          acc.GetNonce(),
			Balance:        acc.GetBalance(),
			BalanceDelta:   acc.GetBalanceDelta(),
			StorageUpdates: storageUpdates,
			Code:           acc.GetCode(),
			CodeMetadata:   acc.GetCodeMetadata(),
		}

		outAcc.OutputTransfers = make([]vm.OutputTransferApi, len(acc.GetOutputTransfers()))
		for i, outTransfer := range acc.GetOutputTransfers() {
			outTransferApi := vm.OutputTransferApi{
				Value:         outTransfer.Value,
				GasLimit:      outTransfer.GasLimit,
				Data:          outTransfer.Data,
				CallType:      outTransfer.CallType,
				SenderAddress: outputAddress,
			}

			if len(outTransfer.SenderAddress) == len(acc.GetAddress()) && !bytes.Equal(outTransfer.SenderAddress, acc.GetAddress()) {
				senderAddr, errEncode := nf.node.EncodeAddressPubkey(outTransfer.SenderAddress)
				if errEncode != nil {
					log.Warn("cannot encode address", "error", errEncode)
					senderAddr = outputAddress
				}
				outTransferApi.SenderAddress = senderAddr
			}

			outAcc.OutputTransfers[i] = outTransferApi
		}

		outputAccounts[outKey] = outAcc
	}

	logs := make([]*vm.LogEntryApi, 0, len(input.Logs))
	for i := 0; i < len(input.Logs); i++ {
		originalLog := input.Logs[i]
		logAddress, err := nf.node.EncodeAddressPubkey(originalLog.Address)
		if err != nil {
			log.Warn("cannot encode address", "error", err)
			logAddress = ""
		}

		logs = append(logs, &vm.LogEntryApi{
			Identifier:     originalLog.Identifier,
			Address:        logAddress,
			Topics:         originalLog.Topics,
			Data:           originalLog.GetFirstDataItem(),
			AdditionalData: originalLog.Data,
		})
	}

	return &vm.VMOutputApi{
		ReturnData:      input.ReturnData,
		ReturnCode:      input.ReturnCode.String(),
		ReturnMessage:   input.ReturnMessage,
		GasRemaining:    input.GasRemaining,
		GasRefund:       input.GasRefund,
		OutputAccounts:  outputAccounts,
		DeletedAccounts: input.DeletedAccounts,
		TouchedAccounts: input.TouchedAccounts,
		Logs:            logs,
	}
}

func queryBlockInfoToApiResource(info common.BlockInfo) apiData.BlockInfo {
	if check.IfNil(info) {
		return apiData.BlockInfo{}
	}

	return apiData.BlockInfo{
		Nonce:    info.GetNonce(),
		Hash:     hex.EncodeToString(info.GetHash()),
		RootHash: hex.EncodeToString(info.GetRootHash()),
	}
}

// GetGenesisNodesPubKeys will return genesis nodes public keys by shard
func (nf *NodeFacade) GetGenesisNodesPubKeys() (map[uint32][]string, map[uint32][]string, error) {
	eligible, waiting := nf.apiResolver.GetGenesisNodesPubKeys()

	if eligible == nil && waiting == nil {
		return nil, nil, ErrNilGenesisNodes
	}

	return eligible, waiting, nil
}

// GetGenesisBalances will return the balances minted on the genesis block
func (nf *NodeFacade) GetGenesisBalances() ([]*common.InitialAccountAPI, error) {
	initialAccounts, err := nf.apiResolver.GetGenesisBalances()
	if err != nil {
		return nil, err
	}
	if len(initialAccounts) == 0 {
		return nil, ErrNilGenesisBalances
	}

	return initialAccounts, nil
}

// GetGasConfigs will return currently using gas schedule configs
func (nf *NodeFacade) GetGasConfigs() (map[string]map[string]uint64, error) {
	gasConfigs := nf.apiResolver.GetGasConfigs()
	if len(gasConfigs) == 0 {
		return nil, ErrEmptyGasConfigs
	}

	return gasConfigs, nil
}

// P2PPrometheusMetricsEnabled returns if p2p prometheus metrics should be enabled or not on the application
func (nf *NodeFacade) P2PPrometheusMetricsEnabled() bool {
	return nf.config.P2PPrometheusMetricsEnabled
}

// IsInterfaceNil returns true if there is no value under the interface
func (nf *NodeFacade) IsInterfaceNil() bool {
	return nf == nil
}
