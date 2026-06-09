package wiop_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover-ray/maths/koalabear/field"
	"github.com/consensys/linea-monorepo/prover-ray/wiop"
	"github.com/consensys/linea-monorepo/prover-ray/wiop/wioptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestQueryMutationSoundness drives every per-query scenario from
// [wioptest.All] with the mutation tester. For each scenario it builds the
// honest witness and then perturbs each committed input — every entry of every
// oracle column and every opening cell — asserting the query's Check rejects at
// least one perturbation.
//
// A query whose Check survives every single-value perturbation does not depend
// on its inputs, which is a soundness gap this flags. Each perturbation is
// applied to a freshly rebuilt scenario so attempts never compound.
func TestQueryMutationSoundness(t *testing.T) {
	for _, build := range wioptest.All() {
		sc := build()
		t.Run(sc.Name, func(t *testing.T) {
			// Completeness sanity: the honest witness must pass.
			rt := wiop.NewRuntime(sc.Sys)
			sc.RunHonest(&rt)
			require.NoError(t, sc.Query.Check(rt),
				"honest witness must pass Check")

			assert.True(t, anyMutationRejected(build),
				"Check must reject at least one single-value perturbation of the honest witness")
		})
	}
}

// mutationTweaks are the perturbations the tester tries per target. AddOne is
// the minimal change, good for exact-equality constraints; RandomValue is the
// general default; addLarge deterministically escapes the slack of range and
// lookup relations, where a small bump can stay in range.
var mutationTweaks = []wioptest.Tweak{wioptest.AddOne, wioptest.RandomValue, addLarge}

// queryTarget identifies one value to corrupt by position, so it can be
// re-resolved against a freshly rebuilt System on each attempt.
type queryTarget struct {
	isCell bool
	idx    int
	row    int
	tweak  wioptest.Tweak
}

// anyMutationRejected reports whether some perturbation of some committed input
// makes the query's Check fail. It returns on the first rejection.
func anyMutationRejected(build func() *wioptest.Scenario) bool {
	for _, tg := range queryTargets(build) {
		if mutationRejectedByCheck(build, tg) {
			return true
		}
	}
	return false
}

// queryTargets enumerates every (committed value, tweak) pair to try. It runs
// the honest witness once to learn the column lengths and which cells are
// assigned, then expands the cross product with mutationTweaks.
func queryTargets(build func() *wioptest.Scenario) []queryTarget {
	sc := build()
	rt := wiop.NewRuntime(sc.Sys)
	sc.RunHonest(&rt)

	var targets []queryTarget
	for i, c := range oracleColumnsOf(sc.Sys) {
		if !rt.HasColumnAssignment(c) {
			continue
		}
		n := rt.GetColumnAssignment(c).Plain.Len()
		for row := 0; row < n; row++ {
			for _, tw := range mutationTweaks {
				targets = append(targets, queryTarget{idx: i, row: row, tweak: tw})
			}
		}
	}
	for j, cell := range cellsOf(sc.Sys) {
		if !rt.HasCellValue(cell) {
			continue
		}
		for _, tw := range mutationTweaks {
			targets = append(targets, queryTarget{isCell: true, idx: j, tweak: tw})
		}
	}
	return targets
}

// mutationRejectedByCheck rebuilds the scenario, applies one perturbation to
// the honest witness, and reports whether Check rejects it. A panic during the
// mutation or the check counts as a rejection: a witness that breaks the run
// has not sailed through to an accepting verifier.
func mutationRejectedByCheck(build func() *wioptest.Scenario, tg queryTarget) (rejected bool) {
	defer func() {
		if r := recover(); r != nil {
			rejected = true
		}
	}()

	sc := build()
	rt := wiop.NewRuntime(sc.Sys)
	sc.RunHonest(&rt)

	m := wioptest.Mutator{Tweak: tg.tweak}
	if tg.isCell {
		m.Cell = cellsOf(sc.Sys)[tg.idx]
	} else {
		m.Column = oracleColumnsOf(sc.Sys)[tg.idx]
		m.Row = tg.row
	}
	m.Apply(rt)

	return sc.Query.Check(rt) != nil
}

// oracleColumnsOf returns every oracle column of sys across all interactive
// rounds, in declaration order. These are the prover-committed witness columns.
func oracleColumnsOf(sys *wiop.System) []*wiop.Column {
	var out []*wiop.Column
	for _, r := range sys.Rounds {
		for _, c := range r.Columns {
			if c.Visibility == wiop.VisibilityOracle {
				out = append(out, c)
			}
		}
	}
	return out
}

// cellsOf returns every cell of sys across all interactive rounds, in
// declaration order. These are the prover's scalar openings.
func cellsOf(sys *wiop.System) []*wiop.Cell {
	var out []*wiop.Cell
	for _, r := range sys.Rounds {
		out = append(out, r.Cells...)
	}
	return out
}

// addLarge adds 2^30 to v, large enough to break range and lookup relations
// whose honest values sit far from their bounds, while staying below the
// koalabear modulus so it cannot wrap back into range.
func addLarge(v field.Gen) field.Gen {
	var e field.Element
	e.SetUint64(1 << 30)
	return v.Add(field.ElemFromBase(e))
}
