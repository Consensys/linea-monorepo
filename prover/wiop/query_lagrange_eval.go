package wiop

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/koalabear/field"
	"github.com/consensys/linea-monorepo/prover/maths/koalabear/polynomials"
)

// LagrangeEval is a [Query] that evaluates a batch of polynomials at a single
// point X and asserts that the results match a set of prover-supplied cells
// (the EvaluationClaims). The polynomials are represented in Lagrange basis
// over the roots-of-unity domain.
//
// For a column C of size n and evaluation point X, the evaluation is:
//
//	(X^n − 1) / n · Σ_{i<n} C[i] / (X − ω^i)
//
// where ω is the canonical n-th root of unity of the ambient field.
//
// LagrangeEval is both [GnarkCheckableQuery] and [AssignableQuery]:
//   - Check and CheckGnark assert that all polynomials and the evaluation point
//     are publicly visible, then verify that the claims match the evaluations.
//   - SelfAssign computes the evaluations from the runtime assignment and
//     stores them in the EvaluationClaims cells.
//
// An important invariant: the round of the EvaluationPoint must be strictly
// later than the round of every polynomial column. The default constructor
// [System.NewLagrangeEval] enforces this automatically.
type LagrangeEval struct {
	baseQuery
	// Polynomials is the ordered list of column views to evaluate.
	// All views must reference columns from the same module.
	Polynomials []*ColumnView
	// EvaluationPoint is the symbolic scalar at which the polynomials are
	// evaluated. It is typically a verifier coin or a public cell.
	EvaluationPoint FieldPromise
	// EvaluationClaims is the parallel list of cells holding the alleged
	// evaluations. EvaluationClaims[i] is the prover's claim for Polynomials[i].
	EvaluationClaims []*Cell
}

// Round implements [Query]. Returns the round of the EvaluationPoint, which
// by the construction invariant is the latest round among all referenced
// objects.
//
// Panics if the EvaluationPoint does not carry a round (i.e. it is not a
// *Cell or *CoinField).
func (le *LagrangeEval) Round() *Round {
	r := roundOf(le.EvaluationPoint)
	if r == nil {
		panic("wiop: LagrangeEval.Round: EvaluationPoint has no associated round; it must be a *Cell or *CoinField")
	}
	return r
}

// IsAlreadyAssigned implements [AssignableQuery]. Reports whether all
// EvaluationClaims cells already hold a runtime assignment.
func (le *LagrangeEval) IsAlreadyAssigned(rt Runtime) bool {
	for _, claim := range le.EvaluationClaims {
		if !rt.HasCellValue(claim) {
			return false
		}
	}
	return true
}

// SelfAssign implements [AssignableQuery]. Evaluates each polynomial at the
// EvaluationPoint and writes the results into the corresponding
// EvaluationClaims cells.
func (le *LagrangeEval) SelfAssign(rt Runtime) {
	evals := le.evalPolynomials(rt)
	for i, claim := range le.EvaluationClaims {
		rt.AssignCell(claim, evals[i])
	}
}

// Check implements [Query]. For each polynomial in Polynomials, evaluates it
// at EvaluationPoint using the barycentric Lagrange formula and asserts that
// the result equals the corresponding EvaluationClaims cell.
//
// Precondition: every polynomial column must be assigned in rt. The method
// returns a descriptive error for the first misassigned column or failing
// claim rather than panicking, so callers can surface the problem cleanly.
func (le *LagrangeEval) Check(rt Runtime) error {

	// Verify that all polynomial columns have been assigned in the runtime.
	for i, pv := range le.Polynomials {
		if !rt.HasColumnAssignment(pv.Column) {
			return fmt.Errorf(
				"wiop: LagrangeEval(%s): polynomial[%d] column %q is not assigned",
				le.context.Path(), i, pv.Column.Context.Path(),
			)
		}
	}

	for i, got := range le.evalPolynomials(rt) {
		claim := rt.GetCellValue(le.EvaluationClaims[i])
		diff := got.Sub(claim)
		if !diff.Ext.IsZero() {
			return fmt.Errorf(
				"wiop: LagrangeEval(%s): polynomial[%d] evaluation mismatch at claim cell %q",
				le.context.Path(), i, le.EvaluationClaims[i].Context.Path(),
			)
		}
	}

	return nil
}

