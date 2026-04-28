//go:build cuda

package plonk2

import (
	"math/big"
	"testing"

	blsfr "github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	blsfft "github.com/consensys/gnark-crypto/ecc/bls12-377/fr/fft"
	bnfr "github.com/consensys/gnark-crypto/ecc/bn254/fr"
	bnfft "github.com/consensys/gnark-crypto/ecc/bn254/fr/fft"
	bwfr "github.com/consensys/gnark-crypto/ecc/bw6-761/fr"
	bwfft "github.com/consensys/gnark-crypto/ecc/bw6-761/fr/fft"
	"github.com/stretchr/testify/require"

	"github.com/consensys/linea-monorepo/prover/gpu"
)

func TestButterfly4Inverse_CUDA(t *testing.T) {
	dev, err := gpu.New()
	require.NoError(t, err, "creating CUDA device should succeed")
	defer dev.Close()

	t.Run("bn254", func(t *testing.T) { testButterfly4InverseBN254(t, dev, 64) })
	t.Run("bls12-377", func(t *testing.T) { testButterfly4InverseBLS12377(t, dev, 64) })
	t.Run("bw6-761", func(t *testing.T) { testButterfly4InverseBW6761(t, dev, 64) })
}

func TestReduceBlindedCoset_CUDA(t *testing.T) {
	dev, err := gpu.New()
	require.NoError(t, err, "creating CUDA device should succeed")
	defer dev.Close()

	t.Run("bn254", func(t *testing.T) { testReduceBlindedCosetBN254(t, dev, 64) })
	t.Run("bls12-377", func(t *testing.T) { testReduceBlindedCosetBLS12377(t, dev, 64) })
	t.Run("bw6-761", func(t *testing.T) { testReduceBlindedCosetBW6761(t, dev, 64) })
}

func TestComputeL1Den_CUDA(t *testing.T) {
	dev, err := gpu.New()
	require.NoError(t, err, "creating CUDA device should succeed")
	defer dev.Close()

	t.Run("bn254", func(t *testing.T) { testComputeL1DenBN254(t, dev, 64) })
	t.Run("bls12-377", func(t *testing.T) { testComputeL1DenBLS12377(t, dev, 64) })
	t.Run("bw6-761", func(t *testing.T) { testComputeL1DenBW6761(t, dev, 64) })
}

func TestPlonkGateAccum_CUDA(t *testing.T) {
	dev, err := gpu.New()
	require.NoError(t, err, "creating CUDA device should succeed")
	defer dev.Close()

	t.Run("bn254", func(t *testing.T) { testPlonkGateAccumBN254(t, dev, 64) })
	t.Run("bls12-377", func(t *testing.T) { testPlonkGateAccumBLS12377(t, dev, 64) })
	t.Run("bw6-761", func(t *testing.T) { testPlonkGateAccumBW6761(t, dev, 64) })
}

func TestPlonkPermBoundary_CUDA(t *testing.T) {
	dev, err := gpu.New()
	require.NoError(t, err, "creating CUDA device should succeed")
	defer dev.Close()

	t.Run("bn254", func(t *testing.T) { testPlonkPermBoundaryBN254(t, dev, 64) })
	t.Run("bls12-377", func(t *testing.T) { testPlonkPermBoundaryBLS12377(t, dev, 64) })
	t.Run("bw6-761", func(t *testing.T) { testPlonkPermBoundaryBW6761(t, dev, 64) })
}

func TestPlonkZComputeFactors_CUDA(t *testing.T) {
	dev, err := gpu.New()
	require.NoError(t, err, "creating CUDA device should succeed")
	defer dev.Close()

	t.Run("bn254", func(t *testing.T) { testPlonkZComputeFactorsBN254(t, dev, 64) })
	t.Run("bls12-377", func(t *testing.T) { testPlonkZComputeFactorsBLS12377(t, dev, 64) })
	t.Run("bw6-761", func(t *testing.T) { testPlonkZComputeFactorsBW6761(t, dev, 64) })
}

func TestZPrefixProduct_CUDA(t *testing.T) {
	dev, err := gpu.New()
	require.NoError(t, err, "creating CUDA device should succeed")
	defer dev.Close()

	t.Run("bn254", func(t *testing.T) { testZPrefixProductBN254(t, dev, 2050) })
	t.Run("bls12-377", func(t *testing.T) { testZPrefixProductBLS12377(t, dev, 2050) })
	t.Run("bw6-761", func(t *testing.T) { testZPrefixProductBW6761(t, dev, 2050) })
}

func testButterfly4InverseBN254(t *testing.T, dev *gpu.Device, n int) {
	b0 := deterministicBN254(n, 101)
	b1 := deterministicBN254(n, 102)
	b2 := deterministicBN254(n, 103)
	b3 := deterministicBN254(n, 104)
	w0, w1, w2, w3 := butterfly4BN254Expected(n, b0, b1, b2, b3)
	runButterfly4BN254(t, dev, b0, b1, b2, b3)
	requireBN254Equal(t, w0, b0, "butterfly b0")
	requireBN254Equal(t, w1, b1, "butterfly b1")
	requireBN254Equal(t, w2, b2, "butterfly b2")
	requireBN254Equal(t, w3, b3, "butterfly b3")
}

func testButterfly4InverseBLS12377(t *testing.T, dev *gpu.Device, n int) {
	b0 := deterministicBLS12377(n, 105)
	b1 := deterministicBLS12377(n, 106)
	b2 := deterministicBLS12377(n, 107)
	b3 := deterministicBLS12377(n, 108)
	w0, w1, w2, w3 := butterfly4BLS12377Expected(n, b0, b1, b2, b3)
	runButterfly4BLS12377(t, dev, b0, b1, b2, b3)
	requireBLS12377Equal(t, w0, b0, "butterfly b0")
	requireBLS12377Equal(t, w1, b1, "butterfly b1")
	requireBLS12377Equal(t, w2, b2, "butterfly b2")
	requireBLS12377Equal(t, w3, b3, "butterfly b3")
}

func testButterfly4InverseBW6761(t *testing.T, dev *gpu.Device, n int) {
	b0 := deterministicBW6761(n, 109)
	b1 := deterministicBW6761(n, 110)
	b2 := deterministicBW6761(n, 111)
	b3 := deterministicBW6761(n, 112)
	w0, w1, w2, w3 := butterfly4BW6761Expected(n, b0, b1, b2, b3)
	runButterfly4BW6761(t, dev, b0, b1, b2, b3)
	requireBW6761Equal(t, w0, b0, "butterfly b0")
	requireBW6761Equal(t, w1, b1, "butterfly b1")
	requireBW6761Equal(t, w2, b2, "butterfly b2")
	requireBW6761Equal(t, w3, b3, "butterfly b3")
}

func testReduceBlindedCosetBN254(t *testing.T, dev *gpu.Device, n int) {
	src := deterministicBN254(n, 121)
	tail := deterministicBN254(3, 122)
	var cosetPowN bnfr.Element
	cosetPowN.SetUint64(17)
	expected := append([]bnfr.Element(nil), src...)
	for i := range tail {
		var scaled bnfr.Element
		scaled.Mul(&tail[i], &cosetPowN)
		expected[i].Add(&expected[i], &scaled)
	}
	out := reduceBlindedCosetBN254(t, dev, src, tail, cosetPowN)
	requireBN254Equal(t, expected, out, "reduce blinded coset")
}

func testReduceBlindedCosetBLS12377(t *testing.T, dev *gpu.Device, n int) {
	src := deterministicBLS12377(n, 123)
	tail := deterministicBLS12377(3, 124)
	var cosetPowN blsfr.Element
	cosetPowN.SetUint64(17)
	expected := append([]blsfr.Element(nil), src...)
	for i := range tail {
		var scaled blsfr.Element
		scaled.Mul(&tail[i], &cosetPowN)
		expected[i].Add(&expected[i], &scaled)
	}
	out := reduceBlindedCosetBLS12377(t, dev, src, tail, cosetPowN)
	requireBLS12377Equal(t, expected, out, "reduce blinded coset")
}

func testReduceBlindedCosetBW6761(t *testing.T, dev *gpu.Device, n int) {
	src := deterministicBW6761(n, 125)
	tail := deterministicBW6761(3, 126)
	var cosetPowN bwfr.Element
	cosetPowN.SetUint64(17)
	expected := append([]bwfr.Element(nil), src...)
	for i := range tail {
		var scaled bwfr.Element
		scaled.Mul(&tail[i], &cosetPowN)
		expected[i].Add(&expected[i], &scaled)
	}
	out := reduceBlindedCosetBW6761(t, dev, src, tail, cosetPowN)
	requireBW6761Equal(t, expected, out, "reduce blinded coset")
}

