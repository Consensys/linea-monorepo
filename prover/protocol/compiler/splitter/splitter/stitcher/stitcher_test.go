package stitcher_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/accessors"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/column/verifiercol"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/splitter/splitter/stitcher"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalEval(t *testing.T) {

	var a, b, c, d ifaces.Column
	var q1, q2, q3, q4, q5, q6, q7, q8, q9, q10, q11, q12 query.LocalOpening

	define := func(builder *wizard.Builder) {
		// declare columns of different sizes
		a = builder.RegisterCommit("A", 2)
		b = builder.RegisterCommit("B", 4)
		c = builder.RegisterCommit("C", 8)
		d = builder.RegisterCommit("D", 16)

		// Local opening at zero
		q1 = builder.LocalOpening("Q00", a)
		q2 = builder.LocalOpening("Q01", b)
		q3 = builder.LocalOpening("Q02", c)
		q4 = builder.LocalOpening("Q03", d)

		// Local opening at but shifted by one
		q5 = builder.LocalOpening("Q10", column.Shift(a, 1))
		q6 = builder.LocalOpening("Q11", column.Shift(b, 1))
		q7 = builder.LocalOpening("Q12", column.Shift(c, 1))
		q8 = builder.LocalOpening("Q13", column.Shift(d, 1))

		// Local opening at but shifted by one
		q9 = builder.LocalOpening("Q20", column.Shift(a, -1))
		q10 = builder.LocalOpening("Q21", column.Shift(b, -1))
		q11 = builder.LocalOpening("Q22", column.Shift(c, -1))
		q12 = builder.LocalOpening("Q23", column.Shift(d, -1))
	}

	comp := wizard.Compile(define, stitcher.Stitcher(4, 8))

	//after stitcing-compilation we expect that the eligible columns and their relevant queries be ignored
	assert.Equal(t, column.Proof.String(), comp.Columns.Status("A").String())
	assert.Equal(t, column.Ignored.String(), comp.Columns.Status("B").String())
	assert.Equal(t, column.Committed.String(), comp.Columns.Status("C").String())
	assert.Equal(t, column.Committed.String(), comp.Columns.Status("D").String())

	assert.Equal(t, false, comp.QueriesParams.IsIgnored(q1.ID))
	assert.Equal(t, true, comp.QueriesParams.IsIgnored(q2.ID))
	assert.Equal(t, false, comp.QueriesParams.IsIgnored(q3.ID))
	assert.Equal(t, false, comp.QueriesParams.IsIgnored(q4.ID))
	assert.Equal(t, false, comp.QueriesParams.IsIgnored(q5.ID))
	assert.Equal(t, true, comp.QueriesParams.IsIgnored(q6.ID))
	assert.Equal(t, false, comp.QueriesParams.IsIgnored(q7.ID))
	assert.Equal(t, false, comp.QueriesParams.IsIgnored(q8.ID))
	assert.Equal(t, false, comp.QueriesParams.IsIgnored(q9.ID))
	assert.Equal(t, true, comp.QueriesParams.IsIgnored(q10.ID))
	assert.Equal(t, false, comp.QueriesParams.IsIgnored(q11.ID))
	assert.Equal(t, false, comp.QueriesParams.IsIgnored(q12.ID))

	// manually compiles the comp
	dummy.Compile(comp)

	proof := wizard.Prove(comp, func(assi *wizard.ProverRuntime) {
		// Assigns all the columns
		assi.AssignColumn(a.GetColID(), smartvectors.ForTest(0, 1))
		assi.AssignColumn(b.GetColID(), smartvectors.ForTest(2, 3, 4, 5))
		assi.AssignColumn(c.GetColID(), smartvectors.ForTest(6, 7, 8, 9, 10, 11, 12, 13))
		assi.AssignColumn(d.GetColID(), smartvectors.ForTest(15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30))

		// And the alleged results
		assi.AssignLocalPoint("Q00", field.NewElement(0))
		assi.AssignLocalPoint("Q01", field.NewElement(2))
		assi.AssignLocalPoint("Q02", field.NewElement(6))
		assi.AssignLocalPoint("Q03", field.NewElement(15))
		assi.AssignLocalPoint("Q10", field.NewElement(1))
		assi.AssignLocalPoint("Q11", field.NewElement(3))
		assi.AssignLocalPoint("Q12", field.NewElement(7))
		assi.AssignLocalPoint("Q13", field.NewElement(16))
		assi.AssignLocalPoint("Q20", field.NewElement(1))
		assi.AssignLocalPoint("Q21", field.NewElement(5))
		assi.AssignLocalPoint("Q22", field.NewElement(13))
		assi.AssignLocalPoint("Q23", field.NewElement(30))
	})

	err := wizard.Verify(comp, proof)
	require.NoError(t, err)

}

