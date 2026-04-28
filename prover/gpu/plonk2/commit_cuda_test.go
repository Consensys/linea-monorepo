//go:build cuda

package plonk2

import (
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	bls12377 "github.com/consensys/gnark-crypto/ecc/bls12-377"
	blsfr "github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	blskzg "github.com/consensys/gnark-crypto/ecc/bls12-377/kzg"
	"github.com/consensys/gnark-crypto/ecc/bn254"
	bnfr "github.com/consensys/gnark-crypto/ecc/bn254/fr"
	bnkzg "github.com/consensys/gnark-crypto/ecc/bn254/kzg"
	bw6761 "github.com/consensys/gnark-crypto/ecc/bw6-761"
	bwfr "github.com/consensys/gnark-crypto/ecc/bw6-761/fr"
	bwkzg "github.com/consensys/gnark-crypto/ecc/bw6-761/kzg"
	"github.com/stretchr/testify/require"

	"github.com/consensys/linea-monorepo/prover/gpu"
)

func TestCommitRaw_CUDA(t *testing.T) {
	dev, err := gpu.New()
	require.NoError(t, err, "creating CUDA device should succeed")
	defer dev.Close()

	t.Run("bn254", func(t *testing.T) { testCommitRawBN254(t, dev) })
	t.Run("bls12-377", func(t *testing.T) { testCommitRawBLS12377(t, dev) })
	t.Run("bw6-761", func(t *testing.T) { testCommitRawBW6761(t, dev) })
}

func TestCommitRawMatchesKZGCommit_CUDA(t *testing.T) {
	dev, err := gpu.New()
	require.NoError(t, err, "creating CUDA device should succeed")
	defer dev.Close()

	t.Run("bn254", func(t *testing.T) { testCommitRawMatchesKZGBN254(t, dev) })
	t.Run("bls12-377", func(t *testing.T) { testCommitRawMatchesKZGBLS12377(t, dev) })
	t.Run("bw6-761", func(t *testing.T) { testCommitRawMatchesKZGBW6761(t, dev) })
}

func TestCommitRawMatchesKZGOpenQuotient_CUDA(t *testing.T) {
	dev, err := gpu.New()
	require.NoError(t, err, "creating CUDA device should succeed")
	defer dev.Close()

	t.Run("bn254", func(t *testing.T) { testCommitRawMatchesKZGOpenBN254(t, dev) })
	t.Run("bls12-377", func(t *testing.T) { testCommitRawMatchesKZGOpenBLS12377(t, dev) })
	t.Run("bw6-761", func(t *testing.T) { testCommitRawMatchesKZGOpenBW6761(t, dev) })
}

func TestG1MSMCommitRaw_CUDA(t *testing.T) {
	dev, err := gpu.New()
	require.NoError(t, err, "creating CUDA device should succeed")
	defer dev.Close()

	t.Run("bn254", func(t *testing.T) { testG1MSMCommitRawBN254(t, dev) })
	t.Run("bls12-377", func(t *testing.T) { testG1MSMCommitRawBLS12377(t, dev) })
	t.Run("bw6-761", func(t *testing.T) { testG1MSMCommitRawBW6761(t, dev) })
}

func testCommitRawBN254(t *testing.T, dev *gpu.Device) {
	points := []bn254.G1Affine{bn254Point(2), bn254Point(3), bn254Point(5), bn254Point(7)}
	scalars := []bnfr.Element{
		bnfr.NewElement(23),
		bnfr.NewElement(29),
		bnfr.NewElement(31),
		bnfr.NewElement(37),
	}
	var expected bn254.G1Jac
	expected.MultiExp(points, scalars, ecc.MultiExpConfig{})
	var expectedAffine bn254.G1Affine
	expectedAffine.FromJacobian(&expected)
	out, err := CommitRaw(dev, CurveBN254, rawBN254G1Slice(points), cloneRaw(rawBN254(scalars)))
	require.NoError(t, err, "raw commitment should succeed")
	requireBN254ProjectiveMatches(t, expectedAffine, out, "raw commitment")
}

