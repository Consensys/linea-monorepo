package compress

import (
	"hash"
	"math/big"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/hash/mimc"
	"github.com/consensys/gnark/std/lookup/logderivlookup"
)

// Pack packs the words as tightly as possible, and works Big Endian: i.e. the first word is the most significant in the packed elem
// it is on the caller to make sure the words are within range
func Pack(api frontend.API, words []frontend.Variable, bitsPerWord int) []frontend.Variable {
	return PackN(api, words, bitsPerWord, (api.Compiler().FieldBitLen()-1)/bitsPerWord)
}

// PackN packs the words wordsPerElem at a time into field elements, and works Big Endian: i.e. the first word is the most significant in the packed elem
// it is on the caller to make sure the words are within range
func PackN(api frontend.API, words []frontend.Variable, bitsPerWord, wordsPerElem int) []frontend.Variable {
	res := make([]frontend.Variable, (len(words)+wordsPerElem-1)/wordsPerElem)

	r := make([]big.Int, wordsPerElem)
	r[wordsPerElem-1].SetInt64(1)
	for i := wordsPerElem - 2; i >= 0; i-- {
		r[i].Lsh(&r[i+1], uint(bitsPerWord))
	}

	for elemI := range res {
		res[elemI] = 0
		for wordI := 0; wordI < wordsPerElem; wordI++ {
			absWordI := elemI*wordsPerElem + wordI
			if absWordI >= len(words) {
				break
			}
			res[elemI] = api.Add(res[elemI], api.Mul(words[absWordI], r[wordI]))
		}
	}
	return res
}

// AssertChecksumEquals takes a MiMC hash of e and asserts it is equal to checksum
func AssertChecksumEquals(api frontend.API, e []frontend.Variable, checksum frontend.Variable) error {
	hsh, err := mimc.NewMiMC(api)
	if err != nil {
		return err
	}
	hsh.Write(e...)
	api.AssertIsEqual(hsh.Sum(), checksum)
	return nil
}

// ChecksumPaddedBytes packs b into field elements, then hashes the field elements along with validLength (encoded into a field element of its own)
func ChecksumPaddedBytes(b []byte, validLength int, hsh hash.Hash, fieldNbBits int) []byte {
	if validLength < 0 || validLength > len(b) {
		panic("invalid length")
	}
	usableBytesPerElem := (fieldNbBits+7)/8 - 1
	buf := make([]byte, usableBytesPerElem+1)
	for i := 0; i < len(b); i += usableBytesPerElem {
		copy(buf[1:], b[i:])
		for j := usableBytesPerElem; j+i > len(b) && j > 0; j-- {
			buf[j] = 0
		}
		hsh.Write(buf)
	}
	big.NewInt(int64(validLength)).FillBytes(buf)
	hsh.Write(buf)

	return hsh.Sum(nil)
}

// ShiftLeft erases shiftAmount many elements from the left of Slice and replaces them in the right with zeros
// it is the caller's responsibility to make sure that 0 \le shift < len(c)
func ShiftLeft(api frontend.API, slice []frontend.Variable, shiftAmount frontend.Variable) []frontend.Variable {
	res := make([]frontend.Variable, len(slice))
	l := logderivlookup.New(api)
	for i := range slice {
		l.Insert(slice[i])
	}
	for range slice {
		l.Insert(0)
	}
	for i := range slice {
		res[i] = l.Lookup(api.Add(i, shiftAmount))[0]
	}
	return res
}
