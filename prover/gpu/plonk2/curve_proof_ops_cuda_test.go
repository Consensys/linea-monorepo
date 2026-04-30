//go:build cuda

package plonk2

import (
	"testing"
	"unsafe"

	bls12377 "github.com/consensys/gnark-crypto/ecc/bls12-377"
	blsfp "github.com/consensys/gnark-crypto/ecc/bls12-377/fp"
	blsfr "github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	blskzg "github.com/consensys/gnark-crypto/ecc/bls12-377/kzg"
	bn254 "github.com/consensys/gnark-crypto/ecc/bn254"
	bnfp "github.com/consensys/gnark-crypto/ecc/bn254/fp"
	bnfr "github.com/consensys/gnark-crypto/ecc/bn254/fr"
	bnkzg "github.com/consensys/gnark-crypto/ecc/bn254/kzg"
	bw6761 "github.com/consensys/gnark-crypto/ecc/bw6-761"
	bwfp "github.com/consensys/gnark-crypto/ecc/bw6-761/fp"
	bwfr "github.com/consensys/gnark-crypto/ecc/bw6-761/fr"
	bwkzg "github.com/consensys/gnark-crypto/ecc/bw6-761/kzg"
	blsplonk "github.com/consensys/gnark/backend/plonk/bls12-377"
	bnplonk "github.com/consensys/gnark/backend/plonk/bn254"
	bwplonk "github.com/consensys/gnark/backend/plonk/bw6-761"
	"github.com/stretchr/testify/require"
)

func TestCurveProofOpsRoundTripFieldSlices_CUDA(t *testing.T) {
	t.Run("bn254", func(t *testing.T) {
		ops, err := newCurveProofOps(CurveBN254)
		require.NoError(t, err, "creating BN254 proof ops should succeed")
		values := []bnfr.Element{bnfr.NewElement(1), bnfr.NewElement(2), bnfr.NewElement(3)}

		raw, err := ops.fieldSliceToRaw(values)
		require.NoError(t, err, "BN254 field-to-raw conversion should succeed")
		got, err := ops.rawToFieldSlice(raw)
		require.NoError(t, err, "BN254 raw-to-field conversion should succeed")
		require.Equal(t, values, got.([]bnfr.Element), "BN254 field roundtrip should preserve values")
	})

	t.Run("bls12-377", func(t *testing.T) {
		ops, err := newCurveProofOps(CurveBLS12377)
		require.NoError(t, err, "creating BLS12-377 proof ops should succeed")
		values := []blsfr.Element{blsfr.NewElement(4), blsfr.NewElement(5), blsfr.NewElement(6)}

		raw, err := ops.fieldSliceToRaw(values)
		require.NoError(t, err, "BLS12-377 field-to-raw conversion should succeed")
		got, err := ops.rawToFieldSlice(raw)
		require.NoError(t, err, "BLS12-377 raw-to-field conversion should succeed")
		require.Equal(t, values, got.([]blsfr.Element), "BLS12-377 field roundtrip should preserve values")
	})

	t.Run("bw6-761", func(t *testing.T) {
		ops, err := newCurveProofOps(CurveBW6761)
		require.NoError(t, err, "creating BW6-761 proof ops should succeed")
		values := []bwfr.Element{bwfr.NewElement(7), bwfr.NewElement(8), bwfr.NewElement(9)}

		raw, err := ops.fieldSliceToRaw(values)
		require.NoError(t, err, "BW6-761 field-to-raw conversion should succeed")
		got, err := ops.rawToFieldSlice(raw)
		require.NoError(t, err, "BW6-761 raw-to-field conversion should succeed")
		require.Equal(t, values, got.([]bwfr.Element), "BW6-761 field roundtrip should preserve values")
	})
}

func TestCurveProofOpsRawCommitmentToDigest_CUDA(t *testing.T) {
	t.Run("bn254", func(t *testing.T) {
		ops, err := newCurveProofOps(CurveBN254)
		require.NoError(t, err, "creating BN254 proof ops should succeed")
		_, _, generator, _ := bn254.Generators()
		var jac bn254.G1Jac
		jac.FromAffine(&generator)

		digest, err := ops.rawCommitmentToDigest(rawProjective(&jac, 3*bnfp.Limbs))
		require.NoError(t, err, "BN254 raw projective conversion should succeed")
		require.Equal(t, bnkzg.Digest(generator), digest.(bnkzg.Digest), "BN254 digest should match the generator")
	})

	t.Run("bls12-377", func(t *testing.T) {
		ops, err := newCurveProofOps(CurveBLS12377)
		require.NoError(t, err, "creating BLS12-377 proof ops should succeed")
		_, _, generator, _ := bls12377.Generators()
		var jac bls12377.G1Jac
		jac.FromAffine(&generator)

		digest, err := ops.rawCommitmentToDigest(rawProjective(&jac, 3*blsfp.Limbs))
		require.NoError(t, err, "BLS12-377 raw projective conversion should succeed")
		require.Equal(t, blskzg.Digest(generator), digest.(blskzg.Digest), "BLS12-377 digest should match the generator")
	})

	t.Run("bw6-761", func(t *testing.T) {
		ops, err := newCurveProofOps(CurveBW6761)
		require.NoError(t, err, "creating BW6-761 proof ops should succeed")
		_, _, generator, _ := bw6761.Generators()
		var jac bw6761.G1Jac
		jac.FromAffine(&generator)

		digest, err := ops.rawCommitmentToDigest(rawProjective(&jac, 3*bwfp.Limbs))
		require.NoError(t, err, "BW6-761 raw projective conversion should succeed")
		require.Equal(t, bwkzg.Digest(generator), digest.(bwkzg.Digest), "BW6-761 digest should match the generator")
	})
}

func TestCurveProofOpsProofSkeletons_CUDA(t *testing.T) {
	for _, tc := range []struct {
		name   string
		curve  Curve
		assert func(*testing.T, any)
	}{
		{
			name:  "bn254",
			curve: CurveBN254,
			assert: func(t *testing.T, proof any) {
				typed := proof.(*bnplonk.Proof)
				require.Len(t, typed.Bsb22Commitments, 2, "BN254 proof should allocate requested BSB22 slots")
			},
		},
		{
			name:  "bls12-377",
			curve: CurveBLS12377,
			assert: func(t *testing.T, proof any) {
				typed := proof.(*blsplonk.Proof)
				require.Len(t, typed.Bsb22Commitments, 2, "BLS12-377 proof should allocate requested BSB22 slots")
			},
		},
		{
			name:  "bw6-761",
			curve: CurveBW6761,
			assert: func(t *testing.T, proof any) {
				typed := proof.(*bwplonk.Proof)
				require.Len(t, typed.Bsb22Commitments, 2, "BW6-761 proof should allocate requested BSB22 slots")
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ops, err := newCurveProofOps(tc.curve)
			require.NoError(t, err, "creating proof ops should succeed")
			tc.assert(t, ops.newProof(2))
		})
	}
}

func rawProjective(point any, words int) []uint64 {
	raw := unsafe.Slice((*uint64)(unsafe.Pointer(reflectablePointer(point))), words)
	out := make([]uint64, len(raw))
	copy(out, raw)
	return out
}

func reflectablePointer(point any) unsafe.Pointer {
	switch p := point.(type) {
	case *bn254.G1Jac:
		return unsafe.Pointer(p)
	case *bls12377.G1Jac:
		return unsafe.Pointer(p)
	case *bw6761.G1Jac:
		return unsafe.Pointer(p)
	default:
		panic("unsupported projective point")
	}
}
