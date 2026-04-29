package distributed_test

import (
	"fmt"
	"math/bits"
	"testing"

	"github.com/consensys/linea-monorepo/prover/backend/files"
	multisethashing "github.com/consensys/linea-monorepo/prover/crypto/multisethashing_koalabear"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/protocol/accessors"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/multilineareval"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/multilinvortex"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/serde"
	"github.com/consensys/linea-monorepo/prover/protocol/sumcheck"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
)

const (
	NbRow = 8
)

// testCompilationParams defines compilation parameters for testing segment compilation
// Note: ColumnProfileMPTS is left nil to avoid profile size constraints during testing
var testCompilationParams = distributed.CompilationParams{
	FixedNbRowPlonkCircuit:       1 << 25,
	FixedNbRowExternalHasher:     1 << 19, // Increased from 1<<22 to handle hash claims
	FixedNbPublicInput:           1 << 10,
	InitialCompilerSize:          1 << 18,
	InitialCompilerSizeConglo:    1 << 18,
	ColumnProfileMPTS:            []int{264, 1400, 256, 24, 12, 28, 8, 8},
	ColumnProfileMPTSPrecomputed: 45,
	FullDebugMode:                false,
}

// DistributedTestCase is an interface for test cases that can be run through
// the distributed wizard.
type DistributedTestCase interface {
	// Name returns the name of the test case
	Name() string
	// Define defines the structure of the wizard (columns, constraints, queries)
	Define(comp *wizard.CompiledIOP)
	// Assign assigns values to the columns at runtime
	Assign(run *wizard.ProverRuntime)
	// Advices returns a list of advices for the module discovery.
	Advices() []*distributed.ModuleDiscoveryAdvice
}

// TestDistributedWizard runs the distributed wizard test over multiple test cases.
// Each test case defines a different set of queries (inclusion, projection, permutation).
func TestDistributedWizard(t *testing.T) {

	testCases := []DistributedTestCase{
		&LookupTestCase{numRow: 1 << NbRow},
		&ProjectionTestCase{numRow: 1 << NbRow},
		&PermutationTestCase{numRow: 1 << NbRow},
		&FibExtTestCase{numRow: 1 << NbRow},
		&MultilinVortexTestCase{numRow: 1 << NbRow},
	}

	for _, tc := range testCases {
		t.Run(tc.Name(), func(t *testing.T) {
			runDistributedWizardTest(t, tc, false)
		})
	}
}

// TestCompileOneSegment tests the compilation of a single segment.
func TestCompileOneSegment(t *testing.T) {

	t.SkipNow()

	var (
		tc      = &PermutationTestCase{numRow: 1 << NbRow}
		defFunc = func(build *wizard.Builder) { tc.Define(build.CompiledIOP) }
		wiop    = wizard.Compile(defFunc)
		disc    = &distributed.StandardModuleDiscoverer{
			TargetWeight: 1 << NbRow,
			Advices:      tc.Advices(),
		}
		distWizard = distributed.DistributeWizard(wiop, disc)
		compiled   = distributed.CompileSegment(distWizard.GLs[0], testCompilationParams)
	)

	profileTree, err := serde.Profile(compiled)
	if err != nil {
		t.Fatal(err)
	}

	profileTree.PruneTree(1 << 20)
	serde.WriteProfileTo(profileTree, files.MustOverwrite("./profiling/comp-profile.txt"))
}

