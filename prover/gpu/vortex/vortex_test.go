package vortex

import (
	"math/rand/v2"
	"testing"

	"github.com/consensys/gnark-crypto/field/koalabear"
	fext "github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/consensys/gnark-crypto/field/koalabear/sis"
	refvortex "github.com/consensys/gnark-crypto/field/koalabear/vortex"
	"github.com/stretchr/testify/require"
)

// ─── Helpers ────────────────────────────────────────────────────────────────

func randKB(rng *rand.Rand) koalabear.Element {
	return koalabear.Element{rng.Uint32N(2130706433)}
}

func randE4(rng *rand.Rand) fext.E4 {
	return fext.E4{
		B0: fext.E2{A0: randKB(rng), A1: randKB(rng)},
		B1: fext.E2{A0: randKB(rng), A1: randKB(rng)},
	}
}

func randMatrix(rng *rand.Rand, nRows, nCols int) [][]koalabear.Element {
	m := make([][]koalabear.Element, nRows)
	for i := range m {
		m[i] = make([]koalabear.Element, nCols)
		for j := range m[i] {
			m[i][j] = randKB(rng)
		}
	}
	return m
}

func deterministicRNG(seed byte) *rand.Rand {
	var s [32]byte
	s[0] = seed
	return rand.New(rand.NewChaCha8(s)) //nolint:gosec // deterministic test RNG
}

// ─── Roundtrip: Commit → Prove → Verify ─────────────────────────────────────

func TestRoundtrip(t *testing.T) {
	assert := require.New(t)
	rng := deterministicRNG(0)

	nCols := 16
	nRows := 8
	rate := 2
	nSelected := 4

	sisParams, err := sis.NewRSis(0, 9, 16, nRows)
	assert.NoError(err)

	params, err := NewParams(nCols, nRows, sisParams, rate, nSelected)
	assert.NoError(err)

	// Random matrix
	m := randMatrix(rng, nRows, nCols)

	// Compute claimed values: Ys[i] = eval(row[i], x) in Lagrange form
	x := randE4(rng)
	ys := make([]fext.E4, nRows)
	for i := range m {
		ys[i], err = EvalBasePolyLagrange(m[i], x)
		assert.NoError(err)
	}

	alpha := randE4(rng)
	selectedCols := []int{0, 1, 2, 3}

	// Commit
	cs, root, err := params.Commit(m)
	assert.NoError(err)

	// Prove
	proof, err := cs.Prove(alpha, selectedCols)
	assert.NoError(err)

	// Verify
	err = params.Verify(root, proof, ys, x, alpha, selectedCols)
	assert.NoError(err, "roundtrip verification failed")
}

// ─── Zero matrix ─────────────────────────────────────────────────────────────

func TestZeroMatrix(t *testing.T) {
	assert := require.New(t)

	nCols := 16
	nRows := 8
	rate := 2

	sisParams, err := sis.NewRSis(0, 9, 16, nRows)
	assert.NoError(err)

	params, err := NewParams(nCols, nRows, sisParams, rate, 4)
	assert.NoError(err)

	// All-zero matrix
	m := make([][]koalabear.Element, nRows)
	for i := range m {
		m[i] = make([]koalabear.Element, nCols)
	}

	x := fext.E4{}
	ys := make([]fext.E4, nRows)
	alpha, _ := new(fext.E4).SetRandom()
	selectedCols := []int{0, 1, 2, 3}

	cs, root, err := params.Commit(m)
	assert.NoError(err)

	proof, err := cs.Prove(*alpha, selectedCols)
	assert.NoError(err)

	err = params.Verify(root, proof, ys, x, *alpha, selectedCols)
	assert.NoError(err, "zero matrix verification failed")
}

// ─── Cross-validation: our API vs gnark-crypto reference ─────────────────────

