package wiop_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/koalabear/field"
	"github.com/consensys/linea-monorepo/prover/wiop"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- baseQuery (via Vanishing) ----

func TestBaseQuery_IsReduced_MarkAsReduced(t *testing.T) {
	sys, r0, _, mod := newTestSystem(t)
	col := mod.NewColumn(sys.Context.Childf("bqCol"), wiop.VisibilityOracle, r0)
	v := mod.NewVanishing(sys.Context.Childf("bqV"), col.View())

	assert.False(t, v.IsReduced())
	v.MarkAsReduced()
	assert.True(t, v.IsReduced())
	v.MarkAsReduced() // idempotent
	assert.True(t, v.IsReduced())
}

func TestBaseQuery_Context(t *testing.T) {
	sys, r0, _, mod := newTestSystem(t)
	col := mod.NewColumn(sys.Context.Childf("ctxCol"), wiop.VisibilityOracle, r0)
	ctx := sys.Context.Childf("ctxV")
	v := mod.NewVanishing(ctx, col.View())
	assert.Equal(t, ctx, v.Context())
}

// ---- Vanishing ----

func TestVanishing_Module_Round(t *testing.T) {
	sys, r0, _, mod := newTestSystem(t)
	col := mod.NewColumn(sys.Context.Childf("vanCol"), wiop.VisibilityOracle, r0)
	v := mod.NewVanishing(sys.Context.Childf("van"), col.View())

	assert.Equal(t, mod, v.Module())
	assert.Equal(t, r0, v.Round())
}

func TestVanishing_String(t *testing.T) {
	sys, r0, _, mod := newTestSystem(t)
	col := mod.NewColumn(sys.Context.Childf("strCol"), wiop.VisibilityOracle, r0)
	ctx := sys.Context.Childf("strV")
	v := mod.NewVanishing(ctx, col.View())
	s := v.String()
	assert.Contains(t, s, "Vanishing")
	assert.Contains(t, s, "strV")
}

func TestVanishing_Check_Scalar_Zero(t *testing.T) {
	// scalar constant zero vanishes
	sys, r0, _, _ := newTestSystem(t)
	rt := wiop.NewRuntime(sys)
	cell := r0.NewCell(sys.Context.Childf("scZero"), false)
	rt.AssignCell(cell, field.ElemZero())

	// zero - zero == 0
	expr := wiop.Sub(cell, cell)
	mod := sys.NewSizedModule(sys.Context.Childf("modScZero"), 1, wiop.PaddingDirectionNone)
	v := mod.NewVanishing(sys.Context.Childf("scVanish"), expr)
	assert.NoError(t, v.Check(rt))
}

func TestVanishing_Check_Scalar_NonZero(t *testing.T) {
	sys, r0, _, _ := newTestSystem(t)
	rt := wiop.NewRuntime(sys)
	cell := r0.NewCell(sys.Context.Childf("scNZ"), false)
	rt.AssignCell(cell, field.ElemFromBase(field.NewFromString("1")))

	mod := sys.NewSizedModule(sys.Context.Childf("modScNZ"), 1, wiop.PaddingDirectionNone)
	v := mod.NewVanishing(sys.Context.Childf("scNZV"), cell)
	err := v.Check(rt)
	assert.Error(t, err)
}

func TestVanishing_Check_Vector_AllZero(t *testing.T) {
	sys, r0, _, mod := newTestSystem(t)
	col := mod.NewColumn(sys.Context.Childf("allZCol"), wiop.VisibilityOracle, r0)
	rt := wiop.NewRuntime(sys)
	rt.AssignColumn(col, baseVec(4, 0))

	v := mod.NewVanishing(sys.Context.Childf("allZVan"), col.View())
	assert.NoError(t, v.Check(rt))
}

func TestVanishing_Check_Vector_NonZero(t *testing.T) {
	sys, r0, _, mod := newTestSystem(t)
	col := mod.NewColumn(sys.Context.Childf("nzVCol"), wiop.VisibilityOracle, r0)
	rt := wiop.NewRuntime(sys)
	rt.AssignColumn(col, baseVec(4, 1))

	v := mod.NewVanishing(sys.Context.Childf("nzVVan"), col.View())
	err := v.Check(rt)
	assert.Error(t, err)
}

