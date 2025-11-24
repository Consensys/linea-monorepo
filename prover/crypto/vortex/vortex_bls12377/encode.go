package vortex

import (
	"math/big"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/zk"
	"github.com/consensys/linea-monorepo/prover/utils/types"
)

// Function to encode a 256-bit frontend.Variable into 8 zk.WrappedVariable objects, each representing 30-bit limbs.
func EncodeFVTo8WVs(api frontend.API, value frontend.Variable) [8]zk.WrappedVariable {
	var res [8]zk.WrappedVariable
	bits := api.ToBinary(value, 256)

	for i := 0; i < 8; i++ {
		limbBits := append(bits[32*i:32*i+30], frontend.Variable(0), frontend.Variable(0))
		res[i] = zk.WrapFrontendVariable(api.FromBinary(limbBits...))
	}

	return res
}

// Function to encode 8 31-bit zk.WrappedVariable into a single 256-bit frontend.Variable
func Encode8WVsToFV(api frontend.API, values [8]zk.WrappedVariable) frontend.Variable {
	apiGen, err := zk.NewGenericApi(api)
	if err != nil {
		panic(err)
	}

	bits := make([]frontend.Variable, 256)

	for i := 0; i < 8; i++ {
		// Convert the 31 bits of the current WrappedVariable to frontend variables
		limbBits := apiGen.ToBinary(values[7-i], 31)
		copy(bits[31*i:], limbBits) // 8 leading padding bits come first
	}
	for i := 248; i < 256; i++ {
		bits[i] = 0 // Explicitly set last 8 bits to zero (most significant bits)
	}

	return api.FromBinary(bits...)
}

// Function to encode 31-bit zk.WrappedVariable into 256-bit frontend.Variable slices
func EncodeWVsToFVs(api frontend.API, values []zk.WrappedVariable) []frontend.Variable {
	var res []frontend.Variable
	for len(values) != 0 {
		var buf [8]zk.WrappedVariable
		// in this case we left pad by zeroes
		if len(values) < 8 {
			copy(buf[8-len(values):], values)
			values = values[:0]
		} else {
			copy(buf[:], values[:8])
			values = values[8:]
		}
		res = append(res, Encode8WVsToFV(api, buf))
	}

	return res
}

func Encode8KoalabearToBigInt(elements [8]field.Element) *big.Int {
	expectedResult := big.NewInt(0)
	for i := 0; i < 8; i++ {
		part := big.NewInt(int64(elements[7-i].Bits()[0]))

		shift := uint(31 * i)                   // Shift based on little-endian order
		part.Lsh(part, shift)                   // Shift left by the appropriate position for little-endian
		expectedResult.Or(expectedResult, part) // Bitwise OR to combine
	}
	return expectedResult
}

func EncodeKoalabearsToBytes(elements []field.Element) []byte {
	var res []byte
	for len(elements) != 0 {
		var buf [8]field.Element
		var bufBytes [32]byte
		// in this case we left pad by zeroes
		if len(elements) < 8 {
			copy(buf[8-len(elements):], elements[:])
			elements = elements[:0]
		} else {
			copy(buf[:], elements[:8])
			elements = elements[8:]
		}
		bytes := Encode8KoalabearToBigInt(buf).Bytes()
		copy(bufBytes[32-len(bytes):], bytes) // left pad with zeroes to 32 bytes
		res = append(res, bufBytes[:]...)
	}
	return res
}

// BLS to Koalabear encoding
const GnarkKoalabearNumElements = 11

func DecodeKoalabearToBLS12377(elements [11]field.Element) types.Bytes32 {
	expectedResult := big.NewInt(0)
	for i := 0; i < 10; i++ {
		part := big.NewInt(int64(elements[10-i].Bits()[0]))

		shift := uint(24 * i)                   // Shift based on little-endian order
		part.Lsh(part, shift)                   // Shift left by the appropriate position for little-endian
		expectedResult.Or(expectedResult, part) // Bitwise OR to combine
	}
	part := big.NewInt(int64(elements[0].Bits()[0]))

	shift := uint(24 * 10) // Shift based on little-endian order
	part.Lsh(part, shift)  // Shift left by the appropriate position for little-endian
	expectedResult.Or(expectedResult, part)
	var res types.Bytes32
	expectedBytes := expectedResult.Bytes()
	copy(res[32-len(expectedBytes):], expectedBytes) // left pad with zeroes to 32 bytes
	return res
}

func EncodeBLS12377ToKoalabear(encoded types.Bytes32) [GnarkKoalabearNumElements]field.Element {
	// Initialize an empty array to store the results
	var elements, res [11]field.Element

	// Convert the bytes32 to big.Int
	value := new(big.Int).SetBytes(encoded[:])

	// Loop to extract each 24-bit chunk
	for i := 0; i < 11; i++ {
		// Extract the corresponding 24-bit chunk by applying a mask
		chunk := new(big.Int).And(value, big.NewInt(0xFFFFFF)) // Mask for 24 bits (0xFFFFFF = 24 ones in binary)

		// Set the extracted chunk to the corresponding field.Element (element[i])
		elements[i].SetBigInt(chunk)

		// Right shift the `value` to move to the next chunk
		value.Rsh(value, 24) // Move to the next 24-bit chunk
	}

	// Since field.Elements are processed in little-endian order, reverse the array
	for i := 0; i < GnarkKoalabearNumElements; i++ {
		res[i] = elements[GnarkKoalabearNumElements-1-i]
	}
	return res
}

// func EncodeBLS12377ToKoalabear(elements types.Bytes32) [GnarkKoalabearNumElements]field.Element {

// 	var res [GnarkKoalabearNumElements]field.Element
// 	for i := 0; i < GnarkKoalabearNumElements; i++ {
// 		var bytes [4]byte
// 		if i != GnarkKoalabearNumElements-1 {
// 			copy(bytes[1:], elements[i*3:(i+1)*3])
// 		} else {
// 			copy(bytes[2:], elements[30:32])
// 		}
// 		res[i].SetBytes(bytes[:])
// 	}
// 	return res
// }

// func DecodeKoalabearToBLS12377(elements [GnarkKoalabearNumElements]field.Element) types.Bytes32 {
// 	var res types.Bytes32
// 	for i := 0; i < GnarkKoalabearNumElements; i++ {
// 		if i != GnarkKoalabearNumElements-1 {
// 			bytes := elements[i].Bytes()
// 			copy(res[i*3:(i+1)*3], bytes[1:])
// 		} else {
// 			bytes := elements[i].Bytes()
// 			copy(res[30:32], bytes[2:])
// 		}
// 	}
// 	return res
// }

// Function to encode 11 31-bit zk.WrappedVariable into a single 256-bit frontend.Variable
func Encode11WVsToFV(api frontend.API, values [GnarkKoalabearNumElements]zk.WrappedVariable) frontend.Variable {
	apiGen, err := zk.NewGenericApi(api)
	if err != nil {
		panic(err)
	}

	bits := make([]frontend.Variable, 256)

	for i := 0; i < GnarkKoalabearNumElements-1; i++ {
		// Convert the 31 bits of the current WrappedVariable to frontend variables
		limbBits := apiGen.ToBinary(values[GnarkKoalabearNumElements-1-i], 24)
		copy(bits[24*i:], limbBits) // 8 leading padding bits come first
	}

	limbBits := apiGen.ToBinary(values[0], 16)
	copy(bits[240:], limbBits) // 8 leading padding bits come first

	return api.FromBinary(bits...)

}
