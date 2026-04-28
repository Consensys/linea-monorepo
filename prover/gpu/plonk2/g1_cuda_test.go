//go:build cuda

package plonk2

import (
	"math/big"
	"testing"
	"unsafe"

	"github.com/consensys/gnark-crypto/ecc"
	bls12377 "github.com/consensys/gnark-crypto/ecc/bls12-377"
	blsfp "github.com/consensys/gnark-crypto/ecc/bls12-377/fp"
	blsfr "github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254"
	bnfp "github.com/consensys/gnark-crypto/ecc/bn254/fp"
	bnfr "github.com/consensys/gnark-crypto/ecc/bn254/fr"
	bw6761 "github.com/consensys/gnark-crypto/ecc/bw6-761"
	bwfp "github.com/consensys/gnark-crypto/ecc/bw6-761/fp"
	bwfr "github.com/consensys/gnark-crypto/ecc/bw6-761/fr"
	"github.com/stretchr/testify/require"

	"github.com/consensys/linea-monorepo/prover/gpu"
)

func TestG1AffinePrimitives_CUDA(t *testing.T) {
	dev, err := gpu.New()
	require.NoError(t, err, "creating CUDA device should succeed")
	defer dev.Close()

	t.Run("bn254", func(t *testing.T) { testG1BN254(t, dev) })
	t.Run("bls12-377", func(t *testing.T) { testG1BLS12377(t, dev) })
	t.Run("bw6-761", func(t *testing.T) { testG1BW6761(t, dev) })
}

func TestG1MSMNaive_CUDA(t *testing.T) {
	dev, err := gpu.New()
	require.NoError(t, err, "creating CUDA device should succeed")
	defer dev.Close()

	t.Run("bn254", func(t *testing.T) { testG1MSMNaiveBN254(t, dev) })
	t.Run("bls12-377", func(t *testing.T) { testG1MSMNaiveBLS12377(t, dev) })
	t.Run("bw6-761", func(t *testing.T) { testG1MSMNaiveBW6761(t, dev) })
}

func TestG1MSMPippengerRaw_CUDA(t *testing.T) {
	dev, err := gpu.New()
	require.NoError(t, err, "creating CUDA device should succeed")
	defer dev.Close()

	t.Run("bn254", func(t *testing.T) { testG1MSMPippengerBN254(t, dev) })
	t.Run("bls12-377", func(t *testing.T) { testG1MSMPippengerBLS12377(t, dev) })
	t.Run("bw6-761", func(t *testing.T) { testG1MSMPippengerBW6761(t, dev) })
}

func testG1BN254(t *testing.T, dev *gpu.Device) {
	p := bn254Point(5)
	q := bn254Point(9)
	var inf, neg, expected bn254.G1Affine
	inf.SetInfinity()
	neg.Neg(&p)

	expected.Add(&p, &q)
	out, err := g1AffineAddRaw(dev, CurveBN254, rawBN254G1(&p), rawBN254G1(&q))
	require.NoError(t, err, "GPU affine add should succeed")
	requireBN254ProjectiveMatches(t, expected, out, "p+q")

	expected.Double(&p)
	out, err = g1AffineDoubleRaw(dev, CurveBN254, rawBN254G1(&p))
	require.NoError(t, err, "GPU affine double should succeed")
	requireBN254ProjectiveMatches(t, expected, out, "2p")

	expected.Add(&inf, &p)
	out, err = g1AffineAddRaw(dev, CurveBN254, rawBN254G1(&inf), rawBN254G1(&p))
	require.NoError(t, err, "GPU infinity add should succeed")
	requireBN254ProjectiveMatches(t, expected, out, "inf+p")

	expected.Add(&p, &neg)
	out, err = g1AffineAddRaw(dev, CurveBN254, rawBN254G1(&p), rawBN254G1(&neg))
	require.NoError(t, err, "GPU inverse add should succeed")
	requireBN254ProjectiveMatches(t, expected, out, "p+(-p)")
}

func testG1BLS12377(t *testing.T, dev *gpu.Device) {
	p := bls12377Point(5)
	q := bls12377Point(9)
	var inf, neg, expected bls12377.G1Affine
	inf.SetInfinity()
	neg.Neg(&p)

	expected.Add(&p, &q)
	out, err := g1AffineAddRaw(dev, CurveBLS12377, rawBLS12377G1(&p), rawBLS12377G1(&q))
	require.NoError(t, err, "GPU affine add should succeed")
	requireBLS12377ProjectiveMatches(t, expected, out, "p+q")

	expected.Double(&p)
	out, err = g1AffineDoubleRaw(dev, CurveBLS12377, rawBLS12377G1(&p))
	require.NoError(t, err, "GPU affine double should succeed")
	requireBLS12377ProjectiveMatches(t, expected, out, "2p")

	expected.Add(&inf, &p)
	out, err = g1AffineAddRaw(dev, CurveBLS12377, rawBLS12377G1(&inf), rawBLS12377G1(&p))
	require.NoError(t, err, "GPU infinity add should succeed")
	requireBLS12377ProjectiveMatches(t, expected, out, "inf+p")

	expected.Add(&p, &neg)
	out, err = g1AffineAddRaw(dev, CurveBLS12377, rawBLS12377G1(&p), rawBLS12377G1(&neg))
	require.NoError(t, err, "GPU inverse add should succeed")
	requireBLS12377ProjectiveMatches(t, expected, out, "p+(-p)")
}