func testCommitRawBLS12377(t *testing.T, dev *gpu.Device) {
	points := []bls12377.G1Affine{bls12377Point(2), bls12377Point(3), bls12377Point(5), bls12377Point(7)}
	scalars := []blsfr.Element{
		blsfr.NewElement(23),
		blsfr.NewElement(29),
		blsfr.NewElement(31),
		blsfr.NewElement(37),
	}
	var expected bls12377.G1Jac
	expected.MultiExp(points, scalars, ecc.MultiExpConfig{})
	var expectedAffine bls12377.G1Affine
	expectedAffine.FromJacobian(&expected)
	out, err := CommitRaw(dev, CurveBLS12377, rawBLS12377G1Slice(points), cloneRaw(rawBLS12377(scalars)))
	require.NoError(t, err, "raw commitment should succeed")
	requireBLS12377ProjectiveMatches(t, expectedAffine, out, "raw commitment")
}

func testCommitRawBW6761(t *testing.T, dev *gpu.Device) {
	points := []bw6761.G1Affine{bw6761Point(2), bw6761Point(3), bw6761Point(5), bw6761Point(7)}
	scalars := []bwfr.Element{
		bwfr.NewElement(23),
		bwfr.NewElement(29),
		bwfr.NewElement(31),
		bwfr.NewElement(37),
	}
	var expected bw6761.G1Jac
	expected.MultiExp(points, scalars, ecc.MultiExpConfig{})
	var expectedAffine bw6761.G1Affine
	expectedAffine.FromJacobian(&expected)
	out, err := CommitRaw(dev, CurveBW6761, rawBW6761G1Slice(points), cloneRaw(rawBW6761(scalars)))
	require.NoError(t, err, "raw commitment should succeed")
	requireBW6761ProjectiveMatches(t, expectedAffine, out, "raw commitment")
}

func testCommitRawMatchesKZGBN254(t *testing.T, dev *gpu.Device) {
	poly := []bnfr.Element{
		bnfr.NewElement(2), bnfr.NewElement(3), bnfr.NewElement(5),
		bnfr.NewElement(7), bnfr.NewElement(11), bnfr.NewElement(13),
		bnfr.NewElement(17), bnfr.NewElement(19),
	}
	srs, err := bnkzg.NewSRS(uint64(len(poly)), big.NewInt(7))
	require.NoError(t, err, "creating BN254 SRS should succeed")
	expected, err := bnkzg.Commit(poly, srs.Pk)
	require.NoError(t, err, "BN254 KZG commit should succeed")
	out, err := CommitRaw(dev, CurveBN254, rawBN254G1Slice(srs.Pk.G1), cloneRaw(rawBN254(poly)))
	require.NoError(t, err, "GPU KZG commitment should succeed")
	requireBN254ProjectiveMatches(t, expected, out, "KZG commitment")
}

func testCommitRawMatchesKZGBLS12377(t *testing.T, dev *gpu.Device) {
	poly := []blsfr.Element{
		blsfr.NewElement(2), blsfr.NewElement(3), blsfr.NewElement(5),
		blsfr.NewElement(7), blsfr.NewElement(11), blsfr.NewElement(13),
		blsfr.NewElement(17), blsfr.NewElement(19),
	}
	srs, err := blskzg.NewSRS(uint64(len(poly)), big.NewInt(7))
	require.NoError(t, err, "creating BLS12-377 SRS should succeed")
	expected, err := blskzg.Commit(poly, srs.Pk)
	require.NoError(t, err, "BLS12-377 KZG commit should succeed")
	out, err := CommitRaw(dev, CurveBLS12377, rawBLS12377G1Slice(srs.Pk.G1), cloneRaw(rawBLS12377(poly)))
	require.NoError(t, err, "GPU KZG commitment should succeed")
	requireBLS12377ProjectiveMatches(t, expected, out, "KZG commitment")
}

func testCommitRawMatchesKZGBW6761(t *testing.T, dev *gpu.Device) {
	poly := []bwfr.Element{
		bwfr.NewElement(2), bwfr.NewElement(3), bwfr.NewElement(5),
		bwfr.NewElement(7), bwfr.NewElement(11), bwfr.NewElement(13),
		bwfr.NewElement(17), bwfr.NewElement(19),
	}
	srs, err := bwkzg.NewSRS(uint64(len(poly)), big.NewInt(7))
	require.NoError(t, err, "creating BW6-761 SRS should succeed")
	expected, err := bwkzg.Commit(poly, srs.Pk)
	require.NoError(t, err, "BW6-761 KZG commit should succeed")
	out, err := CommitRaw(dev, CurveBW6761, rawBW6761G1Slice(srs.Pk.G1), cloneRaw(rawBW6761(poly)))
	require.NoError(t, err, "GPU KZG commitment should succeed")
	requireBW6761ProjectiveMatches(t, expected, out, "KZG commitment")
}

