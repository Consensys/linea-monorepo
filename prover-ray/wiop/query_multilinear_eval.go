package wiop

import (
	"fmt"
	"math/bits"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover-ray/maths/koalabear/field"
	"github.com/consensys/linea-monorepo/prover-ray/maths/koalabear/polynomials"
)

// MultilinearEval is a [Query] that evaluates a batch of multilinear polynomials
// at a single point in the boolean hypercube and asserts that the results match
// a set of prover-supplied cells (the EvaluationClaims).
//
// A multilinear polynomial in n variables is stored as a column of size 2ⁿ,
// whose entries are the evaluations on the boolean hypercube {0,1}ⁿ. The
// coordinate convention follows [polynomials.EvalMultilin]: EvaluationPoints[0]
// selects the most-significant bit of the hypercube index.
//
// Column size must be a power of two. Cyclic shifts ([ColumnView.ShiftingOffset] ≠ 0)
// are not supported, because cyclic shifts do not correspond to a simple
// coordinate transform on the boolean hypercube.
//
// Padding is handled by materialising the full 2ⁿ-length table using the
// module's declared padding direction and padding value.
//
// MultilinearEval is both [GnarkCheckableQuery] and [AssignableQuery]:
//   - Check asserts all columns and claims are assigned, recomputes the
//     evaluations, and verifies claims match.
//   - SelfAssign computes the evaluations and writes them into the claim cells.
//
// Invariant: the round of every EvaluationPoints element must be ≥ the round
// of the claim cells. [System.NewMultilinearEval] enforces this automatically
// by placing claims in the round immediately after the latest column round.
type MultilinearEval struct {
	baseQuery
	// Polynomials is the ordered list of column views to evaluate. All views
	// must have ShiftingOffset == 0 and column sizes must be powers of two.
	Polynomials []*ColumnView
	// EvaluationPoints is the evaluation coordinate vector. Its length equals
	// log₂(column size). EvaluationPoints[0] selects the most-significant bit.
	EvaluationPoints []FieldPromise
	// EvaluationClaims is the parallel list of cells holding the alleged
	// evaluations. EvaluationClaims[i] is the prover's claim for Polynomials[i].
	EvaluationClaims []*Cell
}

// Round implements [Query]. Returns the latest round among all EvaluationPoints,
// which by the construction invariant is the latest round among all referenced
// objects.
func (me *MultilinearEval) Round() *Round {
	var maxRound *Round
	for _, ep := range me.EvaluationPoints {
		r := roundOf(ep)
		if r != nil && (maxRound == nil || r.ID > maxRound.ID) {
			maxRound = r
		}
	}
	if maxRound == nil {
		panic("wiop: MultilinearEval.Round: no EvaluationPoint carries a round; each must be a *Cell or *CoinField")
	}
	return maxRound
}

// IsAlreadyAssigned implements [AssignableQuery]. Reports whether all
// EvaluationClaims cells already hold a runtime assignment.
func (me *MultilinearEval) IsAlreadyAssigned(rt Runtime) bool {
	for _, claim := range me.EvaluationClaims {
		if !rt.HasCellValue(claim) {
			return false
		}
	}
	return true
}

// SelfAssign implements [AssignableQuery]. Evaluates each polynomial at the
// EvaluationPoints and writes the results into the corresponding
// EvaluationClaims cells.
func (me *MultilinearEval) SelfAssign(rt Runtime) {
	for i, v := range me.evalPolynomials(rt) {
		rt.AssignCell(me.EvaluationClaims[i], v)
	}
}

