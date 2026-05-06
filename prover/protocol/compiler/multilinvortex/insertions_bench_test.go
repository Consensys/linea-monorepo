package multilinvortex_test

import (
	"math/rand/v2"
	"testing"
	"time"

	"github.com/consensys/linea-monorepo/prover/crypto/ringsis"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/cleanup"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/multilineareval"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/multilinvortex"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/poseidon2"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/selfrecursion"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/vortex"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
)

// colSpec describes one column: count × size.
type colSpec struct {
	count, size int
}

// productionApprox approximates a mainnet bootstrapper column distribution.
// The total cell count is deliberately kept at ~1/300 of the real 20.5B to
// keep the benchmark under a minute; multiply measured latency by ~300 to
// estimate production overhead.
var productionApprox = []colSpec{
	{count: 2000, size: 1 << 10}, // 2K  → 2.0M cells
	{count: 500, size: 1 << 14},  // 16K → 8.0M cells
	{count: 100, size: 1 << 17},  // 128K → 12.8M cells
	{count: 20, size: 1 << 20},   // 1M  → 20.0M cells
	{count: 5, size: 1 << 22},    // 4M  → 20.0M cells
	{count: 1, size: 1 << 24},    // 16M → 16.0M cells
	// Total: ~78.8M cells (≈ 1/300 of production 20.5B)
}

// productionFull is the actual mainnet column distribution extracted from
// config-mainnet-limitless.toml via arithmetization.Define.
// Total: ~20.54B cells ≈ 164 GiB of field data.
// Requires ~500 GiB peak RSS (column data + RS encoding overhead).
var productionFull = []colSpec{
	{count: 147, size: 1 << 14},   // 16K  → 2.4M cells   (STP)
	{count: 55, size: 1 << 15},    // 32K  → 1.8M cells   (TRM)
	{count: 511, size: 1 << 16},   // 64K  → 33.5M cells  (EUC, EXP, GAS)
	{count: 15328, size: 1 << 17}, // 128K → 2.0B cells   (bit/byte ops, lookups)
	{count: 2276, size: 1 << 18},  // 256K → 596.6M cells (ADD, BIN, OOB, SHF, WCP)
	{count: 274, size: 1 << 19},   // 512K → 143.7M cells (EXT)
	{count: 1025, size: 1 << 20},  // 1M   → 1.1B cells   (MMU, MXP)
	{count: 6525, size: 1 << 21},  // 2M   → 13.7B cells  (HUB, MMIO)
	{count: 357, size: 1 << 23},   // 8M   → 3.0B cells   (ROM, ROM_LEX)
	// Total: ~20.54B cells
}

// benchSISParams matches the SIS instance used by full.go for the bootstrapper
// (LogTwoBound=16, LogTwoDegree=6, i.e. degree-64 ring).
var benchSISParams = ringsis.Params{LogTwoBound: 16, LogTwoDegree: 6}

// insertUnivariateBootstrapperQuery is a compile step that MUST run after
// Arcane (which normalises all committed columns to a uniform size via
// Stitcher + Splitter). It inserts a single UnivariateEval query covering
// every round-0 committed column and registers the matching prover action that
// evaluates them at the Fiat-Shamir challenge. Without this step vortex.Compile
// would skip (no query found) and the prover would do no cryptographic work.
func insertUnivariateBootstrapperQuery(comp *wizard.CompiledIOP) {
	cols := comp.Columns.AllHandleCommittedAt(0)
	if len(cols) == 0 {
		return
	}

	const (
		coinName = coin.Name("UNI_BOOTSTRAP_X")
		qName    = ifaces.QueryID("UNI_BOOTSTRAP_EVAL")
	)

	comp.InsertCoin(1, coinName, coin.FieldExt)
	q := comp.InsertUnivariate(1, qName, cols)
	comp.RegisterProverAction(1, &uniBootstrapProverAction{
		q:        q,
		cols:     cols,
		coinName: coinName,
	})
}

type uniBootstrapProverAction struct {
	q        query.UnivariateEval
	cols     []ifaces.Column
	coinName coin.Name
}

func (a *uniBootstrapProverAction) Run(run *wizard.ProverRuntime) {
	x := run.GetRandomCoinFieldExt(a.coinName)
	ys := make([]fext.Element, len(a.cols))
	parallel.Execute(len(a.cols), func(start, stop int) {
		for k := start; k < stop; k++ {
			sv := run.GetColumn(a.cols[k].GetColID())
			ys[k] = smartvectors.EvaluateBasePolyLagrange(sv, x)
		}
	})
	run.AssignUnivariateExt(a.q.QueryID, x, ys...)
}

