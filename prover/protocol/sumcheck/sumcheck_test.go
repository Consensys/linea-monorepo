package sumcheck

import (
	"math/rand/v2"
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/stretchr/testify/require"
)

// makeRng returns a deterministic RNG so test failures can be reproduced.
func makeRng(seed uint64) *rand.Rand {
	return rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15))
}

// randPoly returns a random multilinear polynomial in n variables, expressed
// directly over fext.
func randPoly(rng *rand.Rand, n int) MultiLin {
	size := 1 << n
	out := make(MultiLin, size)
	for i := range out {
		out[i] = fext.PseudoRand(rng)
	}
	return out
}

// randPoint returns a uniformly random fext-valued point of dimension n.
func randPoint(rng *rand.Rand, n int) []fext.Element {
	out := make([]fext.Element, n)
	for i := range out {
		out[i] = fext.PseudoRand(rng)
	}
	return out
}

func boolPoint(idx, n int) []fext.Element {
	out := make([]fext.Element, n)
	for i := 0; i < n; i++ {
		bit := (idx >> (n - 1 - i)) & 1
		if bit == 1 {
			out[i].SetOne()
		}
	}
	return out
}

// TestEvalEqIndicator checks that eq(q, h) for q,h ∈ {0,1}^n is 1 iff q==h.
func TestEvalEqIndicator(t *testing.T) {
	for _, n := range []int{1, 2, 3, 5} {
		for q := 0; q < (1 << n); q++ {
			for h := 0; h < (1 << n); h++ {
				got := EvalEq(boolPoint(q, n), boolPoint(h, n))
				if q == h {
					require.True(t, got.IsOne(), "n=%d q=%d h=%d expected 1, got %s", n, q, h, got.String())
				} else {
					require.True(t, got.IsZero(), "n=%d q=%d h=%d expected 0, got %s", n, q, h, got.String())
				}
			}
		}
	}
}

// TestBuildEqTableMatchesEvalEq checks that BuildEqTable(r) tabulates eq(r, *)
// correctly: T[idx(h)] == EvalEq(r, h) for every h on the hypercube.
func TestBuildEqTableMatchesEvalEq(t *testing.T) {
	rng := makeRng(0xCAFE_F00D)
	for _, n := range []int{1, 2, 3, 5} {
		r := randPoint(rng, n)
		table := BuildEqTable(r)
		require.Equal(t, 1<<n, len(table))
		for h := 0; h < (1 << n); h++ {
			expected := EvalEq(r, boolPoint(h, n))
			require.True(t, table[h].Equal(&expected),
				"n=%d h=%d table=%s expected=%s", n, h, table[h].String(), expected.String())
		}
	}
}

// TestMultiLinEvaluateOnHypercube checks that evaluating the polynomial at a
// boolean point reproduces the corresponding table entry.
func TestMultiLinEvaluateOnHypercube(t *testing.T) {
	rng := makeRng(0xDEADBEEF)
	for _, n := range []int{1, 2, 3, 5} {
		p := randPoly(rng, n)
		for h := 0; h < (1 << n); h++ {
			got := p.Evaluate(boolPoint(h, n))
			want := p[h]
			require.True(t, got.Equal(&want), "n=%d h=%d", n, h)
		}
	}
}

// TestMultiLinEvaluateOffHypercube cross-checks Evaluate against a direct
// expansion using the Lagrange basis (which is the eq table).
func TestMultiLinEvaluateOffHypercube(t *testing.T) {
	rng := makeRng(0xBADC0FFEE0DDF00D)
	for _, n := range []int{1, 2, 4, 6} {
		p := randPoly(rng, n)
		r := randPoint(rng, n)
		got := p.Evaluate(r)

		// Direct computation: P(r) = Σ_h P[h] · eq(h, r).
		var want fext.Element
		eqTable := BuildEqTable(r) // T[h] = eq(r, h) = eq(h, r) (eq is symmetric)
		var term fext.Element
		for h := 0; h < (1 << n); h++ {
			term.Mul(&p[h], &eqTable[h])
			want.Add(&want, &term)
		}
		require.True(t, got.Equal(&want), "n=%d", n)
	}
}

// TestSumcheckSingleClaim runs the batched sumcheck with N=1 — i.e. a plain
// multilinear evaluation proof — and checks both prover and verifier agree.
func TestSumcheckSingleClaim(t *testing.T) {
	rng := makeRng(42)
	for _, n := range []int{1, 2, 4, 6} {
		p := randPoly(rng, n)
		r := randPoint(rng, n)
		y := p.Evaluate(r)

		claims := []Claim{{Point: r, Eval: y}}
		polys := []MultiLin{p}

		ptr := NewMockTranscript("single-claim")
		proof, err := ProveBatched(claims, polys, ptr)
		require.NoError(t, err, "n=%d", n)

		vtr := NewMockTranscript("single-claim")
		point, evals, err := VerifyBatched(claims, proof, vtr)
		require.NoError(t, err, "n=%d", n)
		require.Len(t, point, n)
		require.Len(t, evals, 1)

		// Residual evaluation is P(c) — sanity-check that this matches a
		// direct evaluation (the prover sent the truth).
		want := p.Evaluate(point)
		require.True(t, evals[0].Equal(&want), "n=%d residual mismatch", n)
	}
}

