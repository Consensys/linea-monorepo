//go:build !fuzzlight

package test_cases_test

import (
	"testing"

	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/protocol/column"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
)

func defineLocal(build *wizard.Builder) {
	P1 := build.RegisterCommit(P1, 4) // overshadowing
	P2 := build.RegisterCommit(P2, 4) // overshadowing

	expr := ifaces.ColumnAsVariable(column.Shift(P1, 1)).Add(ifaces.ColumnAsVariable(P2))
	build.LocalConstraint(LOCAL1, expr)
}

func proveLocal(run *wizard.ProverRuntime) {
	run.AssignColumn(P1, smartvectors.ForTest(1, 2, 3, 4))
	run.AssignColumn(P2, smartvectors.ForTest(-2, 0, 0, 0))
}

func TestLocal(t *testing.T) {
	checkSolved(t, defineLocal, proveLocal, ALL_BUT_ILC, true)
}
