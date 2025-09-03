package serialization_test

import (
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/recursion"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"

	"github.com/consensys/linea-monorepo/prover/protocol/serialization"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// This test is necessary to do a quick check if omitting certain fields in certain
// compiled objects has any effects on generating the proof
func TestBasicProofSanity(t *testing.T) {

	t.Skipf("the test is a development/debug/integration test. It is not needed for CI")

	// Setup
	dw := GetBasicDW()
	testPath := "/tmp/dw"
	defer os.RemoveAll(testPath)

	// Serialize and load to file
	if err := serialization.StoreToDisk(testPath, dw, true); err != nil {
		t.Fatalf("could not serialize %s: %s", testPath, err)
	}

	dw = nil
	runtime.GC()

	// Load the deserialized wizard from the file and sanity check the proof
	distWizard := &distributed.DistributedWizard{}
	if err := serialization.LoadFromDisk(testPath, distWizard, true); err != nil {
		t.Fatalf("could not deserialize %s: %s", testPath, err)
	}

	var (
		numRow                  = 1 << 10
		tc                      = distributeTestCase{numRow: numRow}
		runtimeBoot             = wizard.RunProver(distWizard.Bootstrapper, tc.assign)
		witnessGLs, witnessLPPs = distributed.SegmentRuntime(
			runtimeBoot,
			distWizard.Disc,
			distWizard.BlueprintGLs,
			distWizard.BlueprintLPPs,
		)
		runGLs = RunProverGLs(t, distWizard, witnessGLs)
	)

	for i := range runGLs {
		t.Logf("sanity-checking runGLs[%d]\n", i)
		SanityCheckConglomeration(t, distWizard.CompiledConglomeration, runGLs[i])
	}

	var (
		sharedRandomness = GetSharedRandomness(runGLs)
		runLPPs          = RunProverLPPs(t, distWizard, sharedRandomness, witnessLPPs)
	)

	for i := range runLPPs {
		t.Logf("sanity-checking runLPPs[%d]\n", i)
		SanityCheckConglomeration(t, distWizard.CompiledConglomeration, runLPPs[i])
	}

	RunConglomerationProver(t, distWizard.CompiledConglomeration, runGLs, runLPPs)
}

// GetSharedRandomness computes the shared randomnesses from the runtime
func GetSharedRandomness(runs []*wizard.ProverRuntime) field.Element {
	witnesses := make([]recursion.Witness, len(runs))
	for i := range runs {
		witnesses[i] = recursion.ExtractWitness(runs[i])
	}

	comps := make([]*wizard.CompiledIOP, len(runs))
	for i := range runs {
		comps[i] = runs[i].Spec
	}

	return distributed.GetSharedRandomnessFromWitnesses(comps, witnesses)
}

func RunConglomerationProver(t *testing.T, cong *distributed.ConglomeratorCompilation, runGLs, runLPPs []*wizard.ProverRuntime) {

	var (
		witLPPs = make([]recursion.Witness, len(runLPPs))
		witGLs  = make([]recursion.Witness, len(runGLs))
	)

	for i := range runLPPs {
		witLPPs[i] = recursion.ExtractWitness(runLPPs[i])
	}

	for i := range runGLs {
		witGLs[i] = recursion.ExtractWitness(runGLs[i])
	}

	t.Logf("[%v] Starting to prove conglomerator\n", time.Now())
	proof := cong.Prove(witGLs, witLPPs)
	t.Logf("[%v] Finished proving conglomerator\n", time.Now())

	t.Logf("[%v] start sanity-checking proof\n", time.Now())

	err := wizard.Verify(cong.Wiop, proof)
	if err != nil {
		t.Fatalf("could not verify proof: %v", err)
	}

	t.Logf("[%v] done sanity-checking proof\n", time.Now())
}

// RunProverGLs executes the prover for each GL module segment. It takes in a list of
// compiled GL segments and corresponding witnesses, then runs the prover for each
// segment. The function logs the start and end times of the prover execution for each
// segment. It returns a slice of ProverRuntime instances, each representing the
// result of the prover execution for a segment.
func RunProverGLs(
	t *testing.T,
	distWizard *distributed.DistributedWizard,
	witnessGLs []*distributed.ModuleWitnessGL,
) []*wizard.ProverRuntime {

	var (
		compiledGLs = distWizard.CompiledGLs
		runs        = make([]*wizard.ProverRuntime, len(witnessGLs))
	)

	for i := range witnessGLs {

		var (
			witnessGL = witnessGLs[i]
			moduleGL  *distributed.RecursedSegmentCompilation
		)

		t.Logf("segment(total)=%v module=%v segment.index=%v", i, witnessGL.ModuleName, witnessGL.ModuleIndex)
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
		runs[i] = moduleGL.ProveSegment(witnessGL)
		t.Logf("RUNNING THE GL PROVER - DONE: %v", time.Now())
	}

	return runs
}

// RunProverLPPs runs a prover for a LPP segment. It takes in a DistributedWizard
// object, a slice of RecursedSegmentCompilation objects, and a slice of
// ModuleWitnessLPP objects. It runs the prover for each segment and logs the
// time at which the prover starts and ends. It returns a slice of ProverRuntime
// instances, each representing the result of the prover execution for a segment.
func RunProverLPPs(
	t *testing.T,
	distWizard *distributed.DistributedWizard,
	sharedRandomness field.Element,
	witnessLPPs []*distributed.ModuleWitnessLPP,
) []*wizard.ProverRuntime {

	var (
		runs         = make([]*wizard.ProverRuntime, len(witnessLPPs))
		compiledLPPs = distWizard.CompiledLPPs
	)

	for i := range witnessLPPs {

		var (
			witnessLPP      = witnessLPPs[i]
			moduleLPP       *distributed.RecursedSegmentCompilation
			distModuleNames = [][]distributed.ModuleName{}
		)

		witnessLPP.InitialFiatShamirState = sharedRandomness

		t.Logf("segment(total)=%v module=%v segment.index=%v", i, witnessLPP.ModuleName, witnessLPP.ModuleIndex)

	ModuleLoop:
		for k := range distWizard.LPPs {

			moduleList := distWizard.LPPs[k].ModuleNames()
			distModuleNames = append(distModuleNames, moduleList)

			for l, m := range moduleList {
				if m != witnessLPP.ModuleName[l] {
					continue ModuleLoop
				}
			}

			moduleLPP = compiledLPPs[k]
			break
		}

		if moduleLPP == nil {
			t.Fatalf("module does not exists, moduleName=%v distModuleNames=%v", witnessLPP.ModuleName, distModuleNames)
		}

		t.Logf("RUNNING THE LPP PROVER: %v", time.Now())
		runs[i] = moduleLPP.ProveSegment(witnessLPP)
		t.Logf("RUNNING THE LPP PROVER - DONE: %v", time.Now())
	}

	return runs
}

func SanityCheckConglomeration(t *testing.T, cong *distributed.ConglomeratorCompilation, run *wizard.ProverRuntime) {

	t.Logf("sanity-check for conglomeration")
	stopRound := recursion.VortexQueryRound(cong.ModuleGLIops[0])
	err := wizard.VerifyUntilRound(cong.ModuleGLIops[0], run.ExtractProof(), stopRound+1)

	if err != nil {
		t.Fatalf("could not verify proof: %v", err)
	}
}
