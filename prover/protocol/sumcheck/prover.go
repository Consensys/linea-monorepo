package sumcheck

import (
	"fmt"
	"runtime"
	"sync"

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
	if err := validateBatchedInputs(claims, polys); err != nil {
		return BatchedProof{}, err
	}
	// Absorb claims into the transcript and derive λ.
	for i := range claims {
		t.Append(labelClaimPoint, claims[i].Point...)
		t.Append(labelClaimEval, claims[i].Eval)
	}
	lambda := t.Challenge(labelLambda)
	proof, _, err := proveBatchedCore(claims, polys, lambda, t)
	return proof, err
}

// ProveBatchedWith is like [ProveBatched] but uses a pre-supplied lambda
// rather than deriving it from the transcript. The transcript t should already
// be seeded with lambda so that the per-round challenges are bound to it.
// It returns the proof plus the n per-round challenges forming the residual
// evaluation point.
func ProveBatchedWith(claims []Claim, polys []MultiLin, lambda fext.Element, t Transcript) (BatchedProof, []fext.Element, error) {
	if err := validateBatchedInputs(claims, polys); err != nil {
		return BatchedProof{}, nil, err
	}
	return proveBatchedCore(claims, polys, lambda, t)
}

func validateBatchedInputs(claims []Claim, polys []MultiLin) error {
	N := len(claims)
	if N == 0 {
		return fmt.Errorf("sumcheck: no claims")
	}
	if len(polys) != N {
		return fmt.Errorf("sumcheck: %d polys for %d claims", len(polys), N)
	}
	n := len(claims[0].Point)
	if n == 0 {
		return fmt.Errorf("sumcheck: 0-variable claims are not supported")
	}
	for i := range claims {
		if len(claims[i].Point) != n {
			return fmt.Errorf("sumcheck: claim %d has %d vars, expected %d", i, len(claims[i].Point), n)
		}
		if polys[i].NumVars() != n {
			return fmt.Errorf("sumcheck: poly %d has %d vars, expected %d", i, polys[i].NumVars(), n)
		}
	}
	return nil
}

func proveBatchedCore(claims []Claim, polys []MultiLin, lambda fext.Element, t Transcript) (BatchedProof, []fext.Element, error) {
	N := len(claims)
	n := len(claims[0].Point)

	lambdaPows := make([]fext.Element, N)
	lambdaPows[0].SetOne()
	for i := 1; i < N; i++ {
		lambdaPows[i].Mul(&lambdaPows[i-1], &lambda)
	}

	// Working tables: poly[i] starts as a clone, eq[i] is built from r_i.
	// Clone and EqTable construction are independent per i — parallelize.
	polyTables := make([]MultiLin, N)
	eqTables := make([]MultiLin, N)
	nbWorkers := runtime.GOMAXPROCS(0)
	if nbWorkers > N {
		nbWorkers = N
	}
	chunk := (N + nbWorkers - 1) / nbWorkers
	var wg sync.WaitGroup
	wg.Add(nbWorkers)
	for w := 0; w < nbWorkers; w++ {
		start := w * chunk
		stop := start + chunk
		if stop > N {
			stop = N
		}
		go func(start, stop int) {
			defer wg.Done()
			for i := start; i < stop; i++ {
				polyTables[i] = polys[i].Clone()
				eqTables[i] = BuildEqTable(claims[i].Point)
			}
		}(start, stop)
	}
	wg.Wait()

	proof := BatchedProof{
		RoundPolys: make([][3]fext.Element, n),
	}

	// partials holds per-worker partial (g0,g1,g2) sums for the reduction step.
	type triple = [3]fext.Element
	partials := make([]triple, nbWorkers)
	challenges := make([]fext.Element, n)

	for k := 0; k < n; k++ {
		size := len(polyTables[0])
		mid := size / 2

		// Parallel accumulation of round-polynomial coefficients.
		wg.Add(nbWorkers)
		for w := 0; w < nbWorkers; w++ {
			start := w * chunk
			stop := start + chunk
			if stop > N {
				stop = N
			}
			go func(w, start, stop int) {
				defer wg.Done()
				var g0, g1, g2, t1, e2, p2 fext.Element
				for i := start; i < stop; i++ {
					lp := lambdaPows[i]
					pt := polyTables[i]
					et := eqTables[i]
					for x := 0; x < mid; x++ {
						p0 := pt[x]
						p1 := pt[x+mid]
						e0 := et[x]
						e1 := et[x+mid]

						t1.Mul(&e0, &p0)
						t1.Mul(&t1, &lp)
						g0.Add(&g0, &t1)

						t1.Mul(&e1, &p1)
						t1.Mul(&t1, &lp)
						g1.Add(&g1, &t1)

						e2.Double(&e1)
						e2.Sub(&e2, &e0)
						p2.Double(&p1)
						p2.Sub(&p2, &p0)
						t1.Mul(&e2, &p2)
						t1.Mul(&t1, &lp)
						g2.Add(&g2, &t1)
					}
				}
				partials[w] = triple{g0, g1, g2}
			}(w, start, stop)
		}
		wg.Wait()

		// Serial reduction of worker partials.
		var g0, g1, g2 fext.Element
		for w := 0; w < nbWorkers; w++ {
			g0.Add(&g0, &partials[w][0])
			g1.Add(&g1, &partials[w][1])
			g2.Add(&g2, &partials[w][2])
		}

		proof.RoundPolys[k] = [3]fext.Element{g0, g1, g2}
		t.Append(labelRoundPoly, g0, g1, g2)
		c := t.Challenge(labelRoundChallenge)
		challenges[k] = c

		// Parallel fold.
		wg.Add(nbWorkers)
		for w := 0; w < nbWorkers; w++ {
			start := w * chunk
			stop := start + chunk
			if stop > N {
				stop = N
			}
			go func(start, stop int) {
				defer wg.Done()
				for i := start; i < stop; i++ {
					polyTables[i].Fold(c)
					eqTables[i].Fold(c)
				}
			}(start, stop)
		}
		wg.Wait()
	}

	proof.FinalEvals = make([]fext.Element, N)
	for i := 0; i < N; i++ {
		proof.FinalEvals[i] = polyTables[i][0]
	}
	t.Append(labelFinalEvals, proof.FinalEvals...)

	return proof, challenges, nil
}
