package scToProtocol

import (
	"bytes"
	"encoding/hex"
	"math"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-logger-go"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/vm"
	"github.com/multiversx/mx-chain-go/vm/systemSmartContracts"
)

var _ process.SmartContractToProtocolHandler = (*stakingToPeer)(nil)

var log = logger.GetOrCreate("process/scToProtocol")

// ArgStakingToPeer is struct that contain all components that are needed to create a new stakingToPeer object
type ArgStakingToPeer struct {
	PubkeyConv          core.PubkeyConverter
	Hasher              hashing.Hasher
	Marshalizer         marshal.Marshalizer
	PeerState           state.AccountsAdapter
	BaseState           state.AccountsAdapter
	ArgParser           process.ArgumentsParser
	CurrTxs             dataRetriever.TransactionCacher
	RatingsData         process.RatingsInfoHandler
	EnableEpochsHandler common.EnableEpochsHandler
}

// stakingToPeer defines the component which will translate changes from staking SC state
// to validator statistics trie
type stakingToPeer struct {
	pubkeyConv                      core.PubkeyConverter
	hasher                          hashing.Hasher
	marshalizer                     marshal.Marshalizer
	peerState                       state.AccountsAdapter
	baseState                       state.AccountsAdapter
	argParser                       process.ArgumentsParser
	currTxs                         dataRetriever.TransactionCacher
	startRating                     uint32
	unJailRating                    uint32
	jailRating                      uint32
	enableEpochsHandler             common.EnableEpochsHandler
	modifiedMBShardIDCheckerHandler modifiedMBShardIDCheckerHandler
}

// NewStakingToPeer creates the component which moves from staking sc state to peer state
func NewStakingToPeer(args ArgStakingToPeer) (*stakingToPeer, error) {
	err := checkIfNil(args)
	if err != nil {
		return nil, err
	}

	st := &stakingToPeer{
		pubkeyConv:                      args.PubkeyConv,
		hasher:                          args.Hasher,
		marshalizer:                     args.Marshalizer,
		peerState:                       args.PeerState,
		baseState:                       args.BaseState,
		argParser:                       args.ArgParser,
		currTxs:                         args.CurrTxs,
		startRating:                     args.RatingsData.StartRating(),
		unJailRating:                    args.RatingsData.StartRating(),
		jailRating:                      args.RatingsData.MinRating(),
		enableEpochsHandler:             args.EnableEpochsHandler,
		modifiedMBShardIDCheckerHandler: &modifiedMBShardIDChecker{},
	}

	return st, nil
}

func checkIfNil(args ArgStakingToPeer) error {
	if check.IfNil(args.PubkeyConv) {
		return process.ErrNilPubkeyConverter
	}
	if check.IfNil(args.Hasher) {
		return process.ErrNilHasher
	}
	if check.IfNil(args.Marshalizer) {
		return process.ErrNilMarshalizer
	}
	if check.IfNil(args.PeerState) {
		return process.ErrNilPeerAccountsAdapter
	}
	if check.IfNil(args.BaseState) {
		return process.ErrNilAccountsAdapter
	}
	if check.IfNil(args.ArgParser) {
		return process.ErrNilArgumentParser
	}
	if check.IfNil(args.CurrTxs) {
		return process.ErrNilTxForCurrentBlockHandler
	}
	if check.IfNil(args.RatingsData) {
		return process.ErrNilRatingsInfoHandler
	}
	if check.IfNil(args.EnableEpochsHandler) {
		return process.ErrNilEnableEpochsHandler
	}
	return core.CheckHandlerCompatibility(args.EnableEpochsHandler, []core.EnableEpochFlag{
		common.StakeFlag,
		common.ValidatorToDelegationFlag,
		common.UnJailCleanupFlag,
	})
}

func (stp *stakingToPeer) getPeerAccount(key []byte) (state.PeerAccountHandler, error) {
	account, err := stp.peerState.LoadAccount(key)
	if err != nil {
		return nil, err
	}

	peerAcc, ok := account.(state.PeerAccountHandler)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	return peerAcc, nil
}

func (stp *stakingToPeer) getUserAccount(key []byte) (state.UserAccountHandler, error) {
	account, err := stp.baseState.LoadAccount(key)
	if err != nil {
		return nil, err
	}

	userAcc, ok := account.(state.UserAccountHandler)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	return userAcc, nil
}

func (stp *stakingToPeer) getStorageFromAccount(userAcc state.UserAccountHandler, key []byte) []byte {
	value, _, err := userAcc.RetrieveValue(key)
	if err != nil {
		return nil
	}
	return value
}

