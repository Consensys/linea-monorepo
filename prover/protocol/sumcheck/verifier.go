package sumcheck

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
)

// twoInv is 1/2 in fext, used for the Newton-form interpolation of the
// degree-2 round polynomial.
var twoInv fext.Element

func init() {
	var two fext.Element
	two.SetOne()
	two.Double(&two)
	twoInv.Inverse(&two)
}

// VerifyBatched checks a BatchedProof against the input claims and returns
// the residual sumcheck point c together with the per-claim residual
// evaluations P_i(c) (echoed from the proof). The caller is responsible for
// validating those residual evaluations against an external commitment.
func VerifyBatched(claims []Claim, proof BatchedProof, t Transcript) (point []fext.Element, polyEvals []fext.Element, err error) {
	N := len(claims)
	if N == 0 {
		return nil, nil, fmt.Errorf("sumcheck.VerifyBatched: no claims")
	}
	n := len(claims[0].Point)
	if n == 0 {
		return nil, nil, fmt.Errorf("sumcheck.VerifyBatched: 0-variable claims are not supported")
	}
	for i := range claims {
		if len(claims[i].Point) != n {
			return nil, nil, fmt.Errorf("sumcheck.VerifyBatched: claim %d has %d vars, expected %d",
				i, len(claims[i].Point), n)
		}
	}
	if len(proof.RoundPolys) != n {
		return nil, nil, fmt.Errorf("sumcheck.VerifyBatched: proof has %d round polys, expected %d",
			len(proof.RoundPolys), n)
	}
	if len(proof.FinalEvals) != N {
		return nil, nil, fmt.Errorf("sumcheck.VerifyBatched: proof has %d final evals, expected %d",
			len(proof.FinalEvals), N)
	}

	// Absorb claims and sample lambda — must mirror the prover exactly.
	for i := range claims {
		t.Append(labelClaimPoint, claims[i].Point...)
		t.Append(labelClaimEval, claims[i].Eval)
	}
	lambda := t.Challenge(labelLambda)

	lambdaPows := make([]fext.Element, N)
	lambdaPows[0].SetOne()
	for i := 1; i < N; i++ {
		lambdaPows[i].Mul(&lambdaPows[i-1], &lambda)
	}

	// Initial target: T_0 = Σ_i λ^i · y_i.
	var target, term fext.Element
	for i := 0; i < N; i++ {
		term.Mul(&lambdaPows[i], &claims[i].Eval)
		target.Add(&target, &term)
	}

	challenges := make([]fext.Element, n)
	for k := 0; k < n; k++ {
		g0 := proof.RoundPolys[k][0]
		g1 := proof.RoundPolys[k][1]
		g2 := proof.RoundPolys[k][2]

		// Round consistency: g_k(0) + g_k(1) must equal the running target.
		var sum01 fext.Element
		sum01.Add(&g0, &g1)
		if !sum01.Equal(&target) {
			return nil, nil, fmt.Errorf("sumcheck.VerifyBatched: round %d consistency check failed", k)
		}

		t.Append(labelRoundPoly, g0, g1, g2)
		c := t.Challenge(labelRoundChallenge)
		challenges[k] = c

		// Update target := g_k(c) via degree-2 interpolation through (0,g0),(1,g1),(2,g2).
		// Newton form: g(c) = g0 + c·(g1-g0) + c·(c-1)·(g0 - 2·g1 + g2)/2.
		target = evalRoundPolyAt(g0, g1, g2, c)
	}

	// Final consistency: target must equal Σ_i λ^i · eq(r_i, c) · P_i(c).
	var expected, contrib fext.Element
	for i := 0; i < N; i++ {
		eqVal := EvalEq(claims[i].Point, challenges)
		contrib.Mul(&lambdaPows[i], &eqVal)
		contrib.Mul(&contrib, &proof.FinalEvals[i])
		expected.Add(&expected, &contrib)
	}
	if !expected.Equal(&target) {
		return nil, nil, fmt.Errorf("sumcheck.VerifyBatched: final consistency check failed")
	}

	t.Append(labelFinalEvals, proof.FinalEvals...)

	return challenges, proof.FinalEvals, nil
}

// evalRoundPolyAt evaluates the unique degree-2 polynomial through the points
// (0,g0), (1,g1), (2,g2) at c, using a Newton form:
//
//	g(c) = g0 + c·(g1 - g0) + c·(c-1)·(g0 - 2·g1 + g2)/2
func evalRoundPolyAt(g0, g1, g2, c fext.Element) fext.Element {
	var (
		linDelta, quadDelta, twoG1, cMinus1, quadFactor, term, out fext.Element
	)
	out.Set(&g0)

	// linear term: c · (g1 - g0)
	linDelta.Sub(&g1, &g0)
	term.Mul(&c, &linDelta)
	out.Add(&out, &term)

	// quadratic term: c·(c-1) · (g0 - 2·g1 + g2) / 2
	twoG1.Double(&g1)
	quadDelta.Sub(&g0, &twoG1)
	quadDelta.Add(&quadDelta, &g2)
	quadDelta.Mul(&quadDelta, &twoInv)

	var one fext.Element
	one.SetOne()
	cMinus1.Sub(&c, &one)
	quadFactor.Mul(&c, &cMinus1)
	term.Mul(&quadFactor, &quadDelta)
	out.Add(&out, &term)

	return out
}
