//go:build !fuzzlight

package test_cases_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/splitter/splitter"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/splitter/stitcher"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

func defineRange(build *wizard.Builder) {
	P1 := build.RegisterCommit(P1, 8) // overshadows P1
	P2 := build.RegisterCommit(P2, 8) // overshadows P2
	P3 := build.RegisterCommit(P3, 8) // overshadows P3
	build.Range(RANGE1, P1, 16)
	build.Range(RANGE2, column.Shift(P2, 1), 16)
	build.Range(RANGE4, column.Shift(P3, -2), 16)
}

func proveRange(run *wizard.ProverRuntime) {
	run.AssignColumn(P1, smartvectors.ForTest(14, 15, 1, 0, 3, 4, 6, 12))
	run.AssignColumn(P2, smartvectors.ForTest(1, 15, 14, 0, 3, 4, 6, 12))
	run.AssignColumn(P3, smartvectors.ForTest(0, 1, 2, 3, 4, 5, 6, 7))
}

func TestRange(t *testing.T) {
	checkSolved(t, defineRange, proveRange, join(ALL_SPECIALS, compilationSuite{stitcher.Stitcher(4, 8), splitter.Splitter(8)}, DUMMY), true)
}
