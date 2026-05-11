package multilinvortex_test

import (
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/cleanup"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/multilineareval"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/multilinvortex"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/poseidon2"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/selfrecursion"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/vortex"
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

// TestMultilinVortexPackedRoundtrip verifies that CompileRoundPacked produces a
// correct prove→verify path. It uses the same define/prove helpers as
// TestMultilinVortexRoundtrip but replaces CompileRound with CompileRoundPacked.
func TestMultilinVortexPackedRoundtrip(t *testing.T) {
	cases := []struct {
		name    string
		numCols int
		numVars int
	}{
		{"1col_n4", 1, 4},
		{"2col_n4", 2, 4},
		{"3col_n6", 3, 6},
		{"5col_n6", 5, 6}, // K=5, L=3, KPow2=8 (non-power-of-2 K)
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			rng := rand.New(rand.NewPCG(uint64(tc.numCols*200), uint64(tc.numVars)))
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

			// CompileRoundPacked creates K UCols/RowClaims on the ONE packed column.
			// We must use Batch (not Compile) after a packed round because Compile's
			// duplicate-column check rejects K queries that all reference the same
			// packed UAlpha. Batch handles them as K separate queries on one oracle.
			compiled := wizard.Compile(define,
				multilinvortex.CompileRoundPacked,
				multilineareval.Batch,
				multilinvortex.CompileRound,
				multilineareval.Batch,
				multilinvortex.CompileRound,
				multilineareval.Batch,
				dummy.Compile,
			)
			proof := wizard.Prove(compiled, prove)
			require.NoError(t, wizard.Verify(compiled, proof))
		})
	}
}

// TestMultilinVortexPaddedRoundtrip exercises the padded fast path in
// ProverAction.Run by assigning each input column as a right-zero-padded
// smartvector with len(window) < declaredSize. The non-zero portion must agree
// with the y values fed into the MultilinearEval claim. A correct prove→verify
// proves that skipping rows entirely in the zero suffix produces the same
// commitments and openings as the full-iteration path.
func TestMultilinVortexPaddedRoundtrip(t *testing.T) {
	cases := []struct {
		name       string
		numCols    int
		numVars    int
		actualVars int // log2 of non-zero prefix length; must satisfy 0 < 2^actualVars < 2^numVars
	}{
		{"1col_n6_actual_n3", 1, 6, 3}, // 8-of-64
		{"2col_n6_actual_n4", 2, 6, 4}, // 16-of-64
		{"3col_n8_actual_n5", 3, 8, 5}, // 32-of-256
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			rng := rand.New(rand.NewPCG(uint64(tc.numCols*300+tc.actualVars), uint64(tc.numVars)))
			size := 1 << tc.numVars
			actualSize := 1 << tc.actualVars

			// padded[k] is the full size-N vector: first actualSize entries random,
			// rest zero. window[k] is the non-zero prefix.
			window := make([][]field.Element, tc.numCols)
			padded := make([][]field.Element, tc.numCols)
			for k := 0; k < tc.numCols; k++ {
				window[k] = make([]field.Element, actualSize)
				for j := range window[k] {
					window[k][j] = field.PseudoRand(rng)
				}
				padded[k] = make([]field.Element, size)
				copy(padded[k], window[k])
			}

			point := make([]fext.Element, tc.numVars)
			for i := range point {
				point[i] = fext.PseudoRand(rng)
			}

			ys := make([]fext.Element, tc.numCols)
			for k := 0; k < tc.numCols; k++ {
				vals := make([]fext.Element, size)
				for j, v := range padded[k] {
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
						smartvectors.RightZeroPadded(window[k], size))
				}
				run.AssignMultilinearExtShared("MLEVAL", point, ys...)
			}

			compiled := wizard.Compile(define,
				multilineareval.Compile,
				multilinvortex.CompileRound,
				multilineareval.Batch,
				multilinvortex.CompileRound,
				multilineareval.Batch,
				multilinvortex.CompileRound,
				multilineareval.Batch,
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
				multilinvortex.CompileRound,
				multilineareval.Compile,
				multilinvortex.CompileRound,
				multilineareval.Compile,
				dummy.Compile,
			)
			proof := wizard.Prove(compiled, prove)
			require.NoError(t, wizard.Verify(compiled, proof))
		})
	}
}