func TestVanishing_Check_Vector_CancelledPositions(t *testing.T) {
	// col = [1,1,1,1]; cancel rows 0,1,2,3 → effectively all cancelled → should pass
	sys, r0, _, mod := newTestSystem(t)
	col := mod.NewColumn(sys.Context.Childf("cancelCol"), wiop.VisibilityOracle, r0)
	rt := wiop.NewRuntime(sys)
	rt.AssignColumn(col, baseVec(4, 1))

	// negative indices: -4 = row 0, -3 = row 1, -2 = row 2, -1 = row 3
	v := mod.NewVanishingManual(sys.Context.Childf("cancelVan"), col.View(), -4, -3, -2, -1)
	assert.NoError(t, v.Check(rt))
}

func TestVanishing_NewVanishing_NilCtxPanic(t *testing.T) {
	_, _, _, mod := newTestSystem(t)
	k := wiop.NewConstantField(field.NewFromString("0"))
	assert.Panics(t, func() { mod.NewVanishing(nil, k) })
}

func TestVanishing_NewVanishing_NilExprPanic(t *testing.T) {
	sys, _, _, mod := newTestSystem(t)
	assert.Panics(t, func() { mod.NewVanishing(sys.Context.Childf("nilExpr"), nil) })
}

func TestVanishing_NewVanishingManual_NilCtxPanic(t *testing.T) {
	_, _, _, mod := newTestSystem(t)
	k := wiop.NewConstantField(field.NewFromString("0"))
	assert.Panics(t, func() { mod.NewVanishingManual(nil, k) })
}

func TestVanishing_NewVanishingManual_NilExprPanic(t *testing.T) {
	sys, _, _, mod := newTestSystem(t)
	assert.Panics(t, func() { mod.NewVanishingManual(sys.Context.Childf("nilExpr2"), nil) })
}

func TestVanishing_AutoCancelledPositions_PositiveShift(t *testing.T) {
	// shift +2 → last 2 rows cancelled: -2, -1
	sys, r0, _, mod := newTestSystem(t)
	col := mod.NewColumn(sys.Context.Childf("shiftPosCol"), wiop.VisibilityOracle, r0)
	// col.View().Shift(2) triggers auto-cancel of last 2 rows
	v := mod.NewVanishing(sys.Context.Childf("shiftPosV"), col.View().Shift(2))
	assert.Len(t, v.CancelledPositions, 2)
}

func TestVanishing_AutoCancelledPositions_NegativeShift(t *testing.T) {
	// shift -1 → first row cancelled: 0
	sys, r0, _, mod := newTestSystem(t)
	col := mod.NewColumn(sys.Context.Childf("shiftNegCol"), wiop.VisibilityOracle, r0)
	v := mod.NewVanishing(sys.Context.Childf("shiftNegV"), col.View().Shift(-1))
	assert.Len(t, v.CancelledPositions, 1)
	assert.Equal(t, 0, v.CancelledPositions[0])
}

// ---- LocalOpening ----

func TestLocalOpening_Round_IsAlreadyAssigned_SelfAssign_Check(t *testing.T) {
	sys, r0, _, mod := newTestSystem(t)
	col := mod.NewColumn(sys.Context.Childf("loCol"), wiop.VisibilityOracle, r0)
	lo := col.At(2).Open(sys.Context.Childf("lo"))

	require.NotNil(t, lo)
	assert.Equal(t, r0, lo.Round())

	rt := wiop.NewRuntime(sys)
	// Assign [0,1,2,3]; lo picks index 2 → expected 2
	elems := make([]field.Element, 4)
	for i := range 4 {
		elems[i].SetUint64(uint64(i))
	}
	rt.AssignColumn(col, &wiop.ConcreteVector{Plain: []field.FieldVec{field.VecFromBase(elems)}})

	assert.False(t, lo.IsAlreadyAssigned(rt))
	lo.SelfAssign(rt)
	assert.True(t, lo.IsAlreadyAssigned(rt))

	assert.NoError(t, lo.Check(rt))
}

func TestLocalOpening_Check_Mismatch(t *testing.T) {
	sys, r0, _, mod := newTestSystem(t)
	col := mod.NewColumn(sys.Context.Childf("loMisCol"), wiop.VisibilityOracle, r0)
	lo := col.At(1).Open(sys.Context.Childf("loMis"))

	rt := wiop.NewRuntime(sys)
	rt.AssignColumn(col, baseVec(4, 5))
	// manually assign wrong value to result cell
	rt.AssignCell(lo.Result, field.ElemFromBase(field.NewFromString("9")))

	err := lo.Check(rt)
	assert.Error(t, err)
}

