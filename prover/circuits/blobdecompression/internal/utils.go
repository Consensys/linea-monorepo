package internal

import (
	"slices"

	fr377 "github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/gnark-crypto/ecc/bls12-381/fr"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/lookup/logderivlookup"
)

// Truncate ensures that the slice is 0 starting from the n-th element
func Truncate(api frontend.API, slice []frontend.Variable, n frontend.Variable) []frontend.Variable {
	nYet := frontend.Variable(0)
	res := make([]frontend.Variable, len(slice))
	for i := range slice {
		nYet = api.Add(nYet, api.IsZero(api.Sub(i, n)))
		res[i] = api.MulAcc(api.Mul(1, slice[i]), slice[i], api.Neg(nYet))
	}
	return res
}

func SliceToTable(api frontend.API, slice []frontend.Variable) *logderivlookup.Table {
	table := logderivlookup.New(api)
	for i := range slice {
		table.Insert(slice[i])
	}
	return table
}

// RotateLeft rotates the slice v by n positions to the left, so that res[i] becomes v[(i+n)%len(v)]
func RotateLeft(api frontend.API, v []frontend.Variable, n frontend.Variable) (res []frontend.Variable) {
	res = make([]frontend.Variable, len(v))
	t := SliceToTable(api, v)
	for _, x := range v {
		t.Insert(x)
	}
	for i := range res {
		res[i] = t.Lookup(api.Add(i, n))[0]
	}
	return
}

// Bls12381ScalarToBls12377Scalars interprets its input as a BLS12-381 scalar, with a modular reduction if necessary, returning two BLS12-377 scalars
// r[1] is the lower 252 bits. r[0] is the higher 3 bits.
// useful in circuit "assign" functions
func Bls12381ScalarToBls12377Scalars(v interface{}) (r [2][]byte, err error) {
	var x fr.Element
	_, err = x.SetInterface(v)
	b := x.Bytes()

	r[0] = make([]byte, fr377.Bytes)
	r[0][fr.Bytes-1] = b[0] >> 4

	b[0] &= 0x0f
	r[1] = b[:]
	return
}

func PackedBytesToBits(api frontend.API, bytes []frontend.Variable, packingSize int) []frontend.Variable {
	bytesPerElem := (packingSize + 7) / 8
	firstByteNbBits := packingSize % 8
	if firstByteNbBits == 0 {
		firstByteNbBits = 8
	}
	nbElems := (len(bytes) + bytesPerElem - 1) / bytesPerElem

	if nbElems*bytesPerElem != len(bytes) {
		tmp := bytes
		bytes = make([]frontend.Variable, nbElems*bytesPerElem)
		copy(bytes, tmp)
		for i := len(tmp); i < len(bytes); i++ {
			bytes[i] = 0
		}
	}

	res := make([]frontend.Variable, 0, nbElems*packingSize)

	for i := 0; i < len(bytes); i += bytesPerElem {
		// first byte
		b := api.ToBinary(bytes[i], firstByteNbBits)
		slices.Reverse(b)
		res = append(res, b...)
		// remaining bytes
		for j := 1; j < bytesPerElem; j++ {
			b = api.ToBinary(bytes[i+j], 8)
			slices.Reverse(b)
			res = append(res, b...)
		}
	}

	return res
}

func PackNative(api frontend.API, words []frontend.Variable, bitsPerWord int) []frontend.Variable {
	wordsPerElem := (api.Compiler().FieldBitLen() - 1) / bitsPerWord
	res := make([]frontend.Variable, (len(words)+wordsPerElem-1)/wordsPerElem)
	if len(words) != len(res)*wordsPerElem {
		tmp := words
		words = make([]frontend.Variable, len(res)*wordsPerElem)
		copy(words, tmp)
		for i := len(tmp); i < len(words); i++ {
			words[i] = 0
		}
	}

	if bitsPerWord != 1 {
		panic("not implemented") // we're lazily using FromBinary here
	}
	for i := range res {
		currWords := words[i*wordsPerElem : (i+1)*wordsPerElem]
		slices.Reverse(currWords)
		res[i] = api.FromBinary(currWords...)
	}
	return res
}
