package wiop_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover-ray/maths/koalabear/field"
	"github.com/consensys/linea-monorepo/prover-ray/wiop"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// indexedBaseVec returns a length-n ConcreteVector whose i-th entry is the
// uint64 returned by f(i).
func indexedBaseVec(n int, f func(i int) uint64) *wiop.ConcreteVector {
	elems := make([]field.Element, n)
	for i := range n {
		elems[i].SetUint64(f(i))
	}
	return &wiop.ConcreteVector{Plain: field.VecFromBase(elems)}
}

// ---- happy path ----

func TestLocalConstraint_ColumnAtRowZero_Zero(t *testing.T) {
	sys, r0, _, mod := newTestSystem(t)
	col := mod.NewColumn(sys.Context.Childf("lcZeroCol"), wiop.VisibilityOracle, r0)

	rt := wiop.NewRuntime(sys)
	// col[0] = 0, rest non-zero
	rt.AssignColumn(col, indexedBaseVec(4, func(i int) uint64 {
		if i == 0 {
			return 0
		}
		return uint64(i + 1)
	}))

	v := mod.NewLocalConstraint(sys.Context.Childf("lcZero"), col.View())
	require.NoError(t, v.Check(rt))
}

func TestLocalConstraint_ColumnAtRowZero_NonZero(t *testing.T) {
	sys, r0, _, mod := newTestSystem(t)
	col := mod.NewColumn(sys.Context.Childf("lcNZCol"), wiop.VisibilityOracle, r0)

	rt := wiop.NewRuntime(sys)
	rt.AssignColumn(col, baseVec(4, 7))

	v := mod.NewLocalConstraint(sys.Context.Childf("lcNZ"), col.View())
	err := v.Check(rt)
	assert.Error(t, err)
}

func TestLocalConstraint_ShiftedColumn_PositivePicksRowK(t *testing.T) {
	sys, r0, _, mod := newTestSystem(t)
	col := mod.NewColumn(sys.Context.Childf("lcShiftCol"), wiop.VisibilityOracle, r0)

	rt := wiop.NewRuntime(sys)
	// col = [1, 0, 1, 1]; Shift(1) → looks at col[1] = 0
	rt.AssignColumn(col, indexedBaseVec(4, func(i int) uint64 {
		if i == 1 {
			return 0
		}
		return 1
	}))

	v := mod.NewLocalConstraint(sys.Context.Childf("lcShift"), col.View().Shift(1))
	require.NoError(t, v.Check(rt))
}

func TestLocalConstraint_ShiftedColumn_NegativePicksLastRow(t *testing.T) {
	sys, r0, _, mod := newTestSystem(t)
	col := mod.NewColumn(sys.Context.Childf("lcNegCol"), wiop.VisibilityOracle, r0)

	rt := wiop.NewRuntime(sys)
	// col = [1, 1, 1, 0]; Shift(-1) on size-4 module → col[3] = 0
	rt.AssignColumn(col, indexedBaseVec(4, func(i int) uint64 {
		if i == 3 {
			return 0
		}
		return 1
	}))

	v := mod.NewLocalConstraint(sys.Context.Childf("lcNeg"), col.View().Shift(-1))
	require.NoError(t, v.Check(rt))
}

func TestLocalConstraint_ShiftWrapsAroundModuleSize(t *testing.T) {
	sys, r0, _, mod := newTestSystem(t)
	col := mod.NewColumn(sys.Context.Childf("lcWrapCol"), wiop.VisibilityOracle, r0)

	rt := wiop.NewRuntime(sys)
	// col = [0, 9, 9, 9]; Shift(4) on size-4 module wraps to row 0 = 0
	rt.AssignColumn(col, indexedBaseVec(4, func(i int) uint64 {
		if i == 0 {
			return 0
		}
		return 9
	}))

	v := mod.NewLocalConstraint(sys.Context.Childf("lcWrap"), col.View().Shift(4))
	require.NoError(t, v.Check(rt))
}

