package distributed

import (
	"encoding/json"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/consensys/linea-monorepo/prover/backend/execution"
	"github.com/consensys/linea-monorepo/prover/backend/files"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/recursion"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// TestConglomerationBasic generates a conglomeration proof and checks if it is valid
func TestConglomerationBasic(t *testing.T) {

	var (
		numRow = 1 << 10
		tc     = DistributeTestCase{numRow: numRow}
		disc   = &StandardModuleDiscoverer{
			TargetWeight: 3 * numRow,
			Predivision:  1,
		}
		comp = wizard.Compile(func(build *wizard.Builder) {
			tc.Define(build.CompiledIOP)
		})

		// This tests the compilation of the compiled-IOP
		distWizard                = DistributeWizard(comp, disc)
		compiledGLs, compiledLPPs = compileAllSegments(t, distWizard)

		// This does the compilation of the conglomerator proof
		conglomeration = runConglomerationCompiler(t, slices.Concat(compiledGLs, compiledLPPs))
		runtimeBoot    = wizard.RunProver(distWizard.Bootstrapper, tc.Assign)

		witnessGLs, witnessLPPs = SegmentRuntime(runtimeBoot, &distWizard)
	)

	fmt.Printf("nbWitnessesGL=%d nbWitnessesLPP=%d\n", len(witnessGLs), len(witnessLPPs))

	runGLs := runProverGLs(t, distWizard, compiledGLs, witnessGLs)
	runLPPs := runProverLPPs(t, distWizard, compiledLPPs, witnessLPPs)
	runConglomerationProver(t, conglomeration, runGLs, runLPPs)
}

// TestConglomeration generates a conglomeration proof and checks if it is valid
func TestConglomeration(t *testing.T) {

	// t.Skipf("the test is a development/debug/integration test. It is not needed for CI")

	var (
		zkevm = GetZkEVM()
		disc  = &StandardModuleDiscoverer{
			TargetWeight: 1 << 28,
			Affinities:   GetAffinities(zkevm),
			Predivision:  16,
		}

		// This tests the compilation of the compiled-IOP
		distWizard                = DistributeWizard(zkevm.WizardIOP, disc)
		compiledGLs, compiledLPPs = compileAllSegments(t, distWizard)

		// This does the compilation of the conglomerator proof
		conglomeration = runConglomerationCompiler(t, slices.Concat(compiledGLs, compiledLPPs))
	)

	var (
		reqFile      = files.MustRead("/home/ubuntu/beta-v2-rc11/10556002-10556002-etv0.2.0-stv2.2.2-getZkProof.json")
		cfgFilePath  = "/home/ubuntu/zkevm-monorepo/prover/config/config-sepolia-full.toml"
		req          = &execution.Request{}
		reqDecodeErr = json.NewDecoder(reqFile).Decode(req)
		cfg, cfgErr  = config.NewConfigFromFileUnchecked(cfgFilePath)
	)

	if reqDecodeErr != nil {
		t.Fatalf("could not read the request file: %v", reqDecodeErr)
	}

	if cfgErr != nil {
		t.Fatalf("could not read the config file: err=%v, cfg=%++v", cfgErr, cfg)
	}

	t.Logf("loaded config: %++v", cfg)

	t.Logf("[%v] running the bootstrapper\n", time.Now())

	var (
		witness     = GetZkevmWitness(req, cfg)
		runtimeBoot = wizard.RunProver(distWizard.Bootstrapper, zkevm.GetMainProverStep(witness))
	)

	t.Logf("[%v] done running the bootstrapper\n", time.Now())

	var (
		witnessGLs, witnessLPPs = SegmentRuntime(runtimeBoot, &distWizard)
		runGLs                  = runProverGLs(t, distWizard, compiledGLs, witnessGLs)
		runLPPs                 = runProverLPPs(t, distWizard, compiledLPPs, witnessLPPs)
	)

	runConglomerationProver(t, conglomeration, runGLs, runLPPs)
}

// runConglomerationCompiler compiles the conglomeration proof and returns a new ConglomeratorCompilation object.
// The function takes a slice of RecursedSegmentCompilation objects and compiles the
// ConglomeratorCompilation object with the given number of proofs.
func runConglomerationCompiler(t *testing.T, compiledSegments []*RecursedSegmentCompilation) *ConglomeratorCompilation {

	wiops := make([]*wizard.CompiledIOP, len(compiledSegments))

	for i := range compiledSegments {
		wiops[i] = compiledSegments[i].RecursionComp
	}

	t.Logf("[%v] Starting to compile conglomerator\n", time.Now())
	cong := Conglomerate(20, wiops)
	t.Logf("[%v] Finished compiling conglomerator\n", time.Now())

	return cong
}

// This function runs a prover for a conglomerator compilation. It takes in a ConglomeratorCompilation
// object and two slices of ProverRuntime objects, runGLs and runLPPs. It extracts witnesses from
// these runtimes, then uses the ConglomeratorCompilation object to prove the conglomerator,
// logging the start and end times of the proof process.
func runConglomerationProver(t *testing.T, cong *ConglomeratorCompilation, runGLs, runLPPs []*wizard.ProverRuntime) {

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
	_ = cong.Prove(witGLs, witLPPs)
	t.Logf("[%v] Finished proving conglomerator\n", time.Now())
}