func TestCrossValidationCommitment(t *testing.T) {
	assert := require.New(t)
	rng := deterministicRNG(42)

	nCols := 32
	nRows := 16
	rate := 2
	nSelected := 8

	sisParams, err := sis.NewRSis(0, 9, 16, nRows)
	assert.NoError(err)

	// Our params
	ourParams, err := NewParams(nCols, nRows, sisParams, rate, nSelected)
	assert.NoError(err)

	// Reference params (gnark-crypto)
	refParams, err := refvortex.NewParams(nCols, nRows, sisParams, rate, nSelected)
	assert.NoError(err)

	// Same matrix
	m := randMatrix(rng, nRows, nCols)

	// Commit via our API
	_, ourRoot, err := ourParams.Commit(m)
	assert.NoError(err)

	// Commit via reference
	refState, err := refvortex.Commit(refParams, m)
	assert.NoError(err)
	refRoot := refState.GetCommitment()

	// Roots must match
	assert.Equal(refRoot, ourRoot, "commitment roots differ from gnark-crypto reference")
}

func TestCrossValidationFullProtocol(t *testing.T) {
	assert := require.New(t)
	rng := deterministicRNG(99)

	nCols := 32
	nRows := 16
	rate := 2
	nSelected := 4

	sisParams, err := sis.NewRSis(0, 9, 16, nRows)
	assert.NoError(err)

	ourParams, err := NewParams(nCols, nRows, sisParams, rate, nSelected)
	assert.NoError(err)
	refParams, err := refvortex.NewParams(nCols, nRows, sisParams, rate, nSelected)
	assert.NoError(err)

	m := randMatrix(rng, nRows, nCols)

	// Shared randomness
	x := randE4(rng)
	alpha := randE4(rng)
	selectedCols := make([]int, nSelected)
	for i := range selectedCols {
		selectedCols[i] = rng.IntN(nCols*rate - 1)
	}

	// Compute claimed values
	ys := make([]fext.E4, nRows)
	for i := range m {
		ys[i], err = refvortex.EvalBasePolyLagrange(m[i], x)
		assert.NoError(err)
	}

	// ── Our side ──
	ourCS, ourRoot, err := ourParams.Commit(m)
	assert.NoError(err)

	ourProof, err := ourCS.Prove(alpha, selectedCols)
	assert.NoError(err)

	err = ourParams.Verify(ourRoot, ourProof, ys, x, alpha, selectedCols)
	assert.NoError(err, "our verify failed")

	// ── Reference side ──
	refState, err := refvortex.Commit(refParams, m)
	assert.NoError(err)

	refState.OpenLinComb(alpha)
	refProof, err := refState.OpenColumns(selectedCols)
	assert.NoError(err)

	err = refParams.Verify(refvortex.VerifierInput{
		Proof:           refProof,
		MerkleRoot:      refState.GetCommitment(),
		ClaimedValues:   ys,
		EvaluationPoint: x,
		Alpha:           alpha,
		SelectedColumns: selectedCols,
	})
	assert.NoError(err, "reference verify failed")

	// ── Cross-check ──
	assert.Equal(refState.GetCommitment(), ourRoot, "roots must match")
	assert.Equal(len(refProof.UAlpha), len(ourProof.UAlpha), "UAlpha length mismatch")
	for i := range refProof.UAlpha {
		assert.Equal(refProof.UAlpha[i], ourProof.UAlpha[i], "UAlpha[%d] mismatch", i)
	}
}

// ─── Component tests ─────────────────────────────────────────────────────────

