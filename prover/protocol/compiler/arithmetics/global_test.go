package arithmetics_test

import (
	"testing"

	"github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/accessors"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/column"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler/arithmetics"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler/dummy"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/variables"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/stretchr/testify/require"
)

func TestPocNewGlobalTest(t *testing.T) {
	var (
		P ifaces.ColID   = "P"
		Q ifaces.QueryID = "FIBONNACCI"
	)

	definer := func(build *wizard.Builder) {

		// Number of rows
		n := 1 << 3
		P := build.RegisterCommit(P, n) // overshadows P

		// P(X) = P(X/w) + P(X/w^2)

		expr := ifaces.ColumnAsVariable(column.Shift(P, -1)).
			Add(ifaces.ColumnAsVariable(column.Shift(P, -2))).
			Sub(ifaces.ColumnAsVariable(P))

		_ = build.GlobalConstraint(Q, expr)
	}

	comp := wizard.Compile(
		definer,
		arithmetics.CompileGlobal,
		dummy.Compile,
	)

	hLProver := func(assi *wizard.ProverRuntime) {
		x := smartvectors.ForTest(1, 1, 2, 3, 5, 8, 13, 21)
		assi.AssignColumn(P, x)
	}

	proof := wizard.Prove(comp, hLProver)
	err := wizard.Verify(comp, proof)
	require.NoError(t, err)
}

func TestGlobalWithAccessor(t *testing.T) {

	constant := accessors.AccessorFromConstant(field.NewElement(2))

	definer := func(build *wizard.Builder) {

		// Number of rows
		n := 1 << 3
		P := build.RegisterCommit("P", n) // overshadows P

		// P(X) = P(X/w) + P(X/w^2)
		expr := ifaces.ColumnAsVariable(column.Shift(P, 1)).
			Sub(ifaces.ColumnAsVariable(P).Mul(constant.AsVariable()))

		_ = build.GlobalConstraint("Q", expr)
	}

	comp := wizard.Compile(
		definer,
		arithmetics.CompileGlobal,
		dummy.Compile,
	)

	prover := func(run *wizard.ProverRuntime) {
		run.AssignColumn("P", smartvectors.ForTest(1, 2, 4, 8, 16, 32, 64, 128))
	}

	proof := wizard.Prove(comp, prover)
	err := wizard.Verify(comp, proof)
	require.NoError(t, err)
}

func TestPeriodicSampleGlobalConstraint(t *testing.T) {

	definer := func(build *wizard.Builder) {

		// Number of rows
		n := 1 << 3
		P := build.RegisterCommit("P", n) // overshadows P

		// P(X) = P(X/w) + P(X/w^2)
		expr := ifaces.ColumnAsVariable(P).Mul(variables.NewPeriodicSample(4, 1))

		_ = build.GlobalConstraint("Q", expr)
	}

	comp := wizard.Compile(
		definer,
		arithmetics.CompileGlobal,
		dummy.Compile,
	)

	prover := func(run *wizard.ProverRuntime) {
		run.AssignColumn("P", smartvectors.ForTest(1, 0, 4, 8, 16, 0, 64, 128))
	}

	proof := wizard.Prove(comp, prover)
	err := wizard.Verify(comp, proof)
	require.NoError(t, err)

}

func TestPeriodicSampleAsLagrange(t *testing.T) {

	definer := func(build *wizard.Builder) {

		// Number of rows
		n := 1 << 3
		P := build.RegisterCommit("P", n) // overshadows P

		// P(X) = P(X/w) + P(X/w^2)
		expr := ifaces.ColumnAsVariable(P).Mul(variables.NewPeriodicSample(8, 0))

		_ = build.GlobalConstraint("Q", expr)
	}

	comp := wizard.Compile(
		definer,
		arithmetics.CompileGlobal,
		dummy.Compile,
	)

	prover := func(run *wizard.ProverRuntime) {
		run.AssignColumn("P", smartvectors.ForTest(0, 2, 4, 8, 16, 32, 64, 128))
	}

	proof := wizard.Prove(comp, prover)
	err := wizard.Verify(comp, proof)
	require.NoError(t, err)

}