func TestGlobalConstraintFibonacci(t *testing.T) {

	var a, b, c ifaces.Column
	var q1, q2, q3 query.GlobalConstraint

	define := func(builder *wizard.Builder) {
		// declare columns of different sizes
		a = builder.RegisterCommit("B", 4)
		// a = verifiercol.NewConstantCol(field.One(), 4)
		b = builder.RegisterCommit("C", 8)
		c = builder.RegisterCommit("D", 16)

		fibo := func(col ifaces.Column) *symbolic.Expression {
			col_ := ifaces.ColumnAsVariable(col)
			colNext := ifaces.ColumnAsVariable(column.Shift(col, 1))
			colNextNext := ifaces.ColumnAsVariable(column.Shift(col, 2))
			return colNextNext.Sub(colNext).Sub(col_)
		}

		q1 = builder.GlobalConstraint("Q0", fibo(a))
		q2 = builder.GlobalConstraint("Q1", fibo(b))
		q3 = builder.GlobalConstraint("Q2", fibo(c))
	}

	comp := wizard.Compile(define, stitcher.Stitcher(4, 8))

	//after stitcing-compilation we expect that the eligible columns and their relevant queries be ignored
	assert.Equal(t, true, comp.QueriesNoParams.IsIgnored(q1.ID), "q1 should be ignored")
	assert.Equal(t, false, comp.QueriesNoParams.IsIgnored(q2.ID), "q2 should not be ignored")
	assert.Equal(t, false, comp.QueriesNoParams.IsIgnored(q3.ID), "q3 should not be ignored")

	// manually compiles the comp
	dummy.Compile(comp)

	proof := wizard.Prove(comp, func(assi *wizard.ProverRuntime) {
		// Assigns all the columns
		assi.AssignColumn(a.GetColID(), smartvectors.ForTest(1, 1, 2, 3))
		assi.AssignColumn(b.GetColID(), smartvectors.ForTest(1, 1, 2, 3, 5, 8, 13, 21))
		assi.AssignColumn(c.GetColID(), smartvectors.ForTest(1, 1, 2, 3, 5, 8, 13, 21, 34, 55, 89, 144, 233, 377, 610, 987))
	})

	err := wizard.Verify(comp, proof)
	require.NoError(t, err)

}

func TestLocalConstraintFibonacci(t *testing.T) {

	var a, b, c ifaces.Column
	var q1, q2, q3 query.LocalConstraint

	define := func(builder *wizard.Builder) {
		// declare columns of different sizes
		a = builder.RegisterCommit("B", 4)
		b = builder.RegisterCommit("C", 8)
		c = builder.RegisterCommit("D", 16)

		fibo := func(col ifaces.Column) *symbolic.Expression {
			col_ := ifaces.ColumnAsVariable(col)
			colNext := ifaces.ColumnAsVariable(column.Shift(col, 1))
			colNextNext := ifaces.ColumnAsVariable(column.Shift(col, 2))
			return colNextNext.Sub(colNext).Sub(col_)
		}

		q1 = builder.LocalConstraint("Q0", fibo(a))
		q2 = builder.LocalConstraint("Q1", fibo(b))
		q3 = builder.LocalConstraint("Q2", fibo(c))
	}

	comp := wizard.Compile(define, stitcher.Stitcher(4, 8))

	//after stitcing-compilation we expect that the eligible columns and their relevant queries be ignored
	assert.Equal(t, true, comp.QueriesNoParams.IsIgnored(q1.ID), "q1 should be ignored")
	assert.Equal(t, false, comp.QueriesNoParams.IsIgnored(q2.ID), "q2 should not be ignored")
	assert.Equal(t, false, comp.QueriesNoParams.IsIgnored(q3.ID), "q3 should not be ignored")

	// manually compiles the comp
	dummy.Compile(comp)

	proof := wizard.Prove(comp, func(assi *wizard.ProverRuntime) {
		// Assigns all the columns
		// Todo: Arbitrary changes of col values do not make the test failing
		assi.AssignColumn(a.GetColID(), smartvectors.ForTest(1, 1, 2, 3))
		assi.AssignColumn(b.GetColID(), smartvectors.ForTest(1, 1, 2, 3, 5, 8, 13, 21))
		assi.AssignColumn(c.GetColID(), smartvectors.ForTest(1, 1, 2, 3, 5, 8, 13, 21, 34, 55, 89, 144, 233, 377, 610, 987))
	})

	err := wizard.Verify(comp, proof)
	require.NoError(t, err)

}

