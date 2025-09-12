package runType

import (
	"bytes"
	"math/big"

	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

// UIntToBytes converts given int to big int bytes
func UIntToBytes(n uint32) []byte {
	if n == 0 {
		return []byte{0x0}
	}
	return big.NewInt(int64(n)).Bytes()
}

// FormatEGLDID will return the EGLD as ESDT
func FormatEGLDID(topic []byte) []byte {
	if bytes.Equal(topic, []byte("EGLD")) {
		return []byte(vmcommon.EGLDIdentifier)
	}
	return topic
}
