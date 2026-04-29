package sumcheck

import (
	"math/rand/v2"
	"testing"

	"github.com/consensys/linea-monorepo/prover-ray/maths/koalabear/field"
	"github.com/consensys/linea-monorepo/prover-ray/maths/koalabear/polynomials"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

const testN = 20 // repetitions for randomised tests

func newRng() *rand.Rand { return rand.New(rand.NewPCG(0xdeadbeef, 0)) }

func randExt(rng *rand.Rand) field.Ext   { return field.PseudoRandExt(rng) }
func randBase(rng *rand.Rand) field.Element { return field.PseudoRand(rng) }

func extEq(a, b field.Ext) bool { return a.Equal(&b) }

// evalSum computes Σ_{x ∈ {0,1}ⁿ} eq(qPrime, x) · Gate(tables[0][x], …)
// naively. Used as reference in tests.
func evalSum(gate Gate, tables [][]field.Element, qPrime []field.Ext) field.Ext {
	n := len(tables[0])
	eqTable := make([]field.Ext, n)
	polynomials.FoldedEqTableExt(eqTable, qPrime)

	extTables := make([][]field.Ext, len(tables))
	for k, t := range tables {
		extTables[k] = make([]field.Ext, n)
		for i, e := range t {
			extTables[k][i] = field.Lift(e)
		}
	}

	inputs := make([][]field.Ext, len(tables))
	for k := range inputs {
		inputs[k] = make([]field.Ext, 1)
	}
	gateOut := make([]field.Ext, 1)

	var sum field.Ext
	for x := 0; x < n; x++ {
		for k := range inputs {
			inputs[k][0] = extTables[k][x]
		}
		gate.EvalBatch(gateOut, inputs...)
		var v field.Ext
		v.Mul(&eqTable[x], &gateOut[0])
		sum.Add(&sum, &v)
	}
	return sum
}

// makeProductGate builds a ProductSumGate for a single pair (λ=1).
func makeProductGate() *ProductSumGate {
	var one field.Element
	one.SetOne()
	return &ProductSumGate{Lambdas: []field.Element{one}}
}

// makeProductGateLambdas builds a ProductSumGate with m pairs and lambdas[i]=i+1.
func makeProductGateLambdas(m int) *ProductSumGate {
	lambdas := make([]field.Element, m)
	for i := range lambdas {
		lambdas[i].SetUint64(uint64(i + 1))
	}
	return &ProductSumGate{Lambdas: lambdas}
}

// deterministicChallenge produces a deterministic ext-field challenge from a
// round polynomial (simulating Fiat-Shamir without a real hash).
func deterministicChallenge(rp RoundPoly, round int) field.Ext {
	var seed field.Ext
	seed.SetOne()
	for _, v := range rp {
		seed.Add(&seed, &v)
	}
	var r field.Element
	r.SetUint64(uint64(round + 2))
	seed.MulByElement(&seed, &r)
	return seed
}

// runProof runs a full prover→verifier round trip.
func runProof(
	t *testing.T,
	cfg *ProverConfig,
	gate Gate,
	tables [][]field.Element,
	qPrimes [][]field.Ext,
	mu field.Ext,
	claim field.Ext,
) (proof []RoundPoly, challenges []field.Ext, finalClaims []field.Ext) {
	t.Helper()

	state, err := NewProverStateWithEqMask(cfg, gate, tables, qPrimes, mu, claim)
	if err != nil {
		t.Fatalf("NewProverState: %v", err)
	}

	logN := len(qPrimes[0])
	proof = make([]RoundPoly, logN)
	for i := 0; i < logN; i++ {
		rp := state.ComputeRoundPoly()
		proof[i] = rp
		ch := deterministicChallenge(rp, i)
		state.FoldAndAdvance(rp, ch)
	}
	finalClaims = state.FinalClaims()

	// Run verifier with the same deterministic challenges.
	round := 0
	finalClaim, challenges, err := Verify(claim, proof, gate.Degree(), func(rp RoundPoly) field.Ext {
		ch := deterministicChallenge(rp, round)
		round++
		return ch
	})
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}

	if !finalClaim.Equal(&state.claim) {
		t.Errorf("verifier finalClaim ≠ prover final claim")
	}

	return proof, challenges, finalClaims
}

// extToGen converts []field.Ext to []field.Gen for use with EvalMultilin.
func extToGen(exts []field.Ext) []field.Gen {
	out := make([]field.Gen, len(exts))
	for i, e := range exts {
		out[i] = field.ElemFromExt(e)
	}
	return out
}

