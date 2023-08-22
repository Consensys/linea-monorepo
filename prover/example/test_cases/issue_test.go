//permutation test is passing for wrong inputs; saying that (1, 1, 2, 3, 4, 5, 8, 10) is a permuation of (0, 0, 0, 0, 0, 0, 0, 0)

package test_cases_test

import (
	"testing"

	"github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
)

func defineIssuePermSingleCol(build *wizard.Builder) {
	/*
		just a basic permutation protocol
	*/
	n := 1 << 3
	P1 := build.RegisterCommit(P1, n) // overshadowing
	P2 := build.RegisterCommit(P2, n) // overshadowing
	build.Permutation(PERMUTATION1, []ifaces.Column{P1}, []ifaces.Column{P2})
}

func proveIssuePermSingleCol(run *wizard.ProverRuntime) {
	run.AssignColumn(P1, smartvectors.ForTest(1, 1, 2, 3, 4, 5, 8, 10))
	run.AssignColumn(P2, smartvectors.ForTest(0, 0, 0, 0, 0, 0, 0, 0))
}

func TestIssuePermSingleCol(t *testing.T) {
	checkSolved(t, defineIssuePermSingleCol, proveIssuePermSingleCol, join(ALL_SPECIALS), false, true)
}
