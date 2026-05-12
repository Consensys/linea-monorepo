package multilineareval

// combined.go implements Batch — the cross-size symbolic-expansion
// batcher. All MultilinearEval queries that share the same wizard round are
// folded into ONE sumcheck regardless of their numVars. A polynomial with
// numVars=nᵢ < nmax is embedded into the nmax-variable space by value
// repetition: E(f)[j] = f[j mod 2^nᵢ]. This leaves its multilinear extension
// unchanged (E_MLE(f)(x) = f_MLE(x₁,…,x_{nᵢ}) for any field point x), so the
// claim f(pᵢ) = yᵢ transfers without modification. High coordinates of pᵢ are
// zero-padded; the sumcheck library handles this as normal EQ factors of the
// form (1 − challenges[j]).
//
// After the combined sumcheck the residual for query qᵢ is a MultilinearEval
// at the FIRST nᵢ coordinates of the shared nmax-challenge, so downstream
// multilinvortex passes see queries grouped by numVars exactly as before.

import (
	"fmt"
	"sort"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/sumcheck"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
)

// CombinedContext holds the compiled artifacts for one cross-size combined
// sumcheck. One context is created per wizard round that contains unignored
// MultilinearEval queries.
type CombinedContext struct {
	// InputQueries are all MultilinearEval queries in this batch (mixed numVars).
	InputQueries []query.MultilinearEval
	// MaxNumVars is nmax = max(q.NumVars for q in InputQueries).
	MaxNumVars int
	// Round is the wizard round of all InputQueries.
	Round int
	// LambdaCoin is the single batching coin for the combined sumcheck.
	LambdaCoin coin.Info
	// RoundPolys holds the MaxNumVars round polynomials (3 evaluations each).
	RoundPolys ifaces.Column
	// Residuals[i] is the output query for InputQueries[i].
	// Residuals[i].NumVars == InputQueries[i].NumVars; its evaluation point is
	// the first InputQueries[i].NumVars coordinates of the shared challenge.
	Residuals []query.MultilinearEval
}

func buildCombinedContext(comp *wizard.CompiledIOP, round, nmax int, queries []query.MultilinearEval) *CombinedContext {
	suffix := fmt.Sprintf("nmax%d_r%d_%d", nmax, round, comp.SelfRecursionCount)

	lambdaCoin := comp.InsertCoin(
		round+1,
		coin.Name(fmt.Sprintf("MLEVAL_CMB_LAMBDA_%s", suffix)),
		coin.FieldExt,
	)

	roundPolysCol := comp.InsertProof(
		round+1,
		ifaces.ColID(fmt.Sprintf("MLEVAL_CMB_ROUNDPOLYS_%s", suffix)),
		nextPow2(nmax*3),
		false,
	)

	// "mlvortex_shared_safe_queries" is the multilinvortex ExtraData key for
	// queries whose per-pol points all share the same cCol suffix at prover
	// time. After the combined sumcheck, every residual point is challenges[:nq],
	// which is IDENTICAL across all polys of a single query — so if the input
	// query was shared-safe, the residual is trivially shared-safe too (with
	// the stronger property that ALL per-pol points are equal). Propagate the
	// flag so downstream multilinvortex.CompileRound keeps using the
	// SharedRowEvals fast path.
	const sharedSafeKey = "mlvortex_shared_safe_queries"
	safe, _ := comp.ExtraData[sharedSafeKey].(map[ifaces.QueryID]bool)

	residuals := make([]query.MultilinearEval, len(queries))
	for qIdx, q := range queries {
		residuals[qIdx] = comp.InsertMultilinear(
			round+1,
			ifaces.QueryID(fmt.Sprintf("MLEVAL_CMB_RESIDUAL_q%d_%s", qIdx, suffix)),
			q.Pols,
		)
		if safe[q.Name()] {
			if comp.ExtraData == nil {
				comp.ExtraData = make(map[string]any)
			}
			if safe == nil {
				safe = make(map[ifaces.QueryID]bool)
				comp.ExtraData[sharedSafeKey] = safe
			}
			safe[residuals[qIdx].Name()] = true
		}
	}

	return &CombinedContext{
		InputQueries: queries,
		MaxNumVars:   nmax,
		Round:        round,
		LambdaCoin:   lambdaCoin,
		RoundPolys:   roundPolysCol,
		Residuals:    residuals,
	}
}

