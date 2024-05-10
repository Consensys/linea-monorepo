//go:build !fuzzlight

package test_cases_test

import (
	"testing"

	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
)

func defineIncluMultiCol(build *wizard.Builder) {
	/*
		just a basic Plookup protocol
	*/
	P1 := build.RegisterCommit(P1, 8) // overshadowing
	P2 := build.RegisterCommit(P2, 8) // overshadowing
	P3 := build.RegisterCommit(P3, 4) // overshadowing
	P4 := build.RegisterCommit(P4, 4) // overshadowing
	build.Inclusion(LOOKUP1, []ifaces.Column{P1, P2}, []ifaces.Column{P3, P4})
}

func proveIncluMultiCol(run *wizard.ProverRuntime) {
	run.AssignColumn(P1, smartvectors.ForTest(0, 1, 2, 3, 4, 5, 6, 7))
	run.AssignColumn(P2, smartvectors.ForTest(0, 1, 4, 9, 16, 25, 36, 49))
	run.AssignColumn(P3, smartvectors.ForTest(5, 6, 1, 4))
	run.AssignColumn(P4, smartvectors.ForTest(25, 36, 1, 16))
}

func TestInclusionMultiCol(t *testing.T) {
	checkSolved(t, defineIncluMultiCol, proveIncluMultiCol, ALL_BUT_ILC, false)
}