func testComputeL1DenBN254(t *testing.T, dev *gpu.Device, n int) {
	domain := bnfft.NewDomain(uint64(n))
	var cosetGen bnfr.Element
	cosetGen.SetUint64(17)
	expected := make([]bnfr.Element, n)
	var omega bnfr.Element
	omega.SetOne()
	one := new(bnfr.Element).SetOne()
	for i := range expected {
		expected[i].Mul(&cosetGen, &omega).Sub(&expected[i], one)
		omega.Mul(&omega, &domain.Generator)
	}
	out := computeL1DenBN254(t, dev, n, cosetGen)
	requireBN254Equal(t, expected, out, "compute L1 denominator")
}

func testComputeL1DenBLS12377(t *testing.T, dev *gpu.Device, n int) {
	domain := blsfft.NewDomain(uint64(n))
	var cosetGen blsfr.Element
	cosetGen.SetUint64(17)
	expected := make([]blsfr.Element, n)
	var omega blsfr.Element
	omega.SetOne()
	one := new(blsfr.Element).SetOne()
	for i := range expected {
		expected[i].Mul(&cosetGen, &omega).Sub(&expected[i], one)
		omega.Mul(&omega, &domain.Generator)
	}
	out := computeL1DenBLS12377(t, dev, n, cosetGen)
	requireBLS12377Equal(t, expected, out, "compute L1 denominator")
}

func testComputeL1DenBW6761(t *testing.T, dev *gpu.Device, n int) {
	domain := bwfft.NewDomain(uint64(n))
	var cosetGen bwfr.Element
	cosetGen.SetUint64(17)
	expected := make([]bwfr.Element, n)
	var omega bwfr.Element
	omega.SetOne()
	one := new(bwfr.Element).SetOne()
	for i := range expected {
		expected[i].Mul(&cosetGen, &omega).Sub(&expected[i], one)
		omega.Mul(&omega, &domain.Generator)
	}
	out := computeL1DenBW6761(t, dev, n, cosetGen)
	requireBW6761Equal(t, expected, out, "compute L1 denominator")
}

func testPlonkGateAccumBN254(t *testing.T, dev *gpu.Device, n int) {
	inputs := [9][]bnfr.Element{
		deterministicBN254(n, 131),
		deterministicBN254(n, 132),
		deterministicBN254(n, 133),
		deterministicBN254(n, 134),
		deterministicBN254(n, 135),
		deterministicBN254(n, 136),
		deterministicBN254(n, 137),
		deterministicBN254(n, 138),
		deterministicBN254(n, 139),
	}
	var zhKInv bnfr.Element
	zhKInv.SetUint64(19)
	expected := gateAccumExpectedBN254(inputs, zhKInv)
	out := gateAccumBN254(t, dev, inputs, zhKInv)
	requireBN254Equal(t, expected, out, "gate accumulation")
}

func testPlonkGateAccumBLS12377(t *testing.T, dev *gpu.Device, n int) {
	inputs := [9][]blsfr.Element{
		deterministicBLS12377(n, 141),
		deterministicBLS12377(n, 142),
		deterministicBLS12377(n, 143),
		deterministicBLS12377(n, 144),
		deterministicBLS12377(n, 145),
		deterministicBLS12377(n, 146),
		deterministicBLS12377(n, 147),
		deterministicBLS12377(n, 148),
		deterministicBLS12377(n, 149),
	}
	var zhKInv blsfr.Element
	zhKInv.SetUint64(19)
	expected := gateAccumExpectedBLS12377(inputs, zhKInv)
	out := gateAccumBLS12377(t, dev, inputs, zhKInv)
	requireBLS12377Equal(t, expected, out, "gate accumulation")
}

func testPlonkGateAccumBW6761(t *testing.T, dev *gpu.Device, n int) {
	inputs := [9][]bwfr.Element{
		deterministicBW6761(n, 151),
		deterministicBW6761(n, 152),
		deterministicBW6761(n, 153),
		deterministicBW6761(n, 154),
		deterministicBW6761(n, 155),
		deterministicBW6761(n, 156),
		deterministicBW6761(n, 157),
		deterministicBW6761(n, 158),
		deterministicBW6761(n, 159),
	}
	var zhKInv bwfr.Element
	zhKInv.SetUint64(19)
	expected := gateAccumExpectedBW6761(inputs, zhKInv)
	out := gateAccumBW6761(t, dev, inputs, zhKInv)
	requireBW6761Equal(t, expected, out, "gate accumulation")
}

func testPlonkPermBoundaryBN254(t *testing.T, dev *gpu.Device, n int) {
	inputs := [8][]bnfr.Element{
		deterministicBN254(n, 161),
		deterministicBN254(n, 162),
		deterministicBN254(n, 163),
		nonZeroBN254(deterministicBN254(n, 164)),
		deterministicBN254(n, 165),
		deterministicBN254(n, 166),
		deterministicBN254(n, 167),
		nonZeroBN254(deterministicBN254(n, 168)),
	}
	alpha, beta, gamma, l1Scalar, cosetShift, cosetShiftSq, cosetGen := permBoundaryScalarsBN254()
	domain := bnfft.NewDomain(uint64(n))
	expected := permBoundaryExpectedBN254(
		inputs, domain.Generator, alpha, beta, gamma,
		l1Scalar, cosetShift, cosetShiftSq, cosetGen,
	)
	out := permBoundaryBN254(
		t, dev, inputs, alpha, beta, gamma,
		l1Scalar, cosetShift, cosetShiftSq, cosetGen,
	)
	requireBN254Equal(t, expected, out, "permutation boundary")
}

func testPlonkPermBoundaryBLS12377(t *testing.T, dev *gpu.Device, n int) {
	inputs := [8][]blsfr.Element{
		deterministicBLS12377(n, 171),
		deterministicBLS12377(n, 172),
		deterministicBLS12377(n, 173),
		nonZeroBLS12377(deterministicBLS12377(n, 174)),
		deterministicBLS12377(n, 175),
		deterministicBLS12377(n, 176),
		deterministicBLS12377(n, 177),
		nonZeroBLS12377(deterministicBLS12377(n, 178)),
	}
	alpha, beta, gamma, l1Scalar, cosetShift, cosetShiftSq, cosetGen := permBoundaryScalarsBLS12377()
	domain := blsfft.NewDomain(uint64(n))
	expected := permBoundaryExpectedBLS12377(
		inputs, domain.Generator, alpha, beta, gamma,
		l1Scalar, cosetShift, cosetShiftSq, cosetGen,
	)
	out := permBoundaryBLS12377(
		t, dev, inputs, alpha, beta, gamma,
		l1Scalar, cosetShift, cosetShiftSq, cosetGen,
	)
	requireBLS12377Equal(t, expected, out, "permutation boundary")
}

func testPlonkPermBoundaryBW6761(t *testing.T, dev *gpu.Device, n int) {
	inputs := [8][]bwfr.Element{
		deterministicBW6761(n, 181),
		deterministicBW6761(n, 182),
		deterministicBW6761(n, 183),
		nonZeroBW6761(deterministicBW6761(n, 184)),
		deterministicBW6761(n, 185),
		deterministicBW6761(n, 186),
		deterministicBW6761(n, 187),
		nonZeroBW6761(deterministicBW6761(n, 188)),
	}
	alpha, beta, gamma, l1Scalar, cosetShift, cosetShiftSq, cosetGen := permBoundaryScalarsBW6761()
	domain := bwfft.NewDomain(uint64(n))
	expected := permBoundaryExpectedBW6761(
		inputs, domain.Generator, alpha, beta, gamma,
		l1Scalar, cosetShift, cosetShiftSq, cosetGen,
	)
	out := permBoundaryBW6761(
		t, dev, inputs, alpha, beta, gamma,
		l1Scalar, cosetShift, cosetShiftSq, cosetGen,
	)
	requireBW6761Equal(t, expected, out, "permutation boundary")
}

func testPlonkZComputeFactorsBN254(t *testing.T, dev *gpu.Device, n int) {
	l := deterministicBN254(n, 191)
	r := deterministicBN254(n, 192)
	o := deterministicBN254(n, 193)
	perm := deterministicPermutation(n)
	_, beta, gamma, _, cosetShift, cosetShiftSq, _ := permBoundaryScalarsBN254()
	domain := bnfft.NewDomain(uint64(n))
	expectedNum, expectedDen := zFactorsExpectedBN254(
		l, r, o, perm, domain.Generator, beta, gamma, cosetShift, cosetShiftSq,
	)
	num, den := zFactorsBN254(t, dev, l, r, o, perm, beta, gamma, cosetShift, cosetShiftSq)
	requireBN254Equal(t, expectedNum, num, "Z numerator factors")
	requireBN254Equal(t, expectedDen, den, "Z denominator factors")
}

func testPlonkZComputeFactorsBLS12377(t *testing.T, dev *gpu.Device, n int) {
	l := deterministicBLS12377(n, 194)
	r := deterministicBLS12377(n, 195)
	o := deterministicBLS12377(n, 196)
	perm := deterministicPermutation(n)
	_, beta, gamma, _, cosetShift, cosetShiftSq, _ := permBoundaryScalarsBLS12377()
	domain := blsfft.NewDomain(uint64(n))
	expectedNum, expectedDen := zFactorsExpectedBLS12377(
		l, r, o, perm, domain.Generator, beta, gamma, cosetShift, cosetShiftSq,
	)
	num, den := zFactorsBLS12377(t, dev, l, r, o, perm, beta, gamma, cosetShift, cosetShiftSq)
	requireBLS12377Equal(t, expectedNum, num, "Z numerator factors")
	requireBLS12377Equal(t, expectedDen, den, "Z denominator factors")
}

