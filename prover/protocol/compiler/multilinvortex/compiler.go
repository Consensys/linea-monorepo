// Package multilinvortex implements one round of the multilinear Vortex
// commitment opening. It consumes MultilinearEval queries at a shared point
// c ∈ F^n and produces two new MultilinearEval queries per column:
//
//  1. MultilinearEval(U_α, c_col) — the α-combination of the matrix rows
//     evaluated at the column part of c.
//  2. MultilinearEval(RowEvals, c_row) = y — the row-evaluation polynomial
//     evaluated at the row part of c (this carries the original claim).
//
// A verifier action checks the consistency (Check 3) between U_α and RowEvals:
//
//	Σ_b α^b · RowEvals[b] == v   where v comes from the U_α MultilinearEval params.
//
// Check 1 (RS codeword) and Check 2 (Merkle spot-checks) are deferred to the
// subsequent Vortex compilation pass; the prover is trusted to commit honestly
// in the prototype.
//
// The output MultilinearEval queries feed into the next round of
// multilineareval.Compile and multilinvortex.Compile.
package multilinvortex

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// context holds the compiled artifacts for one batch of MultilinearEval queries.
type context struct {
	// InputQuery is the batched MultilinearEval at the shared point c.
	// After multilineareval.Compile, there is one residual per numVars group.
	InputQuery query.MultilinearEval
	// NumVars is n = n_row + n_col.
	NumVars int
	// NRow is the number of row variables (high bits of the evaluation point).
	NRow int
	// NCol is the number of column variables (low bits).
	NCol int
	// Round is the wizard round of the input query.
	Round int
	// AlphaCoin is the batching coin for the row combination.
	AlphaCoin coin.Info
	// UAlpha[k] is the committed column for the α-combination of column k's rows.
	UAlpha []ifaces.Column
	// RowEvals[k] is the committed column for the row evaluations of column k at c_col.
	RowEvals []ifaces.Column
	// UCols[k] is the MultilinearEval query: UAlpha[k] evaluated at c_col.
	UCols []query.MultilinearEval
	// RowClaims[k] is the MultilinearEval query: RowEvals[k] evaluated at c_row.
	RowClaims []query.MultilinearEval
}

// Compile applies one round of the multilinear Vortex opening protocol using
// the balanced split nRow = ⌈n/2⌉. See [CompileWithFold] for a configurable
// variant.
func Compile(comp *wizard.CompiledIOP) {
	compileWithNRow(comp, -1) // -1 signals "use default ⌈n/2⌉"
}

// CompileWithFold returns a compiler pass that uses nFoldRows row-variables per
// Vortex round instead of the default ⌈n/2⌉.
//
// Setting nFoldRows=1 (WHIR-style minimum) gives a 2-element RowEvals column
// that is immediately terminal (numVars=1), so the RowClaims path exits after
// one round for every polynomial regardless of its original size.  The cost is
// that UAlpha has 2^(n−1) elements instead of 2^⌈n/2⌉, so proof size grows
// unless the Vortex commitment layer shares a single Merkle tree across
// same-size UAlpha columns (future work).
//
// When paired with [multilineareval.CompileAllRound], nFoldRows=1 produces the
// WHIR-style early-exit recursion: every round, ALL polynomials (regardless of
// original size) share the same RowClaims terminal path, and UCols of different
// sizes are naturally re-batched by the next CompileAllRound call.
func CompileWithFold(nFoldRows int) func(*wizard.CompiledIOP) {
	return func(comp *wizard.CompiledIOP) {
		compileWithNRow(comp, nFoldRows)
	}
}

