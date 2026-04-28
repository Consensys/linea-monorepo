package query_test

import (
	"math/rand/v2"
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/stretchr/testify/require"
)

// TestMultilinearEvalWizardRoundtrip exercises the full Compile → Prove →
// Verify path for a MultilinearEval query. The define step registers N
// committed columns of size 2^numVars and a single MultilinearEval query over
// all of them; the prove step evaluates each column at a random fext point and
// assigns the params; the verify step runs the native Check.
func TestMultilinearEvalWizardRoundtrip(t *testing.T) {
	cases := []struct{ N, numVars int }{
		{1, 2},
		{3, 3},
		{4, 4},
	}
	for _, tc := range cases {
		t.Run("", func(t *testing.T) {
			rng := rand.New(rand.NewPCG(uint64(tc.N), uint64(tc.numVars)))
			size := 1 << tc.numVars

			// Build random column data.
			colData := make([][]field.Element, tc.N)
			for i := 0; i < tc.N; i++ {
				colData[i] = make([]field.Element, size)
				for j := range colData[i] {
					colData[i][j] = field.PseudoRand(rng)
				}
			}

			// Random evaluation point.
			point := make([]fext.Element, tc.numVars)
			for i := range point {
				point[i] = fext.PseudoRand(rng)
			}

			// Pre-compute expected Ys (before the wizard runs, for validation).
			expectedYs := make([]fext.Element, tc.N)
			for i := 0; i < tc.N; i++ {
				vals := make([]fext.Element, size)
				for j, v := range colData[i] {
					vals[j].B0.A0 = v // lift base to fext
				}
				expectedYs[i] = evalMultilin(vals, point)
			}

			define := func(b *wizard.Builder) {
				cols := make([]ifaces.Column, tc.N)
				for i := 0; i < tc.N; i++ {
					cols[i] = b.RegisterCommit(ifaces.ColIDf("P_%d", i), size)
				}
				b.CompiledIOP.InsertMultilinear(0, "MLEVAL", tc.numVars, cols)
			}

			prove := func(run *wizard.ProverRuntime) {
				for i := 0; i < tc.N; i++ {
					run.AssignColumn(ifaces.ColIDf("P_%d", i), smartvectors.NewRegular(colData[i]))
				}
				run.AssignMultilinearExt("MLEVAL", point, expectedYs...)
			}

			compiled := wizard.Compile(define, dummy.Compile)
			proof := wizard.Prove(compiled, prove)
			require.NoError(t, wizard.Verify(compiled, proof))
		})
	}
}

// TestMultilinearEvalCheckRejectsWrongY tests that Check returns an error when
// the claimed evaluation value is wrong.
func TestMultilinearEvalCheckRejectsWrongY(t *testing.T) {
	rng := rand.New(rand.NewPCG(0xABCD, 0x1234))
	const numVars = 3
	size := 1 << numVars

	colData := make([]field.Element, size)
	for j := range colData {
		colData[j] = field.PseudoRand(rng)
	}
	point := make([]fext.Element, numVars)
	for i := range point {
		point[i] = fext.PseudoRand(rng)
	}

	vals := make([]fext.Element, size)
	for j, v := range colData {
		vals[j].B0.A0 = v
	}
	correctY := evalMultilin(vals, point)

	var wrongY fext.Element
	var bump fext.Element
	bump.SetOne()
	wrongY.Add(&correctY, &bump)

	define := func(b *wizard.Builder) {
		col := b.RegisterCommit("P", size)
		b.CompiledIOP.InsertMultilinear(0, "MLEVAL_BAD", numVars, []ifaces.Column{col})
	}
	prove := func(run *wizard.ProverRuntime) {
		run.AssignColumn("P", smartvectors.NewRegular(colData))
		run.AssignMultilinearExt("MLEVAL_BAD", point, wrongY)
	}

	compiled := wizard.Compile(define, dummy.Compile)
	proof := wizard.Prove(compiled, prove)
	require.Error(t, wizard.Verify(compiled, proof))
}

// evalMultilin mirrors sumcheck.MultiLin.Evaluate for test use without importing
// the sumcheck package (to keep the test in the query package).
func evalMultilin(vals []fext.Element, point []fext.Element) fext.Element {
	work := make([]fext.Element, len(vals))
	copy(work, vals)
	var t fext.Element
	for _, r := range point {
		mid := len(work) / 2
		for i := 0; i < mid; i++ {
			t.Sub(&work[i+mid], &work[i])
			t.Mul(&t, &r)
			work[i].Add(&work[i], &t)
		}
		work = work[:mid]
	}
	return work[0]
}