// runDistributedWizardTest runs a single distributed wizard test case.
func runDistributedWizardTest(t *testing.T, tc DistributedTestCase, segmentCompilation bool) {

	var (
		defFunc = func(build *wizard.Builder) { tc.Define(build.CompiledIOP) }
		wiop    = wizard.Compile(defFunc)
		disc    = &distributed.StandardModuleDiscoverer{
			TargetWeight: 1 << NbRow,
			Advices:      tc.Advices(),
		}
		distWizard *distributed.DistributedWizard
	)

	// Always create the DistributedWizard first so PostDistribute can run on
	// the bare Bootstrapper before any further compilation.
	distWizard = distributed.DistributeWizard(wiop, disc)

	// Allow test cases to compile Bootstrapper-level protocols (e.g. ML Vortex)
	// that must run after distribution but before proving.
	type postDistributer interface {
		PostDistribute(dw *distributed.DistributedWizard)
	}
	if pd, ok := tc.(postDistributer); ok {
		pd.PostDistribute(distWizard)
	}

	// This tests the compilation of the compiled-IOP
	if segmentCompilation {
		// compile with vortex/ self-recursion/recursion
		distWizard.CompileSegments(testCompilationParams).
			Conglomerate(testCompilationParams)
	}

	// This compilation step is needed for sanity-checking the bootstrapper
	dummy.Compile(distWizard.Bootstrapper)

	// This applies the dummy.Compiler to all parts of the distributed wizard.
	for i := range distWizard.GLs {
		dummy.CompileAtProverLvl()(distWizard.GLs[i].Wiop)
		// Add dummy LPP merkle root public inputs (normally added by Vortex compiler)
		for j := 0; j < 8; j++ {
			name := fmt.Sprintf("LPP_COLUMNS_MERKLE_ROOTS_0_%d", j)
			distWizard.GLs[i].Wiop.InsertPublicInput(name, accessors.NewConstant(field.Zero()))
		}
	}

	for i := range distWizard.LPPs {
		dummy.CompileAtProverLvl()(distWizard.LPPs[i].Wiop)
		// Add dummy LPP merkle root public inputs (normally added by Vortex compiler)
		for j := 0; j < 8; j++ {
			name := fmt.Sprintf("LPP_COLUMNS_MERKLE_ROOTS_0_%d", j)
			distWizard.LPPs[i].Wiop.InsertPublicInput(name, accessors.NewConstant(field.Zero()))
		}
	}

	var (
		runtimeBoot = wizard.RunProver(distWizard.Bootstrapper, tc.Assign, false)
		proof       = runtimeBoot.ExtractProof()
		verBootErr  = wizard.Verify(distWizard.Bootstrapper, proof)
	)

	if verBootErr != nil {
		t.Fatalf("Bootstrapper failed because: %v", verBootErr)
	}

	var (
		allGrandProduct     = fext.One()
		allLogDerivativeSum = fext.Element{}
		allHornerSum        = fext.Element{}
		generalMSet         = multisethashing.MSetHash{}
	)

	witnessGLs, witnessLPPs := distributed.SegmentRuntime(
		runtimeBoot,
		distWizard.Disc,
		distWizard.BlueprintGLs,
		distWizard.BlueprintLPPs,
		field.Octuplet{},
	)

	for i := range witnessGLs {

		var (
			witnessGL = witnessGLs[i]
			moduleGL  *distributed.ModuleGL
		)

		t.Logf("segment(total)=%v module=%v segment.index=%v", i, witnessGL.ModuleName, witnessGL.ModuleIndex)

		for k := range distWizard.ModuleNames {
			if distWizard.ModuleNames[k] != witnessGLs[i].ModuleName {
				continue
			}

			moduleGL = distWizard.GLs[k]
			break
		}

		if moduleGL == nil {
			t.Fatalf("module does not exists")
		}

		var (
			proverRunGL         = wizard.RunProver(moduleGL.Wiop, moduleGL.GetMainProverStep(witnessGLs[i]), false)
			proofGL             = proverRunGL.ExtractProof()
			verRun, verGLErr    = wizard.VerifyWithRuntime(moduleGL.Wiop, proofGL, false)
			generalMSetFromGLFr = distributed.GetPublicInputList(verRun, distributed.GeneralMultiSetPublicInputBase, multisethashing.MSetHashSize)
			generalMSetFromGL   = multisethashing.MSetHash(generalMSetFromGLFr)
		)

		if verGLErr != nil {
			t.Errorf("verifier failed for segment %v, reason=%v", i, verGLErr)
		}

		generalMSet.Add(generalMSetFromGL)
	}

	for i := range witnessLPPs {

		var (
			witnessLPP  = witnessLPPs[i]
			moduleIndex = witnessLPPs[i].ModuleIndex
			moduleLPP   = distWizard.LPPs[moduleIndex]
		)

		// The test uses stubbed values for the initial Fiat-Shamir state for
		// replicability reasons. This is important to help debugging.
		witnessLPP.InitialFiatShamirState = field.NewOctupletFromStrings(
			[8]string{
				"123456789",
				"987654321",
				"111111111",
				"222222222",
				"333333333",
				"444444444",
				"555555555",
				"666666666",
			})

		t.Logf("segment(total)=%v module=%v segment.index=%v", i, witnessLPP.ModuleName, witnessLPP.ModuleIndex)

		var (
			proverRunLPP         = wizard.RunProver(moduleLPP.Wiop, moduleLPP.GetMainProverStep(witnessLPP), false)
			proofLPP             = proverRunLPP.ExtractProof()
			verRun, verLPPErr    = wizard.VerifyWithRuntime(moduleLPP.Wiop, proofLPP, false)
			generalMSetFromLPPFr = distributed.GetPublicInputList(verRun, distributed.GeneralMultiSetPublicInputBase, multisethashing.MSetHashSize)
			generalMSetFromLPP   = multisethashing.MSetHash(generalMSetFromLPPFr)
		)

		if verLPPErr != nil {
			t.Errorf("verifier failed for segment %v, reason=%v", i, verLPPErr)
		}

		generalMSet.Add(generalMSetFromLPP)

		var (
			logDerivativeSum = verRun.GetPublicInput(distributed.LogDerivativeSumPublicInput).Ext
			grandProduct     = verRun.GetPublicInput(distributed.GrandProductPublicInput).Ext
			hornerSum        = verRun.GetPublicInput(distributed.HornerPublicInput).Ext
		)

		t.Logf("log-derivative-sum=%v grand-product=%v horner-sum=%v", logDerivativeSum.String(), grandProduct.String(), hornerSum.String())

		allGrandProduct.Mul(&allGrandProduct, &grandProduct)
		allHornerSum.Add(&allHornerSum, &hornerSum)
		allLogDerivativeSum.Add(&allLogDerivativeSum, &logDerivativeSum)
	}

	if !allGrandProduct.IsOne() {
		t.Errorf("grand-product does not cancel")
	}

	if !allHornerSum.IsZero() {
		t.Errorf("horner does not cancel")
	}

	if !allLogDerivativeSum.IsZero() {
		t.Errorf("log-derivative-sum does not cancel. Has %v", allLogDerivativeSum.String())
	}

	if !generalMSet.IsEmpty() {
		t.Errorf("generalMSet does not cancel: ")
	}
}

