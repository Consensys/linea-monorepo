// Package lookuptologderivsum compiles every unreduced [wiop.TableRelation]
// of kind [wiop.TableRelationInclusion] into a single [wiop.LogDerivativeSum2]
// query whose final result is asserted to be zero. It is the prover-ray
// analogue of linea/prover/protocol/compiler/logderivativesum's
// LookupIntoLogDerivativeSum pass.
//
// The reduction follows the standard log-derivative argument:
//
//   - For each lookup-table fragment T (the "including" side B), commit a
//     multiplicity column M whose value at row i counts how many times that
//     row's value appears across the union of all checked-table A's that
//     reference T.
//
//   - Sample two random extension-field coins:
//
//   - α — used only when the lookup is multi-column, to fold every row
//     into a single field element via random linear combination.
//
//   - γ — used to randomise the denominators (γ + RLC(row)).
//
//   - Emit fractions
//
//     Σ_T  ( −M(row) ) / ( γ + RLC(T(row)) )       (one per T fragment)
//     Σ_S  ( filter_S(row) ) / ( γ + RLC(S(row)) ) (one per A fragment)
//
//     into a single [wiop.LogDerivativeSum2] query. The α-randomised RLC
//     binds the multi-column case via Schwartz–Zippel; the γ-randomised
//     denominator makes every Den non-zero with overwhelming probability,
//     which is what closes the zero-denominator soundness gap that the
//     LogDerivativeSum2 constraint system inherits.
//
//   - Register a verifier action that asserts the LogDerivativeSum2's
//     Result cell is zero (the standard log-derivative identity).
//
// B-side filters (selectors on the including side) are folded into the RLC
// itself by prepending the filter to the B-side and a constant 1 to the
// A-side, mirroring linea/lookup2logderivsum's IsFilteredOnIncluding
// handling.
//
// After Compile runs, every consumed [wiop.TableRelation] is marked reduced
// and a single [wiop.LogDerivativeSum2] query is left in sys for the
// downstream [logderivativesum2] compiler pass to consume.
//
// Scope (MVP): the compiler handles inclusion queries with len(B) == 1
// (single-fragment lookup table). Multiple A fragments per query and
// multiple queries per shared B are both supported. Permutation queries are
// ignored. Multi-fragment B queries are explicitly rejected with a panic
// at compile time so callers can rework them into multiple single-fragment
// queries before running this pass.
package lookuptologderivsum

import (
	"fmt"
	"sort"

	"github.com/consensys/linea-monorepo/prover-ray/maths/koalabear/field"
	"github.com/consensys/linea-monorepo/prover-ray/wiop"
)

// Compile reduces every unreduced inclusion [wiop.TableRelation] in sys to a
// single [wiop.LogDerivativeSum2] query plus a multiplicity column per
// lookup-table fragment, plus a verifier action that asserts the resulting
// LogDerivativeSum2 result equals zero.
//
// All consumed inclusion queries are marked reduced. Permutation queries and
// already-reduced queries are skipped. If sys contains no eligible queries
// the function is a no-op.
//
// Panics if any inclusion query has len(B) != 1 (multi-fragment lookup tables
// are out of scope for this pass).
func Compile(sys *wiop.System) {
	groups := collectGroups(sys)
	if len(groups) == 0 {
		return
	}

	// Allocate the LogDerivativeSum2 in a deterministic group order. We sort
	// by the witness-round ID then by canonical key so the registration order
	// is independent of map iteration.
	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Determine the latest witness round across every group: this dictates
	// where the coin and result rounds live.
	var latestWitness *wiop.Round
	for _, k := range keys {
		g := groups[k]
		if latestWitness == nil || g.witnessRound.ID > latestWitness.ID {
			latestWitness = g.witnessRound
		}
	}

	// We need two more interactive rounds after latestWitness:
	//   - latestWitness + 1: where M, α, γ live.
	//   - latestWitness + 2: where the LogDerivativeSum2 result cell lives.
	coinRound := ensureNextRound(sys, latestWitness)
	resultRound := ensureNextRound(sys, coinRound)
	_ = resultRound // returned for clarity; the LogDerivativeSum2 constructor finds it on its own.

	compCtx := sys.Context.Childf("lookup2logderiv")

	// γ is shared across every group: a single random extension coin is
	// enough to randomise every denominator in the aggregated query.
	gamma := coinRound.NewCoinField(compCtx.Childf("gamma"))

	var (
		fractions  []wiop.Fraction
		mTasks     []*mAssignmentTask
		consumedQs []*wiop.TableRelation
	)
	for _, k := range keys {
		g := groups[k]
		gFractions, gTask := compileGroup(g, gamma, coinRound, compCtx)
		fractions = append(fractions, gFractions...)
		mTasks = append(mTasks, gTask)
		for _, inc := range g.included {
			consumedQs = append(consumedQs, inc.query)
		}
	}

	// Register one prover action per group. They are all independent so we
	// register them under the same coinRound.
	for _, t := range mTasks {
		coinRound.RegisterAction(t)
	}

	ld := sys.NewLogDerivativeSum(compCtx.Childf("aggregated"), fractions)

	// The verifier check: the aggregated log-derivative sum must be zero.
	ld.Result.Round().RegisterVerifierAction(&resultIsZeroVerifierAction{ld: ld})

	// Mark every consumed query as reduced so subsequent compiler passes skip
	// them. We deliberately wait until the LogDerivativeSum2 has been
	// registered so a panic during construction leaves the system unchanged.
	for _, q := range consumedQs {
		q.MarkAsReduced()
	}
}

