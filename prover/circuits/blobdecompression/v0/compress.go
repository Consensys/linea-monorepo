package v0

import (
	"fmt"
	"math/big"

	"github.com/consensys/linea-monorepo/prover/circuits/internal"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/gnark-crypto/hash"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/circuits/blobdecompression/v0/compress"
	"github.com/consensys/linea-monorepo/prover/circuits/blobdecompression/v0/compress/lzss"
	goLzss "github.com/consensys/linea-monorepo/prover/lib/compressor/blob/v0/compress/lzss"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// @alex: careful, the blobHeaderLen must be the unpacked one. This is a
// complexity leak.
func DecompressLZSS(
	api frontend.API,
	dictBytes []byte,
	blobBytesPacked []frontend.Variable,
	blobBytesLen frontend.Variable,
	payloadMaxPackedSize int,
	blobHeaderBytesLen frontend.Variable,
) (
	payloadPackedBytes []frontend.Variable,
	err error,
) {

	// The unpack align step to remove the zeroes contained in the blob byte.
	// @alex: this should be replaced by a more stable function. The function
	// below only works if the number of "useful" bytes in a chunk of 32 bytes
	// is 31. This will change as in theory, we can go as high as 252/256 useful
	// bytes. This also impacts the blobByteLen as we now have to account for
	// the unpacking in it. Wo do so, by computing

	var (
		blobUnpackedBytes = snarkUnpackAlign(blobBytesPacked)
	)

	// This extracts the checksum from the blob and checks that this is matching
	// the checksum of the provided dictionary. @alex: since the dict is
	// provided as a constant it is fine to provide it outside of GKR.

	parsedDictChecksum := compress.ReadNum(api, blobUnpackedBytes[:32], big.NewInt(256))
	dict, _ := assignVarByteSlice(dictBytes, len(dictBytes))

	// @alex: here we do not pad the dict given that it is a constant of the
	// circuit.
	dictChecksum := compress.ChecksumPaddedBytes(dictBytes, len(dictBytes), hash.MIMC_BLS12_377.New(), fr.Bits)
	api.AssertIsEqual(dictChecksum, parsedDictChecksum)

	// @alex: how can I trust that the headerLength is correct?
	// The ShiftLeft basically is there to skip the header of the compressed
	// data since it does not need to be decompressed.
	var (
		compressed             = internal.RotateLeft(api, blobUnpackedBytes, blobHeaderBytesLen) // TODO this should be preceded by at least partially parsing the header
		payloadMaxUnpackedSize = utils.DivCeil(payloadMaxPackedSize, 32) * 31
		payloadUnpacked        = make([]frontend.Variable, payloadMaxUnpackedSize)
	)

	cLen := api.Sub(blobBytesLen, blobHeaderBytesLen)

	_, err = lzss.Decompress(
		api,
		compressed,
		cLen,
		payloadUnpacked,
		dict,
		goLzss.BestCompression,
	)

	if err != nil {
		return nil, fmt.Errorf("could not instantiate the lzss solver: %w", err)
	}

	return snarkPackAlign(payloadUnpacked), nil
}

// Note, this function only works if the number of bits packed in a FE is 248.
// This is only a temporary decision.
func snarkUnpackAlign(bs []frontend.Variable) []frontend.Variable {
	res := make([]frontend.Variable, 0, len(bs))
	for i := range bs {
		if i%32 == 0 {
			continue
		}
		res = append(res, bs[i])
	}
	return res
}

// Note, this function only work if the number of bits packed in a FE is 248
func snarkPackAlign(bs []frontend.Variable) []frontend.Variable {
	var (
		numUnpacked    = len(bs)
		numZeroesToAdd = utils.DivCeil(numUnpacked, 31)
	)

	res := make([]frontend.Variable, 0, numUnpacked+numZeroesToAdd)
	for i := range bs {
		if i%31 == 0 {
			res = append(res, 0)
		}
		res = append(res, bs[i])
	}
	return res
}