func TestReedSolomonEncode(t *testing.T) {
	assert := require.New(t)
	rng := deterministicRNG(7)

	nCols := 64
	rate := 2

	sisParams, err := sis.NewRSis(0, 9, 16, 8)
	assert.NoError(err)

	params, err := NewParams(nCols, 8, sisParams, rate, 4)
	assert.NoError(err)

	// Random row
	row := make([]koalabear.Element, nCols)
	for j := range row {
		row[j] = randKB(rng)
	}

	// Encode via our API
	ourEncoded := make([]koalabear.Element, nCols*rate)
	params.EncodeReedSolomon(row, ourEncoded)

	// Encode via reference
	refParams, err := refvortex.NewParams(nCols, 8, sisParams, rate, 4)
	assert.NoError(err)
	refEncoded := make([]koalabear.Element, nCols*rate)
	refParams.EncodeReedSolomon(row, refEncoded)

	assert.Equal(refEncoded, ourEncoded, "RS encoding mismatch")
}

func TestPoseidon2Compress(t *testing.T) {
	assert := require.New(t)
	rng := deterministicRNG(13)

	var a, b Hash
	for j := 0; j < 8; j++ {
		a[j] = randKB(rng)
		b[j] = randKB(rng)
	}

	// Our API
	ourHash := CompressPoseidon2(a, b)
	// Reference
	refHash := refvortex.CompressPoseidon2(a, b)

	assert.Equal(refHash, ourHash, "Poseidon2 compress mismatch")
}

func TestPoseidon2Sponge(t *testing.T) {
	assert := require.New(t)
	rng := deterministicRNG(17)

	input := make([]koalabear.Element, 48)
	for j := range input {
		input[j] = randKB(rng)
	}

	ourHash := HashPoseidon2(input)
	refHash := refvortex.HashPoseidon2(input)

	assert.Equal(refHash, ourHash, "Poseidon2 sponge mismatch")
}

func TestPoseidon2CompressX16(t *testing.T) {
	assert := require.New(t)
	rng := deterministicRNG(19)

	const (
		width   = 16
		colSize = 16
	)
	input := make([]koalabear.Element, width*colSize)
	for i := range input {
		input[i] = randKB(rng)
	}

	our := make([]Hash, width)
	ref := make([]Hash, width)
	CompressPoseidon2x16(input, colSize, our)
	refvortex.CompressPoseidon2x16(input, colSize, ref)

	assert.Equal(ref, our, "Poseidon2 x16 mismatch")
}

func TestPolyEval(t *testing.T) {
	assert := require.New(t)
	rng := deterministicRNG(23)

	n := 16
	poly := make([]koalabear.Element, n)
	for j := range poly {
		poly[j] = randKB(rng)
	}
	x := randE4(rng)

	ourVal, err := EvalBasePolyLagrange(poly, x)
	assert.NoError(err)
	refVal, err := refvortex.EvalBasePolyLagrange(poly, x)
	assert.NoError(err)

	assert.Equal(refVal, ourVal, "EvalBasePolyLagrange mismatch")
}

func TestEvalBasePolyHorner(t *testing.T) {
	assert := require.New(t)
	rng := deterministicRNG(27)

	n := 16
	poly := make([]koalabear.Element, n)
	for i := range poly {
		poly[i] = randKB(rng)
	}
	x := randE4(rng)

	assert.Equal(refvortex.EvalBasePolyHorner(poly, x), EvalBasePolyHorner(poly, x))
}

func TestBatchPolyEvalLagrange(t *testing.T) {
	assert := require.New(t)
	rng := deterministicRNG(29)

	const (
		numPolys = 5
		n        = 16
	)
	basePolys := make([][]koalabear.Element, numPolys)
	fextPolys := make([][]fext.E4, numPolys)
	for i := 0; i < numPolys; i++ {
		basePolys[i] = make([]koalabear.Element, n)
		fextPolys[i] = make([]fext.E4, n)
		for j := 0; j < n; j++ {
			basePolys[i][j] = randKB(rng)
			fextPolys[i][j] = randE4(rng)
		}
	}
	x := randE4(rng)

	ourBase, err := BatchEvalBasePolyLagrange(basePolys, x)
	assert.NoError(err)
	refBase, err := refvortex.BatchEvalBasePolyLagrange(basePolys, x)
	assert.NoError(err)
	assert.Equal(refBase, ourBase)

	ourFext, err := BatchEvalFextPolyLagrange(fextPolys, x, true)
	assert.NoError(err)
	refFext, err := refvortex.BatchEvalFextPolyLagrange(fextPolys, x, true)
	assert.NoError(err)
	assert.Equal(refFext, ourFext)
}

