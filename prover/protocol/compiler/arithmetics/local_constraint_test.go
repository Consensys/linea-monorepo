package arithmetics_test

import (
	"testing"

	sv "github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/column"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler/arithmetics"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler/dummy"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/stretchr/testify/require"
)

func testCaseFibonnaci() (wizard.DefineFunc, wizard.ProverStep) {

	var (
		P1 ifaces.ColID   = "P1"
		P2 ifaces.ColID   = "P2"
		Q  ifaces.QueryID = "LOCAL"
	)

	definer := func(build *wizard.Builder) {

		// Number of rows
		n := 1 << 3
		P1 := build.RegisterCommit(P1, n)
		P2 := build.RegisterCommit(P2, n)

		// P2(1) = P1(1) + P1(w)
		expr := ifaces.ColumnAsVariable(column.Shift(P1, 1)).
			Add(ifaces.ColumnAsVariable(P1)).
			Sub(ifaces.ColumnAsVariable(P2))

		build.LocalConstraint(Q, expr)
	}

	hLProver := func(assi *wizard.ProverRuntime) {
		p1 := sv.ForTest(1, 1, 2, 3, 5, 8, 13, 21)
		p2 := sv.ForTest(2, 0, 4, 5, 5, 6, 8, 7)
		assi.AssignColumn(P1, p1)
		assi.AssignColumn(P2, p2)
	}

	return definer, hLProver
}

func TestLocalConstraint(t *testing.T) {

	definer, hLProver := testCaseFibonnaci()

	comp := wizard.Compile(
		definer,
		arithmetics.CompileLocal,
		dummy.Compile,
	)

	proof := wizard.Prove(comp, hLProver)
	err := wizard.Verify(comp, proof)
	require.NoError(t, err)
}

func TestLocalConstraintAllLayers(t *testing.T) {

	definer, hLProver := testCaseFibonnaci()

	comp := wizard.Compile(
		definer,
		arithmetics.CompileLocal,
		arithmetics.CompileGlobal,
		dummy.Compile,
	)

	proof := wizard.Prove(comp, hLProver)
	err := wizard.Verify(comp, proof)
	require.NoError(t, err)
}
