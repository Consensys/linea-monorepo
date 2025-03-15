//go:build !fuzzlight

package test_cases_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/stitchsplit"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

func defineRange(build *wizard.Builder) {
	P1 := build.RegisterCommit(P1, 8) // overshadows P1
	// P2 := build.RegisterCommit(P2, 8) // overshadows P2
	// P3 := build.RegisterCommit(P3, 8) // overshadows P3
	build.Range(RANGE1, P1, 16)
	// build.Range(RANGE2, P2.Repeat(2), 16)
	// build.Range(RANGE3, commitment.Interleave(P1, P3), 32)
	// build.Range(RANGE4, P3.Shift(1), 16)
}

func proveRange(run *wizard.ProverRuntime) {
	run.AssignColumn(P1, smartvectors.ForTest(14, 15, 1, 0, 3, 4, 6, 12))
	// run.AssignCommitment(P2, smartvectors.ForTest(1, 15, 14, 0, 3, 4, 6, 12))
	// run.AssignCommitment(P3, smartvectors.ForTest(0, 1, 2, 3, 4, 5, 6, 7))
}

func TestRange(t *testing.T) {
	checkSolved(t, defineRange, proveRange, join(ALL_SPECIALS, compilationSuite{stitchsplit.Splitter(8)}, DUMMY), true)
}