func TestLocalConstraint_ArithmeticOverTwoColumns(t *testing.T) {
	sys, r0, _, mod := newTestSystem(t)
	a := mod.NewColumn(sys.Context.Childf("lcA"), wiop.VisibilityOracle, r0)
	b := mod.NewColumn(sys.Context.Childf("lcB"), wiop.VisibilityOracle, r0)

	rt := wiop.NewRuntime(sys)
	// a[0] = b[0] = 5 ⇒ a - b vanishes at row 0
	rt.AssignColumn(a, baseVec(4, 5))
	rt.AssignColumn(b, baseVec(4, 5))

	expr := wiop.Sub(a.View(), b.View())
	v := mod.NewLocalConstraint(sys.Context.Childf("lcArith"), expr)
	require.NoError(t, v.Check(rt))
}

func TestLocalConstraint_ScalarOnly_CellExpression(t *testing.T) {
	sys, r0, _, mod := newTestSystem(t)
	cell := r0.NewCell(sys.Context.Childf("lcCell"), false)

	rt := wiop.NewRuntime(sys)
	rt.AssignCell(cell, field.ElemZero())

	// No columns at all; just `cell` which is zero.
	v := mod.NewLocalConstraint(sys.Context.Childf("lcScalar"), cell)
	require.NoError(t, v.Check(rt))
}

func TestLocalConstraint_VectorConstantCollapsedToScalar(t *testing.T) {
	sys, _, _, mod := newTestSystem(t)
	// A non-zero vector constant; lowering must collapse it to its scalar
	// equivalent so that Check picks the scalar branch.
	c := wiop.NewConstantVector(mod, field.NewFromString("0"))

	rt := wiop.NewRuntime(sys)
	v := mod.NewLocalConstraint(sys.Context.Childf("lcVecK"), c)
	require.NoError(t, v.Check(rt))
}

// ---- registration / metadata ----

func TestLocalConstraint_RegistersOnModuleVanishings(t *testing.T) {
	sys, r0, _, mod := newTestSystem(t)
	col := mod.NewColumn(sys.Context.Childf("lcRegCol"), wiop.VisibilityOracle, r0)

	before := len(mod.Vanishings)
	v := mod.NewLocalConstraint(sys.Context.Childf("lcReg"), col.View())
	assert.Equal(t, before+1, len(mod.Vanishings))
	assert.Same(t, v, mod.Vanishings[before])
	// "Local" constraints never auto-cancel any rows.
	assert.Empty(t, v.CancelledPositions)
}

func TestLocalConstraint_RoundFollowsColumn(t *testing.T) {
	sys, _, r1, mod := newTestSystem(t)
	col := mod.NewColumn(sys.Context.Childf("lcRoundCol"), wiop.VisibilityOracle, r1)
	v := mod.NewLocalConstraint(sys.Context.Childf("lcRound"), col.View())
	assert.Equal(t, r1, v.Round())
}

// ---- panics ----

func TestLocalConstraint_NilCtx_Panics(t *testing.T) {
	_, _, _, mod := newTestSystem(t)
	k := wiop.NewConstantField(field.NewFromString("0"))
	assert.Panics(t, func() { mod.NewLocalConstraint(nil, k) })
}

func TestLocalConstraint_NilExpr_Panics(t *testing.T) {
	sys, _, _, mod := newTestSystem(t)
	assert.Panics(t, func() { mod.NewLocalConstraint(sys.Context.Childf("nilExprLC"), nil) })
}

func TestLocalConstraint_CrossModule_Panics(t *testing.T) {
	sys, r0, _, modA := newTestSystem(t)
	modB := sys.NewSizedModule(sys.Context.Childf("modB"), 4, wiop.PaddingDirectionNone)
	colB := modB.NewColumn(sys.Context.Childf("crossCol"), wiop.VisibilityOracle, r0)
	// Building a local constraint on modA that references a column from modB
	// violates the single-module invariant and must panic at construction.
	assert.Panics(t, func() {
		modA.NewLocalConstraint(sys.Context.Childf("crossLC"), colB.View())
	})
}

func TestLocalConstraint_NegativeShift_UnsizedModule_Panics(t *testing.T) {
	sys := wiop.NewSystemf("lcUnsized")
	r0 := sys.NewRound()
	// Unsized static module: SetSize has not been called.
	mod := sys.NewModule(sys.Context.Childf("unsizedMod"), wiop.PaddingDirectionNone)
	col := mod.NewColumn(sys.Context.Childf("unsizedCol"), wiop.VisibilityOracle, r0)
	assert.Panics(t, func() {
		mod.NewLocalConstraint(sys.Context.Childf("lcNegUnsized"), col.View().Shift(-1))
	})
}