func testPlonkZComputeFactorsBW6761(t *testing.T, dev *gpu.Device, n int) {
	l := deterministicBW6761(n, 197)
	r := deterministicBW6761(n, 198)
	o := deterministicBW6761(n, 199)
	perm := deterministicPermutation(n)
	_, beta, gamma, _, cosetShift, cosetShiftSq, _ := permBoundaryScalarsBW6761()
	domain := bwfft.NewDomain(uint64(n))
	expectedNum, expectedDen := zFactorsExpectedBW6761(
		l, r, o, perm, domain.Generator, beta, gamma, cosetShift, cosetShiftSq,
	)
	num, den := zFactorsBW6761(t, dev, l, r, o, perm, beta, gamma, cosetShift, cosetShiftSq)
	requireBW6761Equal(t, expectedNum, num, "Z numerator factors")
	requireBW6761Equal(t, expectedDen, den, "Z denominator factors")
}

func testZPrefixProductBN254(t *testing.T, dev *gpu.Device, n int) {
	ratio := nonZeroBN254(deterministicBN254(n, 201))
	expected := zPrefixExpectedBN254(ratio)
	out := zPrefixBN254(t, dev, ratio)
	requireBN254Equal(t, expected, out, "Z prefix product")
}

func testZPrefixProductBLS12377(t *testing.T, dev *gpu.Device, n int) {
	ratio := nonZeroBLS12377(deterministicBLS12377(n, 202))
	expected := zPrefixExpectedBLS12377(ratio)
	out := zPrefixBLS12377(t, dev, ratio)
	requireBLS12377Equal(t, expected, out, "Z prefix product")
}

func testZPrefixProductBW6761(t *testing.T, dev *gpu.Device, n int) {
	ratio := nonZeroBW6761(deterministicBW6761(n, 203))
	expected := zPrefixExpectedBW6761(ratio)
	out := zPrefixBW6761(t, dev, ratio)
	requireBW6761Equal(t, expected, out, "Z prefix product")
}

func runButterfly4BN254(
	t *testing.T,
	dev *gpu.Device,
	b0, b1, b2, b3 []bnfr.Element,
) {
	t.Helper()
	n := len(b0)
	vecs := mustButterflyVectors(t, dev, CurveBN254, rawBN254(b0), rawBN254(b1), rawBN254(b2), rawBN254(b3), n)
	defer freeFrVectors(vecs[:]...)
	omega4Inv, quarter := butterfly4ScalarsBN254(n)
	require.NoError(t, Butterfly4Inverse(
		vecs[0], vecs[1], vecs[2], vecs[3],
		cloneRaw(rawBN254([]bnfr.Element{omega4Inv})),
		cloneRaw(rawBN254([]bnfr.Element{quarter})),
	))
	require.NoError(t, vecs[0].CopyToHostRaw(rawBN254(b0)), "copying b0 should succeed")
	require.NoError(t, vecs[1].CopyToHostRaw(rawBN254(b1)), "copying b1 should succeed")
	require.NoError(t, vecs[2].CopyToHostRaw(rawBN254(b2)), "copying b2 should succeed")
	require.NoError(t, vecs[3].CopyToHostRaw(rawBN254(b3)), "copying b3 should succeed")
}

func runButterfly4BLS12377(
	t *testing.T,
	dev *gpu.Device,
	b0, b1, b2, b3 []blsfr.Element,
) {
	t.Helper()
	n := len(b0)
	vecs := mustButterflyVectors(t, dev, CurveBLS12377, rawBLS12377(b0), rawBLS12377(b1), rawBLS12377(b2), rawBLS12377(b3), n)
	defer freeFrVectors(vecs[:]...)
	omega4Inv, quarter := butterfly4ScalarsBLS12377(n)
	require.NoError(t, Butterfly4Inverse(
		vecs[0], vecs[1], vecs[2], vecs[3],
		cloneRaw(rawBLS12377([]blsfr.Element{omega4Inv})),
		cloneRaw(rawBLS12377([]blsfr.Element{quarter})),
	))
	require.NoError(t, vecs[0].CopyToHostRaw(rawBLS12377(b0)), "copying b0 should succeed")
	require.NoError(t, vecs[1].CopyToHostRaw(rawBLS12377(b1)), "copying b1 should succeed")
	require.NoError(t, vecs[2].CopyToHostRaw(rawBLS12377(b2)), "copying b2 should succeed")
	require.NoError(t, vecs[3].CopyToHostRaw(rawBLS12377(b3)), "copying b3 should succeed")
}

func runButterfly4BW6761(
	t *testing.T,
	dev *gpu.Device,
	b0, b1, b2, b3 []bwfr.Element,
) {
	t.Helper()
	n := len(b0)
	vecs := mustButterflyVectors(t, dev, CurveBW6761, rawBW6761(b0), rawBW6761(b1), rawBW6761(b2), rawBW6761(b3), n)
	defer freeFrVectors(vecs[:]...)
	omega4Inv, quarter := butterfly4ScalarsBW6761(n)
	require.NoError(t, Butterfly4Inverse(
		vecs[0], vecs[1], vecs[2], vecs[3],
		cloneRaw(rawBW6761([]bwfr.Element{omega4Inv})),
		cloneRaw(rawBW6761([]bwfr.Element{quarter})),
	))
	require.NoError(t, vecs[0].CopyToHostRaw(rawBW6761(b0)), "copying b0 should succeed")
	require.NoError(t, vecs[1].CopyToHostRaw(rawBW6761(b1)), "copying b1 should succeed")
	require.NoError(t, vecs[2].CopyToHostRaw(rawBW6761(b2)), "copying b2 should succeed")
	require.NoError(t, vecs[3].CopyToHostRaw(rawBW6761(b3)), "copying b3 should succeed")
}

func reduceBlindedCosetBN254(
	t *testing.T,
	dev *gpu.Device,
	src, tail []bnfr.Element,
	cosetPowN bnfr.Element,
) []bnfr.Element {
	t.Helper()
	srcVec, err := NewFrVector(dev, CurveBN254, len(src))
	require.NoError(t, err, "allocating source vector should succeed")
	defer srcVec.Free()
	dstVec, err := NewFrVector(dev, CurveBN254, len(src))
	require.NoError(t, err, "allocating destination vector should succeed")
	defer dstVec.Free()
	require.NoError(t, srcVec.CopyFromHostRaw(rawBN254(src)), "copying source should succeed")
	require.NoError(t, ReduceBlindedCoset(
		dstVec,
		srcVec,
		cloneRaw(rawBN254(tail)),
		cloneRaw(rawBN254([]bnfr.Element{cosetPowN})),
	))
	out := make([]bnfr.Element, len(src))
	require.NoError(t, dstVec.CopyToHostRaw(rawBN254(out)), "copying output should succeed")
	return out
}

func reduceBlindedCosetBLS12377(
	t *testing.T,
	dev *gpu.Device,
	src, tail []blsfr.Element,
	cosetPowN blsfr.Element,
) []blsfr.Element {
	t.Helper()
	srcVec, err := NewFrVector(dev, CurveBLS12377, len(src))
	require.NoError(t, err, "allocating source vector should succeed")
	defer srcVec.Free()
	dstVec, err := NewFrVector(dev, CurveBLS12377, len(src))
	require.NoError(t, err, "allocating destination vector should succeed")
	defer dstVec.Free()
	require.NoError(t, srcVec.CopyFromHostRaw(rawBLS12377(src)), "copying source should succeed")
	require.NoError(t, ReduceBlindedCoset(
		dstVec,
		srcVec,
		cloneRaw(rawBLS12377(tail)),
		cloneRaw(rawBLS12377([]blsfr.Element{cosetPowN})),
	))
	out := make([]blsfr.Element, len(src))
	require.NoError(t, dstVec.CopyToHostRaw(rawBLS12377(out)), "copying output should succeed")
	return out
}

func reduceBlindedCosetBW6761(
	t *testing.T,
	dev *gpu.Device,
	src, tail []bwfr.Element,
	cosetPowN bwfr.Element,
) []bwfr.Element {
	t.Helper()
	srcVec, err := NewFrVector(dev, CurveBW6761, len(src))
	require.NoError(t, err, "allocating source vector should succeed")
	defer srcVec.Free()
	dstVec, err := NewFrVector(dev, CurveBW6761, len(src))
	require.NoError(t, err, "allocating destination vector should succeed")
	defer dstVec.Free()
	require.NoError(t, srcVec.CopyFromHostRaw(rawBW6761(src)), "copying source should succeed")
	require.NoError(t, ReduceBlindedCoset(
		dstVec,
		srcVec,
		cloneRaw(rawBW6761(tail)),
		cloneRaw(rawBW6761([]bwfr.Element{cosetPowN})),
	))
	out := make([]bwfr.Element, len(src))
	require.NoError(t, dstVec.CopyToHostRaw(rawBW6761(out)), "copying output should succeed")
	return out
}

