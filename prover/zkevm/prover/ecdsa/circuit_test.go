package ecdsa

import (
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/secp256k1/ecdsa"
	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Tests for 16x16 → 26x10 conversion (what gnark's emulated field expects)
// ============================================================================

// convertBE16x16ToLE26x10_Go is the pure Go version of the circuit conversion
// This lets us test the conversion logic without compiling a circuit
func convertBE16x16ToLE26x10_Go(hi, lo [common.NbLimbU128]uint64) [26]uint64 {
	// Reconstruct the full 256-bit value from Hi and Lo
	// Hi and Lo are in big-endian limb order: limb[0] = MSB
	fullValue := new(big.Int)
	for i := 0; i < 8; i++ {
		fullValue.Lsh(fullValue, 16)
		fullValue.Or(fullValue, big.NewInt(int64(hi[i])))
	}
	for i := 0; i < 8; i++ {
		fullValue.Lsh(fullValue, 16)
		fullValue.Or(fullValue, big.NewInt(int64(lo[i])))
	}

	// Decompose into 26 limbs of 10 bits (little-endian: limb[0] = LSB)
	var result [26]uint64
	mask := big.NewInt((1 << 10) - 1) // 0x3FF
	temp := new(big.Int).Set(fullValue)
	for i := 0; i < 26; i++ {
		limb := new(big.Int).And(temp, mask)
		result[i] = limb.Uint64()
		temp.Rsh(temp, 10)
	}
	return result
}

// conversionTestCircuit tests convertBE16x16ToLE26x10 in a gnark circuit
type conversionTestCircuit struct {
	Hi       [common.NbLimbU128]frontend.Variable `gnark:",public"`
	Lo       [common.NbLimbU128]frontend.Variable `gnark:",public"`
	Expected [26]frontend.Variable                `gnark:",public"`
}

func (c *conversionTestCircuit) Define(api frontend.API) error {
	result := convertBE16x16ToLE26x10(api, c.Hi, c.Lo)
	for i := 0; i < 26; i++ {
		api.AssertIsEqual(result[i], c.Expected[i])
	}
	return nil
}

// TestECRecoverWith26x10Limbs verifies that 26×10 limbs (what gnark's emulated field uses)
// produce correct ECRecover results. This proves the data flow:
// 16×16 data → 26×10 conversion → ECRecover → correct public key
// the public key, and signature are from InputFiller in circuit.go.
func TestECRecoverWith26x10Limbs(t *testing.T) {
	//  test data in 16×16 format
	msgHi := [common.NbLimbU128]uint64{0x0100, 0x0000, 0x0000, 0x0000, 0x0000, 0x0000, 0x0000, 0x0000}
	msgLo := [common.NbLimbU128]uint64{0x0000, 0x0000, 0x0000, 0x0000, 0x0000, 0x0000, 0x0000, 0x0000}

	rHi := [common.NbLimbU128]uint64{0xc604, 0x7f94, 0x41ed, 0x7d6d, 0x3045, 0x406e, 0x95c0, 0x7cd8}
	rLo := [common.NbLimbU128]uint64{0x5c77, 0x8e4b, 0x8cef, 0x3ca7, 0xabac, 0x09b9, 0x5c70, 0x9ee5}

	sHi := [common.NbLimbU128]uint64{0xe382, 0x3fca, 0x20f6, 0xbeb6, 0x9822, 0xa037, 0x4ae0, 0x3e6b}
	sLo := [common.NbLimbU128]uint64{0x8b93, 0x3599, 0x1e1b, 0xee71, 0xb5bf, 0x3423, 0x1653, 0x7013}

	v := uint64(27) // EVM format

	// Expected public key (generator point G for secret key = 1)
	expectedPKXHi := [common.NbLimbU128]uint64{0x79be, 0x667e, 0xf9dc, 0xbbac, 0x55a0, 0x6295, 0xce87, 0x0b07}
	expectedPKXLo := [common.NbLimbU128]uint64{0x029b, 0xfcdb, 0x2dce, 0x28d9, 0x59f2, 0x815b, 0x16f8, 0x1798}
	expectedPKYHi := [common.NbLimbU128]uint64{0x483a, 0xda77, 0x26a3, 0xc465, 0x5da4, 0xfbfc, 0x0e11, 0x08a8}
	expectedPKYLo := [common.NbLimbU128]uint64{0xfd17, 0xb448, 0xa685, 0x5419, 0x9c47, 0xd08f, 0xfb10, 0xd4b8}

	// Step 1: Convert 16×16 → 26×10 (what gnark's emulated field would receive)
	msg26x10 := convertBE16x16ToLE26x10_Go(msgHi, msgLo)
	r26x10 := convertBE16x16ToLE26x10_Go(rHi, rLo)
	s26x10 := convertBE16x16ToLE26x10_Go(sHi, sLo)
	expectedPKX26x10 := convertBE16x16ToLE26x10_Go(expectedPKXHi, expectedPKXLo)
	expectedPKY26x10 := convertBE16x16ToLE26x10_Go(expectedPKYHi, expectedPKYLo)

	// Step 2: Reconstruct big.Int values from 26×10 limbs (what gnark would do internally)
	reconstructFrom26x10 := func(limbs [26]uint64) *big.Int {
		result := new(big.Int)
		for i := 25; i >= 0; i-- {
			result.Lsh(result, 10)
			result.Or(result, big.NewInt(int64(limbs[i])))
		}
		return result
	}

	msgReconstructed := reconstructFrom26x10(msg26x10)
	rReconstructed := reconstructFrom26x10(r26x10)
	sReconstructed := reconstructFrom26x10(s26x10)
	expectedPKXReconstructed := reconstructFrom26x10(expectedPKX26x10)
	expectedPKYReconstructed := reconstructFrom26x10(expectedPKY26x10)

	// Step 3: Perform ECRecover with the reconstructed values
	msgHashBytes := msgReconstructed.FillBytes(make([]byte, 32))

	var pk ecdsa.PublicKey
	err := pk.RecoverFrom(msgHashBytes, uint(v-27), rReconstructed, sReconstructed)
	require.NoError(t, err, "ECRecover failed with 26×10 reconstructed data")

	recoveredX := pk.A.X.BigInt(new(big.Int))
	recoveredY := pk.A.Y.BigInt(new(big.Int))

	// Step 4: Verify the recovered public key matches expected
	require.Equal(t, expectedPKXReconstructed, recoveredX, "PKX mismatch after 26×10 conversion")
	require.Equal(t, expectedPKYReconstructed, recoveredY, "PKY mismatch after 26×10 conversion")

}

// TestConvert16x16To26x10_Circuit tests the conversion in an actual gnark circuit
func TestConvert16x16To26x10_Circuit(t *testing.T) {
	testCases := []struct {
		name string
		hi   [common.NbLimbU128]uint64
		lo   [common.NbLimbU128]uint64
	}{
		{
			name: "zero",
			hi:   [common.NbLimbU128]uint64{0, 0, 0, 0, 0, 0, 0, 0},
			lo:   [common.NbLimbU128]uint64{0, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			name: "public_key_x",
			hi:   [common.NbLimbU128]uint64{0x79be, 0x667e, 0xf9dc, 0xbbac, 0x55a0, 0x6295, 0xce87, 0x0b07},
			lo:   [common.NbLimbU128]uint64{0x029b, 0xfcdb, 0x2dce, 0x28d9, 0x59f2, 0x815b, 0x16f8, 0x1798},
		},
		{
			name: "signature_r",
			hi:   [common.NbLimbU128]uint64{0xc604, 0x7f94, 0x41ed, 0x7d6d, 0x3045, 0x406e, 0x95c0, 0x7cd8},
			lo:   [common.NbLimbU128]uint64{0x5c77, 0x8e4b, 0x8cef, 0x3ca7, 0xabac, 0x09b9, 0x5c70, 0x9ee5},
		},
		{
			name: "max_value",
			hi:   [common.NbLimbU128]uint64{0xFFFF, 0xFFFF, 0xFFFF, 0xFFFF, 0xFFFF, 0xFFFF, 0xFFFF, 0xFFFF},
			lo:   [common.NbLimbU128]uint64{0xFFFF, 0xFFFF, 0xFFFF, 0xFFFF, 0xFFFF, 0xFFFF, 0xFFFF, 0xFFFF},
		},
	}

	// Compile circuit once over KoalaBear
	circuit := &conversionTestCircuit{}
	ccs, err := frontend.CompileU32(koalabear.Modulus(), scs.NewBuilder, circuit)
	require.NoError(t, err, "circuit compilation failed")

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Compute expected 26x10 limbs using Go version
			expected := convertBE16x16ToLE26x10_Go(tc.hi, tc.lo)

			// Build witness
			var hi, lo [common.NbLimbU128]frontend.Variable
			var exp [26]frontend.Variable
			for i := 0; i < 8; i++ {
				hi[i] = tc.hi[i]
				lo[i] = tc.lo[i]
			}
			for i := 0; i < 26; i++ {
				exp[i] = expected[i]
			}

			witness := &conversionTestCircuit{Hi: hi, Lo: lo, Expected: exp}
			w, err := frontend.NewWitness(witness, koalabear.Modulus())
			require.NoError(t, err, "witness creation failed")

			err = ccs.IsSolved(w)
			require.NoError(t, err, "circuit not satisfied - conversion mismatch!")

		})
	}
}
