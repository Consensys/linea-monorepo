package sumcheck

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
)

// ProveBatched runs the multilinear sumcheck protocol that reduces N input
// claims (P_i, r_i, y_i) sharing the same number of variables n into a single
// residual point c plus the per-polynomial evaluations P_i(c). The reduction
// proves
//
//	Σ_{x ∈ {0,1}^n} ( Σ_i λ^i · eq(r_i, x) · P_i(x) )  =  Σ_i λ^i · y_i
//
// where λ is the transcript challenge after absorbing the input claims. After
// n rounds, the residual claim is g(c) = Σ_i λ^i · eq(r_i, c) · P_i(c) which
// the verifier checks against the per-poly final evaluations sent in the
// returned BatchedProof.
//
// The transcript is mutated. polys are not mutated (cloned internally).
func ProveBatched(claims []Claim, polys []MultiLin, t Transcript) (BatchedProof, error) {
	N := len(claims)
	if N == 0 {
		return BatchedProof{}, fmt.Errorf("sumcheck.ProveBatched: no claims")
	}
	if len(polys) != N {
		return BatchedProof{}, fmt.Errorf("sumcheck.ProveBatched: %d polys for %d claims", len(polys), N)
	}
	n := len(claims[0].Point)
	if n == 0 {
		return BatchedProof{}, fmt.Errorf("sumcheck.ProveBatched: 0-variable claims are not supported")
	}
	for i := range claims {
		if len(claims[i].Point) != n {
			return BatchedProof{}, fmt.Errorf("sumcheck.ProveBatched: claim %d has %d vars, expected %d",
				i, len(claims[i].Point), n)
		}
		if polys[i].NumVars() != n {
			return BatchedProof{}, fmt.Errorf("sumcheck.ProveBatched: poly %d has %d vars, expected %d",
				i, polys[i].NumVars(), n)
		}
	}

	// Absorb claims.
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

	// Working tables: poly[i] starts as a clone, eq[i] is built from r_i.
	// Both get folded in lock-step at every round challenge.
	polyTables := make([]MultiLin, N)
	eqTables := make([]MultiLin, N)
	for i := 0; i < N; i++ {
		polyTables[i] = polys[i].Clone()
		eqTables[i] = BuildEqTable(claims[i].Point)
	}

	proof := BatchedProof{
		RoundPolys: make([][3]fext.Element, n),
	}

	var (
		g0, g1, g2 fext.Element
		t1, e2, p2 fext.Element
	)

	for k := 0; k < n; k++ {
		g0.SetZero()
		g1.SetZero()
		g2.SetZero()

		size := len(polyTables[0])
		mid := size / 2

		for i := 0; i < N; i++ {
			lp := lambdaPows[i]
			pt := polyTables[i]
			et := eqTables[i]
			for x := 0; x < mid; x++ {
				p0 := pt[x]
				p1 := pt[x+mid]
				e0 := et[x]
				e1 := et[x+mid]

				// g0 += λ^i · e0 · p0
				t1.Mul(&e0, &p0)
				t1.Mul(&t1, &lp)
				g0.Add(&g0, &t1)

				// g1 += λ^i · e1 · p1
				t1.Mul(&e1, &p1)
				t1.Mul(&t1, &lp)
				g1.Add(&g1, &t1)

				// g2 += λ^i · (2e1 - e0) · (2p1 - p0)
				e2.Double(&e1)
				e2.Sub(&e2, &e0)
				p2.Double(&p1)
				p2.Sub(&p2, &p0)
				t1.Mul(&e2, &p2)
				t1.Mul(&t1, &lp)
				g2.Add(&g2, &t1)
			}
		}

		proof.RoundPolys[k] = [3]fext.Element{g0, g1, g2}
		t.Append(labelRoundPoly, g0, g1, g2)
		c := t.Challenge(labelRoundChallenge)

		for i := 0; i < N; i++ {
			polyTables[i].Fold(c)
			eqTables[i].Fold(c)
		}
	}

	proof.FinalEvals = make([]fext.Element, N)
	for i := 0; i < N; i++ {
		proof.FinalEvals[i] = polyTables[i][0]
	}
	t.Append(labelFinalEvals, proof.FinalEvals...)

	return proof, nil
}