// LookupTestCase tests the basic distributed wizard with inclusion query (lookup).
// The testcase generates 2 triplets of columns a, b, c such that a + b = c
// and the two modules are joined by a lookup.
type LookupTestCase struct {
	numRow int
	wiop   *wizard.CompiledIOP
}

func (d *LookupTestCase) Name() string {
	return "Lookup"
}

// Define defines the structure of the distributed wizard. The structure is
// composed of 2 modules that are connected by a lookup. The two modules are
// identical and are defined as a + b = c.
func (d *LookupTestCase) Define(comp *wizard.CompiledIOP) {

	d.wiop = comp

	// Define the first module
	a0 := comp.InsertCommit(0, "a0", d.numRow, true)
	b0 := comp.InsertCommit(0, "b0", d.numRow, true)
	c0 := comp.InsertCommit(0, "c0", d.numRow, true)

	// Importantly, the second module must be slightly different than the first
	// one because else it will create a wierd edge case in the conglomeration:
	// as we would have two GL modules with the same verifying key and we would
	// not be able to infer a module from a VK.
	//
	// We differentiate the modules by adding a duplicate constraints for GL0
	a1 := comp.InsertCommit(0, "a1", d.numRow, true)
	b1 := comp.InsertCommit(0, "b1", d.numRow, true)
	c1 := comp.InsertCommit(0, "c1", d.numRow, true)

	comp.InsertGlobal(0, "global-0", symbolic.Sub(c0, b0, a0))
	comp.InsertGlobal(0, "global-duplicate", symbolic.Sub(c0, b0, a0))
	comp.InsertGlobal(0, "global-1", symbolic.Sub(c1, b1, a1))

	comp.InsertInclusion(0, "inclusion-0", []ifaces.Column{c0, b0, a0}, []ifaces.Column{c1, b1, a1})
}

// Assign sets up the column assignments for the LookupTestCase.
func (d *LookupTestCase) Assign(run *wizard.ProverRuntime) {
	run.AssignColumn("a0", smartvectors.RightZeroPadded(vector.Repeat(field.NewElement(1), d.numRow-2), d.numRow))
	run.AssignColumn("b0", smartvectors.RightZeroPadded(vector.Repeat(field.NewElement(2), d.numRow-2), d.numRow))
	run.AssignColumn("c0", smartvectors.RightZeroPadded(vector.Repeat(field.NewElement(3), d.numRow-2), d.numRow))
	run.AssignColumn("a1", smartvectors.RightZeroPadded(vector.Repeat(field.NewElement(1), d.numRow-2), d.numRow))
	run.AssignColumn("b1", smartvectors.RightZeroPadded(vector.Repeat(field.NewElement(2), d.numRow-2), d.numRow))
	run.AssignColumn("c1", smartvectors.RightZeroPadded(vector.Repeat(field.NewElement(3), d.numRow-2), d.numRow))
}

// Advices returns the advices for the LookupTestCase.
func (d *LookupTestCase) Advices() []*distributed.ModuleDiscoveryAdvice {
	return []*distributed.ModuleDiscoveryAdvice{
		distributed.SameSizeAdvice("module-0", d.wiop.Columns.GetHandle("a1")),
		distributed.SameSizeAdvice("module-0", d.wiop.Columns.GetHandle("a0")),
	}
}

// ProjectionTestCase includes a projection query in addition to
// the global and inclusion constraints. Projection queries are compiled
// into Horner evaluations.
type ProjectionTestCase struct {
	numRow int
	wiop   *wizard.CompiledIOP
}

func (d *ProjectionTestCase) Name() string {
	return "Projection"
}

// Define defines the structure with projection query. It has:
// - Global constraints a + b = c for both modules
// - An inclusion query between modules
// - A projection query between filtered columns
func (d *ProjectionTestCase) Define(comp *wizard.CompiledIOP) {

	d.wiop = comp

	// Define the first module
	a0 := comp.InsertCommit(0, "a0", d.numRow/2, true)
	b0 := comp.InsertCommit(0, "b0", d.numRow/2, true)
	c0 := comp.InsertCommit(0, "c0", d.numRow/2, true)

	// Second module with slight differences
	a1 := comp.InsertCommit(0, "a1", d.numRow, true)
	b1 := comp.InsertCommit(0, "b1", d.numRow, true)
	c1 := comp.InsertCommit(0, "c1", d.numRow, true)

	// Projection columns: projA with filterA projects to projB with filterB
	filterA := comp.InsertCommit(0, "filterA", d.numRow/2, true)
	filterB := comp.InsertCommit(0, "filterB", d.numRow, true)

	comp.InsertGlobal(0, "global-0", symbolic.Sub(c0, b0, a0))
	comp.InsertGlobal(0, "global-1", symbolic.Sub(c1, b1, a1))

	// Projection query: filtered values from projA should match filtered values from projB
	comp.InsertProjection(
		"projection-0",
		query.ProjectionInput{
			ColumnA: []ifaces.Column{a0, b0, c0},
			ColumnB: []ifaces.Column{a1, b1, c1},
			FilterA: filterA,
			FilterB: filterB,
		},
	)
}

