package encoding

import (
	"math/big"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2_koalabear"
)

// multipliers9 contains precomputed powers of 2 for 9-chunk (30-bit) encoding.
// Used for encoding 9 Koalabear elements into a single BLS12-377 element.
var multipliers9 = [KoalabearChunks]*big.Int{
	big.NewInt(1),                                 // 2^0
	big.NewInt(1 << ChunkBits9),                   // 2^30
	new(big.Int).Lsh(big.NewInt(1), ChunkBits9*2), // 2^60
	new(big.Int).Lsh(big.NewInt(1), ChunkBits9*3), // 2^90
	new(big.Int).Lsh(big.NewInt(1), ChunkBits9*4), // 2^120
	new(big.Int).Lsh(big.NewInt(1), ChunkBits9*5), // 2^150
	new(big.Int).Lsh(big.NewInt(1), ChunkBits9*6), // 2^180
	new(big.Int).Lsh(big.NewInt(1), ChunkBits9*7), // 2^210
	new(big.Int).Lsh(big.NewInt(1), ChunkBits9*8), // 2^240
}

// Encode8WVsToFV encodes 8 Koalabear zk.WrappedVariable into a single BLS12-377 frontend.Variable.
// Each input is treated as a 31-bit value. This is the circuit equivalent of EncodeKoalabearOctupletToFrElement.
func Encode8WVsToFV(api frontend.API, values [8]frontend.Variable) frontend.Variable {
	var result frontend.Variable = 0

	for i := 0; i < 8; i++ {
		value := values[7-i]
		result = api.Add(result, api.Mul(value, multipliers8[i]))
	}

	return result
}

// EncodeWVsToFVs encodes a slice of Koalabear zk.WrappedVariable into BLS12-377 frontend.Variable slices.
// Elements are packed 8 at a time, with left-padding of zeros if needed.
func EncodeKoalasToFVs(api frontend.API, values []frontend.Variable) []frontend.Variable {
	var res []frontend.Variable
	for len(values) != 0 {
		var buf [8]frontend.Variable
		for i := 0; i < 8; i++ {
			buf[i] = 0
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

// EncodeFVTo8WVs decodes a BLS12-377 frontend.Variable into 8 Koalabear zk.WrappedVariable.
// Each output represents a 30-bit limb extracted from the input.
// Note: This extracts 30-bit chunks, which differs from the 31-bit encoding in Encode8WVsToFV.
func EncodeFVTo8WVs(api frontend.API, value frontend.Variable) poseidon2_koalabear.Octuplet {
	var res [8]frontend.Variable
	bits := api.ToBinary(value, 256)

	for i := 0; i < 8; i++ {
		limbBits := bits[32*i : 32*i+30]
		res[i] = api.FromBinary(limbBits...)
	}

	return res
}

// Encode9KoalasToFV encodes 9 Koalabear zk.WrappedVariable into a single BLS12-377 frontend.Variable.
// Each input is treated as a 30-bit value. This is the circuit equivalent of DecodeKoalabearToBLS12Root.
// Used for Merkle root round-trip encoding in the gnark verifier.
func Encode9KoalasToFV(api frontend.API, values [KoalabearChunks]frontend.Variable) frontend.Variable {
	var result frontend.Variable = 0

	for i := 0; i < KoalabearChunks; i++ {
		value := values[KoalabearChunks-1-i]
		result = api.Add(result, api.Mul(value, multipliers9[i]))
	}

	return result
}