// evalPolynomials evaluates each polynomial in le.Polynomials at the
// EvaluationPoint, applying the cyclic-shift adjustment for each [ColumnView].
// It is the shared kernel used by both [Check] and [SelfAssign].
func (le *LagrangeEval) evalPolynomials(rt Runtime) []field.FieldElem {
	evalPoint := le.EvaluationPoint.EvaluateSingle(rt)
	results := make([]field.FieldElem, len(le.Polynomials))
	for i, pv := range le.Polynomials {
		// Adjust the evaluation point for the column view's cyclic shift.
		// C'[j] = C[(j+k) mod n]  implies  C'(z) = C(ω^k · z),
		// so we evaluate the original column data at ω^k · z instead.
		z := evalPoint.Value
		data := rt.GetColumnAssignment(pv.Column).Plain[0]
		if k := pv.ShiftingOffset; k != 0 {
			var (
				n      = pv.Column.Module.Size()
				omega  = field.RootOfUnityBy(n)
				omegaK field.Element
			)
			omegaK.ExpInt64(omega, int64(k))
			z = z.Mul(field.ElemFromBase(omegaK))
		}
		results[i] = polynomials.EvalLagrange(data, z)
	}
	return results
}

// CheckGnark implements [GnarkCheckableQuery]. Asserts inside a gnark circuit
// that each polynomial evaluates to the claimed value at the EvaluationPoint.
//
// Precondition: all polynomials and the EvaluationPoint must be publicly
// visible; panics otherwise.
//
// TODO: Implement once the gnark layer is defined.
func (le *LagrangeEval) CheckGnark(_ frontend.API, _ GnarkRuntime) {
	panic("wiop: LagrangeEval.CheckGnark not yet implemented")
}

// NewLagrangeEval constructs and registers a [LagrangeEval] query on sys.
// One fresh [Cell] is allocated automatically for each polynomial, placed in
// the round immediately following the latest polynomial column's round.
//
// The evaluation point x must carry a round (i.e. be a *Cell or *CoinField).
// The invariant that x's round is ≥ the claim cells' round is the caller's
// responsibility.
//
// Panics if ctx is nil, polys is empty, no round follows the latest column
// round, or x does not carry a round.
func (sys *System) NewLagrangeEval(ctx *ContextFrame, polys []*ColumnView, x FieldPromise) *LagrangeEval {
	if ctx == nil {
		panic("wiop: System.NewLagrangeEval requires a non-nil ContextFrame")
	}
	if len(polys) == 0 {
		panic("wiop: System.NewLagrangeEval requires at least one polynomial")
	}

	// Identify the latest round among the polynomial columns.
	var maxColRound *Round
	for _, pv := range polys {
		r := pv.Round()
		if maxColRound == nil || r.ID > maxColRound.ID {
			maxColRound = r
		}
	}

	// The EvaluationClaims live in the round immediately following the latest
	// column round. This round must already exist in the system; NewLagrangeEval
	// does not create new rounds.
	claimRound, ok := maxColRound.Next()
	if !ok {
		panic(fmt.Sprintf(
			"wiop: System.NewLagrangeEval: no round follows round %d (the latest polynomial column round); "+
				"call sys.NewRound() before registering this query",
			maxColRound.ID,
		))
	}

	// Allocate one claim cell per polynomial. The cell inherits the extension
	// status of its corresponding column view.
	claims := make([]*Cell, len(polys))
	for i, pv := range polys {
		cellCtx := ctx.Childf("claim[%d]", i)
		claims[i] = claimRound.NewCell(cellCtx, pv.IsExtension())
	}

	return sys.newLagrangeEval(ctx, polys, x, claims)
}

// NewLagrangeEvalFrom constructs and registers a [LagrangeEval] query on sys,
// using caller-supplied EvaluationClaims instead of freshly allocated ones.
// This is used when the claim cells already exist (e.g. when two queries share
// the same set of alleged evaluations).
//
// Panics if ctx is nil, polys is empty, or len(claims) != len(polys).
func (sys *System) NewLagrangeEvalFrom(ctx *ContextFrame, polys []*ColumnView, x FieldPromise, claims []*Cell) *LagrangeEval {
	if ctx == nil {
		panic("wiop: System.NewLagrangeEvalFrom requires a non-nil ContextFrame")
	}
	if len(polys) == 0 {
		panic("wiop: System.NewLagrangeEvalFrom requires at least one polynomial")
	}
	if len(claims) != len(polys) {
		panic(fmt.Sprintf(
			"wiop: System.NewLagrangeEvalFrom: claim count (%d) must equal polynomial count (%d)",
			len(claims), len(polys),
		))
	}
	return sys.newLagrangeEval(ctx, polys, x, claims)
}

// newLagrangeEval is the shared constructor used by [System.NewLagrangeEval]
// and [System.NewLagrangeEvalFrom]. It builds the [LagrangeEval], appends it
// to sys.LagrangeEvals, and returns it.
func (sys *System) newLagrangeEval(ctx *ContextFrame, polys []*ColumnView, x FieldPromise, claims []*Cell) *LagrangeEval {
	le := &LagrangeEval{
		baseQuery: baseQuery{
			context:     ctx,
			Annotations: make(Annotations),
		},
		Polynomials:      polys,
		EvaluationPoint:  x,
		EvaluationClaims: claims,
	}
	sys.LagrangeEvals = append(sys.LagrangeEvals, le)
	return le
}
