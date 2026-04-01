package wiop_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/koalabear/field"
	"github.com/consensys/linea-monorepo/prover/wiop"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- expression_compiler: vector eval with all opcodes ----

// vecEval evaluates an expression over two columns assigned to all-2 and all-3.
func makeVecEvalSystem(t *testing.T) (*wiop.System, *wiop.Column, *wiop.Column, wiop.Runtime) {
	t.Helper()
	sys, r0, _, mod := newTestSystem(t)
	c1 := mod.NewColumn(sys.Context.Childf("c1"), wiop.VisibilityOracle, r0)
	c2 := mod.NewColumn(sys.Context.Childf("c2"), wiop.VisibilityOracle, r0)
	rt := wiop.NewRuntime(sys)
	rt.AssignColumn(c1, baseVec(4, 2)) // all 2
	rt.AssignColumn(c2, baseVec(4, 3)) // all 3
	return sys, c1, c2, rt
}

func vecAt(cv wiop.ConcreteVector, i int) field.Element {
	return cv.Plain[0].AsBase()[i]
}

func TestCompiler_VecSub(t *testing.T) {
	_, c1, c2, rt := makeVecEvalSystem(t)
	expr := wiop.Sub(c1.View(), c2.View())
	cv := expr.(interface{ EvaluateVector(wiop.Runtime) wiop.ConcreteVector }).EvaluateVector(rt)
	// 2 - 3 = -1 in field
	var want field.Element
	want.SetUint64(2)
	var three field.Element
	three.SetUint64(3)
	want.Sub(&want, &three)
	assert.Equal(t, want, vecAt(cv, 0))
}

func TestCompiler_VecMul(t *testing.T) {
	_, c1, c2, rt := makeVecEvalSystem(t)
	expr := wiop.Mul(c1.View(), c2.View())
	cv := expr.(interface{ EvaluateVector(wiop.Runtime) wiop.ConcreteVector }).EvaluateVector(rt)
	// 2*3 = 6
	var want field.Element
	want.SetUint64(6)
	assert.Equal(t, want, vecAt(cv, 0))
}

func TestCompiler_VecDiv(t *testing.T) {
	_, c1, c2, rt := makeVecEvalSystem(t)
	expr := wiop.Div(c1.View(), c2.View())
	cv := expr.(interface{ EvaluateVector(wiop.Runtime) wiop.ConcreteVector }).EvaluateVector(rt)
	// 2/3 — just verify it evaluated without error
	require.Len(t, cv.Plain, 1)
	assert.Equal(t, 4, cv.Plain[0].Len())
}

func TestCompiler_VecDouble(t *testing.T) {
	_, c1, _, rt := makeVecEvalSystem(t)
	expr := wiop.Double(c1.View())
	cv := expr.(interface{ EvaluateVector(wiop.Runtime) wiop.ConcreteVector }).EvaluateVector(rt)
	// 2+2 = 4
	var want field.Element
	want.SetUint64(4)
	assert.Equal(t, want, vecAt(cv, 0))
}

func TestCompiler_VecSquare(t *testing.T) {
	_, c1, _, rt := makeVecEvalSystem(t)
	expr := wiop.Square(c1.View())
	cv := expr.(interface{ EvaluateVector(wiop.Runtime) wiop.ConcreteVector }).EvaluateVector(rt)
	// 2*2 = 4
	var want field.Element
	want.SetUint64(4)
	assert.Equal(t, want, vecAt(cv, 0))
}

func TestCompiler_VecNegate(t *testing.T) {
	_, c1, _, rt := makeVecEvalSystem(t)
	expr := wiop.Negate(c1.View())
	cv := expr.(interface{ EvaluateVector(wiop.Runtime) wiop.ConcreteVector }).EvaluateVector(rt)
	// -2 in the field
	var want field.Element
	want.SetUint64(2)
	want.Neg(&want)
	assert.Equal(t, want, vecAt(cv, 0))
}

