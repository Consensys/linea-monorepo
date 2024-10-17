package query_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"

	"github.com/stretchr/testify/require"
)

func TestGlobal(t *testing.T) {
	runTest(t, pythagoreTriplet, true)
	runTest(t, fibonacci, true)
	runTest(t, testDummyShifted, true)
}

func runTest(t *testing.T, gen GlobalConstraintGenerator, expectedCorrect bool) {
	def, assi := gen()
	// Applies the dummy compiler
	compiled := wizard.Compile(def, dummy.Compile)
	proof := wizard.Prove(compiled, assi)
	err := wizard.Verify(compiled, proof)
	if expectedCorrect {
		require.NoError(t, err)
	} else {
		require.Error(t, err)
	}
}

type GlobalConstraintGenerator func() (wizard.DefineFunc, wizard.ProverStep)

/*
No annulator remove the term (X-1)(X-omega) from the constraint. Making it invalid
*/
func fibonacci() (wizard.DefineFunc, wizard.ProverStep) {

	var (
		P ifaces.ColID   = "X"
		Q ifaces.QueryID = "FIBONNACCI"
	)

	definer := func(build *wizard.Builder) {

		// Number of rows
		n := 1 << 3
		P := build.RegisterCommit(P, n) // overshadows P

		expr := ifaces.ColumnAsVariable(column.Shift(P, -1)).
			Add(ifaces.ColumnAsVariable(column.Shift(P, -2))).
			Sub(ifaces.ColumnAsVariable(P))

		build.GlobalConstraint(Q, expr)
	}

	hLProver := func(assi *wizard.ProverRuntime) {
		x := smartvectors.ForTest(1, 1, 2, 3, 5, 8, 13, 21)
		assi.AssignColumn(P, x)
	}

	return definer, hLProver
}

func pythagoreTriplet() (wizard.DefineFunc, wizard.ProverStep) {

	var (
		X ifaces.ColID   = "X"
		Y ifaces.ColID   = "Y"
		Q ifaces.QueryID = "Q"
	)

	define := func(build *wizard.Builder) {

		n := 1 << 2

		X := build.RegisterCommit(X, n) // overshadows P
		Y := build.RegisterCommit(Y, n) // overshadows P

		expr := ifaces.ColumnAsVariable(X).Square().
			Add(ifaces.ColumnAsVariable(Y).Square()).
			Sub(symbolic.NewConstant(25))

		build.GlobalConstraint(Q, expr)
	}

	hLProver := func(assi *wizard.ProverRuntime) {

		x := smartvectors.ForTest(0, 5, 3, 4)
		y := smartvectors.ForTest(5, 0, 4, 3)

		assi.AssignColumn(X, x)
		assi.AssignColumn(Y, y)
	}

	return define, hLProver
}

func testDummyShifted() (wizard.DefineFunc, wizard.ProverStep) {
	var (
		X, Y ifaces.ColID = "X", "Y"
	)
	define := func(build *wizard.Builder) {
		A := build.RegisterCommit(X, 4)
		B := build.RegisterCommit(Y, 4)

		expr := symbolic.Sub(column.Shift(A, 1),
			symbolic.Mul(2, column.Shift(B, 1)))

		build.InsertGlobal(0, "Q", expr)
	}
	Prover := func(run *wizard.ProverRuntime) {
		x := smartvectors.ForTest(2, 8, 4, 0)
		y := smartvectors.ForTest(1, 4, 2, 0)
		run.AssignColumn(X, x)
		run.AssignColumn(Y, y)
	}
	return define, Prover
}
