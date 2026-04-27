//go:build cuda

package plonk_test

// Validation tests for the new SW affine GPU primitives (foundation for
// batched-affine bucket accumulation). Compares GPU results to gnark-crypto's
// host reference exactly.

import (
	"crypto/rand"
	"math/big"
	"testing"

	bls12377 "github.com/consensys/gnark-crypto/ecc/bls12-377"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fp"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"

	"github.com/consensys/linea-monorepo/prover/gpu"
	"github.com/consensys/linea-monorepo/prover/gpu/plonk"
	"github.com/stretchr/testify/require"
)

// randomG1Affine returns a uniformly random G1 affine point on BLS12-377.
func randomG1Affine(t *testing.T) bls12377.G1Affine {
	t.Helper()
	var sBytes [32]byte
	_, err := rand.Read(sBytes[:])
	require.NoError(t, err)
	var s fr.Element
	s.SetBytes(sBytes[:])
	_, _, g1Aff, _ := bls12377.Generators()
	var p bls12377.G1Affine
	var sBig big.Int
	s.BigInt(&sBig)
	p.ScalarMultiplication(&g1Aff, &sBig)
	return p
}

// affineToLimbs returns the 12-uint64 limb layout matching the GPU G1AffineSW
// struct (x[0..6] then y[0..6], Montgomery form).
func affineToLimbs(p bls12377.G1Affine) [12]uint64 {
	var out [12]uint64
	for i := 0; i < 6; i++ {
		out[i] = p.X[i]
	}
	for i := 0; i < 6; i++ {
		out[6+i] = p.Y[i]
	}
	return out
}

// limbsToAffine inverts affineToLimbs.
func limbsToAffine(limbs [12]uint64) bls12377.G1Affine {
	var out bls12377.G1Affine
	for i := 0; i < 6; i++ {
		out.X[i] = limbs[i]
	}
	for i := 0; i < 6; i++ {
		out.Y[i] = limbs[6+i]
	}
	return out
}

// TestGPUSWPairAdd validates the GPU SW affine pair-add primitive against
// gnark-crypto's host reference.
func TestGPUSWPairAdd(t *testing.T) {
	dev, err := gpu.New()
	require.NoError(t, err)
	defer dev.Close()
	_ = dev

	for trial := 0; trial < 10; trial++ {
		p0 := randomG1Affine(t)
		p1 := randomG1Affine(t)

		// Expected: gnark-crypto host add (in Jacobian, then back to affine).
		var expectedJac bls12377.G1Jac
		expectedJac.FromAffine(&p0)
		var p1Jac bls12377.G1Jac
		p1Jac.FromAffine(&p1)
		expectedJac.AddAssign(&p1Jac)
		var expected bls12377.G1Affine
		expected.FromJacobian(&expectedJac)

		p0Limbs := affineToLimbs(p0)
		p1Limbs := affineToLimbs(p1)
		var outLimbs [12]uint64

		err := plonk.TestSWPairAddGPU(&p0Limbs, &p1Limbs, &outLimbs)
		require.NoError(t, err, "GPU pair-add error")

		got := limbsToAffine(outLimbs)
		require.Truef(t, got.Equal(&expected),
			"trial %d: GPU pair-add mismatch\n  p0 = %v\n  p1 = %v\n  got = %v\n  expected = %v",
			trial, p0, p1, got, expected)
	}
}

// TestGPUSWToTEExtended validates the GPU SW→TE extended conversion. Result on
// GPU is in extended coordinates (X, Y, T, Z); we convert to (x_te, y_te)
// affine and compare to the reference TE conversion derived from the same
// formulas as g1_te.go.
func TestGPUSWToTEExtended(t *testing.T) {
	dev, err := gpu.New()
	require.NoError(t, err)
	defer dev.Close()
	_ = dev

	for trial := 0; trial < 5; trial++ {
		p := randomG1Affine(t)

		pLimbs := affineToLimbs(p)
		var outLimbs [24]uint64

		err := plonk.TestSWToTEGPU(&pLimbs, &outLimbs)
		require.NoError(t, err, "GPU SW→TE error")

		// Recover (x_te, y_te) affine = (X/Z, Y/Z).
		var X, Y, Z fp.Element
		for i := 0; i < 6; i++ {
			X[i] = outLimbs[i]
			Y[i] = outLimbs[6+i]
			Z[i] = outLimbs[18+i]
		}
		var zInv, xTe, yTe fp.Element
		zInv.Inverse(&Z)
		xTe.Mul(&X, &zInv)
		yTe.Mul(&Y, &zInv)

		// Reference (mirrors the constants and math in g1_te.go).
		var teSqrtThree, teInvSqrtMinusA fp.Element
		teSqrtThree[0] = 0x3fabdfd08894e1e4
		teSqrtThree[1] = 0xcbf921ddcc1f55aa
		teSqrtThree[2] = 0xd17deff1460edc0c
		teSqrtThree[3] = 0xd394e81e7897028d
		teSqrtThree[4] = 0xc29c995d0912681a
		teSqrtThree[5] = 0x01515e6caff9d568
		teInvSqrtMinusA[0] = 0x3b092ce1fd76a6bd
		teInvSqrtMinusA[1] = 0x925230d9bba32683
		teInvSqrtMinusA[2] = 0x872d5d2fe991a197
		teInvSqrtMinusA[3] = 0x8367c527a82b2ab0
		teInvSqrtMinusA[4] = 0xe285bbb3ef662a15
		teInvSqrtMinusA[5] = 0x0160527a9283e729

		var one fp.Element
		one.SetOne()
		var xPlusOne, d1, d2 fp.Element
		xPlusOne.Add(&p.X, &one)
		d1.Mul(&p.Y, &teInvSqrtMinusA)
		d2.Add(&xPlusOne, &teSqrtThree)

		var d1Inv, d2Inv fp.Element
		d1Inv.Inverse(&d1)
		d2Inv.Inverse(&d2)

		var expXte, expYte, xMinusSqrt3 fp.Element
		expXte.Mul(&xPlusOne, &d1Inv)
		xMinusSqrt3.Sub(&xPlusOne, &teSqrtThree)
		expYte.Mul(&xMinusSqrt3, &d2Inv)

		require.Truef(t, xTe.Equal(&expXte),
			"trial %d: x_te mismatch\n  got = %v\n  expected = %v",
			trial, xTe, expXte)
		require.Truef(t, yTe.Equal(&expYte),
			"trial %d: y_te mismatch\n  got = %v\n  expected = %v",
			trial, yTe, expYte)
	}
}
