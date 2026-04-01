package wiop_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/koalabear/field"
	"github.com/consensys/linea-monorepo/prover/wiop"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTable_Basic(t *testing.T) {
	sys, r0, _, mod := newTestSystem(t)
	col := mod.NewColumn(sys.Context.Childf("tblCol"), wiop.VisibilityOracle, r0)
	tab := wiop.NewTable(col.View())
	assert.Equal(t, mod, tab.Module())
	assert.Equal(t, r0, tab.Round())
	assert.Equal(t, 1, tab.Width())
	assert.Nil(t, tab.Selector)
}

func TestNewTable_EmptyPanic(t *testing.T) {
	assert.Panics(t, func() { wiop.NewTable() })
}

func TestNewTable_MixedModulePanic(t *testing.T) {
	sys := wiop.NewSystemf("tblMix")
	r := sys.NewRound()
	mod1 := sys.NewSizedModule(sys.Context.Childf("m1"), 4, wiop.PaddingDirectionNone)
	mod2 := sys.NewSizedModule(sys.Context.Childf("m2"), 4, wiop.PaddingDirectionNone)
	c1 := mod1.NewColumn(sys.Context.Childf("c1"), wiop.VisibilityOracle, r)
	c2 := mod2.NewColumn(sys.Context.Childf("c2"), wiop.VisibilityOracle, r)
	assert.Panics(t, func() { wiop.NewTable(c1.View(), c2.View()) })
}

func TestNewFilteredTable_Basic(t *testing.T) {
	sys, r0, _, mod := newTestSystem(t)
	col := mod.NewColumn(sys.Context.Childf("ftCol"), wiop.VisibilityOracle, r0)
	sel := mod.NewColumn(sys.Context.Childf("ftSel"), wiop.VisibilityOracle, r0)
	tab := wiop.NewFilteredTable(sel.View(), col.View())
	assert.Equal(t, sel.View().Column, tab.Selector.Column)
}

func TestNewFilteredTable_NilSelectorPanic(t *testing.T) {
	sys, r0, _, mod := newTestSystem(t)
	col := mod.NewColumn(sys.Context.Childf("ftNilCol"), wiop.VisibilityOracle, r0)
	assert.Panics(t, func() { wiop.NewFilteredTable(nil, col.View()) })
}

func TestNewFilteredTable_SelectorMixedModulePanic(t *testing.T) {
	sys := wiop.NewSystemf("ftsMix")
	r := sys.NewRound()
	mod1 := sys.NewSizedModule(sys.Context.Childf("m1"), 4, wiop.PaddingDirectionNone)
	mod2 := sys.NewSizedModule(sys.Context.Childf("m2"), 4, wiop.PaddingDirectionNone)
	c := mod1.NewColumn(sys.Context.Childf("c"), wiop.VisibilityOracle, r)
	s := mod2.NewColumn(sys.Context.Childf("s"), wiop.VisibilityOracle, r)
	assert.Panics(t, func() { wiop.NewFilteredTable(s.View(), c.View()) })
}

func TestTableRelationKind_String(t *testing.T) {
	assert.Equal(t, "Permutation", wiop.TableRelationPermutation.String())
	assert.Equal(t, "Inclusion", wiop.TableRelationInclusion.String())
	assert.Contains(t, wiop.TableRelationKind(99).String(), "TableRelationKind")
}

// ---- Permutation ----

func TestPermutation_Check_Match(t *testing.T) {
	sys, r0, _, mod := newTestSystem(t)
	colA := mod.NewColumn(sys.Context.Childf("permA"), wiop.VisibilityOracle, r0)
	colB := mod.NewColumn(sys.Context.Childf("permB"), wiop.VisibilityOracle, r0)

	// A and B are same constant vector → trivially a permutation of each other
	tabA := wiop.NewTable(colA.View())
	tabB := wiop.NewTable(colB.View())
	perm := sys.NewPermutation(sys.Context.Childf("perm"), []wiop.Table{tabA}, []wiop.Table{tabB})
	require.NotNil(t, perm)

	rt := wiop.NewRuntime(sys)
	rt.AssignColumn(colA, baseVec(4, 7))
	rt.AssignColumn(colB, baseVec(4, 7))

	assert.NoError(t, perm.Check(rt))
}