// Batch batches ALL unignored MultilinearEval queries sharing the
// same wizard round into a single combined sumcheck, regardless of numVars.
// This is the "symbolic expansion" strategy: smaller polynomials are embedded
// into the nmax-variable space so they participate in one shared sumcheck. The
// sumcheck cost is proportional to Σᵢ 2^nᵢ (not 2^nmax), because the expanded
// polynomials have trivial contributions for the high variables (the EQ factor
// (1−challenges[j]) makes those rounds cheap). After the combined sumcheck each
// query's residual is at the first nᵢ coordinates of the shared challenge.
func Batch(comp *wizard.CompiledIOP) {
	type grp struct {
		queries []query.MultilinearEval
		nmax    int
	}
	grps := map[int]*grp{}
	var rounds []int

	for _, name := range comp.QueriesParams.AllUnignoredKeys() {
		q, ok := comp.QueriesParams.Data(name).(query.MultilinearEval)
		if !ok {
			continue
		}
		comp.QueriesParams.MarkAsIgnored(name)
		r := comp.QueriesParams.Round(name)
		if grps[r] == nil {
			grps[r] = &grp{}
			rounds = append(rounds, r)
		}
		g := grps[r]
		g.queries = append(g.queries, q)
		for _, nv := range q.NumVars {
			if nv > g.nmax {
				g.nmax = nv
			}
		}
	}

	if len(rounds) == 0 {
		return
	}
	sort.Ints(rounds)

	for _, r := range rounds {
		g := grps[r]
		ctx := buildCombinedContext(comp, r, g.nmax, g.queries)
		comp.RegisterProverAction(r+1, &CombinedProverAction{Ctx: ctx})
		comp.RegisterVerifierAction(comp.NumRounds()-1, &CombinedVerifierAction{Ctx: ctx})
	}
}

// BatchIgnored is like [Batch] but processes currently-IGNORED
// MultilinearEval queries instead of unignored ones. Call this in PostDistribute
// (after DistributeWizard) to pick up queries that were pre-marked as ignored so
// that FilterCompiledIOP did not see them.
func BatchIgnored(comp *wizard.CompiledIOP) {
	type grp struct {
		queries []query.MultilinearEval
		nmax    int
	}
	grps := map[int]*grp{}
	var rounds []int

	for _, name := range comp.QueriesParams.AllKeys() {
		if !comp.QueriesParams.IsIgnored(name) {
			continue
		}
		q, ok := comp.QueriesParams.Data(name).(query.MultilinearEval)
		if !ok {
			continue
		}
		r := comp.QueriesParams.Round(name)
		if grps[r] == nil {
			grps[r] = &grp{}
			rounds = append(rounds, r)
		}
		g := grps[r]
		g.queries = append(g.queries, q)
		for _, nv := range q.NumVars {
			if nv > g.nmax {
				g.nmax = nv
			}
		}
	}

	if len(rounds) == 0 {
		return
	}
	sort.Ints(rounds)

	for _, r := range rounds {
		g := grps[r]
		ctx := buildCombinedContext(comp, r, g.nmax, g.queries)
		comp.RegisterProverAction(r+1, &CombinedProverAction{Ctx: ctx})
		comp.RegisterVerifierAction(comp.NumRounds()-1, &CombinedVerifierAction{Ctx: ctx})
	}
}

// CombinedProverAction runs the prover side of the combined sumcheck.
type CombinedProverAction struct {
	Ctx *CombinedContext
}

func (p *CombinedProverAction) Run(run *wizard.ProverRuntime) {
	ctx := p.Ctx
	nmax := ctx.MaxNumVars

	lambda := run.GetRandomCoinFieldExt(ctx.LambdaCoin.Name)

	// Collect (col, nq, point, eval) tuples.
	type colEntry struct {
		col   ifaces.Column
		nq    int
		point []fext.Element // length nmax, zero-padded for nq < nmax
		eval  fext.Element
	}
	var entries []colEntry

	for _, q := range ctx.InputQueries {
		params := run.GetMultilinearParams(q.Name())
		for j, col := range q.Pols {
			nq := q.NumVars[j]
			extPoint := make([]fext.Element, nmax)
			copy(extPoint, params.Points[j])
			entries = append(entries, colEntry{
				col:   col,
				nq:    nq,
				point: extPoint,
				eval:  params.Ys[j],
			})
		}
	}

	// Build compact polynomial tables in parallel. Each poly is kept at its
	// native size 2^nq instead of being expanded to 2^nmax. The sumcheck
	// round polynomials are identical to the expanded form because the extra
	// (high) variables all have r=0, so their eq factors sum to 1 over
	// {0,1}^(nmax-nq) and cancel out.
	claims := make([]sumcheck.Claim, len(entries))
	polys := make([]sumcheck.MultiLin, len(entries))
	polyVars := make([]int, len(entries))
	parallel.Execute(len(entries), func(start, stop int) {
		for i := start; i < stop; i++ {
			e := entries[i]
			origVec := run.GetColumn(e.col.GetColID()).IntoRegVecSaveAllocExt()
			// Copy to own the slice for in-place folding.
			compact := make([]fext.Element, len(origVec))
			copy(compact, origVec)
			claims[i] = sumcheck.Claim{Point: e.point, Eval: e.eval}
			polys[i] = sumcheck.MultiLin(compact)
			polyVars[i] = e.nq
		}
	})

	t := sumcheck.NewMockTranscript(transcriptSeed)
	t.Append("lambda", lambda)

	// Compact polys are owned — pass to mixed prover which folds in-place.
	proof, challenges, err := sumcheck.ProveBatchedWithOwnedMixed(claims, polys, lambda, t, polyVars)
	if err != nil {
		panic(fmt.Sprintf("multilineareval combinedProver: %v", err))
	}

	// Write round polynomials.
	colSize := ctx.RoundPolys.Size()
	flat := make([]fext.Element, colSize)
	for k := 0; k < nmax; k++ {
		flat[k*3+0] = proof.RoundPolys[k][0]
		flat[k*3+1] = proof.RoundPolys[k][1]
		flat[k*3+2] = proof.RoundPolys[k][2]
	}
	run.AssignColumn(ctx.RoundPolys.GetColID(), smartvectors.NewRegularExt(flat))

	// Assign per-query residuals. FinalEvals[i] = expanded_poly_i(challenges)
	// = original_poly_i(challenges[:nq]), so we truncate the challenge point per poly.
	evalIdx := 0
	for qIdx, q := range ctx.InputQueries {
		finalEvals := proof.FinalEvals[evalIdx : evalIdx+len(q.Pols)]
		residualPoints := make([][]fext.Element, len(q.Pols))
		for j := range q.Pols {
			nq := q.NumVars[j]
			residualPoints[j] = challenges[:nq]
		}
		run.AssignMultilinearExt(ctx.Residuals[qIdx].Name(), residualPoints, finalEvals...)
		evalIdx += len(q.Pols)
	}
}

