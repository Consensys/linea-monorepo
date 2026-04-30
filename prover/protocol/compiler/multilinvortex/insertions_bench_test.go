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
// The total cell count is deliberately kept at ~1/64 of the real 23.75B to
// keep the benchmark under a minute; multiply measured latency by 64 to
// estimate production overhead.
var productionApprox = []colSpec{
	{count: 2000, size: 1 << 10}, // 2K  → 2.0M cells
	{count: 500, size: 1 << 14},  // 16K → 8.0M cells
	{count: 100, size: 1 << 17},  // 128K → 12.8M cells
	{count: 20, size: 1 << 20},   // 1M  → 20.0M cells
	{count: 5, size: 1 << 22},    // 4M  → 20.0M cells
	{count: 1, size: 1 << 24},    // 16M → 16.0M cells
	// Total: ~78.8M cells (≈ 1/300 of production 23.75B)
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

// BenchmarkMLBootstrapperProver compares the multilinear-vortex bootstrapper
// commitment scheme against the production univariate-vortex pipeline on the
// same synthetic column distribution.
//
// Both endpoints are cryptographically equivalent:
//   - ML path:         InsertBootstrapperOpenings + 7×(Compile+CompileAllRound)
//     Seven rounds fully reduce any numVars ≤ 128; the final
//     dummy.Compile is a no-op terminator.
//   - Univariate path: Arcane(1M) + vortex(round 1, blowUp=2, 256 cols)
//   - 3× [SelfRecurse + CleanUp + poseidon2 + Arcane + vortex]
//     matching the four-round full.go production pipeline.
//   - Baseline:        dummy compile — column-assign overhead only (lower bound).
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

	// ML path: Compile halves numVars first so that CompileAllRound only
	// expands within the halved space. Seven pairs cover any numVars ≤ 128.
	// dummy.Compile is a no-op terminator (all queries resolved after 7 rounds).
	compiledML := wizard.Compile(
		define,
		multilinvortex.InsertBootstrapperOpenings,
		multilinvortex.Compile,
		multilinvortex.CommitMLColumns,
		multilinvortex.CommitOriginalMLColumns, // binds round-0 cols to FS before α
		multilineareval.CompileAllRound,
		multilinvortex.Compile,
		multilinvortex.CommitMLColumns,
		multilineareval.CompileAllRound,
		multilinvortex.Compile,
		multilinvortex.CommitMLColumns,
		multilineareval.CompileAllRound,
		multilinvortex.Compile,
		multilinvortex.CommitMLColumns,
		multilineareval.CompileAllRound,
		multilinvortex.Compile,
		multilinvortex.CommitMLColumns,
		multilineareval.CompileAllRound,
		multilinvortex.Compile,
		multilinvortex.CommitMLColumns,
		multilineareval.CompileAllRound,
		multilinvortex.Compile,
		multilinvortex.CommitMLColumns,
		multilineareval.CompileAllRound,
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

	b.Run("Baseline", func(b *testing.B) {
		for range b.N {
			start := time.Now()
			wizard.Prove(compiledBaseline, prove)
			b.ReportMetric(float64(time.Since(start).Milliseconds()), "ms/op")
		}
	})
}
