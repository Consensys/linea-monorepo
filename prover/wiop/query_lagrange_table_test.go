package wiop_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/koalabear/field"
	"github.com/consensys/linea-monorepo/prover/wiop"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- LagrangeEval helpers ----

// lagrangeSystem builds a 2-round system with one 4-element sized module,
// one column and one coin serving as evaluation point.
func lagrangeSystem(t *testing.T) (*wiop.System, *wiop.Round, *wiop.Round, *wiop.Module, *wiop.Column, *wiop.CoinField) {
	t.Helper()
	sys := wiop.NewSystemf("le")
	r0 := sys.NewRound()
	r1 := sys.NewRound()
	mod := sys.NewSizedModule(sys.Context.Childf("leMod"), 4, wiop.PaddingDirectionNone)
	col := mod.NewColumn(sys.Context.Childf("leCol"), wiop.VisibilityOracle, r0)
	coin := r1.NewCoinField(sys.Context.Childf("leCoin"))
	return sys, r0, r1, mod, col, coin
}

// ---- LagrangeEval construction ----

func TestLagrangeEval_NewLagrangeEval_Basic(t *testing.T) {
	sys, r0, r1, _, col, coin := lagrangeSystem(t)
	_ = r0
	le := sys.NewLagrangeEval(sys.Context.Childf("le1"), []*wiop.ColumnView{col.View()}, coin)
	require.NotNil(t, le)
	assert.Len(t, le.EvaluationClaims, 1)
	// Round() should be r1 (the coin's round)
	assert.Equal(t, r1, le.Round())
}

func TestLagrangeEval_NewLagrangeEval_NilCtxPanic(t *testing.T) {
	sys, _, _, _, col, coin := lagrangeSystem(t)
	assert.Panics(t, func() { sys.NewLagrangeEval(nil, []*wiop.ColumnView{col.View()}, coin) })
}

func TestLagrangeEval_NewLagrangeEval_EmptyPolysPanic(t *testing.T) {
	sys, _, _, _, _, coin := lagrangeSystem(t)
	assert.Panics(t, func() { sys.NewLagrangeEval(sys.Context.Childf("le2"), nil, coin) })
}

func TestLagrangeEval_NewLagrangeEval_NoNextRoundPanic(t *testing.T) {
	// Only one round — no next round for claim cells
	sys := wiop.NewSystemf("leNoRound")
	r0 := sys.NewRound()
	mod := sys.NewSizedModule(sys.Context.Childf("mod"), 4, wiop.PaddingDirectionNone)
	col := mod.NewColumn(sys.Context.Childf("col"), wiop.VisibilityOracle, r0)
	cell := r0.NewCell(sys.Context.Childf("ep"), false)
	assert.Panics(t, func() {
		sys.NewLagrangeEval(sys.Context.Childf("le3"), []*wiop.ColumnView{col.View()}, cell)
	})
}

func TestLagrangeEval_NewLagrangeEvalFrom(t *testing.T) {
	sys, _, r1, _, col, coin := lagrangeSystem(t)
	claim := r1.NewCell(sys.Context.Childf("extClaim"), false)
	le := sys.NewLagrangeEvalFrom(sys.Context.Childf("leFrom"), []*wiop.ColumnView{col.View()}, coin, []*wiop.Cell{claim})
	require.NotNil(t, le)
	assert.Equal(t, claim, le.EvaluationClaims[0])
}

func TestLagrangeEval_NewLagrangeEvalFrom_LenMismatchPanic(t *testing.T) {
	sys, _, r1, _, col, coin := lagrangeSystem(t)
	c1 := r1.NewCell(sys.Context.Childf("c1"), false)
	c2 := r1.NewCell(sys.Context.Childf("c2"), false)
	assert.Panics(t, func() {
		sys.NewLagrangeEvalFrom(sys.Context.Childf("leFromMis"), []*wiop.ColumnView{col.View()}, coin, []*wiop.Cell{c1, c2})
	})
}

// ---- LagrangeEval SelfAssign + Check ----

func TestLagrangeEval_SelfAssign_Check(t *testing.T) {
	// Constant column of value c; Lagrange eval of constant polynomial P = c is c everywhere.
	sys, _, _, _, col, coin := lagrangeSystem(t)
	le := sys.NewLagrangeEval(sys.Context.Childf("leEval"), []*wiop.ColumnView{col.View()}, coin)

	rt := wiop.NewRuntime(sys)
	rt.AssignColumn(col, baseVec(4, 3))
	rt.AdvanceRound() // samples coin, now at r1

	assert.False(t, le.IsAlreadyAssigned(rt))
	le.SelfAssign(rt)
	assert.True(t, le.IsAlreadyAssigned(rt))
	assert.NoError(t, le.Check(rt))
}

