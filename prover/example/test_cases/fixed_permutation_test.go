//go:build !fuzzlight

package test_cases_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

var (
	Q1 ifaces.ColID = "Q1"
	Q2 ifaces.ColID = "Q2"
	Q3 ifaces.ColID = "Q3"
)

func defineFixedPerm(build *wizard.Builder) {

	/*
		just a basic fixedpermutation protocol with three splittings
	*/
	n := 1 << 3
	//first splittings
	A1 := build.RegisterCommit(P1, n)
	A2 := build.RegisterCommit(P2, n)
	A3 := build.RegisterCommit(P3, n)

	//second splittings
	B1 := build.RegisterCommit(Q1, n)
	B2 := build.RegisterCommit(Q2, n)
	B3 := build.RegisterCommit(Q3, n)

	//given fixed polynomial in three splittings
	s1 := smartvectors.ForTest(6, 0, 2, 1, 9, 10, 23, 15)
	s2 := smartvectors.ForTest(3, 5, 11, 13, 21, 20, 19, 8)
	s3 := smartvectors.ForTest(4, 7, 12, 14, 16, 17, 18, 22)

	build.FixedPermutation(FIXEDPERMUTATION1, []ifaces.ColAssignment{s1, s2, s3}, []ifaces.Column{A1, A2, A3}, []ifaces.Column{B1, B2, B3})
}

func proveCorrectFixedPerm(run *wizard.ProverRuntime) {
	run.AssignColumn(P1, smartvectors.ForTest(5, 1, 1, 15, 1, 7, 9, 1))
	run.AssignColumn(P2, smartvectors.ForTest(1, 1, 1, 99, 53, 1, 1, 1))
	run.AssignColumn(P3, smartvectors.ForTest(1, 1, 1, 1, 1, 1, 1, 1))

	run.AssignColumn(Q1, smartvectors.ForTest(9, 5, 1, 1, 1, 1, 1, 1))
	run.AssignColumn(Q2, smartvectors.ForTest(15, 7, 99, 1, 1, 1, 1, 1))
	run.AssignColumn(Q3, smartvectors.ForTest(1, 1, 53, 1, 1, 1, 1, 1))
}
func proveWrongFixedPerm(run *wizard.ProverRuntime) {
	run.AssignColumn(P1, smartvectors.ForTest(5, 1, 1, 15, 1, 7, 9, 1))
	run.AssignColumn(P2, smartvectors.ForTest(1, 1, 1, 99, 53, 1, 1, 1))
	run.AssignColumn(P3, smartvectors.ForTest(0, 0, 1, 1, 1, 1, 1, 1))

	run.AssignColumn(Q1, smartvectors.ForTest(9, 5, 1, 1, 1, 1, 1, 1))
	run.AssignColumn(Q2, smartvectors.ForTest(15, 7, 99, 1, 1, 1, 1, 1))
	run.AssignColumn(Q3, smartvectors.ForTest(1, 1, 53, 1, 1, 1, 1, 1))
}

func TestFixedPerm(t *testing.T) {
	checkSolved(t, defineFixedPerm, proveCorrectFixedPerm, ALL_BUT_ILC, true)
	checkSolved(t, defineFixedPerm, proveWrongFixedPerm, ALL_BUT_ILC, false, true)
}