// Assign assigns the columns for the projection test case.
// The projection query maps filtered rows from (a0, b0, c0) to (a1, b1, c1).
// Each column has sequential values 1, 2, 3, ...
func (d *ProjectionTestCase) Assign(run *wizard.ProverRuntime) {
	sizeSmall := d.numRow / 2
	sizeLarge := d.numRow

	// Create sequential values for the small table (a0, b0, c0)
	// a0: 1, 2, 3, ..., sizeSmall
	// b0: 1, 2, 3, ..., sizeSmall
	// c0: 2, 4, 6, ..., 2*sizeSmall (since c = a + b for the global constraint)
	a0Vals := make([]field.Element, sizeSmall)
	b0Vals := make([]field.Element, sizeSmall)
	c0Vals := make([]field.Element, sizeSmall)
	for i := 0; i < sizeSmall; i++ {
		a0Vals[i] = field.NewElement(uint64(i + 1))
		b0Vals[i] = field.NewElement(uint64(i + 1))
		c0Vals[i] = field.NewElement(uint64(2 * (i + 1))) // c = a + b
	}

	// Create values for the large table (a1, b1, c1)
	// The first sizeSmall rows contain the same values as (a0, b0, c0)
	// The rest are padded with zeros
	a1Vals := make([]field.Element, sizeLarge)
	b1Vals := make([]field.Element, sizeLarge)
	c1Vals := make([]field.Element, sizeLarge)
	for i := 0; i < sizeSmall; i++ {
		a1Vals[i] = field.NewElement(uint64(i + 1))
		b1Vals[i] = field.NewElement(uint64(i + 1))
		c1Vals[i] = field.NewElement(uint64(2 * (i + 1))) // c = a + b
	}
	// Remaining positions are zero (default)

	run.AssignColumn("a0", smartvectors.NewRegular(a0Vals))
	run.AssignColumn("b0", smartvectors.NewRegular(b0Vals))
	run.AssignColumn("c0", smartvectors.NewRegular(c0Vals))
	run.AssignColumn("a1", smartvectors.NewRegular(a1Vals))
	run.AssignColumn("b1", smartvectors.NewRegular(b1Vals))
	run.AssignColumn("c1", smartvectors.NewRegular(c1Vals))

	// Assign filters:
	// filterA: all 1s (select all rows from the small table)
	// filterB: 1 for the first sizeSmall rows, 0 for the rest
	filterAVals := make([]field.Element, sizeSmall)
	filterBVals := make([]field.Element, sizeLarge)

	for i := 0; i < sizeSmall; i++ {
		filterAVals[i] = field.One()
		filterBVals[i] = field.One()
	}
	// filterBVals[sizeSmall:] remains zero

	run.AssignColumn("filterA", smartvectors.NewRegular(filterAVals))
	run.AssignColumn("filterB", smartvectors.NewRegular(filterBVals))
}

// Advices returns a list of advices for the module discovery
func (d *ProjectionTestCase) Advices() []*distributed.ModuleDiscoveryAdvice {
	return []*distributed.ModuleDiscoveryAdvice{
		distributed.SameSizeAdvice("module-0", d.wiop.Columns.GetHandle("a0")),
		distributed.SameSizeAdvice("module-0", d.wiop.Columns.GetHandle("a1")),
	}
}

// PermutationTestCase includes a permutation query in addition to
// the global constraints. Permutation queries are compiled into
// grand product arguments.
type PermutationTestCase struct {
	numRow int
	wiop   *wizard.CompiledIOP
}

func (d *PermutationTestCase) Name() string {
	return "Permutation"
}

// Define defines the structure with permutation query. It has:
// - Global constraints a + b = c for both modules
// - A permutation query asserting that (a0, b0, c0) and (a1, b1, c1) contain the same rows
func (d *PermutationTestCase) Define(comp *wizard.CompiledIOP) {

	d.wiop = comp

	// Define the first module
	a0 := comp.InsertCommit(0, "a0", d.numRow, true)
	b0 := comp.InsertCommit(0, "b0", d.numRow, true)
	c0 := comp.InsertCommit(0, "c0", d.numRow, true)

	// Second module
	a1 := comp.InsertCommit(0, "a1", d.numRow, true)
	b1 := comp.InsertCommit(0, "b1", d.numRow, true)
	c1 := comp.InsertCommit(0, "c1", d.numRow, true)

	comp.InsertGlobal(0, "global-0", symbolic.Sub(c0, b0, a0))
	comp.InsertGlobal(0, "global-1", symbolic.Sub(c1, b1, a1))

	// Permutation query: (a0, b0, c0) and (a1, b1, c1) should contain the same rows (possibly reordered)
	comp.InsertPermutation(0, "permutation-0", []ifaces.Column{a0, b0, c0}, []ifaces.Column{a1, b1, c1})
}