func computeL1DenBN254(t *testing.T, dev *gpu.Device, n int, cosetGen bnfr.Element) []bnfr.Element {
	t.Helper()
	domain, err := NewFFTDomain(dev, fftSpecBN254(n))
	require.NoError(t, err, "allocating domain should succeed")
	defer domain.Free()
	outVec, err := NewFrVector(dev, CurveBN254, n)
	require.NoError(t, err, "allocating L1 denominator vector should succeed")
	defer outVec.Free()
	require.NoError(t, ComputeL1Den(outVec, domain, cloneRaw(rawBN254([]bnfr.Element{cosetGen}))))
	out := make([]bnfr.Element, n)
	require.NoError(t, outVec.CopyToHostRaw(rawBN254(out)), "copying L1 denominator should succeed")
	return out
}

func computeL1DenBLS12377(t *testing.T, dev *gpu.Device, n int, cosetGen blsfr.Element) []blsfr.Element {
	t.Helper()
	domain, err := NewFFTDomain(dev, fftSpecBLS12377(n))
	require.NoError(t, err, "allocating domain should succeed")
	defer domain.Free()
	outVec, err := NewFrVector(dev, CurveBLS12377, n)
	require.NoError(t, err, "allocating L1 denominator vector should succeed")
	defer outVec.Free()
	require.NoError(t, ComputeL1Den(outVec, domain, cloneRaw(rawBLS12377([]blsfr.Element{cosetGen}))))
	out := make([]blsfr.Element, n)
	require.NoError(t, outVec.CopyToHostRaw(rawBLS12377(out)), "copying L1 denominator should succeed")
	return out
}

func computeL1DenBW6761(t *testing.T, dev *gpu.Device, n int, cosetGen bwfr.Element) []bwfr.Element {
	t.Helper()
	domain, err := NewFFTDomain(dev, fftSpecBW6761(n))
	require.NoError(t, err, "allocating domain should succeed")
	defer domain.Free()
	outVec, err := NewFrVector(dev, CurveBW6761, n)
	require.NoError(t, err, "allocating L1 denominator vector should succeed")
	defer outVec.Free()
	require.NoError(t, ComputeL1Den(outVec, domain, cloneRaw(rawBW6761([]bwfr.Element{cosetGen}))))
	out := make([]bwfr.Element, n)
	require.NoError(t, outVec.CopyToHostRaw(rawBW6761(out)), "copying L1 denominator should succeed")
	return out
}

func gateAccumBN254(
	t *testing.T,
	dev *gpu.Device,
	inputs [9][]bnfr.Element,
	zhKInv bnfr.Element,
) []bnfr.Element {
	t.Helper()
	vecs := mustVectors(t, dev, CurveBN254, rawInputsBN254(inputs), len(inputs[0]))
	defer freeFrVectors(vecs...)
	require.NoError(t, PlonkGateAccum(
		vecs[0], vecs[1], vecs[2], vecs[3], vecs[4], vecs[5], vecs[6], vecs[7], vecs[8],
		cloneRaw(rawBN254([]bnfr.Element{zhKInv})),
	))
	out := make([]bnfr.Element, len(inputs[0]))
	require.NoError(t, vecs[0].CopyToHostRaw(rawBN254(out)), "copying gate output should succeed")
	return out
}

func gateAccumBLS12377(
	t *testing.T,
	dev *gpu.Device,
	inputs [9][]blsfr.Element,
	zhKInv blsfr.Element,
) []blsfr.Element {
	t.Helper()
	vecs := mustVectors(t, dev, CurveBLS12377, rawInputsBLS12377(inputs), len(inputs[0]))
	defer freeFrVectors(vecs...)
	require.NoError(t, PlonkGateAccum(
		vecs[0], vecs[1], vecs[2], vecs[3], vecs[4], vecs[5], vecs[6], vecs[7], vecs[8],
		cloneRaw(rawBLS12377([]blsfr.Element{zhKInv})),
	))
	out := make([]blsfr.Element, len(inputs[0]))
	require.NoError(t, vecs[0].CopyToHostRaw(rawBLS12377(out)), "copying gate output should succeed")
	return out
}

func gateAccumBW6761(
	t *testing.T,
	dev *gpu.Device,
	inputs [9][]bwfr.Element,
	zhKInv bwfr.Element,
) []bwfr.Element {
	t.Helper()
	vecs := mustVectors(t, dev, CurveBW6761, rawInputsBW6761(inputs), len(inputs[0]))
	defer freeFrVectors(vecs...)
	require.NoError(t, PlonkGateAccum(
		vecs[0], vecs[1], vecs[2], vecs[3], vecs[4], vecs[5], vecs[6], vecs[7], vecs[8],
		cloneRaw(rawBW6761([]bwfr.Element{zhKInv})),
	))
	out := make([]bwfr.Element, len(inputs[0]))
	require.NoError(t, vecs[0].CopyToHostRaw(rawBW6761(out)), "copying gate output should succeed")
	return out
}

func permBoundaryBN254(
	t *testing.T,
	dev *gpu.Device,
	inputs [8][]bnfr.Element,
	alpha, beta, gamma, l1Scalar, cosetShift, cosetShiftSq, cosetGen bnfr.Element,
) []bnfr.Element {
	t.Helper()
	n := len(inputs[0])
	domain, err := NewFFTDomain(dev, fftSpecBN254(n))
	require.NoError(t, err, "allocating domain should succeed")
	defer domain.Free()
	raws := [][]uint64{
		make([]uint64, n*4),
		rawBN254(inputs[0]), rawBN254(inputs[1]), rawBN254(inputs[2]), rawBN254(inputs[3]),
		rawBN254(inputs[4]), rawBN254(inputs[5]), rawBN254(inputs[6]), rawBN254(inputs[7]),
	}
	vecs := mustVectors(t, dev, CurveBN254, raws, n)
	defer freeFrVectors(vecs...)
	require.NoError(t, PlonkPermBoundary(
		vecs[0], vecs[1], vecs[2], vecs[3], vecs[4], vecs[5], vecs[6], vecs[7], vecs[8],
		domain,
		cloneRaw(rawBN254([]bnfr.Element{alpha})),
		cloneRaw(rawBN254([]bnfr.Element{beta})),
		cloneRaw(rawBN254([]bnfr.Element{gamma})),
		cloneRaw(rawBN254([]bnfr.Element{l1Scalar})),
		cloneRaw(rawBN254([]bnfr.Element{cosetShift})),
		cloneRaw(rawBN254([]bnfr.Element{cosetShiftSq})),
		cloneRaw(rawBN254([]bnfr.Element{cosetGen})),
	))
	out := make([]bnfr.Element, n)
	require.NoError(t, vecs[0].CopyToHostRaw(rawBN254(out)), "copying permutation output should succeed")
	return out
}

func permBoundaryBLS12377(
	t *testing.T,
	dev *gpu.Device,
	inputs [8][]blsfr.Element,
	alpha, beta, gamma, l1Scalar, cosetShift, cosetShiftSq, cosetGen blsfr.Element,
) []blsfr.Element {
	t.Helper()
	n := len(inputs[0])
	domain, err := NewFFTDomain(dev, fftSpecBLS12377(n))
	require.NoError(t, err, "allocating domain should succeed")
	defer domain.Free()
	raws := [][]uint64{
		make([]uint64, n*4),
		rawBLS12377(inputs[0]), rawBLS12377(inputs[1]), rawBLS12377(inputs[2]), rawBLS12377(inputs[3]),
		rawBLS12377(inputs[4]), rawBLS12377(inputs[5]), rawBLS12377(inputs[6]), rawBLS12377(inputs[7]),
	}
	vecs := mustVectors(t, dev, CurveBLS12377, raws, n)
	defer freeFrVectors(vecs...)
	require.NoError(t, PlonkPermBoundary(
		vecs[0], vecs[1], vecs[2], vecs[3], vecs[4], vecs[5], vecs[6], vecs[7], vecs[8],
		domain,
		cloneRaw(rawBLS12377([]blsfr.Element{alpha})),
		cloneRaw(rawBLS12377([]blsfr.Element{beta})),
		cloneRaw(rawBLS12377([]blsfr.Element{gamma})),
		cloneRaw(rawBLS12377([]blsfr.Element{l1Scalar})),
		cloneRaw(rawBLS12377([]blsfr.Element{cosetShift})),
		cloneRaw(rawBLS12377([]blsfr.Element{cosetShiftSq})),
		cloneRaw(rawBLS12377([]blsfr.Element{cosetGen})),
	))
	out := make([]blsfr.Element, n)
	require.NoError(t, vecs[0].CopyToHostRaw(rawBLS12377(out)), "copying permutation output should succeed")
	return out
}

