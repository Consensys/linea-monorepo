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
	proof, _, err := proveBatchedCore(claims, polys, false, lambda, t)
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
	return proveBatchedCore(claims, polys, false, lambda, t)
}

// ProveBatchedWithOwned is like [ProveBatchedWith] but takes ownership of the
// polys slices, folding them in-place without cloning. The caller must not use
// polys after this call. This saves N × 2^n fext.Element allocations (the
// polyTables clone), which is significant for large N (e.g. 5252 × 4096 = 328 MB).
func ProveBatchedWithOwned(claims []Claim, polys []MultiLin, lambda fext.Element, t Transcript) (BatchedProof, []fext.Element, error) {
	if err := validateBatchedInputs(claims, polys); err != nil {
		return BatchedProof{}, nil, err
	}
	return proveBatchedCore(claims, polys, true, lambda, t)
}

// ProveBatchedWithOwnedMixed is like [ProveBatchedWithOwned] but supports
// polynomials with fewer than nmax variables (compact representation).
// polyVars[i] is the actual number of variables for poly i (1 ≤ polyVars[i] ≤ n).
// polys[i] must have size 2^polyVars[i]; claims[i].Point must have length n
// (zero-padded high coordinates for compact polys). Ownership of polys is taken.
//
// The round polynomials produced are identical to the expanded representation
// (embed compact poly by repetition): sum_{xhigh ∈ {0,1}^(n-nq)} (1-x_j) = 1,
// so the extra dimensions cancel and each compact poly contributes the same g_k
// as its fully-expanded counterpart at 2^n entries.
func ProveBatchedWithOwnedMixed(claims []Claim, polys []MultiLin, lambda fext.Element, t Transcript, polyVars []int) (BatchedProof, []fext.Element, error) {
	if err := validateMixedInputs(claims, polys, polyVars); err != nil {
		return BatchedProof{}, nil, err
	}
	return proveBatchedCoreMixed(claims, polys, lambda, t, polyVars)
}

func validateMixedInputs(claims []Claim, polys []MultiLin, polyVars []int) error {
	N := len(claims)
	if N == 0 {
		return fmt.Errorf("sumcheck: no claims")
	}
	if len(polys) != N || len(polyVars) != N {
		return fmt.Errorf("sumcheck: %d polys, %d polyVars for %d claims", len(polys), len(polyVars), N)
	}
	n := len(claims[0].Point)
	if n == 0 {
		return fmt.Errorf("sumcheck: 0-variable claims are not supported")
	}
	for i := range claims {
		if len(claims[i].Point) != n {
			return fmt.Errorf("sumcheck: claim %d point length %d != %d", i, len(claims[i].Point), n)
		}
		nq := polyVars[i]
		if nq < 1 || nq > n {
			return fmt.Errorf("sumcheck: polyVars[%d]=%d out of [1,%d]", i, nq, n)
		}
		if polys[i].NumVars() != nq {
			return fmt.Errorf("sumcheck: poly %d has %d vars, expected polyVars=%d", i, polys[i].NumVars(), nq)
		}
	}
	return nil
}

// proveBatchedCoreMixed runs the batched sumcheck with per-poly compact tables.
// Each poly i has polyVars[i] real variables; its table has size 2^polyVars[i].
// For rounds k ≥ polyVars[i] the poly is trivially constant (scalar), contributing
// g_k(X) = poly_scalar * eq_scalar * (1-X) to the round polynomial.
func proveBatchedCoreMixed(claims []Claim, polys []MultiLin, lambda fext.Element, t Transcript, polyVars []int) (BatchedProof, []fext.Element, error) {
	N := len(claims)
	n := len(claims[0].Point) // nmax

	lambdaPows := make([]fext.Element, N)
	lambdaPows[0].SetOne()
	for i := 1; i < N; i++ {
		lambdaPows[i].Mul(&lambdaPows[i-1], &lambda)
	}

	// Build compact poly and eq tables. Eq table for poly i uses only the
	// first polyVars[i] coordinates of the claim point (high coords are zero,
	// contributing a factor of 1 to the sum which cancels perfectly).
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
				polyTables[i] = polys[i]
				eqTables[i] = BuildEqTable(claims[i].Point[:polyVars[i]])
			}
		}(start, stop)
	}
	wg.Wait()

	proof := BatchedProof{
		RoundPolys: make([][3]fext.Element, n),
	}

	type triple = [3]fext.Element
	partials := make([]triple, nbWorkers)
	challenges := make([]fext.Element, n)

	for k := 0; k < n; k++ {
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
					myMid := len(pt) / 2
					if myMid == 0 {
						// Trivial round: poly and eq are both scalars.
						// Contribution: g(X) = pt[0]*et[0]*(1-X), so
						// g(0)=pt[0]*et[0], g(1)=0, g(2)=-pt[0]*et[0].
						t1.Mul(&pt[0], &et[0])
						t1.Mul(&t1, &lp)
						g0.Add(&g0, &t1)
						g2.Sub(&g2, &t1)
					} else {
						for x := 0; x < myMid; x++ {
							p0 := pt[x]
							p1 := pt[x+myMid]
							e0 := et[x]
							e1 := et[x+myMid]

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
				}
				partials[w] = triple{g0, g1, g2}
			}(w, start, stop)
		}
		wg.Wait()

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

		// Parallel fold: for non-trivial polys, halve the tables; for trivial
		// (scalar) polys, only scale the eq by (1-c) — the poly scalar is
		// invariant to the high-variable folds.
		var oneMinusC fext.Element
		oneMinusC.SetOne()
		oneMinusC.Sub(&oneMinusC, &c)

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
					if len(polyTables[i]) > 1 {
						polyTables[i].Fold(c)
						eqTables[i].Fold(c)
					} else {
						eqTables[i][0].Mul(&eqTables[i][0], &oneMinusC)
					}
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

// ownPolys: if true, polys are folded in-place (no clone); caller must not use them after.
func proveBatchedCore(claims []Claim, polys []MultiLin, ownPolys bool, lambda fext.Element, t Transcript) (BatchedProof, []fext.Element, error) {
	N := len(claims)
	n := len(claims[0].Point)

	lambdaPows := make([]fext.Element, N)
	lambdaPows[0].SetOne()
	for i := 1; i < N; i++ {
		lambdaPows[i].Mul(&lambdaPows[i-1], &lambda)
	}

	// Working tables: poly[i] starts as a clone (or in-place if ownPolys), eq[i] from r_i.
	// Construction is independent per i — parallelize.
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
				if ownPolys {
					polyTables[i] = polys[i]
				} else {
					polyTables[i] = polys[i].Clone()
				}
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
