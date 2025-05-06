package transaction

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/receipt"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/data/vm"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	logger "github.com/multiversx/mx-chain-logger-go"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/state"
)

var log = logger.GetOrCreate("process/transaction")
var _ process.TransactionProcessor = (*txProcessor)(nil)

// RefundGasMessage is the message returned in the data field of a receipt,
// for move balance transactions that provide more gas than needed
const RefundGasMessage = "refundedGas"

type relayedFees struct {
	totalFee, remainingFee, relayerFee *big.Int
}

// txProcessor implements TransactionProcessor interface and can modify account states according to a transaction
type txProcessor struct {
	*baseTxProcessor
	txFeeHandler        process.TransactionFeeHandler
	receiptForwarder    process.IntermediateTransactionHandler
	badTxForwarder      process.IntermediateTransactionHandler
	argsParser          process.ArgumentsParser
	scrForwarder        process.IntermediateTransactionHandler
	signMarshalizer     marshal.Marshalizer
	enableEpochsHandler common.EnableEpochsHandler
	txLogsProcessor     process.TransactionLogProcessor
}

// ArgsNewTxProcessor defines the arguments needed for new tx processor
type ArgsNewTxProcessor struct {
	Accounts            state.AccountsAdapter
	Hasher              hashing.Hasher
	PubkeyConv          core.PubkeyConverter
	Marshalizer         marshal.Marshalizer
	SignMarshalizer     marshal.Marshalizer
	ShardCoordinator    sharding.Coordinator
	ScProcessor         process.SmartContractProcessor
	TxFeeHandler        process.TransactionFeeHandler
	TxTypeHandler       process.TxTypeHandler
	EconomicsFee        process.FeeHandler
	ReceiptForwarder    process.IntermediateTransactionHandler
	BadTxForwarder      process.IntermediateTransactionHandler
	ArgsParser          process.ArgumentsParser
	ScrForwarder        process.IntermediateTransactionHandler
	EnableRoundsHandler process.EnableRoundsHandler
	EnableEpochsHandler common.EnableEpochsHandler
	TxVersionChecker    process.TxVersionCheckerHandler
	GuardianChecker     process.GuardianChecker
	TxLogsProcessor     process.TransactionLogProcessor
}

