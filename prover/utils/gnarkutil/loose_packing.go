package gnarkutil

import (
	"encoding/binary"
	"errors"
	hashinterface "hash"
	"math/big"

	"github.com/consensys/gnark-crypto/hash"
	"github.com/consensys/gnark/constraint/solver"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// SubStringsHashHint takes the following input:
//  - len(substrings)
//  - substring ends [0 .. maxNbSubStrings]
//  - inputBytes ...
// It returns H(inputBytes[: ends[0]]), H(inputBytes[ends[0], ends[1]]), ...
func SubStringsHashHint(maxNbSubStrings int, hsh hashinterface.Hash) solver.Hint {
	return func(_ *big.Int, ins []*big.Int, outs []*big.Int) error {
		if len(outs) != maxNbSubStrings {
			return errors.New("expected exactly maxNbBatches outputs")
		}

		nbBatches := ins[0].Int64()
		ends := utils.BigsToInts(ins[1 : 1+maxNbSubStrings])
		in := append(utils.BigsToBytes(ins[1+maxNbSubStrings:]), make([]byte, 31)...) // pad with 31 bytes to avoid out of range panic TODO try removing this

		subStart := 0
		zeros := make([]byte, hsh.BlockSize())

		for i := range outs[:nbBatches] {
			hsh.Reset()
			hsh.Write(in[subStart:ends[i]])

			subLen := ends[i] - subStart
			nbTrailingZeros := (subLen-1)/31*31 + 31 // ⌈subLen/31⌉ * 31
			hsh.Write(zeros[:nbTrailingZeros])

			outs[i].SetBytes(hsh.Sum(nil))
			subStart = ends[i]
		}

		return nil
	}
}

// ins: nbBatches, [end byte positions], payload...
// the result for each batch is <data (31 bytes)> ... <data (31 bytes)>
// for soundness some kind of length indicator must later be incorporated.
func PartialChecksumBatchesPackedHint(maxNbBatches int) solver.Hint {
	return func(_ *big.Int, ins []*big.Int, outs []*big.Int) error {
		if len(outs) != maxNbBatches {
			return errors.New("expected exactly maxNbBatches outputs")
		}

		nbBatches := ins[0].Int64()
		ends := utils.BigsToInts(ins[1 : 1+maxNbBatches])
		in := append(utils.BigsToBytes(ins[1+maxNbBatches:]), make([]byte, 31)...) // pad with 31 bytes to avoid out of range panic TODO try removing this

		hsh := hash.MIMC_BLS12_377.New()
		buf := make([]byte, 32)
		batchStart := 0

		for i := range outs[:nbBatches] {
			partialChecksumLooselyPackedBytes(in[batchStart:ends[i]], buf, hsh)
			outs[i].SetBytes(buf)
			batchStart = ends[i]
		}

		return nil
	}

}

func partialChecksumLooselyPackedBytes(b []byte, buf []byte, h hashinterface.Hash) {
	pack := func(b []byte, buffStartIndex int) {
		for i := range buf[:buffStartIndex] {
			buf[i] = 0
		}
		buf := buf[buffStartIndex:]
		for n := copy(buf, b); n < len(buf); n++ {
			buf[n] = 0
		}
	}

	pack(b, 1)
	for i := len(buf) - 1; i < len(b); i += len(buf) - 1 {
		h.Reset()
		h.Write(buf)
		pack(b[i:], 1)
		h.Write(buf)
		pack(h.Sum(nil), 0)
	}
}

// ChecksumLooselyPackedBytes produces the results expected by CheckBatchesSums, but more generalized
// b is partitioned into elements of length len(buf)-1 and hashed together, with zero padding on the right if necessary.
// the first bytes of the result are put in buf.
// if b consists of only one "element", the result is not hashed
func ChecksumLooselyPackedBytes(b []byte, buf []byte, h hashinterface.Hash) {
	partialChecksumLooselyPackedBytes(b, buf, h)

	// hash the length along with the partial sum
	var numBuf [8]byte
	binary.BigEndian.PutUint64(numBuf[:], uint64(len(b)))
	h.Reset()
	h.Write(numBuf[:])
	h.Write(buf)

	res := h.Sum(nil)

	for i := 0; i < len(buf)-len(res); i++ { // one final "packing"
		buf[i] = 0
	}

	copy(buf[len(buf)-len(res):], res)
}