// combineClaims computes Σ_j mu^j * claims[j].
func combineClaims(claims []field.Ext, mu field.Ext) field.Ext {
	var sum, muPow field.Ext
	muPow.SetOne()
	for _, c := range claims {
		var term field.Ext
		term.Mul(&c, &muPow)
		sum.Add(&sum, &term)
		muPow.Mul(&muPow, &mu)
	}
	return sum
}

// ---------------------------------------------------------------------------
// Unit tests: ProductSumGate
// ---------------------------------------------------------------------------

func TestProductSumGateEvalBatch(t *testing.T) {
	rng := newRng()
	const n = 64

	lambdas := make([]field.Element, 3)
	for i := range lambdas {
		lambdas[i] = randBase(rng)
	}
	gate := &ProductSumGate{Lambdas: lambdas}

	inputs := make([][]field.Ext, 2*len(lambdas))
	for k := range inputs {
		inputs[k] = make([]field.Ext, n)
		for j := range inputs[k] {
			inputs[k][j] = randExt(rng)
		}
	}

	res := make([]field.Ext, n)
	gate.EvalBatch(res, inputs...)

	for j := 0; j < n; j++ {
		var want field.Ext
		for i, λ := range lambdas {
			var prod, tmp field.Ext
			prod.Mul(&inputs[2*i][j], &inputs[2*i+1][j])
			tmp.MulByElement(&prod, &λ)
			want.Add(&want, &tmp)
		}
		if !extEq(res[j], want) {
			t.Fatalf("j=%d: got %v want %v", j, res[j], want)
		}
	}
}

// ---------------------------------------------------------------------------
// Unit tests: RoundPoly.EvalAt
// ---------------------------------------------------------------------------

func TestRoundPolyEvalAtKnownPoints(t *testing.T) {
	rng := newRng()

	for range testN {
		p0 := randExt(rng)
		p1 := randExt(rng)
		p2 := randExt(rng)

		var claim field.Ext
		claim.Add(&p0, &p1)

		rp := RoundPoly{p0, p2}

		var zero, one, two field.Ext
		zero.SetZero()
		one.SetOne()
		two.Add(&one, &one)

		if got := rp.EvalAt(zero, claim); !extEq(got, p0) {
			t.Fatalf("EvalAt(0): got %v want %v", got, p0)
		}
		if got := rp.EvalAt(one, claim); !extEq(got, p1) {
			t.Fatalf("EvalAt(1): got %v want %v", got, p1)
		}
		if got := rp.EvalAt(two, claim); !extEq(got, p2) {
			t.Fatalf("EvalAt(2): got %v want %v", got, p2)
		}
	}
}

// ---------------------------------------------------------------------------
// Integration tests: prover + verifier
// ---------------------------------------------------------------------------

func TestSumcheckSoundnessSingle(t *testing.T) {
	rng := newRng()
	gate := makeProductGate()

	for logN := 1; logN <= 10; logN++ {
		n := 1 << logN
		cfg := NewProverConfig(logN, 2, 1)

		A := field.VecPseudoRandBase(rng, n)
		B := field.VecPseudoRandBase(rng, n)

		qPrime := make([]field.Ext, logN)
		for i := range qPrime {
			qPrime[i] = field.Lift(randBase(rng))
		}

		claim := evalSum(gate, [][]field.Element{A, B}, qPrime)
		_, challenges, finalClaims := runProof(t, cfg, gate, [][]field.Element{A, B}, [][]field.Ext{qPrime}, field.Ext{}, claim)

		gotA := polynomials.EvalMultilin(field.VecFromBase(A), extToGen(challenges))
		gotB := polynomials.EvalMultilin(field.VecFromBase(B), extToGen(challenges))

		if !extEq(gotA.AsExt(), finalClaims[1]) {
			t.Errorf("logN=%d: A(challenges) mismatch", logN)
		}
		if !extEq(gotB.AsExt(), finalClaims[2]) {
			t.Errorf("logN=%d: B(challenges) mismatch", logN)
		}

		gotEq := polynomials.EvalEqExt(challenges, qPrime)
		if !extEq(gotEq, finalClaims[0]) {
			t.Errorf("logN=%d: eq(challenges,qPrime) mismatch", logN)
		}
	}
}

