package vortex

import (
	"testing"

	smt "github.com/consensys/linea-monorepo/prover-ray/crypto/koalabear/smt"
	"github.com/consensys/linea-monorepo/prover-ray/maths/koalabear/field"
	"github.com/consensys/linea-monorepo/prover-ray/maths/koalabear/polynomials"
)

// buildMultilinCoords returns logN random ext-field coordinates.
func buildMultilinCoords(logN int) []field.Ext {
	h := make([]field.Ext, logN)
	for i := range h {
		h[i] = field.RandomElementExt()
	}
	return h
}

func getProofVortexNCommitmentsWithMerkle(_ *testing.T, nCommitments, nbPolys, polySize, rate int, WithSis []bool) (
	*Params,
	*OpeningProof,
	VerifierInput,
	[]Commitment,
	[][]smt.Proof,
) {

	var vi VerifierInput
	vi.X = field.RandomElementExt()
	vi.Alpha = field.RandomElementExt()
	vi.EntryList = []int{1, 5, 19, 645}
	vi.Ys = make([][]field.Ext, nCommitments)

	polyLists := make([][][]field.Element, nCommitments)
	for i := range polyLists {
		polys := make([][]field.Element, nbPolys)
		ys := make([]field.Ext, nbPolys)
		for j := range polys {
			polys[j] = field.VecRandomBase(polySize)
			ys[j] = polynomials.EvalLagrange(field.VecFromBase(polys[j]), field.ElemFromExt(vi.X)).AsExt()
		}
		polyLists[i] = polys
		vi.Ys[i] = ys
	}

	// Commits to it
	commitments := make([]Commitment, nCommitments)
	trees := make([]*smt.Tree, nCommitments)
	logTwoDegree := 9
	logTwoBound := 16
	vortexInstance := NewParams(
		rate,
		polySize,
		nbPolys*nCommitments,
		logTwoDegree,
		logTwoBound)

	encodedMatrices := make([]EncodedMatrix, nCommitments)
	for j := range commitments {
		if WithSis[j] {
			encodedMatrices[j], commitments[j], trees[j], _ = vortexInstance.CommitMerkleWithSIS(polyLists[j])
		} else {
			encodedMatrices[j], commitments[j], trees[j], _ = vortexInstance.CommitMerkleWithoutSIS(polyLists[j])
		}
	}

	// Generate the proof
	proof, merkleProofs := Prove(vi.EntryList, encodedMatrices, trees, vi.Alpha)

	return vortexInstance, proof, vi, commitments, merkleProofs
}

func TestVerifier(t *testing.T) {

	nCommitments := 4
	nbPolys := 15
	polySize := 1 << 10
	rate := 2
	WithSis := make([]bool, nCommitments)

	WithSis[0] = true
	WithSis[1] = true
	WithSis[2] = false
	WithSis[3] = true
	params, proof, vi, commitments, merkleProofs := getProofVortexNCommitmentsWithMerkle(t, nCommitments, nbPolys,
		polySize, rate, WithSis)

	err := Verify(params, proof, &vi, commitments, merkleProofs, WithSis)
	if err != nil {
		t.Fatal(err)
	}
}

func TestVerifierMultilinear(t *testing.T) {
	const nCommitments = 3
	const nbPolys = 8
	const polySize = 1 << 8 // 256 — logN = 8
	const rate = 2
	logN := 8

	withSIS := []bool{true, false, true}

	h := buildMultilinCoords(logN)
	alpha := field.RandomElementExt()
	entryList := []int{3, 17, 42, 100}

	polyLists := make([][][]field.Element, nCommitments)
	vi := VerifierInputMultilinear{
		Alpha:     alpha,
		H:         h,
		EntryList: entryList,
		Ys:        make([][]field.Ext, nCommitments),
	}

	coords := make([]field.Gen, logN)
	for i, hi := range h {
		coords[i] = field.ElemFromExt(hi)
	}

	for i := range nCommitments {
		polys := make([][]field.Element, nbPolys)
		ys := make([]field.Ext, nbPolys)
		for j := range nbPolys {
			polys[j] = field.VecRandomBase(polySize)
			ys[j] = polynomials.EvalMultilin(field.VecFromBase(polys[j]), coords).AsExt()
		}
		polyLists[i] = polys
		vi.Ys[i] = ys
	}

	params := NewParams(rate, polySize, nbPolys*nCommitments, 9, 16)

	commitments := make([]Commitment, nCommitments)
	trees := make([]*smt.Tree, nCommitments)
	encodedMatrices := make([]EncodedMatrix, nCommitments)
	for i := range nCommitments {
		if withSIS[i] {
			encodedMatrices[i], commitments[i], trees[i], _ = params.CommitMerkleWithSIS(polyLists[i])
		} else {
			encodedMatrices[i], commitments[i], trees[i], _ = params.CommitMerkleWithoutSIS(polyLists[i])
		}
	}

	proof, merkleProofs := Prove(vi.EntryList, encodedMatrices, trees, vi.Alpha)

	if err := VerifyMultilinear(params, proof, &vi, commitments, merkleProofs, withSIS); err != nil {
		t.Fatal(err)
	}
}