// Assign assigns the columns for the permutation test case.
// Each column has sequential values 1, 2, 3, ... with c = a + b.
// The second table (a1, b1, c1) is a permutation (reversed) of the first (a0, b0, c0).
func (d *PermutationTestCase) Assign(run *wizard.ProverRuntime) {
	// Create sequential values for the first table (a0, b0, c0)
	// a0: 1, 2, 3, ..., numRow
	// b0: 1, 2, 3, ..., numRow
	// c0: 2, 4, 6, ..., 2*numRow (since c = a + b)
	a0Vals := make([]field.Element, d.numRow)
	b0Vals := make([]field.Element, d.numRow)
	c0Vals := make([]field.Element, d.numRow)

	for i := 0; i < d.numRow; i++ {
		a0Vals[i] = field.NewElement(uint64(i + 1))
		b0Vals[i] = field.NewElement(uint64(i + 1))
		c0Vals[i] = field.NewElement(uint64(2 * (i + 1))) // c = a + b
	}

	// Create the second table (a1, b1, c1) as a permutation (reversed) of the first
	a1Vals := make([]field.Element, d.numRow)
	b1Vals := make([]field.Element, d.numRow)
	c1Vals := make([]field.Element, d.numRow)

	for i := 0; i < d.numRow; i++ {
		// Reverse the order - this is a general permutation that crosses segments
		a1Vals[d.numRow-1-i] = field.NewElement(uint64(i + 1))
		b1Vals[d.numRow-1-i] = field.NewElement(uint64(i + 1))
		c1Vals[d.numRow-1-i] = field.NewElement(uint64(2 * (i + 1)))
	}

	run.AssignColumn("a0", smartvectors.NewRegular(a0Vals))
	run.AssignColumn("b0", smartvectors.NewRegular(b0Vals))
	run.AssignColumn("c0", smartvectors.NewRegular(c0Vals))
	run.AssignColumn("a1", smartvectors.NewRegular(a1Vals))
	run.AssignColumn("b1", smartvectors.NewRegular(b1Vals))
	run.AssignColumn("c1", smartvectors.NewRegular(c1Vals))
}

// Advices returns a list of advices for the module discovery
func (d *PermutationTestCase) Advices() []*distributed.ModuleDiscoveryAdvice {
	return []*distributed.ModuleDiscoveryAdvice{
		distributed.SameSizeAdvice("module-0", d.wiop.Columns.GetHandle("a0")),
		distributed.SameSizeAdvice("module-0", d.wiop.Columns.GetHandle("a1")),
	}
}

// FibExtTestCase tests that ReceivedValuesGlobal correctly handles extension-field
// values across segment boundaries.
//
// The column extFib forms an arithmetic sequence over F_p^4:
//
//	extFib[i] = (i+1) * delta   where delta = 1 + 2u
//
// The global constraint extFib[i] - extFib[i-1] - delta = 0 forces each segment
// to receive the last value of the previous segment as a genuine extension-field
// element (non-zero upper coordinates), exercising the non-base code path in
// ReceivedValuesGlobal.
//
// At segment 0 the implicit "received" value is 0_fext, so extFib[0] = delta.
type FibExtTestCase struct {
	numRow int
	wiop   *wizard.CompiledIOP
}

func (d *FibExtTestCase) Name() string {
	return "FibExt"
}

// Define registers the single extension-field column and the step constraint.
func (d *FibExtTestCase) Define(comp *wizard.CompiledIOP) {
	d.wiop = comp

	// extFib lives in F_p^4 (isBase = false).
	extFib := comp.InsertCommit(0, "extFib", d.numRow, false)

	// delta = 1 + 2u  — a non-trivial extension-field element so that the
	// values sent across segment boundaries are genuinely extension-field and
	// not merely lifted base-field elements.
	delta := fext.NewFromInt(1, 2, 0, 0)

	// extFib[i] - extFib[i-1] - delta = 0  for all i.
	// At row 0 of each segment extFib[-1] is the last value of the previous
	// segment, transmitted as a ReceivedValuesGlobal extension-field entry.
	comp.InsertGlobal(0, "extFib-step",
		symbolic.Sub(
			ifaces.ColumnAsVariable(extFib),
			ifaces.ColumnAsVariable(column.Shift(extFib, -1)),
			symbolic.NewConstant(delta),
		),
	)
}

// Assign fills extFib with extFib[i] = (i+1)*delta.
//
// The first segment implicitly receives 0_fext, so the constraint
// extFib[0] - 0 - delta = 0  forces extFib[0] = delta.  By induction
// every subsequent value follows from the step constraint.
func (d *FibExtTestCase) Assign(run *wizard.ProverRuntime) {
	delta := fext.NewFromInt(1, 2, 0, 0)

	extFibVals := make([]fext.Element, d.numRow)
	extFibVals[0] = delta
	for i := 1; i < d.numRow; i++ {
		extFibVals[i].Add(&extFibVals[i-1], &delta)
	}

	run.AssignColumn("extFib", smartvectors.NewRegularExt(extFibVals))
}

// Advices returns the module-discovery advice for FibExtTestCase.
func (d *FibExtTestCase) Advices() []*distributed.ModuleDiscoveryAdvice {
	return []*distributed.ModuleDiscoveryAdvice{
		distributed.SameSizeAdvice("fib-ext-module", d.wiop.Columns.GetHandle("extFib")),
	}
}

