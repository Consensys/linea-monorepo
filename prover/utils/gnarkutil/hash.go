package gnarkutil

import (
	"fmt"
	hashinterface "hash"

	gcHash "github.com/consensys/gnark-crypto/hash"
)

type hashCompressorWrapper struct {
	hashinterface.Hash
}

// prependZeros so that len(res) = n.
func prependZeros(b []byte, n int) ([]byte, error) {
	if len(b) == n {
		return b, nil
	}
	if len(b) > n {
		return nil, fmt.Errorf("expected a maximum byte length of %d, got %d", n, len(b))
	}
	return append(make([]byte, n-len(b)), b...), nil
}

func (h hashCompressorWrapper) Compress(left []byte, right []byte) (compressed []byte, err error) {
	if left, err = prependZeros(left, h.BlockSize()); err != nil {
		return
	}
	if right, err = prependZeros(right, h.BlockSize()); err != nil {
		return
	}
	h.Reset()
	if _, err = h.Write(left); err != nil {
		return
	}
	if _, err = h.Write(right); err != nil {
		return
	}
	return h.Sum(nil), nil
}

// HashAsCompressor returns the hash compression function
// a, b ↦ H(a, b).
// NB! This is inefficient and is only intended for backwards compatibility.
func HashAsCompressor(hsh hashinterface.Hash) gcHash.Compressor {
	return hashCompressorWrapper{hsh}
}
