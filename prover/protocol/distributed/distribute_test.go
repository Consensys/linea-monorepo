package distributed_test

import (
	"fmt"
	"testing"
	"time"

	multisethashing "github.com/consensys/linea-monorepo/prover/crypto/multisethashing_koalabear"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/protocol/accessors"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
)

const (
	NbRow = 8
)

// testCompilationParams defines compilation parameters for testing segment compilation
// Note: ColumnProfileMPTS is left nil to avoid profile size constraints during testing
var testCompilationParams = distributed.CompilationParams{
	FixedNbRowPlonkCircuit:       1 << 18,
	FixedNbRowExternalHasher:     1 << 14,
	FixedNbPublicInput:           1 << 10,
	InitialCompilerSize:          1 << 18,
	InitialCompilerSizeConglo:    1 << 13,
	ColumnProfileMPTS:            nil, // nil disables profile checking
	ColumnProfileMPTSPrecomputed: 0,
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
}

// TestDistributedWizard runs the distributed wizard test over multiple test cases.
// Each test case defines a different set of queries (inclusion, projection, permutation).
func TestDistributedWizard(t *testing.T) {

	testCases := []DistributedTestCase{
		&LookupTestCase{numRow: 1 << NbRow},
		&ProjectionTestCase{numRow: 1 << NbRow},
		&PermutationTestCase{numRow: 1 << NbRow},
	}

	for _, tc := range testCases {
		t.Run(tc.Name(), func(t *testing.T) {
			runDistributedWizardTest(t, tc, false)
		})
	}
}

func TestDistributedWizardWithSegmentCompilation(t *testing.T) {
	t.Skipf(" the test is skipped since vortex is not yet implemented for extension/post-recursion")

	testCases := []DistributedTestCase{
		&LookupTestCase{numRow: 1 << NbRow},
	}

	for _, tc := range testCases {
		t.Run(tc.Name(), func(t *testing.T) {
			runDistributedWizardTest(t, tc, true)
		})
	}
}

