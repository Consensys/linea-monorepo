package multilineareval_test

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

// TestCompileAllRoundMixedSizes verifies that CompileAllRound correctly batches
// queries of different numVars into a single combined sumcheck and that the
// residuals are assigned at the correct point prefixes.
func TestCompileAllRoundMixedSizes(t *testing.T) {
	cases := []struct {
		name    string
		nVars   []int // numVars for each query (mixed sizes)
		nCols   []int // number of columns per query
	}{
		{"same_size", []int{4, 4}, []int{2, 2}},
		{"two_sizes_3_5", []int{3, 5}, []int{1, 1}},
		{"two_sizes_3_5_multicol", []int{3, 5}, []int{2, 3}},
		{"three_sizes_2_4_6", []int{2, 4, 6}, []int{1, 2, 1}},
		{"four_sizes", []int{2, 3, 4, 5}, []int{1, 1, 1, 1}},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			rng := rand.New(rand.NewPCG(42, 7))

			// Build column data.
			colData := make([][][]field.Element, len(tc.nVars))
			points := make([][]fext.Element, len(tc.nVars))
			ys := make([][]fext.Element, len(tc.nVars))

			for q, nv := range tc.nVars {
				size := 1 << nv
				nc := tc.nCols[q]
				colData[q] = make([][]field.Element, nc)
				for p := 0; p < nc; p++ {
					colData[q][p] = make([]field.Element, size)
					for j := range colData[q][p] {
						colData[q][p][j] = field.PseudoRand(rng)
					}
				}

				points[q] = make([]fext.Element, nv)
				for i := range points[q] {
					points[q][i] = fext.PseudoRand(rng)
				}

				ys[q] = make([]fext.Element, nc)
				for p := 0; p < nc; p++ {
					vals := make([]fext.Element, size)
					for j, v := range colData[q][p] {
						vals[j].B0.A0 = v
					}
					ys[q][p] = evalMultilin(vals, points[q])
				}
			}

			define := func(b *wizard.Builder) {
				for q, nv := range tc.nVars {
					cols := make([]ifaces.Column, tc.nCols[q])
					for p := 0; p < tc.nCols[q]; p++ {
						cols[p] = b.RegisterCommit(ifaces.ColIDf("Q%d_P%d", q, p), 1<<nv)
					}
					b.CompiledIOP.InsertMultilinear(0,
						ifaces.QueryIDf("MLEVAL_%d", q),
						nv,
						cols,
					)
				}
			}

			prove := func(run *wizard.ProverRuntime) {
				for q, nv := range tc.nVars {
					for p := 0; p < tc.nCols[q]; p++ {
						run.AssignColumn(ifaces.ColIDf("Q%d_P%d", q, p),
							smartvectors.NewRegular(colData[q][p]))
					}
					_ = nv
					run.AssignMultilinearExt(ifaces.QueryIDf("MLEVAL_%d", q),
						points[q], ys[q]...)
				}
			}

			compiled := wizard.Compile(define,
				multilineareval.CompileAllRound,
				dummy.Compile,
			)
			proof := wizard.Prove(compiled, prove)
			require.NoError(t, wizard.Verify(compiled, proof))
		})
	}
}

