package compilers_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover-ray/wiop"
	"github.com/consensys/linea-monorepo/prover-ray/wiop/wioptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMutationSoundness_VanishingScenarios is the basis of the mutation-driven
// soundness tester. For every honest vanishing-scenario workload it perturbs a
// single committed witness entry via [wioptest.Mutator], reruns the full
// compilation pipeline, and asserts the compiled verifier rejects the corrupted
// run.
//
// A witness column bound by a vanishing constraint must have at least one entry
// whose corruption the verifier detects. If no single-entry mutation of any
// oracle column is caught, the witness is effectively unconstrained — a
// soundness gap this test flags.
func TestMutationSoundness_VanishingScenarios(t *testing.T) {
	for _, build := range wioptest.VanishingScenarios() {
		sc := build()
		t.Run(sc.Name, func(t *testing.T) {
			lens := honestOracleColumnLens(build)
			require.NotEmpty(t, lens,
				"scenario must expose at least one oracle witness column")

			caught := false
		probe:
			for colIdx, n := range lens {
				for _, row := range probeRows(n) {
					if mutationRejected(build, colIdx, row) {
						caught = true
						break probe
					}
				}
			}
			assert.True(t, caught,
				"a single-entry mutation of a constrained witness column must be rejected by the compiled verifier")
		})
	}
}

// honestOracleColumnLens runs the honest workload once and returns the assigned
// length of each round-0 oracle column, in declaration order. The lengths bound
// the rows the mutator may target.
func honestOracleColumnLens(build func() *wioptest.VanishingScenario) []int {
	sc := build()
	compileFullPipeline(sc.Sys)
	cols := oracleColumns(sc.Sys)
	rt := wiop.NewRuntime(sc.Sys)
	sc.AssignHonest(&rt)

	lens := make([]int, len(cols))
	for i, c := range cols {
		lens[i] = rt.GetColumnAssignment(c).Plain.Len()
	}
	return lens
}

// mutationRejected rebuilds the scenario, corrupts entry row of the colIdx-th
// oracle column, and reports whether the corrupted run is rejected. A panic in
// a prover action counts as rejection: a witness that no longer lets the prover
// build a valid proof has not sailed through to an accepting verifier.
func mutationRejected(build func() *wioptest.VanishingScenario, colIdx, row int) (rejected bool) {
	defer func() {
		if r := recover(); r != nil {
			rejected = true
		}
	}()

	sc := build()
	compileFullPipeline(sc.Sys)
	col := oracleColumns(sc.Sys)[colIdx]
	wioptest.Mutator{Column: col, Row: row}.Compile(sc.Sys)

	rt := wiop.NewRuntime(sc.Sys)
	sc.AssignHonest(&rt)
	return wioptest.RunAndVerify(&rt) != nil
}

// oracleColumns returns the round-0 oracle columns of sys in declaration order.
// These are the prover-committed witness columns a mutator can corrupt.
func oracleColumns(sys *wiop.System) []*wiop.Column {
	var out []*wiop.Column
	for _, c := range sys.Rounds[0].Columns {
		if c.Visibility == wiop.VisibilityOracle {
			out = append(out, c)
		}
	}
	return out
}

// probeRows returns the distinct rows to corrupt for a column of length n:
// the first, middle, and last entries. Probing the boundaries and the interior
// covers boundary-pinned, recurrence, and full-domain constraints without
// scanning every row of large columns.
func probeRows(n int) []int {
	if n <= 0 {
		return nil
	}
	rows := []int{0}
	if mid := n / 2; mid != 0 {
		rows = append(rows, mid)
	}
	if last := n - 1; last != 0 && last != n/2 {
		rows = append(rows, last)
	}
	return rows
}
