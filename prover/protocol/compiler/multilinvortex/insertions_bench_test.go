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
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/multilineareval"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/multilinvortex"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/vortex"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
	"github.com/sirupsen/logrus"
)

func TestMain(m *testing.M) {
	logrus.SetLevel(logrus.ErrorLevel)
	m.Run()
}

// colSpec describes one column: count × size.
type colSpec struct {
	count, size int
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

// benchBootstrapperProver is the shared implementation for all bootstrapper
// benchmarks. It compiles the ML and UniNR pipelines against the given column
// distribution, then runs each sub-benchmark for b.N iterations.
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

	// UniNR path: single Arcane+Vortex round, no self-recursion.
	// Directly comparable to ML: same witness, same statement, one commitment round.
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

	b.Run("UniNR", func(b *testing.B) {
		for range b.N {
			start := time.Now()
			wizard.Prove(compiledUniNR, prove)
			b.ReportMetric(float64(time.Since(start).Milliseconds()), "ms/op")
		}
	})
}

// BenchmarkFullScaleProver runs ML vs UniNR on the actual mainnet column
// distribution (20.54B cells, ~164 GiB of field data).
// Expected peak RSS: ~500 GiB.  Expected wall time: ~25 min (ML), ~60 min
// (UniNR) on a 192-core machine.  Run with -benchtime=1x -timeout=2h.
func BenchmarkFullScaleProver(b *testing.B) {
	benchBootstrapperProver(b, productionFull)
}

// scalingByPolySizeDists defines homogeneous distributions sweeping polynomial
// size from 2^14 to 2^24.  Total cell count is held at ~8M so all benchmarks
// complete in a few minutes each.
var scalingByPolySizeDists = []struct {
	name string
	dist []colSpec
}{
	{"nv14_500cols", []colSpec{{count: 500, size: 1 << 14}}}, // 500 × 16 K  =  8.0 M
	{"nv17_62cols", []colSpec{{count: 62, size: 1 << 17}}},   //  62 × 128 K =  7.9 M
	{"nv20_8cols", []colSpec{{count: 8, size: 1 << 20}}},     //   8 × 1 M   =  8.0 M
	{"nv22_2cols", []colSpec{{count: 2, size: 1 << 22}}},     //   2 × 4 M   =  8.0 M
	{"nv23_1col", []colSpec{{count: 1, size: 1 << 23}}},      //   1 × 8 M   =  8.0 M
}

// BenchmarkScalingByPolySize compares ML and UniNR prover time as polynomial
// size grows from 2^14 to 2^24, with total cell count fixed at ~8M.
// This isolates the effect of polynomial size on each pipeline:
//   - ML: each numVars gets its own evaluation point; no Arcane normalization.
//   - UniNR: Arcane stitches small columns or splits large ones to target 1M,
//     then one Vortex round.
//
// Run with: go test -bench "BenchmarkScalingByPolySize/(ML|UniNR)" -benchtime=1x -timeout=2h -v
func BenchmarkScalingByPolySize(b *testing.B) {
	for _, tc := range scalingByPolySizeDists {
		tc := tc
		b.Run(tc.name, func(b *testing.B) {
			benchBootstrapperProver(b, tc.dist)
		})
	}
}
