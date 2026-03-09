package finalwrap

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	frBn254 "github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/linea-monorepo/prover/circuits"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2_koalabear"
	"github.com/consensys/linea-monorepo/prover/crypto/ringsis"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/cleanup"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/globalcs"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/localcs"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/logdata"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/logderivativesum"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/mpts"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/poseidon2"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/recursion"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/selfrecursion"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/splitextension"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/univariates"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/vortex"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/tree"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/stretchr/testify/require"
)

func init() {
	if err := poseidon2_koalabear.RegisterGates(); err != nil {
		panic(err)
	}
}

// toyWizardSuite returns a minimal compilation suite for a toy wizard IOP.
// It compiles through Vortex with IsLastRound=true so the compiled IOP
// is suitable for wrapping in a BN254 final wrap circuit.
func toyWizardSuite() []func(*wizard.CompiledIOP) {
	return []func(*wizard.CompiledIOP){
		logderivativesum.CompileLookups,
		localcs.Compile,
		globalcs.Compile,
		univariates.Naturalize,
		mpts.Compile(),
		splitextension.CompileSplitExtToBase,
		vortex.Compile(
			2, true, // IsLastRound=true since this is the final wizard proof
			vortex.ForceNumOpenedColumns(4),
			vortex.WithSISParams(&ringsis.StdParams),
			vortex.WithOptionalSISHashingThreshold(64),
		),
	}
}

// realisticWizardParams mirrors the "realistic-segment" benchmark case.
// Shared between BN254 (this test) and BLS (e2e_bls_test.go) for comparison.
type realisticWizardParams struct {
	NumPermutations int
	NumLookups      int
	NumProjections  int
	NumFibo         int
	NumCol          int
	NumRow          int
}

// defaultRealisticParams matches the "realistic-segment" case from
// BenchmarkCompilerWithSelfRecursionAndGnarkVerifier in standard_benchmark_test.go:
// 5 permutations, 50 lookups, 5 projections, 200 fibonacci cols, NumCol=3.
// NumRow=1<<16 equals targetColSize so Arcane produces one sub-column per raw column
// (no splitting needed). This is 16× more data than the old 1<<14 default.
var defaultRealisticParams = realisticWizardParams{
	NumPermutations: 5,
	NumLookups:      50,
	NumProjections:  5,
	NumFibo:         200,
	NumCol:          3,
	NumRow:          1 << 16,
}

// realisticWizardDefine defines a wizard IOP with multiple module types
// (permutations, lookups, projections, fibonacci).
func realisticWizardDefine(p realisticWizardParams) func(*wizard.Builder) {
	return func(bui *wizard.Builder) {
		comp := bui.CompiledIOP

		// Permutation modules
		for n := 0; n < p.NumPermutations; n++ {
			var a, b []ifaces.Column
			for i := 0; i < p.NumCol; i++ {
				ai := comp.InsertCommit(0, ifaces.ColID(fmt.Sprintf("PERM_%d_A_%d", n, i)), p.NumRow, true)
				bi := comp.InsertCommit(0, ifaces.ColID(fmt.Sprintf("PERM_%d_B_%d", n, i)), p.NumRow, true)
				a = append(a, ai)
				b = append(b, bi)
			}
			comp.InsertPermutation(0, ifaces.QueryID(fmt.Sprintf("PERM_%d_Q", n)), a, b)
		}

		// Lookup (inclusion) modules
		for n := 0; n < p.NumLookups; n++ {
			var s, t []ifaces.Column
			for i := 0; i < p.NumCol; i++ {
				si := comp.InsertCommit(0, ifaces.ColID(fmt.Sprintf("LOOKUP_%d_S_%d", n, i)), p.NumRow, true)
				ti := comp.InsertCommit(0, ifaces.ColID(fmt.Sprintf("LOOKUP_%d_T_%d", n, i)), p.NumRow, true)
				s = append(s, si)
				t = append(t, ti)
			}
			comp.InsertInclusion(0, ifaces.QueryID(fmt.Sprintf("LOOKUP_%d_Q", n)), t, s)
		}

		// Projection modules
		for n := 0; n < p.NumProjections; n++ {
			aFilter := comp.InsertCommit(0, ifaces.ColID(fmt.Sprintf("PROJ_%d_AF", n)), p.NumRow, true)
			bFilter := comp.InsertCommit(0, ifaces.ColID(fmt.Sprintf("PROJ_%d_BF", n)), p.NumRow, true)
			var a, b []ifaces.Column
			for i := 0; i < p.NumCol; i++ {
				ai := comp.InsertCommit(0, ifaces.ColID(fmt.Sprintf("PROJ_%d_A_%d", n, i)), p.NumRow, true)
				bi := comp.InsertCommit(0, ifaces.ColID(fmt.Sprintf("PROJ_%d_B_%d", n, i)), p.NumRow, true)
				a = append(a, ai)
				b = append(b, bi)
			}
			comp.InsertProjection(ifaces.QueryID(fmt.Sprintf("PROJ_%d_Q", n)), query.ProjectionInput{
				ColumnA: a,
				ColumnB: b,
				FilterA: aFilter,
				FilterB: bFilter,
			})
		}

		// Fibonacci modules
		for n := 0; n < p.NumFibo; n++ {
			a := comp.InsertCommit(0, ifaces.ColID(fmt.Sprintf("FIBO_%d", n)), p.NumRow, true)
			comp.InsertGlobal(0, ifaces.QueryID(fmt.Sprintf("FIBO_%d_G", n)),
				sym.Sub(a, column.Shift(a, -1), column.Shift(a, -2)))
			comp.InsertLocal(0, ifaces.QueryID(fmt.Sprintf("FIBO_%d_L0", n)),
				sym.Sub(a, 1))
			comp.InsertLocal(0, ifaces.QueryID(fmt.Sprintf("FIBO_%d_L1", n)),
				sym.Sub(column.Shift(a, 1), 1))
		}
	}
}

