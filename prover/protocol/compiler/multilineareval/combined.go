package multilineareval

// combined.go implements CompileAllRound — the cross-size symbolic-expansion
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
)

// combinedContext holds the compiled artifacts for one cross-size combined
// sumcheck. One context is created per wizard round that contains unignored
// MultilinearEval queries.
type combinedContext struct {
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

	skipped bool `serde:"omit"`
}

func buildCombinedContext(comp *wizard.CompiledIOP, round, nmax int, queries []query.MultilinearEval) *combinedContext {
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

	residuals := make([]query.MultilinearEval, len(queries))
	for qIdx, q := range queries {
		residuals[qIdx] = comp.InsertMultilinear(
			round+1,
			ifaces.QueryID(fmt.Sprintf("MLEVAL_CMB_RESIDUAL_q%d_%s", qIdx, suffix)),
			q.NumVars,
			q.Pols,
		)
	}

	return &combinedContext{
		InputQueries: queries,
		MaxNumVars:   nmax,
		Round:        round,
		LambdaCoin:   lambdaCoin,
		RoundPolys:   roundPolysCol,
		Residuals:    residuals,
	}
}

// CompileAllRound batches ALL unignored MultilinearEval queries sharing the
// same wizard round into a single combined sumcheck, regardless of numVars.
// This is the "symbolic expansion" strategy: smaller polynomials are embedded
// into the nmax-variable space so they participate in one shared sumcheck. The
// sumcheck cost is proportional to Σᵢ 2^nᵢ (not 2^nmax), because the expanded
// polynomials have trivial contributions for the high variables (the EQ factor
// (1−challenges[j]) makes those rounds cheap). After the combined sumcheck each
// query's residual is at the first nᵢ coordinates of the shared challenge.
func CompileAllRound(comp *wizard.CompiledIOP) {
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
		if q.NumVars > g.nmax {
			g.nmax = q.NumVars
		}
	}

	if len(rounds) == 0 {
		return
	}
	sort.Ints(rounds)

	for _, r := range rounds {
		g := grps[r]
		ctx := buildCombinedContext(comp, r, g.nmax, g.queries)
		comp.RegisterProverAction(r+1, &combinedProverAction{ctx: ctx})
		comp.RegisterVerifierAction(comp.NumRounds()-1, &combinedVerifierAction{ctx: ctx})
	}
}

// CompileAllRoundIgnored is like [CompileAllRound] but processes currently-IGNORED
// MultilinearEval queries instead of unignored ones. Call this in PostDistribute
// (after DistributeWizard) to pick up queries that were pre-marked as ignored so
// that FilterCompiledIOP did not see them.
func CompileAllRoundIgnored(comp *wizard.CompiledIOP) {
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
		if q.NumVars > g.nmax {
			g.nmax = q.NumVars
		}
	}

	if len(rounds) == 0 {
		return
	}
	sort.Ints(rounds)

	for _, r := range rounds {
		g := grps[r]
		ctx := buildCombinedContext(comp, r, g.nmax, g.queries)
		comp.RegisterProverAction(r+1, &combinedProverAction{ctx: ctx})
		comp.RegisterVerifierAction(comp.NumRounds()-1, &combinedVerifierAction{ctx: ctx})
	}
}

// combinedProverAction runs the prover side of the combined sumcheck.
type combinedProverAction struct {
	ctx *combinedContext
}

func (p *combinedProverAction) Run(run *wizard.ProverRuntime) {
	ctx := p.ctx
	nmax := ctx.MaxNumVars

	lambda := run.GetRandomCoinFieldExt(ctx.LambdaCoin.Name)

	var claims []sumcheck.Claim
	var polys []sumcheck.MultiLin

	for _, q := range ctx.InputQueries {
		params := run.GetMultilinearParams(q.Name())
		nq := q.NumVars
		shift := nmax - nq

		// Extend evaluation point to nmax with zero coordinates.
		extPoint := make([]fext.Element, nmax)
		copy(extPoint, params.Point)
		for j, col := range q.Pols {
			// Expand column to 2^nmax: E(f)[i] = f[i >> (nmax-nq)].
			// In MSB-first convention this replicates each f entry across all
			// combinations of the LOW (nmax-nq) variables, so
			// E(f)_MLE(x₁,...,x_{nmax}) = f_MLE(x₁,...,x_{nq}) for any point.
			// The claim transfers unchanged since extPoint high coords are zero.
			origVec := run.GetColumn(col.GetColID()).IntoRegVecSaveAllocExt()
			expanded := make([]fext.Element, 1<<nmax)
			for i := range expanded {
				expanded[i] = origVec[i>>shift]
			}
			claims = append(claims, sumcheck.Claim{Point: extPoint, Eval: params.Ys[j]})
			polys = append(polys, sumcheck.MultiLin(expanded))
		}
	}

	t := sumcheck.NewMockTranscript(transcriptSeed)
	t.Append("lambda", lambda)

	proof, challenges, err := sumcheck.ProveBatchedWith(claims, polys, lambda, t)
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
	// = original_poly_i(challenges[:nq]), so we truncate the challenge point.
	evalIdx := 0
	for qIdx, q := range ctx.InputQueries {
		nq := q.NumVars
		finalEvals := proof.FinalEvals[evalIdx : evalIdx+len(q.Pols)]
		run.AssignMultilinearExt(ctx.Residuals[qIdx].Name(), challenges[:nq], finalEvals...)
		evalIdx += len(q.Pols)
	}
}

// combinedVerifierAction verifies the combined sumcheck.
type combinedVerifierAction struct {
	ctx     *combinedContext
	skipped bool `serde:"omit"`
}

func (v *combinedVerifierAction) Run(run wizard.Runtime) error {
	ctx := v.ctx
	nmax := ctx.MaxNumVars

	lambda := run.GetRandomCoinFieldExt(ctx.LambdaCoin.Name)

	// Build extended claims: all evaluation points padded to nmax with zeros.
	var claims []sumcheck.Claim
	for _, q := range ctx.InputQueries {
		params := run.GetMultilinearParams(q.Name())
		extPoint := make([]fext.Element, nmax)
		copy(extPoint, params.Point)
		for j := range q.Pols {
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
		nq := q.NumVars
		residualParams := run.GetMultilinearParams(ctx.Residuals[qIdx].Name())
		if len(residualParams.Point) != nq {
			return fmt.Errorf("multilineareval combinedVerifier: residual %d point length %d != %d",
				qIdx, len(residualParams.Point), nq)
		}
		for i := 0; i < nq; i++ {
			if !residualParams.Point[i].Equal(&challenges[i]) {
				return fmt.Errorf("multilineareval combinedVerifier: residual %d point[%d] mismatch", qIdx, i)
			}
		}
	}
	return nil
}

func (v *combinedVerifierAction) RunGnark(_ frontend.API, _ wizard.GnarkRuntime) {
	panic("multilineareval.combinedVerifierAction.RunGnark: not yet implemented")
}

func (v *combinedVerifierAction) Skip()           { v.skipped = true }
func (v *combinedVerifierAction) IsSkipped() bool { return v.skipped }
