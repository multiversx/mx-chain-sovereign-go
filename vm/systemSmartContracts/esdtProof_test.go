package systemSmartContracts

import (
	"bytes"
	"encoding/binary"
	"math/big"
	"sort"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/vm"
	"github.com/multiversx/mx-chain-go/vm/mock"
)

func TestEsdtProof_registerProofTicker(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei := createDefaultEei()
	args.Eei = eei
	eei.gasRemaining = 9999
	e, _ := NewESDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc(funcRegisterProofTicker, nil)
	output := e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, output)
	require.Equal(t, eei.returnMessage, "not enough arguments")

	vmInput.CallValue = big.NewInt(0).Set(e.baseIssuingCost)
	vmInput.Arguments = [][]byte{[]byte("tokenName")}
	eei.returnMessage = ""
	output = e.Execute(vmInput)
	require.Equal(t, vmcommon.FunctionWrongSignature, output)
	require.Equal(t, eei.returnMessage, "arguments length mismatch")

	vmInput.Arguments = [][]byte{[]byte("tokenName"), []byte("PROOF")}
	eei.returnMessage = ""
	output = e.Execute(vmInput)
	require.Equal(t, vmcommon.Ok, output)
	require.Equal(t, len(eei.output), 1)
	token, _ := e.getExistingToken(eei.output[0])
	require.Equal(t, token.TokenType, []byte("proof"))
	require.Equal(t, token.TickerName, []byte("PROOF"))
}

func TestEsdtProof_createProof_ErrorCases(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei := createDefaultEei()
	args.Eei = eei
	eei.gasRemaining = 9999
	e, _ := NewESDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc(funcCreateProof, [][]byte{[]byte("PROOF")})
	output := e.Execute(vmInput)
	require.Equal(t, vmcommon.FunctionWrongSignature, output)
	require.Equal(t, eei.returnMessage, "not enough arguments")

	proofTicker := []byte("PROOF")

	vmInput.Arguments = [][]byte{proofTicker, []byte(proofTypeMicro), []byte("proofData")}
	vmInput.CallValue = big.NewInt(0)
	eei.returnMessage = ""
	output = e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, output)
	require.Equal(t, eei.returnMessage, "no ticker with given name")
}

func TestEsdtProof_generateMicroPFTValue(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	e, _ := NewESDTSmartContract(args)

	pftValue, _ := e.generateMicroPFTValue([]byte("proof"))
	require.Len(t, pftValue, 44)
}

func TestEsdtProof_createMicroPFT_Error(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	e, _ := NewESDTSmartContract(args)

	err := e.createMicroPFT([]byte("caller"), []byte("ticker"), [][]byte{[]byte("part1"), []byte("part2")})
	require.Error(t, err, "expected 1 part for createMicroPFT")
}

func TestEsdtProof_createMicroProofFT(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()

	proofToken := []byte("sov-PROOF-000000")
	proofData := []byte("proofData")
	timestamp := uint64(111)
	pftStorageValue := generatePFTValue(proofData, timestamp)
	tokenNonce := uint64(0)

	args.Eei = &mock.SystemEIStub{
		BlockChainHookCalled: func() vm.BlockchainHook {
			return &testscommon.BlockChainHookStub{
				LastTimeStampCalled: func() uint64 {
					return timestamp
				},
			}
		},
		GetStorageCalled: func(key []byte) []byte {
			token := &ESDTDataV2{
				OwnerAddress: []byte("owner"),
				SpecialRoles: []*ESDTRoles{
					{
						Address: []byte("owner"),
						Roles:   [][]byte{[]byte(core.ESDTRoleNFTCreate)},
					},
				},
				TokenType: []byte("proof"),
			}
			tokenBytes, _ := args.Marshalizer.Marshal(token)
			return tokenBytes
		},
		GetStorageFromAddressCalled: func(address []byte, key []byte) []byte {
			switch {
			case bytes.Contains(key, roleKeyPrefix):
				esdtRoles := &ESDTRoles{
					Roles: [][]byte{[]byte(core.ESDTRoleNFTCreate)},
				}
				esdtRolesBytes, _ := args.Marshalizer.Marshal(esdtRoles)
				return esdtRolesBytes
			case bytes.Equal(key, getNonceKey(proofToken)):
				require.Equal(t, core.SystemAccountAddress, address)

				return big.NewInt(0).SetUint64(tokenNonce).Bytes()
			default:
				require.Fail(t, "invalid key")
				return nil
			}
		},
		SetStorageForAddressCalled: func(address []byte, key []byte, value []byte) {
			require.Equal(t, core.SystemAccountAddress, address)

			switch {
			case bytes.Equal(key, getNonceKey(proofToken)):
				tokenNonce++
			case strings.Contains(string(key), pftPrefix):
				pftStorageKey := generatePFTStorageKey(proofToken, tokenNonce)
				require.Equal(t, pftStorageKey, key)
				require.Equal(t, pftStorageValue, value)
			default:
				require.Fail(t, "invalid key")
			}
		},
	}
	e, _ := NewESDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc(funcCreateProof, [][]byte{proofToken, []byte(proofTypeMicro), proofData})
	output := e.Execute(vmInput)
	require.Equal(t, vmcommon.Ok, output)
}

