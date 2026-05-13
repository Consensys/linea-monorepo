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
// Check 2 (Merkle spot-checks on UAlpha) is handled by CommitMLColumns.
// Check 1 (UAlpha = α-combination of committed original data) and the
// corresponding original-column Merkle commitment are handled by
// CommitOriginalMLColumns.
//
// The output MultilinearEval queries feed into the next round of
// multilineareval.Compile and multilinvortex.Compile.
package multilinvortex

import (
	"fmt"
	"math/bits"

	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// Context holds the compiled artifacts for one batch of MultilinearEval queries.
type Context struct {
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
	// When Packed=true the slice has length 1: a single column of size KPow2·2^NCol
	// storing all K UAlpha vectors end-to-end (block k at [k·nColSize:(k+1)·nColSize]).
	UAlpha []ifaces.Column
	// RowEvals[k] is the committed column for the row evaluations of column k at c_col.
	// When Packed=true the slice has length 1 (same layout as UAlpha).
	RowEvals []ifaces.Column
	// UCols[k] is the MultilinearEval query for UAlpha[k] evaluated at c_col.
	// When Packed=true each UCols[k] references the single packed UAlpha column at
	// the locator-extended point (l_k ‖ c_col) with L+NCol variables.
	UCols []query.MultilinearEval
	// RowClaims[k] is the MultilinearEval query for RowEvals[k] evaluated at c_row.
	// When Packed=true each RowClaims[k] references the packed RowEvals column at
	// (l_k ‖ c_row) with L+NRow variables.
	RowClaims []query.MultilinearEval

	// Packed indicates that all K UAlpha/RowEvals columns are packed into a single
	// column per type using the locator-tuple embedding P(l_k, x) = f_k(x).
	// CommitMLColumns builds ONE Merkle tree per packed column instead of K trees.
	Packed bool
	// L = ceil(log2(K)) when Packed; number of locator bits prepended to each
	// evaluation point. 0 when Packed=false.
	L int
	// KPow2 = 2^L when Packed (K rounded up to the next power of two).
	KPow2 int

	// SharedInput indicates that all K input polynomials reference the SAME
	// committed column (e.g. K duplicate Q polys from
	// InsertBootstrapperOpeningsPacked). In this mode UAlpha has length 1
	// (UAlpha is independent of the per-pol cCol), while RowEvals may still
	// have length K when per-pol cCol slices can differ.
	SharedInput bool

	// SharedRowEvals strengthens SharedInput with a per-pol cCol-shared
	// guarantee: cCol_k is identical for every k, so RowEvals also collapses
	// to a single column. In this mode the Context has UAlpha=[1], RowEvals=[1],
	// UCols=[1] (single claim at the shared cCol), and RowClaims=[1] holding
	// a SINGLE MultilinearEval query with K duplicate-RowEvals polys at K
	// different cRow_k points. Downstream Batch sees one input query, emits
	// one residual query, and the propagated structure remains shared.
	SharedRowEvals bool
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
// When paired with [multilineareval.Batch], nFoldRows=1 produces the
// WHIR-style early-exit recursion: every round, ALL polynomials (regardless of
// original size) share the same RowClaims terminal path, and UCols of different
// sizes are naturally re-batched by the next Batch call.
func CompileWithFold(nFoldRows int) func(*wizard.CompiledIOP) {
	return func(comp *wizard.CompiledIOP) {
		compileWithNRow(comp, nFoldRows)
	}
}

// CompileRound is a convenience wrapper that runs one complete ML Vortex round:
// Compile + CommitMLColumns + CommitOriginalMLColumns.
// It is safe to call multiple times; CommitOriginalMLColumns is idempotent and
// becomes a no-op after the first round.
func CompileRound(comp *wizard.CompiledIOP) {
	Compile(comp)
	CommitMLColumns(comp)
	CommitOriginalMLColumns(comp)
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

func hasCommittedInputs(comp *wizard.CompiledIOP, q query.MultilinearEval) bool {
	if len(q.Pols) == 0 {
		return false
	}
	hasCommitted := comp.Columns.Status(q.Pols[0].GetColID()) == column.Committed
	for _, pol := range q.Pols[1:] {
		if (comp.Columns.Status(pol.GetColID()) == column.Committed) != hasCommitted {
			panic(fmt.Sprintf("multilinvortex: mixed committed/non-committed inputs in query %v", q.QueryID))
		}
	}
	return hasCommitted
}

func terminalShortcutAllowed(comp *wizard.CompiledIOP, q query.MultilinearEval, numVars int) bool {
	if numVars > 1 {
		return false
	}
	if !hasCommittedInputs(comp, q) {
		return true
	}
	if numVars == 0 {
		panic(fmt.Sprintf("multilinvortex: committed constant query %v requires an explicit opening path", q.QueryID))
	}
	return false
}

// chooseNRow picks the row half of the multilinear-variable split so that the
// COMMITTED codeword matrix is as close to square as possible.
//
// CommitOriginalMLColumns stacks K_orig polys vertically, each viewed as
// (2^nRow × 2^nCol). The codeword (after RS blowup b) has:
//
//	codeword_rows = K_orig · 2^nRow
//	codeword_cols = b · 2^nCol
//
// With nRow + nCol = nv and b = vortexBlowup, "square" requires
//
//	log2(K_orig) + nRow = log2(b) + nCol
//	2·nRow = nv + log2(b) - log2(K_orig)
//	nRow  = (nv + log2(b) - log2(K_orig)) / 2
//
// When K_orig = 1 (no stacking) the formula reduces to ⌈(nv + log2(b))/2⌉,
// which for b = 2 gives the legacy ⌈nv/2⌉+ε. Larger K_orig shifts nRow DOWN
// (more cols, fewer rows) to undo the K-fold row inflation.
//
// nFoldRows ≥ 1 overrides the balanced value (used by CompileWithFold).
//
// nRow is clamped to [1, numVars-1] so that both halves remain non-degenerate.
func chooseNRow(numVars, kOrig, nFoldRows int) int {
	if numVars <= 1 {
		return 1
	}
	if nFoldRows >= 1 {
		nRow := nFoldRows
		if nRow >= numVars {
			nRow = numVars - 1
		}
		return nRow
	}
	if kOrig < 1 {
		kOrig = 1
	}
	// logK = ⌈log₂(kOrig)⌉; bits.Len(0)=0 handles kOrig=1 cleanly.
	logK := 0
	if kOrig > 1 {
		logK = bits.Len(uint(kOrig - 1))
	}
	// vortexBlowup is 2 throughout this package (see CommitMerkleWithSIS callers).
	const logBlowup = 1
	// Solve 2·nRow = nv + logBlowup − logK for nRow (integer-div rounds DOWN
	// when the exact balance is non-integer; this matches the legacy
	// ⌈nv/2⌉ behaviour for K = 1 and keeps the row count from inflating
	// the SIS leaf size).
	nRow := (numVars + logBlowup - logK) / 2
	if nRow < 1 {
		nRow = 1
	}
	if nRow >= numVars {
		nRow = numVars - 1
	}
	return nRow
}

// compileWithNRow is the shared implementation. nFoldRows < 1 means "use ⌈n/2⌉".
func compileWithNRow(comp *wizard.CompiledIOP, nFoldRows int) {
	// Collect ProverActions per round so independent same-round actions can be
	// batched into one parallel execution, reducing sequential overhead.
	pendingProver := make(map[int][]*ProverAction)

	idx := 0
	for _, name := range comp.QueriesParams.AllUnignoredKeys() {
		q, ok := comp.QueriesParams.Data(name).(query.MultilinearEval)
		if !ok {
			continue
		}
		comp.QueriesParams.MarkAsIgnored(name)
		r := comp.QueriesParams.Round(name)

		n := mustAllSameNumVars(q)
		if terminalShortcutAllowed(comp, q, n) {
			comp.RegisterVerifierAction(comp.NumRounds()-1, &TerminalVerifierAction{Q: q})
			idx++
			continue
		}

		// kOrig is the number of UNIQUE input columns — what actually gets
		// stacked in the orig commit matrix. Shared-input contexts collapse
		// K duplicate Q polys to kOrig = 1; legacy contexts use K as-is.
		kOrig := len(q.Pols)
		if allSameInput(q.Pols) {
			kOrig = 1
		}
		nRow := chooseNRow(n, kOrig, nFoldRows)
		var ctx *Context
		if allSameInput(q.Pols) {
			if isSharedSafeQuery(comp, q.Name()) {
				ctx = buildContextSharedRowEvals(comp, r, q, nRow, idx)
			} else {
				ctx = buildContextSharedInput(comp, r, q, nRow, idx)
			}
		} else {
			ctx = buildContext(comp, r, q, nRow, idx)
		}
		pendingProver[r+1] = append(pendingProver[r+1], &ProverAction{Ctx: ctx})
		comp.RegisterVerifierAction(comp.NumRounds()-1, &VerifierAction{Ctx: ctx})

		// Store context so later passes (e.g. CommitOriginalMLColumns) can read it.
		if comp.ExtraData == nil {
			comp.ExtraData = make(map[string]any)
		}
		ctxs, _ := comp.ExtraData[mlvortexContextsKey].([]*Context)
		comp.ExtraData[mlvortexContextsKey] = append(ctxs, ctx)

		idx++
	}

	for round, actions := range pendingProver {
		if len(actions) == 1 {
			comp.RegisterProverAction(round, actions[0])
		} else {
			comp.RegisterProverAction(round, &proverActionBatch{actions: actions})
		}
	}
}

// proverActionBatch runs multiple independent ProverAction instances concurrently.
// Each action reads from different input columns and writes to different output
// columns, so there is no shared mutable state beyond the mutex-protected runtime
// calls (GetColumn, AssignColumn, GetRandomCoinFieldExt, etc.).
type proverActionBatch struct {
	actions []*ProverAction
}

func (b *proverActionBatch) Run(run *wizard.ProverRuntime) {
	// Run sequentially: each ProverAction already saturates all cores via
	// parallel.Execute, so concurrent execution only over-subscribes the CPU
	// and hurts throughput.
	for _, a := range b.actions {
		a.Run(run)
	}
}

// mlvortexContextsKey indexes the slice of *Context values stored in ExtraData
// by compileWithNRow, for use by CommitOriginalMLColumns.
const mlvortexContextsKey = "mlvortex_contexts"

// sharedSafeQueriesKey indexes a map[ifaces.QueryID]bool stored in ExtraData.
// A query in the set is guaranteed (by its producer) to have ALL per-pol
// evaluation points sharing the same cCol suffix at prover time. This lets
// buildContextSharedRowEvals collapse RowEvals to ONE proof column instead of K.
const sharedSafeQueriesKey = "mlvortex_shared_safe_queries"

func buildContext(comp *wizard.CompiledIOP, round int, q query.MultilinearEval, nRow, idx int) *Context {
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

	return &Context{
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

// isSharedSafeQuery reports whether the producer of q has marked it as
// guaranteed-shared-cCol via sharedSafeQueriesKey. This is the contract that
// lets buildContextSharedRowEvals emit a single RowEvals proof column.
func isSharedSafeQuery(comp *wizard.CompiledIOP, name ifaces.QueryID) bool {
	if comp.ExtraData == nil {
		return false
	}
	set, _ := comp.ExtraData[sharedSafeQueriesKey].(map[ifaces.QueryID]bool)
	return set[name]
}

// markSharedSafeQuery records that q's per-pol points all share the same
// cCol suffix at prover time. Used by buildContextSharedRowEvals to propagate
// the guarantee onto the downstream queries it emits (so subsequent
// CompileRound passes can keep using the SharedRowEvals fast path).
func markSharedSafeQuery(comp *wizard.CompiledIOP, name ifaces.QueryID) {
	if comp.ExtraData == nil {
		comp.ExtraData = make(map[string]any)
	}
	set, _ := comp.ExtraData[sharedSafeQueriesKey].(map[ifaces.QueryID]bool)
	if set == nil {
		set = make(map[ifaces.QueryID]bool)
		comp.ExtraData[sharedSafeQueriesKey] = set
	}
	set[name] = true
}

// buildContextSharedRowEvals is the strongest shared-input mode. It assumes
// the caller guarantees that every per-pol point shares the same cCol suffix,
// so both UAlpha AND RowEvals collapse to a SINGLE proof column. The K
// downstream claims are encoded as exactly TWO queries:
//
//   - One UCols query on the shared UAlpha at the shared cCol (single claim).
//   - One RowClaims query on the shared RowEvals with K duplicate polys at
//     K different cRow_k points (the per-pol claim that still varies with k).
//
// Both are marked SharedSafe so the next CompileRound (after Batch) finds
// the same structure: 1 residual on UAlpha, 1 residual on RowEvals with K
// duplicate polys at K identical points (Batch puts every poly at
// challenges[:nq]).
func buildContextSharedRowEvals(comp *wizard.CompiledIOP, round int, q query.MultilinearEval, nRow, idx int) *Context {
	n := mustAllSameNumVars(q)
	nCol := n - nRow
	K := len(q.Pols)

	suffix := fmt.Sprintf("n%d_r%d_%d_i%d_share2", n, round, comp.SelfRecursionCount, idx)

	alpha := comp.InsertCoin(
		round+1,
		coin.Name(fmt.Sprintf("MLVORTEX_ALPHA_%s", suffix)),
		coin.FieldExt,
	)

	uAlphaCol := comp.InsertProof(
		round+1,
		ifaces.ColID(fmt.Sprintf("MLVORTEX_UALPHA_share2_%s", suffix)),
		1<<nCol,
		false,
	)
	rowEvalsCol := comp.InsertProof(
		round+1,
		ifaces.ColID(fmt.Sprintf("MLVORTEX_ROWEVAL_share2_%s", suffix)),
		1<<nRow,
		false,
	)

	// Single UCols query at the shared cCol.
	uColsQuery := comp.InsertMultilinear(
		round+1,
		ifaces.QueryID(fmt.Sprintf("MLVORTEX_UCOL_share2_%s", suffix)),
		[]ifaces.Column{uAlphaCol},
	)
	// Single RowClaims query with K duplicate polys at K different cRow_k.
	rowClaimsPols := make([]ifaces.Column, K)
	for k := range rowClaimsPols {
		rowClaimsPols[k] = rowEvalsCol
	}
	rowClaimsQuery := comp.InsertMultilinear(
		round+1,
		ifaces.QueryID(fmt.Sprintf("MLVORTEX_ROW_share2_%s", suffix)),
		rowClaimsPols,
	)

	// Propagate the cCol-shared guarantee. The RowClaims query has K
	// duplicate polys at K distinct points; after Batch each residual point
	// becomes challenges[:nq] — identical across all K, so the next round's
	// query trivially satisfies the same shared-cCol property.
	markSharedSafeQuery(comp, rowClaimsQuery.Name())
	// The UCols query has only one pol; it isn't "shared-input" so the
	// downstream compileWithNRow will take the standard buildContext path.

	return &Context{
		InputQuery:     q,
		NumVars:        n,
		NRow:           nRow,
		NCol:           nCol,
		Round:          round,
		AlphaCoin:      alpha,
		UAlpha:         []ifaces.Column{uAlphaCol},
		RowEvals:       []ifaces.Column{rowEvalsCol},
		UCols:          []query.MultilinearEval{uColsQuery},
		RowClaims:      []query.MultilinearEval{rowClaimsQuery},
		SharedInput:    true,
		SharedRowEvals: true,
	}
}

// allSameInput reports whether every column in pols has the same ColID. When
// true, the K UAlpha vectors are all equal (they depend only on the shared
// column and α), so we can commit a single UAlpha proof column instead of K
// duplicates. RowEvals are equal when the per-pol cCol slices are also shared
// (a stronger condition checked at runtime).
func allSameInput(pols []ifaces.Column) bool {
	if len(pols) < 2 {
		return false
	}
	id := pols[0].GetColID()
	for _, c := range pols[1:] {
		if c.GetColID() != id {
			return false
		}
	}
	return true
}

// buildContextSharedInput builds a Context for the SharedInput regime: K input
// claims on the SAME column. Emits exactly ONE UAlpha proof column and ONE
// RowEvals proof column (the unique values shared by all k), and K
// MultilinearEval queries (UCols and RowClaims) each referencing the single
// UAlpha/RowEvals. This collapses the K-fold proof-column duplication that
// arises whenever a packed bootstrapper merges K original claims onto a
// single packed Q.
func buildContextSharedInput(comp *wizard.CompiledIOP, round int, q query.MultilinearEval, nRow, idx int) *Context {
	n := mustAllSameNumVars(q)
	nCol := n - nRow
	K := len(q.Pols)

	suffix := fmt.Sprintf("n%d_r%d_%d_i%d_shared", n, round, comp.SelfRecursionCount, idx)

	alpha := comp.InsertCoin(
		round+1,
		coin.Name(fmt.Sprintf("MLVORTEX_ALPHA_%s", suffix)),
		coin.FieldExt,
	)

	// UAlpha is a function of (Q, α) only, so it is shared across all K
	// claims regardless of the per-pol point. One proof column suffices.
	uAlphaCol := comp.InsertProof(
		round+1,
		ifaces.ColID(fmt.Sprintf("MLVORTEX_UALPHA_shared_%s", suffix)),
		1<<nCol,
		false,
	)
	// RowEvals depends on cCol_k, which CAN differ per k whenever any
	// original poly has nv < nCol (locator bits then spill into cCol). We
	// keep K separate RowEvals columns to handle this general case; when
	// cCol_k is shared (e.g. every original nv ≥ nCol), the prover fast
	// path computes the value once and writes the same content K times.
	rowEvals := make([]ifaces.Column, K)
	uCols := make([]query.MultilinearEval, K)
	rowClaims := make([]query.MultilinearEval, K)
	for k := range q.Pols {
		rowEvals[k] = comp.InsertProof(
			round+1,
			ifaces.ColID(fmt.Sprintf("MLVORTEX_ROWEVAL_%d_%s", k, suffix)),
			1<<nRow,
			false,
		)
		uCols[k] = comp.InsertMultilinear(
			round+1,
			ifaces.QueryID(fmt.Sprintf("MLVORTEX_UCOL_%d_%s", k, suffix)),
			[]ifaces.Column{uAlphaCol},
		)
		rowClaims[k] = comp.InsertMultilinear(
			round+1,
			ifaces.QueryID(fmt.Sprintf("MLVORTEX_ROW_%d_%s", k, suffix)),
			[]ifaces.Column{rowEvals[k]},
		)
	}

	return &Context{
		InputQuery:  q,
		NumVars:     n,
		NRow:        nRow,
		NCol:        nCol,
		Round:       round,
		AlphaCoin:   alpha,
		UAlpha:      []ifaces.Column{uAlphaCol},
		RowEvals:    rowEvals,
		UCols:       uCols,
		RowClaims:   rowClaims,
		SharedInput: true,
	}
}

// CompileRoundPacked is like CompileRound but uses the locator-tuple packing
// strategy: instead of K separate UAlpha/RowEvals proof columns per query, it
// creates ONE packed column of size KPow2·2^NCol (resp. KPow2·2^NRow) where
// KPow2 = 2^ceil(log2(K)). CommitMLColumns then builds ONE Merkle tree for the
// packed column instead of K trees, reducing both proof size and gnark circuit cost.
//
// Each downstream ML claim references the packed column at the locator-extended
// point (l_k ‖ c_col) where l_k is the L-bit big-endian binary encoding of k.
// These K claims are handled correctly by multilineareval.Batch.
func CompileRoundPacked(comp *wizard.CompiledIOP) {
	compileWithNRowPacked(comp, -1)
	CommitMLColumns(comp)
	CommitOriginalMLColumns(comp)
}

// compileWithNRowPacked is the packed variant of compileWithNRow.
func compileWithNRowPacked(comp *wizard.CompiledIOP, nFoldRows int) {
	pendingProver := make(map[int][]*ProverAction)

	idx := 0
	for _, name := range comp.QueriesParams.AllUnignoredKeys() {
		q, ok := comp.QueriesParams.Data(name).(query.MultilinearEval)
		if !ok {
			continue
		}
		comp.QueriesParams.MarkAsIgnored(name)
		r := comp.QueriesParams.Round(name)

		n := mustAllSameNumVars(q)
		if terminalShortcutAllowed(comp, q, n) {
			comp.RegisterVerifierAction(comp.NumRounds()-1, &TerminalVerifierAction{Q: q})
			idx++
			continue
		}

		// In compileWithNRowPacked the K input polys are locator-packed into
		// ONE proof column for committed inputs; for non-committed inputs the
		// K UAlpha/RowEvals are themselves packed into a single Merkle tree
		// of effective K = 1 column. Either way kOrig = 1 for the balance calc.
		nRow := chooseNRow(n, 1, nFoldRows)
		var ctx *Context
		if hasCommittedInputs(comp, q) {
			ctx = buildContext(comp, r, q, nRow, idx)
		} else {
			ctx = buildContextPacked(comp, r, q, nRow, idx)
		}
		pendingProver[r+1] = append(pendingProver[r+1], &ProverAction{Ctx: ctx})
		comp.RegisterVerifierAction(comp.NumRounds()-1, &VerifierAction{Ctx: ctx})

		if comp.ExtraData == nil {
			comp.ExtraData = make(map[string]any)
		}
		ctxs, _ := comp.ExtraData[mlvortexContextsKey].([]*Context)
		comp.ExtraData[mlvortexContextsKey] = append(ctxs, ctx)

		idx++
	}

	for round, actions := range pendingProver {
		if len(actions) == 1 {
			comp.RegisterProverAction(round, actions[0])
		} else {
			comp.RegisterProverAction(round, &proverActionBatch{actions: actions})
		}
	}
}

// buildContextPacked builds a Context using the locator-tuple packing strategy.
// All K UAlpha/RowEvals vectors are packed into a single proof column each.
//
// Packed layout (UAlpha):  packed[k·nColSize:(k+1)·nColSize] = UAlpha_k  (k=0…K-1)
//                          packed[k·nColSize:(k+1)·nColSize] = 0         (k=K…KPow2-1)
//
// UCols[k] claims packed UAlpha at point (l_k ‖ c_col) where l_k is the
// L-bit big-endian binary encoding of k; MLE evaluation gives UAlpha_k(c_col).
func buildContextPacked(comp *wizard.CompiledIOP, round int, q query.MultilinearEval, nRow, idx int) *Context {
	n := mustAllSameNumVars(q)
	nCol := n - nRow
	K := len(q.Pols)

	// L = ceil(log2(K)); KPow2 = 2^L.
	L := bits.Len(uint(K - 1)) // bits.Len(0)=0 when K=1; ceil(log2(K)) otherwise
	if K == 1 {
		L = 0
	}
	KPow2 := 1 << L

	suffix := fmt.Sprintf("n%d_r%d_%d_i%d", n, round, comp.SelfRecursionCount, idx)

	alpha := comp.InsertCoin(
		round+1,
		coin.Name(fmt.Sprintf("MLVORTEX_ALPHA_%s", suffix)),
		coin.FieldExt,
	)

	// ONE packed proof column for all K UAlpha vectors (KPow2 × 2^nCol elements).
	packedUAlpha := comp.InsertProof(
		round+1,
		ifaces.ColID(fmt.Sprintf("MLVORTEX_UALPHA_packed_%s", suffix)),
		KPow2*(1<<nCol),
		false,
	)
	// ONE packed proof column for all K RowEvals vectors (KPow2 × 2^nRow elements).
	packedRowEvals := comp.InsertProof(
		round+1,
		ifaces.ColID(fmt.Sprintf("MLVORTEX_ROWEVAL_packed_%s", suffix)),
		KPow2*(1<<nRow),
		false,
	)

	// K downstream ML claims; each at the locator-extended point (l_k ‖ c_col/c_row).
	uCols := make([]query.MultilinearEval, K)
	rowClaims := make([]query.MultilinearEval, K)
	for k := range q.Pols {
		uCols[k] = comp.InsertMultilinear(
			round+1,
			ifaces.QueryID(fmt.Sprintf("MLVORTEX_UCOL_%d_%s", k, suffix)),
			[]ifaces.Column{packedUAlpha},
		)
		rowClaims[k] = comp.InsertMultilinear(
			round+1,
			ifaces.QueryID(fmt.Sprintf("MLVORTEX_ROW_%d_%s", k, suffix)),
			[]ifaces.Column{packedRowEvals},
		)
	}

	return &Context{
		InputQuery: q,
		NumVars:    n,
		NRow:       nRow,
		NCol:       nCol,
		Round:      round,
		AlphaCoin:  alpha,
		UAlpha:     []ifaces.Column{packedUAlpha},
		RowEvals:   []ifaces.Column{packedRowEvals},
		UCols:      uCols,
		RowClaims:  rowClaims,
		Packed:     true,
		L:          L,
		KPow2:      KPow2,
	}
}
