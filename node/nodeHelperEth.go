package node

import (
	"bytes"
	"encoding/hex"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/node/external"
	"github.com/multiversx/mx-chain-go/process/factory"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/multiversx/mx-chain-vm-common-go/builtInFunctions/convertEncoding/common"
	"github.com/multiversx/mx-chain-vm-common-go/parsers"
	"strings"
)

// processEthereumOriginalData processes the original Ethereum data and formats it to the MultiversX standard
func processEthereumOriginalData(txArgs *external.ArgsCreateTransaction, _ []byte, receiverAddress []byte) ([]byte, error) {
	if core.IsSmartContractAddress(receiverAddress) {
		if core.IsEmptyAddress(receiverAddress) {
			return processDeployData(txArgs.OriginalDataField)
		}

		vmType, err := vmcommon.ParseVMTypeFromContractAddress(receiverAddress)
		if err != nil {
			return nil, err
		}

		if bytes.Equal(factory.EVMVirtualMachine, vmType) {
			return processCallData(txArgs.OriginalDataField)
		}
	}
	return txArgs.OriginalDataField, nil
}

func processDeployData(originalData []byte) ([]byte, error) {
	codeMetadata := vmcommon.GetEVMContractCodeMetadata()

	codeHex := hex.EncodeToString(originalData)
	vmTypeHex := hex.EncodeToString(factory.EVMVirtualMachine)
	codeMetadataHex := hex.EncodeToString(codeMetadata.ToBytes())

	return []byte(strings.Join([]string{codeHex, vmTypeHex, codeMetadataHex}, common.PartsSeparator)), nil
}

func processCallData(originalData []byte) ([]byte, error) {
	functionName, args, err := parsers.ParseEthereumCallInput(originalData)
	if err != nil {
		return nil, err
	}

	functionNameHex := hex.EncodeToString(functionName)
	if len(args) > 0 {
		argsHex := hex.EncodeToString(args)
		return []byte(strings.Join([]string{functionNameHex, argsHex}, common.PartsSeparator)), nil
	}
	return []byte(functionNameHex), nil
}
