package internal

import (
	"fmt"
	hashinterface "hash"

	gcHash "github.com/consensys/gnark-crypto/hash"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/hash"
)

type snarkHashCompressorWrapper struct {
	hash.FieldHasher
}

func (h snarkHashCompressorWrapper) Compress(left frontend.Variable, right frontend.Variable) frontend.Variable {
	h.Reset()
	h.Write(left, right)
	return h.Sum()
}

// SnarkHashAsCompressor returns the hash compression function
// a, b ↦ H(a, b).
// NB! This is inefficient and is only intended for backwards compatibility.
func SnarkHashAsCompressor(hsh hash.FieldHasher) hash.Compressor {
	return snarkHashCompressorWrapper{hsh}
}
