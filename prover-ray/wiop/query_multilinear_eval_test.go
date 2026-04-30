package wiop_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover-ray/maths/koalabear/field"
	"github.com/consensys/linea-monorepo/prover-ray/maths/koalabear/polynomials"
	"github.com/consensys/linea-monorepo/prover-ray/wiop"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mleSystem builds a two-round system with a sized module of size 2^logN
// (must be a power of two) and one base-field column, returning the coin
// slice for the evaluation coordinates.
func mleSystem(t *testing.T, logN int) (
	sys *wiop.System,
	r0, r1 *wiop.Round,
	mod *wiop.Module,
	col *wiop.Column,
	coins []*wiop.CoinField,
) {
	t.Helper()
	sys = wiop.NewSystemf("mle")
	r0 = sys.NewRound()
	r1 = sys.NewRound()
	mod = sys.NewSizedModule(sys.Context.Childf("mleMod"), 1<<logN, wiop.PaddingDirectionNone)
	col = mod.NewColumn(sys.Context.Childf("mleCol"), wiop.VisibilityOracle, r0)
	coins = make([]*wiop.CoinField, logN)
	for i := range coins {
		coins[i] = r1.NewCoinField(sys.Context.Childf("mleCoin[%d]", i))
	}
	return
}

// coinsToPromises converts []*CoinField to []FieldPromise.
func coinsToPromises(coins []*wiop.CoinField) []wiop.FieldPromise {
	fps := make([]wiop.FieldPromise, len(coins))
	for i, c := range coins {
		fps[i] = c
	}
	return fps
}

// ---- Construction ----

func TestMultilinearEval_Construction_Basic(t *testing.T) {
	sys, _, r1, _, col, coins := mleSystem(t, 3)
	me := sys.NewMultilinearEval(sys.Context.Childf("me"), []*wiop.ColumnView{col.View()}, coinsToPromises(coins))
	require.NotNil(t, me)
	assert.Len(t, me.EvaluationClaims, 1)
	assert.True(t, me.EvaluationClaims[0].IsExtension())
	assert.Equal(t, r1, me.Round())
}

func TestMultilinearEval_Construction_NilCtxPanic(t *testing.T) {
	sys, _, _, _, col, coins := mleSystem(t, 2)
	assert.Panics(t, func() {
		sys.NewMultilinearEval(nil, []*wiop.ColumnView{col.View()}, coinsToPromises(coins))
	})
}

func TestMultilinearEval_Construction_EmptyPolysPanic(t *testing.T) {
	sys, _, _, _, _, coins := mleSystem(t, 2)
	assert.Panics(t, func() {
		sys.NewMultilinearEval(sys.Context.Childf("me"), nil, coinsToPromises(coins))
	})
}

func TestMultilinearEval_Construction_NoNextRoundPanic(t *testing.T) {
	// Only one round — no next round for claim cells.
	sys := wiop.NewSystemf("mleOneRound")
	r0 := sys.NewRound()
	mod := sys.NewSizedModule(sys.Context.Childf("mod"), 4, wiop.PaddingDirectionNone)
	col := mod.NewColumn(sys.Context.Childf("col"), wiop.VisibilityOracle, r0)
	coin := r0.NewCoinField(sys.Context.Childf("coin"))
	assert.Panics(t, func() {
		sys.NewMultilinearEval(sys.Context.Childf("me"), []*wiop.ColumnView{col.View()}, []wiop.FieldPromise{coin, coin})
	})
}

func TestMultilinearEval_Construction_ShiftedViewPanic(t *testing.T) {
	sys, _, _, _, col, coins := mleSystem(t, 2)
	shifted := col.View().Shift(1)
	assert.Panics(t, func() {
		sys.NewMultilinearEval(sys.Context.Childf("me"), []*wiop.ColumnView{shifted}, coinsToPromises(coins))
	})
}