// runDistributedWizardTest runs a single distributed wizard test case.
func runDistributedWizardTest(t *testing.T, tc DistributedTestCase, segmentCompilation bool) {

	var (
		defFunc = func(build *wizard.Builder) { tc.Define(build.CompiledIOP) }
		wiop    = wizard.Compile(defFunc)
		disc    = &distributed.StandardModuleDiscoverer{
			TargetWeight: NbRow,
			Predivision:  1,
		}
		distWizard *distributed.DistributedWizard
	)

	// This tests the compilation of the compiled-IOP
	if segmentCompilation {
		// compile with vortex/ self-recursion/recursion
		distWizard = distributed.DistributeWizard(wiop, disc).
			CompileSegments(testCompilationParams).
			Conglomerate(testCompilationParams)
	} else {
		// dummy compilation
		distWizard = distributed.DistributeWizard(wiop, disc)
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
		runtimeBoot = wizard.RunProver(distWizard.Bootstrapper, tc.Assign)
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
			proverRunGL         = wizard.RunProver(moduleGL.Wiop, moduleGL.GetMainProverStep(witnessGLs[i]))
			proofGL             = proverRunGL.ExtractProof()
			verRun, verGLErr    = wizard.VerifyWithRuntime(moduleGL.Wiop, proofGL)
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
			proverRunLPP         = wizard.RunProver(moduleLPP.Wiop, moduleLPP.GetMainProverStep(witnessLPP))
			proofLPP             = proverRunLPP.ExtractProof()
			verRun, verLPPErr    = wizard.VerifyWithRuntime(moduleLPP.Wiop, proofLPP)
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

	for i := 0; i < len(generalMSet); i++ {
		if !generalMSet.IsEmpty() {
			t.Errorf("generalMSet does not cancel: ")
		}
	}
}

// LookupTestCase tests the basic distributed wizard with inclusion query (lookup).
// The testcase generates 2 triplets of columns a, b, c such that a + b = c
// and the two modules are joined by a lookup.
type LookupTestCase struct {
	numRow int
}

func (d *LookupTestCase) Name() string {
	return "Lookup"
}

// Define defines the structure of the distributed wizard. The structure is
// composed of 2 modules that are connected by a lookup. The two modules are
// identical and are defined as a + b = c.
func (d *LookupTestCase) Define(comp *wizard.CompiledIOP) {

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

// ProjectionTestCase includes a projection query in addition to
// the global and inclusion constraints. Projection queries are compiled
// into Horner evaluations.
type ProjectionTestCase struct {
	numRow int
}

func (d *ProjectionTestCase) Name() string {
	return "Projection"
}

// Define defines the structure with projection query. It has:
// - Global constraints a + b = c for both modules
// - An inclusion query between modules
// - A projection query between filtered columns
func (d *ProjectionTestCase) Define(comp *wizard.CompiledIOP) {

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

// PermutationTestCase includes a permutation query in addition to
// the global constraints. Permutation queries are compiled into
// grand product arguments.
type PermutationTestCase struct {
	numRow int
}

func (d *PermutationTestCase) Name() string {
	return "Permutation"
}

// Define defines the structure with permutation query. It has:
// - Global constraints a + b = c for both modules
// - A permutation query asserting that (a0, b0, c0) and (a1, b1, c1) contain the same rows
func (d *PermutationTestCase) Define(comp *wizard.CompiledIOP) {

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

// runProverGLs executes the prover for each GL module segment. It takes in a list of
// compiled GL segments and corresponding witnesses, then runs the prover for each
// segment. The function logs the start and end times of the prover execution for each
// segment. It returns a slice of ProverRuntime instances, each representing the
// result of the prover execution for a segment.
func runProverGLs(
	t *testing.T,
	distWizard *distributed.DistributedWizard,
	witnessGLs []*distributed.ModuleWitnessGL,
) (proofs []*distributed.SegmentProof) {

	var (
		compiledGLs = distWizard.CompiledGLs
	)

	proofs = make([]*distributed.SegmentProof, len(witnessGLs))

	for i := range witnessGLs {

		var (
			witnessGL = witnessGLs[i]
			moduleGL  *distributed.RecursedSegmentCompilation
		)

		t.Logf("segment(total)=%v module=%v module.index=%v segment.index=%v", i, witnessGL.ModuleName, witnessGL.ModuleIndex, witnessGL.SegmentModuleIndex)
		for k := range distWizard.ModuleNames {
			if distWizard.ModuleNames[k] != witnessGLs[i].ModuleName {
				continue
			}
			moduleGL = compiledGLs[k]
		}

		if moduleGL == nil {
			t.Fatalf("module does not exists, module=%v, distWizard.ModuleNames=%v", witnessGL.ModuleName, distWizard.ModuleNames)
		}

		moduleGL.RecursionCompForCheck = distWizard.CompiledGLs[0].RecursionComp

		t.Logf("RUNNING THE GL PROVER: %v", time.Now())
		proofs[i] = moduleGL.ProveSegment(witnessGL)
		t.Logf("RUNNING THE GL PROVER - DONE: %v", time.Now())

	}

	return proofs
}

// runProverLPPs runs a prover for a LPP segment. It takes in a DistributedWizard
// object, a slice of RecursedSegmentCompilation objects, and a slice of
// ModuleWitnessLPP objects. It runs the prover for each segment and logs the
// time at which the prover starts and ends. It returns a slice of ProverRuntime
// instances, each representing the result of the prover execution for a segment.
func runProverLPPs(
	t *testing.T,
	distWizard *distributed.DistributedWizard,
	sharedRandomness field.Octuplet,
	witnessLPPs []*distributed.ModuleWitnessLPP,
) []*distributed.SegmentProof {

	var (
		proofs = make([]*distributed.SegmentProof, len(witnessLPPs))
	)

	for i := range witnessLPPs {

		var (
			witnessLPP  = witnessLPPs[i]
			moduleIndex = witnessLPP.ModuleIndex
			moduleLPP   = distWizard.CompiledLPPs[moduleIndex]
		)

		moduleLPP.RecursionCompForCheck = distWizard.CompiledGLs[0].RecursionComp

		witnessLPP.InitialFiatShamirState = sharedRandomness

		t.Logf("segment(total)=%v module=%v module.index=%v segment.index=%v", i, witnessLPP.ModuleName, witnessLPP.ModuleIndex, witnessLPP.SegmentModuleIndex)
		t.Logf("RUNNING THE LPP PROVER: %v", time.Now())
		proofs[i] = moduleLPP.ProveSegment(witnessLPP)
		t.Logf("RUNNING THE LPP PROVER - DONE: %v", time.Now())
	}

	return proofs
}
