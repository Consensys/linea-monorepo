package wiop

import "fmt"

// NewLocalConstraint registers a [Vanishing] that is enforced only at logical
// row 0 — a "local constraint" in the sense of prover/protocol/query.LocalConstraint.
//
// Every column reference in expr is interpreted as the column's value at row
// 0. A reference of the form col.View().Shift(k) (a [*ColumnView] with
// ShiftingOffset k) becomes the column's value at row (k mod n), where n is
// the module size; this matches the prover-side convention that "to evaluate
// a local constraint at a different point, the column must be shifted first".
//
// Internally, every [*ColumnView] in expr is rewritten to a [*ColumnPosition]
// and every vector [*Constant] is collapsed to its scalar form. The lowered
// expression is necessarily scalar, so the returned [*Vanishing] is checked
// through the scalar branch of [Vanishing.Check]. This keeps a single Query
// type for both "global" (multi-valued) and "local" (scalar) vanishing
// predicates.
//
// All columns referenced in expr must belong to m. Negative shifts require m
// to be statically sized: for an unsized or dynamic module the row-0 position
// cannot be normalised at construction time and a negative shift would not
// fit into a non-negative [ColumnPosition.Position].
//
// Panics if ctx or expr is nil, if any column reference belongs to a
// different module, or if a negative shift is used on an unsized/dynamic
// module.
func (m *Module) NewLocalConstraint(ctx *ContextFrame, expr Expression) *Vanishing {
	if ctx == nil {
		panic("wiop: Module.NewLocalConstraint requires a non-nil ContextFrame")
	}
	if expr == nil {
		panic("wiop: Module.NewLocalConstraint requires a non-nil Expression")
	}
	scalar := m.lowerToRowZero(expr)
	if scalar.IsMultiValued() {
		panic(fmt.Sprintf(
			"wiop: Module.NewLocalConstraint(%s): lowered expression is unexpectedly multi-valued",
			ctx.Path(),
		))
	}
	return m.newVanishing(ctx, scalar, nil)
}

// lowerToRowZero traverses expr bottom-up via [EditExpression] and produces
// the scalar expression that the prover's LocalConstraint would evaluate.
// The rewrite rules are:
//
//   - [*ColumnView]{Column: c, ShiftingOffset: k}
//     → [*ColumnPosition]{Column: c, Position: (k mod n)}.
//
//   - vector [*Constant]{Value: v, module: !=nil}
//     → scalar [*Constant]{Value: v, module: nil} (the same value at any row).
//
//   - every other leaf ([*Cell], [*CoinField], [*ColumnPosition], scalar
//     [*Constant]) is passed through unchanged.
//
//   - [*ArithmeticOperation] is rebuilt with the rewritten operands.
//
// All [*ColumnView] and [*ColumnPosition] leaves must reference columns
// belonging to m; otherwise the call panics.
func (m *Module) lowerToRowZero(expr Expression) Expression {
	return EditExpression(expr, func(curr Expression, newChildren []Expression) Expression {
		switch e := curr.(type) {
		case *ColumnView:
			m.assertOwnsColumn(e.Column)
			return columnViewToRowZeroPosition(e)
		case *ColumnPosition:
			m.assertOwnsColumn(e.Column)
			return e
		case *Constant:
			if e.module != nil {
				return NewConstantField(e.Value)
			}
			return e
		default:
			return DefaultConstruct(curr, newChildren)
		}
	})
}

// assertOwnsColumn panics if col is not owned by m. Used during local-constraint
// lowering to enforce the single-module invariant of [Vanishing].
func (m *Module) assertOwnsColumn(col *Column) {
	if col.Module == m {
		return
	}
	panic(fmt.Sprintf(
		"wiop: Module.NewLocalConstraint: column %q belongs to module %q, not %q",
		col.Context.Path(), col.Module.Context.Path(), m.Context.Path(),
	))
}

// columnViewToRowZeroPosition converts a [*ColumnView] into the
// [*ColumnPosition] it would evaluate to at logical row 0. The returned
// position lies in [0, n) when the parent module is sized. For an unsized or
// dynamic module the offset is passed through directly; a negative offset is
// rejected upfront because it cannot be normalised without knowing n.
func columnViewToRowZeroPosition(cv *ColumnView) *ColumnPosition {
	off := cv.ShiftingOffset
	m := cv.Column.Module
	if m.IsSized() {
		n := m.Size()
		off = ((off % n) + n) % n
	} else if off < 0 {
		panic(fmt.Sprintf(
			"wiop: NewLocalConstraint: negative shift %d on column %q requires a sized module",
			cv.ShiftingOffset, cv.Column.Context.Path(),
		))
	}
	return &ColumnPosition{Column: cv.Column, Position: off}
}