// CombinedVerifierAction verifies the combined sumcheck.
type CombinedVerifierAction struct {
	Ctx     *CombinedContext
	Skipped bool `serde:"omit"`
}

func (v *CombinedVerifierAction) Run(run wizard.Runtime) error {
	ctx := v.Ctx
	nmax := ctx.MaxNumVars

	lambda := run.GetRandomCoinFieldExt(ctx.LambdaCoin.Name)

	// Build extended claims: all evaluation points padded to nmax with zeros.
	var claims []sumcheck.Claim
	for _, q := range ctx.InputQueries {
		params := run.GetMultilinearParams(q.Name())
		for j := range q.Pols {
			extPoint := make([]fext.Element, nmax)
			copy(extPoint, params.Points[j])
			claims = append(claims, sumcheck.Claim{Point: extPoint, Eval: params.Ys[j]})
		}
	}

	// Read round polynomials from the proof column.
	roundPolys := make([][3]fext.Element, nmax)
	for k := 0; k < nmax; k++ {
		roundPolys[k][0] = run.GetColumnAtExt(ctx.RoundPolys.GetColID(), k*3+0)
		roundPolys[k][1] = run.GetColumnAtExt(ctx.RoundPolys.GetColID(), k*3+1)
		roundPolys[k][2] = run.GetColumnAtExt(ctx.RoundPolys.GetColID(), k*3+2)
	}

	// Assemble FinalEvals from per-query residuals (in the same order as polys).
	var finalEvals []fext.Element
	for qIdx := range ctx.InputQueries {
		residualParams := run.GetMultilinearParams(ctx.Residuals[qIdx].Name())
		finalEvals = append(finalEvals, residualParams.Ys...)
	}

	proof := sumcheck.BatchedProof{RoundPolys: roundPolys, FinalEvals: finalEvals}

	t := sumcheck.NewMockTranscript(transcriptSeed)
	t.Append("lambda", lambda)

	challenges, _, err := sumcheck.VerifyBatchedWith(claims, proof, lambda, t)
	if err != nil {
		return fmt.Errorf("multilineareval combinedVerifier: %w", err)
	}

	// Confirm each residual's point matches the corresponding prefix of challenges.
	for qIdx, q := range ctx.InputQueries {
		residualParams := run.GetMultilinearParams(ctx.Residuals[qIdx].Name())
		for j := range q.Pols {
			nq := q.NumVars[j]
			if len(residualParams.Points[j]) != nq {
				return fmt.Errorf("multilineareval combinedVerifier: residual %d poly %d point length %d != %d",
					qIdx, j, len(residualParams.Points[j]), nq)
			}
			for i := 0; i < nq; i++ {
				if !residualParams.Points[j][i].Equal(&challenges[i]) {
					return fmt.Errorf("multilineareval combinedVerifier: residual %d poly %d point[%d] mismatch", qIdx, j, i)
				}
			}
		}
	}
	return nil
}

func (v *CombinedVerifierAction) RunGnark(_ frontend.API, _ wizard.GnarkRuntime) {
	panic("multilineareval.combinedVerifierAction.RunGnark: not yet implemented")
}

func (v *CombinedVerifierAction) Skip()           { v.Skipped = true }
func (v *CombinedVerifierAction) IsSkipped() bool { return v.Skipped }