// realisticWizardAssign returns a prover function that assigns all columns.
func realisticWizardAssign(p realisticWizardParams) func(*wizard.ProverRuntime) {
	return func(run *wizard.ProverRuntime) {
		zero := smartvectors.NewConstant(field.Zero(), p.NumRow)
		ones := smartvectors.NewConstant(field.One(), p.NumRow)

		for n := 0; n < p.NumPermutations; n++ {
			for i := 0; i < p.NumCol; i++ {
				run.AssignColumn(ifaces.ColID(fmt.Sprintf("PERM_%d_A_%d", n, i)), zero)
				run.AssignColumn(ifaces.ColID(fmt.Sprintf("PERM_%d_B_%d", n, i)), zero)
			}
		}

		for n := 0; n < p.NumLookups; n++ {
			for i := 0; i < p.NumCol; i++ {
				run.AssignColumn(ifaces.ColID(fmt.Sprintf("LOOKUP_%d_S_%d", n, i)), zero)
				run.AssignColumn(ifaces.ColID(fmt.Sprintf("LOOKUP_%d_T_%d", n, i)), zero)
			}
		}

		for n := 0; n < p.NumProjections; n++ {
			run.AssignColumn(ifaces.ColID(fmt.Sprintf("PROJ_%d_AF", n)), ones)
			run.AssignColumn(ifaces.ColID(fmt.Sprintf("PROJ_%d_BF", n)), ones)
			for i := 0; i < p.NumCol; i++ {
				run.AssignColumn(ifaces.ColID(fmt.Sprintf("PROJ_%d_A_%d", n, i)), zero)
				run.AssignColumn(ifaces.ColID(fmt.Sprintf("PROJ_%d_B_%d", n, i)), zero)
			}
		}

		fibo := make([]field.Element, p.NumRow)
		fibo[0] = field.One()
		fibo[1] = field.One()
		for i := 2; i < p.NumRow; i++ {
			fibo[i].Add(&fibo[i-1], &fibo[i-2])
		}
		fiboVec := smartvectors.NewRegular(fibo)
		for n := 0; n < p.NumFibo; n++ {
			run.AssignColumn(ifaces.ColID(fmt.Sprintf("FIBO_%d", n)), fiboVec)
		}
	}
}

// realisticWizardCompile compiles a realistic wizard IOP using the pipeline:
// Arcane → Vortex → (SelfRecursion → Poseidon2 → Arcane → Vortex) × nbIteration (IsLastRound=true on final).
// This matches realisticBLSWizardCompile in e2e_bls_test.go for fair comparison.
func realisticWizardCompile(p realisticWizardParams, nbIteration int) *wizard.CompiledIOP {
	const (
		rsInverseRate     = 16
		nbOpenedColumns   = 64
		initTargetColSize = 1 << 16
		midTargetRowSize  = 1 << 8
		lastTargetRowSize = 1 << 10
	)

	comp := wizard.Compile(
		realisticWizardDefine(p),
		compiler.Arcane(
			compiler.WithTargetColSize(initTargetColSize),
			compiler.WithStitcherMinSize(1<<1),
		),
		vortex.Compile(
			rsInverseRate, false,
			vortex.WithOptionalSISHashingThreshold(512),
			vortex.ForceNumOpenedColumns(nbOpenedColumns),
			vortex.WithSISParams(&ringsis.StdParams),
		),
	)

	for i := 0; i < nbIteration-1; i++ {
		selfrecursion.SelfRecurse(comp)

		stats := logdata.GetWizardStats(comp)
		rowSize := utils.NextPowerOfTwo(utils.DivCeil(stats.NumCellsCommitted, midTargetRowSize))
		wizard.ContinueCompilation(comp,
			poseidon2.CompilePoseidon2,
			compiler.Arcane(
				compiler.WithTargetColSize(rowSize),
				compiler.WithStitcherMinSize(1<<1),
			),
		)
		wizard.ContinueCompilation(comp,
			vortex.Compile(
				rsInverseRate, false,
				vortex.ForceNumOpenedColumns(nbOpenedColumns),
				vortex.WithSISParams(&ringsis.StdParams),
			),
		)

	}

	// Last iteration: SelfRecursion → Poseidon2 → Arcane → Vortex (IsLastRound=true).
	selfrecursion.SelfRecurse(comp)
	stats := logdata.GetWizardStats(comp)
	rowSize := utils.NextPowerOfTwo(utils.DivCeil(stats.NumCellsCommitted, lastTargetRowSize))
	wizard.ContinueCompilation(comp,
		poseidon2.CompilePoseidon2,
		compiler.Arcane(
			compiler.WithTargetColSize(rowSize),
			compiler.WithStitcherMinSize(1<<1),
		),
	)
	// IsLastRound=true
	wizard.ContinueCompilation(comp,
		vortex.Compile(
			rsInverseRate, true, // IsLastRound=true
			vortex.ForceNumOpenedColumns(nbOpenedColumns),
			vortex.WithOptionalSISHashingThreshold(1<<20),
		),
	)

	return comp
}