func generatePFTValue(proofData []byte, timestamp uint64) []byte {
	hasher := &hashingMocks.HasherMock{}

	hashBytes := hasher.Compute(string(proofData))

	timestampBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timestampBytes, timestamp)

	pftValue := make([]byte, 0, hashSize+timestampSize+algoIDMicroPFTSize)
	pftValue = append(pftValue, hashBytes...)
	pftValue = append(pftValue, timestampBytes...)
	pftValue = append(pftValue, algoIDMicroPFT...)

	return pftValue
}

func TestEsdtProof_generateDPFTValue(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	e, _ := NewESDTSmartContract(args)

	dpftValue, _ := e.generateDPFTValue([]byte("proof"), [][]byte{[]byte("parents")})
	require.Len(t, dpftValue, 80)
}

func TestEsdtProof_createDPFT_Error(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	e, _ := NewESDTSmartContract(args)

	err := e.createDPFT([]byte("caller"), []byte("ticker"), [][]byte{[]byte("part1")})
	require.Error(t, err, "expected at least 2 arguments for createDPFT")
}

func TestEsdtProof_createDAGProofFT(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()

	proofToken := []byte("sov-PROOF-1a2b3c")
	proofData := []byte("proofData")
	parentKeys := [][]byte{[]byte("sov-PROOF-000000-1"), []byte("sov-PROOF-000000-2")}
	timestamp := uint64(111)
	dpftStorageValue := generateDPFTValue(proofData, parentKeys, timestamp)
	tokenNonce := uint64(0)

	args.Eei = &mock.SystemEIStub{
		BlockChainHookCalled: func() vm.BlockchainHook {
			return &testscommon.BlockChainHookStub{
				LastTimeStampCalled: func() uint64 {
					return timestamp
				},
			}
		},
		GetStorageCalled: func(key []byte) []byte {
			token := &ESDTDataV2{
				OwnerAddress: []byte("owner"),
				SpecialRoles: []*ESDTRoles{
					{
						Address: []byte("owner"),
						Roles:   [][]byte{[]byte(core.ESDTRoleNFTCreate)},
					},
				},
				TokenType: []byte("proof"),
			}
			tokenBytes, _ := args.Marshalizer.Marshal(token)
			return tokenBytes
		},
		GetStorageFromAddressCalled: func(address []byte, key []byte) []byte {
			switch {
			case bytes.Contains(key, roleKeyPrefix):
				esdtRoles := &ESDTRoles{
					Roles: [][]byte{[]byte(core.ESDTRoleNFTCreate)},
				}
				esdtRolesBytes, _ := args.Marshalizer.Marshal(esdtRoles)
				return esdtRolesBytes
			case bytes.Equal(key, getNonceKey(proofToken)):
				require.Equal(t, core.SystemAccountAddress, address)

				return big.NewInt(0).SetUint64(tokenNonce).Bytes()
			case strings.Contains(string(key), "PROOF-000000"):
				return key
			default:
				require.Fail(t, "invalid key")
				return nil
			}
		},
		SetStorageForAddressCalled: func(address []byte, key []byte, value []byte) {
			require.Equal(t, core.SystemAccountAddress, address)

			switch {
			case bytes.Equal(key, getNonceKey(proofToken)):
				tokenNonce++
			case strings.Contains(string(key), pftPrefix):
				pftStorageKey := generatePFTStorageKey(proofToken, tokenNonce)
				require.Equal(t, pftStorageKey, key)
				require.Equal(t, dpftStorageValue, value)
			default:
				require.Fail(t, "invalid key")
			}
		},
	}
	e, _ := NewESDTSmartContract(args)

	arguments := [][]byte{proofToken, []byte(proofTypeDAG), proofData, big.NewInt(int64(len(parentKeys))).Bytes()}
	arguments = append(arguments, parentKeys...)
	vmInput := getDefaultVmInputForFunc(funcCreateProof, arguments)
	output := e.Execute(vmInput)
	require.Equal(t, vmcommon.Ok, output)
}