func TestPermutation_Check_Mismatch(t *testing.T) {
	sys, r0, _, mod := newTestSystem(t)
	colA := mod.NewColumn(sys.Context.Childf("permMisA"), wiop.VisibilityOracle, r0)
	colB := mod.NewColumn(sys.Context.Childf("permMisB"), wiop.VisibilityOracle, r0)

	tabA := wiop.NewTable(colA.View())
	tabB := wiop.NewTable(colB.View())
	perm := sys.NewPermutation(sys.Context.Childf("permMis"), []wiop.Table{tabA}, []wiop.Table{tabB})

	rt := wiop.NewRuntime(sys)
	rt.AssignColumn(colA, baseVec(4, 1))
	rt.AssignColumn(colB, baseVec(4, 2)) // different values

	err := perm.Check(rt)
	assert.Error(t, err)
}

func TestPermutation_Round(t *testing.T) {
	sys := wiop.NewSystemf("permRound")
	r0 := sys.NewRound()
	r1 := sys.NewRound()
	mod0 := sys.NewSizedModule(sys.Context.Childf("m0"), 4, wiop.PaddingDirectionNone)
	mod1 := sys.NewSizedModule(sys.Context.Childf("m1"), 4, wiop.PaddingDirectionNone)
	cA := mod0.NewColumn(sys.Context.Childf("cA"), wiop.VisibilityOracle, r0)
	cB := mod1.NewColumn(sys.Context.Childf("cB"), wiop.VisibilityOracle, r1)
	tabA := wiop.NewTable(cA.View())
	tabB := wiop.NewTable(cB.View())
	perm := sys.NewPermutation(sys.Context.Childf("perm"), []wiop.Table{tabA}, []wiop.Table{tabB})
	assert.Equal(t, r1, perm.Round())
}

func TestPermutation_NewPermutation_NilCtxPanic(t *testing.T) {
	sys, r0, _, mod := newTestSystem(t)
	col := mod.NewColumn(sys.Context.Childf("p"), wiop.VisibilityOracle, r0)
	tab := wiop.NewTable(col.View())
	assert.Panics(t, func() { sys.NewPermutation(nil, []wiop.Table{tab}, []wiop.Table{tab}) })
}

func TestPermutation_NewPermutation_EmptyAPanic(t *testing.T) {
	sys, r0, _, mod := newTestSystem(t)
	col := mod.NewColumn(sys.Context.Childf("pEA"), wiop.VisibilityOracle, r0)
	tab := wiop.NewTable(col.View())
	assert.Panics(t, func() {
		sys.NewPermutation(sys.Context.Childf("q"), nil, []wiop.Table{tab})
	})
}

func TestPermutation_NewPermutation_SelectorPanic(t *testing.T) {
	sys, r0, _, mod := newTestSystem(t)
	col := mod.NewColumn(sys.Context.Childf("pSel"), wiop.VisibilityOracle, r0)
	sel := mod.NewColumn(sys.Context.Childf("pSelSel"), wiop.VisibilityOracle, r0)
	tab := wiop.NewFilteredTable(sel.View(), col.View())
	colB := mod.NewColumn(sys.Context.Childf("pSelB"), wiop.VisibilityOracle, r0)
	tabB := wiop.NewTable(colB.View())
	assert.Panics(t, func() {
		sys.NewPermutation(sys.Context.Childf("q"), []wiop.Table{tab}, []wiop.Table{tabB})
	})
}

func TestPermutation_NewPermutation_WidthMismatchPanic(t *testing.T) {
	sys, r0, _, mod := newTestSystem(t)
	c1 := mod.NewColumn(sys.Context.Childf("pW1"), wiop.VisibilityOracle, r0)
	c2 := mod.NewColumn(sys.Context.Childf("pW2"), wiop.VisibilityOracle, r0)
	c3 := mod.NewColumn(sys.Context.Childf("pW3"), wiop.VisibilityOracle, r0)
	tab1 := wiop.NewTable(c1.View())
	tab2 := wiop.NewTable(c2.View(), c3.View()) // width 2 vs 1
	assert.Panics(t, func() {
		sys.NewPermutation(sys.Context.Childf("q"), []wiop.Table{tab1}, []wiop.Table{tab2})
	})
}

// ---- Inclusion ----

func TestInclusion_Check_Match(t *testing.T) {
	sys, r0, _, mod := newTestSystem(t)
	colA := mod.NewColumn(sys.Context.Childf("incA"), wiop.VisibilityOracle, r0)
	colB := mod.NewColumn(sys.Context.Childf("incB"), wiop.VisibilityOracle, r0)

	// A ⊆ B: both are same values → trivially included
	tabA := wiop.NewTable(colA.View())
	tabB := wiop.NewTable(colB.View())
	inc := sys.NewInclusion(sys.Context.Childf("inc"), []wiop.Table{tabA}, []wiop.Table{tabB})
	require.NotNil(t, inc)

	rt := wiop.NewRuntime(sys)
	rt.AssignColumn(colA, baseVec(4, 3))
	rt.AssignColumn(colB, baseVec(4, 3))

	assert.NoError(t, inc.Check(rt))
}

