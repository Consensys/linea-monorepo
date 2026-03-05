package finalwrap

import (
	"context"
	"fmt"
	"testing"

	frBn254 "github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/linea-monorepo/prover/circuits"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2_koalabear"
	"github.com/consensys/linea-monorepo/prover/crypto/ringsis"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler"
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

var defaultRealisticParams = realisticWizardParams{
	NumPermutations: 1,
	NumLookups:      5,
	NumProjections:  1,
	NumFibo:         10,
	NumCol:          2,
	NumRow:          1 << 14,
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
// Arcane → Vortex → SelfRecursion → Poseidon2 → Arcane → Vortex (IsLastRound=true).
// This matches realisticBLSWizardCompile in e2e_bls_test.go for fair comparison.
func realisticWizardCompile(p realisticWizardParams) *wizard.CompiledIOP {
	const (
		rsInverseRate     = 16
		nbOpenedColumns   = 64
		initTargetColSize = 1 << 16
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

	// IsLastRound=true → BN254-native FS + Merkle hashing
	wizard.ContinueCompilation(comp,
		vortex.Compile(
			rsInverseRate, true,
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
// wizard IOP (Arcane → Vortex → SelfRecursion × 1) → BN254 final wrap.
// This is the BN254-native equivalent of TestFullOldPipelineBLS in e2e_bls_test.go.
func TestFinalWrapEndToEnd(t *testing.T) {
	params := defaultRealisticParams
	srsProvider := circuits.NewUnsafeSRSProvider()

	t.Log("Compiling realistic wizard IOP (Arcane → Vortex → SelfRecursion × 1)...")
	comp := realisticWizardCompile(params)
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
		rsInverseRate    = 16
		nbOpenedColumns  = 64
		targetColSize    = 1 << 16
		// normTotalColumns pins the Merkle tree depth to log2(64)=6 regardless
		// of how many polynomial columns the wizard circuit happens to produce.
		// Without this, two leaves from circuits of different sizes would have
		// different Merkle depths, making the tree-aggregation verification
		// circuit non-uniform and breaking the "one fixed circuit per level"
		// property.  Any N ≥ actual CommittedRowsCount works; 64 matches the
		// current count (52 → NextPowerOfTwo → 64) and is the minimum safe value
		// for this test circuit.  Production leaf suites need an analogous
		// ForceNumTotalColumns sized to their actual column budget.
		normTotalColumns = 64
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

	for depth := 1; depth <= 3; depth++ {
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
