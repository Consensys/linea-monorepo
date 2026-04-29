package plonk2

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRawLayoutForCurve_WordCounts(t *testing.T) {
	for _, tc := range []struct {
		curve      Curve
		scalar     int
		affine     int
		projective int
	}{
		{curve: CurveBN254, scalar: 4, affine: 8, projective: 12},
		{curve: CurveBLS12377, scalar: 4, affine: 12, projective: 18},
		{curve: CurveBW6761, scalar: 6, affine: 24, projective: 36},
	} {
		t.Run(tc.curve.String(), func(t *testing.T) {
			layout, err := RawLayoutForCurve(tc.curve)
			require.NoError(t, err)
			require.Equal(t, tc.scalar, layout.ScalarWords, "scalar word count should match Fr limbs")
			require.Equal(t, tc.affine, layout.AffinePointWords, "affine point should be X,Y base limbs")
			require.Equal(t, tc.projective, layout.ProjectivePointWords, "projective point should be X,Y,Z base limbs")
		})
	}
}

func TestRawLayoutForCurve_UnsupportedCurve(t *testing.T) {
	_, err := RawLayoutForCurve(Curve(99))
	require.Error(t, err, "unsupported curve should fail")
}

func TestRawLayoutValidation(t *testing.T) {
	layout, count, err := validateRawAffinePoints(CurveBN254, make([]uint64, 16))
	require.NoError(t, err)
	require.Equal(t, 8, layout.AffinePointWords, "BN254 affine width should be used")
	require.Equal(t, 2, count, "two affine points should be detected")

	_, _, err = validateRawAffinePoints(CurveBN254, make([]uint64, 15))
	require.Error(t, err, "truncated affine point buffer should fail")

	require.NoError(t, validateRawScalarsExact(CurveBN254, make([]uint64, 8), 2))
	require.Error(t, validateRawScalarsExact(CurveBN254, make([]uint64, 7), 2))

	count, err = validateRawScalarsAtMost(CurveBN254, make([]uint64, 4), 2)
	require.NoError(t, err)
	require.Equal(t, 1, count, "short scalar slices should be allowed for resident MSM")

	_, err = validateRawScalarsAtMost(CurveBN254, make([]uint64, 12), 2)
	require.Error(t, err, "scalar slices longer than resident SRS should fail")
}
