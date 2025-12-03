package encoding

import (
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils/types"
)

func EncodeKoalabearOctupletToFrElement(elements [8]field.Element) fr.Element {
	var res fr.Element
	var bres, part big.Int
	for i := 0; i < 8; i++ {
		part.SetInt64(int64(elements[7-i].Bits()[0]))
		shift := uint(31 * i)  // Shift based on little-endian order
		part.Lsh(&part, shift) // Shift left by the appropriate position for little-endian
		bres.Add(&bres, &part) // Bitwise OR to combine
	}
	res.SetBigInt(&bres)
	return res
}

func EncodeKoalabearsToFrElement(elements []field.Element) []fr.Element {
	var res []fr.Element
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
		res = append(res, EncodeKoalabearOctupletToFrElement(buf))
	}
	return res
}

func EncodeFrElementToOctuplet(element fr.Element) field.Octuplet {
	mask := uint64(1073741823)
	bits := element.Bits()
	var res field.Octuplet
	for i := 0; i < 4; i++ {
		res[2*i].SetUint64(bits[i] & mask)
		res[2*i+1].SetUint64((bits[i] >> 32) & mask)
	}
	return res
}

// BLS to Koalabear encoding, 1 BLS -- > 9 Koalabear --> 1 BLS
// The following encoding and decoding are used in the compiler circuits
// Perform a lossless round-trip transformation between a Merkle Root (bls12.Element) and its decomposition into columns,
// ensuring the input Root matches the output Root.
const KoalabearChunks = 9

func EncodeBLS12RootToKoalabear(encoded fr.Element) [KoalabearChunks]field.Element {
	// Initialize an empty array to store the results
	var elements, res [KoalabearChunks]field.Element

	bytes := encoded.Bytes()
	// Convert the bytes32 to big.Int
	value := new(big.Int).SetBytes(bytes[:])

	// Loop to extract each 30-bit chunk
	for i := 0; i < KoalabearChunks; i++ {
		// Extract the corresponding 30-bit chunk by applying a mask
		chunk := new(big.Int).And(value, big.NewInt(0x3FFFFFFF)) // Mask for 30 bits (0x3FFFFFFF = 30 ones in binary)

		// Set the extracted chunk to the corresponding field.Element (element[i])
		elements[i].SetBigInt(chunk)

		// Right shift the `value` to move to the next chunk
		value.Rsh(value, 30) // Move to the next 30-bit chunk
	}

	// Since field.Elements are processed in little-endian order, reverse the array
	for i := 0; i < KoalabearChunks; i++ {
		res[i] = elements[KoalabearChunks-1-i]
	}
	return res
}

func DecodeKoalabearToBLS12Root(elements [KoalabearChunks]field.Element) fr.Element {
	expectedResult := big.NewInt(0)
	for i := 0; i < KoalabearChunks-1; i++ {
		part := big.NewInt(int64(elements[KoalabearChunks-1-i].Bits()[0]))

		shift := uint(30 * i)                   // Shift based on little-endian order
		part.Lsh(part, shift)                   // Shift left by the appropriate position for little-endian
		expectedResult.Or(expectedResult, part) // Bitwise OR to combine
	}
	part := big.NewInt(int64(elements[0].Bits()[0]))

	shift := uint(30 * (KoalabearChunks - 1)) // Shift based on little-endian order
	part.Lsh(part, shift)                     // Shift left by the appropriate position for little-endian
	expectedResult.Or(expectedResult, part)
	var res types.Bytes32
	expectedBytes := expectedResult.Bytes()
	copy(res[32-len(expectedBytes):], expectedBytes) // left pad with zeroes to 32 bytes

	var resElem fr.Element
	resElem.SetBytes(res[:])
	return resElem
}
