package common

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// LeftPadToFrBytes pads a []bytes element by adding zeroes to
// the left until the slice has fr.Bytes bytes.
func LeftPadToFrBytes(b []byte) []byte {
	if len(b) > fr.Bytes {
		utils.Panic("Passed a string of %v element but the max is {}", len(b), fr.Bytes)
	}
	c := append(make([]byte, fr.Bytes-len(b)), b...)
	return c
}