// MultilinVortexTestCase tests that the multilinear Vortex protocol runs
// correctly through the distributed wizard. The MLEval query is marked IGNORED
// in Define so DistributeWizard never sees it. PostDistribute compiles the full
// ML protocol on the Bootstrapper after distribution, avoiding the round>0
// column check inside FilterCompiledIOP.
type MultilinVortexTestCase struct {
	numRow  int
	numVars int
	wiop    *wizard.CompiledIOP
}

func (d *MultilinVortexTestCase) Name() string { return "MultilinVortex" }

// Define registers structural columns plus a single MLEval column. The MLEval
// query is immediately marked as IGNORED so the distribution infrastructure
// never encounters it during FilterCompiledIOP.
func (d *MultilinVortexTestCase) Define(comp *wizard.CompiledIOP) {
	d.wiop = comp
	d.numVars = bits.Len(uint(d.numRow)) - 1

	// Two pseudo-modules differentiated by a duplicate constraint so they
	// produce distinct verifying keys (same pattern as LookupTestCase).
	a0 := comp.InsertCommit(0, "mlv_a0", d.numRow, true)
	b0 := comp.InsertCommit(0, "mlv_b0", d.numRow, true)
	c0 := comp.InsertCommit(0, "mlv_c0", d.numRow, true)
	a1 := comp.InsertCommit(0, "mlv_a1", d.numRow, true)
	b1 := comp.InsertCommit(0, "mlv_b1", d.numRow, true)
	c1 := comp.InsertCommit(0, "mlv_c1", d.numRow, true)

	comp.InsertGlobal(0, "mlv_global_0", symbolic.Sub(c0, b0, a0))
	comp.InsertGlobal(0, "mlv_global_0_dup", symbolic.Sub(c0, b0, a0))
	comp.InsertGlobal(0, "mlv_global_1", symbolic.Sub(c1, b1, a1))

	// Inclusion makes a0/b0/c0/a1/b1/c1 LPP columns so CompileSegment places
	// them at round 0 of the GL module, giving Vortex a round-0 Merkle root.
	comp.InsertInclusion(0, "mlv_self_incl", []ifaces.Column{c0, b0, a0}, []ifaces.Column{c1, b1, a1})

	col := comp.InsertCommit(0, "mlv_col", d.numRow, true)
	comp.InsertMultilinear(0, "MLV_EVAL", []ifaces.Column{col})
	// Mark as ignored so FilterCompiledIOP never sees it. The ML protocol is
	// compiled on the Bootstrapper by PostDistribute after DistributeWizard.
	comp.QueriesParams.MarkAsIgnored("MLV_EVAL")
}

// PostDistribute compiles the full multilinear Vortex opening on the
// Bootstrapper after DistributeWizard has run. This avoids the round>0 column
// panic in FilterCompiledIOP while still providing a proper ML proof on the
// Bootstrapper.
func (d *MultilinVortexTestCase) PostDistribute(dw *distributed.DistributedWizard) {
	// CompileAllRoundIgnored picks up the pre-ignored MLV_EVAL and batches all
	// same-round queries (mixed sizes) into one combined sumcheck per round.
	// CompileAllRound then re-batches Vortex residuals at each subsequent round.
	// 4 pairs of (CompileAllRound + Compile) are sufficient for numVars ≤ 8.
	wizard.ContinueCompilation(
		dw.Bootstrapper,
		multilineareval.CompileAllRoundIgnored,
		multilinvortex.Compile,
		multilineareval.CompileAllRound,
		multilinvortex.Compile,
		multilineareval.CompileAllRound,
		multilinvortex.Compile,
		multilineareval.CompileAllRound,
		multilinvortex.Compile,
		multilineareval.CompileAllRound,
	)
}

// Assign assigns the structural columns and the MLEval params for the Bootstrapper.
func (d *MultilinVortexTestCase) Assign(run *wizard.ProverRuntime) {
	// RightZeroPadded mirrors the LookupTestCase pattern: the last two rows
	// are zero so the inclusion query (mlv_self_incl) can match across modules.
	run.AssignColumn("mlv_a0", smartvectors.RightZeroPadded(vector.Repeat(field.One(), d.numRow-2), d.numRow))
	run.AssignColumn("mlv_b0", smartvectors.RightZeroPadded(vector.Repeat(field.NewElement(2), d.numRow-2), d.numRow))
	run.AssignColumn("mlv_c0", smartvectors.RightZeroPadded(vector.Repeat(field.NewElement(3), d.numRow-2), d.numRow))
	run.AssignColumn("mlv_a1", smartvectors.RightZeroPadded(vector.Repeat(field.One(), d.numRow-2), d.numRow))
	run.AssignColumn("mlv_b1", smartvectors.RightZeroPadded(vector.Repeat(field.NewElement(2), d.numRow-2), d.numRow))
	run.AssignColumn("mlv_c1", smartvectors.RightZeroPadded(vector.Repeat(field.NewElement(3), d.numRow-2), d.numRow))

	// MLEval column: values 1, 2, ..., numRow.
	colVals := make([]field.Element, d.numRow)
	for i := range colVals {
		colVals[i] = field.NewElement(uint64(i + 1))
	}
	run.AssignColumn("mlv_col", smartvectors.NewRegular(colVals))

	// Fixed evaluation point: (1, 2, ..., numVars) lifted into the extension field.
	point := make([]fext.Element, d.numVars)
	for i := range point {
		point[i] = fext.NewFromInt(int64(i+1), 0, 0, 0)
	}

	// Compute y = MultilinEval(colVals)(point).
	valsExt := make([]fext.Element, d.numRow)
	for i, v := range colVals {
		valsExt[i].B0.A0 = v
	}
	y := sumcheck.MultiLin(valsExt).Evaluate(point)

	run.AssignMultilinearExtShared("MLV_EVAL", point, y)
}

