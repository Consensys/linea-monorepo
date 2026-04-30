package multilinvortex_test

import (
	"math/rand/v2"
	"testing"
	"time"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/multilineareval"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/multilinvortex"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// colSpec describes one column: count × size.
type colSpec struct {
	count, size int
}

// productionApprox approximates a mainnet bootstrapper column distribution.
// The total cell count is deliberately kept at ~1/64 of the real 23.75B to
// keep the benchmark under a minute; multiply measured latency by 64 to
// estimate production overhead.
var productionApprox = []colSpec{
	{count: 2000, size: 1 << 10},  // 2K  → 2.0M cells
	{count: 500, size: 1 << 14},   // 16K → 8.0M cells
	{count: 100, size: 1 << 17},   // 128K → 12.8M cells
	{count: 20, size: 1 << 20},    // 1M  → 20.0M cells
	{count: 5, size: 1 << 22},     // 4M  → 20.0M cells
	{count: 1, size: 1 << 24},     // 16M → 16.0M cells
	// Total: ~78.8M cells (≈ 1/300 of production 23.75B)
}

// BenchmarkMLBootstrapperProver measures the full InsertBootstrapperOpenings +
// CompileAllRound + Compile prove cycle with a synthetic column distribution
// that approximates the production bootstrapper at ~1/300 scale.
func BenchmarkMLBootstrapperProver(b *testing.B) {
	rng := rand.New(rand.NewPCG(42, 0))

	// Build column data once (outside the bench loop).
	type colEntry struct {
		id   ifaces.ColID
		data []field.Element
	}
	var entries []colEntry
	colIdx := 0
	for _, spec := range productionApprox {
		for k := 0; k < spec.count; k++ {
			id := ifaces.ColIDf("COL_%d", colIdx)
			data := make([]field.Element, spec.size)
			for i := range data {
				data[i] = field.PseudoRand(rng)
			}
			entries = append(entries, colEntry{id: id, data: data})
			colIdx++
		}
	}

	define := func(b *wizard.Builder) {
		for _, e := range entries {
			b.RegisterCommit(e.id, len(e.data))
		}
	}

	// Compile→CompileAllRound order: Compile halves numVars first so that
	// CompileAllRound only expands within the halved space, avoiding the
	// O(2^nmax × numCols) memory explosion that would occur if CompileAllRound
	// ran first on heterogeneous-size columns.
	compiledML := wizard.Compile(
		define,
		multilinvortex.InsertBootstrapperOpenings,
		multilinvortex.Compile,
		multilineareval.CompileAllRound,
		multilinvortex.Compile,
		multilineareval.CompileAllRound,
		multilinvortex.Compile,
		multilineareval.CompileAllRound,
		multilinvortex.Compile,
		multilineareval.CompileAllRound,
		multilinvortex.Compile,
		multilineareval.CompileAllRound,
		multilinvortex.Compile,
		multilineareval.CompileAllRound,
		multilinvortex.Compile,
		multilineareval.CompileAllRound,
		dummy.Compile,
	)

	compiledBaseline := wizard.Compile(define, dummy.Compile)

	prove := func(run *wizard.ProverRuntime) {
		for _, e := range entries {
			run.AssignColumn(e.id, smartvectors.NewRegular(e.data))
		}
	}

	b.Run("ML", func(b *testing.B) {
		for range b.N {
			start := time.Now()
			wizard.Prove(compiledML, prove)
			b.ReportMetric(float64(time.Since(start).Milliseconds()), "ms/op")
		}
	})

	b.Run("Baseline", func(b *testing.B) {
		for range b.N {
			start := time.Now()
			wizard.Prove(compiledBaseline, prove)
			b.ReportMetric(float64(time.Since(start).Milliseconds()), "ms/op")
		}
	})
}
