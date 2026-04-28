package multilineareval_test

import (
	"math/rand/v2"
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/multilineareval"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/stretchr/testify/require"
)

// TestMultilinearEvalCompilerRoundtrip verifies that the multilineareval
// compiler pass correctly batches multiple MultilinearEval queries into a
// sumcheck, producing a residual MultilinearEval query that the dummy verifier
// then checks.
func TestMultilinearEvalCompilerRoundtrip(t *testing.T) {
	cases := []struct {
		name        string
		numQueries  int
		polsPerQuery int
		numVars     int
	}{
		{"single_query_1_poly", 1, 1, 3},
		{"single_query_3_poly", 1, 3, 4},
		{"two_queries_2_poly", 2, 2, 3},
		{"three_queries_mixed", 3, 2, 4},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			rng := rand.New(rand.NewPCG(uint64(tc.numQueries), uint64(tc.numVars)))
			size := 1 << tc.numVars

			// Build column data for each query and poly.
			colData := make([][][]field.Element, tc.numQueries)
			for q := 0; q < tc.numQueries; q++ {
				colData[q] = make([][]field.Element, tc.polsPerQuery)
				for p := 0; p < tc.polsPerQuery; p++ {
					colData[q][p] = make([]field.Element, size)
					for j := range colData[q][p] {
						colData[q][p][j] = field.PseudoRand(rng)
					}
				}
			}

			// Random evaluation points — one per query.
			points := make([][]fext.Element, tc.numQueries)
			for q := 0; q < tc.numQueries; q++ {
				points[q] = make([]fext.Element, tc.numVars)
				for i := range points[q] {
					points[q][i] = fext.PseudoRand(rng)
				}
			}

			// Pre-compute expected evaluations.
			ys := make([][]fext.Element, tc.numQueries)
			for q := 0; q < tc.numQueries; q++ {
				ys[q] = make([]fext.Element, tc.polsPerQuery)
				for p := 0; p < tc.polsPerQuery; p++ {
					vals := make([]fext.Element, size)
					for j, v := range colData[q][p] {
						vals[j].B0.A0 = v
					}
					ys[q][p] = evalMultilin(vals, points[q])
				}
			}

			define := func(b *wizard.Builder) {
				for q := 0; q < tc.numQueries; q++ {
					cols := make([]ifaces.Column, tc.polsPerQuery)
					for p := 0; p < tc.polsPerQuery; p++ {
						cols[p] = b.RegisterCommit(ifaces.ColIDf("Q%d_P%d", q, p), size)
					}
					b.CompiledIOP.InsertMultilinear(0,
						ifaces.QueryIDf("MLEVAL_%d", q),
						tc.numVars,
						cols,
					)
				}
			}

			prove := func(run *wizard.ProverRuntime) {
				for q := 0; q < tc.numQueries; q++ {
					for p := 0; p < tc.polsPerQuery; p++ {
						run.AssignColumn(ifaces.ColIDf("Q%d_P%d", q, p),
							smartvectors.NewRegular(colData[q][p]))
					}
					run.AssignMultilinearExt(ifaces.QueryIDf("MLEVAL_%d", q),
						points[q], ys[q]...)
				}
			}

			compiled := wizard.Compile(define,
				multilineareval.Compile,
				dummy.Compile,
			)
			proof := wizard.Prove(compiled, prove)
			require.NoError(t, wizard.Verify(compiled, proof))
		})
	}
}

// evalMultilin is a local helper that evaluates a multilinear polynomial
// (given as a table of 2^n fext values, MSB-first) at a point.
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