func TestMultilinearEval_Construction_WrongNbCoordsPanic(t *testing.T) {
	// Column size 4 requires log₂(4)=2 coords; provide only 1.
	sys, _, _, _, col, _ := mleSystem(t, 2)
	sys.NewRound() // add a round so claims can be placed
	r1 := sys.Rounds[1]
	coin := r1.NewCoinField(sys.Context.Childf("onlyCoin"))
	assert.Panics(t, func() {
		sys.NewMultilinearEval(sys.Context.Childf("me"), []*wiop.ColumnView{col.View()}, []wiop.FieldPromise{coin})
	})
}

func TestMultilinearEval_Construction_SizeMismatchPanic(t *testing.T) {
	// Two columns of different sizes.
	sys := wiop.NewSystemf("mleMismatch")
	r0 := sys.NewRound()
	sys.NewRound()
	r1 := sys.Rounds[1]
	mod4 := sys.NewSizedModule(sys.Context.Childf("mod4"), 4, wiop.PaddingDirectionNone)
	mod8 := sys.NewSizedModule(sys.Context.Childf("mod8"), 8, wiop.PaddingDirectionNone)
	col4 := mod4.NewColumn(sys.Context.Childf("col4"), wiop.VisibilityOracle, r0)
	col8 := mod8.NewColumn(sys.Context.Childf("col8"), wiop.VisibilityOracle, r0)
	c1 := r1.NewCoinField(sys.Context.Childf("c1"))
	c2 := r1.NewCoinField(sys.Context.Childf("c2"))
	assert.Panics(t, func() {
		sys.NewMultilinearEval(
			sys.Context.Childf("me"),
			[]*wiop.ColumnView{col4.View(), col8.View()},
			[]wiop.FieldPromise{c1, c2},
		)
	})
}

func TestMultilinearEval_NewMultilinearEvalFrom_Basic(t *testing.T) {
	sys, _, r1, _, col, coins := mleSystem(t, 2)
	extClaim := r1.NewCell(sys.Context.Childf("extClaim"), true)
	me := sys.NewMultilinearEvalFrom(
		sys.Context.Childf("meFrom"),
		[]*wiop.ColumnView{col.View()},
		coinsToPromises(coins),
		[]*wiop.Cell{extClaim},
	)
	require.NotNil(t, me)
	assert.Equal(t, extClaim, me.EvaluationClaims[0])
}

func TestMultilinearEval_NewMultilinearEvalFrom_LenMismatchPanic(t *testing.T) {
	sys, _, r1, _, col, coins := mleSystem(t, 2)
	c1 := r1.NewCell(sys.Context.Childf("c1"), true)
	c2 := r1.NewCell(sys.Context.Childf("c2"), true)
	assert.Panics(t, func() {
		sys.NewMultilinearEvalFrom(
			sys.Context.Childf("meMis"),
			[]*wiop.ColumnView{col.View()},
			coinsToPromises(coins),
			[]*wiop.Cell{c1, c2},
		)
	})
}

// ---- SelfAssign + Check (base column, extension-field coins) ----

func TestMultilinearEval_SelfAssign_Check_Base(t *testing.T) {
	const logN = 3
	sys, _, _, _, col, coins := mleSystem(t, logN)
	me := sys.NewMultilinearEval(sys.Context.Childf("me"), []*wiop.ColumnView{col.View()}, coinsToPromises(coins))

	// Build a random table.
	elems := make([]field.Element, 1<<logN)
	for i := range elems {
		elems[i].SetUint64(uint64(i + 1))
	}

	rt := wiop.NewRuntime(sys)
	rt.AssignColumn(col, &wiop.ConcreteVector{Plain: field.VecFromBase(elems)})
	rt.AdvanceRound() // derives coin values

	assert.False(t, me.IsAlreadyAssigned(rt))
	me.SelfAssign(rt)
	assert.True(t, me.IsAlreadyAssigned(rt))
	require.NoError(t, me.Check(rt))
}

// ---- SelfAssign matches EvalMultilin ----