func TestTerminalCommittedQueriesRoundtrip(t *testing.T) {
	cases := []struct {
		name       string
		compile0   func(*wizard.CompiledIOP)
	}{
		{name: "unpacked", compile0: multilinvortex.CompileRound},
		{name: "packed-fallback", compile0: multilinvortex.CompileRoundPacked},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			rng := rand.New(rand.NewPCG(11, 19))
			const numCols = 2
			const numVars = 1
			size := 1 << numVars

			colData := make([][]field.Element, numCols)
			for k := 0; k < numCols; k++ {
				colData[k] = make([]field.Element, size)
				for j := range colData[k] {
					colData[k][j] = field.PseudoRand(rng)
				}
			}

			point := []fext.Element{fext.PseudoRand(rng)}
			ys := make([]fext.Element, numCols)
			for k := 0; k < numCols; k++ {
				vals := make([]fext.Element, size)
				for j, v := range colData[k] {
					vals[j].B0.A0 = v
				}
				ys[k] = evalMultilin(vals, point)
			}

			define := func(b *wizard.Builder) {
				cols := make([]ifaces.Column, numCols)
				for k := 0; k < numCols; k++ {
					cols[k] = b.RegisterCommit(ifaces.ColIDf("COL_%d", k), size)
				}
				b.CompiledIOP.InsertMultilinear(0, "MLEVAL", cols)
			}
			prove := func(run *wizard.ProverRuntime) {
				for k := 0; k < numCols; k++ {
					run.AssignColumn(ifaces.ColIDf("COL_%d", k), smartvectors.NewRegular(colData[k]))
				}
				run.AssignMultilinearExtShared("MLEVAL", point, ys...)
			}

			compiled := wizard.Compile(
				define,
				tc.compile0,
				multilineareval.Batch,
				multilinvortex.CompileRound,
				dummy.Compile,
			)
			proof := wizard.Prove(compiled, prove)
			require.NoError(t, wizard.Verify(compiled, proof))
		})
	}
}

func TestCompileRoundPackedFallsBackToUnpackedForCommittedInputs(t *testing.T) {
	define := func(b *wizard.Builder) {
		cols := []ifaces.Column{
			b.RegisterCommit("COL_0", 16),
			b.RegisterCommit("COL_1", 16),
		}
		b.CompiledIOP.InsertMultilinear(0, "MLEVAL", cols)
	}

	compiled := wizard.Compile(define, multilinvortex.CompileRoundPacked)

	var hasPackedUAlpha bool
	var hasUnpackedUAlpha bool
	for _, name := range compiled.Columns.AllKeysProof() {
		switch {
		case strings.HasPrefix(string(name), "MLVORTEX_UALPHA_packed_"):
			hasPackedUAlpha = true
		case strings.HasPrefix(string(name), "MLVORTEX_UALPHA_0_"):
			hasUnpackedUAlpha = true
		}
	}

	require.False(t, hasPackedUAlpha, "committed inputs must not use packed UAlpha columns")
	require.True(t, hasUnpackedUAlpha, "committed inputs should fall back to unpacked UAlpha columns")
}