// UpdateProtocol applies changes from staking smart contract to peer state and creates the actual peer changes
func (stp *stakingToPeer) UpdateProtocol(body *block.Body, nonce uint64) error {
	affectedStates, err := stp.getAllModifiedStates(body)
	if err != nil {
		return err
	}

	if len(affectedStates) == 0 {
		return nil
	}

	stakingSCAccount, err := stp.getUserAccount(vm.StakingSCAddress)
	if err != nil {
		return err
	}

	for _, key := range affectedStates {
		if len(key) != stp.pubkeyConv.Len() {
			continue
		}

		blsPubKey := []byte(key)
		log.Trace("get on StakingScAddress called", "blsKey", blsPubKey)

		buff := stp.getStorageFromAccount(stakingSCAccount, blsPubKey)
		// no data under key -> peer can be deleted from trie
		var existingAcc vmcommon.AccountHandler
		existingAcc, err = stp.peerState.GetExistingAccount(blsPubKey)
		shouldDeleteAccount := len(buff) == 0 && !check.IfNil(existingAcc) && err == nil
		if shouldDeleteAccount {
			err = stp.peerState.RemoveAccount(blsPubKey)
			if err != nil {
				log.Debug("staking to protocol RemoveAccount", "error", err, "blsPubKey", hex.EncodeToString(blsPubKey))
				continue
			}
			log.Debug("remove account from validator statistics", "blsPubKey", blsPubKey)

			continue
		}

		if len(buff) == 0 {
			continue
		}

		var stakingData systemSmartContracts.StakedDataV2_0
		err = stp.marshalizer.Unmarshal(&stakingData, buff)
		if err != nil {
			return err
		}

		err = stp.updatePeerState(stakingData, blsPubKey, nonce)
		if err != nil {
			return err
		}
	}

	return nil
}

func (stp *stakingToPeer) updatePeerStateV1(
	stakingData systemSmartContracts.StakedDataV2_0,
	blsPubKey []byte,
	nonce uint64,
) error {
	if stakingData.StakedNonce == math.MaxUint64 {
		return nil
	}

	account, err := stp.getPeerAccount(blsPubKey)
	if err != nil {
		return err
	}

	if !bytes.Equal(account.GetRewardAddress(), stakingData.RewardAddress) {
		err = account.SetRewardAddress(stakingData.RewardAddress)
		if err != nil {
			return err
		}
	}

	isValidator := account.GetList() == string(common.EligibleList) || account.GetList() == string(common.WaitingList)
	isJailed := stakingData.JailedNonce >= stakingData.UnJailedNonce && stakingData.JailedNonce > 0
	isStakingV4Started := stp.enableEpochsHandler.IsFlagEnabled(common.StakingV4StartedFlag)

	if !isJailed {
		if stakingData.StakedNonce == nonce && !isValidator {
			account.SetListAndIndex(account.GetShardId(), string(common.NewList), uint32(stakingData.RegisterNonce), isStakingV4Started)
			account.SetTempRating(stp.startRating)
			account.SetUnStakedEpoch(common.DefaultUnstakedEpoch)
		}

		if stakingData.UnStakedNonce == nonce && account.GetList() != string(common.InactiveList) {
			account.SetListAndIndex(account.GetShardId(), string(common.LeavingList), uint32(stakingData.UnStakedNonce), isStakingV4Started)
			account.SetUnStakedEpoch(stakingData.UnStakedEpoch)
		}
	}

	if stakingData.UnJailedNonce == nonce {
		if account.GetTempRating() < stp.unJailRating {
			account.SetTempRating(stp.unJailRating)
		}

		if !isValidator && account.GetUnStakedEpoch() == common.DefaultUnstakedEpoch {
			account.SetListAndIndex(account.GetShardId(), string(common.NewList), uint32(stakingData.UnJailedNonce), isStakingV4Started)
		}
	}

	err = stp.peerState.SaveAccount(account)
	if err != nil {
		return err
	}

	return nil
}