// TestMakeCS verifies that the BN254 final wrap circuit can be compiled
// around a toy wizard IOP.
func TestMakeCS(t *testing.T) {
	define := func(bui *wizard.Builder) {
		a := bui.RegisterCommit(ifaces.ColID("FW_A"), 64)
		b := bui.RegisterCommit(ifaces.ColID("FW_B"), 64)
		bui.Inclusion(ifaces.QueryID("FW_Q"), []ifaces.Column{a}, []ifaces.Column{b})
	}

	comp := wizard.Compile(define, toyWizardSuite()...)
	require.NotNil(t, comp)
	t.Logf("Toy wizard IOP compiled: %d rounds", comp.NumRounds())

	ccs, err := MakeCS(comp)
	require.NoError(t, err, "BN254 final wrap circuit should compile")
	t.Logf("Final wrap BN254 circuit: %d constraints", ccs.GetNbConstraints())
}

// TestToyFinalWrapEndToEnd tests the full flow with a toy wizard IOP:
// wizard IOP compile → prove → verify → BN254 wrap compile → PLONK setup → prove → verify
func TestToyFinalWrapEndToEnd(t *testing.T) {
	define := func(bui *wizard.Builder) {
		a := bui.RegisterCommit(ifaces.ColID("E2E_A"), 64)
		b := bui.RegisterCommit(ifaces.ColID("E2E_B"), 64)
		bui.Inclusion(ifaces.QueryID("E2E_Q"), []ifaces.Column{a}, []ifaces.Column{b})
	}

	comp := wizard.Compile(define, toyWizardSuite()...)
	require.NotNil(t, comp)
	t.Logf("Wizard IOP compiled: %d rounds", comp.NumRounds())

	proverFunc := func(run *wizard.ProverRuntime) {
		run.AssignColumn(ifaces.ColID("E2E_A"), smartvectors.NewConstant(field.Zero(), 64))
		run.AssignColumn(ifaces.ColID("E2E_B"), smartvectors.NewConstant(field.Zero(), 64))
	}
	wizardProof := wizard.Prove(comp, proverFunc)

	err := wizard.Verify(comp, wizardProof)
	require.NoError(t, err, "wizard proof should verify")
	t.Log("Wizard proof verified OK")

	ccs, err := MakeCS(comp)
	require.NoError(t, err, "final wrap circuit should compile")
	t.Logf("Final wrap circuit: %d constraints", ccs.GetNbConstraints())

	srsProvider := circuits.NewUnsafeSRSProvider()
	setup, err := circuits.MakeSetup(
		context.Background(),
		circuits.FinalWrapCircuitID,
		ccs,
		srsProvider,
		nil,
	)
	require.NoError(t, err, "PLONK setup should succeed")
	t.Log("PLONK setup complete")

	var publicInput frBn254.Element
	proof, err := MakeProof(&setup, comp, wizardProof, publicInput)
	require.NoError(t, err, "BN254 final wrap proof should succeed")
	require.NotNil(t, proof)

	t.Log("BN254 final wrap proof generated and verified successfully (toy)")
}

// TestFinalWrapEndToEnd tests the BN254 pipeline:
// wizard IOP (Arcane → Vortex → SelfRecursion × 2) → BN254 final wrap.
// This is the BN254-native equivalent of TestFullOldPipelineBLS in e2e_bls_test.go.
func TestFinalWrapEndToEnd(t *testing.T) {
	params := defaultRealisticParams
	srsProvider := circuits.NewUnsafeSRSProvider()

	t.Log("Compiling realistic wizard IOP (Arcane → Vortex → SelfRecursion × 2)...")
	comp := realisticWizardCompile(params, 2)
	require.NotNil(t, comp)
	t.Logf("Wizard IOP compiled: %d rounds", comp.NumRounds())

	t.Log("Proving wizard IOP...")
	wizardProof := wizard.Prove(comp, realisticWizardAssign(params))

	err := wizard.Verify(comp, wizardProof)
	require.NoError(t, err, "wizard proof should verify")
	t.Log("Wizard proof verified OK")

	t.Log("Compiling BN254 final wrap circuit...")
	ccs, err := MakeCS(comp)
	require.NoError(t, err, "final wrap circuit should compile")
	t.Logf("Final wrap circuit: %d constraints", ccs.GetNbConstraints())

	t.Log("Creating PLONK setup (BN254)...")
	setup, err := circuits.MakeSetup(
		context.Background(),
		circuits.FinalWrapCircuitID,
		ccs,
		srsProvider,
		nil,
	)
	require.NoError(t, err, "PLONK setup should succeed")

	t.Log("Generating BN254 PLONK proof...")
	var publicInput frBn254.Element
	proof, err := MakeProof(&setup, comp, wizardProof, publicInput)
	require.NoError(t, err, "BN254 final wrap proof should succeed")
	require.NotNil(t, proof)

	t.Log("BN254 final wrap proof generated and verified successfully")
}