func TestSumcheckFinalClaimsMatchEvalMultilin(t *testing.T) {
	rng := newRng()
	gate := makeProductGate()

	for range testN {
		logN := 1 + int(rng.Uint32()%8)
		n := 1 << logN
		cfg := NewProverConfig(logN, 2, 1)

		A := field.VecPseudoRandBase(rng, n)
		B := field.VecPseudoRandBase(rng, n)

		qPrime := make([]field.Ext, logN)
		for i := range qPrime {
			qPrime[i] = field.Lift(randBase(rng))
		}
		claim := evalSum(gate, [][]field.Element{A, B}, qPrime)

		_, challenges, finalClaims := runProof(t, cfg, gate, [][]field.Element{A, B}, [][]field.Ext{qPrime}, field.Ext{}, claim)

		gens := extToGen(challenges)
		wantA := polynomials.EvalMultilin(field.VecFromBase(A), gens)
		wantB := polynomials.EvalMultilin(field.VecFromBase(B), gens)

		if !extEq(wantA.AsExt(), finalClaims[1]) {
			t.Errorf("A(challenges) mismatch at logN=%d", logN)
		}
		if !extEq(wantB.AsExt(), finalClaims[2]) {
			t.Errorf("B(challenges) mismatch at logN=%d", logN)
		}
	}
}

func TestSumcheckMultiInstance(t *testing.T) {
	rng := newRng()
	gate := makeProductGate()

	for _, m := range []int{2, 3, 5} {
		logN := 6
		n := 1 << logN
		cfg := NewProverConfig(logN, 2, 1)

		A := field.VecPseudoRandBase(rng, n)
		B := field.VecPseudoRandBase(rng, n)
		tables := [][]field.Element{A, B}

		qPrimes := make([][]field.Ext, m)
		claims := make([]field.Ext, m)
		for j := range qPrimes {
			q := make([]field.Ext, logN)
			for i := range q {
				q[i] = field.Lift(randBase(rng))
			}
			qPrimes[j] = q
			claims[j] = evalSum(gate, tables, q)
		}

		mu := field.Lift(randBase(rng))
		combinedClaim := combineClaims(claims, mu)

		runProof(t, cfg, gate, tables, qPrimes, mu, combinedClaim)
	}
}

func TestSumcheckMultipleLambdas(t *testing.T) {
	rng := newRng()

	logN := 5
	n := 1 << logN
	m := 3
	cfg := NewProverConfig(logN, 2*m, 1)
	gate := makeProductGateLambdas(m)

	tables := make([][]field.Element, 2*m)
	for k := range tables {
		tables[k] = field.VecPseudoRandBase(rng, n)
	}

	qPrime := make([]field.Ext, logN)
	for i := range qPrime {
		qPrime[i] = field.Lift(randBase(rng))
	}

	claim := evalSum(gate, tables, qPrime)
	runProof(t, cfg, gate, tables, [][]field.Ext{qPrime}, field.Ext{}, claim)
}

func TestSumcheckParallel(t *testing.T) {
	rng := newRng()
	gate := makeProductGate()

	logN := 8
	n := 1 << logN

	A := field.VecPseudoRandBase(rng, n)
	B := field.VecPseudoRandBase(rng, n)

	qPrime := make([]field.Ext, logN)
	for i := range qPrime {
		qPrime[i] = field.Lift(randBase(rng))
	}
	claim := evalSum(gate, [][]field.Element{A, B}, qPrime)

	cfg1 := NewProverConfig(logN, 2, 1)
	cfg4 := NewProverConfig(logN, 2, 4)

	_, _, fc1 := runProof(t, cfg1, gate, [][]field.Element{A, B}, [][]field.Ext{qPrime}, field.Ext{}, claim)
	_, _, fc4 := runProof(t, cfg4, gate, [][]field.Element{A, B}, [][]field.Ext{qPrime}, field.Ext{}, claim)

	for i := range fc1 {
		if !fc1[i].Equal(&fc4[i]) {
			t.Errorf("finalClaims[%d] differ between 1-CPU and 4-CPU runs", i)
		}
	}
}

// ---------------------------------------------------------------------------
// Tests: NewProverStateWithMask
// ---------------------------------------------------------------------------

// identityGate is a local degree-1 gate used only in tests.
type identityGate struct{}

func (identityGate) Degree() int { return 1 }
func (identityGate) EvalBatch(res []field.Ext, inputs ...[]field.Ext) { copy(res, inputs[0]) }

