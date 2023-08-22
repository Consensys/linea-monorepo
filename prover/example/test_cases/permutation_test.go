package test_cases_test

import (
	"testing"

	"github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
)

func definePermSingleCol(build *wizard.Builder) {
	/*
		just a basic permutation protocol
	*/
	n := 1 << 4
	P1 := build.RegisterCommit(P1, n) // overshadowing
	P2 := build.RegisterCommit(P2, n) // overshadowing
	build.Permutation(PERMUTATION1, []ifaces.Column{P1}, []ifaces.Column{P2})
}

func provePermSingleCol(run *wizard.ProverRuntime) {
	run.AssignColumn(P1, smartvectors.ForTest(1, 1, 2, 3, 4, 5, 8, 0, 1, 1, 2, 3, 4, 5, 8, 0))
	run.AssignColumn(P2, smartvectors.ForTest(8, 1, 2, 1, 3, 0, 4, 5, 1, 1, 2, 3, 4, 5, 8, 0))
}

func TestPermSingleCol(t *testing.T) {
	checkSolved(t, definePermSingleCol, provePermSingleCol, ALL_BUT_ILC, true)
}
