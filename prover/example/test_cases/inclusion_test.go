//go:build !fuzzlight

package test_cases_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

func defineInclu(build *wizard.Builder) {
	n := 1 << 2

	P1 := build.RegisterCommit(P1, n)   // overshadows P1
	P2 := build.RegisterCommit(P2, n*2) // overshadows P2

	build.Inclusion(LOOKUP1, []ifaces.Column{P1}, []ifaces.Column{P2})
}

func proveInclu(run *wizard.ProverRuntime) {
	p1 := smartvectors.ForTest(1, 2, 3, 4)
	p2 := smartvectors.ForTest(2, 1, 3, 4, 4, 4, 2, 3)

	run.AssignColumn(P1, p1)
	run.AssignColumn(P2, p2)
}

func TestInclu(t *testing.T) {
	checkSolved(t, defineInclu, proveInclu, ALL_BUT_ILC, false)
}