// evalPolyAtPoint evaluates Σ_h z^h · coeffs[h], the univariate polynomial
// with base-field coefficient vector coeffs at extension-field point z.
func evalPolyAtPoint(coeffs []field.Element, z field.Ext) field.Ext {
	var result, zPow field.Ext
	zPow.SetOne()
	for _, c := range coeffs {
		var term field.Ext
		term.MulByElement(&zPow, &c)
		result.Add(&result, &term)
		zPow.Mul(&zPow, &z)
	}
	return result
}

// runProofWithMask is the NewProverStateWithMask analogue of runProof.
func runProofWithMask(
	t *testing.T,
	cfg *ProverConfig,
	gate Gate,
	tables [][]field.Element,
	mask []field.Ext,
	claim field.Ext,
) (proof []RoundPoly, challenges []field.Ext, finalClaims []field.Ext) {
	t.Helper()

	state, err := NewProverStateWithMask(cfg, gate, tables, mask, claim)
	if err != nil {
		t.Fatalf("NewProverStateWithMask: %v", err)
	}

	logN := state.logN
	proof = make([]RoundPoly, logN)
	for i := range logN {
		rp := state.ComputeRoundPoly()
		proof[i] = rp
		ch := deterministicChallenge(rp, i)
		state.FoldAndAdvance(rp, ch)
	}
	finalClaims = state.FinalClaims()

	round := 0
	finalClaim, challenges, err := Verify(claim, proof, gate.Degree(), func(rp RoundPoly) field.Ext {
		ch := deterministicChallenge(rp, round)
		round++
		return ch
	})
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if !finalClaim.Equal(&state.claim) {
		t.Errorf("verifier finalClaim ≠ prover final claim")
	}

	return proof, challenges, finalClaims
}

// TestSumcheckGeomMaskSingle verifies that a single-point geometric mask
// sumcheck is sound and that the final mask claim matches EvalMonomialMaskExt.
func TestSumcheckGeomMaskSingle(t *testing.T) {
	rng := newRng()

	for logN := 1; logN <= 8; logN++ {
		n := 1 << logN
		cfg := NewProverConfig(logN, 1, 1)

		coeffs := field.VecPseudoRandBase(rng, n)
		r := randExt(rng)

		// Build geometric mask M[h] = r^h.
		var one field.Ext
		one.SetOne()
		mask := make([]field.Ext, n)
		polynomials.BuildMonomialMaskExt(mask, []field.Ext{r}, []field.Ext{one})

		// Reference claim: Y = Σ_h mask[h] * coeffs[h] = P(r).
		claim := evalPolyAtPoint(coeffs, r)
		var altClaim field.Ext
		for h := range n {
			c := field.Lift(coeffs[h])
			var term field.Ext
			term.Mul(&mask[h], &c)
			altClaim.Add(&altClaim, &term)
		}
		if !claim.Equal(&altClaim) {
			t.Fatalf("logN=%d: claim sanity check failed", logN)
		}

		_, challenges, finalClaims := runProofWithMask(t, cfg, identityGate{}, [][]field.Element{coeffs}, mask, claim)

		// finalClaims[0] = mask(h'), finalClaims[1] = coeffs_MLE(h').
		wantMask := polynomials.EvalMonomialMaskExt(r, challenges)
		if !finalClaims[0].Equal(&wantMask) {
			t.Errorf("logN=%d: mask final claim mismatch", logN)
		}

		hGen := make([]field.Gen, logN)
		for i, h := range challenges {
			hGen[i] = field.ElemFromExt(h)
		}
		wantCoeffsGen := polynomials.EvalMultilin(field.VecFromBase(coeffs), hGen)
		wantCoeffsExt := wantCoeffsGen.AsExt()
		if !finalClaims[1].Equal(&wantCoeffsExt) {
			t.Errorf("logN=%d: coeffs final claim mismatch", logN)
		}
	}
}