// TestDynamicColumnSizes places three commit-and-prove pipelines side-by-side
// for a witness with four different column sizes (64, 256, 1K, 4K elements).
// The aim is to make the code-complexity difference tangible, not to benchmark.
//
//   - ML:         same chain regardless of the size distribution; zero config.
//   - UniNR:      single Vortex round — needs Arcane + a custom helper.
//   - Univariate: production-style 4-round pipeline with SelfRecurse — needs
//                 four Arcane passes each with hand-tuned target sizes, plus
//                 SelfRecurse/CleanUp/Poseidon2 boilerplate between rounds.
func TestDynamicColumnSizes(t *testing.T) {
	dist := []colSpec{
		{count: 16, size: 1 << 6},  // 64-element  → stitched 16→1 at 1K target
		{count: 4, size: 1 << 8},   // 256-element → stitched  4→1 at 1K target
		{count: 2, size: 1 << 10},  // 1K          → already at target
		{count: 1, size: 1 << 12},  // 4K          → split     1→4 at 1K target
	}

	rng := rand.New(rand.NewPCG(7, 0))
	type entry struct {
		id   ifaces.ColID
		data []field.Element
	}
	var entries []entry
	idx := 0
	for _, spec := range dist {
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

	// ── ML ───────────────────────────────────────────────────────────────────
	// InsertBootstrapperOpenings discovers nv=6, 8, 10, 12 and creates one ML
	// group per size. Five CompileRound+Batch passes reduce any numVars ≤ 12 to
	// a single claim. Adding a new column size requires zero changes here.
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

	// ── UniNR (single Vortex round, no recursion) ────────────────────────────
	// WithTargetColSize(1<<10) must match the distribution: too large wastes
	// cells; too small causes excessive splitting. insertUnivariateBootstrapper-
	// Query is a custom helper (not part of any framework package) that must run
	// after Arcane so all columns already have uniform size.
	compiledUniNR := wizard.Compile(
		define,
		compiler.Arcane(
			compiler.WithTargetColSize(1<<10),
			compiler.WithStitcherMinSize(16),
		),
		insertUnivariateBootstrapperQuery, // custom helper — not in any framework package
		vortex.Compile(
			2, false,
			vortex.ForceNumOpenedColumns(32),
			vortex.WithSISParams(&benchSISParams),
			vortex.WithOptionalSISHashingThreshold(1),
		),
		dummy.Compile,
	)

	// ── Univariate (4 rounds + SelfRecurse, production-style) ────────────────
	// Each round needs Arcane with a target size calibrated to what SelfRecurse
	// produced in the previous round (discovered by probing comp.Columns at
	// compile time). Any change to the distribution may invalidate these
	// constants. The SelfRecurse+CleanUp+Poseidon2 triple is mandatory boilerplate
	// between every pair of Vortex rounds.
	compiledUni := wizard.Compile(
		define,
		// Round 1: normalise heterogeneous sizes to 1K.
		compiler.Arcane(compiler.WithTargetColSize(1<<10), compiler.WithStitcherMinSize(16)),
		insertUnivariateBootstrapperQuery,
		vortex.Compile(2, false,
			vortex.ForceNumOpenedColumns(32),
			vortex.WithSISParams(&benchSISParams),
			vortex.WithOptionalSISHashingThreshold(1),
		),
		selfrecursion.SelfRecurse, cleanup.CleanUp, poseidon2.CompilePoseidon2,
		// Round 2: SelfRecurse emits columns ≤ 2K; target chosen to fit.
		compiler.Arcane(compiler.WithTargetColSize(1<<9), compiler.WithStitcherMinSize(16)),
		insertUnivariateBootstrapperQuery,
		vortex.Compile(4, false,
			vortex.ForceNumOpenedColumns(16),
			vortex.WithSISParams(&benchSISParams),
			vortex.WithOptionalSISHashingThreshold(1),
		),
		selfrecursion.SelfRecurse, cleanup.CleanUp, poseidon2.CompilePoseidon2,
		// Round 3: sizes grow after the second recursion; 256 fits them.
		compiler.Arcane(compiler.WithTargetColSize(1<<8), compiler.WithStitcherMinSize(16)),
		insertUnivariateBootstrapperQuery,
		vortex.Compile(8, false,
			vortex.ForceNumOpenedColumns(8),
			vortex.WithSISParams(&benchSISParams),
			vortex.WithOptionalSISHashingThreshold(1),
		),
		selfrecursion.SelfRecurse, cleanup.CleanUp, poseidon2.CompilePoseidon2,
		// Round 4 (final): PremarkAsSelfRecursed so the gnark verifier can
		// consume this proof without another recursion pass.
		compiler.Arcane(compiler.WithTargetColSize(1<<7), compiler.WithStitcherMinSize(16)),
		insertUnivariateBootstrapperQuery,
		vortex.Compile(16, false,
			vortex.ForceNumOpenedColumns(8),
			vortex.WithOptionalSISHashingThreshold(1<<20),
			vortex.PremarkAsSelfRecursed(),
		),
		dummy.Compile,
	)

	t.Run("ML", func(t *testing.T) {
		proof := wizard.Prove(compiledML, prove)
		require.NoError(t, wizard.Verify(compiledML, proof))
	})
	t.Run("UniNR", func(t *testing.T) {
		proof := wizard.Prove(compiledUniNR, prove)
		require.NoError(t, wizard.Verify(compiledUniNR, proof))
	})
	t.Run("Univariate", func(t *testing.T) {
		proof := wizard.Prove(compiledUni, prove)
		require.NoError(t, wizard.Verify(compiledUni, proof))
	})
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