func permBoundaryBW6761(
	t *testing.T,
	dev *gpu.Device,
	inputs [8][]bwfr.Element,
	alpha, beta, gamma, l1Scalar, cosetShift, cosetShiftSq, cosetGen bwfr.Element,
) []bwfr.Element {
	t.Helper()
	n := len(inputs[0])
	domain, err := NewFFTDomain(dev, fftSpecBW6761(n))
	require.NoError(t, err, "allocating domain should succeed")
	defer domain.Free()
	raws := [][]uint64{
		make([]uint64, n*6),
		rawBW6761(inputs[0]), rawBW6761(inputs[1]), rawBW6761(inputs[2]), rawBW6761(inputs[3]),
		rawBW6761(inputs[4]), rawBW6761(inputs[5]), rawBW6761(inputs[6]), rawBW6761(inputs[7]),
	}
	vecs := mustVectors(t, dev, CurveBW6761, raws, n)
	defer freeFrVectors(vecs...)
	require.NoError(t, PlonkPermBoundary(
		vecs[0], vecs[1], vecs[2], vecs[3], vecs[4], vecs[5], vecs[6], vecs[7], vecs[8],
		domain,
		cloneRaw(rawBW6761([]bwfr.Element{alpha})),
		cloneRaw(rawBW6761([]bwfr.Element{beta})),
		cloneRaw(rawBW6761([]bwfr.Element{gamma})),
		cloneRaw(rawBW6761([]bwfr.Element{l1Scalar})),
		cloneRaw(rawBW6761([]bwfr.Element{cosetShift})),
		cloneRaw(rawBW6761([]bwfr.Element{cosetShiftSq})),
		cloneRaw(rawBW6761([]bwfr.Element{cosetGen})),
	))
	out := make([]bwfr.Element, n)
	require.NoError(t, vecs[0].CopyToHostRaw(rawBW6761(out)), "copying permutation output should succeed")
	return out
}

func zFactorsBN254(
	t *testing.T,
	dev *gpu.Device,
	l, r, o []bnfr.Element,
	perm []int64,
	beta, gamma, cosetShift, cosetShiftSq bnfr.Element,
) ([]bnfr.Element, []bnfr.Element) {
	t.Helper()
	n := len(l)
	domain, err := NewFFTDomain(dev, fftSpecBN254(n))
	require.NoError(t, err, "allocating domain should succeed")
	defer domain.Free()
	permDev, err := NewDeviceInt64(dev, perm)
	require.NoError(t, err, "uploading permutation should succeed")
	defer permDev.Free()
	vecs := mustVectors(t, dev, CurveBN254, [][]uint64{rawBN254(l), rawBN254(r), rawBN254(o)}, n)
	defer freeFrVectors(vecs...)
	require.NoError(t, PlonkZComputeFactors(
		vecs[0], vecs[1], vecs[2], permDev, domain,
		cloneRaw(rawBN254([]bnfr.Element{beta})),
		cloneRaw(rawBN254([]bnfr.Element{gamma})),
		cloneRaw(rawBN254([]bnfr.Element{cosetShift})),
		cloneRaw(rawBN254([]bnfr.Element{cosetShiftSq})),
		uint(log2Exact(n)),
	))
	num := make([]bnfr.Element, n)
	den := make([]bnfr.Element, n)
	require.NoError(t, vecs[0].CopyToHostRaw(rawBN254(num)), "copying numerator should succeed")
	require.NoError(t, vecs[1].CopyToHostRaw(rawBN254(den)), "copying denominator should succeed")
	return num, den
}

func zFactorsBLS12377(
	t *testing.T,
	dev *gpu.Device,
	l, r, o []blsfr.Element,
	perm []int64,
	beta, gamma, cosetShift, cosetShiftSq blsfr.Element,
) ([]blsfr.Element, []blsfr.Element) {
	t.Helper()
	n := len(l)
	domain, err := NewFFTDomain(dev, fftSpecBLS12377(n))
	require.NoError(t, err, "allocating domain should succeed")
	defer domain.Free()
	permDev, err := NewDeviceInt64(dev, perm)
	require.NoError(t, err, "uploading permutation should succeed")
	defer permDev.Free()
	vecs := mustVectors(t, dev, CurveBLS12377, [][]uint64{rawBLS12377(l), rawBLS12377(r), rawBLS12377(o)}, n)
	defer freeFrVectors(vecs...)
	require.NoError(t, PlonkZComputeFactors(
		vecs[0], vecs[1], vecs[2], permDev, domain,
		cloneRaw(rawBLS12377([]blsfr.Element{beta})),
		cloneRaw(rawBLS12377([]blsfr.Element{gamma})),
		cloneRaw(rawBLS12377([]blsfr.Element{cosetShift})),
		cloneRaw(rawBLS12377([]blsfr.Element{cosetShiftSq})),
		uint(log2Exact(n)),
	))
	num := make([]blsfr.Element, n)
	den := make([]blsfr.Element, n)
	require.NoError(t, vecs[0].CopyToHostRaw(rawBLS12377(num)), "copying numerator should succeed")
	require.NoError(t, vecs[1].CopyToHostRaw(rawBLS12377(den)), "copying denominator should succeed")
	return num, den
}

func zFactorsBW6761(
	t *testing.T,
	dev *gpu.Device,
	l, r, o []bwfr.Element,
	perm []int64,
	beta, gamma, cosetShift, cosetShiftSq bwfr.Element,
) ([]bwfr.Element, []bwfr.Element) {
	t.Helper()
	n := len(l)
	domain, err := NewFFTDomain(dev, fftSpecBW6761(n))
	require.NoError(t, err, "allocating domain should succeed")
	defer domain.Free()
	permDev, err := NewDeviceInt64(dev, perm)
	require.NoError(t, err, "uploading permutation should succeed")
	defer permDev.Free()
	vecs := mustVectors(t, dev, CurveBW6761, [][]uint64{rawBW6761(l), rawBW6761(r), rawBW6761(o)}, n)
	defer freeFrVectors(vecs...)
	require.NoError(t, PlonkZComputeFactors(
		vecs[0], vecs[1], vecs[2], permDev, domain,
		cloneRaw(rawBW6761([]bwfr.Element{beta})),
		cloneRaw(rawBW6761([]bwfr.Element{gamma})),
		cloneRaw(rawBW6761([]bwfr.Element{cosetShift})),
		cloneRaw(rawBW6761([]bwfr.Element{cosetShiftSq})),
		uint(log2Exact(n)),
	))
	num := make([]bwfr.Element, n)
	den := make([]bwfr.Element, n)
	require.NoError(t, vecs[0].CopyToHostRaw(rawBW6761(num)), "copying numerator should succeed")
	require.NoError(t, vecs[1].CopyToHostRaw(rawBW6761(den)), "copying denominator should succeed")
	return num, den
}

func zPrefixBN254(t *testing.T, dev *gpu.Device, ratio []bnfr.Element) []bnfr.Element {
	t.Helper()
	n := len(ratio)
	vecs := mustVectors(t, dev, CurveBN254, [][]uint64{
		make([]uint64, n*4),
		rawBN254(ratio),
		make([]uint64, n*4),
	}, n)
	defer freeFrVectors(vecs...)
	require.NoError(t, ZPrefixProduct(vecs[0], vecs[1], vecs[2]), "Z prefix should succeed")
	out := make([]bnfr.Element, n)
	require.NoError(t, vecs[0].CopyToHostRaw(rawBN254(out)), "copying Z prefix should succeed")
	return out
}

func zPrefixBLS12377(t *testing.T, dev *gpu.Device, ratio []blsfr.Element) []blsfr.Element {
	t.Helper()
	n := len(ratio)
	vecs := mustVectors(t, dev, CurveBLS12377, [][]uint64{
		make([]uint64, n*4),
		rawBLS12377(ratio),
		make([]uint64, n*4),
	}, n)
	defer freeFrVectors(vecs...)
	require.NoError(t, ZPrefixProduct(vecs[0], vecs[1], vecs[2]), "Z prefix should succeed")
	out := make([]blsfr.Element, n)
	require.NoError(t, vecs[0].CopyToHostRaw(rawBLS12377(out)), "copying Z prefix should succeed")
	return out
}

func zPrefixBW6761(t *testing.T, dev *gpu.Device, ratio []bwfr.Element) []bwfr.Element {
	t.Helper()
	n := len(ratio)
	vecs := mustVectors(t, dev, CurveBW6761, [][]uint64{
		make([]uint64, n*6),
		rawBW6761(ratio),
		make([]uint64, n*6),
	}, n)
	defer freeFrVectors(vecs...)
	require.NoError(t, ZPrefixProduct(vecs[0], vecs[1], vecs[2]), "Z prefix should succeed")
	out := make([]bwfr.Element, n)
	require.NoError(t, vecs[0].CopyToHostRaw(rawBW6761(out)), "copying Z prefix should succeed")
	return out
}

