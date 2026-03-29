package wiop

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
)

// LocalOpening is a [Query] that asserts the value stored in a prover-supplied
// [Cell] (Result) equals the column assignment at a fixed row index (Pol).
//
// The predicate is: Result == Pol.Column[Pol.Position].
//
// The Result cell lives in the same round as the referenced column: the prover
// commits both the column and the opening value in one step, which is the
// natural structure for a compile-time-fixed opening position.
//
// LocalOpening implements both [GnarkCheckableQuery] and [AssignableQuery]:
//   - SelfAssign reads Column[Position] from the runtime and writes it to Result.
//   - CheckGnark asserts the equality inside a gnark arithmetic circuit.
//
// Use [ColumnPosition.Open] to construct and register an instance.
type LocalOpening struct {
	baseQuery
	// Pol is the column position being opened.
	Pol *ColumnPosition
	// Result is the cell holding the prover's claimed opening value. It lives
	// in the same round as Pol.Column.
	Result *Cell
}

// Round implements [Query]. Returns the round of the referenced column, which
// equals the round of the Result cell.
func (lo *LocalOpening) Round() *Round {
	return lo.Pol.Column.Round()
}

// IsAlreadyAssigned implements [AssignableQuery]. Reports whether the Result
// cell already holds a runtime assignment.
//
// TODO: Implement once Runtime is defined.
func (lo *LocalOpening) IsAlreadyAssigned(_ Runtime) bool {
	panic("wiop: LocalOpening.IsAlreadyAssigned not yet implemented")
}

// SelfAssign implements [AssignableQuery]. Reads Column[Position] from the
// runtime and writes the value into Result.
//
// TODO: Implement once Runtime is defined.
func (lo *LocalOpening) SelfAssign(_ Runtime) {
	panic("wiop: LocalOpening.SelfAssign not yet implemented")
}

// Check implements [Query]. Verifies that Result equals the column assignment
// at Pol.Position.
//
// TODO: Implement once Runtime is defined.
func (lo *LocalOpening) Check(_ Runtime) error {
	panic(fmt.Sprintf("wiop: LocalOpening(%s).Check not yet implemented", lo.Pol.Column.Context.Path()))
}

// CheckGnark implements [GnarkCheckableQuery]. Asserts inside a gnark circuit
// that Result equals the column's gnark variable at Pol.Position.
//
// TODO: Implement once the gnark layer is defined.
func (lo *LocalOpening) CheckGnark(_ frontend.API, _ GnarkRuntime) {
	panic("wiop: LocalOpening.CheckGnark not yet implemented")
}

// Open constructs and registers a [LocalOpening] query for this column
// position. A fresh [Cell] is allocated automatically for the result, placed
// in the same round as the column. The cell's extension flag inherits from
// the parent column.
//
// Panics if ctx or the receiver is nil.
func (pos *ColumnPosition) Open(ctx *ContextFrame) *LocalOpening {
	if pos == nil {
		panic("wiop: ColumnPosition.Open requires a non-nil ColumnPosition")
	}
	if ctx == nil {
		panic("wiop: ColumnPosition.Open requires a non-nil ContextFrame")
	}

	colRound := pos.Column.Round()
	result := colRound.NewCell(ctx.Childf("result"), pos.Column.IsExtension)

	lo := &LocalOpening{
		baseQuery: baseQuery{
			context:     ctx,
			Annotations: make(Annotations),
		},
		Pol:    pos,
		Result: result,
	}

	module := pos.Column.Module
	module.LocalOpenings = append(module.LocalOpenings, lo)
	return lo
}