// Check implements [Query]. Verifies that every polynomial column is assigned,
// recomputes the multilinear evaluations at EvaluationPoints, and returns an
// error if any claim cell is unassigned or does not match the computed value.
func (me *MultilinearEval) Check(rt Runtime) error {
	for i, pv := range me.Polynomials {
		if !rt.HasColumnAssignment(pv.Column) {
			return fmt.Errorf(
				"wiop: MultilinearEval(%s): polynomial[%d] column %q is not assigned",
				me.context.Path(), i, pv.Column.Context.Path(),
			)
		}
	}

	for i, got := range me.evalPolynomials(rt) {
		claimCell := me.EvaluationClaims[i]
		if !rt.HasCellValue(claimCell) {
			return fmt.Errorf(
				"wiop: MultilinearEval(%s): claim cell[%d] %q is not assigned",
				me.context.Path(), i, claimCell.Context.Path(),
			)
		}
		claim := rt.GetCellValue(claimCell)
		diff := got.Sub(claim)
		if !diff.IsZero() {
			return fmt.Errorf(
				"wiop: MultilinearEval(%s): polynomial[%d] evaluation mismatch at claim cell %q",
				me.context.Path(), i, claimCell.Context.Path(),
			)
		}
	}
	return nil
}

// CheckGnark implements [GnarkCheckableQuery]. Not yet implemented.
func (me *MultilinearEval) CheckGnark(_ frontend.API, _ GnarkRuntime) {
	panic("wiop: MultilinearEval.CheckGnark not yet implemented")
}

// evalPolynomials evaluates each polynomial in me.Polynomials at the
// EvaluationPoints. It is the shared kernel for [Check] and [SelfAssign].
func (me *MultilinearEval) evalPolynomials(rt Runtime) []field.Gen {
	coords := make([]field.Gen, len(me.EvaluationPoints))
	for i, ep := range me.EvaluationPoints {
		coords[i] = ep.EvaluateSingle(rt).Value
	}

	results := make([]field.Gen, len(me.Polynomials))
	for i, pv := range me.Polynomials {
		col := pv.Column
		m := col.Module
		n := m.RuntimeSize(rt)
		cv := rt.GetColumnAssignment(col)
		table := materializeTable(cv, m.Padding, n)
		results[i] = polynomials.EvalMultilin(table, coords)
	}
	return results
}

// materializeTable builds the full-length field.Vec for a concrete column
// assignment, expanding padding to produce a slice of exactly n elements.
func materializeTable(cv *ConcreteVector, padding PaddingDirection, n int) field.Vec {
	if cv.Plain.IsBase() {
		tbl := make([]field.Element, n)
		for j := 0; j < n; j++ {
			tbl[j] = cv.ElementAtN(padding, n, j).AsBase()
		}
		return field.VecFromBase(tbl)
	}
	tbl := make([]field.Ext, n)
	for j := 0; j < n; j++ {
		e := cv.ElementAtN(padding, n, j)
		if e.IsBase() {
			tbl[j] = field.Lift(e.AsBase())
		} else {
			tbl[j] = e.AsExt()
		}
	}
	return field.VecFromExt(tbl)
}