// Advices returns module-discovery advice for MultilinVortexTestCase.
func (d *MultilinVortexTestCase) Advices() []*distributed.ModuleDiscoveryAdvice {
	return []*distributed.ModuleDiscoveryAdvice{
		distributed.SameSizeAdvice("mlv-module-0", d.wiop.Columns.GetHandle("mlv_a0")),
		distributed.SameSizeAdvice("mlv-module-0", d.wiop.Columns.GetHandle("mlv_a1")),
		// mlv_col forms a singleton QBM (MLV_EVAL is ignored) so it needs
		// explicit advice to pass analyzeWithAdvices.
		distributed.SameSizeAdvice("mlv-module-0", d.wiop.Columns.GetHandle("mlv_col")),
	}
}

// MultilinVortexMixedTestCase exercises the cross-size batching path with two
// multilinear polynomials of different sizes: one big (numVarsBig = numVars of
// the large module) and one small (numVarsSmall = numVarsBig - 2, i.e. quarter
// the size). The two query pipelines under test are selected by useNewPipeline:
//   - false: classic per-size pipeline (CompileIgnored + Compile per-size)
//   - true:  cross-size pipeline (CompileAllRoundIgnored + CompileAllRound)
type MultilinVortexMixedTestCase struct {
	numRowBig      int
	numRowSmall    int
	useNewPipeline bool
	wiop           *wizard.CompiledIOP
}

func (d *MultilinVortexMixedTestCase) Name() string {
	if d.useNewPipeline {
		return "MultilinVortexMixed/New"
	}
	return "MultilinVortexMixed/Old"
}

// Define registers two sets of structural columns (big and small modules) each
// with an inclusion query so Vortex gets round-0 Merkle roots for both. The two
// MLEval queries are immediately marked IGNORED so FilterCompiledIOP never sees
// them; PostDistribute compiles the ML protocol after DistributeWizard.
func (d *MultilinVortexMixedTestCase) Define(comp *wizard.CompiledIOP) {
	d.wiop = comp
	d.numRowSmall = d.numRowBig / 4 // numVarsSmall = numVarsBig - 2

	// Big module structural columns (size numRowBig).
	a0b := comp.InsertCommit(0, "mix_a0b", d.numRowBig, true)
	b0b := comp.InsertCommit(0, "mix_b0b", d.numRowBig, true)
	c0b := comp.InsertCommit(0, "mix_c0b", d.numRowBig, true)
	a1b := comp.InsertCommit(0, "mix_a1b", d.numRowBig, true)
	b1b := comp.InsertCommit(0, "mix_b1b", d.numRowBig, true)
	c1b := comp.InsertCommit(0, "mix_c1b", d.numRowBig, true)
	comp.InsertGlobal(0, "mix_global_b0", symbolic.Sub(c0b, b0b, a0b))
	comp.InsertGlobal(0, "mix_global_b0_dup", symbolic.Sub(c0b, b0b, a0b))
	comp.InsertGlobal(0, "mix_global_b1", symbolic.Sub(c1b, b1b, a1b))
	comp.InsertInclusion(0, "mix_incl_big", []ifaces.Column{c0b, b0b, a0b}, []ifaces.Column{c1b, b1b, a1b})

	// Small module structural columns (size numRowSmall).
	a0s := comp.InsertCommit(0, "mix_a0s", d.numRowSmall, true)
	b0s := comp.InsertCommit(0, "mix_b0s", d.numRowSmall, true)
	c0s := comp.InsertCommit(0, "mix_c0s", d.numRowSmall, true)
	a1s := comp.InsertCommit(0, "mix_a1s", d.numRowSmall, true)
	b1s := comp.InsertCommit(0, "mix_b1s", d.numRowSmall, true)
	c1s := comp.InsertCommit(0, "mix_c1s", d.numRowSmall, true)
	comp.InsertGlobal(0, "mix_global_s0", symbolic.Sub(c0s, b0s, a0s))
	comp.InsertGlobal(0, "mix_global_s0_dup", symbolic.Sub(c0s, b0s, a0s))
	comp.InsertGlobal(0, "mix_global_s1", symbolic.Sub(c1s, b1s, a1s))
	comp.InsertInclusion(0, "mix_incl_small", []ifaces.Column{c0s, b0s, a0s}, []ifaces.Column{c1s, b1s, a1s})

	// ML columns: big (size numRowBig) and small (size numRowSmall).
	bigCol := comp.InsertCommit(0, "mix_big_col", d.numRowBig, true)
	smallCol := comp.InsertCommit(0, "mix_small_col", d.numRowSmall, true)

	comp.InsertMultilinear(0, "MIX_EVAL_BIG", []ifaces.Column{bigCol})
	comp.InsertMultilinear(0, "MIX_EVAL_SMALL", []ifaces.Column{smallCol})
	comp.QueriesParams.MarkAsIgnored("MIX_EVAL_BIG")
	comp.QueriesParams.MarkAsIgnored("MIX_EVAL_SMALL")
}