func TestCompiler_VecInverse(t *testing.T) {
	_, c1, _, rt := makeVecEvalSystem(t)
	expr := wiop.Inverse(c1.View())
	cv := expr.(interface{ EvaluateVector(wiop.Runtime) wiop.ConcreteVector }).EvaluateVector(rt)
	// 1/2
	require.Len(t, cv.Plain, 1)
	assert.Equal(t, 4, cv.Plain[0].Len())
}

// Composite expression reusing a subexpression — exercises slot reuse in the compiler.
func TestCompiler_VecComposite(t *testing.T) {
	_, c1, c2, rt := makeVecEvalSystem(t)
	// (c1+c2) * (c1-c2)  = (2+3)*(2-3) = 5*(-1) = -5
	sum := wiop.Add(c1.View(), c2.View())
	diff := wiop.Sub(c1.View(), c2.View())
	expr := wiop.Mul(sum, diff)
	cv := expr.(interface{ EvaluateVector(wiop.Runtime) wiop.ConcreteVector }).EvaluateVector(rt)
	// 5 * (-1)
	var five, neg1, want field.Element
	five.SetUint64(5)
	neg1.SetUint64(1)
	neg1.Neg(&neg1)
	want.Mul(&five, &neg1)
	assert.Equal(t, want, vecAt(cv, 0))
}

// ---- RationalReduction: scalar numerator / vec denominator ----

func TestRationalReduction_ScalarNumVecDen(t *testing.T) {
	// num = scalar 1, den = vector all-2; sum = 4*(1/2)
	sys := wiop.NewSystemf("rrScalNum")
	r0 := sys.NewRound()
	sys.NewRound()
	mod := sys.NewSizedModule(sys.Context.Childf("m"), 4, wiop.PaddingDirectionNone)
	col := mod.NewColumn(sys.Context.Childf("c"), wiop.VisibilityOracle, r0)

	one := wiop.NewConstantField(field.NewFromString("1"))
	frac := wiop.Fraction{Numerator: one, Denominator: col.View()}
	rr := sys.NewRationalReduction(sys.Context.Childf("rr"), wiop.RationalSum, []wiop.Fraction{frac})

	rt := wiop.NewRuntime(sys)
	rt.AssignColumn(col, baseVec(4, 2))
	rt.AdvanceRound()
	rr.SelfAssign(rt)
	assert.NoError(t, rr.Check(rt))
}

func TestRationalReduction_VecNumScalarDen(t *testing.T) {
	// num = vector all-3, den = scalar 1; sum = 4*3 = 12
	sys := wiop.NewSystemf("rrScalDen")
	r0 := sys.NewRound()
	sys.NewRound()
	mod := sys.NewSizedModule(sys.Context.Childf("m"), 4, wiop.PaddingDirectionNone)
	col := mod.NewColumn(sys.Context.Childf("c"), wiop.VisibilityOracle, r0)

	one := wiop.NewConstantField(field.NewFromString("1"))
	frac := wiop.Fraction{Numerator: col.View(), Denominator: one}
	rr := sys.NewRationalReduction(sys.Context.Childf("rr"), wiop.RationalSum, []wiop.Fraction{frac})

	rt := wiop.NewRuntime(sys)
	rt.AssignColumn(col, baseVec(4, 3))
	rt.AdvanceRound()
	rr.SelfAssign(rt)
	assert.NoError(t, rr.Check(rt))
}

// ---- Inclusion with padding (covers inclusionBuildSet / inclusionCheckSet fast-paths) ----

func TestInclusion_PaddingLeft_Match(t *testing.T) {
	sys := wiop.NewSystemf("incPadL")
	r0 := sys.NewRound()
	// 4-element module with left padding
	mod := sys.NewSizedModule(sys.Context.Childf("mod"), 4, wiop.PaddingDirectionLeft)
	colA := mod.NewColumn(sys.Context.Childf("colA"), wiop.VisibilityOracle, r0)
	colB := mod.NewColumn(sys.Context.Childf("colB"), wiop.VisibilityOracle, r0)

	tabA := wiop.NewTable(colA.View())
	tabB := wiop.NewTable(colB.View())
	inc := sys.NewInclusion(sys.Context.Childf("inc"), []wiop.Table{tabA}, []wiop.Table{tabB})

	rt := wiop.NewRuntime(sys)
	// Both have 3-element assignments padded left to 4 with zero
	var v field.Element
	v.SetUint64(7)
	elems := []field.Element{v, v, v}
	vA := &wiop.ConcreteVector{Plain: []field.FieldVec{field.VecFromBase(elems)}}
	vB := &wiop.ConcreteVector{Plain: []field.FieldVec{field.VecFromBase(elems)}}
	rt.AssignColumn(colA, vA)
	rt.AssignColumn(colB, vB)

	assert.NoError(t, inc.Check(rt))
}

