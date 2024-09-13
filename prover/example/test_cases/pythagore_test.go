//go:build !fuzzlight

package test_cases_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
)

func definePythagore(build *wizard.Builder) {
	n := 1 << 2

	P1 := build.RegisterCommit(P1, n) // overshadows P
	P2 := build.RegisterCommit(P2, n) // overshadows P

	expr := ifaces.ColumnAsVariable(P1).Square().
		Add(ifaces.ColumnAsVariable(P2).Square()).
		Sub(symbolic.NewConstant(25))

	build.GlobalConstraint(GLOBAL1, expr)
}

func provePythagore(run *wizard.ProverRuntime) {
	x := smartvectors.ForTest(0, 5, 3, 4)
	y := smartvectors.ForTest(5, 0, 4, 3)

	run.AssignColumn(P1, x)
	run.AssignColumn(P2, y)
}

func TestPythagore(t *testing.T) {
	checkSolved(t, definePythagore, provePythagore, DUMMY, true)
}