// PostDistribute selects the old (per-size) or new (cross-size) ML pipeline.
func (d *MultilinVortexMixedTestCase) PostDistribute(dw *distributed.DistributedWizard) {
	if d.useNewPipeline {
		wizard.ContinueCompilation(
			dw.Bootstrapper,
			multilineareval.CompileAllRoundIgnored,
			multilinvortex.Compile,
			multilineareval.CompileAllRound,
			multilinvortex.Compile,
			multilineareval.CompileAllRound,
			multilinvortex.Compile,
			multilineareval.CompileAllRound,
			multilinvortex.Compile,
			multilineareval.CompileAllRound,
		)
	} else {
		wizard.ContinueCompilation(
			dw.Bootstrapper,
			multilineareval.CompileIgnored,
			multilinvortex.Compile,
			multilineareval.Compile,
			multilinvortex.Compile,
			multilineareval.Compile,
			multilinvortex.Compile,
			multilinvortex.Compile,
			multilineareval.Compile,
		)
	}
}

// Assign assigns structural and ML column values for both modules.
func (d *MultilinVortexMixedTestCase) Assign(run *wizard.ProverRuntime) {
	numVarsBig := bits.Len(uint(d.numRowBig)) - 1
	numVarsSmall := bits.Len(uint(d.numRowSmall)) - 1

	// Structural columns satisfy c = a + b (i.e. 3 = 1 + 2) with last 2 rows
	// zero for inclusion query matching across modules.
	type abcGroup struct{ a, b, c ifaces.ColID; size int }
	for _, g := range []abcGroup{
		{"mix_a0b", "mix_b0b", "mix_c0b", d.numRowBig},
		{"mix_a1b", "mix_b1b", "mix_c1b", d.numRowBig},
		{"mix_a0s", "mix_b0s", "mix_c0s", d.numRowSmall},
		{"mix_a1s", "mix_b1s", "mix_c1s", d.numRowSmall},
	} {
		run.AssignColumn(g.a, smartvectors.RightZeroPadded(vector.Repeat(field.One(), g.size-2), g.size))
		run.AssignColumn(g.b, smartvectors.RightZeroPadded(vector.Repeat(field.NewElement(2), g.size-2), g.size))
		run.AssignColumn(g.c, smartvectors.RightZeroPadded(vector.Repeat(field.NewElement(3), g.size-2), g.size))
	}

	// Big ML column: values 1..numRowBig.
	bigVals := make([]field.Element, d.numRowBig)
	for i := range bigVals {
		bigVals[i] = field.NewElement(uint64(i + 1))
	}
	run.AssignColumn("mix_big_col", smartvectors.NewRegular(bigVals))

	// Small ML column: values 1..numRowSmall.
	smallVals := make([]field.Element, d.numRowSmall)
	for i := range smallVals {
		smallVals[i] = field.NewElement(uint64(i + 1))
	}
	run.AssignColumn("mix_small_col", smartvectors.NewRegular(smallVals))

	// Evaluation point for big polynomial.
	bigPoint := make([]fext.Element, numVarsBig)
	for i := range bigPoint {
		bigPoint[i] = fext.NewFromInt(int64(i+1), 0, 0, 0)
	}
	bigExt := make([]fext.Element, d.numRowBig)
	for i, v := range bigVals {
		bigExt[i].B0.A0 = v
	}
	bigY := sumcheck.MultiLin(bigExt).Evaluate(bigPoint)
	run.AssignMultilinearExtShared("MIX_EVAL_BIG", bigPoint, bigY)

	// Evaluation point for small polynomial.
	smallPoint := make([]fext.Element, numVarsSmall)
	for i := range smallPoint {
		smallPoint[i] = fext.NewFromInt(int64(i+1), 0, 0, 0)
	}
	smallExt := make([]fext.Element, d.numRowSmall)
	for i, v := range smallVals {
		smallExt[i].B0.A0 = v
	}
	smallY := sumcheck.MultiLin(smallExt).Evaluate(smallPoint)
	run.AssignMultilinearExtShared("MIX_EVAL_SMALL", smallPoint, smallY)
}

// Advices groups each module's columns under its own module name.
func (d *MultilinVortexMixedTestCase) Advices() []*distributed.ModuleDiscoveryAdvice {
	return []*distributed.ModuleDiscoveryAdvice{
		distributed.SameSizeAdvice("mix-module-big", d.wiop.Columns.GetHandle("mix_a0b")),
		distributed.SameSizeAdvice("mix-module-big", d.wiop.Columns.GetHandle("mix_a1b")),
		distributed.SameSizeAdvice("mix-module-big", d.wiop.Columns.GetHandle("mix_big_col")),
		distributed.SameSizeAdvice("mix-module-small", d.wiop.Columns.GetHandle("mix_a0s")),
		distributed.SameSizeAdvice("mix-module-small", d.wiop.Columns.GetHandle("mix_a1s")),
		distributed.SameSizeAdvice("mix-module-small", d.wiop.Columns.GetHandle("mix_small_col")),
	}
}
