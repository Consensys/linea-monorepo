package test_cases_test

import (
	"testing"

	"github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/column"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
)

func defineInterleaved(build *wizard.Builder) {
	n := 1 << 2
	P1 := build.RegisterCommit(P1, n) // overshadows P1
	P2 := build.RegisterCommit(P2, n) // overshadows P2
	I := column.Interleave(P1, column.Shift(P2, 2))
	build.Inclusion(LOOKUP1, []ifaces.Column{column.Shift(P1, -1)}, []ifaces.Column{I})
}

func proveInterleaved(run *wizard.ProverRuntime) {
	p1 := smartvectors.ForTest(1, 2, 3, 4)
	p2 := smartvectors.ForTest(1, 2, 2, 1)

	run.AssignColumn(P1, p1)
	run.AssignColumn(P2, p2)
}

func TestInterleaved(t *testing.T) {
	checkSolved(t, defineInterleaved, proveInterleaved, ALL_BUT_ILC, false)
}