// realisticLeafCompile compiles the base wizard IOP for use as a leaf in tree
// aggregation. Uses PremarkAsSelfRecursed so the output can be consumed by
// recursion.DefineRecursionOf at the tree level.
func realisticLeafCompile(p realisticWizardParams) *wizard.CompiledIOP {
	const (
		rsInverseRate   = 16
		nbOpenedColumns = 64
		targetColSize   = 1 << 16
		// normTotalColumns ensures CommittedRowsCount is padded to a fixed value
		// so all leaves produce Vortex opened columns of the same size.
		// With the realistic-segment params (5 perms, 50 lookups, 5 projs, 200 fibo,
		// NumCol=3, NumRow=1<<16) there are ~570 raw columns → ~850 after compilation
		// → NextPowerOfTwo(850)=1024. Use 2048 to match the inner-node normalization
		// in TreeAggregationCompilationSuite (ForceNumTotalColumns(2048)).
		normTotalColumns = 2048
	)

	return wizard.Compile(
		realisticWizardDefine(p),
		compiler.Arcane(
			compiler.WithTargetColSize(targetColSize),
			compiler.WithStitcherMinSize(1<<1),
		),
		vortex.Compile(
			rsInverseRate, false,
			vortex.WithOptionalSISHashingThreshold(512),
			vortex.ForceNumOpenedColumns(nbOpenedColumns),
			vortex.ForceNumTotalColumns(normTotalColumns),
			vortex.WithSISParams(&ringsis.StdParams),
			vortex.PremarkAsSelfRecursed(),
		),
	)
}

