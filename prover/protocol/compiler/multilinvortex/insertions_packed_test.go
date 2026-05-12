package multilinvortex_test

import (
	"math/rand/v2"
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/multilineareval"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/multilinvortex"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/stretchr/testify/require"
)

// TestInsertBootstrapperOpeningsPacked_Roundtrip verifies that the cross-size
// packed bootstrapper produces a correct prove→verify path through the
// existing ML Vortex pipeline. The same data goes through:
//
//	InsertBootstrapperOpeningsPacked → CompileRound + Batch (×5) → dummy.Compile
//
// Each test case uses columns of MIXED sizes — the case this pass exists for.
func TestInsertBootstrapperOpeningsPacked_Roundtrip(t *testing.T) {
	type colDef struct{ size int }
	cases := []struct {
		name string
		cols []colDef
	}{
		{
			name: "single_small",
			cols: []colDef{{1 << 4}},
		},
		{
			name: "two_equal_size",
			cols: []colDef{{1 << 5}, {1 << 5}},
		},
		{
			name: "two_different_sizes",
			cols: []colDef{{1 << 6}, {1 << 4}},
		},
		{
			name: "mixed_three_sizes",
			cols: []colDef{{1 << 7}, {1 << 5}, {1 << 5}, {1 << 4}, {1 << 3}},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			rng := rand.New(rand.NewPCG(0xBEEF, uint64(len(tc.cols))))

			// Build column data.
			colData := make([][]field.Element, len(tc.cols))
			for k, c := range tc.cols {
				colData[k] = make([]field.Element, c.size)
				for j := range colData[k] {
					colData[k][j] = field.PseudoRand(rng)
				}
			}

			define := func(b *wizard.Builder) {
				for k, c := range tc.cols {
					b.RegisterCommit(ifaces.ColIDf("COL_%d", k), c.size)
					_ = c
				}
			}

			prove := func(run *wizard.ProverRuntime) {
				for k := range tc.cols {
					run.AssignColumn(ifaces.ColIDf("COL_%d", k),
						smartvectors.NewRegular(colData[k]))
				}
				// NOTE: with InsertBootstrapperOpeningsPacked, the prover does
				// NOT call run.AssignMultilinear* directly — the registered
				// PackedBootstrapperProverAction does it automatically at round 1.
			}

			compiled := wizard.Compile(define,
				multilinvortex.InsertBootstrapperOpeningsPacked,
				multilinvortex.CompileRound, multilineareval.Batch,
				multilinvortex.CompileRound, multilineareval.Batch,
				multilinvortex.CompileRound, multilineareval.Batch,
				multilinvortex.CompileRound, multilineareval.Batch,
				multilinvortex.CompileRound, multilineareval.Batch,
				dummy.Compile,
			)
			proof := wizard.Prove(compiled, prove)
			require.NoError(t, wizard.Verify(compiled, proof))
		})
	}
}