func testG1BW6761(t *testing.T, dev *gpu.Device) {
	p := bw6761Point(5)
	q := bw6761Point(9)
	var inf, neg, expected bw6761.G1Affine
	inf.SetInfinity()
	neg.Neg(&p)

	expected.Add(&p, &q)
	out, err := g1AffineAddRaw(dev, CurveBW6761, rawBW6761G1(&p), rawBW6761G1(&q))
	require.NoError(t, err, "GPU affine add should succeed")
	requireBW6761ProjectiveMatches(t, expected, out, "p+q")

	expected.Double(&p)
	out, err = g1AffineDoubleRaw(dev, CurveBW6761, rawBW6761G1(&p))
	require.NoError(t, err, "GPU affine double should succeed")
	requireBW6761ProjectiveMatches(t, expected, out, "2p")

	expected.Add(&inf, &p)
	out, err = g1AffineAddRaw(dev, CurveBW6761, rawBW6761G1(&inf), rawBW6761G1(&p))
	require.NoError(t, err, "GPU infinity add should succeed")
	requireBW6761ProjectiveMatches(t, expected, out, "inf+p")

	expected.Add(&p, &neg)
	out, err = g1AffineAddRaw(dev, CurveBW6761, rawBW6761G1(&p), rawBW6761G1(&neg))
	require.NoError(t, err, "GPU inverse add should succeed")
	requireBW6761ProjectiveMatches(t, expected, out, "p+(-p)")
}

func testG1MSMNaiveBN254(t *testing.T, dev *gpu.Device) {
	points := []bn254.G1Affine{bn254Point(2), bn254Point(3), bn254Point(5), bn254Point(7)}
	scalars := []bnfr.Element{bnfr.NewElement(11), bnfr.NewElement(13), bnfr.NewElement(17), bnfr.NewElement(19)}
	var expected bn254.G1Jac
	expected.MultiExp(points, scalars, ecc.MultiExpConfig{})
	expected.ScalarMultiplication(&expected, montgomeryRawMultiplier(bnfr.Modulus(), bnfr.Limbs))
	var expectedAffine bn254.G1Affine
	expectedAffine.FromJacobian(&expected)
	out, err := g1MSMNaiveRaw(dev, CurveBN254, rawBN254G1Slice(points), cloneRaw(rawBN254(scalars)), len(points))
	require.NoError(t, err, "GPU naive MSM should succeed")
	requireBN254ProjectiveMatches(t, expectedAffine, out, "naive MSM")
}

func testG1MSMNaiveBLS12377(t *testing.T, dev *gpu.Device) {
	points := []bls12377.G1Affine{bls12377Point(2), bls12377Point(3), bls12377Point(5), bls12377Point(7)}
	scalars := []blsfr.Element{blsfr.NewElement(11), blsfr.NewElement(13), blsfr.NewElement(17), blsfr.NewElement(19)}
	var expected bls12377.G1Jac
	expected.MultiExp(points, scalars, ecc.MultiExpConfig{})
	expected.ScalarMultiplication(&expected, montgomeryRawMultiplier(blsfr.Modulus(), blsfr.Limbs))
	var expectedAffine bls12377.G1Affine
	expectedAffine.FromJacobian(&expected)
	out, err := g1MSMNaiveRaw(
		dev,
		CurveBLS12377,
		rawBLS12377G1Slice(points),
		cloneRaw(rawBLS12377(scalars)),
		len(points),
	)
	require.NoError(t, err, "GPU naive MSM should succeed")
	requireBLS12377ProjectiveMatches(t, expectedAffine, out, "naive MSM")
}

func testG1MSMNaiveBW6761(t *testing.T, dev *gpu.Device) {
	points := []bw6761.G1Affine{bw6761Point(2), bw6761Point(3), bw6761Point(5), bw6761Point(7)}
	scalars := []bwfr.Element{bwfr.NewElement(11), bwfr.NewElement(13), bwfr.NewElement(17), bwfr.NewElement(19)}
	var expected bw6761.G1Jac
	expected.MultiExp(points, scalars, ecc.MultiExpConfig{})
	expected.ScalarMultiplication(&expected, montgomeryRawMultiplier(bwfr.Modulus(), bwfr.Limbs))
	var expectedAffine bw6761.G1Affine
	expectedAffine.FromJacobian(&expected)
	out, err := g1MSMNaiveRaw(dev, CurveBW6761, rawBW6761G1Slice(points), cloneRaw(rawBW6761(scalars)), len(points))
	require.NoError(t, err, "GPU naive MSM should succeed")
	requireBW6761ProjectiveMatches(t, expectedAffine, out, "naive MSM")
}