func TestInclusion_Check_Mismatch(t *testing.T) {
	sys, r0, _, mod := newTestSystem(t)
	colA := mod.NewColumn(sys.Context.Childf("incMisA"), wiop.VisibilityOracle, r0)
	colB := mod.NewColumn(sys.Context.Childf("incMisB"), wiop.VisibilityOracle, r0)

	tabA := wiop.NewTable(colA.View())
	tabB := wiop.NewTable(colB.View())
	inc := sys.NewInclusion(sys.Context.Childf("incMis"), []wiop.Table{tabA}, []wiop.Table{tabB})

	rt := wiop.NewRuntime(sys)
	rt.AssignColumn(colA, baseVec(4, 1))
	rt.AssignColumn(colB, baseVec(4, 2)) // A not in B

	err := inc.Check(rt)
	assert.Error(t, err)
}

func TestInclusion_NewInclusion_NilCtxPanic(t *testing.T) {
	sys, r0, _, mod := newTestSystem(t)
	col := mod.NewColumn(sys.Context.Childf("incNilC"), wiop.VisibilityOracle, r0)
	tab := wiop.NewTable(col.View())
	assert.Panics(t, func() { sys.NewInclusion(nil, []wiop.Table{tab}, []wiop.Table{tab}) })
}

func TestInclusion_NewInclusion_EmptyIncludedPanic(t *testing.T) {
	sys, r0, _, mod := newTestSystem(t)
	col := mod.NewColumn(sys.Context.Childf("incEI"), wiop.VisibilityOracle, r0)
	tab := wiop.NewTable(col.View())
	assert.Panics(t, func() {
		sys.NewInclusion(sys.Context.Childf("q"), nil, []wiop.Table{tab})
	})
}

// ---- Permutation with padding (covers padAnchorRow, tableHasZeroShift) ----

func TestPermutation_Check_PaddingRight(t *testing.T) {
	sys := wiop.NewSystemf("permPad")
	r0 := sys.NewRound()
	mod := sys.NewSizedModule(sys.Context.Childf("mod"), 4, wiop.PaddingDirectionRight)
	colA := mod.NewColumn(sys.Context.Childf("colA"), wiop.VisibilityOracle, r0)
	colB := mod.NewColumn(sys.Context.Childf("colB"), wiop.VisibilityOracle, r0)

	tabA := wiop.NewTable(colA.View())
	tabB := wiop.NewTable(colB.View())
	perm := sys.NewPermutation(sys.Context.Childf("perm"), []wiop.Table{tabA}, []wiop.Table{tabB})

	// 2-element data padded to 4 with zeros
	var e field.Element
	e.SetUint64(5)
	data := []field.Element{e, e}
	vecA := &wiop.ConcreteVector{Plain: []field.FieldVec{field.VecFromBase(data)}}
	vecB := &wiop.ConcreteVector{Plain: []field.FieldVec{field.VecFromBase(data)}}

	rt := wiop.NewRuntime(sys)
	rt.AssignColumn(colA, vecA)
	rt.AssignColumn(colB, vecB)

	assert.NoError(t, perm.Check(rt))
}

// ---- Inclusion with selector (covers inclusionBuildSet/inclusionCheckSet selector branch) ----

func TestInclusion_Check_WithSelector(t *testing.T) {
	sys, r0, _, mod := newTestSystem(t)
	colA := mod.NewColumn(sys.Context.Childf("selA"), wiop.VisibilityOracle, r0)
	colB := mod.NewColumn(sys.Context.Childf("selB"), wiop.VisibilityOracle, r0)
	selA := mod.NewColumn(sys.Context.Childf("selAsel"), wiop.VisibilityOracle, r0)

	// selA = [1,0,0,0] → only first row of A is selected
	// A[0] = 1, B is all-1 → first row of A is in B
	tabA := wiop.NewFilteredTable(selA.View(), colA.View())
	tabB := wiop.NewTable(colB.View())
	inc := sys.NewInclusion(sys.Context.Childf("incSel"), []wiop.Table{tabA}, []wiop.Table{tabB})

	rt := wiop.NewRuntime(sys)
	rt.AssignColumn(colA, baseVec(4, 1))
	rt.AssignColumn(colB, baseVec(4, 1))

	// sel = [1,0,0,0]
	var one, zero field.Element
	one.SetUint64(1)
	selVec := &wiop.ConcreteVector{Plain: []field.FieldVec{field.VecFromBase([]field.Element{one, zero, zero, zero})}}
	rt.AssignColumn(selA, selVec)

	assert.NoError(t, inc.Check(rt))
}

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