// NewTxProcessor creates a new txProcessor engine
func NewTxProcessor(args ArgsNewTxProcessor) (*txProcessor, error) {
	if check.IfNil(args.Accounts) {
		return nil, process.ErrNilAccountsAdapter
	}
	if check.IfNil(args.Hasher) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(args.PubkeyConv) {
		return nil, process.ErrNilPubkeyConverter
	}
	if check.IfNil(args.Marshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(args.ShardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	if check.IfNil(args.ScProcessor) {
		return nil, process.ErrNilSmartContractProcessor
	}
	if check.IfNil(args.TxFeeHandler) {
		return nil, process.ErrNilUnsignedTxHandler
	}
	if check.IfNil(args.TxTypeHandler) {
		return nil, process.ErrNilTxTypeHandler
	}
	if check.IfNil(args.EconomicsFee) {
		return nil, process.ErrNilEconomicsFeeHandler
	}
	if check.IfNil(args.ReceiptForwarder) {
		return nil, process.ErrNilReceiptHandler
	}
	if check.IfNil(args.BadTxForwarder) {
		return nil, process.ErrNilBadTxHandler
	}
	if check.IfNil(args.ArgsParser) {
		return nil, process.ErrNilArgumentParser
	}
	if check.IfNil(args.ScrForwarder) {
		return nil, process.ErrNilIntermediateTransactionHandler
	}
	if check.IfNil(args.SignMarshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(args.EnableRoundsHandler) {
		return nil, process.ErrNilEnableRoundsHandler
	}
	if check.IfNil(args.EnableEpochsHandler) {
		return nil, process.ErrNilEnableEpochsHandler
	}
	err := core.CheckHandlerCompatibility(args.EnableEpochsHandler, []core.EnableEpochFlag{
		common.PenalizedTooMuchGasFlag,
		common.MetaProtectionFlag,
		common.AddFailedRelayedTxToInvalidMBsFlag,
		common.RelayedTransactionsFlag,
		common.RelayedTransactionsV2Flag,
		common.RelayedNonceFixFlag,
		common.FixRelayedBaseCostFlag,
	})
	if err != nil {
		return nil, err
	}
	if check.IfNil(args.TxVersionChecker) {
		return nil, process.ErrNilTransactionVersionChecker
	}
	if check.IfNil(args.GuardianChecker) {
		return nil, process.ErrNilGuardianChecker
	}
	if check.IfNil(args.TxLogsProcessor) {
		return nil, process.ErrNilTxLogsProcessor
	}

	baseTxProcess := &baseTxProcessor{
		accounts:            args.Accounts,
		shardCoordinator:    args.ShardCoordinator,
		pubkeyConv:          args.PubkeyConv,
		economicsFee:        args.EconomicsFee,
		hasher:              args.Hasher,
		marshalizer:         args.Marshalizer,
		scProcessor:         args.ScProcessor,
		enableEpochsHandler: args.EnableEpochsHandler,
		txVersionChecker:    args.TxVersionChecker,
		guardianChecker:     args.GuardianChecker,
		txTypeHandler:       args.TxTypeHandler,
	}

	txProc := &txProcessor{
		baseTxProcessor:     baseTxProcess,
		txFeeHandler:        args.TxFeeHandler,
		receiptForwarder:    args.ReceiptForwarder,
		badTxForwarder:      args.BadTxForwarder,
		argsParser:          args.ArgsParser,
		scrForwarder:        args.ScrForwarder,
		signMarshalizer:     args.SignMarshalizer,
		enableEpochsHandler: args.EnableEpochsHandler,
		txLogsProcessor:     args.TxLogsProcessor,
	}

	return txProc, nil
}

// ProcessTransaction modifies the account states in respect with the transaction data
func (txProc *txProcessor) ProcessTransaction(tx data.TransactionHandler) (vmcommon.ReturnCode, error) {
	if check.IfNil(tx) {
		return 0, process.ErrNilTransaction
	}

	acntSnd, acntDst, err := txProc.getAccounts(tx.GetSndAddr(), tx.GetRcvAddr())
	if err != nil {
		return 0, err
	}

	txHash, err := core.CalculateHash(txProc.marshalizer, txProc.hasher, tx)
	if err != nil {
		return 0, err
	}

	process.DisplayProcessTxDetails(
		"ProcessTransaction: sender account details",
		acntSnd,
		tx,
		txHash,
		txProc.pubkeyConv,
	)

	txType, dstShardTxType, isRelayedV3 := txProc.txTypeHandler.ComputeTransactionType(tx)
	err = txProc.checkTxValues(tx, acntSnd, acntDst, false)
	if err != nil {
		if errors.Is(err, process.ErrInsufficientFunds) {
			receiptErr := txProc.executingFailedTransaction(tx, acntSnd, acntDst, err)
			if receiptErr != nil {
				return 0, receiptErr
			}
		}

		if errors.Is(err, process.ErrUserNameDoesNotMatch) && txProc.enableEpochsHandler.IsFlagEnabled(common.RelayedTransactionsFlag) {
			receiptErr := txProc.executingFailedTransaction(tx, acntSnd, acntDst, err)
			if receiptErr != nil {
				return vmcommon.UserError, receiptErr
			}
		}

		if errors.Is(err, process.ErrUserNameDoesNotMatchInCrossShardTx) {
			errProcessIfErr := txProc.processIfTxErrorCrossShard(tx, err.Error())
			if errProcessIfErr != nil {
				return 0, errProcessIfErr
			}
			return vmcommon.UserError, nil
		}
		return vmcommon.UserError, err
	}

	if isRelayedV3 {
		err = txProc.verifyRelayedTxV3(tx)
		if err != nil {
			return vmcommon.UserError, err
		}
	}

	switch txType {
	case process.MoveBalance:
		err = txProc.processMoveBalance(tx, acntSnd, acntDst, dstShardTxType, nil, false)
		if err != nil {
			return vmcommon.UserError, txProc.executeAfterFailedMoveBalanceTransaction(tx, err)
		}
		return vmcommon.Ok, err
	case process.SCDeployment:
		return txProc.processSCDeployment(tx, acntSnd)
	case process.SCInvoking:
		return txProc.processSCInvoking(tx, acntSnd, acntDst)
	case process.BuiltInFunctionCall:
		return txProc.processBuiltInFunctionCall(tx, acntSnd, acntDst)
	case process.RelayedTx:
		return txProc.processRelayedTx(tx, acntSnd, acntDst)
	case process.RelayedTxV2:
		return txProc.processRelayedTxV2(tx, acntSnd, acntDst)
	}

	return vmcommon.UserError, txProc.executingFailedTransaction(tx, acntSnd, acntDst, process.ErrWrongTransaction)
}

func (txProc *txProcessor) executeAfterFailedMoveBalanceTransaction(
	tx data.TransactionHandler,
	txError error,
) error {
	if core.IsGetNodeFromDBError(txError) {
		return txError
	}

	acntSnd, err := txProc.getAccountFromAddress(tx.GetSndAddr())
	if err != nil {
		return err
	}

	if errors.Is(txError, process.ErrInvalidMetaTransaction) || errors.Is(txError, process.ErrAccountNotPayable) {
		snapshot := txProc.accounts.JournalLen()
		var txHash []byte
		txHash, err = core.CalculateHash(txProc.marshalizer, txProc.hasher, tx)
		if err != nil {
			return err
		}

		err = txProc.scProcessor.ProcessIfError(acntSnd, txHash, tx, txError.Error(), nil, snapshot, 0)
		if err != nil {
			return err
		}

		if check.IfNil(acntSnd) {
			return nil
		}

		err = txProc.badTxForwarder.AddIntermediateTransactions([]data.TransactionHandler{tx}, txHash)
		if err != nil {
			return err
		}

		return process.ErrFailedTransaction
	}

	return txError
}

func (txProc *txProcessor) executingFailedTransaction(
	tx data.TransactionHandler,
	acntSnd state.UserAccountHandler,
	acntDst state.UserAccountHandler,
	txError error,
) error {
	if check.IfNil(acntSnd) {
		return nil
	}

	feePayer, isRelayedV3, err := txProc.getFeePayer(tx, acntSnd, acntDst)
	if err != nil {
		return err
	}

	txFee := txProc.economicsFee.ComputeTxFee(tx)
	err = feePayer.SubFromBalance(txFee)
	if err != nil {
		return err
	}

	txHash, err := core.CalculateHash(txProc.marshalizer, txProc.hasher, tx)
	if err != nil {
		return err
	}

	acntSnd.IncreaseNonce(1)
	err = txProc.badTxForwarder.AddIntermediateTransactions([]data.TransactionHandler{tx}, txHash)
	if err != nil {
		return err
	}

	log.Trace("executingFailedTransaction", "fail reason(error)", txError, "tx hash", txHash)

	rpt := &receipt.Receipt{
		Value:   big.NewInt(0).Set(txFee),
		SndAddr: tx.GetSndAddr(),
		Data:    []byte(txError.Error()),
		TxHash:  txHash,
	}

	err = txProc.receiptForwarder.AddIntermediateTransactions([]data.TransactionHandler{rpt}, txHash)
	if err != nil {
		return err
	}

	txProc.txFeeHandler.ProcessTransactionFee(txFee, big.NewInt(0), txHash)

	err = txProc.accounts.SaveAccount(feePayer)
	if err != nil {
		return err
	}

	if isRelayedV3 {
		// for relayed v3, the nonce was increased for sender, but fees consumed from relayer
		err = txProc.accounts.SaveAccount(acntSnd)
		if err != nil {
			return err
		}
	}

	return process.ErrFailedTransaction
}

func (txProc *txProcessor) createReceiptWithReturnedGas(
	txHash []byte,
	tx data.TransactionHandler,
	feePayer state.UserAccountHandler,
	moveBalanceCost *big.Int,
	totalProvided *big.Int,
	destShardTxType process.TransactionType,
	isUserTxOfRelayed bool,
) error {
	if check.IfNil(feePayer) || isUserTxOfRelayed {
		return nil
	}
	shouldCreateReceiptBackwardCompatible := !txProc.enableEpochsHandler.IsFlagEnabled(common.MetaProtectionFlag) && core.IsSmartContractAddress(tx.GetRcvAddr())
	if destShardTxType != process.MoveBalance || shouldCreateReceiptBackwardCompatible {
		return nil
	}

	refundValue := big.NewInt(0).Sub(totalProvided, moveBalanceCost)

	zero := big.NewInt(0)
	if refundValue.Cmp(zero) == 0 {
		return nil
	}

	rpt := &receipt.Receipt{
		Value:   big.NewInt(0).Set(refundValue),
		SndAddr: feePayer.AddressBytes(),
		Data:    []byte(RefundGasMessage),
		TxHash:  txHash,
	}

	return txProc.receiptForwarder.AddIntermediateTransactions([]data.TransactionHandler{rpt}, txHash)
}

func (txProc *txProcessor) processTxFee(
	tx data.TransactionHandler,
	feePayer, acntDst state.UserAccountHandler,
	dstShardTxType process.TransactionType,
	isUserTxOfRelayed bool,
) (*big.Int, *big.Int, error) {
	if check.IfNil(feePayer) {
		return big.NewInt(0), big.NewInt(0), nil
	}

	if isUserTxOfRelayed {
		totalCost := txProc.computeInnerTxFee(tx)

		err := feePayer.SubFromBalance(totalCost)
		if err != nil {
			return nil, nil, err
		}

		err = txProc.accounts.SaveAccount(feePayer)
		if err != nil {
			return nil, nil, err
		}

		if dstShardTxType == process.MoveBalance {
			return totalCost, totalCost, nil
		}

		moveBalanceGasLimit := txProc.economicsFee.ComputeGasLimit(tx)
		currentShardFee := txProc.economicsFee.ComputeFeeForProcessing(tx, moveBalanceGasLimit)
		if txProc.enableEpochsHandler.IsFlagEnabled(common.FixRelayedBaseCostFlag) {
			currentShardFee = txProc.economicsFee.ComputeMoveBalanceFee(tx)
		}

		return currentShardFee, totalCost, nil
	}

	moveBalanceFee := txProc.economicsFee.ComputeMoveBalanceFee(tx)
	totalCost := txProc.economicsFee.ComputeTxFee(tx)

	if !txProc.enableEpochsHandler.IsFlagEnabled(common.PenalizedTooMuchGasFlag) {
		totalCost = core.SafeMul(tx.GetGasLimit(), tx.GetGasPrice())
	}

	isCrossShardSCCall := check.IfNil(acntDst) && len(tx.GetData()) > 0 && core.IsSmartContractAddress(tx.GetRcvAddr())
	if dstShardTxType != process.MoveBalance ||
		(!txProc.enableEpochsHandler.IsFlagEnabled(common.MetaProtectionFlag) && isCrossShardSCCall) {

		err := feePayer.SubFromBalance(totalCost)
		if err != nil {
			return nil, nil, err
		}
	} else {
		err := feePayer.SubFromBalance(moveBalanceFee)
		if err != nil {
			return nil, nil, err
		}
	}

	err := txProc.accounts.SaveAccount(feePayer)
	if err != nil {
		return nil, nil, err
	}

	return moveBalanceFee, totalCost, nil
}

func (txProc *txProcessor) checkIfValidTxToMetaChain(tx data.TransactionHandler) error {

	destShardId := txProc.shardCoordinator.ComputeId(tx.GetRcvAddr())
	if destShardId != core.MetachainShardId {
		return nil
	}

	// it is not allowed to send transactions to metachain if those are not of type smart contract
	if len(tx.GetData()) == 0 {
		return process.ErrInvalidMetaTransaction
	}

	if txProc.enableEpochsHandler.IsFlagEnabled(common.MetaProtectionFlag) {
		// additional check
		if tx.GetGasLimit() < txProc.economicsFee.ComputeGasLimit(tx)+core.MinMetaTxExtraGasCost {
			return fmt.Errorf("%w: not enough gas", process.ErrInvalidMetaTransaction)
		}
	}

	return nil
}

func (txProc *txProcessor) processMoveBalance(
	tx data.TransactionHandler,
	acntSrc, acntDst state.UserAccountHandler,
	destShardTxType process.TransactionType,
	originalTxHash []byte,
	isUserTxOfRelayed bool,
) error {

	feePayer, _, err := txProc.getFeePayer(tx, acntSrc, acntDst)
	if err != nil {
		return nil
	}
	moveBalanceCost, totalCost, err := txProc.processTxFee(tx, feePayer, acntDst, destShardTxType, isUserTxOfRelayed)
	if err != nil {
		return err
	}

	// is sender address in node shard
	if !check.IfNil(acntSrc) {
		acntSrc.IncreaseNonce(1)
		err = acntSrc.SubFromBalance(tx.GetValue())
		if err != nil {
			return err
		}

		err = txProc.accounts.SaveAccount(acntSrc)
		if err != nil {
			return err
		}
	}

	isPayable, err := txProc.scProcessor.IsPayable(tx.GetSndAddr(), tx.GetRcvAddr())
	if err != nil {
		errRefund := txProc.revertConsumedValueFromSender(tx, acntSrc, isUserTxOfRelayed)
		if errRefund != nil {
			log.Error("failed to return funds to sender after check if receiver is payable", "error", errRefund)
		}
		return err
	}
	if !isPayable {
		err = txProc.revertConsumedValueFromSender(tx, acntSrc, isUserTxOfRelayed)
		if err != nil {
			log.Error("failed to return funds to sender while transferring to non payable sc", "error", err)
		}

		return process.ErrAccountNotPayable
	}

	err = txProc.checkIfValidTxToMetaChain(tx)
	if err != nil {
		errLocal := txProc.revertConsumedValueFromSender(tx, acntSrc, isUserTxOfRelayed)
		if errLocal != nil {
			log.Error("failed to return funds to sender while sending invalid tx to metachain", "error", errLocal)
		}

		return err
	}

	// is receiver address in node shard
	if !check.IfNil(acntDst) {
		err = acntDst.AddToBalance(tx.GetValue())
		if err != nil {
			return err
		}

		err = txProc.accounts.SaveAccount(acntDst)
		if err != nil {
			return err
		}
	}

	txHash, err := core.CalculateHash(txProc.marshalizer, txProc.hasher, tx)
	if err != nil {
		return err
	}

	err = txProc.createReceiptWithReturnedGas(txHash, tx, feePayer, moveBalanceCost, totalCost, destShardTxType, isUserTxOfRelayed)
	if err != nil {
		return err
	}

	if isUserTxOfRelayed {
		txProc.txFeeHandler.ProcessTransactionFeeRelayedUserTx(moveBalanceCost, big.NewInt(0), txHash, originalTxHash)
	} else {
		txProc.txFeeHandler.ProcessTransactionFee(moveBalanceCost, big.NewInt(0), txHash)
	}

	return nil
}

func (txProc *txProcessor) revertConsumedValueFromSender(
	tx data.TransactionHandler,
	acntSrc state.UserAccountHandler,
	isUserTxOfRelayed bool,
) error {
	if !isUserTxOfRelayed {
		return nil
	}

	if !txProc.enableEpochsHandler.IsFlagEnabled(common.FixRelayedMoveBalanceToNonPayableSCFlag) {
		return nil
	}

	if check.IfNil(acntSrc) {
		return nil
	}

	err := acntSrc.AddToBalance(tx.GetValue())
	if err != nil {
		return err
	}

	return txProc.accounts.SaveAccount(acntSrc)
}

func (txProc *txProcessor) processSCDeployment(
	tx data.TransactionHandler,
	acntSrc state.UserAccountHandler,
) (vmcommon.ReturnCode, error) {
	return txProc.scProcessor.DeploySmartContract(tx, acntSrc)
}

func (txProc *txProcessor) processSCInvoking(
	tx data.TransactionHandler,
	acntSrc, acntDst state.UserAccountHandler,
) (vmcommon.ReturnCode, error) {
	return txProc.scProcessor.ExecuteSmartContractTransaction(tx, acntSrc, acntDst)
}

func (txProc *txProcessor) processBuiltInFunctionCall(
	tx data.TransactionHandler,
	acntSrc, acntDst state.UserAccountHandler,
) (vmcommon.ReturnCode, error) {
	return txProc.scProcessor.ExecuteBuiltInFunction(tx, acntSrc, acntDst)
}

func makeUserTxFromRelayedTxV2Args(args [][]byte) *transaction.Transaction {
	userTx := &transaction.Transaction{}
	userTx.RcvAddr = args[0]
	userTx.Nonce = big.NewInt(0).SetBytes(args[1]).Uint64()
	userTx.Data = args[2]
	userTx.Signature = args[3]
	userTx.Value = big.NewInt(0)
	return userTx
}

func (txProc *txProcessor) finishExecutionOfRelayedTx(
	relayerAcnt, acntDst state.UserAccountHandler,
	tx data.TransactionHandler,
	userTx data.TransactionHandler,
) (vmcommon.ReturnCode, error) {
	computedFees := txProc.computeRelayedTxFees(tx, userTx)
	txHash, err := txProc.processTxAtRelayer(
		relayerAcnt,
		computedFees.totalFee,
		computedFees.relayerFee,
		tx,
		tx.GetValue())
	if err != nil {
		return 0, err
	}

	if check.IfNil(acntDst) {
		return vmcommon.Ok, nil
	}

	err = txProc.addFeeAndValueToDest(acntDst, tx.GetValue(), computedFees.remainingFee)
	if err != nil {
		return 0, err
	}

	return txProc.processUserTx(tx, userTx, tx.GetValue(), tx.GetNonce(), tx.GetSndAddr(), txHash)
}

func (txProc *txProcessor) processTxAtRelayer(
	relayerAcnt state.UserAccountHandler,
	totalFee *big.Int,
	relayerFee *big.Int,
	tx data.TransactionHandler,
	valueToSubFromRelayer *big.Int,
) ([]byte, error) {
	txHash, err := core.CalculateHash(txProc.marshalizer, txProc.hasher, tx)
	if err != nil {
		return nil, err
	}

	if !check.IfNil(relayerAcnt) {
		err = relayerAcnt.SubFromBalance(valueToSubFromRelayer)
		if err != nil {
			return nil, err
		}

		err = relayerAcnt.SubFromBalance(totalFee)
		if err != nil {
			return nil, err
		}

		relayerAcnt.IncreaseNonce(1)
		err = txProc.accounts.SaveAccount(relayerAcnt)
		if err != nil {
			return nil, err
		}

		txProc.txFeeHandler.ProcessTransactionFee(relayerFee, big.NewInt(0), txHash)
	}

	return txHash, nil
}

func (txProc *txProcessor) addFeeAndValueToDest(acntDst state.UserAccountHandler, txValue *big.Int, remainingFee *big.Int) error {
	if check.IfNil(acntDst) {
		return nil
	}

	err := acntDst.AddToBalance(txValue)
	if err != nil {
		return err
	}

	err = acntDst.AddToBalance(remainingFee)
	if err != nil {
		return err
	}

	return txProc.accounts.SaveAccount(acntDst)
}

func (txProc *txProcessor) verifyRelayedTxV3(tx data.TransactionHandler) error {
	if !txProc.enableEpochsHandler.IsFlagEnabled(common.RelayedTransactionsV3Flag) {
		return fmt.Errorf("%w, %s", process.ErrTransactionNotExecutable, process.ErrRelayedTxV3Disabled)
	}

	if !txProc.shardCoordinator.SameShard(tx.GetRelayerAddr(), tx.GetSndAddr()) {
		return fmt.Errorf("%w, %s", process.ErrTransactionNotExecutable, process.ErrShardIdMissmatch)
	}

	if bytes.Equal(tx.GetRelayerAddr(), tx.GetGuardianAddr()) {
		return fmt.Errorf("%w, %s", process.ErrTransactionNotExecutable, process.ErrRelayedByGuardianNotAllowed)
	}

	relayerAccount, err := txProc.getAccountFromAddress(tx.GetRelayerAddr())
	if err != nil {
		return fmt.Errorf("%w, %s", process.ErrTransactionNotExecutable, err)
	}

	if !check.IfNil(relayerAccount) && relayerAccount.IsGuarded() {
		return fmt.Errorf("%w, %s", process.ErrTransactionNotExecutable, process.ErrGuardedRelayerNotAllowed)
	}

	return nil
}

func (txProc *txProcessor) processRelayedTxV2(
	tx data.TransactionHandler,
	relayerAcnt, acntDst state.UserAccountHandler,
) (vmcommon.ReturnCode, error) {
	if !txProc.enableEpochsHandler.IsFlagEnabled(common.RelayedTransactionsV2Flag) {
		return vmcommon.UserError, txProc.executingFailedTransaction(tx, relayerAcnt, acntDst, process.ErrRelayedTxV2Disabled)
	}
	if tx.GetValue().Cmp(big.NewInt(0)) != 0 {
		return vmcommon.UserError, txProc.executingFailedTransaction(tx, relayerAcnt, acntDst, process.ErrRelayedTxV2ZeroVal)
	}

	_, args, err := txProc.argsParser.ParseCallData(string(tx.GetData()))
	if err != nil {
		return vmcommon.UserError, txProc.executingFailedTransaction(tx, relayerAcnt, acntDst, err)
	}
	if len(args) != 4 {
		return vmcommon.UserError, txProc.executingFailedTransaction(tx, relayerAcnt, acntDst, process.ErrInvalidArguments)
	}

	userTx := makeUserTxFromRelayedTxV2Args(args)
	userTx.GasPrice = tx.GetGasPrice()
	userTx.GasLimit = tx.GetGasLimit() - txProc.economicsFee.ComputeGasLimit(tx)
	userTx.SndAddr = tx.GetRcvAddr()

	return txProc.finishExecutionOfRelayedTx(relayerAcnt, acntDst, tx, userTx)
}

func (txProc *txProcessor) processRelayedTx(
	tx data.TransactionHandler,
	relayerAcnt, acntDst state.UserAccountHandler,
) (vmcommon.ReturnCode, error) {

	_, args, err := txProc.argsParser.ParseCallData(string(tx.GetData()))
	if err != nil {
		return 0, err
	}
	if len(args) != 1 {
		return vmcommon.UserError, txProc.executingFailedTransaction(tx, relayerAcnt, acntDst, process.ErrInvalidArguments)
	}
	if !txProc.enableEpochsHandler.IsFlagEnabled(common.RelayedTransactionsFlag) {
		return vmcommon.UserError, txProc.executingFailedTransaction(tx, relayerAcnt, acntDst, process.ErrRelayedTxDisabled)
	}

	userTx := &transaction.Transaction{}
	err = txProc.signMarshalizer.Unmarshal(userTx, args[0])
	if err != nil {
		return vmcommon.UserError, txProc.executingFailedTransaction(tx, relayerAcnt, acntDst, err)
	}
	if !bytes.Equal(userTx.SndAddr, tx.GetRcvAddr()) {
		return vmcommon.UserError, txProc.executingFailedTransaction(tx, relayerAcnt, acntDst, process.ErrRelayedTxBeneficiaryDoesNotMatchReceiver)
	}

	if userTx.Value.Cmp(tx.GetValue()) < 0 {
		return vmcommon.UserError, txProc.executingFailedTransaction(tx, relayerAcnt, acntDst, process.ErrRelayedTxValueHigherThenUserTxValue)
	}
	if userTx.GasPrice != tx.GetGasPrice() {
		return vmcommon.UserError, txProc.executingFailedTransaction(tx, relayerAcnt, acntDst, process.ErrRelayedGasPriceMissmatch)
	}

	remainingGasLimit := tx.GetGasLimit() - txProc.economicsFee.ComputeGasLimit(tx)
	if userTx.GasLimit != remainingGasLimit {
		return vmcommon.UserError, txProc.executingFailedTransaction(tx, relayerAcnt, acntDst, process.ErrRelayedTxGasLimitMissmatch)
	}

	return txProc.finishExecutionOfRelayedTx(relayerAcnt, acntDst, tx, userTx)
}

func (txProc *txProcessor) computeRelayedTxFees(tx, userTx data.TransactionHandler) relayedFees {
	relayerFee := txProc.economicsFee.ComputeMoveBalanceFee(tx)
	totalFee := txProc.economicsFee.ComputeTxFee(tx)
	if txProc.enableEpochsHandler.IsFlagEnabled(common.FixRelayedBaseCostFlag) {
		userFee := txProc.computeInnerTxFeeAfterBaseCostFix(userTx)

		totalFee = totalFee.Add(relayerFee, userFee)
	}
	remainingFee := big.NewInt(0).Sub(totalFee, relayerFee)

	computedFees := relayedFees{
		totalFee:     totalFee,
		remainingFee: remainingFee,
		relayerFee:   relayerFee,
	}

	return computedFees
}

func (txProc *txProcessor) removeValueAndConsumedFeeFromUser(
	userTx data.TransactionHandler,
	relayedTxValue *big.Int,
	originalTxHash []byte,
	originalTx data.TransactionHandler,
	executionErr error,
) error {
	userAcnt, err := txProc.getAccountFromAddress(userTx.GetSndAddr())
	if err != nil {
		return err
	}
	if check.IfNil(userAcnt) {
		return process.ErrNilUserAccount
	}
	err = userAcnt.SubFromBalance(relayedTxValue)
	if err != nil {
		return err
	}

	consumedFee := txProc.computeInnerTxFee(userTx)

	err = userAcnt.SubFromBalance(consumedFee)
	if err != nil {
		return err
	}

	if txProc.shouldIncreaseNonce(executionErr) {
		userAcnt.IncreaseNonce(1)
	}

	err = txProc.addNonExecutableLog(executionErr, originalTxHash, originalTx)
	if err != nil {
		return err
	}

	err = txProc.accounts.SaveAccount(userAcnt)
	if err != nil {
		return err
	}

	return nil
}

func (txProc *txProcessor) addNonExecutableLog(executionErr error, originalTxHash []byte, originalTx data.TransactionHandler) error {
	if !isNonExecutableError(executionErr) {
		return nil
	}

	logEntry := &vmcommon.LogEntry{
		Identifier: []byte(core.SignalErrorOperation),
		Address:    originalTx.GetRcvAddr(),
	}

	return txProc.txLogsProcessor.SaveLog(originalTxHash, originalTx, []*vmcommon.LogEntry{logEntry})

}

func (txProc *txProcessor) processMoveBalanceCostRelayedUserTx(
	userTx data.TransactionHandler,
	userScr *smartContractResult.SmartContractResult,
	userAcc state.UserAccountHandler,
	originalTxHash []byte,
) error {
	moveBalanceGasLimit := txProc.economicsFee.ComputeGasLimit(userTx)
	moveBalanceUserFee := txProc.economicsFee.ComputeFeeForProcessing(userTx, moveBalanceGasLimit)
	if txProc.enableEpochsHandler.IsFlagEnabled(common.FixRelayedBaseCostFlag) {
		moveBalanceUserFee = txProc.economicsFee.ComputeMoveBalanceFee(userTx)
	}

	userScrHash, err := core.CalculateHash(txProc.marshalizer, txProc.hasher, userScr)
	if err != nil {
		return err
	}

	txProc.txFeeHandler.ProcessTransactionFeeRelayedUserTx(moveBalanceUserFee, big.NewInt(0), userScrHash, originalTxHash)
	return userAcc.SubFromBalance(moveBalanceUserFee)
}

func (txProc *txProcessor) processUserTx(
	originalTx data.TransactionHandler,
	userTx data.TransactionHandler,
	relayedTxValue *big.Int,
	relayedNonce uint64,
	relayerAddr []byte,
	originalTxHash []byte,
) (vmcommon.ReturnCode, error) {

	acntSnd, acntDst, err := txProc.getAccounts(userTx.GetSndAddr(), userTx.GetRcvAddr())
	if err != nil {
		errRemove := txProc.removeValueAndConsumedFeeFromUser(userTx, relayedTxValue, originalTxHash, originalTx, err)
		if errRemove != nil {
			return vmcommon.UserError, errRemove
		}
		return vmcommon.UserError, txProc.executeFailedRelayedUserTx(
			userTx,
			relayerAddr,
			relayedTxValue,
			relayedNonce,
			originalTx,
			originalTxHash,
			err.Error())
	}

	txType, dstShardTxType, _ := txProc.txTypeHandler.ComputeTransactionType(userTx)
	err = txProc.checkTxValues(userTx, acntSnd, acntDst, true)
	if err != nil {
		errRemove := txProc.removeValueAndConsumedFeeFromUser(userTx, relayedTxValue, originalTxHash, originalTx, err)
		if errRemove != nil {
			return vmcommon.UserError, errRemove
		}
		return vmcommon.UserError, txProc.executeFailedRelayedUserTx(
			userTx,
			relayerAddr,
			relayedTxValue,
			relayedNonce,
			originalTx,
			originalTxHash,
			err.Error())
	}

	scrFromTx, err := txProc.makeSCRFromUserTx(userTx, relayerAddr, relayedTxValue, originalTxHash)
	if err != nil {
		return 0, err
	}

	returnCode := vmcommon.Ok
	switch txType {
	case process.MoveBalance:
		err = txProc.processMoveBalance(userTx, acntSnd, acntDst, dstShardTxType, originalTxHash, true)
	case process.SCDeployment:
		err = txProc.processMoveBalanceCostRelayedUserTx(userTx, scrFromTx, acntSnd, originalTxHash)
		if err != nil {
			break
		}

		returnCode, err = txProc.scProcessor.DeploySmartContract(scrFromTx, acntSnd)
	case process.SCInvoking:
		err = txProc.processMoveBalanceCostRelayedUserTx(userTx, scrFromTx, acntSnd, originalTxHash)
		if err != nil {
			break
		}

		returnCode, err = txProc.scProcessor.ExecuteSmartContractTransaction(scrFromTx, acntSnd, acntDst)
	case process.BuiltInFunctionCall:
		err = txProc.processMoveBalanceCostRelayedUserTx(userTx, scrFromTx, acntSnd, originalTxHash)
		if err != nil {
			break
		}

		returnCode, err = txProc.scProcessor.ExecuteBuiltInFunction(scrFromTx, acntSnd, acntDst)
	default:
		err = process.ErrWrongTransaction
		errRemove := txProc.removeValueAndConsumedFeeFromUser(userTx, relayedTxValue, originalTxHash, originalTx, err)
		if errRemove != nil {
			return vmcommon.UserError, errRemove
		}
		return vmcommon.UserError, txProc.executeFailedRelayedUserTx(
			userTx,
			relayerAddr,
			relayedTxValue,
			relayedNonce,
			originalTx,
			originalTxHash,
			err.Error())
	}

	if errors.Is(err, process.ErrInvalidMetaTransaction) || errors.Is(err, process.ErrAccountNotPayable) {
		return vmcommon.UserError, txProc.executeFailedRelayedUserTx(
			userTx,
			relayerAddr,
			relayedTxValue,
			relayedNonce,
			originalTx,
			originalTxHash,
			err.Error())
	}

	if errors.Is(err, process.ErrFailedTransaction) {
		// in case of failed inner user tx transaction we should just simply return execution failed and
		// not failed transaction - as the actual transaction (the relayed we correctly executed) and thus
		// it should not lend in the invalid miniblock
		return vmcommon.ExecutionFailed, nil
	}

	if err != nil {
		log.Error("processUserTx", "protocolError", err)
		return vmcommon.ExecutionFailed, err
	}

	// no need to add the smart contract result From TX to the intermediate transactions in case of error
	// returning value is resolved inside smart contract processor or above by executeFailedRelayedUserTx
	if returnCode != vmcommon.Ok {
		return returnCode, nil
	}

	err = txProc.scrForwarder.AddIntermediateTransactions([]data.TransactionHandler{scrFromTx}, originalTxHash)
	if err != nil {
		return 0, err
	}

	return vmcommon.Ok, nil
}

func (txProc *baseTxProcessor) isCrossTxFromMe(adrSrc, adrDst []byte) bool {
	shardForSrc := txProc.shardCoordinator.ComputeId(adrSrc)
	shardForDst := txProc.shardCoordinator.ComputeId(adrDst)
	shardForCurrentNode := txProc.shardCoordinator.SelfId()

	srcInShard := shardForSrc == shardForCurrentNode
	dstInShard := shardForDst == shardForCurrentNode

	return srcInShard && !dstInShard
}

func (txProc *txProcessor) makeSCRFromUserTx(
	tx data.TransactionHandler,
	relayerAdr []byte,
	relayedTxValue *big.Int,
	txHash []byte,
) (*smartContractResult.SmartContractResult, error) {
	scr := &smartContractResult.SmartContractResult{
		Nonce:          tx.GetNonce(),
		Value:          tx.GetValue(),
		RcvAddr:        tx.GetRcvAddr(),
		SndAddr:        tx.GetSndAddr(),
		RelayerAddr:    relayerAdr,
		RelayedValue:   big.NewInt(0).Set(relayedTxValue),
		Data:           tx.GetData(),
		PrevTxHash:     txHash,
		OriginalTxHash: txHash,
		GasLimit:       tx.GetGasLimit(),
		GasPrice:       tx.GetGasPrice(),
		CallType:       vm.DirectCall,
	}

	var err error
	scr.GasLimit, err = core.SafeSubUint64(scr.GasLimit, txProc.economicsFee.ComputeGasLimit(tx))
	if err != nil {
		return nil, err
	}

	return scr, nil
}

func (txProc *txProcessor) executeFailedRelayedUserTx(
	userTx data.TransactionHandler,
	relayerAdr []byte,
	relayedTxValue *big.Int,
	relayedNonce uint64,
	originalTx data.TransactionHandler,
	originalTxHash []byte,
	errorMsg string,
) error {

	scrForRelayer := &smartContractResult.SmartContractResult{
		Nonce:          relayedNonce,
		Value:          big.NewInt(0).Set(relayedTxValue),
		RcvAddr:        relayerAdr,
		SndAddr:        userTx.GetSndAddr(),
		PrevTxHash:     originalTxHash,
		OriginalTxHash: originalTxHash,
		ReturnMessage:  []byte(errorMsg),
	}

	relayerAcnt, err := txProc.getAccountFromAddress(relayerAdr)
	if err != nil {
		return err
	}

	err = txProc.scrForwarder.AddIntermediateTransactions([]data.TransactionHandler{scrForRelayer}, originalTxHash)
	if err != nil {
		return err
	}

	totalFee := txProc.economicsFee.ComputeFeeForProcessing(userTx, userTx.GetGasLimit())
	moveBalanceGasLimit := txProc.economicsFee.ComputeGasLimit(userTx)
	gasToUse := userTx.GetGasLimit() - moveBalanceGasLimit
	processingUserFee := txProc.economicsFee.ComputeFeeForProcessing(userTx, gasToUse)
	if txProc.enableEpochsHandler.IsFlagEnabled(common.FixRelayedBaseCostFlag) {
		moveBalanceUserFee := txProc.economicsFee.ComputeMoveBalanceFee(userTx)
		totalFee = big.NewInt(0).Add(moveBalanceUserFee, processingUserFee)
	}

	senderShardID := txProc.shardCoordinator.ComputeId(userTx.GetSndAddr())
	if senderShardID != txProc.shardCoordinator.SelfId() {
		if txProc.enableEpochsHandler.IsFlagEnabled(common.FixRelayedBaseCostFlag) {
			totalFee.Sub(totalFee, processingUserFee)
		} else {
			moveBalanceUserFee := txProc.economicsFee.ComputeFeeForProcessing(userTx, moveBalanceGasLimit)
			totalFee.Sub(totalFee, moveBalanceUserFee)
		}
	}

	txProc.txFeeHandler.ProcessTransactionFee(totalFee, big.NewInt(0), originalTxHash)

	if !check.IfNil(relayerAcnt) {
		err = relayerAcnt.AddToBalance(scrForRelayer.Value)
		if err != nil {
			return err
		}

		if txProc.enableEpochsHandler.IsFlagEnabled(common.AddFailedRelayedTxToInvalidMBsFlag) {
			err = txProc.badTxForwarder.AddIntermediateTransactions([]data.TransactionHandler{originalTx}, originalTxHash)
			if err != nil {
				return err
			}
		}

		err = txProc.accounts.SaveAccount(relayerAcnt)
		if err != nil {
			return err
		}
	}

	return nil
}

func (txProc *txProcessor) shouldIncreaseNonce(executionErr error) bool {
	if !txProc.enableEpochsHandler.IsFlagEnabled(common.RelayedNonceFixFlag) {
		return true
	}

	if isNonExecutableError(executionErr) {
		return false
	}

	return true
}

func isNonExecutableError(executionErr error) bool {
	return errors.Is(executionErr, process.ErrLowerNonceInTransaction) ||
		errors.Is(executionErr, process.ErrHigherNonceInTransaction) ||
		errors.Is(executionErr, process.ErrTransactionNotExecutable)
}

// IsInterfaceNil returns true if there is no value under the interface
func (txProc *txProcessor) IsInterfaceNil() bool {
	return txProc == nil
}
