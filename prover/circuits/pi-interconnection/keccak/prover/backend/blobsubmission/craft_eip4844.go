package blobsubmission

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/utils"
	"github.com/ethereum/go-ethereum/crypto/kzg4844"
)

// Prepare a response object by computing all the fields except for the proof.
func CraftResponse() (*Response, error) {
	panic("not implemented")
}

// Blob is populated with the compressedStream (with padding)
func compressedStreamToBlob(compressedStream []byte) (blob kzg4844.Blob, err error) {
	// Error is returned when len(compressedStream) is larger than the 4844 data blob [131072]byte
	if len(compressedStream) > len(blob) {
		return blob, fmt.Errorf("compressedStream length (%d) exceeds blob length (%d)", len(compressedStream), len(blob))
	}

	// Copy compressedStream to blob, padding with zeros
	copy(blob[:len(compressedStream)], compressedStream)

	// Sanity-check that the blob is right-padded with zeroes
	for i := len(compressedStream); i < len(blob); i++ {
		if blob[i] != 0 {
			utils.Panic("blob not padded correctly at index blob[%d]", i)
		}
	}
	return blob, nil

}