// collectGroups walks sys.TableRelations once, picks out every unreduced
// inclusion query, and groups them by the canonical identity of their
// (single) B-side fragment.
func collectGroups(sys *wiop.System) map[string]*lookupGroup {
	groups := make(map[string]*lookupGroup)
	for _, q := range sys.TableRelations {
		if q.IsReduced() {
			continue
		}
		if q.Kind != wiop.TableRelationInclusion {
			continue
		}
		if len(q.B) != 1 {
			panic(fmt.Sprintf(
				"wiop/compilers/lookuptologderivsum: query %q has len(B)=%d; "+
					"only single-fragment lookup tables are supported in this pass",
				q.Context().Path(), len(q.B),
			))
		}
		key := canonicalIncludingKey(q.B[0])
		g, ok := groups[key]
		if !ok {
			g = &lookupGroup{
				including: includingTable{
					cols:     q.B[0].Columns,
					selector: q.B[0].Selector,
					module:   q.B[0].Module(),
				},
			}
			if !allIncludingColumnsShareModule(g.including) {
				panic(fmt.Sprintf(
					"wiop/compilers/lookuptologderivsum: query %q has a B fragment whose "+
						"columns or selector live on different modules",
					q.Context().Path(),
				))
			}
			groups[key] = g
		}
		// Update the witness round from B and from every A fragment of this
		// query.
		g.updateWitnessRound(q.B[0].Round())
		for _, tabA := range q.A {
			g.updateWitnessRound(tabA.Round())
			g.addIncluded(q, tabA)
		}
	}
	return groups
}

// compileGroup builds the fraction list and the multiplicity-assignment
// prover task for a single B-grouped collection of inclusion queries.
func compileGroup(
	g *lookupGroup,
	gamma *wiop.CoinField,
	coinRound *wiop.Round,
	compCtx *wiop.ContextFrame,
) ([]wiop.Fraction, *mAssignmentTask) {
	gCtx := compCtx.Childf("group-%p", g)

	// Consistency: every A fragment must have the same width as B.
	for i, inc := range g.included {
		if len(inc.cols) != g.including.width() {
			panic(fmt.Sprintf(
				"wiop/compilers/lookuptologderivsum: included fragment %d in group %s has width %d "+
					"but the lookup table has width %d",
				i, gCtx.Path(), len(inc.cols), g.including.width(),
			))
		}
	}

	// α is needed whenever we have to fold more than one column down to a
	// single field element. The "effective" B-side width is the original
	// width plus one when the IsFilteredOnIncluding trick prepends the
	// B-selector to the row (and a constant 1 to every A-side).
	prependOnesToA := g.including.selector != nil
	effectiveWidth := g.including.width()
	if prependOnesToA {
		effectiveWidth++
	}

	var alpha *wiop.CoinField
	if effectiveWidth > 1 {
		alpha = coinRound.NewCoinField(gCtx.Childf("alpha"))
	}

	// Build the symbolic random linear combination of the B-side columns.
	// When the B-side carries a selector we prepend it and prepend a
	// constant-1 to every A side, mirroring the standard
	// IsFilteredOnIncluding trick.
	var bRLC wiop.Expression
	if prependOnesToA {
		bRLC = randomLinearCombination(alpha, g.including.selector, viewExprs(g.including.cols))
	} else {
		bRLC = randomLinearCombinationOfViews(alpha, g.including.cols)
	}

	// M lives on the same module as B, in the coin round (which is also the
	// round where M is assigned by the prover).
	mCol := g.including.module.NewColumn(
		gCtx.Childf("M"),
		wiop.VisibilityOracle,
		coinRound,
	)

	// The T-side fraction:  −M / (γ + bRLC).
	negM := wiop.Negate(mCol.View())
	bDen := wiop.Add(gamma, bRLC)
	fractions := []wiop.Fraction{{
		Numerator:   negM,
		Denominator: bDen,
	}}

	// The A-side fractions: one per A fragment.
	for _, inc := range g.included {
		var sRLC wiop.Expression
		if prependOnesToA {
			oneOnA := wiop.NewConstantVector(inc.cols[0].Module(), field.One())
			sRLC = randomLinearCombination(alpha, oneOnA, viewExprs(inc.cols))
		} else {
			sRLC = randomLinearCombinationOfViews(alpha, inc.cols)
		}
		sDen := wiop.Add(gamma, sRLC)
		// Numerator is the constant 1 broadcast over the A fragment's module
		// so the fraction is vector-valued on the A side.
		oneNum := wiop.NewConstantVector(inc.cols[0].Module(), field.One())
		var filter wiop.Expression
		if inc.selector != nil {
			filter = inc.selector
		}
		fractions = append(fractions, wiop.Fraction{
			Filter:      filter,
			Numerator:   oneNum,
			Denominator: sDen,
		})
	}

	// The prover task that will fill M with multiplicities once the witness
	// is in place. M is referenced by both the symbolic recurrence and by
	// the prover task, so we share the same *Column pointer.
	task := &mAssignmentTask{
		m:               mCol,
		bCols:           g.including.cols,
		bSelector:       g.including.selector,
		included:        append([]includedSpec{}, g.included...),
		prependOneOnAOk: prependOnesToA,
	}

	return fractions, task
}

