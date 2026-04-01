package wiop_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/koalabear/field"
	"github.com/consensys/linea-monorepo/prover/wiop"
	"github.com/stretchr/testify/assert"
)

// evalWithBaseZ builds a system where the evaluation point is a *Cell whose
// assigned value is a base-field element. This triggers the z.IsBase() branches
// in evalLagrangePadded.
func evalWithBaseZ(t *testing.T, padding wiop.PaddingDirection, ext bool) {
	t.Helper()
	sys := wiop.NewSystemf("padTest")
	r0 := sys.NewRound()
	r1 := sys.NewRound()
	mod := sys.NewSizedModule(sys.Context.Childf("mod"), 4, padding)

	var col *wiop.Column
	if ext {
		col = mod.NewExtensionColumn(sys.Context.Childf("col"), wiop.VisibilityOracle, r0)
	} else {
		col = mod.NewColumn(sys.Context.Childf("col"), wiop.VisibilityOracle, r0)
	}

	// Evaluation point is a Cell holding a base-field value → z.IsBase() = true.
	ep := r1.NewCell(sys.Context.Childf("ep"), false)
	le := sys.NewLagrangeEval(sys.Context.Childf("le"), []*wiop.ColumnView{col.View()}, ep)

	rt := wiop.NewRuntime(sys)

	if ext {
		// Extension column: assign extension elements all equal to 2.
		var e field.Ext
		var two field.Element
		two.SetUint64(2)
		e.B0.A0 = two
		elems := make([]field.Ext, 3)
		for i := range elems {
			elems[i] = e
		}
		rt.AssignColumn(col, &wiop.ConcreteVector{Plain: []field.FieldVec{field.VecFromExt(elems)}})
	} else {
		// Base column: assign 3 elements all equal to 5, padded to 4.
		var v field.Element
		v.SetUint64(5)
		elems := []field.Element{v, v, v}
		rt.AssignColumn(col, &wiop.ConcreteVector{Plain: []field.FieldVec{field.VecFromBase(elems)}})
	}

	// Advance to r1 so we can assign the Cell in r1.
	rt.AdvanceRound()

	// Assign a base-field value to the evaluation-point cell (now at r1, so valid).
	var z field.Element
	z.SetUint64(3)
	rt.AssignCell(ep, field.ElemFromBase(z))

	le.SelfAssign(rt)
	assert.NoError(t, le.Check(rt))
}

// TestEvalLagrangePaddedBaseBase: base data, left padding, base z.
func TestEvalLagrangePaddedBaseBase(t *testing.T) {
	evalWithBaseZ(t, wiop.PaddingDirectionLeft, false)
}

// TestEvalLagrangePaddedExtBase: extension data, right padding, base z.
func TestEvalLagrangePaddedExtBase(t *testing.T) {
	evalWithBaseZ(t, wiop.PaddingDirectionRight, true)
}

// TestEvalLagrangePaddedExtExt: extension data, left padding, extension z (coin).
func TestEvalLagrangePaddedExtExt(t *testing.T) {
	sys := wiop.NewSystemf("extExt")
	r0 := sys.NewRound()
	r1 := sys.NewRound()
	mod := sys.NewSizedModule(sys.Context.Childf("mod"), 4, wiop.PaddingDirectionLeft)
	col := mod.NewExtensionColumn(sys.Context.Childf("col"), wiop.VisibilityOracle, r0)
	// Coin gives an extension element → z.IsBase() = false.
	coin := r1.NewCoinField(sys.Context.Childf("coin"))
	le := sys.NewLagrangeEval(sys.Context.Childf("le"), []*wiop.ColumnView{col.View()}, coin)

	rt := wiop.NewRuntime(sys)

	var e field.Ext
	var two field.Element
	two.SetUint64(2)
	e.B0.A0 = two
	elems := make([]field.Ext, 3)
	for i := range elems {
		elems[i] = e
	}
	rt.AssignColumn(col, &wiop.ConcreteVector{Plain: []field.FieldVec{field.VecFromExt(elems)}})
	rt.AdvanceRound()

	le.SelfAssign(rt)
	assert.NoError(t, le.Check(rt))
}

// TestEvalPolynomials_Shift: covers the k != 0 branch in evalPolynomials.
func TestEvalPolynomials_Shift(t *testing.T) {
	sys := wiop.NewSystemf("shift")
	r0 := sys.NewRound()
	r1 := sys.NewRound()
	mod := sys.NewSizedModule(sys.Context.Childf("mod"), 4, wiop.PaddingDirectionNone)
	col := mod.NewColumn(sys.Context.Childf("col"), wiop.VisibilityOracle, r0)
	coin := r1.NewCoinField(sys.Context.Childf("coin"))

	// Shifted view: ShiftingOffset = 1
	shifted := col.View().Shift(1)
	le := sys.NewLagrangeEval(sys.Context.Childf("le"), []*wiop.ColumnView{shifted}, coin)

	rt := wiop.NewRuntime(sys)
	rt.AssignColumn(col, baseVec(4, 5))
	rt.AdvanceRound()

	le.SelfAssign(rt)
	assert.NoError(t, le.Check(rt))
}
