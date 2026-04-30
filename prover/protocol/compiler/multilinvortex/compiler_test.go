package multilinvortex_test

import (
	"math/rand/v2"
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/multilineareval"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/multilinvortex"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/stretchr/testify/require"
)

// TestMultilinVortexRoundtrip verifies the Compile → Prove → Verify path for
// the multilineareval + multilinvortex combined pipeline. The define step
// inserts committed columns with MultilinearEval queries. After both compiler
// passes the residual MultilinearEval claims are consumed by dummy.Compile.
func TestMultilinVortexRoundtrip(t *testing.T) {
	cases := []struct {
		name    string
		numCols int
		numVars int
	}{
		{"1col_n4", 1, 4},
		{"2col_n4", 2, 4},
		{"3col_n6", 3, 6},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			rng := rand.New(rand.NewPCG(uint64(tc.numCols*100), uint64(tc.numVars)))
			size := 1 << tc.numVars

			colData := make([][]field.Element, tc.numCols)
			for k := 0; k < tc.numCols; k++ {
				colData[k] = make([]field.Element, size)
				for j := range colData[k] {
					colData[k][j] = field.PseudoRand(rng)
				}
			}

			point := make([]fext.Element, tc.numVars)
			for i := range point {
				point[i] = fext.PseudoRand(rng)
			}

			ys := make([]fext.Element, tc.numCols)
			for k := 0; k < tc.numCols; k++ {
				vals := make([]fext.Element, size)
				for j, v := range colData[k] {
					vals[j].B0.A0 = v
				}
				ys[k] = evalMultilin(vals, point)
			}

			define := func(b *wizard.Builder) {
				cols := make([]ifaces.Column, tc.numCols)
				for k := 0; k < tc.numCols; k++ {
					cols[k] = b.RegisterCommit(ifaces.ColIDf("COL_%d", k), size)
				}
				b.CompiledIOP.InsertMultilinear(0, "MLEVAL", cols)
			}

			prove := func(run *wizard.ProverRuntime) {
				for k := 0; k < tc.numCols; k++ {
					run.AssignColumn(ifaces.ColIDf("COL_%d", k),
						smartvectors.NewRegular(colData[k]))
				}
				run.AssignMultilinearExtShared("MLEVAL", point, ys...)
			}

			compiled := wizard.Compile(define,
				multilineareval.Compile,
				multilinvortex.Compile,
				multilineareval.Compile, // batch the two residual claims
				dummy.Compile,
			)
			proof := wizard.Prove(compiled, prove)
			require.NoError(t, wizard.Verify(compiled, proof))
		})
	}
}

// TestCommitMLColumnsCheck2 verifies that CommitMLColumns produces a prover
// action that fills the Merkle root/opened/paths proof columns, and a verifier
// action that successfully re-hashes and checks the Merkle paths (Check 2).
func TestCommitMLColumnsCheck2(t *testing.T) {
	cases := []struct {
		name    string
		numCols int
		numVars int
	}{
		{"1col_n4", 1, 4},
		{"2col_n4", 2, 4},
		{"4col_n6", 4, 6},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			rng := rand.New(rand.NewPCG(uint64(tc.numCols*100+1), uint64(tc.numVars+1)))
			size := 1 << tc.numVars

			colData := make([][]field.Element, tc.numCols)
			for k := 0; k < tc.numCols; k++ {
				colData[k] = make([]field.Element, size)
				for j := range colData[k] {
					colData[k][j] = field.PseudoRand(rng)
				}
			}

			point := make([]fext.Element, tc.numVars)
			for i := range point {
				point[i] = fext.PseudoRand(rng)
			}
			ys := make([]fext.Element, tc.numCols)
			for k := 0; k < tc.numCols; k++ {
				vals := make([]fext.Element, size)
				for j, v := range colData[k] {
					vals[j].B0.A0 = v
				}
				ys[k] = evalMultilin(vals, point)
			}

			define := func(b *wizard.Builder) {
				cols := make([]ifaces.Column, tc.numCols)
				for k := 0; k < tc.numCols; k++ {
					cols[k] = b.RegisterCommit(ifaces.ColIDf("COL_%d", k), size)
				}
				b.CompiledIOP.InsertMultilinear(0, "MLEVAL", cols)
			}
			prove := func(run *wizard.ProverRuntime) {
				for k := 0; k < tc.numCols; k++ {
					run.AssignColumn(ifaces.ColIDf("COL_%d", k),
						smartvectors.NewRegular(colData[k]))
				}
				run.AssignMultilinearExtShared("MLEVAL", point, ys...)
			}

			compiled := wizard.Compile(define,
				multilineareval.Compile,
				multilinvortex.Compile,
				multilinvortex.CommitMLColumns,
				multilinvortex.CommitOriginalMLColumns,
				multilineareval.Compile,
				multilinvortex.Compile,
				multilinvortex.CommitMLColumns,
				multilineareval.Compile,
				dummy.Compile,
			)
			proof := wizard.Prove(compiled, prove)
			require.NoError(t, wizard.Verify(compiled, proof))
		})
	}
}

func evalMultilin(vals []fext.Element, point []fext.Element) fext.Element {
	work := make([]fext.Element, len(vals))
	copy(work, vals)
	var tmp fext.Element
	for _, r := range point {
		mid := len(work) / 2
		for i := 0; i < mid; i++ {
			tmp.Sub(&work[i+mid], &work[i])
			tmp.Mul(&tmp, &r)
			work[i].Add(&work[i], &tmp)
		}
		work = work[:mid]
	}
	return work[0]
}