// TestProofSizeProgression shows real cell/column counts at each step of
// realisticWizardCompile to illustrate how self-recursion compresses the proof.
// Compile-only — no proving.
func TestProofSizeProgression(t *testing.T) {
	const (
		rsInverseRate     = 16
		nbOpenedColumns   = 64
		initTargetColSize = 1 << 16
		midTargetRowSize  = 1 << 8
		lastTargetRowSize = 1 << 10
		nbIteration       = 3
	)
	p := defaultRealisticParams
	// KoalaBear field element = 4 bytes (32-bit prime field).
	// Intermediate rounds use KoalaBear; final round adds BN254 elements (32 bytes each),
	// but proof columns are all KoalaBear-sized so 4 bytes/cell is a good approximation.
	const bytesPerCell = 4
	fmtSize := func(cells int) string {
		bytes := cells * bytesPerCell
		switch {
		case bytes >= 1<<20:
			return fmt.Sprintf("%.2f MB", float64(bytes)/float64(1<<20))
		case bytes >= 1<<10:
			return fmt.Sprintf("%.2f KB", float64(bytes)/float64(1<<10))
		default:
			return fmt.Sprintf("%d B", bytes)
		}
	}
	logStats := func(label string, comp *wizard.CompiledIOP) {
		s := logdata.GetWizardStats(comp)
		t.Logf("%-44s committed: cols=%5d  cells=%10d  | proof: cols=%4d  cells=%8d  (%s)",
			label,
			s.NumColumnsCommitted, s.NumCellsCommitted,
			s.NumColumnsProof, s.NumCellsProof,
			fmtSize(s.NumCellsProof),
		)
	}

	comp := wizard.Compile(
		realisticWizardDefine(p),
		compiler.Arcane(
			compiler.WithTargetColSize(initTargetColSize),
			compiler.WithStitcherMinSize(1<<1),
		),
	)
	logStats("after initial Arcane", comp)

	wizard.ContinueCompilation(comp,
		vortex.Compile(
			rsInverseRate, false,
			vortex.WithOptionalSISHashingThreshold(512),
			vortex.ForceNumOpenedColumns(nbOpenedColumns),
			vortex.WithSISParams(&ringsis.StdParams),
		),
	)
	logStats("after initial Vortex", comp)

	for i := 0; i < nbIteration-1; i++ {
		selfrecursion.SelfRecurse(comp)
		logStats(fmt.Sprintf("iter %d: after SelfRecurse", i), comp)

		s := logdata.GetWizardStats(comp)
		rowSize := utils.NextPowerOfTwo(utils.DivCeil(s.NumCellsCommitted, midTargetRowSize))
		t.Logf("  → computed rowSize for Arcane = %d", rowSize)

		isLastRound := i == nbIteration-1
		wizard.ContinueCompilation(comp, poseidon2.CompilePoseidon2)
		logStats(fmt.Sprintf("iter %d: after Poseidon2", i), comp)
		wizard.ContinueCompilation(comp,
			compiler.Arcane(
				compiler.WithTargetColSize(rowSize),
				compiler.WithStitcherMinSize(1<<1),
			),
		)
		logStats(fmt.Sprintf("iter %d: after Arcane", i), comp)

		wizard.ContinueCompilation(comp,
			vortex.Compile(
				rsInverseRate, isLastRound,
				vortex.ForceNumOpenedColumns(nbOpenedColumns),
				// Use SIS for all rounds so intermediate Poseidon2 linear-hash
				// queries are replaced by ring-polynomial SIS checks. This avoids
				// the O(nbOpened × colHeight / 8) Poseidon2 query blowup in
				// subsequent iterations. Only Merkle-path Poseidon2 queries remain.
				vortex.WithSISParams(&ringsis.StdParams),
				vortex.WithOptionalSISHashingThreshold(1),
			),
		)
		logStats(fmt.Sprintf("iter %d: after Vortex (lastRound=%v)", i, isLastRound), comp)
	}

	selfrecursion.SelfRecurse(comp)
	logStats(fmt.Sprintf("iter %d: after SelfRecurse", nbIteration-1), comp)

	s := logdata.GetWizardStats(comp)
	rowSize := utils.NextPowerOfTwo(utils.DivCeil(s.NumCellsCommitted, lastTargetRowSize))
	t.Logf("  → computed rowSize for Arcane = %d", rowSize)

	isLastRound := true
	wizard.ContinueCompilation(comp, poseidon2.CompilePoseidon2)
	logStats(fmt.Sprintf("iter %d: after Poseidon2", nbIteration-1), comp)
	wizard.ContinueCompilation(comp,
		compiler.Arcane(
			compiler.WithTargetColSize(rowSize),
			compiler.WithStitcherMinSize(1<<1),
		),
	)
	logStats(fmt.Sprintf("iter %d: after Arcane", nbIteration-1), comp)

	wizard.ContinueCompilation(comp,
		vortex.Compile(
			rsInverseRate, isLastRound,
			vortex.ForceNumOpenedColumns(nbOpenedColumns),
			// Use SIS for all rounds so intermediate Poseidon2 linear-hash
			// queries are replaced by ring-polynomial SIS checks. This avoids
			// the O(nbOpened × colHeight / 8) Poseidon2 query blowup in
			// subsequent iterations. Only Merkle-path Poseidon2 queries remain.
			vortex.WithOptionalSISHashingThreshold(1<<20),
		),
	)
	logStats(fmt.Sprintf("iter %d: after Vortex (lastRound=%v)", nbIteration-1, isLastRound), comp)
}

// TestConstraintsVsDepth measures BN254 constraint count at different tree depths
// WITHOUT proving — compile-only to quickly answer whether cost is O(1) with depth.
//
// Expected results:
//   - If the SelfRecurse + Arcane normalization caps proof size, constraint count
//     should be approximately constant for depth ≥ 1.
//   - If it grows, the compression isn't capping the output size.
func TestConstraintsVsDepth(t *testing.T) {
	params := defaultRealisticParams

	t.Log("Compiling leaf wizard (PremarkAsSelfRecursed)...")
	leafComp := realisticLeafCompile(params)
	t.Logf("Leaf wizard compiled: %d rounds", leafComp.NumRounds())

	for depth := 1; depth <= 5; depth++ {
		t.Run(fmt.Sprintf("depth=%d", depth), func(t *testing.T) {
			t.Logf("Compiling tree aggregation depth=%d...", depth)
			treeAgg := tree.CompileTreeAggregation(leafComp, depth)
			t.Logf("Tree depth=%d root rounds=%d", depth, treeAgg.RootCompiledIOP().NumRounds())

			ccs, err := MakeCS(treeAgg.RootCompiledIOP())
			require.NoError(t, err, "BN254 wrap circuit should compile at depth=%d", depth)
			t.Logf("depth=%d => %d BN254 constraints", depth, ccs.GetNbConstraints())
		})
	}
}

