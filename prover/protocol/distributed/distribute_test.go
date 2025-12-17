package distributed_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/consensys/linea-monorepo/prover/crypto/hasher_factory"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/protocol/accessors"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
)

// TestDistributedWizardBasic attempts to compiler the wizard distribution.
func TestDistributedWizardBasic(t *testing.T) {

	var (
		z       = DistributeTestCase{numRow: 1 << 20}
		defFunc = func(build *wizard.Builder) { z.Define(build.CompiledIOP) }
		wiop    = wizard.Compile(defFunc)
		disc    = &distributed.StandardModuleDiscoverer{
			TargetWeight: 1 << 20,
			Predivision:  1,
		}

		// This tests the compilation of the compiled-IOP
		distWizard = distributed.DistributeWizard(wiop, disc)
	)

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
		runtimeBoot = wizard.RunProver(distWizard.Bootstrapper, z.Assign, false)
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
		generalMSet         = hasher_factory.MSetHash{}
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
			// moduleIndex = witnessGLs[i].ModuleIndex
			// moduleName  = witnessGLs[i].ModuleName
			moduleGL *distributed.ModuleGL
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
			generalMSetFromGLFr = distributed.GetPublicInputList(verRun, distributed.GeneralMultiSetPublicInputBase, hasher_factory.MSetHashSize)
			generalMSetFromGL   = hasher_factory.MSetHash(generalMSetFromGLFr)
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
			proverRunLPP         = wizard.RunProver(moduleLPP.Wiop, moduleLPP.GetMainProverStep(witnessLPP), false)
			proofLPP             = proverRunLPP.ExtractProof()
			verRun, verLPPErr    = wizard.VerifyWithRuntime(moduleLPP.Wiop, proofLPP, false)
			generalMSetFromLPPFr = distributed.GetPublicInputList(verRun, distributed.GeneralMultiSetPublicInputBase, hasher_factory.MSetHashSize)
			generalMSetFromLPP   = hasher_factory.MSetHash(generalMSetFromLPPFr)
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

// DistributeTestCase is an implementation of the testcase interface. The
// testcase generates 2 triplets of columns a, b, c such that a + b = c
// and the two modules are joined by a lookup.
type DistributeTestCase struct {
	numRow int
}

// Define defines the structure of the distributed wizard. The structure is
// composed of 2 modules that are connected by a lookup. The two modules are
// identical and are defined as a + b = c. The a, b and c are each defined as
// a commit in the wizard. The lookup is defined as a global constraint that
// enforces the equality of the two modules.
func (d DistributeTestCase) Define(comp *wizard.CompiledIOP) {

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

// Assign sets up the column assignments for the DistributeTestCase
// within the ProverRuntime. It assigns constant values to six columns
// ('a0', 'b0', 'c0', 'a1', 'b1', 'c1') where each column is assigned
// a smart vector with a constant field element value and a specified
// number of rows (d.numRow). This function helps initialize the columns
// with predetermined values for the testing setup.
func (d DistributeTestCase) Assign(run *wizard.ProverRuntime) {
	run.AssignColumn("a0", smartvectors.RightZeroPadded(vector.Repeat(field.NewElement(1), d.numRow-2), d.numRow))
	run.AssignColumn("b0", smartvectors.RightZeroPadded(vector.Repeat(field.NewElement(2), d.numRow-2), d.numRow))
	run.AssignColumn("c0", smartvectors.RightZeroPadded(vector.Repeat(field.NewElement(3), d.numRow-2), d.numRow))
	run.AssignColumn("a1", smartvectors.RightZeroPadded(vector.Repeat(field.NewElement(1), d.numRow-2), d.numRow))
	run.AssignColumn("b1", smartvectors.RightZeroPadded(vector.Repeat(field.NewElement(2), d.numRow-2), d.numRow))
	run.AssignColumn("c1", smartvectors.RightZeroPadded(vector.Repeat(field.NewElement(3), d.numRow-2), d.numRow))
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

		witnessLPP.InitialFiatShamirState = sharedRandomness

		t.Logf("segment(total)=%v module=%v module.index=%v segment.index=%v", i, witnessLPP.ModuleName, witnessLPP.ModuleIndex, witnessLPP.SegmentModuleIndex)
		t.Logf("RUNNING THE LPP PROVER: %v", time.Now())
		proofs[i] = moduleLPP.ProveSegment(witnessLPP)
		t.Logf("RUNNING THE LPP PROVER - DONE: %v", time.Now())
	}

	return proofs
}
