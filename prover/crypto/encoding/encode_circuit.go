package encoding

import (
	"math/big"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/zk"
)

// NEW CODE FOR KOALABEAR ENCODING
// Function to encode 8 31-bit zk.WrappedVariable into a single 256-bit frontend.Variable
func Encode8WVsToFV(api frontend.API, values [8]zk.WrappedVariable) frontend.Variable {

	var result frontend.Variable = 0

	// Precompute all multipliers as constants
	multipliers := [8]*big.Int{
		big.NewInt(1),                        // 2^0
		big.NewInt(1 << 31),                  // 2^31
		new(big.Int).Lsh(big.NewInt(1), 62),  // 2^62
		new(big.Int).Lsh(big.NewInt(1), 93),  // 2^93
		new(big.Int).Lsh(big.NewInt(1), 124), // 2^124
		new(big.Int).Lsh(big.NewInt(1), 155), // 2^155
		new(big.Int).Lsh(big.NewInt(1), 186), // 2^186
		new(big.Int).Lsh(big.NewInt(1), 217), // 2^217
	}

	for i := 0; i < 8; i++ {
		value := values[7-i].AsNative()
		// Add the value to the result, scaled by the current multiplier
		result = api.Add(result, api.Mul(value, multipliers[i]))
	}

	return result
}

// Function to encode 31-bit zk.WrappedVariable into 256-bit frontend.Variable slices
func EncodeWVsToFVs(api frontend.API, values []zk.WrappedVariable) []frontend.Variable {
	var res []frontend.Variable
	for len(values) != 0 {
		var buf [8]zk.WrappedVariable
		for i := 0; i < 8; i++ {
			buf[i] = zk.ValueOf(0)
		}
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

// Function to encode a 256-bit frontend.Variable into 8 zk.WrappedVariable objects, each representing 30-bit limbs.
func EncodeFVTo8WVs(api frontend.API, value frontend.Variable) [8]zk.WrappedVariable {

	var res [8]zk.WrappedVariable
	bits := api.ToBinary(value, 256)

	apiGen, err := zk.NewGenericApi(api)
	if err != nil {
		panic(err)
	}

	for i := 0; i < 8; i++ {
		limbBits := append(bits[32*i : 32*i+30])
		res[i] = apiGen.FromBinary(limbBits...)
	}

	return res
}

// BLS to Koalabear encoding, 1 BLS -- > 9 Koalabear --> 1 BLS
// The Encode9WVsToFV function is used in the gnark verifier
// Function to encode 9 31-bit zk.WrappedVariable into a single 256-bit frontend.Variable
func Encode9WVsToFV(api frontend.API, values [KoalabearChunks]zk.WrappedVariable) frontend.Variable {
	apiGen, err := zk.NewGenericApi(api)
	if err != nil {
		panic(err)
	}

	bits := make([]frontend.Variable, 256)

	for i := 0; i < KoalabearChunks-1; i++ {
		// Convert the 31 bits of the current WrappedVariable to frontend variables
		limbBits := apiGen.ToBinary(values[KoalabearChunks-1-i], 30)
		copy(bits[30*i:], limbBits)
	}

	limbBits := apiGen.ToBinary(values[0], 16)
	copy(bits[240:], limbBits)

	return api.FromBinary(bits...)

}