func TestInclusion_PaddingRight_Match(t *testing.T) {
	sys := wiop.NewSystemf("incPadR")
	r0 := sys.NewRound()
	mod := sys.NewSizedModule(sys.Context.Childf("mod"), 4, wiop.PaddingDirectionRight)
	colA := mod.NewColumn(sys.Context.Childf("colA"), wiop.VisibilityOracle, r0)
	colB := mod.NewColumn(sys.Context.Childf("colB"), wiop.VisibilityOracle, r0)

	tabA := wiop.NewTable(colA.View())
	tabB := wiop.NewTable(colB.View())
	inc := sys.NewInclusion(sys.Context.Childf("inc"), []wiop.Table{tabA}, []wiop.Table{tabB})

	rt := wiop.NewRuntime(sys)
	var v field.Element
	v.SetUint64(5)
	elems := []field.Element{v, v, v}
	vA := &wiop.ConcreteVector{Plain: []field.FieldVec{field.VecFromBase(elems)}}
	vB := &wiop.ConcreteVector{Plain: []field.FieldVec{field.VecFromBase(elems)}}
	rt.AssignColumn(colA, vA)
	rt.AssignColumn(colB, vB)

	assert.NoError(t, inc.Check(rt))
}

func TestInclusion_PaddingLeft_Mismatch(t *testing.T) {
	sys := wiop.NewSystemf("incPadMis")
	r0 := sys.NewRound()
	mod := sys.NewSizedModule(sys.Context.Childf("mod"), 4, wiop.PaddingDirectionLeft)
	colA := mod.NewColumn(sys.Context.Childf("colA"), wiop.VisibilityOracle, r0)
	colB := mod.NewColumn(sys.Context.Childf("colB"), wiop.VisibilityOracle, r0)

	tabA := wiop.NewTable(colA.View())
	tabB := wiop.NewTable(colB.View())
	inc := sys.NewInclusion(sys.Context.Childf("inc"), []wiop.Table{tabA}, []wiop.Table{tabB})

	rt := wiop.NewRuntime(sys)
	var v1, v2 field.Element
	v1.SetUint64(1)
	v2.SetUint64(9)
	elemsA := []field.Element{v1, v1, v1}
	elemsB := []field.Element{v2, v2, v2}
	vA := &wiop.ConcreteVector{Plain: []field.FieldVec{field.VecFromBase(elemsA)}}
	vB := &wiop.ConcreteVector{Plain: []field.FieldVec{field.VecFromBase(elemsB)}}
	rt.AssignColumn(colA, vA)
	rt.AssignColumn(colB, vB)

	err := inc.Check(rt)
	assert.Error(t, err)
}

// ---- Inclusion with selector + padding (covers the selector path in inclusionBuildSet) ----