// ─── Rate 8 ──────────────────────────────────────────────────────────────────

func TestRoundtripRate8(t *testing.T) {
	assert := require.New(t)
	rng := deterministicRNG(31)

	nCols := 16
	nRows := 4
	rate := 8
	nSelected := 4

	sisParams, err := sis.NewRSis(0, 9, 16, nRows)
	assert.NoError(err)

	params, err := NewParams(nCols, nRows, sisParams, rate, nSelected)
	assert.NoError(err)

	m := randMatrix(rng, nRows, nCols)

	x := randE4(rng)
	ys := make([]fext.E4, nRows)
	for i := range m {
		ys[i], err = EvalBasePolyLagrange(m[i], x)
		assert.NoError(err)
	}

	alpha := randE4(rng)
	selectedCols := []int{0, 5, 30, 100}

	cs, root, err := params.Commit(m)
	assert.NoError(err)

	proof, err := cs.Prove(alpha, selectedCols)
	assert.NoError(err)

	err = params.Verify(root, proof, ys, x, alpha, selectedCols)
	assert.NoError(err, "rate-8 roundtrip verification failed")
}

// ─── Larger matrix ───────────────────────────────────────────────────────────

func TestLargerMatrix(t *testing.T) {
	assert := require.New(t)
	rng := deterministicRNG(37)

	nCols := 256
	nRows := 64
	rate := 2
	nSelected := 16

	sisParams, err := sis.NewRSis(0, 9, 16, nRows)
	assert.NoError(err)

	params, err := NewParams(nCols, nRows, sisParams, rate, nSelected)
	assert.NoError(err)

	m := randMatrix(rng, nRows, nCols)

	x := randE4(rng)
	ys := make([]fext.E4, nRows)
	for i := range m {
		ys[i], err = EvalBasePolyLagrange(m[i], x)
		assert.NoError(err)
	}

	alpha := randE4(rng)
	selectedCols := make([]int, nSelected)
	for i := range selectedCols {
		selectedCols[i] = rng.IntN(nCols*rate - 1)
	}

	cs, root, err := params.Commit(m)
	assert.NoError(err)

	proof, err := cs.Prove(alpha, selectedCols)
	assert.NoError(err)

	err = params.Verify(root, proof, ys, x, alpha, selectedCols)
	assert.NoError(err, "large matrix roundtrip failed")
}

// ─── Benchmark ───────────────────────────────────────────────────────────────

func BenchmarkVortex(b *testing.B) {
	rng := deterministicRNG(0)

	nCols := 1024
	nRows := 128
	rate := 2
	nSelected := 32

	sisParams, _ := sis.NewRSis(0, 9, 16, nRows)
	params, _ := NewParams(nCols, nRows, sisParams, rate, nSelected)
	m := randMatrix(rng, nRows, nCols)

	alpha := randE4(rng)
	x := randE4(rng)
	selectedCols := make([]int, nSelected)
	for i := range selectedCols {
		selectedCols[i] = rng.IntN(nCols*rate - 1)
	}

	b.Run("Commit", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _, _ = params.Commit(m)
		}
	})

	cs, root, _ := params.Commit(m)
	ys := make([]fext.E4, nRows)
	for i := range m {
		ys[i], _ = EvalBasePolyLagrange(m[i], x)
	}

	b.Run("Prove", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cs, _, _ = params.Commit(m)
			_, _ = cs.Prove(alpha, selectedCols)
		}
	})

	proof, _ := cs.Prove(alpha, selectedCols)

	b.Run("Verify", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = params.Verify(root, proof, ys, x, alpha, selectedCols)
		}
	})
}