func testCommitRawMatchesKZGOpenBN254(t *testing.T, dev *gpu.Device) {
	poly := []bnfr.Element{
		bnfr.NewElement(3), bnfr.NewElement(4), bnfr.NewElement(8),
		bnfr.NewElement(15), bnfr.NewElement(16), bnfr.NewElement(23),
		bnfr.NewElement(42), bnfr.NewElement(108),
	}
	point := bnfr.NewElement(9)
	srs, err := bnkzg.NewSRS(uint64(len(poly)), big.NewInt(7))
	require.NoError(t, err, "creating BN254 SRS should succeed")
	proof, err := bnkzg.Open(poly, point, srs.Pk)
	require.NoError(t, err, "BN254 KZG open should succeed")
	h := quotientBN254(poly, proof.ClaimedValue, point)
	out, err := CommitRaw(dev, CurveBN254, rawBN254G1Slice(srs.Pk.G1[:len(h)]), cloneRaw(rawBN254(h)))
	require.NoError(t, err, "GPU quotient commitment should succeed")
	requireBN254ProjectiveMatches(t, proof.H, out, "KZG opening quotient")
}

func testCommitRawMatchesKZGOpenBLS12377(t *testing.T, dev *gpu.Device) {
	poly := []blsfr.Element{
		blsfr.NewElement(3), blsfr.NewElement(4), blsfr.NewElement(8),
		blsfr.NewElement(15), blsfr.NewElement(16), blsfr.NewElement(23),
		blsfr.NewElement(42), blsfr.NewElement(108),
	}
	point := blsfr.NewElement(9)
	srs, err := blskzg.NewSRS(uint64(len(poly)), big.NewInt(7))
	require.NoError(t, err, "creating BLS12-377 SRS should succeed")
	proof, err := blskzg.Open(poly, point, srs.Pk)
	require.NoError(t, err, "BLS12-377 KZG open should succeed")
	h := quotientBLS12377(poly, proof.ClaimedValue, point)
	out, err := CommitRaw(dev, CurveBLS12377, rawBLS12377G1Slice(srs.Pk.G1[:len(h)]), cloneRaw(rawBLS12377(h)))
	require.NoError(t, err, "GPU quotient commitment should succeed")
	requireBLS12377ProjectiveMatches(t, proof.H, out, "KZG opening quotient")
}

func testCommitRawMatchesKZGOpenBW6761(t *testing.T, dev *gpu.Device) {
	poly := []bwfr.Element{
		bwfr.NewElement(3), bwfr.NewElement(4), bwfr.NewElement(8),
		bwfr.NewElement(15), bwfr.NewElement(16), bwfr.NewElement(23),
		bwfr.NewElement(42), bwfr.NewElement(108),
	}
	point := bwfr.NewElement(9)
	srs, err := bwkzg.NewSRS(uint64(len(poly)), big.NewInt(7))
	require.NoError(t, err, "creating BW6-761 SRS should succeed")
	proof, err := bwkzg.Open(poly, point, srs.Pk)
	require.NoError(t, err, "BW6-761 KZG open should succeed")
	h := quotientBW6761(poly, proof.ClaimedValue, point)
	out, err := CommitRaw(dev, CurveBW6761, rawBW6761G1Slice(srs.Pk.G1[:len(h)]), cloneRaw(rawBW6761(h)))
	require.NoError(t, err, "GPU quotient commitment should succeed")
	requireBW6761ProjectiveMatches(t, proof.H, out, "KZG opening quotient")
}