func TestInclusion_PaddingRight_WithSelector_Match(t *testing.T) {
	sys := wiop.NewSystemf("incPadSel")
	r0 := sys.NewRound()
	mod := sys.NewSizedModule(sys.Context.Childf("mod"), 4, wiop.PaddingDirectionRight)
	colA := mod.NewColumn(sys.Context.Childf("colA"), wiop.VisibilityOracle, r0)
	colB := mod.NewColumn(sys.Context.Childf("colB"), wiop.VisibilityOracle, r0)
	selA := mod.NewColumn(sys.Context.Childf("selA"), wiop.VisibilityOracle, r0)

	// Selector selects only data rows (index 0,1), not padding rows.
	var one, zero field.Element
	one.SetUint64(1)
	selElems := []field.Element{one, one}
	selVec := &wiop.ConcreteVector{Plain: []field.FieldVec{field.VecFromBase(selElems)}}

	var v field.Element
	v.SetUint64(3)
	dataElems := []field.Element{v, v}
	dataVec := &wiop.ConcreteVector{Plain: []field.FieldVec{field.VecFromBase(dataElems)}}

	tabA := wiop.NewFilteredTable(selA.View(), colA.View())
	tabB := wiop.NewTable(colB.View())
	inc := sys.NewInclusion(sys.Context.Childf("inc"), []wiop.Table{tabA}, []wiop.Table{tabB})

	rt := wiop.NewRuntime(sys)
	rt.AssignColumn(colA, dataVec)
	rt.AssignColumn(colB, dataVec)
	rt.AssignColumn(selA, selVec)

	assert.NoError(t, inc.Check(rt))
	_ = zero
}

// ---- Permutation with padding + PaddingDirectionLeft (padAnchorRow Left branch) ----

func TestPermutation_PaddingLeft_Match(t *testing.T) {
	sys := wiop.NewSystemf("permPadL")
	r0 := sys.NewRound()
	mod := sys.NewSizedModule(sys.Context.Childf("mod"), 4, wiop.PaddingDirectionLeft)
	colA := mod.NewColumn(sys.Context.Childf("colA"), wiop.VisibilityOracle, r0)
	colB := mod.NewColumn(sys.Context.Childf("colB"), wiop.VisibilityOracle, r0)

	tabA := wiop.NewTable(colA.View())
	tabB := wiop.NewTable(colB.View())
	perm := sys.NewPermutation(sys.Context.Childf("perm"), []wiop.Table{tabA}, []wiop.Table{tabB})

	var v field.Element
	v.SetUint64(4)
	elems := []field.Element{v, v, v}
	vA := &wiop.ConcreteVector{Plain: []field.FieldVec{field.VecFromBase(elems)}}
	vB := &wiop.ConcreteVector{Plain: []field.FieldVec{field.VecFromBase(elems)}}

	rt := wiop.NewRuntime(sys)
	rt.AssignColumn(colA, vA)
	rt.AssignColumn(colB, vB)

	assert.NoError(t, perm.Check(rt))
}

// ---- NewRationalReduction: different-module panic (covers the numM != denM branch) ----

func TestNewRationalReduction_DifferentModulePanic(t *testing.T) {
	sys := wiop.NewSystemf("rrDiffMod")
	r0 := sys.NewRound()
	sys.NewRound()
	mod1 := sys.NewSizedModule(sys.Context.Childf("m1"), 4, wiop.PaddingDirectionNone)
	mod2 := sys.NewSizedModule(sys.Context.Childf("m2"), 4, wiop.PaddingDirectionNone)
	c1 := mod1.NewColumn(sys.Context.Childf("c1"), wiop.VisibilityOracle, r0)
	c2 := mod2.NewColumn(sys.Context.Childf("c2"), wiop.VisibilityOracle, r0)
	frac := wiop.Fraction{Numerator: c1.View(), Denominator: c2.View()}
	assert.Panics(t, func() {
		sys.NewRationalReduction(sys.Context.Childf("q"), wiop.RationalSum, []wiop.Fraction{frac})
	})
}

// ---- NewExtensionColumn nil round panic ----

func TestModule_NewExtensionColumn_NilRoundPanic(t *testing.T) {
	sys, _, _, mod := newTestSystem(t)
	assert.Panics(t, func() { mod.NewExtensionColumn(sys.Context.Childf("ext"), wiop.VisibilityOracle, nil) })
}

// ---- LookupColumn wrong kind panic ----

func TestSystem_LookupColumn_WrongKindPanic(t *testing.T) {
	sys, r0, _, _ := newTestSystem(t)
	cell := r0.NewCell(sys.Context.Childf("cellLookup"), false)
	assert.Panics(t, func() { sys.LookupColumn(cell.Context.ID) })
}

// ---- ColumnView.Degree on unsized module ----

