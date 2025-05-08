package systemSmartContracts

import (
	"bytes"
	"encoding/binary"
	"math/big"
	"sort"
	"strconv"
	"strings"

	"github.com/multiversx/mx-chain-core-go/core"
	esdtCore "github.com/multiversx/mx-chain-core-go/data/esdt"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/pkg/errors"

	"github.com/multiversx/mx-chain-go/vm"
)

// Constants related to Proof Tokens (PFTs) - Adapt prefixes as needed to avoid collision
const (
	// Function names
	funcRegisterProofTicker = "registerProofTicker"
	funcCreateProof         = "createProof"

	// Proof Types (as expected in input data)
	proofTypeMicro = "uPFT"
	proofTypeDAG   = "dPFT"

	// Storage prefixes and suffixes
	pftPrefix = "PFT:" // Prefix for storing PFT data in the global state trie

	// PFT Data constants
	algoIDMicroPFTSize = 4          // bytes
	algoIDMicroPFT     = "0001"     // Example Algo ID for µPFT
	flagsDPFTSize      = 8          // bytes
	flagsDPFT          = "00000001" // Example Flags for dPFT
	hashSize           = 32         // bytes (SHA256)
	timestampSize      = 8          // bytes (uint64)

	// Costs (These are placeholders - should be defined properly in gas schedule)
	createMicroPFTBaseCost   = 100_000
	createDPFTBaseCost       = 200_000
	deleteProofBaseCost      = 50_000
	deleteCostPerProof       = 1_000
	parentCheckCostPerParent = 10_000 // Cost for each parent existence check

	deleteCutOff = 259_200_000
)

