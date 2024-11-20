//go:build !fuzzlight

package test_cases_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/globalcs"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
)

func defineFibo(build *wizard.Builder) {

	// Number of rows
	n := 1 << 3
	p1 := build.RegisterCommit(P1, n) // overshadows P

	// P(X) = P(X/w) + P(X/w^2)
	expr := sym.Sub(
		p1,
		column.Shift(p1, -1),
		column.Shift(p1, -2),
	)

	_ = build.GlobalConstraint(GLOBAL1, expr)
}

func proveFibo(run *wizard.ProverRuntime) {
	x := smartvectors.ForTest(1, 1, 2, 3, 5, 8, 13, 21)
	run.AssignColumn(P1, x)
}

func TestFibo(t *testing.T) {
	checkSolved(t, defineFibo, proveFibo, join(compilationSuite{globalcs.Compile}, DUMMY), true)
}