func testG1MSMCommitRawBN254(t *testing.T, dev *gpu.Device) {
	srs, err := bnkzg.NewSRS(8, big.NewInt(7))
	require.NoError(t, err, "creating BN254 SRS should succeed")
	msm, err := NewG1MSM(dev, CurveBN254, rawBN254G1Slice(srs.Pk.G1))
	require.NoError(t, err, "creating resident BN254 MSM should succeed")
	defer msm.Close()

	poly := []bnfr.Element{
		bnfr.NewElement(2), bnfr.NewElement(3), bnfr.NewElement(5),
		bnfr.NewElement(7), bnfr.NewElement(11), bnfr.NewElement(13),
		bnfr.NewElement(17), bnfr.NewElement(19),
	}
	expected, err := bnkzg.Commit(poly, srs.Pk)
	require.NoError(t, err, "BN254 KZG commit should succeed")
	out, err := msm.CommitRaw(cloneRaw(rawBN254(poly)))
	require.NoError(t, err, "resident BN254 commitment should succeed")
	requireBN254ProjectiveMatches(t, expected, out, "resident KZG commitment")

	require.NoError(t, msm.ReleaseWorkBuffers(), "releasing BN254 MSM work buffers should succeed")
	shortPoly := poly[:4]
	expected, err = bnkzg.Commit(shortPoly, srs.Pk)
	require.NoError(t, err, "BN254 short KZG commit should succeed")
	out, err = msm.CommitRaw(cloneRaw(rawBN254(shortPoly)))
	require.NoError(t, err, "resident BN254 short commitment should succeed")
	requireBN254ProjectiveMatches(t, expected, out, "resident short KZG commitment")

	require.NoError(t, msm.PinWorkBuffers(), "pinning BN254 MSM work buffers should succeed")
	expected, err = bnkzg.Commit(poly, srs.Pk)
	require.NoError(t, err, "BN254 repeated KZG commit should succeed")
	out, err = msm.CommitRaw(cloneRaw(rawBN254(poly)))
	require.NoError(t, err, "resident BN254 repeated commitment should succeed")
	requireBN254ProjectiveMatches(t, expected, out, "resident repeated KZG commitment")
}

func testG1MSMCommitRawBLS12377(t *testing.T, dev *gpu.Device) {
	srs, err := blskzg.NewSRS(8, big.NewInt(7))
	require.NoError(t, err, "creating BLS12-377 SRS should succeed")
	msm, err := NewG1MSM(dev, CurveBLS12377, rawBLS12377G1Slice(srs.Pk.G1))
	require.NoError(t, err, "creating resident BLS12-377 MSM should succeed")
	defer msm.Close()

	poly := []blsfr.Element{
		blsfr.NewElement(2), blsfr.NewElement(3), blsfr.NewElement(5),
		blsfr.NewElement(7), blsfr.NewElement(11), blsfr.NewElement(13),
		blsfr.NewElement(17), blsfr.NewElement(19),
	}
	expected, err := blskzg.Commit(poly, srs.Pk)
	require.NoError(t, err, "BLS12-377 KZG commit should succeed")
	out, err := msm.CommitRaw(cloneRaw(rawBLS12377(poly)))
	require.NoError(t, err, "resident BLS12-377 commitment should succeed")
	requireBLS12377ProjectiveMatches(t, expected, out, "resident KZG commitment")

	require.NoError(t, msm.ReleaseWorkBuffers(), "releasing BLS12-377 MSM work buffers should succeed")
	shortPoly := poly[:4]
	expected, err = blskzg.Commit(shortPoly, srs.Pk)
	require.NoError(t, err, "BLS12-377 short KZG commit should succeed")
	out, err = msm.CommitRaw(cloneRaw(rawBLS12377(shortPoly)))
	require.NoError(t, err, "resident BLS12-377 short commitment should succeed")
	requireBLS12377ProjectiveMatches(t, expected, out, "resident short KZG commitment")
}