// ensureNextRound returns the round immediately following r, allocating one
// via [wiop.System.NewRound] if necessary.
func ensureNextRound(sys *wiop.System, r *wiop.Round) *wiop.Round {
	if next, ok := r.Next(); ok {
		return next
	}
	return sys.NewRound()
}

// randomLinearCombinationOfViews builds the symbolic expression
// cols[0] + α·cols[1] + α²·cols[2] + … . When alpha is nil the slice must
// have exactly one element (the single-column case); the function returns
// that column directly.
func randomLinearCombinationOfViews(alpha *wiop.CoinField, cols []*wiop.ColumnView) wiop.Expression {
	exprs := viewExprs(cols)
	if alpha == nil {
		if len(exprs) != 1 {
			panic("wiop/compilers/lookuptologderivsum: alpha is nil but width > 1")
		}
		return exprs[0]
	}
	return rlcExpression(alpha, exprs)
}

// randomLinearCombination is the same as [randomLinearCombinationOfViews]
// except the first term is taken from `head` (used to fold the
// IsFilteredOnIncluding prepended column into the same RLC).
func randomLinearCombination(alpha *wiop.CoinField, head wiop.Expression, rest []wiop.Expression) wiop.Expression {
	if alpha == nil {
		if len(rest) != 0 {
			panic("wiop/compilers/lookuptologderivsum: alpha is nil but a tail is present")
		}
		return head
	}
	exprs := append([]wiop.Expression{head}, rest...)
	return rlcExpression(alpha, exprs)
}

// rlcExpression returns exprs[0] + α·exprs[1] + α²·exprs[2] + … . Requires
// alpha != nil and len(exprs) >= 1.
func rlcExpression(alpha *wiop.CoinField, exprs []wiop.Expression) wiop.Expression {
	if len(exprs) == 0 {
		panic("wiop/compilers/lookuptologderivsum: rlcExpression requires at least one term")
	}
	acc := exprs[0]
	pow := wiop.Expression(alpha)
	for i := 1; i < len(exprs); i++ {
		acc = wiop.Add(acc, wiop.Mul(pow, exprs[i]))
		if i+1 < len(exprs) {
			pow = wiop.Mul(pow, alpha)
		}
	}
	return acc
}

// viewExprs lifts a slice of *ColumnView into the equivalent slice of
// wiop.Expression so it can be passed to [rlcExpression].
func viewExprs(cols []*wiop.ColumnView) []wiop.Expression {
	out := make([]wiop.Expression, len(cols))
	for i, cv := range cols {
		out[i] = cv
	}
	return out
}

// resultIsZeroVerifierAction asserts that the aggregated [wiop.LogDerivativeSum2]
// result cell holds the zero field element. This is the standard
// log-derivative identity: the sum of A-side fractions cancels the sum of
// T-side fractions exactly when every selected A row is in the union of
// selected B rows.
type resultIsZeroVerifierAction struct {
	ld *wiop.LogDerivativeSum
}

// Check implements [wiop.VerifierAction].
func (a *resultIsZeroVerifierAction) Check(rt wiop.Runtime) error {
	v := rt.GetCellValue(a.ld.Result)
	if !v.IsZero() {
		return fmt.Errorf(
			"wiop/compilers/lookuptologderivsum: aggregated lookup result for query %q must be zero",
			a.ld.Context().Path(),
		)
	}
	return nil
}
