package wiop

import "fmt"

// RationalReductionKind identifies the aggregation mode of a [RationalReduction].
type RationalReductionKind int

const (
	// RationalSum computes Result = ∑_k ∑_row Num_k[row] / Den_k[row].
	// This unifies LogDerivativeSum and InnerProduct (InnerProduct is the
	// special case where each denominator is the constant 1).
	RationalSum RationalReductionKind = iota
	// RationalProduct computes Result = ∏_k ∏_row Num_k[row] / Den_k[row].
	// This unifies GrandProduct.
	RationalProduct
)

// String implements [fmt.Stringer].
func (k RationalReductionKind) String() string {
	switch k {
	case RationalSum:
		return "Sum"
	case RationalProduct:
		return "Product"
	default:
		return fmt.Sprintf("RationalReductionKind(%d)", int(k))
	}
}

// Fraction is a value type pairing a Numerator and a Denominator expression.
// Both fields are always non-nil: use [Module.NewConstant] with value 1 to
// express a unit numerator or denominator.
//
// At least one of the two must be vector-valued (IsMultiValued() == true). If
// both are vector-valued they must reference the same module. These invariants
// are enforced by [System.NewRationalReduction].
type Fraction struct {
	// Numerator is the top expression. Always non-nil.
	Numerator Expression
	// Denominator is the bottom expression. Always non-nil.
	Denominator Expression
}

// RationalReduction is a [Query] that reduces a list of [Fraction] objects
// to a single field-element result stored in a [Cell]. The reduction is
// row-wise then fraction-wise according to [RationalReduction.Kind]:
//
//   - [RationalSum]:     Result = ∑_k ∑_row Num_k[row] / Den_k[row]
//   - [RationalProduct]: Result = ∏_k ∏_row Num_k[row] / Den_k[row]
//
// Different fractions may span different modules; no cross-fraction module
// constraint is imposed. The Result cell is allocated automatically in the
// round immediately following the latest column round across all fractions.
//
// RationalReduction implements [AssignableQuery] but not [GnarkCheckableQuery]:
// a compiler pass must reduce it before gnark verification.
//
// Use [System.NewRationalReduction] to construct and register an instance.
type RationalReduction struct {
	baseQuery
	// Kind is the aggregation mode (Sum or Product).
	Kind RationalReductionKind
	// Fractions is the ordered list of rational expression pairs. Contains
	// at least one entry.
	Fractions []Fraction
	// Result is the cell holding the prover's claimed aggregated value.
	// Allocated automatically by the constructor.
	Result *Cell
}

// Round implements [Query]. Returns the round of the [Result] cell, which is
// the round immediately following the latest column round across all fractions.
func (rr *RationalReduction) Round() *Round {
	return rr.Result.Round()
}

// IsAlreadyAssigned implements [AssignableQuery]. Reports whether the Result
// cell already holds a runtime assignment.
//
// TODO: Implement once Runtime is defined.
func (rr *RationalReduction) IsAlreadyAssigned(_ Runtime) bool {
	panic("wiop: RationalReduction.IsAlreadyAssigned not yet implemented")
}

// SelfAssign implements [AssignableQuery]. Computes the rational reduction
// from the runtime column assignments and writes the result into Result.
//
// TODO: Implement once Runtime is defined.
func (rr *RationalReduction) SelfAssign(_ Runtime) {
	panic("wiop: RationalReduction.SelfAssign not yet implemented")
}

// Check implements [Query]. Verifies that the Result cell holds the correct
// aggregated value computed from the fraction expressions.
//
// TODO: Implement once Runtime is defined.
func (rr *RationalReduction) Check(_ Runtime) error {
	panic(fmt.Sprintf("wiop: RationalReduction(%s).Check not yet implemented", rr.Kind))
}

// NewRationalReduction constructs and registers a [RationalReduction] query on
// sys. A fresh [Cell] is allocated automatically for the result, placed in the
// round immediately following the latest column round across all fractions.
//
// Invariants enforced at construction:
//   - len(fractions) ≥ 1.
//   - For each Fraction: Numerator and Denominator are both non-nil.
//   - For each Fraction: at least one of Numerator or Denominator is
//     vector-valued (IsMultiValued() == true, i.e. Module() != nil).
//   - For each Fraction: if both are vector-valued, they must share the same module.
//
// Panics if ctx is nil, any invariant is violated, or no round follows the
// latest fraction column round (call [System.NewRound] first in that case).
func (sys *System) NewRationalReduction(ctx *ContextFrame, kind RationalReductionKind, fractions []Fraction) *RationalReduction {
	if ctx == nil {
		panic("wiop: System.NewRationalReduction requires a non-nil ContextFrame")
	}
	if len(fractions) == 0 {
		panic("wiop: System.NewRationalReduction requires at least one Fraction")
	}

	var maxFracRound *Round
	for i, f := range fractions {
		if f.Numerator == nil {
			panic(fmt.Sprintf("wiop: System.NewRationalReduction: fraction %d has a nil Numerator", i))
		}
		if f.Denominator == nil {
			panic(fmt.Sprintf("wiop: System.NewRationalReduction: fraction %d has a nil Denominator", i))
		}
		numM := f.Numerator.Module()
		denM := f.Denominator.Module()
		if numM == nil && denM == nil {
			panic(fmt.Sprintf(
				"wiop: System.NewRationalReduction: fraction %d has no vector-valued expression; "+
					"at least one of Numerator or Denominator must be vector-valued (IsMultiValued() == true)",
				i,
			))
		}
		if numM != nil && denM != nil && numM != denM {
			panic(fmt.Sprintf(
				"wiop: System.NewRationalReduction: fraction %d Numerator module %q and Denominator module %q differ; "+
					"both must share the same module when vector-valued",
				i, numM.Context.Path(), denM.Context.Path(),
			))
		}
		for _, expr := range [2]Expression{f.Numerator, f.Denominator} {
			if r := maxRoundInExpr(expr); r != nil && (maxFracRound == nil || r.ID > maxFracRound.ID) {
				maxFracRound = r
			}
		}
	}

	if maxFracRound == nil {
		panic("wiop: System.NewRationalReduction: no column, cell, or coin found in any fraction expression; " +
			"at least one expression must be round-bearing")
	}

	resultRound, ok := maxFracRound.Next()
	if !ok {
		panic(fmt.Sprintf(
			"wiop: System.NewRationalReduction: no round follows round %d (the latest fraction column round); "+
				"call sys.NewRound() before registering this query",
			maxFracRound.ID,
		))
	}

	result := resultRound.NewCell(ctx.Childf("result"), false)

	rr := &RationalReduction{
		baseQuery: baseQuery{
			context:     ctx,
			Annotations: make(Annotations),
		},
		Kind:      kind,
		Fractions: fractions,
		Result:    result,
	}
	sys.RationalReductions = append(sys.RationalReductions, rr)
	return rr
}