func generateDPFTValue(proofData []byte, parentKeys [][]byte, timestamp uint64) []byte {
	hasher := &hashingMocks.HasherMock{}

	hashBytes := hasher.Compute(string(proofData))
	parentHashBytes := computeParentListHash(hasher, parentKeys)

	timestampBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timestampBytes, timestamp)

	pftValue := make([]byte, 0, hashSize+timestampSize+algoIDMicroPFTSize)
	pftValue = append(pftValue, hashBytes...)
	pftValue = append(pftValue, timestampBytes...)
	pftValue = append(pftValue, parentHashBytes...)
	pftValue = append(pftValue, flagsDPFT...)

	return pftValue
}

func computeParentListHash(hasher *hashingMocks.HasherMock, parentKeys [][]byte) []byte {
	sort.Slice(parentKeys, func(i, j int) bool {
		return bytes.Compare(parentKeys[i], parentKeys[j]) < 0
	})
	var combined []byte
	for _, pk := range parentKeys {
		combined = append(combined, pk...)
	}
	return hasher.Compute(string(combined))
}

func TestEsdtProof_deleteProofsErrorCases(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei := createDefaultEei()
	args.Eei = eei
	e, _ := NewESDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("deleteProof", nil)
	output := e.Execute(vmInput)
	require.Equal(t, vmcommon.FunctionWrongSignature, output)
	require.Equal(t, eei.returnMessage, "not enough arguments")
}

func TestEsdtProof_deleteProofs(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()

	proofTokens := [][]byte{[]byte("sov-PROOF-000000-1"), []byte("sov-PROOF-000000-2"), []byte("INVALID")}
	proofData := []byte("proofData")
	proofTimestamps := []uint64{uint64(111), uint64(222)}
	currentTimestamp := deleteCutOff + proofTimestamps[0] + 1

	countGetProofWasCalled := 0
	countSetProofWasCalled := 0

	args.Eei = &mock.SystemEIStub{
		BlockChainHookCalled: func() vm.BlockchainHook {
			return &testscommon.BlockChainHookStub{
				LastTimeStampCalled: func() uint64 {
					return currentTimestamp
				},
			}
		},
		GetStorageFromAddressCalled: func(address []byte, key []byte) []byte {
			require.Equal(t, core.SystemAccountAddress, address)

			switch {
			case strings.Contains(string(key), pftPrefix):
				timestamp := proofTimestamps[countGetProofWasCalled]
				countGetProofWasCalled++
				return generatePFTValue(proofData, timestamp)
			default:
				require.Fail(t, "invalid key")
				return nil
			}
		},
		SetStorageForAddressCalled: func(address []byte, key []byte, value []byte) {
			require.Equal(t, core.SystemAccountAddress, address)

			switch {
			case strings.Contains(string(key), pftPrefix):
				countSetProofWasCalled++
				require.Nil(t, value)
			default:
				require.Fail(t, "invalid key")
			}
		},
	}
	e, _ := NewESDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("deleteProof", proofTokens)
	output := e.Execute(vmInput)
	require.Equal(t, vmcommon.Ok, output)
	require.Equal(t, countGetProofWasCalled, 2) // get 2 valid proofs
	require.Equal(t, countSetProofWasCalled, 1) // set (delete) only one because second one is not expired
}