func mustButterflyVectors(
	t *testing.T,
	dev *gpu.Device,
	curve Curve,
	b0Raw, b1Raw, b2Raw, b3Raw []uint64,
	n int,
) [4]*FrVector {
	t.Helper()
	raws := [4][]uint64{b0Raw, b1Raw, b2Raw, b3Raw}
	var vecs [4]*FrVector
	for i := range vecs {
		v, err := NewFrVector(dev, curve, n)
		require.NoError(t, err, "allocating butterfly vector should succeed")
		require.NoError(t, v.CopyFromHostRaw(raws[i]), "copying butterfly input should succeed")
		vecs[i] = v
	}
	return vecs
}

func mustVectors(t *testing.T, dev *gpu.Device, curve Curve, raws [][]uint64, n int) []*FrVector {
	t.Helper()
	vecs := make([]*FrVector, len(raws))
	for i := range vecs {
		v, err := NewFrVector(dev, curve, n)
		require.NoError(t, err, "allocating vector should succeed")
		require.NoError(t, v.CopyFromHostRaw(raws[i]), "copying input should succeed")
		vecs[i] = v
	}
	return vecs
}

func freeFrVectors(vecs ...*FrVector) {
	for _, v := range vecs {
		if v != nil {
			v.Free()
		}
	}
}

func butterfly4BN254Expected(
	n int,
	b0, b1, b2, b3 []bnfr.Element,
) ([]bnfr.Element, []bnfr.Element, []bnfr.Element, []bnfr.Element) {
	omega4Inv, quarter := butterfly4ScalarsBN254(n)
	o0, o1, o2, o3 := make([]bnfr.Element, n), make([]bnfr.Element, n), make([]bnfr.Element, n), make([]bnfr.Element, n)
	for i := range o0 {
		var t0, t1, t2, t3 bnfr.Element
		t0.Add(&b0[i], &b2[i])
		t1.Sub(&b0[i], &b2[i])
		t2.Add(&b1[i], &b3[i])
		t3.Sub(&b1[i], &b3[i]).Mul(&t3, &omega4Inv)
		o0[i].Add(&t0, &t2).Mul(&o0[i], &quarter)
		o1[i].Add(&t1, &t3).Mul(&o1[i], &quarter)
		o2[i].Sub(&t0, &t2).Mul(&o2[i], &quarter)
		o3[i].Sub(&t1, &t3).Mul(&o3[i], &quarter)
	}
	return o0, o1, o2, o3
}

func butterfly4BLS12377Expected(
	n int,
	b0, b1, b2, b3 []blsfr.Element,
) ([]blsfr.Element, []blsfr.Element, []blsfr.Element, []blsfr.Element) {
	omega4Inv, quarter := butterfly4ScalarsBLS12377(n)
	o0, o1, o2, o3 := make([]blsfr.Element, n), make([]blsfr.Element, n), make([]blsfr.Element, n), make([]blsfr.Element, n)
	for i := range o0 {
		var t0, t1, t2, t3 blsfr.Element
		t0.Add(&b0[i], &b2[i])
		t1.Sub(&b0[i], &b2[i])
		t2.Add(&b1[i], &b3[i])
		t3.Sub(&b1[i], &b3[i]).Mul(&t3, &omega4Inv)
		o0[i].Add(&t0, &t2).Mul(&o0[i], &quarter)
		o1[i].Add(&t1, &t3).Mul(&o1[i], &quarter)
		o2[i].Sub(&t0, &t2).Mul(&o2[i], &quarter)
		o3[i].Sub(&t1, &t3).Mul(&o3[i], &quarter)
	}
	return o0, o1, o2, o3
}

func butterfly4BW6761Expected(
	n int,
	b0, b1, b2, b3 []bwfr.Element,
) ([]bwfr.Element, []bwfr.Element, []bwfr.Element, []bwfr.Element) {
	omega4Inv, quarter := butterfly4ScalarsBW6761(n)
	o0, o1, o2, o3 := make([]bwfr.Element, n), make([]bwfr.Element, n), make([]bwfr.Element, n), make([]bwfr.Element, n)
	for i := range o0 {
		var t0, t1, t2, t3 bwfr.Element
		t0.Add(&b0[i], &b2[i])
		t1.Sub(&b0[i], &b2[i])
		t2.Add(&b1[i], &b3[i])
		t3.Sub(&b1[i], &b3[i]).Mul(&t3, &omega4Inv)
		o0[i].Add(&t0, &t2).Mul(&o0[i], &quarter)
		o1[i].Add(&t1, &t3).Mul(&o1[i], &quarter)
		o2[i].Sub(&t0, &t2).Mul(&o2[i], &quarter)
		o3[i].Sub(&t1, &t3).Mul(&o3[i], &quarter)
	}
	return o0, o1, o2, o3
}

func gateAccumExpectedBN254(inputs [9][]bnfr.Element, zhKInv bnfr.Element) []bnfr.Element {
	out := append([]bnfr.Element(nil), inputs[0]...)
	for i := range out {
		var tmp, lr bnfr.Element
		tmp.Mul(&inputs[1][i], &inputs[6][i])
		out[i].Add(&out[i], &tmp)
		tmp.Mul(&inputs[2][i], &inputs[7][i])
		out[i].Add(&out[i], &tmp)
		lr.Mul(&inputs[6][i], &inputs[7][i])
		tmp.Mul(&inputs[3][i], &lr)
		out[i].Add(&out[i], &tmp)
		tmp.Mul(&inputs[4][i], &inputs[8][i])
		out[i].Add(&out[i], &tmp)
		out[i].Add(&out[i], &inputs[5][i])
		out[i].Mul(&out[i], &zhKInv)
	}
	return out
}

func gateAccumExpectedBLS12377(inputs [9][]blsfr.Element, zhKInv blsfr.Element) []blsfr.Element {
	out := append([]blsfr.Element(nil), inputs[0]...)
	for i := range out {
		var tmp, lr blsfr.Element
		tmp.Mul(&inputs[1][i], &inputs[6][i])
		out[i].Add(&out[i], &tmp)
		tmp.Mul(&inputs[2][i], &inputs[7][i])
		out[i].Add(&out[i], &tmp)
		lr.Mul(&inputs[6][i], &inputs[7][i])
		tmp.Mul(&inputs[3][i], &lr)
		out[i].Add(&out[i], &tmp)
		tmp.Mul(&inputs[4][i], &inputs[8][i])
		out[i].Add(&out[i], &tmp)
		out[i].Add(&out[i], &inputs[5][i])
		out[i].Mul(&out[i], &zhKInv)
	}
	return out
}

func gateAccumExpectedBW6761(inputs [9][]bwfr.Element, zhKInv bwfr.Element) []bwfr.Element {
	out := append([]bwfr.Element(nil), inputs[0]...)
	for i := range out {
		var tmp, lr bwfr.Element
		tmp.Mul(&inputs[1][i], &inputs[6][i])
		out[i].Add(&out[i], &tmp)
		tmp.Mul(&inputs[2][i], &inputs[7][i])
		out[i].Add(&out[i], &tmp)
		lr.Mul(&inputs[6][i], &inputs[7][i])
		tmp.Mul(&inputs[3][i], &lr)
		out[i].Add(&out[i], &tmp)
		tmp.Mul(&inputs[4][i], &inputs[8][i])
		out[i].Add(&out[i], &tmp)
		out[i].Add(&out[i], &inputs[5][i])
		out[i].Mul(&out[i], &zhKInv)
	}
	return out
}

func permBoundaryExpectedBN254(
	inputs [8][]bnfr.Element,
	omega, alpha, beta, gamma, l1Scalar, cosetShift, cosetShiftSq, cosetGen bnfr.Element,
) []bnfr.Element {
	out := make([]bnfr.Element, len(inputs[0]))
	omegaI := new(bnfr.Element).SetOne()
	one := new(bnfr.Element).SetOne()
	for i := range out {
		var xI, id1, id2, id3 bnfr.Element
		xI.Mul(&cosetGen, omegaI)
		id1.Mul(&beta, &xI)
		id2.Mul(&id1, &cosetShift)
		id3.Mul(&id1, &cosetShiftSq)

		var t1, t2, t3, num, den, tmp bnfr.Element
		t1.Add(&inputs[0][i], &id1).Add(&t1, &gamma)
		t2.Add(&inputs[1][i], &id2).Add(&t2, &gamma)
		t3.Add(&inputs[2][i], &id3).Add(&t3, &gamma)
		num.Mul(&inputs[3][i], &t1).Mul(&num, &t2).Mul(&num, &t3)

		tmp.Mul(&beta, &inputs[4][i])
		t1.Add(&inputs[0][i], &tmp).Add(&t1, &gamma)
		tmp.Mul(&beta, &inputs[5][i])
		t2.Add(&inputs[1][i], &tmp).Add(&t2, &gamma)
		tmp.Mul(&beta, &inputs[6][i])
		t3.Add(&inputs[2][i], &tmp).Add(&t3, &gamma)
		den.Mul(&inputs[3][(i+1)%len(out)], &t1).Mul(&den, &t2).Mul(&den, &t3)

		var ordering, l1, local, alphaLocal, sum, zMinusOne bnfr.Element
		ordering.Sub(&den, &num)
		l1.Mul(&l1Scalar, &inputs[7][i])
		zMinusOne.Sub(&inputs[3][i], one)
		local.Mul(&zMinusOne, &l1)
		alphaLocal.Mul(&alpha, &local)
		sum.Add(&ordering, &alphaLocal)
		out[i].Mul(&alpha, &sum)
		omegaI.Mul(omegaI, &omega)
	}
	return out
}