// TestCompileAllRoundWithVortex verifies the full pipeline:
// CompileAllRound (cross-size batching) + multilinvortex.Compile (default fold)
// + another round of CompileAllRound + multilinvortex.Compile, etc.
// This is the symbolic-expansion + balanced-fold strategy end-to-end.
func TestCompileAllRoundWithVortex(t *testing.T) {
	rng := rand.New(rand.NewPCG(99, 3))

	// Two polynomials: size 2^3=8 and size 2^5=32.
	nVars := []int{3, 5}
	nCols := []int{2, 2}

	colData := make([][][]field.Element, len(nVars))
	points := make([][]fext.Element, len(nVars))
	ys := make([][]fext.Element, len(nVars))

	for q, nv := range nVars {
		size := 1 << nv
		nc := nCols[q]
		colData[q] = make([][]field.Element, nc)
		for p := 0; p < nc; p++ {
			colData[q][p] = make([]field.Element, size)
			for j := range colData[q][p] {
				colData[q][p][j] = field.PseudoRand(rng)
			}
		}
		points[q] = make([]fext.Element, nv)
		for i := range points[q] {
			points[q][i] = fext.PseudoRand(rng)
		}
		ys[q] = make([]fext.Element, nc)
		for p := 0; p < nc; p++ {
			vals := make([]fext.Element, size)
			for j, v := range colData[q][p] {
				vals[j].B0.A0 = v
			}
			ys[q][p] = evalMultilin(vals, points[q])
		}
	}

	define := func(b *wizard.Builder) {
		for q, nv := range nVars {
			cols := make([]ifaces.Column, nCols[q])
			for p := 0; p < nCols[q]; p++ {
				cols[p] = b.RegisterCommit(ifaces.ColIDf("Q%d_P%d", q, p), 1<<nv)
			}
			b.CompiledIOP.InsertMultilinear(0, ifaces.QueryIDf("MLEVAL_%d", q), nv, cols)
		}
	}

	prove := func(run *wizard.ProverRuntime) {
		for q := range nVars {
			for p := 0; p < nCols[q]; p++ {
				run.AssignColumn(ifaces.ColIDf("Q%d_P%d", q, p),
					smartvectors.NewRegular(colData[q][p]))
			}
			run.AssignMultilinearExt(ifaces.QueryIDf("MLEVAL_%d", q), points[q], ys[q]...)
		}
	}

	// Symbolic expansion + balanced Vortex fold: run enough rounds to fully
	// reduce a 5-variable polynomial (needs ~3 CompileAllRound+Vortex pairs).
	compiled := wizard.Compile(define,
		multilineareval.CompileAllRound,
		multilinvortex.Compile,
		multilineareval.CompileAllRound,
		multilinvortex.Compile,
		multilineareval.CompileAllRound,
		multilinvortex.Compile,
		multilineareval.CompileAllRound,
		dummy.Compile,
	)
	proof := wizard.Prove(compiled, prove)
	require.NoError(t, wizard.Verify(compiled, proof))
}

// TestCompileWithFold1MixedSizes tests the WHIR-style fold=1 strategy paired
// with CompileAllRound. With fold=1 the RowClaims path is always terminal
// (RowEvals has 2 elements), so ALL polynomials share that path after one
// Vortex round. UCols of different sizes are re-batched by the next
// CompileAllRound. This requires more rounds but maximises cross-size batching.
func TestCompileWithFold1MixedSizes(t *testing.T) {
	rng := rand.New(rand.NewPCG(77, 13))

	nVars := []int{3, 5}
	nCols := []int{1, 1}

	colData := make([][][]field.Element, len(nVars))
	points := make([][]fext.Element, len(nVars))
	ys := make([][]fext.Element, len(nVars))

	for q, nv := range nVars {
		size := 1 << nv
		nc := nCols[q]
		colData[q] = make([][]field.Element, nc)
		for p := 0; p < nc; p++ {
			colData[q][p] = make([]field.Element, size)
			for j := range colData[q][p] {
				colData[q][p][j] = field.PseudoRand(rng)
			}
		}
		points[q] = make([]fext.Element, nv)
		for i := range points[q] {
			points[q][i] = fext.PseudoRand(rng)
		}
		ys[q] = make([]fext.Element, nc)
		for p := 0; p < nc; p++ {
			vals := make([]fext.Element, size)
			for j, v := range colData[q][p] {
				vals[j].B0.A0 = v
			}
			ys[q][p] = evalMultilin(vals, points[q])
		}
	}

	define := func(b *wizard.Builder) {
		for q, nv := range nVars {
			cols := make([]ifaces.Column, nCols[q])
			for p := 0; p < nCols[q]; p++ {
				cols[p] = b.RegisterCommit(ifaces.ColIDf("Q%d_P%d", q, p), 1<<nv)
			}
			b.CompiledIOP.InsertMultilinear(0, ifaces.QueryIDf("MLEVAL_%d", q), nv, cols)
		}
	}

	prove := func(run *wizard.ProverRuntime) {
		for q := range nVars {
			for p := 0; p < nCols[q]; p++ {
				run.AssignColumn(ifaces.ColIDf("Q%d_P%d", q, p),
					smartvectors.NewRegular(colData[q][p]))
			}
			run.AssignMultilinearExt(ifaces.QueryIDf("MLEVAL_%d", q), points[q], ys[q]...)
		}
	}

	// WHIR early exit: fold=1 means nRow=1 at every Vortex round.
	// A 5-variable polynomial needs 4 rounds of (CompileAllRound + CompileWithFold(1))
	// to reduce all UCols to terminal. RowClaims are terminal after round 1.
	fold1 := multilinvortex.CompileWithFold(1)
	compiled := wizard.Compile(define,
		multilineareval.CompileAllRound,
		fold1,
		multilineareval.CompileAllRound,
		fold1,
		multilineareval.CompileAllRound,
		fold1,
		multilineareval.CompileAllRound,
		fold1,
		multilineareval.CompileAllRound,
		dummy.Compile,
	)
	proof := wizard.Prove(compiled, prove)
	require.NoError(t, wizard.Verify(compiled, proof))
}