// benchBootstrapperProver is the shared implementation for both the small-scale
// and full-scale bootstrapper benchmarks.  It compiles both the ML and
// univariate pipelines against the given column distribution, then runs each
// sub-benchmark for b.N iterations.
func benchBootstrapperProver(b *testing.B, dist []colSpec) {
	b.Helper()
	rng := rand.New(rand.NewPCG(42, 0))

	// Build column data once (outside the bench loop).
	type colEntry struct {
		id   ifaces.ColID
		data []field.Element
	}
	var entries []colEntry
	colIdx := 0
	for _, spec := range dist {
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

	// ML path: five rounds of CompileRound+Batch cover any numVars ≤ 32.
	compiledML := wizard.Compile(
		define,
		multilinvortex.InsertBootstrapperOpenings,
		multilinvortex.CompileRound,
		multilineareval.Batch,
		multilinvortex.CompileRound,
		multilineareval.Batch,
		multilinvortex.CompileRound,
		multilineareval.Batch,
		multilinvortex.CompileRound,
		multilineareval.Batch,
		multilinvortex.CompileRound,
		multilineareval.Batch,
		dummy.Compile,
	)

	// Univariate path: mirrors the four-round full.go bootstrapper pipeline.
	// insertUnivariateBootstrapperQuery adds the single UnivariateEval query
	// that vortex requires; it must come after Arcane so the columns are
	// already normalised to a uniform size.
	compiledUnivariate := wizard.Compile(
		define,
		// Round 1 — stitch/split heterogeneous columns to 1M rows, commit.
		compiler.Arcane(
			compiler.WithTargetColSize(1<<20),
			compiler.WithStitcherMinSize(16),
		),
		insertUnivariateBootstrapperQuery,
		vortex.Compile(
			2, false,
			vortex.ForceNumOpenedColumns(256),
			vortex.WithSISParams(&benchSISParams),
			vortex.WithOptionalSISHashingThreshold(1),
		),
		// Round 2
		selfrecursion.SelfRecurse,
		cleanup.CleanUp,
		poseidon2.CompilePoseidon2,
		compiler.Arcane(
			compiler.WithTargetColSize(1<<17),
			compiler.WithStitcherMinSize(16),
		),
		vortex.Compile(
			8, false,
			vortex.ForceNumOpenedColumns(86),
			vortex.WithSISParams(&benchSISParams),
			vortex.WithOptionalSISHashingThreshold(1),
		),
		// Round 3
		selfrecursion.SelfRecurse,
		cleanup.CleanUp,
		poseidon2.CompilePoseidon2,
		compiler.Arcane(
			compiler.WithTargetColSize(1<<15),
			compiler.WithStitcherMinSize(16),
		),
		vortex.Compile(
			16, false,
			vortex.ForceNumOpenedColumns(64),
			vortex.WithSISParams(&benchSISParams),
			vortex.WithOptionalSISHashingThreshold(1),
		),
		// Round 4 — final: mark as self-recursed so the gnark verifier can
		// consume the proof without another self-recursion pass.
		selfrecursion.SelfRecurse,
		cleanup.CleanUp,
		poseidon2.CompilePoseidon2,
		compiler.Arcane(
			compiler.WithTargetColSize(1<<14),
			compiler.WithStitcherMinSize(16),
		),
		vortex.Compile(
			16, false,
			vortex.ForceNumOpenedColumns(64),
			vortex.WithOptionalSISHashingThreshold(1<<20),
			vortex.PremarkAsSelfRecursed(),
		),
	)

	// UniNR path: single vortex round (same params as round 1 of the full
	// univariate pipeline) with no self-recursion. Directly comparable to the
	// ML path: same witness, same statement, one commitment round + dummy.
	compiledUniNR := wizard.Compile(
		define,
		compiler.Arcane(
			compiler.WithTargetColSize(1<<20),
			compiler.WithStitcherMinSize(16),
		),
		insertUnivariateBootstrapperQuery,
		vortex.Compile(
			2, false,
			vortex.ForceNumOpenedColumns(256),
			vortex.WithSISParams(&benchSISParams),
			vortex.WithOptionalSISHashingThreshold(1),
		),
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

	b.Run("Univariate", func(b *testing.B) {
		for range b.N {
			start := time.Now()
			wizard.Prove(compiledUnivariate, prove)
			b.ReportMetric(float64(time.Since(start).Milliseconds()), "ms/op")
		}
	})

	b.Run("UniNR", func(b *testing.B) {
		for range b.N {
			start := time.Now()
			wizard.Prove(compiledUniNR, prove)
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

// BenchmarkMLBootstrapperProver compares the ML and univariate pipelines on a
// scaled-down (~1/300) approximation of the production column distribution.
// Run with -benchtime=1x -v; multiply ms/op by ~300 to estimate production.
func BenchmarkMLBootstrapperProver(b *testing.B) {
	benchBootstrapperProver(b, productionApprox)
}

// BenchmarkFullScaleProver runs the same ML vs univariate comparison on the
// actual mainnet column distribution (20.54B cells, ~164 GiB of field data).
// Expected peak RSS: ~500 GiB.  Expected wall time: ~25 min (ML), ~60 min
// (Univariate) on a 192-core machine.  Run with -benchtime=1x -timeout=2h.
func BenchmarkFullScaleProver(b *testing.B) {
	benchBootstrapperProver(b, productionFull)
}

// dynamicDist matches the distribution in TestDynamicColumnSizes: four column
// sizes spanning two orders of magnitude.
var dynamicDist = []colSpec{
	{count: 16, size: 1 << 6},
	{count: 4, size: 1 << 8},
	{count: 2, size: 1 << 10},
	{count: 1, size: 1 << 12},
}

// dynamicDistExtended adds a fifth size group (nv=14 → 16K elements) to
// dynamicDist, modelling a new zkEVM module being added between releases.
var dynamicDistExtended = []colSpec{
	{count: 16, size: 1 << 6},
	{count: 4, size: 1 << 8},
	{count: 2, size: 1 << 10},
	{count: 1, size: 1 << 12},
	{count: 3, size: 1 << 14}, // new size group
}

// BenchmarkDynamicColumnSizes measures prover time for the heterogeneous
// distribution used in TestDynamicColumnSizes. Unlike BenchmarkMLBootstrapper-
// Prover (which uses production Arcane params for the Univariate path), this
// benchmark uses Arcane target sizes calibrated to the actual post-SelfRecurse
// column distribution so each path is fairly configured.
// Run with: go test -bench BenchmarkDynamicColumnSizes -benchtime=1x -v
func BenchmarkDynamicColumnSizes(b *testing.B) {
	b.Helper()
	rng := rand.New(rand.NewPCG(7, 0))

	type entry struct {
		id   ifaces.ColID
		data []field.Element
	}
	var entries []entry
	idx := 0
	for _, spec := range dynamicDist {
		for k := 0; k < spec.count; k++ {
			id := ifaces.ColIDf("COL_%d", idx)
			data := make([]field.Element, spec.size)
			for i := range data {
				data[i] = field.PseudoRand(rng)
			}
			entries = append(entries, entry{id: id, data: data})
			idx++
		}
	}

	define := func(b *wizard.Builder) {
		for _, e := range entries {
			b.RegisterCommit(e.id, len(e.data))
		}
	}
	prove := func(run *wizard.ProverRuntime) {
		for _, e := range entries {
			run.AssignColumn(e.id, smartvectors.NewRegular(e.data))
		}
	}

	compiledML := wizard.Compile(
		define,
		multilinvortex.InsertBootstrapperOpenings,
		multilinvortex.CompileRound, multilineareval.Batch,
		multilinvortex.CompileRound, multilineareval.Batch,
		multilinvortex.CompileRound, multilineareval.Batch,
		multilinvortex.CompileRound, multilineareval.Batch,
		multilinvortex.CompileRound, multilineareval.Batch,
		dummy.Compile,
	)
	compiledUniNR := wizard.Compile(
		define,
		compiler.Arcane(compiler.WithTargetColSize(1<<10), compiler.WithStitcherMinSize(16)),
		insertUnivariateBootstrapperQuery,
		vortex.Compile(2, false,
			vortex.ForceNumOpenedColumns(32),
			vortex.WithSISParams(&benchSISParams),
			vortex.WithOptionalSISHashingThreshold(1),
		),
		dummy.Compile,
	)
	compiledUni := wizard.Compile(
		define,
		compiler.Arcane(compiler.WithTargetColSize(1<<10), compiler.WithStitcherMinSize(16)),
		insertUnivariateBootstrapperQuery,
		vortex.Compile(2, false,
			vortex.ForceNumOpenedColumns(32),
			vortex.WithSISParams(&benchSISParams),
			vortex.WithOptionalSISHashingThreshold(1),
		),
		selfrecursion.SelfRecurse, cleanup.CleanUp, poseidon2.CompilePoseidon2,
		compiler.Arcane(compiler.WithTargetColSize(1<<9), compiler.WithStitcherMinSize(16)),
		insertUnivariateBootstrapperQuery,
		vortex.Compile(4, false,
			vortex.ForceNumOpenedColumns(16),
			vortex.WithSISParams(&benchSISParams),
			vortex.WithOptionalSISHashingThreshold(1),
		),
		selfrecursion.SelfRecurse, cleanup.CleanUp, poseidon2.CompilePoseidon2,
		compiler.Arcane(compiler.WithTargetColSize(1<<8), compiler.WithStitcherMinSize(16)),
		insertUnivariateBootstrapperQuery,
		vortex.Compile(8, false,
			vortex.ForceNumOpenedColumns(8),
			vortex.WithSISParams(&benchSISParams),
			vortex.WithOptionalSISHashingThreshold(1),
		),
		selfrecursion.SelfRecurse, cleanup.CleanUp, poseidon2.CompilePoseidon2,
		compiler.Arcane(compiler.WithTargetColSize(1<<7), compiler.WithStitcherMinSize(16)),
		insertUnivariateBootstrapperQuery,
		vortex.Compile(16, false,
			vortex.ForceNumOpenedColumns(8),
			vortex.WithOptionalSISHashingThreshold(1<<20),
			vortex.PremarkAsSelfRecursed(),
		),
		dummy.Compile,
	)
	compiledBaseline := wizard.Compile(define, dummy.Compile)

	b.Run("ML", func(b *testing.B) {
		for range b.N {
			start := time.Now()
			wizard.Prove(compiledML, prove)
			b.ReportMetric(float64(time.Since(start).Milliseconds()), "ms/op")
		}
	})
	b.Run("UniNR", func(b *testing.B) {
		for range b.N {
			start := time.Now()
			wizard.Prove(compiledUniNR, prove)
			b.ReportMetric(float64(time.Since(start).Milliseconds()), "ms/op")
		}
	})
	b.Run("Univariate", func(b *testing.B) {
		for range b.N {
			start := time.Now()
			wizard.Prove(compiledUni, prove)
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

// BenchmarkRecompilationOverhead demonstrates the cost of "recompilation on
// distribution change": when the column distribution changes between proof
// batches the compiled IOP is stale and must be rebuilt.
//
// We model this by compiling + proving for two distributions:
//   - base:     4 sizes (nv = 6, 8, 10, 12)
//   - extended: same + a new 5th group (nv = 14, added by a new zkEVM module)
//
// KEY OBSERVATIONS
//
// ML pipeline (InsertBootstrapperOpenings + CompileRound×5):
//   - Zero pipeline-code change between base and extended — InsertBootstrapperOpenings
//     discovers the new nv=14 group automatically.
//   - Compilation is lighter: no Arcane, no SelfRecurse passes.
//
// Univariate pipeline (Arcane + Vortex + SelfRecurse×4):
//   - WithTargetColSize(1<<10) splits 16K→16 sub-cols automatically, so the
//     pipeline code also compiles both distributions unchanged.
//   - BUT the developer must manually verify ForceNumOpenedColumns, round targets,
//     and round count after each distribution change — no auto-calibration knob.
//   - Compilation includes multiple SelfRecurse+CleanUp+Poseidon2 passes.
//
// Sub-benchmarks report compile_ms / prove_ms / total_ms separately.
// Run with: go test -bench BenchmarkRecompilationOverhead -benchtime=1x -v
func BenchmarkRecompilationOverhead(b *testing.B) {
	for _, dist := range []struct {
		name string
		spec []colSpec
	}{
		{"base", dynamicDist},
		{"extended", dynamicDistExtended},
	} {
		distName := dist.name
		distSpec := dist.spec

		rng := rand.New(rand.NewPCG(99, 0))
		type entry struct {
			id   ifaces.ColID
			data []field.Element
		}
		var entries []entry
		cidx := 0
		for _, spec := range distSpec {
			for k := 0; k < spec.count; k++ {
				id := ifaces.ColIDf("RCOL_%d", cidx)
				d := make([]field.Element, spec.size)
				for i := range d {
					d[i] = field.PseudoRand(rng)
				}
				entries = append(entries, entry{id: id, data: d})
				cidx++
			}
		}

		define := func(bld *wizard.Builder) {
			for _, e := range entries {
				bld.RegisterCommit(e.id, len(e.data))
			}
		}
		prove := func(run *wizard.ProverRuntime) {
			for _, e := range entries {
				run.AssignColumn(e.id, smartvectors.NewRegular(e.data))
			}
		}

		type pipeline struct {
			name    string
			compile func() *wizard.CompiledIOP
		}

		pipelines := []pipeline{
			{
				name: "ML",
				compile: func() *wizard.CompiledIOP {
					// SAME code for base and extended — zero developer changes needed.
					return wizard.Compile(define,
						multilinvortex.InsertBootstrapperOpenings,
						multilinvortex.CompileRound, multilineareval.Batch,
						multilinvortex.CompileRound, multilineareval.Batch,
						multilinvortex.CompileRound, multilineareval.Batch,
						multilinvortex.CompileRound, multilineareval.Batch,
						multilinvortex.CompileRound, multilineareval.Batch,
						dummy.Compile,
					)
				},
			},
			{
				name: "Univariate",
				// NOTE: Arcane(target=1<<10) splits the new 16K group into 16×1K
				// sub-columns automatically, so the code below compiles both
				// distributions unchanged — same as ML in this regard.
				// The developer advantage of ML is that InsertBootstrapperOpenings
				// self-calibrates soundness and open counts; Univariate needs manual
				// re-inspection of every parameter after a distribution change.
				compile: func() *wizard.CompiledIOP {
					return wizard.Compile(define,
						compiler.Arcane(compiler.WithTargetColSize(1<<10), compiler.WithStitcherMinSize(16)),
						insertUnivariateBootstrapperQuery,
						vortex.Compile(2, false,
							vortex.ForceNumOpenedColumns(32),
							vortex.WithSISParams(&benchSISParams),
							vortex.WithOptionalSISHashingThreshold(1),
						),
						selfrecursion.SelfRecurse, cleanup.CleanUp, poseidon2.CompilePoseidon2,
						compiler.Arcane(compiler.WithTargetColSize(1<<9), compiler.WithStitcherMinSize(16)),
						insertUnivariateBootstrapperQuery,
						vortex.Compile(4, false,
							vortex.ForceNumOpenedColumns(16),
							vortex.WithSISParams(&benchSISParams),
							vortex.WithOptionalSISHashingThreshold(1),
						),
						selfrecursion.SelfRecurse, cleanup.CleanUp, poseidon2.CompilePoseidon2,
						compiler.Arcane(compiler.WithTargetColSize(1<<8), compiler.WithStitcherMinSize(16)),
						insertUnivariateBootstrapperQuery,
						vortex.Compile(8, false,
							vortex.ForceNumOpenedColumns(8),
							vortex.WithSISParams(&benchSISParams),
							vortex.WithOptionalSISHashingThreshold(1),
						),
						selfrecursion.SelfRecurse, cleanup.CleanUp, poseidon2.CompilePoseidon2,
						compiler.Arcane(compiler.WithTargetColSize(1<<7), compiler.WithStitcherMinSize(16)),
						insertUnivariateBootstrapperQuery,
						vortex.Compile(16, false,
							vortex.ForceNumOpenedColumns(8),
							vortex.WithOptionalSISHashingThreshold(1<<20),
							vortex.PremarkAsSelfRecursed(),
						),
						dummy.Compile,
					)
				},
			},
		}

		for _, p := range pipelines {
			p := p
			b.Run(p.name+"/"+distName, func(b *testing.B) {
				for range b.N {
					tCompile := time.Now()
					compiled := p.compile()
					compileMs := time.Since(tCompile).Milliseconds()

					tProve := time.Now()
					wizard.Prove(compiled, prove)
					proveMs := time.Since(tProve).Milliseconds()

					b.ReportMetric(float64(compileMs), "compile_ms")
					b.ReportMetric(float64(proveMs), "prove_ms")
					b.ReportMetric(float64(compileMs+proveMs), "total_ms")
				}
			})
		}
	}
}