// NewMultilinearEval constructs and registers a [MultilinearEval] query on sys.
//
// One fresh [Cell] is allocated automatically for each polynomial, placed in
// the round immediately following the latest polynomial column's round.
// Claim cells are always extension-field to accommodate extension-field
// evaluation points.
//
// len(evalPoints) must equal log₂(column size). The column size must be a
// power of two, and all columns must have the same size. Cyclic shifts
// (ShiftingOffset ≠ 0) are rejected.
//
// Panics if ctx or polys is nil/empty, if any column view is shifted, if the
// column size is not a power of two, if len(evalPoints) ≠ log₂(size), or if
// no round follows the latest column round.
func (sys *System) NewMultilinearEval(ctx *ContextFrame, polys []*ColumnView,
	evalPoints []FieldPromise) *MultilinearEval {

	if ctx == nil {
		panic("wiop: System.NewMultilinearEval requires a non-nil ContextFrame")
	}
	if len(polys) == 0 {
		panic("wiop: System.NewMultilinearEval requires at least one polynomial")
	}
	if len(evalPoints) == 0 {
		panic("wiop: System.NewMultilinearEval requires at least one evaluation point")
	}

	var maxColRound *Round
	var refSize int
	for k, pv := range polys {
		if pv.ShiftingOffset != 0 {
			panic(fmt.Sprintf(
				"wiop: System.NewMultilinearEval: polynomial[%d] %q has a non-zero ShiftingOffset (%d); "+
					"cyclic shifts are not supported for multilinear evaluation",
				k, pv.Column.Context.Path(), pv.ShiftingOffset,
			))
		}

		size := pv.Column.Module.Size()
		if size == 0 {
			panic(fmt.Sprintf(
				"wiop: System.NewMultilinearEval: polynomial[%d] %q belongs to an unsized module; "+
					"call Module.SetSize before registering this query",
				k, pv.Column.Context.Path(),
			))
		}
		if bits.OnesCount(uint(size)) != 1 {
			panic(fmt.Sprintf(
				"wiop: System.NewMultilinearEval: polynomial[%d] %q has size %d which is not a power of two",
				k, pv.Column.Context.Path(), size,
			))
		}

		logN := bits.TrailingZeros(uint(size))
		if len(evalPoints) != logN {
			panic(fmt.Sprintf(
				"wiop: System.NewMultilinearEval: polynomial[%d] %q has size %d (log=%d) "+
					"but %d evaluation points were provided",
				k, pv.Column.Context.Path(), size, logN, len(evalPoints),
			))
		}

		if refSize == 0 {
			refSize = size
		} else if size != refSize {
			panic(fmt.Sprintf(
				"wiop: System.NewMultilinearEval: polynomial[%d] %q has size %d but polynomial[0] has size %d; "+
					"all columns must have the same size",
				k, pv.Column.Context.Path(), size, refSize,
			))
		}

		r := pv.Round()
		if maxColRound == nil || r.ID > maxColRound.ID {
			maxColRound = r
		}
	}

	claimRound, ok := maxColRound.Next()
	if !ok {
		panic(fmt.Sprintf(
			"wiop: System.NewMultilinearEval: no round follows round %d (the latest polynomial column round); "+
				"call sys.NewRound() before registering this query",
			maxColRound.ID,
		))
	}

	// Claim cells are always extension-field: evaluation points are extension
	// coins in any non-trivial protocol, so the result is extension-field.
	claims := make([]*Cell, len(polys))
	for i := range polys {
		claims[i] = claimRound.NewCell(ctx.Childf("claim[%d]", i), true)
	}

	return sys.newMultilinearEval(ctx, polys, evalPoints, claims)
}

// NewMultilinearEvalFrom constructs and registers a [MultilinearEval] query on
// sys using caller-supplied EvaluationClaims instead of freshly allocated ones.
// This is used when claim cells already exist (e.g. when two queries share the
// same alleged evaluations).
//
// Panics if ctx is nil, polys is empty, or len(claims) ≠ len(polys).
func (sys *System) NewMultilinearEvalFrom(ctx *ContextFrame, polys []*ColumnView,
	evalPoints []FieldPromise, claims []*Cell) *MultilinearEval {
	if ctx == nil {
		panic("wiop: System.NewMultilinearEvalFrom requires a non-nil ContextFrame")
	}
	if len(polys) == 0 {
		panic("wiop: System.NewMultilinearEvalFrom requires at least one polynomial")
	}
	if len(claims) != len(polys) {
		panic(fmt.Sprintf(
			"wiop: System.NewMultilinearEvalFrom: claim count (%d) must equal polynomial count (%d)",
			len(claims), len(polys),
		))
	}
	return sys.newMultilinearEval(ctx, polys, evalPoints, claims)
}

// newMultilinearEval is the shared constructor used by [System.NewMultilinearEval]
// and [System.NewMultilinearEvalFrom].
func (sys *System) newMultilinearEval(ctx *ContextFrame, polys []*ColumnView,
	evalPoints []FieldPromise, claims []*Cell) *MultilinearEval {
	me := &MultilinearEval{
		baseQuery: baseQuery{
			context:     ctx,
			Annotations: make(Annotations),
		},
		Polynomials:      polys,
		EvaluationPoints: evalPoints,
		EvaluationClaims: claims,
	}
	sys.MultilinearEvals = append(sys.MultilinearEvals, me)
	return me
}