func testG1MSMPippengerBN254(t *testing.T, dev *gpu.Device) {
	points := []bn254.G1Affine{
		bn254Point(2), bn254Point(3), bn254Point(5), bn254Point(7),
		bn254Point(11), bn254Point(13), bn254Point(17), bn254Point(19),
		bn254Point(23),
	}
	scalars := []bnfr.Element{
		bnfr.NewElement(11), bnfr.NewElement(13), bnfr.NewElement(17),
		bnfr.NewElement(19), bnfr.NewElement(23), bnfr.NewElement(29),
		bnfr.NewElement(31), bnfr.NewElement(37), bnfr.NewElement(41),
	}
	var expected bn254.G1Jac
	expected.MultiExp(points, scalars, ecc.MultiExpConfig{})
	expected.ScalarMultiplication(&expected, montgomeryRawMultiplier(bnfr.Modulus(), bnfr.Limbs))
	var expectedAffine bn254.G1Affine
	expectedAffine.FromJacobian(&expected)
	out, err := g1MSMPippengerRaw(
		dev,
		CurveBN254,
		rawBN254G1Slice(points),
		cloneRaw(rawBN254(scalars)),
		len(points),
		5,
	)
	require.NoError(t, err, "GPU Pippenger MSM should succeed")
	requireBN254ProjectiveMatches(t, expectedAffine, out, "Pippenger MSM")
}

func testG1MSMPippengerBLS12377(t *testing.T, dev *gpu.Device) {
	points := []bls12377.G1Affine{
		bls12377Point(2), bls12377Point(3), bls12377Point(5),
		bls12377Point(7), bls12377Point(11), bls12377Point(13),
		bls12377Point(17), bls12377Point(19), bls12377Point(23),
	}
	scalars := []blsfr.Element{
		blsfr.NewElement(11), blsfr.NewElement(13), blsfr.NewElement(17),
		blsfr.NewElement(19), blsfr.NewElement(23), blsfr.NewElement(29),
		blsfr.NewElement(31), blsfr.NewElement(37), blsfr.NewElement(41),
	}
	var expected bls12377.G1Jac
	expected.MultiExp(points, scalars, ecc.MultiExpConfig{})
	expected.ScalarMultiplication(&expected, montgomeryRawMultiplier(blsfr.Modulus(), blsfr.Limbs))
	var expectedAffine bls12377.G1Affine
	expectedAffine.FromJacobian(&expected)
	out, err := g1MSMPippengerRaw(
		dev,
		CurveBLS12377,
		rawBLS12377G1Slice(points),
		cloneRaw(rawBLS12377(scalars)),
		len(points),
		5,
	)
	require.NoError(t, err, "GPU Pippenger MSM should succeed")
	requireBLS12377ProjectiveMatches(t, expectedAffine, out, "Pippenger MSM")
}

func testG1MSMPippengerBW6761(t *testing.T, dev *gpu.Device) {
	points := []bw6761.G1Affine{
		bw6761Point(2), bw6761Point(3), bw6761Point(5), bw6761Point(7),
		bw6761Point(11), bw6761Point(13), bw6761Point(17), bw6761Point(19),
		bw6761Point(23),
	}
	scalars := []bwfr.Element{
		bwfr.NewElement(11), bwfr.NewElement(13), bwfr.NewElement(17),
		bwfr.NewElement(19), bwfr.NewElement(23), bwfr.NewElement(29),
		bwfr.NewElement(31), bwfr.NewElement(37), bwfr.NewElement(41),
	}
	var expected bw6761.G1Jac
	expected.MultiExp(points, scalars, ecc.MultiExpConfig{})
	expected.ScalarMultiplication(&expected, montgomeryRawMultiplier(bwfr.Modulus(), bwfr.Limbs))
	var expectedAffine bw6761.G1Affine
	expectedAffine.FromJacobian(&expected)
	out, err := g1MSMPippengerRaw(
		dev,
		CurveBW6761,
		rawBW6761G1Slice(points),
		cloneRaw(rawBW6761(scalars)),
		len(points),
		5,
	)
	require.NoError(t, err, "GPU Pippenger MSM should succeed")
	requireBW6761ProjectiveMatches(t, expectedAffine, out, "Pippenger MSM")
}

func bn254Point(s int64) bn254.G1Affine {
	var p bn254.G1Affine
	p.ScalarMultiplicationBase(big.NewInt(s))
	return p
}

