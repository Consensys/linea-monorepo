package internal

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/utils"
)

// CopyHexEncodedBytes panics if the string won't fit. If the string is too small, is will be 0-padded on the left.
func CopyHexEncodedBytes(dst []frontend.Variable, hex string) error {
	b, err := utils.HexDecodeString(hex)
	if err != nil {
		return err
	}

	slack := len(dst) - len(b)
	for i := 0; i < slack; i++ {
		dst[i] = 0
	}

	utils.Copy(dst[slack:], b) // This will panic if b is too long

	return nil
}
