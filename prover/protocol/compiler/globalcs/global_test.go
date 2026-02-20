package globalcs_test

import (
	"fmt"
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/accessors"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/globalcs"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/variables"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
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

		expr := symbolic.Sub(P, symbolic.Add(column.Shift(P, -1), column.Shift(P, -2)))
		_ = build.GlobalConstraint(Q, expr)
	}

	comp := wizard.Compile(
		definer,
		globalcs.Compile,
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

func TestPocNewGlobalTestExt(t *testing.T) {
	var (
		P ifaces.ColID   = "P"
		Q ifaces.QueryID = "FIBONNACCI"
	)

	definer := func(build *wizard.Builder) {

		// Number of rows
		n := 1 << 2
		P := build.RegisterCommitExt(P, n) // overshadows P

		// P(X) = P(X/w) + P(X/w^2)

		expr := symbolic.Sub(P, symbolic.Add(column.Shift(P, -1), column.Shift(P, -2)))
		_ = build.GlobalConstraint(Q, expr)
	}

	comp := wizard.Compile(
		definer,
		globalcs.Compile,
		dummy.Compile,
	)

	hLProver := func(assi *wizard.ProverRuntime) {
		x := smartvectors.ForTestFromQuads(
			1, 2, 5, 13,
			1, 3, 8, 21,
			2, 5, 13, 34,
			3, 8, 21, 55)
		assi.AssignColumn(P, x)
	}

	proof := wizard.Prove(comp, hLProver)
	err := wizard.Verify(comp, proof)
	require.NoError(t, err)
}

func TestGlobalWithAccessor(t *testing.T) {

	constant := accessors.NewConstant(field.NewElement(2))

	definer := func(build *wizard.Builder) {

		// Number of rows
		n := 1 << 3
		P := build.RegisterCommit("P", n) // overshadows P

		// P(wX) = 2 * P(X/w)
		expr := symbolic.Sub(
			column.Shift(P, 1),
			symbolic.Mul(constant, P),
		)

		_ = build.GlobalConstraint("Q", expr)
	}

	comp := wizard.Compile(
		definer,
		// globalcs.Compile,
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
		expr := symbolic.Mul(P, variables.NewPeriodicSample(4, 1))
		_ = build.GlobalConstraint("Q", expr)
	}

	comp := wizard.Compile(
		definer,
		globalcs.Compile,
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
		expr := symbolic.Mul(P, variables.NewPeriodicSample(8, 0))
		_ = build.GlobalConstraint("Q", expr)
	}

	comp := wizard.Compile(
		definer,
		globalcs.Compile,
		dummy.Compile,
	)

	prover := func(run *wizard.ProverRuntime) {
		run.AssignColumn("P", smartvectors.ForTest(0, 2, 4, 8, 16, 32, 64, 128))
	}

	proof := wizard.Prove(comp, prover)
	err := wizard.Verify(comp, proof)
	require.NoError(t, err)

}

func TestGlobalDegree3(t *testing.T) {

	definer := func(build *wizard.Builder) {
		// Number of rows
		n := 1 << 3
		P := build.RegisterCommit("P", n)
		P3 := build.RegisterCommit("P3", n)
		// P(X) = P(X/w) + P(X/w^2)
		expr := symbolic.Sub(P3, symbolic.Mul(P, P, P))
		_ = build.GlobalConstraint("Q", expr)
	}

	comp := wizard.Compile(
		definer,
		globalcs.Compile,
		dummy.Compile,
	)

	prover := func(run *wizard.ProverRuntime) {
		run.AssignColumn("P", smartvectors.ForTest(1, 2, 3, 4, 5, 6, 7, 8))
		run.AssignColumn("P3", smartvectors.ForTest(1, 8, 27, 64, 125, 216, 343, 512))
	}

	proof := wizard.Prove(comp, prover)
	err := wizard.Verify(comp, proof)
	require.NoError(t, err)
}

// TestQuotientLargeScale exercises QuotientCtx.Run at a scale similar to the
// real prover. Uses base-field columns so the evaluation goes through
// vmBase.execute (the hot path in production).
func TestQuotientLargeScale(t *testing.T) {

	const (
		domainSize = 1 << 22
		numCols    = 8000 // unique committed columns
	)

	definer := func(build *wizard.Builder) {
		cols := make([]ifaces.Column, numCols)
		for i := 0; i < numCols; i++ {
			cols[i] = build.RegisterCommit(ifaces.ColID(fmt.Sprintf("C_%d", i)), domainSize)
		}

		// Degree-4 constraints: C_i * C_{i+1} * C_{i+2} * C_{i+3} = 0
		// This gives ratio=4, meaning 4 coset evaluations per constraint group.
		for i := 0; i < numCols-3; i++ {
			expr := symbolic.Mul(cols[i], cols[i+1], cols[i+2], cols[i+3])
			build.GlobalConstraint(ifaces.QueryID(fmt.Sprintf("Q_%d", i)), expr)
		}
	}

	t.Log("Compiling...")
	comp := wizard.Compile(
		definer,
		globalcs.Compile,
		dummy.Compile,
	)

	prover := func(run *wizard.ProverRuntime) {
		shared := smartvectors.Rand(domainSize)
		for i := 0; i < numCols; i++ {
			run.AssignColumn(ifaces.ColID(fmt.Sprintf("C_%d", i)), shared)
		}
	}

	t.Log("Proving (quotient computation)...")
	_ = wizard.Prove(comp, prover)
	// Verification is skipped: random witnesses don't satisfy the constraints,
	// but we only care about exercising QuotientCtx.Run via vmBase.
}
