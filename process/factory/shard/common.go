package shard

import (
	"github.com/multiversx/mx-chain-core-go/core"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
)

// CreateMapOpCodeAddressIsAllowed create an opcode map for allowed addresses
func CreateMapOpCodeAddressIsAllowed(config config.VirtualMachineConfig, pubKeyConverter core.PubkeyConverter) (map[string]map[string]struct{}, error) {
	mapOpcodeAddressIsAllowed := make(map[string]map[string]struct{})

	transferAndExecuteByUserAddresses := config.TransferAndExecuteByUserAddresses
	if len(transferAndExecuteByUserAddresses) == 0 {
		return nil, process.ErrTransferAndExecuteByUserAddressesAreNil
	}

	mapOpcodeAddressIsAllowed[managedMultiTransferESDTNFTExecuteByUser] = make(map[string]struct{})
	for _, address := range transferAndExecuteByUserAddresses {
		decodedAddress, errDecode := pubKeyConverter.Decode(address)
		if errDecode != nil {
			return nil, errDecode
		}
		mapOpcodeAddressIsAllowed[managedMultiTransferESDTNFTExecuteByUser][string(decodedAddress)] = struct{}{}
	}

	return mapOpcodeAddressIsAllowed, nil
}