func (stp *stakingToPeer) updatePeerState(
	stakingData systemSmartContracts.StakedDataV2_0,
	blsPubKey []byte,
	nonce uint64,
) error {
	if !stp.enableEpochsHandler.IsFlagEnabled(common.StakeFlag) {
		return stp.updatePeerStateV1(stakingData, blsPubKey, nonce)
	}

	account, isNew, err := state.GetPeerAccountAndReturnIfNew(stp.peerState, blsPubKey)
	if err != nil {
		return err
	}

	isStakingV4Started := stp.enableEpochsHandler.IsFlagEnabled(common.StakingV4StartedFlag)

	isUnJailForInactive := !isNew && !stakingData.Staked &&
		stakingData.UnJailedNonce == nonce && account.GetList() == string(common.JailedList)
	if isUnJailForInactive {
		log.Debug("unJail for inactive node changed status to inactive list", "blsKey", account.AddressBytes(), "unStakedEpoch", stakingData.UnStakedEpoch)
		account.SetListAndIndex(account.GetShardId(), string(common.InactiveList), uint32(stakingData.UnJailedNonce), isStakingV4Started)
		if account.GetTempRating() < stp.unJailRating {
			account.SetTempRating(stp.unJailRating)
		}
		account.SetUnStakedEpoch(stakingData.UnStakedEpoch)

		if stp.enableEpochsHandler.IsFlagEnabled(common.ValidatorToDelegationFlag) && !bytes.Equal(account.GetRewardAddress(), stakingData.RewardAddress) {
			log.Debug("new reward address", "blsKey", blsPubKey, "rwdAddr", stakingData.RewardAddress)
			err = account.SetRewardAddress(stakingData.RewardAddress)
			if err != nil {
				return err
			}
		}

		return stp.peerState.SaveAccount(account)
	}

	if stakingData.StakedNonce == math.MaxUint64 {
		return nil
	}

	if !bytes.Equal(account.GetRewardAddress(), stakingData.RewardAddress) {
		log.Debug("new reward address", "blsKey", blsPubKey, "rwdAddr", stakingData.RewardAddress)
		err = account.SetRewardAddress(stakingData.RewardAddress)
		if err != nil {
			return err
		}
	}

	if isNew {
		log.Debug("new node", "blsKey", blsPubKey)
	}

	newNodesList := common.NewList
	if isStakingV4Started {
		newNodesList = common.AuctionList
	}

	isValidator := account.GetList() == string(common.EligibleList) || account.GetList() == string(common.WaitingList)
	if !stakingData.Jailed {
		if stakingData.StakedNonce == nonce && !isValidator {
			log.Debug("node is staked, changed status to", "list", newNodesList, "blsKey", blsPubKey)
			account.SetListAndIndex(account.GetShardId(), string(newNodesList), uint32(stakingData.StakedNonce), isStakingV4Started)
			account.SetTempRating(stp.startRating)
			account.SetUnStakedEpoch(common.DefaultUnstakedEpoch)
		}

		if stakingData.UnStakedNonce == nonce && account.GetList() != string(common.InactiveList) {
			log.Debug("node is unStaked, changed status to leaving list", "blsKey", blsPubKey)
			account.SetListAndIndex(account.GetShardId(), string(common.LeavingList), uint32(stakingData.UnStakedNonce), isStakingV4Started)
			account.SetUnStakedEpoch(stakingData.UnStakedEpoch)
		}
	}

	if stakingData.UnJailedNonce == nonce {
		if account.GetTempRating() < stp.unJailRating {
			log.Debug("node is unJailed, setting temp rating to start rating", "blsKey", blsPubKey)
			account.SetTempRating(stp.unJailRating)
			if stp.enableEpochsHandler.IsFlagEnabled(common.UnJailCleanupFlag) {
				account.SetConsecutiveProposerMisses(0)
			}
		}

		isNewValidator := !isValidator && stakingData.Staked
		if isNewValidator {
			log.Debug("node is unJailed and staked, changing status to", "list", newNodesList, "blsKey", blsPubKey)
			account.SetListAndIndex(account.GetShardId(), string(newNodesList), uint32(stakingData.UnJailedNonce), isStakingV4Started)
		}

		if account.GetList() == string(common.JailedList) {
			log.Debug("node is unJailed and not staked, changing status to inactive list", "blsKey", blsPubKey)
			account.SetListAndIndex(account.GetShardId(), string(common.InactiveList), uint32(stakingData.UnJailedNonce), isStakingV4Started)
			account.SetUnStakedEpoch(stakingData.UnStakedEpoch)
		}
	}

	if stakingData.JailedNonce == nonce && account.GetList() != string(common.InactiveList) {
		log.Debug("node is jailed, setting status to leaving", "blsKey", blsPubKey)
		account.SetListAndIndex(account.GetShardId(), string(common.LeavingList), uint32(stakingData.JailedNonce), isStakingV4Started)
		account.SetTempRating(stp.jailRating)
	}

	err = stp.peerState.SaveAccount(account)
	if err != nil {
		return err
	}

	return nil
}

func (stp *stakingToPeer) getAllModifiedStates(body *block.Body) ([]string, error) {
	affectedStates := make([]string, 0)

	for _, miniBlock := range body.MiniBlocks {
		if miniBlock.Type != block.SmartContractResultBlock {
			continue
		}
		if !stp.modifiedMBShardIDCheckerHandler.isModifiedStateMBValid(miniBlock) {
			continue
		}

		for _, txHash := range miniBlock.TxHashes {
			tx, err := stp.currTxs.GetTx(txHash)
			if err != nil {
				continue
			}

			if !bytes.Equal(tx.GetRcvAddr(), vm.StakingSCAddress) {
				continue
			}

			scr, ok := tx.(*smartContractResult.SmartContractResult)
			if !ok {
				return nil, process.ErrWrongTypeAssertion
			}

			storageUpdates, err := stp.argParser.GetStorageUpdates(string(scr.Data))
			if err != nil {
				continue
			}

			for _, storageUpdate := range storageUpdates {
				affectedStates = append(affectedStates, string(storageUpdate.Offset))
			}
		}
	}

	return affectedStates, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (stp *stakingToPeer) IsInterfaceNil() bool {
	return stp == nil
}
