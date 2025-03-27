package incomingEventsProc

import (
	"encoding/hex"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/errors"
	sovBlock "github.com/multiversx/mx-chain-go/process/block/sovereign"
	"github.com/multiversx/mx-chain-go/process/block/sovereign/incomingHeader/dto"
)

type eventData struct {
	nonce                uint64
	functionCallWithArgs []byte
	gasLimit             uint64
}

// EventProcDepositTokensArgs holds necessary args for deposit event processor
type EventProcDepositTokensArgs struct {
	Marshaller    marshal.Marshalizer
	Hasher        hashing.Hasher
	DataCodec     sovBlock.DataCodecHandler
	TopicsChecker sovBlock.TopicsCheckerHandler
}

type eventProcDepositTokens struct {
	marshaller    marshal.Marshalizer
	hasher        hashing.Hasher
	dataCodec     sovBlock.DataCodecHandler
	topicsChecker sovBlock.TopicsCheckerHandler
}

// NewEventProcDepositTokens creates a new event processor for deposit token operations
func NewEventProcDepositTokens(args EventProcDepositTokensArgs) (*eventProcDepositTokens, error) {
	if check.IfNil(args.Marshaller) {
		return nil, core.ErrNilMarshalizer
	}
	if check.IfNil(args.Hasher) {
		return nil, core.ErrNilHasher
	}
	if check.IfNil(args.DataCodec) {
		return nil, errors.ErrNilDataCodec
	}
	if check.IfNil(args.TopicsChecker) {
		return nil, errors.ErrNilTopicsChecker
	}

	return &eventProcDepositTokens{
		marshaller:    args.Marshaller,
		hasher:        args.Hasher,
		dataCodec:     args.DataCodec,
		topicsChecker: args.TopicsChecker,
	}, nil
}

// ProcessEvent handles incoming token deposit events and returns the corresponding incoming SCR info.
// Each deposit event is identified by dto.EventIDDepositIncomingTransfer.
//
// Expected event data:
// - Data []byte – Serialized event details (nonce, gas, function call with arguments).
// - Topics [][]byte – A variable-length list of incoming token deposit topics, where:
//   - topic[0] = dto.TopicIDDepositIncomingTransfer.
//   - topic[1] = Receiver address.
//   - topic[2:N] = List of token data.
func (dep *eventProcDepositTokens) ProcessEvent(event data.EventHandler) (*dto.EventResult, error) {
	topics := event.GetTopics()
	err := dep.topicsChecker.CheckValidity(topics)
	if err != nil {
		return nil, err
	}

	receivedEventData, err := dep.createEventData(event.GetData())
	if err != nil {
		return nil, err
	}

	scrData, err := dep.createSCRData(topics)
	if err != nil {
		return nil, err
	}

	scrData = append(scrData, receivedEventData.functionCallWithArgs...)
	scr := &smartContractResult.SmartContractResult{
		Nonce:          receivedEventData.nonce,
		OriginalTxHash: nil, // TODO:  Implement this in MX-14321 task
		RcvAddr:        topics[1],
		SndAddr:        core.ESDTSCAddress,
		Data:           scrData,
		Value:          big.NewInt(0),
		GasLimit:       receivedEventData.gasLimit,
	}

	hash, err := core.CalculateHash(dep.marshaller, dep.hasher, scr)
	if err != nil {
		return nil, err
	}

	return &dto.EventResult{
		SCR: &dto.SCRInfo{
			SCR:  scr,
			Hash: hash,
		},
	}, nil
}

func (dep *eventProcDepositTokens) createEventData(data []byte) (*eventData, error) {
	evData, err := dep.dataCodec.DeserializeEventData(data)
	if err != nil {
		return nil, err
	}

	gasLimit, functionCallWithArgs := extractSCTransferInfo(evData.TransferData)
	return &eventData{
		nonce:                evData.Nonce,
		functionCallWithArgs: functionCallWithArgs,
		gasLimit:             gasLimit,
	}, nil
}

func extractSCTransferInfo(transferData *sovereign.TransferData) (uint64, []byte) {
	gasLimit := uint64(0)
	functionCallWithArgs := make([]byte, 0)
	if transferData != nil {
		gasLimit = transferData.GasLimit

		functionCallWithArgs = append(functionCallWithArgs, []byte("@")...)
		functionCallWithArgs = append(functionCallWithArgs, hex.EncodeToString(transferData.Function)...)
		functionCallWithArgs = append(functionCallWithArgs, extractArguments(transferData.Args)...)
	}

	return gasLimit, functionCallWithArgs
}

func extractArguments(arguments [][]byte) []byte {
	if len(arguments) == 0 {
		return make([]byte, 0)
	}

	args := make([]byte, 0)
	for _, arg := range arguments {
		args = append(args, []byte("@")...)
		args = append(args, hex.EncodeToString(arg)...)
	}

	return args
}

func (dep *eventProcDepositTokens) createSCRData(topics [][]byte) ([]byte, error) {
	numTokensToTransfer := len(topics[dto.TokensIndex:]) / dto.NumTransferTopics
	numTokensToTransferBytes := big.NewInt(int64(numTokensToTransfer)).Bytes()

	ret := []byte(core.BuiltInFunctionMultiESDTNFTTransfer +
		"@" + hex.EncodeToString(numTokensToTransferBytes))

	for idx := dto.TokensIndex; idx < len(topics); idx += dto.NumTransferTopics {
		tokenData, err := dep.getTokenDataBytes(topics[idx+1], topics[idx+2])
		if err != nil {
			return nil, err
		}

		transfer := []byte("@" +
			hex.EncodeToString(topics[idx]) + // tokenID
			"@" + hex.EncodeToString(topics[idx+1]) + // nonce
			"@" + hex.EncodeToString(tokenData)) // value/tokenData

		ret = append(ret, transfer...)
	}

	return ret, nil
}

func (dep *eventProcDepositTokens) getTokenDataBytes(tokenNonce []byte, tokenData []byte) ([]byte, error) {
	esdtTokenData, err := dep.dataCodec.DeserializeTokenData(tokenData)
	if err != nil {
		return nil, err
	}

	if esdtTokenData.TokenType == core.Fungible {
		return esdtTokenData.Amount.Bytes(), nil
	}

	nonce, err := common.ByteSliceToUint64(tokenNonce)
	if err != nil {
		return nil, err
	}

	digitalToken := &esdt.ESDigitalToken{
		Type:  uint32(esdtTokenData.TokenType),
		Value: esdtTokenData.Amount,
		TokenMetaData: &esdt.MetaData{
			Nonce:      nonce,
			Name:       esdtTokenData.Name,
			Creator:    esdtTokenData.Creator,
			Royalties:  uint32(esdtTokenData.Royalties.Uint64()),
			Hash:       esdtTokenData.Hash,
			URIs:       esdtTokenData.Uris,
			Attributes: esdtTokenData.Attributes,
		},
	}

	return dep.marshaller.Marshal(digitalToken)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (dep *eventProcDepositTokens) IsInterfaceNil() bool {
	return dep == nil
}