func bls12377Point(s int64) bls12377.G1Affine {
	var p bls12377.G1Affine
	p.ScalarMultiplicationBase(big.NewInt(s))
	return p
}

func bw6761Point(s int64) bw6761.G1Affine {
	var p bw6761.G1Affine
	p.ScalarMultiplicationBase(big.NewInt(s))
	return p
}

func rawBN254G1(p *bn254.G1Affine) []uint64 {
	return cloneRaw(unsafe.Slice((*uint64)(unsafe.Pointer(p)), 2*bnfp.Limbs))
}

func rawBLS12377G1(p *bls12377.G1Affine) []uint64 {
	return cloneRaw(unsafe.Slice((*uint64)(unsafe.Pointer(p)), 2*blsfp.Limbs))
}

func rawBW6761G1(p *bw6761.G1Affine) []uint64 {
	return cloneRaw(unsafe.Slice((*uint64)(unsafe.Pointer(p)), 2*bwfp.Limbs))
}

func rawBN254G1Slice(points []bn254.G1Affine) []uint64 {
	out := make([]uint64, 0, len(points)*2*bnfp.Limbs)
	for i := range points {
		out = append(out, rawBN254G1(&points[i])...)
	}
	return out
}

func rawBLS12377G1Slice(points []bls12377.G1Affine) []uint64 {
	out := make([]uint64, 0, len(points)*2*blsfp.Limbs)
	for i := range points {
		out = append(out, rawBLS12377G1(&points[i])...)
	}
	return out
}

func rawBW6761G1Slice(points []bw6761.G1Affine) []uint64 {
	out := make([]uint64, 0, len(points)*2*bwfp.Limbs)
	for i := range points {
		out = append(out, rawBW6761G1(&points[i])...)
	}
	return out
}

func montgomeryRawMultiplier(modulus *big.Int, limbs int) *big.Int {
	r := new(big.Int).Lsh(big.NewInt(1), uint(64*limbs))
	return r.Mod(r, modulus)
}

func requireBN254ProjectiveMatches(tb testing.TB, want bn254.G1Affine, raw []uint64, label string) {
	tb.Helper()
	var got [3]bnfp.Element
	copy(unsafe.Slice((*uint64)(unsafe.Pointer(&got[0])), 3*bnfp.Limbs), raw)
	if want.IsInfinity() {
		require.True(tb, got[2].IsZero(), "%s should be projective infinity", label)
		return
	}
	require.False(tb, got[2].IsZero(), "%s should not be projective infinity", label)
	var z2, z3, wantX, wantY bnfp.Element
	z2.Square(&got[2])
	z3.Mul(&z2, &got[2])
	wantX.Mul(&want.X, &z2)
	wantY.Mul(&want.Y, &z3)
	require.True(tb, wantX.Equal(&got[0]), "%s projective X should match", label)
	require.True(tb, wantY.Equal(&got[1]), "%s projective Y should match", label)
}

func requireBLS12377ProjectiveMatches(tb testing.TB, want bls12377.G1Affine, raw []uint64, label string) {
	tb.Helper()
	var got [3]blsfp.Element
	copy(unsafe.Slice((*uint64)(unsafe.Pointer(&got[0])), 3*blsfp.Limbs), raw)
	if want.IsInfinity() {
		require.True(tb, got[2].IsZero(), "%s should be projective infinity", label)
		return
	}
	require.False(tb, got[2].IsZero(), "%s should not be projective infinity", label)
	var z2, z3, wantX, wantY blsfp.Element
	z2.Square(&got[2])
	z3.Mul(&z2, &got[2])
	wantX.Mul(&want.X, &z2)
	wantY.Mul(&want.Y, &z3)
	require.True(tb, wantX.Equal(&got[0]), "%s projective X should match", label)
	require.True(tb, wantY.Equal(&got[1]), "%s projective Y should match", label)
}

func requireBW6761ProjectiveMatches(tb testing.TB, want bw6761.G1Affine, raw []uint64, label string) {
	tb.Helper()
	var got [3]bwfp.Element
	copy(unsafe.Slice((*uint64)(unsafe.Pointer(&got[0])), 3*bwfp.Limbs), raw)
	if want.IsInfinity() {
		require.True(tb, got[2].IsZero(), "%s should be projective infinity", label)
		return
	}
	require.False(tb, got[2].IsZero(), "%s should not be projective infinity", label)
	var z2, z3, wantX, wantY bwfp.Element
	z2.Square(&got[2])
	z3.Mul(&z2, &got[2])
	wantX.Mul(&want.X, &z2)
	wantY.Mul(&want.Y, &z3)
	require.True(tb, wantX.Equal(&got[0]), "%s projective X should match", label)
	require.True(tb, wantY.Equal(&got[1]), "%s projective Y should match", label)
}