func TestMultilinearEval_AgainstEvalMultilin(t *testing.T) {
	const logN = 4
	sys, _, _, _, col, coins := mleSystem(t, logN)
	me := sys.NewMultilinearEval(sys.Context.Childf("me"), []*wiop.ColumnView{col.View()}, coinsToPromises(coins))

	n := 1 << logN
	elems := make([]field.Element, n)
	for i := range elems {
		elems[i].SetUint64(uint64(i * i))
	}

	rt := wiop.NewRuntime(sys)
	rt.AssignColumn(col, &wiop.ConcreteVector{Plain: field.VecFromBase(elems)})
	rt.AdvanceRound()

	me.SelfAssign(rt)
	require.NoError(t, me.Check(rt))

	// Compare the claim value against EvalMultilin using the same coin values.
	coords := make([]field.Gen, logN)
	for i, c := range coins {
		coords[i] = rt.GetCoinValue(c)
	}
	want := polynomials.EvalMultilin(field.VecFromBase(elems), coords)
	got := rt.GetCellValue(me.EvaluationClaims[0])
	diff := got.Sub(want)
	assert.True(t, diff.IsZero(), "claim does not match EvalMultilin: got %v want %v", got, want)
}

// ---- Batch (multiple polynomials) ----

func TestMultilinearEval_Batch(t *testing.T) {
	const logN = 3
	sys := wiop.NewSystemf("mleBatch")
	r0 := sys.NewRound()
	r1 := sys.NewRound()
	mod := sys.NewSizedModule(sys.Context.Childf("mod"), 1<<logN, wiop.PaddingDirectionNone)
	col0 := mod.NewColumn(sys.Context.Childf("col0"), wiop.VisibilityOracle, r0)
	col1 := mod.NewColumn(sys.Context.Childf("col1"), wiop.VisibilityOracle, r0)

	coins := make([]*wiop.CoinField, logN)
	for i := range coins {
		coins[i] = r1.NewCoinField(sys.Context.Childf("coin[%d]", i))
	}

	me := sys.NewMultilinearEval(
		sys.Context.Childf("meBatch"),
		[]*wiop.ColumnView{col0.View(), col1.View()},
		coinsToPromises(coins),
	)
	assert.Len(t, me.EvaluationClaims, 2)

	n := 1 << logN
	elems0 := make([]field.Element, n)
	elems1 := make([]field.Element, n)
	for i := range elems0 {
		elems0[i].SetUint64(uint64(i + 1))
		elems1[i].SetUint64(uint64(n - i))
	}

	rt := wiop.NewRuntime(sys)
	rt.AssignColumn(col0, &wiop.ConcreteVector{Plain: field.VecFromBase(elems0)})
	rt.AssignColumn(col1, &wiop.ConcreteVector{Plain: field.VecFromBase(elems1)})
	rt.AdvanceRound()

	me.SelfAssign(rt)
	require.NoError(t, me.Check(rt))
}

// ---- Padding ----

func TestMultilinearEval_PaddingRight(t *testing.T) {
	const logN = 2 // module size 4, provide 3 elements → right-padded
	sys := wiop.NewSystemf("mlePadRight")
	r0 := sys.NewRound()
	r1 := sys.NewRound()
	mod := sys.NewSizedModule(sys.Context.Childf("mod"), 1<<logN, wiop.PaddingDirectionRight)
	col := mod.NewColumn(sys.Context.Childf("col"), wiop.VisibilityOracle, r0)
	coins := make([]*wiop.CoinField, logN)
	for i := range coins {
		coins[i] = r1.NewCoinField(sys.Context.Childf("coin[%d]", i))
	}

	me := sys.NewMultilinearEval(sys.Context.Childf("me"), []*wiop.ColumnView{col.View()}, coinsToPromises(coins))

	var v field.Element
	v.SetUint64(7)
	elems := []field.Element{v, v, v} // 3 elements; 4th will be padding zero

	rt := wiop.NewRuntime(sys)
	rt.AssignColumn(col, &wiop.ConcreteVector{Plain: field.VecFromBase(elems)})
	rt.AdvanceRound()

	me.SelfAssign(rt)
	require.NoError(t, me.Check(rt))
}