// TestSumcheckBatched runs the batched protocol for several (N, n) shapes and
// verifies both correctness and that the residual evaluations are P_i(c).
func TestSumcheckBatched(t *testing.T) {
	cases := []struct {
		N, n int
		seed uint64
	}{
		{N: 2, n: 3, seed: 1},
		{N: 4, n: 4, seed: 2},
		{N: 8, n: 5, seed: 3},
		{N: 16, n: 8, seed: 4},
	}
	for _, tc := range cases {
		t.Run("", func(t *testing.T) {
			rng := makeRng(tc.seed)
			polys := make([]MultiLin, tc.N)
			claims := make([]Claim, tc.N)
			for i := 0; i < tc.N; i++ {
				polys[i] = randPoly(rng, tc.n)
				r := randPoint(rng, tc.n)
				claims[i] = Claim{Point: r, Eval: polys[i].Evaluate(r)}
			}

			ptr := NewMockTranscript("batched")
			proof, err := ProveBatched(claims, polys, ptr)
			require.NoError(t, err)

			vtr := NewMockTranscript("batched")
			point, evals, err := VerifyBatched(claims, proof, vtr)
			require.NoError(t, err)
			require.Len(t, point, tc.n)
			require.Len(t, evals, tc.N)

			for i := 0; i < tc.N; i++ {
				want := polys[i].Evaluate(point)
				require.True(t, evals[i].Equal(&want), "poly %d residual mismatch", i)
			}
		})
	}
}

// TestSumcheckRejectsBadEval checks that a single bit-flip on a claimed
// evaluation makes the verifier reject (round-0 consistency fires first
// because the initial target Σ λ^i y_i is wrong).
func TestSumcheckRejectsBadEval(t *testing.T) {
	rng := makeRng(7)
	const N, n = 3, 4
	polys := make([]MultiLin, N)
	claims := make([]Claim, N)
	for i := 0; i < N; i++ {
		polys[i] = randPoly(rng, n)
		r := randPoint(rng, n)
		claims[i] = Claim{Point: r, Eval: polys[i].Evaluate(r)}
	}

	ptr := NewMockTranscript("reject-eval")
	proof, err := ProveBatched(claims, polys, ptr)
	require.NoError(t, err)

	// Mutate the FIRST claimed evaluation. Both prover and verifier transcripts
	// are constructed from `claims`, so a tampered claim breaks consistency.
	var bump fext.Element
	bump.SetOne()
	tampered := make([]Claim, N)
	copy(tampered, claims)
	tampered[0].Eval.Add(&tampered[0].Eval, &bump)

	vtr := NewMockTranscript("reject-eval")
	_, _, err = VerifyBatched(tampered, proof, vtr)
	require.Error(t, err)
}

// TestSumcheckRejectsBadRoundPoly mutates a single round polynomial coefficient
// and checks the verifier rejects.
func TestSumcheckRejectsBadRoundPoly(t *testing.T) {
	rng := makeRng(11)
	const N, n = 4, 4
	polys := make([]MultiLin, N)
	claims := make([]Claim, N)
	for i := 0; i < N; i++ {
		polys[i] = randPoly(rng, n)
		r := randPoint(rng, n)
		claims[i] = Claim{Point: r, Eval: polys[i].Evaluate(r)}
	}

	ptr := NewMockTranscript("reject-round")
	proof, err := ProveBatched(claims, polys, ptr)
	require.NoError(t, err)

	// Tamper with the LAST round poly's evaluation at X=2. This breaks the
	// final consistency check because the verifier's target diverges from the
	// honest Σ λ^i · eq · P_i(c).
	var bump fext.Element
	bump.SetOne()
	proof.RoundPolys[n-1][2].Add(&proof.RoundPolys[n-1][2], &bump)

	vtr := NewMockTranscript("reject-round")
	_, _, err = VerifyBatched(claims, proof, vtr)
	require.Error(t, err)
}

// TestFoldEqualsEvaluate proves the obvious-but-load-bearing invariant that
// repeatedly folding gives the same answer as Evaluate.
func TestFoldEqualsEvaluate(t *testing.T) {
	rng := makeRng(0x5EED)
	for _, n := range []int{1, 3, 5} {
		p := randPoly(rng, n)
		r := randPoint(rng, n)
		want := p.Evaluate(r)

		work := p.Clone()
		for _, ri := range r {
			work.Fold(ri)
		}
		require.Equal(t, 1, len(work))
		require.True(t, work[0].Equal(&want))
	}
}