// TestProfileConstraintsByAction profiles the BN254 constraint count broken
// down by SubVerifier action type, coin round, FS init, and PI digest.
// Uses depth=2 tree aggregation.
// Run with: go test -run TestProfileConstraintsByAction -v
func TestProfileConstraintsByAction(t *testing.T) {
	params := defaultRealisticParams
	t.Log("Compiling leaf wizard...")
	leafComp := realisticLeafCompile(params)
	t.Log("Compiling depth=2 tree aggregation...")
	treeAgg := tree.CompileTreeAggregation(leafComp, 2)
	rootComp := treeAgg.RootCompiledIOP()

	// Build circuit with all profilers attached.
	circuit := Allocate(rootComp)

	// Track SubVerifier actions
	actionCounts := map[string]int{}
	circuit.WizardVerifier.ActionProfiler = func(round int, stepType string, delta int) {
		actionCounts[stepType] += delta
	}

	// Track GenerateCoinsForRound per round
	coinRoundCounts := map[int]int{}
	circuit.WizardVerifier.CoinRoundProfiler = func(round int, delta int) {
		coinRoundCounts[round] += delta
	}

	// Track FS init cost
	var fsInitCount int
	circuit.WizardVerifier.FSInitProfiler = func(delta int) {
		fsInitCount = delta
	}

	// Track computePIDigest cost
	var piDigestCount int
	circuit.PIDigestProfiler = func(delta int) {
		piDigestCount = delta
	}

	t.Log("Compiling BN254 circuit with profilers...")
	ccs, err := frontend.Compile(ecc.BN254.ScalarField(), scs.NewBuilder, circuit)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}
	total := ccs.GetNbConstraints()

	// --- SubVerifier action breakdown ---
	type entry struct {
		name  string
		count int
	}
	var actionRows []entry
	actionTotal := 0
	for k, v := range actionCounts {
		actionRows = append(actionRows, entry{k, v})
		actionTotal += v
	}
	sort.Slice(actionRows, func(i, j int) bool { return actionRows[i].count > actionRows[j].count })

	// --- Coin round breakdown ---
	coinTotal := 0
	var coinRounds []int
	for r := range coinRoundCounts {
		coinRounds = append(coinRounds, r)
		coinTotal += coinRoundCounts[r]
	}
	sort.Ints(coinRounds)

	// --- Summary ---
	knownTotal := actionTotal + coinTotal + fsInitCount + piDigestCount
	unknownTotal := total - knownTotal

	t.Logf("Total BN254 constraints: %d", total)
	t.Log("═══════════════════════════════════════════════════════════════════")
	t.Logf("  FS init (FiatShamirSetup absorb):  %12d  (%5.1f%%)", fsInitCount, 100*float64(fsInitCount)/float64(total))
	t.Logf("  GenerateCoinsForRound (all rounds): %12d  (%5.1f%%)", coinTotal, 100*float64(coinTotal)/float64(total))
	for _, r := range coinRounds {
		c := coinRoundCounts[r]
		t.Logf("    round %-3d: %12d  (%5.1f%%)", r, c, 100*float64(c)/float64(total))
	}
	t.Logf("  SubVerifier actions total:          %12d  (%5.1f%%)", actionTotal, 100*float64(actionTotal)/float64(total))
	t.Logf("  computePIDigest:                    %12d  (%5.1f%%)", piDigestCount, 100*float64(piDigestCount)/float64(total))
	t.Logf("  Untracked (AssertIsEqual etc.):     %12d  (%5.1f%%)", unknownTotal, 100*float64(unknownTotal)/float64(total))
	t.Log("───────────────────────────────────────────────────────────────────")
	t.Log("SubVerifier action breakdown:")
	t.Logf("  %-60s %12s  %6s", "Action type", "Constraints", "%total")
	t.Log("  " + fmt.Sprintf("%s", "─────────────────────────────────────────────────────────────────"))
	for _, r := range actionRows {
		t.Logf("  %-60s %12d  %5.1f%%", r.name, r.count, 100*float64(r.count)/float64(total))
	}
}

