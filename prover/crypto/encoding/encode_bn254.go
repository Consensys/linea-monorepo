// Package encoding provides functions for encoding/decoding between
// Koalabear field elements (31-bit) and BN254 field elements (254-bit).
//
// Two encoding schemes are supported:
// - 8-chunk encoding: 8 x 31-bit = 248 bits (for general use)
// - 9-chunk encoding: 9 x 30-bit = 270 bits (for Merkle root round-trips)
//
// The BN254 scalar field modulus is ~2^254, which accommodates both schemes:
// - 8 x 31 = 248 < 254 bits (safe)
// - For root encoding, actual values are < p < 2^254, so ceil(254/30) = 9
//   chunks with the 9th chunk holding at most 14 bits (fits in KoalaBear).
package encoding

import (
	"math/big"

	bn254fr "github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/linea-monorepo/prover/maths/field"
)

// BN254RootChunks is the number of 30-bit chunks for BN254 root encoding.
// The BN254 scalar field is ~254 bits, so ceil(254/30) = 9 chunks.
const BN254RootChunks = 9

// EncodeKoalabearOctupletToBN254FrElement encodes 8 Koalabear field elements into a single BN254 field element.
// Each Koalabear element is treated as a 31-bit value, packed into the 254-bit BN254 field.
// Optimized to reduce big.Int allocations by reusing variables.
func EncodeKoalabearOctupletToBN254FrElement(elements [8]field.Element) bn254fr.Element {
	var res bn254fr.Element
	var bres, bElement, bScaled big.Int

	for i := 0; i < 8; i++ {
		elements[7-i].BigInt(&bElement)

		// Reuse bScaled instead of allocating new big.Int for multiplication result
		bScaled.Mul(&bElement, multipliers8[i])
		bres.Add(&bres, &bScaled)
	}

	res.SetBigInt(&bres)
	return res
}

// EncodeKoalabearsToBN254FrElement encodes a slice of Koalabear field elements into BN254 field elements.
// Elements are packed 8 at a time, with left-padding of zeros if the input length is not a multiple of 8.
// Optimized with pre-allocated result slice to avoid repeated slice growth.
func EncodeKoalabearsToBN254FrElement(elements []field.Element) []bn254fr.Element {
	if len(elements) == 0 {
		return nil
	}
	// Pre-allocate result slice with exact capacity needed
	numResults := (len(elements) + 7) / 8
	res := make([]bn254fr.Element, 0, numResults)

	for len(elements) != 0 {
		var buf [8]field.Element
		// in this case we left pad by zeroes
		if len(elements) < 8 {
			copy(buf[8-len(elements):], elements[:])
			elements = elements[:0]
		} else {
			copy(buf[:], elements[:8])
			elements = elements[8:]
		}
		res = append(res, EncodeKoalabearOctupletToBN254FrElement(buf))
	}
	return res
}

// EncodeBN254FrElementToOctuplet decodes a BN254 field element into 8 Koalabear field elements.
// It is used for randomness generation, not necessarily the inverse of EncodeKoalabearOctupletToBN254FrElement.
// Note: This extracts 30-bit chunks, which differs from the 31-bit encoding in EncodeKoalabearOctupletToBN254FrElement.
func EncodeBN254FrElementToOctuplet(element bn254fr.Element) field.Octuplet {
	bits := element.Bits()
	var res field.Octuplet
	for i := 0; i < 4; i++ {
		// Each 64-bit word in element gives two 30-bit chunks
		res[2*i].SetUint64(bits[i] & Mask30Bits)
		res[2*i+1].SetUint64((bits[i] >> 32) & Mask30Bits)
	}
	return res
}

// EncodeBN254RootToKoalabear decomposes a BN254 field element (Merkle root) into 9 Koalabear field elements.
// Each chunk is 30 bits, allowing for lossless round-trip encoding.
// The BN254 scalar field is ~254 bits, so the 9th chunk holds at most 14 bits.
func EncodeBN254RootToKoalabear(encoded bn254fr.Element) [BN254RootChunks]field.Element {
	var elements, res [BN254RootChunks]field.Element

	bytes := encoded.Bytes()
	var value, chunk big.Int
	value.SetBytes(bytes[:])

	// Extract each 30-bit chunk using reusable big.Int
	for i := 0; i < BN254RootChunks; i++ {
		chunk.And(&value, mask30BitsBigInt)
		elements[i].SetBigInt(&chunk)
		value.Rsh(&value, ChunkBits9)
	}

	// Reverse to match expected ordering (little-endian to big-endian)
	for i := 0; i < BN254RootChunks; i++ {
		res[i] = elements[BN254RootChunks-1-i]
	}
	return res
}

// DecodeBN254KoalabearToRoot reconstructs a BN254 field element from 9 Koalabear field elements.
// This is the inverse of EncodeBN254RootToKoalabear.
func DecodeBN254KoalabearToRoot(elements [BN254RootChunks]field.Element) bn254fr.Element {
	var expectedResult, bElement big.Int

	// Process all chunks in a single loop using reusable big.Int
	for i := 0; i < BN254RootChunks; i++ {
		elements[BN254RootChunks-1-i].BigInt(&bElement)
		bElement.Lsh(&bElement, uint(ChunkBits9*i))
		expectedResult.Or(&expectedResult, &bElement)
	}

	var res bn254fr.Element
	expectedBytes := expectedResult.Bytes()
	var buf [32]byte
	copy(buf[32-len(expectedBytes):], expectedBytes) // left pad with zeroes to 32 bytes
	res.SetBytes(buf[:])
	return res
}
