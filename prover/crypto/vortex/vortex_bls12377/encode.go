package vortex

import (
	"math/big"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/zk"
)

// Function to encode 8 31-bit zk.WrappedVariable into a single 256-bit frontend.Variable
func EncodeWVsToFV(api frontend.API, values [8]zk.WrappedVariable) frontend.Variable {
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
		bits[i] = frontend.Variable(0) // Explicitly set last 8 bits to zero (most significant bits)
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
		res = append(res, EncodeWVsToFV(api, buf))
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