func TestLocalOpening_Check_ColumnNotAssigned(t *testing.T) {
	sys, r0, _, mod := newTestSystem(t)
	col := mod.NewColumn(sys.Context.Childf("loUnassigned"), wiop.VisibilityOracle, r0)
	lo := col.At(0).Open(sys.Context.Childf("loUnassignedQ"))

	rt := wiop.NewRuntime(sys)
	// don't assign column
	rt.AssignCell(lo.Result, field.ElemZero())
	err := lo.Check(rt)
	assert.Error(t, err)
}

func TestColumnPosition_Open_NilReceiverPanic(t *testing.T) {
	sys := wiop.NewSystemf("s")
	sys.NewRound()
	var cp *wiop.ColumnPosition
	assert.Panics(t, func() { cp.Open(sys.Context.Childf("p")) })
}

func TestColumnPosition_Open_NilCtxPanic(t *testing.T) {
	sys, r0, _, mod := newTestSystem(t)
	col := mod.NewColumn(sys.Context.Childf("openNilCtx"), wiop.VisibilityOracle, r0)
	assert.Panics(t, func() { col.At(0).Open(nil) })
}

// ---- RationalReduction ----

func TestRationalReduction_Sum(t *testing.T) {
	// 4-row column of all-2; denominator = constant 1; sum = 4*2 = 8
	sys := wiop.NewSystemf("rrSys")
	r0 := sys.NewRound()
	r1 := sys.NewRound() // result cell goes here
	_ = r1
	mod := sys.NewSizedModule(sys.Context.Childf("rrMod"), 4, wiop.PaddingDirectionNone)
	col := mod.NewColumn(sys.Context.Childf("rrCol"), wiop.VisibilityOracle, r0)

	one := wiop.NewConstantVector(mod, field.NewFromString("1"))
	frac := wiop.Fraction{Numerator: col.View(), Denominator: one}
	rr := sys.NewRationalReduction(sys.Context.Childf("rrQ"), wiop.RationalSum, []wiop.Fraction{frac})
	require.NotNil(t, rr)

	rt := wiop.NewRuntime(sys)
	rt.AssignColumn(col, baseVec(4, 2))
	rt.AdvanceRound()

	assert.False(t, rr.IsAlreadyAssigned(rt))
	rr.SelfAssign(rt)
	assert.True(t, rr.IsAlreadyAssigned(rt))

	assert.NoError(t, rr.Check(rt))
	assert.Equal(t, r1, rr.Round())
}

func TestRationalReduction_Product(t *testing.T) {
	// 4-row column of all-2; denominator = constant 1; product = 2^4 = 16
	sys := wiop.NewSystemf("rrProdSys")
	r0 := sys.NewRound()
	sys.NewRound()
	mod := sys.NewSizedModule(sys.Context.Childf("rrProdMod"), 4, wiop.PaddingDirectionNone)
	col := mod.NewColumn(sys.Context.Childf("rrProdCol"), wiop.VisibilityOracle, r0)

	one := wiop.NewConstantVector(mod, field.NewFromString("1"))
	frac := wiop.Fraction{Numerator: col.View(), Denominator: one}
	rr := sys.NewRationalReduction(sys.Context.Childf("rrProdQ"), wiop.RationalProduct, []wiop.Fraction{frac})

	rt := wiop.NewRuntime(sys)
	rt.AssignColumn(col, baseVec(4, 2))
	rt.AdvanceRound()
	rr.SelfAssign(rt)
	assert.NoError(t, rr.Check(rt))
}

func TestRationalReduction_Check_Mismatch(t *testing.T) {
	sys := wiop.NewSystemf("rrMisSys")
	r0 := sys.NewRound()
	sys.NewRound()
	mod := sys.NewSizedModule(sys.Context.Childf("rrMisMod"), 4, wiop.PaddingDirectionNone)
	col := mod.NewColumn(sys.Context.Childf("rrMisCol"), wiop.VisibilityOracle, r0)

	one := wiop.NewConstantVector(mod, field.NewFromString("1"))
	frac := wiop.Fraction{Numerator: col.View(), Denominator: one}
	rr := sys.NewRationalReduction(sys.Context.Childf("rrMisQ"), wiop.RationalSum, []wiop.Fraction{frac})

	rt := wiop.NewRuntime(sys)
	rt.AssignColumn(col, baseVec(4, 1))
	rt.AdvanceRound()
	// assign wrong value
	rt.AssignCell(rr.Result, field.ElemFromBase(field.NewFromString("99")))

	err := rr.Check(rt)
	assert.Error(t, err)
}

