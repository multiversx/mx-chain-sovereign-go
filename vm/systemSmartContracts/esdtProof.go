package systemSmartContracts

import (
	"bytes"
	"math/big"
	"sort"
	"strconv"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/vm"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/pkg/errors"
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
	parentCheckCostPerParent = 10_000 // Cost for each parent existence check
	storageCostPerByte       = 1_000  // Placeholder cost for storing data
)

// registerProofTicker handles the registration of a new proof ticker and assigns the create role.
// Expected input: registerProofTicker@MYTICKER
func (e *esdt) registerProofTicker(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	returnCode := e.checkBasicCreateArguments(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	tokenType := []byte("proof")
	tokenIdentifier, token, err := e.createNewToken(
		args.CallerAddr,
		args.Arguments[0],
		args.Arguments[1],
		big.NewInt(0),
		0,
		[][]byte{},
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

// --- Create Proof (Dispatcher) ---

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
	dataGas := uint64(len(proofData) * storageCostPerByte)
	err = e.eei.UseGas(dataGas)
	if err != nil {
		return err
	}

	newNonce := e.incrementLatestNonce(ticker)
	pftKey := e.generatePFTStorageKey(ticker, newNonce)
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

	timestampBytes := big.NewInt(0).SetUint64(e.eei.BlockChainHook().LastTimeStamp()).Bytes()

	// Value = [32b_hash | 4b_algo_id | 8b_timestamp]
	pftValue := make([]byte, 0, hashSize+algoIDMicroPFTSize+timestampSize)
	pftValue = append(pftValue, hashBytes...)
	pftValue = append(pftValue, algoIDMicroPFT...)
	pftValue = append(pftValue, timestampBytes...)

	return pftValue
}

// createDPFT handles the creation of a DAGProof token, linking to parents.
// Expected input parts: [ "createProof", "MYTICKER", "dPFT", "proof_data", "<list_parent_keys>" ]
func (e *esdt) createDPFT(caller []byte, ticker []byte, parts [][]byte) error {
	if len(parts) < 2 {
		return errors.New("expected at least 2 arguments for createDPFT")
	}

	err := e.eei.UseGas(createDPFTBaseCost)
	if err != nil {
		return err
	}

	proofData := parts[0]
	dataGas := uint64(len(proofData) * storageCostPerByte)
	err = e.eei.UseGas(dataGas)
	if err != nil {
		return err
	}

	parents := parts[1:]
	lenParents := len(parents)
	err = e.eei.UseGas(uint64(lenParents * parentCheckCostPerParent))
	if err != nil {
		return err
	}

	newNonce := e.incrementLatestNonce(ticker)
	pftKey := e.generatePFTStorageKey(ticker, newNonce)

	for _, pkBytes := range parents {
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

	timestampBytes := big.NewInt(0).SetUint64(e.eei.BlockChainHook().LastTimeStamp()).Bytes()

	// Value = [32b_digest | 32b_parent_hash | 8b_flags | 8b_timestamp]
	pftValue := make([]byte, 0, hashSize+hashSize+flagsDPFTSize+timestampSize)
	pftValue = append(pftValue, proofHash...)
	pftValue = append(pftValue, parentHashBytes...)
	pftValue = append(pftValue, flagsDPFT...)
	pftValue = append(pftValue, timestampBytes...)

	return pftValue, nil
}

// generatePFTStorageKey creates the key used to store PFT data in the global state trie.
func (e *esdt) generatePFTStorageKey(ticker []byte, nonce uint64) []byte {
	key := pftPrefix + string(ticker) + "-" + strconv.FormatUint(nonce, 10)
	return []byte(key)
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

func (e *esdt) getESDTRolesForAcnt(address []byte, key []byte) (*ESDTRoles, error) {
	roles := &ESDTRoles{
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

func doesRoleExist(roles *ESDTRoles, role []byte) (int, bool) {
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

var noncePrefixProof = []byte(core.ProtectedKeyPrefix + core.ESDTNFTLatestNonceIdentifier)

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