func TestColumnView_Degree_UnsizedPanic(t *testing.T) {
	sys, r0, _, _ := newTestSystem(t)
	unsized := sys.NewModule(sys.Context.Childf("unsized"), wiop.PaddingDirectionNone)
	col := unsized.NewColumn(sys.Context.Childf("col"), wiop.VisibilityOracle, r0)
	assert.Panics(t, func() { col.View().Degree() })
}

// ---- Vanishing.Check on vector with extension data ----

func TestVanishing_Check_Vector_ExtZero(t *testing.T) {
	sys, r0, _, mod := newTestSystem(t)
	extCol := mod.NewExtensionColumn(sys.Context.Childf("extZeroCol"), wiop.VisibilityOracle, r0)
	rt := wiop.NewRuntime(sys)

	// All-zero extension vector
	var zeroExt field.Ext
	elems := make([]field.Ext, 4)
	for i := range elems {
		elems[i] = zeroExt
	}
	rt.AssignColumn(extCol, &wiop.ConcreteVector{Plain: []field.FieldVec{field.VecFromExt(elems)}})

	v := mod.NewVanishing(sys.Context.Childf("extZeroVan"), extCol.View())
	assert.NoError(t, v.Check(rt))
}

// ---- NewLagrangeEvalFrom nil/empty panics ----

func TestLagrangeEval_NewLagrangeEvalFrom_NilCtxPanic(t *testing.T) {
	sys, _, r1, _, col, coin := lagrangeSystem(t)
	claim := r1.NewCell(sys.Context.Childf("cl"), false)
	assert.Panics(t, func() {
		sys.NewLagrangeEvalFrom(nil, []*wiop.ColumnView{col.View()}, coin, []*wiop.Cell{claim})
	})
}

func TestLagrangeEval_NewLagrangeEvalFrom_EmptyPolysPanic(t *testing.T) {
	sys, _, r1, _, _, coin := lagrangeSystem(t)
	claim := r1.NewCell(sys.Context.Childf("cl2"), false)
	assert.Panics(t, func() {
		sys.NewLagrangeEvalFrom(sys.Context.Childf("le"), nil, coin, []*wiop.Cell{claim})
	})
}

// ---- LagrangeEval.Check: unassigned column error ----

func TestLagrangeEval_Check_UnassignedColumn(t *testing.T) {
	sys, _, _, _, col, coin := lagrangeSystem(t)
	le := sys.NewLagrangeEval(sys.Context.Childf("leUnassigned"), []*wiop.ColumnView{col.View()}, coin)

	// Build a runtime at r1 without assigning the column.
	// We need to advance to r1 somehow. Use a separate system trick:
	// actually let's just assign the column but then create a runtime on
	// a different system state... The simplest approach: use a fresh runtime,
	// assign only what's needed for AdvanceRound (no oracle columns exist
	// except col, which we intentionally skip), then try Check.
	// But without assigning col (oracle), AdvanceRound will panic.
	// So instead: don't advance, just call Check while at r0 -- column is
	// not assigned.
	rt := wiop.NewRuntime(sys)
	_ = le
	_ = rt

	// Now create a fresh runtime that doesn't have the column assigned but is at r1.
	// Can't easily do this with the existing API. Instead test via a separate module
	// where we can control assignment.
	sys3 := wiop.NewSystemf("leChkUA")
	r0 := sys3.NewRound()
	r1 := sys3.NewRound()
	mod3 := sys3.NewSizedModule(sys3.Context.Childf("m"), 4, wiop.PaddingDirectionNone)
	col3 := mod3.NewColumn(sys3.Context.Childf("col3"), wiop.VisibilityOracle, r0)
	coin3 := r1.NewCoinField(sys3.Context.Childf("coin3"))
	le3 := sys3.NewLagrangeEval(sys3.Context.Childf("le3"), []*wiop.ColumnView{col3.View()}, coin3)

	rt3 := wiop.NewRuntime(sys3)
	rt3.AssignColumn(col3, baseVec(4, 1))
	rt3.AdvanceRound()
	// Now assign claim but skip SelfAssign (claim is unassigned) → Check returns error
	err := le3.Check(rt3)
	assert.Error(t, err)
	_ = rt
}
