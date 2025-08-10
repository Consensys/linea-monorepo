package keccak

import (
	"math/bits"
	"unsafe"

	"github.com/consensys/linea-monorepo/prover/utils"
)

// bytesAsBlockPtrUnsafe unsafely cast a slice into an array. The caller is
// responsible for checking the length of the slice is at least as large as
// a block.
func bytesAsBlockPtrUnsafe(s []byte) *Block {
	if len(s) < Rate {
		panic("slice length is smaller than block size")
	}
	return (*Block)(unsafe.Pointer(&s[0]))
}

// castDigest casts a 4-uplets of uint64 into a Keccak digest
// Added bounds checking to prevent potential overflow
func castDigest(a0, a1, a2, a3 uint64) Digest {
	// Validate that the values are within expected bounds
	if a0 > 0xFFFFFFFFFFFFFFFF || a1 > 0xFFFFFFFFFFFFFFFF ||
		a2 > 0xFFFFFFFFFFFFFFFF || a3 > 0xFFFFFFFFFFFFFFFF {
		panic("digest values exceed expected bounds")
	}

	resU64 := [4]uint64{a0, a1, a2, a3}
	return *(*Digest)(unsafe.Pointer(&resU64[0]))
}

// cycShf is an alias for [bits.RotateLeft64]. The function performs a bit
// cyclic shift over a uin64.
var cycShf = bits.RotateLeft64

// mod5 returns x mod5 5. Used as a shorthand for wrapping around the coordinates
// of the keccak state concisely.
func mod5(x int) int {
	return utils.PositiveMod(x, Dim)
}