// TestAggregateFinalWrapEndToEnd tests the full tree-aggregation pipeline for N=2:
//
//	2 leaf wizard proofs (PremarkAsSelfRecursed)
//	  → tree.CompileTreeAggregation(depth=1): DefineRecursionOf(MaxNumProof=2) on KoalaBear
//	    → 1 root proof (IsLastRound=true)
//	      → BN254 PLONK final wrap (L1)
func TestAggregateFinalWrapEndToEnd(t *testing.T) {
	params := defaultRealisticParams
	srsProvider := circuits.NewUnsafeSRSProvider()

	// Step 1: compile the leaf wizard
	t.Log("Compiling leaf wizard (PremarkAsSelfRecursed)...")
	leafComp := realisticLeafCompile(params)
	t.Logf("Leaf wizard compiled: %d rounds", leafComp.NumRounds())

	// Step 2: compile tree aggregation (depth=1 for N=2, single final level)
	t.Log("Compiling tree aggregation (depth=1 for N=2)...")
	treeAgg := tree.CompileTreeAggregation(leafComp, 1)
	t.Logf("Tree compiled: depth=%d, root rounds=%d", treeAgg.Depth(), treeAgg.RootCompiledIOP().NumRounds())

	// Step 3: generate 2 leaf witnesses by running the prover to the vortex
	// query round and extracting the recursion witness.
	t.Log("Generating 2 leaf witnesses...")
	stoppingRound := recursion.VortexQueryRound(leafComp) + 1
	run1 := wizard.RunProverUntilRound(leafComp, realisticWizardAssign(params), stoppingRound)
	wit1 := recursion.ExtractWitness(run1)
	run2 := wizard.RunProverUntilRound(leafComp, realisticWizardAssign(params), stoppingRound)
	wit2 := recursion.ExtractWitness(run2)

	// Step 4: prove the tree — produces a single root proof
	t.Log("Proving tree aggregation (N=2)...")
	rootProof, err := treeAgg.ProveTree([]recursion.Witness{wit1, wit2})
	require.NoError(t, err, "tree proving should succeed")

	err = wizard.Verify(treeAgg.RootCompiledIOP(), rootProof)
	require.NoError(t, err, "root proof should verify")
	t.Log("Root proof verified OK")

	// Step 5: compile BN254 final wrap circuit around the root compiled IOP
	t.Log("Compiling BN254 final wrap circuit...")
	ccs, err := MakeCS(treeAgg.RootCompiledIOP())
	require.NoError(t, err, "final wrap circuit should compile")
	t.Logf("Final wrap circuit: %d constraints", ccs.GetNbConstraints())

	// Step 6: PLONK setup
	t.Log("Creating PLONK setup (BN254)...")
	setup, err := circuits.MakeSetup(
		context.Background(),
		circuits.FinalWrapCircuitID,
		ccs,
		srsProvider,
		nil,
	)
	require.NoError(t, err, "PLONK setup should succeed")

	// Step 7: prove and verify
	t.Log("Generating BN254 PLONK proof...")
	var publicInput frBn254.Element
	proof, err := MakeProof(&setup, treeAgg.RootCompiledIOP(), rootProof, publicInput)
	require.NoError(t, err, "BN254 aggregate final wrap proof should succeed")
	require.NotNil(t, proof)

	t.Log("BN254 aggregate final wrap proof (N=2) generated and verified successfully")
}

// TestAggregateFinalWrapEndToEndDepth2 tests the full tree-aggregation pipeline for N=4:
//
//	4 leaf wizard proofs
//	  → tree.CompileTreeAggregation(depth=2): 2 aggregation levels
//	    → 1 root proof (IsLastRound=true)
//	      → BN254 PLONK final wrap (L1)
//
// This validates that depth=2 still produces a valid proof and measures
// whether constraint count is O(1) vs depth (compared to depth=1 above).
func TestAggregateFinalWrapEndToEndDepth2(t *testing.T) {
	params := defaultRealisticParams
	srsProvider := circuits.NewUnsafeSRSProvider()

	// Step 1: compile the leaf wizard
	t.Log("Compiling leaf wizard (PremarkAsSelfRecursed)...")
	leafComp := realisticLeafCompile(params)
	t.Logf("Leaf wizard compiled: %d rounds", leafComp.NumRounds())

	// Step 2: compile tree aggregation depth=2 (2 levels, verifies 4 leaf proofs)
	t.Log("Compiling tree aggregation (depth=2 for N=4)...")
	treeAgg := tree.CompileTreeAggregation(leafComp, 2)
	t.Logf("Tree compiled: depth=%d, root rounds=%d", treeAgg.Depth(), treeAgg.RootCompiledIOP().NumRounds())

	// Step 3: compile final wrap circuit and log constraint count
	t.Log("Compiling BN254 final wrap circuit...")
	ccs, err := MakeCS(treeAgg.RootCompiledIOP())
	require.NoError(t, err, "final wrap circuit should compile at depth=2")
	t.Logf("Final wrap circuit (depth=2): %d constraints", ccs.GetNbConstraints())

	// Step 4: generate 4 leaf witnesses
	t.Log("Generating 4 leaf witnesses...")
	stoppingRound := recursion.VortexQueryRound(leafComp) + 1
	wits := make([]recursion.Witness, 4)
	for i := range wits {
		run := wizard.RunProverUntilRound(leafComp, realisticWizardAssign(params), stoppingRound)
		wits[i] = recursion.ExtractWitness(run)
	}

	// Step 5: prove the tree
	t.Log("Proving tree aggregation (N=4, depth=2)...")
	rootProof, err := treeAgg.ProveTree(wits)
	require.NoError(t, err, "tree proving should succeed at depth=2")

	err = wizard.Verify(treeAgg.RootCompiledIOP(), rootProof)
	require.NoError(t, err, "root proof should verify at depth=2")
	t.Log("Root proof verified OK (depth=2)")

	// Step 6: PLONK setup + prove + verify
	t.Log("Creating PLONK setup (BN254, depth=2)...")
	setup, err := circuits.MakeSetup(
		context.Background(),
		circuits.FinalWrapCircuitID,
		ccs,
		srsProvider,
		nil,
	)
	require.NoError(t, err, "PLONK setup should succeed at depth=2")

	t.Log("Generating BN254 PLONK proof (depth=2)...")
	var publicInput frBn254.Element
	proof, err := MakeProof(&setup, treeAgg.RootCompiledIOP(), rootProof, publicInput)
	require.NoError(t, err, "BN254 aggregate final wrap proof should succeed at depth=2")
	require.NotNil(t, proof)

	t.Log("BN254 aggregate final wrap proof (N=4, depth=2) generated and verified successfully")
}