func permBoundaryExpectedBLS12377(
	inputs [8][]blsfr.Element,
	omega, alpha, beta, gamma, l1Scalar, cosetShift, cosetShiftSq, cosetGen blsfr.Element,
) []blsfr.Element {
	out := make([]blsfr.Element, len(inputs[0]))
	omegaI := new(blsfr.Element).SetOne()
	one := new(blsfr.Element).SetOne()
	for i := range out {
		var xI, id1, id2, id3 blsfr.Element
		xI.Mul(&cosetGen, omegaI)
		id1.Mul(&beta, &xI)
		id2.Mul(&id1, &cosetShift)
		id3.Mul(&id1, &cosetShiftSq)

		var t1, t2, t3, num, den, tmp blsfr.Element
		t1.Add(&inputs[0][i], &id1).Add(&t1, &gamma)
		t2.Add(&inputs[1][i], &id2).Add(&t2, &gamma)
		t3.Add(&inputs[2][i], &id3).Add(&t3, &gamma)
		num.Mul(&inputs[3][i], &t1).Mul(&num, &t2).Mul(&num, &t3)

		tmp.Mul(&beta, &inputs[4][i])
		t1.Add(&inputs[0][i], &tmp).Add(&t1, &gamma)
		tmp.Mul(&beta, &inputs[5][i])
		t2.Add(&inputs[1][i], &tmp).Add(&t2, &gamma)
		tmp.Mul(&beta, &inputs[6][i])
		t3.Add(&inputs[2][i], &tmp).Add(&t3, &gamma)
		den.Mul(&inputs[3][(i+1)%len(out)], &t1).Mul(&den, &t2).Mul(&den, &t3)

		var ordering, l1, local, alphaLocal, sum, zMinusOne blsfr.Element
		ordering.Sub(&den, &num)
		l1.Mul(&l1Scalar, &inputs[7][i])
		zMinusOne.Sub(&inputs[3][i], one)
		local.Mul(&zMinusOne, &l1)
		alphaLocal.Mul(&alpha, &local)
		sum.Add(&ordering, &alphaLocal)
		out[i].Mul(&alpha, &sum)
		omegaI.Mul(omegaI, &omega)
	}
	return out
}

func permBoundaryExpectedBW6761(
	inputs [8][]bwfr.Element,
	omega, alpha, beta, gamma, l1Scalar, cosetShift, cosetShiftSq, cosetGen bwfr.Element,
) []bwfr.Element {
	out := make([]bwfr.Element, len(inputs[0]))
	omegaI := new(bwfr.Element).SetOne()
	one := new(bwfr.Element).SetOne()
	for i := range out {
		var xI, id1, id2, id3 bwfr.Element
		xI.Mul(&cosetGen, omegaI)
		id1.Mul(&beta, &xI)
		id2.Mul(&id1, &cosetShift)
		id3.Mul(&id1, &cosetShiftSq)

		var t1, t2, t3, num, den, tmp bwfr.Element
		t1.Add(&inputs[0][i], &id1).Add(&t1, &gamma)
		t2.Add(&inputs[1][i], &id2).Add(&t2, &gamma)
		t3.Add(&inputs[2][i], &id3).Add(&t3, &gamma)
		num.Mul(&inputs[3][i], &t1).Mul(&num, &t2).Mul(&num, &t3)

		tmp.Mul(&beta, &inputs[4][i])
		t1.Add(&inputs[0][i], &tmp).Add(&t1, &gamma)
		tmp.Mul(&beta, &inputs[5][i])
		t2.Add(&inputs[1][i], &tmp).Add(&t2, &gamma)
		tmp.Mul(&beta, &inputs[6][i])
		t3.Add(&inputs[2][i], &tmp).Add(&t3, &gamma)
		den.Mul(&inputs[3][(i+1)%len(out)], &t1).Mul(&den, &t2).Mul(&den, &t3)

		var ordering, l1, local, alphaLocal, sum, zMinusOne bwfr.Element
		ordering.Sub(&den, &num)
		l1.Mul(&l1Scalar, &inputs[7][i])
		zMinusOne.Sub(&inputs[3][i], one)
		local.Mul(&zMinusOne, &l1)
		alphaLocal.Mul(&alpha, &local)
		sum.Add(&ordering, &alphaLocal)
		out[i].Mul(&alpha, &sum)
		omegaI.Mul(omegaI, &omega)
	}
	return out
}

func zFactorsExpectedBN254(
	l, r, o []bnfr.Element,
	perm []int64,
	omega, beta, gamma, cosetShift, cosetShiftSq bnfr.Element,
) ([]bnfr.Element, []bnfr.Element) {
	n := len(l)
	num := make([]bnfr.Element, n)
	den := make([]bnfr.Element, n)
	omegaI := new(bnfr.Element).SetOne()
	log2n := log2Exact(n)
	for i := range num {
		var betaID0, betaID1, betaID2 bnfr.Element
		betaID0.Mul(&beta, omegaI)
		betaID1.Mul(&cosetShift, &betaID0)
		betaID2.Mul(&cosetShiftSq, &betaID0)
		var t1, t2, t3, tmp bnfr.Element
		t1.Add(&l[i], &betaID0).Add(&t1, &gamma)
		t2.Add(&r[i], &betaID1).Add(&t2, &gamma)
		t3.Add(&o[i], &betaID2).Add(&t3, &gamma)
		num[i].Mul(&t1, &t2).Mul(&num[i], &t3)

		sid0 := permIdentityBN254(perm[i], n, log2n, omega, cosetShift, cosetShiftSq)
		sid1 := permIdentityBN254(perm[n+i], n, log2n, omega, cosetShift, cosetShiftSq)
		sid2 := permIdentityBN254(perm[2*n+i], n, log2n, omega, cosetShift, cosetShiftSq)
		tmp.Mul(&beta, &sid0)
		t1.Add(&l[i], &tmp).Add(&t1, &gamma)
		tmp.Mul(&beta, &sid1)
		t2.Add(&r[i], &tmp).Add(&t2, &gamma)
		tmp.Mul(&beta, &sid2)
		t3.Add(&o[i], &tmp).Add(&t3, &gamma)
		den[i].Mul(&t1, &t2).Mul(&den[i], &t3)
		omegaI.Mul(omegaI, &omega)
	}
	return num, den
}

func zFactorsExpectedBLS12377(
	l, r, o []blsfr.Element,
	perm []int64,
	omega, beta, gamma, cosetShift, cosetShiftSq blsfr.Element,
) ([]blsfr.Element, []blsfr.Element) {
	n := len(l)
	num := make([]blsfr.Element, n)
	den := make([]blsfr.Element, n)
	omegaI := new(blsfr.Element).SetOne()
	log2n := log2Exact(n)
	for i := range num {
		var betaID0, betaID1, betaID2 blsfr.Element
		betaID0.Mul(&beta, omegaI)
		betaID1.Mul(&cosetShift, &betaID0)
		betaID2.Mul(&cosetShiftSq, &betaID0)
		var t1, t2, t3, tmp blsfr.Element
		t1.Add(&l[i], &betaID0).Add(&t1, &gamma)
		t2.Add(&r[i], &betaID1).Add(&t2, &gamma)
		t3.Add(&o[i], &betaID2).Add(&t3, &gamma)
		num[i].Mul(&t1, &t2).Mul(&num[i], &t3)

		sid0 := permIdentityBLS12377(perm[i], n, log2n, omega, cosetShift, cosetShiftSq)
		sid1 := permIdentityBLS12377(perm[n+i], n, log2n, omega, cosetShift, cosetShiftSq)
		sid2 := permIdentityBLS12377(perm[2*n+i], n, log2n, omega, cosetShift, cosetShiftSq)
		tmp.Mul(&beta, &sid0)
		t1.Add(&l[i], &tmp).Add(&t1, &gamma)
		tmp.Mul(&beta, &sid1)
		t2.Add(&r[i], &tmp).Add(&t2, &gamma)
		tmp.Mul(&beta, &sid2)
		t3.Add(&o[i], &tmp).Add(&t3, &gamma)
		den[i].Mul(&t1, &t2).Mul(&den[i], &t3)
		omegaI.Mul(omegaI, &omega)
	}
	return num, den
}