func TestRationalReductionKind_String(t *testing.T) {
	assert.Equal(t, "Sum", wiop.RationalSum.String())
	assert.Equal(t, "Product", wiop.RationalProduct.String())
	assert.Contains(t, wiop.RationalReductionKind(99).String(), "RationalReductionKind")
}

func TestNewRationalReduction_NilCtxPanic(t *testing.T) {
	sys := wiop.NewSystemf("s")
	sys.NewRound()
	sys.NewRound()
	mod := sys.NewSizedModule(sys.Context.Childf("m"), 4, wiop.PaddingDirectionNone)
	r := sys.Rounds[0]
	col := mod.NewColumn(sys.Context.Childf("c"), wiop.VisibilityOracle, r)
	one := wiop.NewConstantVector(mod, field.NewFromString("1"))
	frac := wiop.Fraction{Numerator: col.View(), Denominator: one}
	assert.Panics(t, func() { sys.NewRationalReduction(nil, wiop.RationalSum, []wiop.Fraction{frac}) })
}

func TestNewRationalReduction_EmptyFractionsPanic(t *testing.T) {
	sys := wiop.NewSystemf("s")
	sys.NewRound()
	assert.Panics(t, func() {
		sys.NewRationalReduction(sys.Context.Childf("q"), wiop.RationalSum, nil)
	})
}

func TestNewRationalReduction_NilNumeratorPanic(t *testing.T) {
	sys := wiop.NewSystemf("s")
	r0 := sys.NewRound()
	sys.NewRound()
	mod := sys.NewSizedModule(sys.Context.Childf("m"), 4, wiop.PaddingDirectionNone)
	col := mod.NewColumn(sys.Context.Childf("c"), wiop.VisibilityOracle, r0)
	frac := wiop.Fraction{Numerator: nil, Denominator: col.View()}
	assert.Panics(t, func() {
		sys.NewRationalReduction(sys.Context.Childf("q"), wiop.RationalSum, []wiop.Fraction{frac})
	})
}

func TestNewRationalReduction_BothScalarPanic(t *testing.T) {
	sys := wiop.NewSystemf("s")
	sys.NewRound()
	sys.NewRound()
	k := wiop.NewConstantField(field.NewFromString("1"))
	frac := wiop.Fraction{Numerator: k, Denominator: k}
	assert.Panics(t, func() {
		sys.NewRationalReduction(sys.Context.Childf("q"), wiop.RationalSum, []wiop.Fraction{frac})
	})
}

func TestNewRationalReduction_NoNextRoundPanic(t *testing.T) {
	// Only one round — no next round for result cell
	sys := wiop.NewSystemf("s")
	r0 := sys.NewRound()
	mod := sys.NewSizedModule(sys.Context.Childf("m"), 4, wiop.PaddingDirectionNone)
	col := mod.NewColumn(sys.Context.Childf("c"), wiop.VisibilityOracle, r0)
	one := wiop.NewConstantVector(mod, field.NewFromString("1"))
	frac := wiop.Fraction{Numerator: col.View(), Denominator: one}
	assert.Panics(t, func() {
		sys.NewRationalReduction(sys.Context.Childf("q"), wiop.RationalSum, []wiop.Fraction{frac})
	})
}

// ---- maxRoundInExpr (via Vanishing.Round) ----

func TestMaxRoundInExpr_MultipleRounds(t *testing.T) {
	sys := wiop.NewSystemf("s")
	r0 := sys.NewRound()
	r1 := sys.NewRound()
	mod0 := sys.NewSizedModule(sys.Context.Childf("m0"), 4, wiop.PaddingDirectionNone)
	mod1 := sys.NewSizedModule(sys.Context.Childf("m1"), 4, wiop.PaddingDirectionNone)
	col0 := mod0.NewColumn(sys.Context.Childf("c0"), wiop.VisibilityOracle, r0)
	col1 := mod1.NewColumn(sys.Context.Childf("c1"), wiop.VisibilityOracle, r1)

	// Add two vector expressions from different modules is valid structurally;
	// Vanishing.Round walks the tree and picks the max.
	expr := wiop.Add(col0.View(), col1.View())
	// Use a scratch module just to hang the Vanishing on
	mod := sys.NewSizedModule(sys.Context.Childf("scratch"), 4, wiop.PaddingDirectionNone)
	v := mod.NewVanishing(sys.Context.Childf("v"), expr)
	assert.Equal(t, r1, v.Round())
}