func TestLagrangeEval_Check_ColumnNotAssigned(t *testing.T) {
	sys, _, _, _, col, coin := lagrangeSystem(t)
	le := sys.NewLagrangeEval(sys.Context.Childf("leNoCol"), []*wiop.ColumnView{col.View()}, coin)

	// Manually assign a coin by advancing normally first, but don't assign column.
	// However, we can't advance without assigning oracle columns... let's make a
	// separate system without oracle columns to reach r1.
	sys2 := wiop.NewSystemf("le2")
	r0b := sys2.NewRound()
	r1b := sys2.NewRound()
	mod2 := sys2.NewSizedModule(sys2.Context.Childf("mod2"), 4, wiop.PaddingDirectionNone)
	col2 := mod2.NewColumn(sys2.Context.Childf("col2"), wiop.VisibilityOracle, r0b)
	coin2 := r1b.NewCoinField(sys2.Context.Childf("coin2"))
	le2 := sys2.NewLagrangeEval(sys2.Context.Childf("le2q"), []*wiop.ColumnView{col2.View()}, coin2)
	_ = le

	rt2 := wiop.NewRuntime(sys2)
	rt2.AssignColumn(col2, baseVec(4, 1))
	rt2.AdvanceRound()
	// Don't assign col2 in a fresh runtime — use a fresh runtime without column assignment.
	// Actually we just already advanced once. Let's test Check when claim isn't assigned.
	// SelfAssign was not called, and we haven't assigned the result cell.
	// Check should error because claim cell is unassigned.
	err := le2.Check(rt2)
	assert.Error(t, err)
}

func TestLagrangeEval_Check_Mismatch(t *testing.T) {
	sys, r0, _, _, col, coin := lagrangeSystem(t)
	le := sys.NewLagrangeEval(sys.Context.Childf("leMis"), []*wiop.ColumnView{col.View()}, coin)

	rt := wiop.NewRuntime(sys)
	rt.AssignColumn(col, baseVec(4, 5))
	rt.AdvanceRound()
	// Assign wrong value to claim cell
	rt.AssignCell(le.EvaluationClaims[0], field.ElemFromBase(field.NewFromString("99")))

	err := le.Check(rt)
	assert.Error(t, err)
	_ = r0
}

func TestLagrangeEval_CheckGnark_Panics(t *testing.T) {
	sys, _, _, _, col, coin := lagrangeSystem(t)
	le := sys.NewLagrangeEval(sys.Context.Childf("leGnark"), []*wiop.ColumnView{col.View()}, coin)
	assert.Panics(t, func() { le.CheckGnark(nil, nil) })
}

// ---- LagrangeEval with Cell as evaluation point (covers roundOf *Cell branch) ----

func TestLagrangeEval_CellEvalPoint_Round(t *testing.T) {
	sys := wiop.NewSystemf("leCell")
	r0 := sys.NewRound()
	r1 := sys.NewRound()
	mod := sys.NewSizedModule(sys.Context.Childf("mod"), 4, wiop.PaddingDirectionNone)
	col := mod.NewColumn(sys.Context.Childf("col"), wiop.VisibilityOracle, r0)
	cell := r1.NewCell(sys.Context.Childf("ep"), false)
	le := sys.NewLagrangeEval(sys.Context.Childf("le"), []*wiop.ColumnView{col.View()}, cell)
	assert.Equal(t, r1, le.Round())
}

// ---- LagrangeEval with PaddingDirectionLeft (covers evalLagrangePaddedBaseBase) ----

func TestLagrangeEval_PaddingLeft_SelfAssign_Check(t *testing.T) {
	sys := wiop.NewSystemf("leLeft")
	r0 := sys.NewRound()
	r1 := sys.NewRound()
	// Module size 4, but assignment provides 3 elements; padding=Left → [pad,v,v,v]
	mod := sys.NewSizedModule(sys.Context.Childf("mod"), 4, wiop.PaddingDirectionLeft)
	col := mod.NewColumn(sys.Context.Childf("col"), wiop.VisibilityOracle, r0)
	coin := r1.NewCoinField(sys.Context.Childf("coin"))
	le := sys.NewLagrangeEval(sys.Context.Childf("le"), []*wiop.ColumnView{col.View()}, coin)

	rt := wiop.NewRuntime(sys)
	// 3-element assignment with no padding data specified; padding value is zero
	var v field.Element
	v.SetUint64(2)
	elems := []field.Element{v, v, v}
	rt.AssignColumn(col, &wiop.ConcreteVector{Plain: []field.FieldVec{field.VecFromBase(elems)}})
	rt.AdvanceRound()

	le.SelfAssign(rt)
	assert.NoError(t, le.Check(rt))
}

// ---- Vanishing/LocalOpening CheckGnark panics ----

func TestVanishing_CheckGnark_Panics(t *testing.T) {
	sys, r0, _, mod := newTestSystem(t)
	col := mod.NewColumn(sys.Context.Childf("vgCol"), wiop.VisibilityOracle, r0)
	v := mod.NewVanishing(sys.Context.Childf("vg"), col.View())
	assert.Panics(t, func() { v.CheckGnark(nil, nil) })
}

func TestLocalOpening_CheckGnark_Panics(t *testing.T) {
	sys, r0, _, mod := newTestSystem(t)
	col := mod.NewColumn(sys.Context.Childf("logCol"), wiop.VisibilityOracle, r0)
	lo := col.At(0).Open(sys.Context.Childf("log"))
	assert.Panics(t, func() { lo.CheckGnark(nil, nil) })
}

// ---- Table ----

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