func zFactorsExpectedBW6761(
	l, r, o []bwfr.Element,
	perm []int64,
	omega, beta, gamma, cosetShift, cosetShiftSq bwfr.Element,
) ([]bwfr.Element, []bwfr.Element) {
	n := len(l)
	num := make([]bwfr.Element, n)
	den := make([]bwfr.Element, n)
	omegaI := new(bwfr.Element).SetOne()
	log2n := log2Exact(n)
	for i := range num {
		var betaID0, betaID1, betaID2 bwfr.Element
		betaID0.Mul(&beta, omegaI)
		betaID1.Mul(&cosetShift, &betaID0)
		betaID2.Mul(&cosetShiftSq, &betaID0)
		var t1, t2, t3, tmp bwfr.Element
		t1.Add(&l[i], &betaID0).Add(&t1, &gamma)
		t2.Add(&r[i], &betaID1).Add(&t2, &gamma)
		t3.Add(&o[i], &betaID2).Add(&t3, &gamma)
		num[i].Mul(&t1, &t2).Mul(&num[i], &t3)

		sid0 := permIdentityBW6761(perm[i], n, log2n, omega, cosetShift, cosetShiftSq)
		sid1 := permIdentityBW6761(perm[n+i], n, log2n, omega, cosetShift, cosetShiftSq)
		sid2 := permIdentityBW6761(perm[2*n+i], n, log2n, omega, cosetShift, cosetShiftSq)
		tmp.Mul(&beta, &sid0)
		t1.Add(&l[i], &tmp).Add(&t1, &gamma)
		tmp.Mul(&beta, &sid1)
		t2.Add(&r[i], &tmp).Add(&t2, &gamma)
		tmp.Mul(&beta, &sid2)
		t3.Add(&o[i], &tmp).Add(&t3, &gamma)
		den[i].Mul(&t1, &t2).Mul(&den[i], &t3)
		omegaI.Mul(omegaI, &omega)
	}
	return num, den
}

func zPrefixExpectedBN254(ratio []bnfr.Element) []bnfr.Element {
	out := make([]bnfr.Element, len(ratio))
	var acc bnfr.Element
	acc.SetOne()
	for i := range out {
		out[i].Set(&acc)
		acc.Mul(&acc, &ratio[i])
	}
	return out
}

func zPrefixExpectedBLS12377(ratio []blsfr.Element) []blsfr.Element {
	out := make([]blsfr.Element, len(ratio))
	var acc blsfr.Element
	acc.SetOne()
	for i := range out {
		out[i].Set(&acc)
		acc.Mul(&acc, &ratio[i])
	}
	return out
}

func zPrefixExpectedBW6761(ratio []bwfr.Element) []bwfr.Element {
	out := make([]bwfr.Element, len(ratio))
	var acc bwfr.Element
	acc.SetOne()
	for i := range out {
		out[i].Set(&acc)
		acc.Mul(&acc, &ratio[i])
	}
	return out
}

func permIdentityBN254(
	idx int64,
	n int,
	log2n int,
	omega, cosetShift, cosetShiftSq bnfr.Element,
) bnfr.Element {
	pos := int(idx) & (n - 1)
	coset := int(idx) >> log2n
	var out bnfr.Element
	out.Exp(omega, big.NewInt(int64(pos)))
	if coset == 1 {
		out.Mul(&out, &cosetShift)
	} else if coset == 2 {
		out.Mul(&out, &cosetShiftSq)
	}
	return out
}

func permIdentityBLS12377(
	idx int64,
	n int,
	log2n int,
	omega, cosetShift, cosetShiftSq blsfr.Element,
) blsfr.Element {
	pos := int(idx) & (n - 1)
	coset := int(idx) >> log2n
	var out blsfr.Element
	out.Exp(omega, big.NewInt(int64(pos)))
	if coset == 1 {
		out.Mul(&out, &cosetShift)
	} else if coset == 2 {
		out.Mul(&out, &cosetShiftSq)
	}
	return out
}

func permIdentityBW6761(
	idx int64,
	n int,
	log2n int,
	omega, cosetShift, cosetShiftSq bwfr.Element,
) bwfr.Element {
	pos := int(idx) & (n - 1)
	coset := int(idx) >> log2n
	var out bwfr.Element
	out.Exp(omega, big.NewInt(int64(pos)))
	if coset == 1 {
		out.Mul(&out, &cosetShift)
	} else if coset == 2 {
		out.Mul(&out, &cosetShiftSq)
	}
	return out
}

func permBoundaryScalarsBN254() (bnfr.Element, bnfr.Element, bnfr.Element, bnfr.Element, bnfr.Element, bnfr.Element, bnfr.Element) {
	var alpha, beta, gamma, l1Scalar, cosetShift, cosetShiftSq, cosetGen bnfr.Element
	alpha.SetUint64(3)
	beta.SetUint64(5)
	gamma.SetUint64(7)
	l1Scalar.SetUint64(11)
	cosetShift.SetUint64(13)
	cosetShiftSq.Mul(&cosetShift, &cosetShift)
	cosetGen.SetUint64(17)
	return alpha, beta, gamma, l1Scalar, cosetShift, cosetShiftSq, cosetGen
}

func permBoundaryScalarsBLS12377() (blsfr.Element, blsfr.Element, blsfr.Element, blsfr.Element, blsfr.Element, blsfr.Element, blsfr.Element) {
	var alpha, beta, gamma, l1Scalar, cosetShift, cosetShiftSq, cosetGen blsfr.Element
	alpha.SetUint64(3)
	beta.SetUint64(5)
	gamma.SetUint64(7)
	l1Scalar.SetUint64(11)
	cosetShift.SetUint64(13)
	cosetShiftSq.Mul(&cosetShift, &cosetShift)
	cosetGen.SetUint64(17)
	return alpha, beta, gamma, l1Scalar, cosetShift, cosetShiftSq, cosetGen
}

func permBoundaryScalarsBW6761() (bwfr.Element, bwfr.Element, bwfr.Element, bwfr.Element, bwfr.Element, bwfr.Element, bwfr.Element) {
	var alpha, beta, gamma, l1Scalar, cosetShift, cosetShiftSq, cosetGen bwfr.Element
	alpha.SetUint64(3)
	beta.SetUint64(5)
	gamma.SetUint64(7)
	l1Scalar.SetUint64(11)
	cosetShift.SetUint64(13)
	cosetShiftSq.Mul(&cosetShift, &cosetShift)
	cosetGen.SetUint64(17)
	return alpha, beta, gamma, l1Scalar, cosetShift, cosetShiftSq, cosetGen
}

func deterministicPermutation(n int) []int64 {
	out := make([]int64, 3*n)
	mod := 3 * n
	for i := range out {
		out[i] = int64((7*i + 5) % mod)
	}
	return out
}

func log2Exact(n int) int {
	log := 0
	for (1 << log) < n {
		log++
	}
	return log
}

func rawInputsBN254(inputs [9][]bnfr.Element) [][]uint64 {
	return [][]uint64{
		rawBN254(inputs[0]), rawBN254(inputs[1]), rawBN254(inputs[2]),
		rawBN254(inputs[3]), rawBN254(inputs[4]), rawBN254(inputs[5]),
		rawBN254(inputs[6]), rawBN254(inputs[7]), rawBN254(inputs[8]),
	}
}

func rawInputsBLS12377(inputs [9][]blsfr.Element) [][]uint64 {
	return [][]uint64{
		rawBLS12377(inputs[0]), rawBLS12377(inputs[1]), rawBLS12377(inputs[2]),
		rawBLS12377(inputs[3]), rawBLS12377(inputs[4]), rawBLS12377(inputs[5]),
		rawBLS12377(inputs[6]), rawBLS12377(inputs[7]), rawBLS12377(inputs[8]),
	}
}

func rawInputsBW6761(inputs [9][]bwfr.Element) [][]uint64 {
	return [][]uint64{
		rawBW6761(inputs[0]), rawBW6761(inputs[1]), rawBW6761(inputs[2]),
		rawBW6761(inputs[3]), rawBW6761(inputs[4]), rawBW6761(inputs[5]),
		rawBW6761(inputs[6]), rawBW6761(inputs[7]), rawBW6761(inputs[8]),
	}
}

func butterfly4ScalarsBN254(n int) (bnfr.Element, bnfr.Element) {
	domain := bnfft.NewDomain(4 * uint64(n))
	var omega4, omega4Inv, quarter bnfr.Element
	omega4.Exp(domain.Generator, big.NewInt(int64(n)))
	omega4Inv.Inverse(&omega4)
	quarter.SetUint64(4)
	quarter.Inverse(&quarter)
	return omega4Inv, quarter
}

func butterfly4ScalarsBLS12377(n int) (blsfr.Element, blsfr.Element) {
	domain := blsfft.NewDomain(4 * uint64(n))
	var omega4, omega4Inv, quarter blsfr.Element
	omega4.Exp(domain.Generator, big.NewInt(int64(n)))
	omega4Inv.Inverse(&omega4)
	quarter.SetUint64(4)
	quarter.Inverse(&quarter)
	return omega4Inv, quarter
}

func butterfly4ScalarsBW6761(n int) (bwfr.Element, bwfr.Element) {
	domain := bwfft.NewDomain(4 * uint64(n))
	var omega4, omega4Inv, quarter bwfr.Element
	omega4.Exp(domain.Generator, big.NewInt(int64(n)))
	omega4Inv.Inverse(&omega4)
	quarter.SetUint64(4)
	quarter.Inverse(&quarter)
	return omega4Inv, quarter
}
