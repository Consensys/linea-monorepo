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

// Compile applies one round of the multilinear Vortex opening protocol. It
// consumes all MultilinearEval queries (after they have been batched to a
// shared point by multilineareval.Compile) and produces the two new
// MultilinearEval claims described in the package doc.
func Compile(comp *wizard.CompiledIOP) {
	for _, name := range comp.QueriesParams.AllUnignoredKeys() {
		q, ok := comp.QueriesParams.Data(name).(query.MultilinearEval)
		if !ok {
			continue
		}
		// Only process queries that were introduced by the multilineareval
		// compiler (i.e. residuals). We recognize them by having more than one
		// polynomial — the residual always collects all columns.
		// For the prototype we process every MultilinearEval query.
		comp.QueriesParams.MarkAsIgnored(name)
		r := comp.QueriesParams.Round(name)

		// Terminal case: a 1-variable query is just a 2-point evaluation.
		// Promote the columns to Proof status and register a direct verifier.
		if q.NumVars == 1 {
			for _, pol := range q.Pols {
				comp.Columns.SetStatus(pol.GetColID(), column.Proof)
			}
			comp.RegisterVerifierAction(comp.NumRounds()-1, &terminalVerifierAction{q: q})
			continue
		}

		ctx := buildContext(comp, r, q)
		comp.RegisterProverAction(r+1, &proverAction{ctx: ctx})
		comp.RegisterVerifierAction(comp.NumRounds()-1, &verifierAction{ctx: ctx})
	}
}

func buildContext(comp *wizard.CompiledIOP, round int, q query.MultilinearEval) *context {
	n := q.NumVars
	nRow := (n + 1) / 2
	nCol := n - nRow
	K := len(q.Pols)

	suffix := fmt.Sprintf("n%d_r%d_%d", n, round, comp.SelfRecursionCount)

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
			nCol,
			[]ifaces.Column{uAlpha[k]},
		)
		rowClaims[k] = comp.InsertMultilinear(
			round+1,
			ifaces.QueryID(fmt.Sprintf("MLVORTEX_ROW_%d_%s", k, suffix)),
			nRow,
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