func testG1MSMCommitRawBW6761(t *testing.T, dev *gpu.Device) {
	srs, err := bwkzg.NewSRS(8, big.NewInt(7))
	require.NoError(t, err, "creating BW6-761 SRS should succeed")
	msm, err := NewG1MSM(dev, CurveBW6761, rawBW6761G1Slice(srs.Pk.G1))
	require.NoError(t, err, "creating resident BW6-761 MSM should succeed")
	defer msm.Close()

	poly := []bwfr.Element{
		bwfr.NewElement(2), bwfr.NewElement(3), bwfr.NewElement(5),
		bwfr.NewElement(7), bwfr.NewElement(11), bwfr.NewElement(13),
		bwfr.NewElement(17), bwfr.NewElement(19),
	}
	expected, err := bwkzg.Commit(poly, srs.Pk)
	require.NoError(t, err, "BW6-761 KZG commit should succeed")
	out, err := msm.CommitRaw(cloneRaw(rawBW6761(poly)))
	require.NoError(t, err, "resident BW6-761 commitment should succeed")
	requireBW6761ProjectiveMatches(t, expected, out, "resident KZG commitment")

	require.NoError(t, msm.ReleaseWorkBuffers(), "releasing BW6-761 MSM work buffers should succeed")
	shortPoly := poly[:4]
	expected, err = bwkzg.Commit(shortPoly, srs.Pk)
	require.NoError(t, err, "BW6-761 short KZG commit should succeed")
	out, err = msm.CommitRaw(cloneRaw(rawBW6761(shortPoly)))
	require.NoError(t, err, "resident BW6-761 short commitment should succeed")
	requireBW6761ProjectiveMatches(t, expected, out, "resident short KZG commitment")
}

func TestCommitRawRejectsMalformedInputs_CUDA(t *testing.T) {
	dev, err := gpu.New()
	require.NoError(t, err, "creating CUDA device should succeed")
	defer dev.Close()

	_, err = CommitRaw(dev, CurveBN254, nil, nil)
	require.Error(t, err, "empty point buffer should fail")

	pointAffine := bn254Point(2)
	point := rawBN254G1(&pointAffine)
	scalar := cloneRaw(rawBN254([]bnfr.Element{bnfr.NewElement(3)}))

	_, err = CommitRaw(dev, CurveBN254, point[:len(point)-1], scalar)
	require.Error(t, err, "truncated point buffer should fail")

	_, err = CommitRaw(dev, CurveBN254, point, scalar[:len(scalar)-1])
	require.Error(t, err, "truncated scalar buffer should fail")

	_, err = CommitRaw(dev, Curve(99), point, scalar)
	require.Error(t, err, "unsupported curve should fail")
}

func TestCommitRawMatchesScalarMultipleForOnePoint_CUDA(t *testing.T) {
	dev, err := gpu.New()
	require.NoError(t, err, "creating CUDA device should succeed")
	defer dev.Close()

	point := bn254Point(17)
	scalar := bnfr.NewElement(41)
	out, err := CommitRaw(dev, CurveBN254, rawBN254G1(&point), cloneRaw(rawBN254([]bnfr.Element{scalar})))
	require.NoError(t, err, "single-point raw commitment should succeed")

	var expected bn254.G1Affine
	scalarRegular := scalar.ToBigIntRegular(new(big.Int))
	expected.ScalarMultiplication(&point, scalarRegular)
	requireBN254ProjectiveMatches(t, expected, out, "single-point raw commitment")
}

func quotientBN254(poly []bnfr.Element, claimed, point bnfr.Element) []bnfr.Element {
	f := append([]bnfr.Element(nil), poly...)
	f[0].Sub(&f[0], &claimed)
	for i := len(f) - 2; i >= 0; i-- {
		var t bnfr.Element
		t.Mul(&f[i+1], &point)
		f[i].Add(&f[i], &t)
	}
	return f[1:]
}

func quotientBLS12377(poly []blsfr.Element, claimed, point blsfr.Element) []blsfr.Element {
	f := append([]blsfr.Element(nil), poly...)
	f[0].Sub(&f[0], &claimed)
	for i := len(f) - 2; i >= 0; i-- {
		var t blsfr.Element
		t.Mul(&f[i+1], &point)
		f[i].Add(&f[i], &t)
	}
	return f[1:]
}

func quotientBW6761(poly []bwfr.Element, claimed, point bwfr.Element) []bwfr.Element {
	f := append([]bwfr.Element(nil), poly...)
	f[0].Sub(&f[0], &claimed)
	for i := len(f) - 2; i >= 0; i-- {
		var t bwfr.Element
		t.Mul(&f[i+1], &point)
		f[i].Add(&f[i], &t)
	}
	return f[1:]
}