// TestSumcheckGeomMaskMultiPoint verifies a multi-point MPTS-style sumcheck:
// Y = Σ_j λ^j Σ_h geom(r_j, h) P(h) using a combined geometric mask.
func TestSumcheckGeomMaskMultiPoint(t *testing.T) {
	rng := newRng()
	const logN = 7
	const m = 4
	n := 1 << logN
	cfg := NewProverConfig(logN, 1, 1)

	coeffs := field.VecPseudoRandBase(rng, n)
	rs := field.VecPseudoRandExt(rng, m)

	// lambdas[j] = λ^j.
	lambdaBase := randExt(rng)
	lambdas := make([]field.Ext, m)
	lambdas[0].SetOne()
	for j := 1; j < m; j++ {
		lambdas[j].Mul(&lambdas[j-1], &lambdaBase)
	}

	// Combined mask M[h] = Σ_j λ^j · r_j^h.
	mask := make([]field.Ext, n)
	polynomials.BuildMonomialMaskExt(mask, rs, lambdas)

	// Reference claim: Y = Σ_j λ^j · P(r_j).
	var claim field.Ext
	for j := range m {
		Prj := evalPolyAtPoint(coeffs, rs[j])
		var term field.Ext
		term.Mul(&lambdas[j], &Prj)
		claim.Add(&claim, &term)
	}

	_, challenges, finalClaims := runProofWithMask(t, cfg, identityGate{}, [][]field.Element{coeffs}, mask, claim)

	// Mask final claim must equal Σ_j λ^j · EvalMonomialMaskExt(r_j, h').
	var wantMask field.Ext
	for j := range m {
		g := polynomials.EvalMonomialMaskExt(rs[j], challenges)
		var term field.Ext
		term.Mul(&lambdas[j], &g)
		wantMask.Add(&wantMask, &term)
	}
	if !finalClaims[0].Equal(&wantMask) {
		t.Errorf("mask final claim mismatch")
	}

	hGen := make([]field.Gen, logN)
	for i, h := range challenges {
		hGen[i] = field.ElemFromExt(h)
	}
	wantCoeffsGen := polynomials.EvalMultilin(field.VecFromBase(coeffs), hGen)
	wantCoeffsExt := wantCoeffsGen.AsExt()
	if !finalClaims[1].Equal(&wantCoeffsExt) {
		t.Errorf("coeffs final claim mismatch")
	}
}

// TestSumcheckNoMask verifies the nil-mask path: Σ_h A[h]·B[h] without any
// external weighting. The mask final claim must be 1.
func TestSumcheckNoMask(t *testing.T) {
	rng := newRng()

	for logN := 1; logN <= 8; logN++ {
		n := 1 << logN
		cfg := NewProverConfig(logN, 2, 1)

		A := field.VecPseudoRandBase(rng, n)
		B := field.VecPseudoRandBase(rng, n)

		// Y = Σ_h A[h] * B[h].
		var claim field.Ext
		for h := range n {
			a, b := field.Lift(A[h]), field.Lift(B[h])
			var term field.Ext
			term.Mul(&a, &b)
			claim.Add(&claim, &term)
		}

		_, challenges, finalClaims := runProofWithMask(t, cfg, makeProductGate(), [][]field.Element{A, B}, nil, claim)

		// Mask claim must be 1 (all-ones folds to 1 under any challenge).
		var one field.Ext
		one.SetOne()
		if !finalClaims[0].Equal(&one) {
			t.Errorf("logN=%d: mask final claim = %v, want 1", logN, finalClaims[0])
		}

		hGen := make([]field.Gen, logN)
		for i, h := range challenges {
			hGen[i] = field.ElemFromExt(h)
		}
		wantAGen := polynomials.EvalMultilin(field.VecFromBase(A), hGen)
		wantBGen := polynomials.EvalMultilin(field.VecFromBase(B), hGen)
		wantAExt, wantBExt := wantAGen.AsExt(), wantBGen.AsExt()
		if !finalClaims[1].Equal(&wantAExt) {
			t.Errorf("logN=%d: A final claim mismatch", logN)
		}
		if !finalClaims[2].Equal(&wantBExt) {
			t.Errorf("logN=%d: B final claim mismatch", logN)
		}
	}
}

// ---------------------------------------------------------------------------
// Benchmarks
// ---------------------------------------------------------------------------

func BenchmarkProveProductSum(b *testing.B) {
	const logN = 22
	n := 1 << logN
	rng := newRng()
	gate := makeProductGate()
	cfg := NewProverConfig(logN, 2, 0)

	A := field.VecPseudoRandBase(rng, n)
	B := field.VecPseudoRandBase(rng, n)

	qPrime := make([]field.Ext, logN)
	for i := range qPrime {
		qPrime[i] = field.Lift(randBase(rng))
	}

	b.ResetTimer()
	for range b.N {
		claim := evalSum(gate, [][]field.Element{A, B}, qPrime)
		state, _ := NewProverStateWithEqMask(cfg, gate, [][]field.Element{A, B}, [][]field.Ext{qPrime}, field.Ext{}, claim)

		proof := make([]RoundPoly, logN)
		for i := 0; i < logN; i++ {
			rp := state.ComputeRoundPoly()
			proof[i] = rp
			ch := deterministicChallenge(rp, i)
			state.FoldAndAdvance(rp, ch)
		}
	}
}