// registerProofTicker handles the registration of a new proof ticker and assigns the create role.
// Expected input: registerProofTicker@MYTICKER
func (e *esdt) registerProofTicker(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	returnCode := e.checkBasicCreateArguments(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	if len(args.Arguments) != 2 {
		e.eei.AddReturnMessage("arguments length mismatch")
		return vmcommon.FunctionWrongSignature
	}

	tokenType := []byte("proof")
	tokenIdentifier, token, err := e.createNewToken(
		args.CallerAddr,
		args.Arguments[0],
		args.Arguments[1],
		big.NewInt(0),
		0,
		[][]byte{[]byte(canCreateMultiShard), []byte("true")},
		tokenType)
	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	allRoles, err := e.getAllRolesForTokenType(core.DynamicMetaESDT)
	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	properties, returnCode := e.setRolesForTokenAndAddress(token, args.CallerAddr, allRoles)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	returnCode = e.prepareAndSendRoleChangeData(tokenIdentifier, args.CallerAddr, allRoles, properties)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	err = e.saveToken(tokenIdentifier, token)
	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	logEntry := &vmcommon.LogEntry{
		Identifier: []byte(args.Function),
		Address:    args.CallerAddr,
		Topics:     [][]byte{tokenIdentifier, args.Arguments[0], args.Arguments[1], tokenType, big.NewInt(int64(0)).Bytes()},
	}
	e.eei.Finish(tokenIdentifier)
	e.eei.AddLogEntry(logEntry)

	return vmcommon.Ok
}

// createProof acts as a dispatcher based on the proof type provided in the input.
// Expected input: createProof@MYTICKER@PROOF_TYPE@... (specific format depends on PROOF_TYPE)
func (e *esdt) createProof(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if len(args.Arguments) < 3 {
		e.eei.AddReturnMessage("not enough arguments")
		return vmcommon.FunctionWrongSignature
	}
	_, returnCode := e.getTokenInfoAfterInputChecks(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	ticker := args.Arguments[0]
	proofType := string(args.Arguments[1])

	err := e.checkAllowedToExecute(args.CallerAddr, ticker, []byte(core.ESDTRoleNFTCreate))
	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	switch proofType {
	case proofTypeMicro:
		err = e.createMicroPFT(args.CallerAddr, ticker, args.Arguments[2:])
	case proofTypeDAG:
		err = e.createDPFT(args.CallerAddr, ticker, args.Arguments[2:])
	default:
		err = errors.New("unknown proof type")
	}

	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

// deleteProof@List<Proofs>
// it will delete the proofs given in the list which are older than 3 days
func (e *esdt) deleteProofs(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if len(args.Arguments) < 1 {
		e.eei.AddReturnMessage("not enough arguments")
		return vmcommon.FunctionWrongSignature
	}

	err := e.eei.UseGas(uint64(deleteProofBaseCost + len(args.Arguments)*deleteCostPerProof))
	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return vmcommon.OutOfGas
	}

	currentTimestamp := e.eei.BlockChainHook().LastTimeStamp()
	for _, pkBytes := range args.Arguments {
		if !validateProofTokenID(pkBytes) {
			continue
		}

		pkKey := pftPrefix + string(pkBytes)
		timeStamp := e.getProofTimestamp([]byte(pkKey))
		if currentTimestamp-timeStamp > deleteCutOff {
			e.eei.SetStorageForAddress(core.SystemAccountAddress, []byte(pkKey), nil)
		}
	}

	return vmcommon.Ok
}

func validateProofTokenID(token []byte) bool {
	splits := strings.Split(string(token), "-")
	if len(splits) != 4 {
		return false
	}

	if !esdtCore.IsValidTokenPrefix(splits[0]) {
		return false
	}

	if !esdtCore.IsTickerValid(splits[1]) {
		return false
	}

	if !esdtCore.IsRandomSeqValid(splits[2]) {
		return false
	}

	if len(splits[3]) > 20 {
		return false
	}

	return true
}

// createMicroPFT handles the creation of a simple MicroProof token.
// Expected input parts (from createProof): [ "createProof", "MYTICKER", "proof_data" ]
func (e *esdt) createMicroPFT(caller []byte, ticker []byte, parts [][]byte) error {
	if len(parts) != 1 {
		return errors.New("expected 1 part for createMicroPFT")
	}

	err := e.eei.UseGas(createMicroPFTBaseCost)
	if err != nil {
		return err
	}

	proofData := parts[0]
	newNonce := e.incrementLatestNonce(ticker)
	pftKey := generatePFTStorageKey(ticker, newNonce)
	pftValue := e.generateMicroPFTValue(proofData)

	e.eei.SetStorageForAddress(core.SystemAccountAddress, pftKey, pftValue)

	newEntry := vmcommon.LogEntry{
		Identifier: []byte("createMicroPFT"),
		Address:    caller,
		Topics:     [][]byte{ticker, big.NewInt(0).SetUint64(newNonce).Bytes(), pftKey},
		Data:       [][]byte{pftValue},
	}
	e.eei.AddLogEntry(&newEntry)

	return nil
}

func (e *esdt) generateMicroPFTValue(proofData []byte) []byte {
	hashBytes := e.hasher.Compute(string(proofData))

	timestampBytes := make([]byte, timestampSize)
	binary.BigEndian.PutUint64(timestampBytes, e.eei.BlockChainHook().LastTimeStamp())

	// Value = [32b_hash | 8b_timestamp | 4b_algo_id ]
	pftValue := make([]byte, 0, hashSize+timestampSize+algoIDMicroPFTSize)
	pftValue = append(pftValue, hashBytes...)
	pftValue = append(pftValue, timestampBytes...)
	pftValue = append(pftValue, algoIDMicroPFT...)

	return pftValue
}

// createDPFT handles the creation of a DAGProof token, linking to parents.
// Expected input parts: [ "createProof", "MYTICKER", "dPFT", "proof_data", "<list_parent_token_ids>" ]
func (e *esdt) createDPFT(caller []byte, ticker []byte, parts [][]byte) error {
	if len(parts) < 2 {
		return errors.New("expected at least 2 arguments for createDPFT")
	}

	err := e.eei.UseGas(createDPFTBaseCost)
	if err != nil {
		return err
	}

	proofData := parts[0]
	parents := parts[1:]
	err = e.eei.UseGas(uint64(len(parents) * parentCheckCostPerParent))
	if err != nil {
		return err
	}

	newNonce := e.incrementLatestNonce(ticker)
	pftKey := generatePFTStorageKey(ticker, newNonce)

	for _, pkBytes := range parents {
		if !validateProofTokenID(pkBytes) {
			return errors.New("invalid token id")
		}

		pkKey := append([]byte(pftPrefix), pkBytes...)
		parentValue := e.eei.GetStorageFromAddress(core.SystemAccountAddress, pkKey)
		if len(parentValue) == 0 {
			return errors.New("parent does not exist " + string(pkBytes))
		}
	}

	pftValue, err := e.generateDPFTValue(proofData, parents)
	if err != nil {
		return errors.Wrap(err, "failed to generate dPFT value")
	}

	e.eei.SetStorageForAddress(core.SystemAccountAddress, pftKey, pftValue)

	newEntry := vmcommon.LogEntry{
		Identifier: []byte("createDPFT"),
		Address:    caller,
		Topics:     [][]byte{ticker, big.NewInt(0).SetUint64(newNonce).Bytes(), pftKey},
		Data:       [][]byte{pftValue},
	}
	e.eei.AddLogEntry(&newEntry)

	return nil
}

func (e *esdt) generateDPFTValue(dPFTData []byte, sortedParentKeys [][]byte) ([]byte, error) {
	proofHash := e.hasher.Compute(string(dPFTData))
	parentHashBytes := e.computeParentListHash(sortedParentKeys)

	timestampBytes := make([]byte, timestampSize)
	binary.BigEndian.PutUint64(timestampBytes, e.eei.BlockChainHook().LastTimeStamp())

	// Value = [32b_digest | 8b_timestamp | 32b_parent_hash | 8b_flags]
	pftValue := make([]byte, 0, hashSize+timestampSize+hashSize+flagsDPFTSize)
	pftValue = append(pftValue, proofHash...)
	pftValue = append(pftValue, timestampBytes...)
	pftValue = append(pftValue, parentHashBytes...)
	pftValue = append(pftValue, flagsDPFT...)

	return pftValue, nil
}

// generatePFTStorageKey creates the key used to store PFT data in the global state trie.
func generatePFTStorageKey(ticker []byte, nonce uint64) []byte {
	key := pftPrefix + string(ticker) + "-" + strconv.FormatUint(nonce, 10)
	return []byte(key)
}

func (e *esdt) getProofTimestamp(pkKey []byte) uint64 {
	proofData := e.eei.GetStorageFromAddress(core.SystemAccountAddress, pkKey)
	if len(proofData) < 40 {
		return e.eei.BlockChainHook().LastTimeStamp()
	}

	timeStampData := proofData[32:40]
	return big.NewInt(0).SetBytes(timeStampData).Uint64()
}

// computeParentListHash calculates the SHA256 hash of a sorted list of parent keys.
func (e *esdt) computeParentListHash(parentKeys [][]byte) []byte {
	sort.Slice(parentKeys, func(i, j int) bool {
		return bytes.Compare(parentKeys[i], parentKeys[j]) < 0
	})

	var combined []byte
	for _, pk := range parentKeys {
		combined = append(combined, pk...)
	}

	return e.hasher.Compute(string(combined))
}

func (e *esdt) getESDTRolesForAcnt(address []byte, key []byte) (*esdtCore.ESDTRoles, error) {
	roles := &esdtCore.ESDTRoles{
		Roles: make([][]byte, 0),
	}

	marshaledData := e.eei.GetStorageFromAddress(address, key)
	if len(marshaledData) == 0 {
		return roles, nil
	}

	err := e.marshalizer.Unmarshal(roles, marshaledData)
	if err != nil {
		return nil, err
	}

	return roles, nil
}

func doesRoleExist(roles *esdtCore.ESDTRoles, role []byte) (int, bool) {
	for i, currentRole := range roles.Roles {
		if bytes.Equal(currentRole, role) {
			return i, true
		}
	}
	return -1, false
}

var roleKeyPrefix = []byte(core.ProtectedKeyPrefix + core.ESDTRoleIdentifier + core.ESDTKeyIdentifier)

func (e *esdt) checkAllowedToExecute(address []byte, tokenID []byte, action []byte) error {
	esdtTokenRoleKey := append(roleKeyPrefix, tokenID...)
	roles, err := e.getESDTRolesForAcnt(address, esdtTokenRoleKey)
	if err != nil {
		return err
	}
	_, exist := doesRoleExist(roles, action)
	if !exist {
		return vm.ErrInvalidArgument
	}

	return nil
}

var noncePrefixProof = []byte(core.ESDTNFTLatestNonceIdentifier)

func getNonceKey(tokenID []byte) []byte {
	return append(noncePrefixProof, tokenID...)
}

func (e *esdt) incrementLatestNonce(tokenID []byte) uint64 {
	nonceKey := getNonceKey(tokenID)
	nonceData := e.eei.GetStorageFromAddress(core.SystemAccountAddress, nonceKey)

	nonce := big.NewInt(0).SetBytes(nonceData).Uint64()
	nonce++

	e.eei.SetStorageForAddress(core.SystemAccountAddress, nonceKey, big.NewInt(0).SetUint64(nonce).Bytes())

	return nonce
}