func TestGlobalMixedRounds(t *testing.T) {

	var a0, a1, a2, b0, b1, b2 ifaces.Column
	var q0, q1, q2 query.LocalConstraint

	define := func(builder *wizard.Builder) {
		// declare columns of different sizes
		a0 = builder.RegisterCommit("A0", 4)
		a1 = builder.RegisterCommit("A1", 4)
		a2 = builder.RegisterCommit("A2", 4)
		_ = builder.RegisterRandomCoin("COIN", coin.Field)
		b0 = builder.RegisterCommit("B0", 4)
		b1 = builder.RegisterCommit("B1", 4)
		b2 = builder.RegisterCommit("B2", 4)

		q0 = builder.LocalConstraint("Q0", ifaces.ColumnAsVariable(a0).Sub(ifaces.ColumnAsVariable(b0)))
		q1 = builder.LocalConstraint("Q1", ifaces.ColumnAsVariable(a1).Sub(ifaces.ColumnAsVariable(b1)))
		q2 = builder.LocalConstraint("Q2", ifaces.ColumnAsVariable(a2).Sub(ifaces.ColumnAsVariable(b2)))
	}

	comp := wizard.Compile(define, stitcher.Stitcher(4, 8))

	//after stitcing-compilation we expect that the eligible columns and their relevant queries be ignored
	assert.Equal(t, true, comp.QueriesNoParams.IsIgnored(q0.ID), "q0 should be ignored")
	assert.Equal(t, true, comp.QueriesNoParams.IsIgnored(q1.ID), "q1 should be ignored")
	assert.Equal(t, true, comp.QueriesNoParams.IsIgnored(q2.ID), "q2 should be ignored")

	// manually compiles the comp
	dummy.Compile(comp)

	proof := wizard.Prove(comp, func(assi *wizard.ProverRuntime) {
		// Assigns all the columns
		assi.AssignColumn(a0.GetColID(), smartvectors.ForTest(1, 1, 2, 3))
		assi.AssignColumn(a1.GetColID(), smartvectors.ForTest(1, 1, 2, 3))
		assi.AssignColumn(a2.GetColID(), smartvectors.ForTest(1, 1, 2, 3))
		_ = assi.GetRandomCoinField("COIN") // triggers going to the next round
		assi.AssignColumn(b0.GetColID(), smartvectors.ForTest(1, 1, 2, 3))
		assi.AssignColumn(b1.GetColID(), smartvectors.ForTest(1, 1, 2, 3))
		assi.AssignColumn(b2.GetColID(), smartvectors.ForTest(1, 1, 2, 3))
	})

	err := wizard.Verify(comp, proof)
	require.NoError(t, err)
}

func TestWithVerifCol(t *testing.T) {
	var a, b, c, verifcol1, verifcol2 ifaces.Column
	var q1, q2 query.GlobalConstraint
	var q3 query.LocalConstraint

	define := func(builder *wizard.Builder) {
		// declare columns of different sizes
		a = builder.RegisterCommit("B", 4)
		b = builder.RegisterCommit("C", 4)
		// a new round
		_ = builder.RegisterRandomCoin("COIN", coin.Field)
		c = builder.RegisterCommit("D", 4)
		// verifiercols
		verifcol1 = verifiercol.NewConstantCol(field.NewElement(3), 4)
		accessors := genAccessors([]int{1, 7, 5, 3})
		verifcol2 = verifiercol.NewFromAccessors(accessors, field.Zero(), 4)

		expr := symbolic.Sub(symbolic.Mul(a, verifcol1), b)
		q1 = builder.GlobalConstraint("Q0", expr)

		expr = symbolic.Sub(symbolic.Add(a, verifcol2), c)
		q2 = builder.GlobalConstraint("Q1", expr)

		q3 = builder.LocalConstraint("Q2", expr)
	}

	comp := wizard.Compile(define, stitcher.Stitcher(4, 8))

	//after stitcing-compilation we expect that the eligible columns and their relevant queries be ignored
	assert.Equal(t, true, comp.QueriesNoParams.IsIgnored(q1.ID), "q1 should be ignored")
	assert.Equal(t, true, comp.QueriesNoParams.IsIgnored(q2.ID), "q2 should  be ignored")
	assert.Equal(t, true, comp.QueriesNoParams.IsIgnored(q3.ID), "q2 should  be ignored")

	// manually compiles the comp
	dummy.Compile(comp)

	proof := wizard.Prove(comp, func(assi *wizard.ProverRuntime) {
		// Assigns all the columns
		assi.AssignColumn(a.GetColID(), smartvectors.ForTest(1, 1, 2, 3))
		assi.AssignColumn(b.GetColID(), smartvectors.ForTest(3, 3, 6, 9))
		_ = assi.GetRandomCoinField("COIN") // triggers going to the next round
		assi.AssignColumn(c.GetColID(), smartvectors.ForTest(2, 8, 7, 6))
	})

	err := wizard.Verify(comp, proof)
	require.NoError(t, err)

}

func genAccessors(a []int) (res []ifaces.Accessor) {
	for i := range a {
		t := accessors.NewConstant(field.NewElement(uint64(a[i])))
		res = append(res, t)
	}
	return res
}
