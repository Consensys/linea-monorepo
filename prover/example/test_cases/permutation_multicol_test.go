//go:build !fuzzlight

package test_cases_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

func definePermutationMultiCol(build *wizard.Builder) {
	n := 8
	/*
		just a basic permutation protocol
	*/
	P1 := build.RegisterCommit(P1, n) // overshadows
	P2 := build.RegisterCommit(P2, n) // overshadows
	P3 := build.RegisterCommit(P3, n) // overshadows
	P4 := build.RegisterCommit(P4, n) // overshadows
	build.Permutation(PERMUTATION1, []ifaces.Column{P1, P2}, []ifaces.Column{P3, P4})
}

func provePermutationMultiCol(run *wizard.ProverRuntime) {
	run.AssignColumn(P1, smartvectors.ForTest(1, 2, 3, 4, 5, 6, 7, 8))
	run.AssignColumn(P2, smartvectors.ForTest(10, 11, 12, 13, 14, 15, 16, 17))
	run.AssignColumn(P3, smartvectors.ForTest(7, 8, 6, 5, 4, 3, 2, 1))
	run.AssignColumn(P4, smartvectors.ForTest(16, 17, 15, 14, 13, 12, 11, 10))
}

func TestPermutationMultiCol(t *testing.T) {
	checkSolved(t, definePermutationMultiCol, provePermutationMultiCol, ALL_BUT_ILC, true)
}