// mustAllSameNumVars returns the shared numVars for all polynomials in q.
// Panics if the polynomials have mixed sizes, since multilinvortex requires a
// uniform row/column split.
func mustAllSameNumVars(q query.MultilinearEval) int {
	if len(q.NumVars) == 0 {
		panic("multilinvortex: MultilinearEval has no polynomials")
	}
	n := q.NumVars[0]
	for k, nk := range q.NumVars {
		if nk != n {
			panic(fmt.Sprintf("multilinvortex: mixed numVars in query %v: poly 0 has %d, poly %d has %d",
				q.QueryID, n, k, nk))
		}
	}
	return n
}

// compileWithNRow is the shared implementation. nFoldRows < 1 means "use ⌈n/2⌉".
func compileWithNRow(comp *wizard.CompiledIOP, nFoldRows int) {
	idx := 0
	for _, name := range comp.QueriesParams.AllUnignoredKeys() {
		q, ok := comp.QueriesParams.Data(name).(query.MultilinearEval)
		if !ok {
			continue
		}
		comp.QueriesParams.MarkAsIgnored(name)
		r := comp.QueriesParams.Round(name)

		if mustAllSameNumVars(q) == 1 {
			for _, pol := range q.Pols {
				comp.Columns.SetStatus(pol.GetColID(), column.Proof)
			}
			comp.RegisterVerifierAction(comp.NumRounds()-1, &terminalVerifierAction{q: q})
			idx++
			continue
		}

		// Determine nRow: clamp nFoldRows to [1, n-1].
		n := mustAllSameNumVars(q)
		nRow := (n + 1) / 2 // default balanced
		if nFoldRows >= 1 {
			nRow = nFoldRows
			if nRow >= n {
				nRow = n - 1
			}
		}

		ctx := buildContext(comp, r, q, nRow, idx)
		comp.RegisterProverAction(r+1, &proverAction{ctx: ctx})
		comp.RegisterVerifierAction(comp.NumRounds()-1, &verifierAction{ctx: ctx})
		idx++
	}
}

func buildContext(comp *wizard.CompiledIOP, round int, q query.MultilinearEval, nRow, idx int) *context {
	n := mustAllSameNumVars(q)
	nCol := n - nRow
	K := len(q.Pols)

	suffix := fmt.Sprintf("n%d_r%d_%d_i%d", n, round, comp.SelfRecursionCount, idx)

	alpha := comp.InsertCoin(
		round+1,
		coin.Name(fmt.Sprintf("MLVORTEX_ALPHA_%s", suffix)),
		coin.FieldExt,
	)

	uAlpha := make([]ifaces.Column, K)
	rowEvals := make([]ifaces.Column, K)
	uCols := make([]query.MultilinearEval, K)
	rowClaims := make([]query.MultilinearEval, K)

	for k, pol := range q.Pols {
		uAlpha[k] = comp.InsertProof(
			round+1,
			ifaces.ColID(fmt.Sprintf("MLVORTEX_UALPHA_%d_%s", k, suffix)),
			1<<nCol,
			false,
		)
		rowEvals[k] = comp.InsertProof(
			round+1,
			ifaces.ColID(fmt.Sprintf("MLVORTEX_ROWEVAL_%d_%s", k, suffix)),
			1<<nRow,
			false,
		)

		uCols[k] = comp.InsertMultilinear(
			round+1,
			ifaces.QueryID(fmt.Sprintf("MLVORTEX_UCOL_%d_%s", k, suffix)),
			[]ifaces.Column{uAlpha[k]},
		)
		rowClaims[k] = comp.InsertMultilinear(
			round+1,
			ifaces.QueryID(fmt.Sprintf("MLVORTEX_ROW_%d_%s", k, suffix)),
			[]ifaces.Column{rowEvals[k]},
		)
		_ = pol // column already referenced via q.Pols[k]
	}

	return &context{
		InputQuery: q,
		NumVars:    n,
		NRow:       nRow,
		NCol:       nCol,
		Round:      round,
		AlphaCoin:  alpha,
		UAlpha:     uAlpha,
		RowEvals:   rowEvals,
		UCols:      uCols,
		RowClaims:  rowClaims,
	}
}