// TestProductionProofSizeProgression mirrors fullSecondCompilationSuite from
// zkevm/full.go step-by-step, logging proof sizes at each stage.
// realisticLeafCompile produces the same PremarkAsSelfRecursed output as
// fullInitialCompilationSuite, so this test begins where fullSecondCompilationSuite starts.
func TestProductionProofSizeProgression(t *testing.T) {
	p := defaultRealisticParams
	const bytesPerCell = 4
	fmtSize := func(cells int) string {
		bytes := cells * bytesPerCell
		switch {
		case bytes >= 1<<20:
			return fmt.Sprintf("%.2f MB", float64(bytes)/float64(1<<20))
		case bytes >= 1<<10:
			return fmt.Sprintf("%.2f KB", float64(bytes)/float64(1<<10))
		default:
			return fmt.Sprintf("%d B", bytes)
		}
	}
	logStats := func(label string, comp *wizard.CompiledIOP) {
		s := logdata.GetWizardStats(comp)
		t.Logf("%-58s committed: cols=%5d  cells=%10d  | proof: cols=%4d  cells=%8d  (%s)",
			label,
			s.NumColumnsCommitted, s.NumCellsCommitted,
			s.NumColumnsProof, s.NumCellsProof,
			fmtSize(s.NumCellsProof),
		)
	}

	// zkevm/full.go sisInstance = ringsis.Params{LogTwoBound: 16, LogTwoDegree: 6}
	sisInst := ringsis.Params{LogTwoBound: 16, LogTwoDegree: 6}

	// Equivalent to the output of fullInitialCompilationSuite (ends with PremarkAsSelfRecursed).
	comp := realisticLeafCompile(p)
	logStats("leaf (≡ fullInitialCompilationSuite output)", comp)

	// fullSecondCompilationSuite step 1:
	// CleanUp → Poseidon2 → Arcane(1<<22) → Vortex(rsRate=2, nbOpened=256, SIS)
	wizard.ContinueCompilation(comp, cleanup.CleanUp, poseidon2.CompilePoseidon2)
	logStats("fullSecond step1: after CleanUp+Poseidon2", comp)
	wizard.ContinueCompilation(comp, compiler.Arcane(
		compiler.WithTargetColSize(1<<22),
		compiler.WithStitcherMinSize(16),
	))
	logStats("fullSecond step1: after Arcane(1<<22)", comp)
	wizard.ContinueCompilation(comp, vortex.Compile(2, false,
		vortex.ForceNumOpenedColumns(256),
		vortex.WithSISParams(&sisInst),
	))
	logStats("fullSecond step1: after Vortex(rsRate=2, nbOpened=256, SIS)", comp)

	// fullSecondCompilationSuite step 2:
	// SelfRecurse → CleanUp → Poseidon2 → Arcane(1<<17) → Vortex(rsRate=16, nbOpened=64, SIS)
	selfrecursion.SelfRecurse(comp)
	logStats("fullSecond selfrecurse1: after SelfRecurse", comp)
	wizard.ContinueCompilation(comp, cleanup.CleanUp, poseidon2.CompilePoseidon2)
	logStats("fullSecond step2: after CleanUp+Poseidon2", comp)
	wizard.ContinueCompilation(comp, compiler.Arcane(
		compiler.WithTargetColSize(1<<17),
		compiler.WithStitcherMinSize(16),
	))
	logStats("fullSecond step2: after Arcane(1<<17)", comp)
	wizard.ContinueCompilation(comp, vortex.Compile(16, false,
		vortex.ForceNumOpenedColumns(64),
		vortex.WithSISParams(&sisInst),
	))
	logStats("fullSecond step2: after Vortex(rsRate=16, nbOpened=64, SIS)", comp)

	// fullSecondCompilationSuite step 3:
	// SelfRecurse → CleanUp → Poseidon2 → Arcane(1<<12) → Vortex(PremarkAsSelfRecursed)
	selfrecursion.SelfRecurse(comp)
	logStats("fullSecond selfrecurse2: after SelfRecurse", comp)
	wizard.ContinueCompilation(comp, cleanup.CleanUp, poseidon2.CompilePoseidon2)
	logStats("fullSecond step3: after CleanUp+Poseidon2", comp)
	wizard.ContinueCompilation(comp, compiler.Arcane(
		compiler.WithTargetColSize(1<<12),
		compiler.WithStitcherMinSize(16),
	))
	logStats("fullSecond step3: after Arcane(1<<12)", comp)
	wizard.ContinueCompilation(comp, vortex.Compile(16, false,
		vortex.ForceNumOpenedColumns(64),
		vortex.WithOptionalSISHashingThreshold(1<<20),
		vortex.PremarkAsSelfRecursed(),
	))
	logStats("fullSecond step3: after Vortex(PremarkAsSelfRecursed) [→ tree-agg leaf]", comp)
}
