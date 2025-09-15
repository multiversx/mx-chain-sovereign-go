package runType

import (
	"math/big"
)

// UIntToBytes converts given int to big int bytes
func UIntToBytes(n uint32) []byte {
	if n == 0 {
		return []byte{0x0}
	}
	return big.NewInt(int64(n)).Bytes()
}