func TestMultilinearEval_PaddingLeft(t *testing.T) {
	const logN = 2
	sys := wiop.NewSystemf("mlePadLeft")
	r0 := sys.NewRound()
	r1 := sys.NewRound()
	mod := sys.NewSizedModule(sys.Context.Childf("mod"), 1<<logN, wiop.PaddingDirectionLeft)
	col := mod.NewColumn(sys.Context.Childf("col"), wiop.VisibilityOracle, r0)
	coins := make([]*wiop.CoinField, logN)
	for i := range coins {
		coins[i] = r1.NewCoinField(sys.Context.Childf("coin[%d]", i))
	}

	me := sys.NewMultilinearEval(sys.Context.Childf("me"), []*wiop.ColumnView{col.View()}, coinsToPromises(coins))

	var v field.Element
	v.SetUint64(3)
	elems := []field.Element{v, v, v}

	rt := wiop.NewRuntime(sys)
	rt.AssignColumn(col, &wiop.ConcreteVector{Plain: field.VecFromBase(elems)})
	rt.AdvanceRound()

	me.SelfAssign(rt)
	require.NoError(t, me.Check(rt))
}

// ---- Check error paths ----

func TestMultilinearEval_Check_ColumnUnassigned(t *testing.T) {
	sys, _, _, _, col, coins := mleSystem(t, 2)
	me := sys.NewMultilinearEval(sys.Context.Childf("me"), []*wiop.ColumnView{col.View()}, coinsToPromises(coins))

	rt := wiop.NewRuntime(sys)
	// Column not assigned → Check must error immediately.
	err := me.Check(rt)
	assert.Error(t, err)
}

func TestMultilinearEval_Check_ClaimUnassigned(t *testing.T) {
	sys, _, _, _, col, coins := mleSystem(t, 2)
	me := sys.NewMultilinearEval(sys.Context.Childf("me"), []*wiop.ColumnView{col.View()}, coinsToPromises(coins))

	rt := wiop.NewRuntime(sys)
	rt.AssignColumn(col, baseVec(4, 1))
	rt.AdvanceRound()
	// SelfAssign not called → claim cell unassigned → Check must error.
	err := me.Check(rt)
	assert.Error(t, err)
}

func TestMultilinearEval_Check_WrongClaim(t *testing.T) {
	sys, _, _, _, col, coins := mleSystem(t, 2)
	me := sys.NewMultilinearEval(sys.Context.Childf("me"), []*wiop.ColumnView{col.View()}, coinsToPromises(coins))

	rt := wiop.NewRuntime(sys)
	rt.AssignColumn(col, baseVec(4, 5))
	rt.AdvanceRound()

	// Assign a wrong value to the claim cell.
	rt.AssignCell(me.EvaluationClaims[0], field.ElemFromBase(field.NewFromString("99")))
	err := me.Check(rt)
	assert.Error(t, err)
}

func TestMultilinearEval_CheckGnark_Panics(t *testing.T) {
	sys, _, _, _, col, coins := mleSystem(t, 2)
	me := sys.NewMultilinearEval(sys.Context.Childf("me"), []*wiop.ColumnView{col.View()}, coinsToPromises(coins))
	assert.Panics(t, func() { me.CheckGnark(nil, nil) })
}

// ---- System registration ----

func TestMultilinearEval_RegisteredInSystem(t *testing.T) {
	sys, _, _, _, col, coins := mleSystem(t, 2)
	me := sys.NewMultilinearEval(sys.Context.Childf("me"), []*wiop.ColumnView{col.View()}, coinsToPromises(coins))
	require.Len(t, sys.MultilinearEvals, 1)
	assert.Equal(t, me, sys.MultilinearEvals[0])
}
