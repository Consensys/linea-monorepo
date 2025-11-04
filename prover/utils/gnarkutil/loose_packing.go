package gnarkutil

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	hashinterface "hash"
	"io"
	"math/big"

	"github.com/consensys/gnark-crypto/hash"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// PartialChecksumBatchesLooselyPackedHint takes
// ins: nbBatches, [end byte positions], payload...
// The result for each batch is a hash of <data (31 bytes)> ... <data (31 bytes)>, padded with zeros at the end.
// For collision resistance each such result shall be hashed along with a length indicator.
func PartialChecksumBatchesLooselyPackedHint(maxNbBatches int, h hash.Compressor, ins, outs []*big.Int) error {
	if len(outs) != maxNbBatches {
		return errors.New("expected exactly maxNbBatches outputs")
	}

	nbBatches := ins[0].Int64()
	ends := utils.BigsToInts(ins[1 : 1+maxNbBatches])
	in := append(utils.BigsToBytes(ins[1+maxNbBatches:]), make([]byte, 31)...) // pad with 31 bytes to avoid out of range panic TODO try removing this

	batchStart := 0

	for i := range outs[:nbBatches] {
		outs[i].SetBytes(PartialChecksumLooselyPackedBytes(in[batchStart:ends[i]], h))
		batchStart = ends[i]
	}

	return nil
}

func PartialChecksumLooselyPackedBytes(b []byte, h hash.Compressor) []byte {
	buf := make([]byte, h.BlockSize())
	res := make([]byte, h.BlockSize())

	// pack with the requisite 0 prefix and suffix.
	pack := func(b []byte, buffStartIndex int) {
		for i := range buf[:buffStartIndex] {
			buf[i] = 0
		}
		buf := buf[buffStartIndex:]
		for n := copy(buf, b); n < len(buf); n++ {
			buf[n] = 0
		}
	}

	var err error

	pack(b, 1)
	copy(res, buf)

	for i := len(buf) - 1; i < len(b); i += len(buf) - 1 {
		pack(b[i:], 1)
		if res, err = h.Compress(res, buf); err != nil {
			panic(err)
		}
	}
	return res
}

func ChecksumLooselyPackedBytes(b []byte, h hash.Compressor) []byte {
	// hash the length along with the partial sum
	num := make([]byte, h.BlockSize())
	binary.BigEndian.PutUint64(num[len(num)-8:], uint64(len(b)))
	if res, err := h.Compress(num, PartialChecksumLooselyPackedBytes(b, h)); err != nil {
		panic(err)
	} else {
		return res
	}
}

func PartialMiMCChecksumLooselyPackedBytes(b []byte) []byte {
	return PartialChecksumLooselyPackedBytes(b, hashAsCompressor{hash.MIMC_BLS12_377.New()})
}

// ChecksumMiMCLooselyPackedBytes produces the results expected by CheckBatchesSums, but more generalized.
// b is partitioned into elements of length len(buf)-1 and hashed together, with zero padding on the right if necessary.
// The first bytes of the result are put in buf.
// If b consists of only one "element", the result is not hashed.
func ChecksumMiMCLooselyPackedBytes(b []byte) []byte {
	return ChecksumLooselyPackedBytes(b, hashAsCompressor{hash.MIMC_BLS12_377.New()})
}

type hashAsCompressor struct {
	hashinterface.Hash
}

// NewHashAsCompressor returns the given length-extended hash function as a collision-resistant compressor.
// NB! This is inefficient, using twice the number of required permutations.
// It should only be used for retro-compatibility.
func NewHashAsCompressor(hash hashinterface.Hash) hash.Compressor {
	return hashAsCompressor{hash}
}

func (h hashAsCompressor) Compress(left []byte, right []byte) (compressed []byte, err error) {
	if len(left) != h.BlockSize() {
		return nil, fmt.Errorf("expected a %d-byte hash block, got %d", h.BlockSize(), len(left))
	}
	if len(right) != h.BlockSize() {
		return nil, fmt.Errorf("expected a %d-byte hash block, got %d", h.BlockSize(), len(right))
	}
	h.Reset()
	if _, err = h.Write(left); err != nil {
		panic(err) // as per the hash interface contract, this should never happen.
	}
	if _, err = h.Write(right); err != nil {
		panic(err)
	}
	return h.Sum(nil), nil
}

// PackLoose packs the input bytes into blocks of elements,
// with each element's leftmost byte equal to 0.
func PackLoose(out io.Writer, input []byte, elemNbBytes, blockNbElems int) error {
	elemNbBytesUnpadded := elemNbBytes - 1
	nbElems := (len(input) + elemNbBytesUnpadded - 1) / elemNbBytesUnpadded
	nbBlocks := (nbElems + blockNbElems - 1) / blockNbElems
	for i := range nbElems {
		// left pad the element
		if _, err := out.Write([]byte{0}); err != nil {
			return err
		}
		if _, err := out.Write(input[i*elemNbBytesUnpadded : min(i*elemNbBytesUnpadded+elemNbBytesUnpadded, len(input))]); err != nil {
			return err
		}
	}
	// right pad
	if _, err := out.Write(make([]byte, nbBlocks*blockNbElems*elemNbBytes-len(input))); err != nil {
		return err
	}

}
