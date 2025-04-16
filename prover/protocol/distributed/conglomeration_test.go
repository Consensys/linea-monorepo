package distributed

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"reflect"
	"testing"
	"time"

	"github.com/consensys/linea-monorepo/prover/backend/execution"
	"github.com/consensys/linea-monorepo/prover/backend/files"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/recursion"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// TestConglomeration generates a conglomeration proof and checks if it is valid
func TestConglomeration(t *testing.T) {

	// t.Skipf("the test is a development/debug/integration test. It is not needed for CI")

	var (
		// #nosec G404 --we don't need a cryptographic RNG for testing purpose
		rng              = rand.New(utils.NewRandSource(0))
		sharedRandomness = field.PseudoRand(rng)
		zkevm            = GetZkEVM()
		disc             = &StandardModuleDiscoverer{
			TargetWeight: 1 << 28,
			Affinities:   GetAffinities(zkevm),
			Predivision:  16,
		}

		// This tests the compilation of the compiled-IOP
		distWizard = DistributeWizard(zkevm.WizardIOP, disc)

		// Minimal witness size to compile
		// minCompilationSize = 1 << 10
		compiledGLs  = make([]*RecursedSegmentCompilation, len(distWizard.GLs))
		compiledLPPs = make([]*RecursedSegmentCompilation, len(distWizard.LPPs))
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
		proof       = runtimeBoot.ExtractProof()
		verBootErr  = wizard.Verify(distWizard.Bootstrapper, proof)
	)

	t.Logf("[%v] done running the bootstrapper\n", time.Now())

	if verBootErr != nil {
		t.Fatalf("")
	}

	var (
		witnessGLs, witnessLPPs = SegmentRuntime(runtimeBoot, &distWizard)
		proofGLs                = make([]recursion.Witness, len(witnessGLs))
		proofLPPs               = make([]recursion.Witness, len(witnessLPPs))
		moduleWiop              = make([]*wizard.CompiledIOP, 0)
	)

	for i := range distWizard.LPPs {
		fmt.Printf("[%v] Starting to compile module LPP for %v\n", time.Now(), distWizard.LPPs[i].ModuleNames())
		compiledLPPs[i] = CompileSegment(distWizard.LPPs[i])
		fmt.Printf("[%v] Done compiling module LPP for %v\n", time.Now(), distWizard.LPPs[i].ModuleNames())
		moduleWiop = append(moduleWiop, compiledLPPs[i].RecursionComp)
	}

	// This applies the dummy.Compiler to all parts of the distributed wizard.
	for i := range distWizard.GLs {
		fmt.Printf("[%v] Starting to compile module GL for %v\n", time.Now(), distWizard.ModuleNames[i])
		compiledGLs[i] = CompileSegment(distWizard.GLs[i])
		fmt.Printf("[%v] Done compiling module GL for %v\n", time.Now(), distWizard.ModuleNames[i])
		moduleWiop = append(moduleWiop, compiledGLs[i].RecursionComp)
	}

	// This does the compilation of the conglomerator proof
	conglomeration := Conglomerate(max(20, len(witnessGLs)+len(witnessLPPs)), moduleWiop)

	for i := range witnessGLs {

		var (
			witnessGL = witnessGLs[i]
			moduleGL  *RecursedSegmentCompilation
		)

		t.Logf("segment(total)=%v module=%v segment.index=%v", i, witnessGL.ModuleName, witnessGL.ModuleIndex)

		for k := range distWizard.ModuleNames {

			if distWizard.ModuleNames[k] != witnessGLs[i].ModuleName {
				continue
			}

			moduleGL = compiledGLs[k]
		}

		if moduleGL == nil {
			t.Fatalf("module does not exists")
		}

		t.Logf("RUNNING THE GL PROVER: %v", time.Now())

		run := moduleGL.ProveSegment(witnessGL)
		proofGLs[i] = recursion.ExtractWitness(run)

		t.Logf("RUNNING THE GL PROVER - DONE: %v", time.Now())
	}

	for i := range witnessLPPs {

		var (
			witnessLPP = witnessLPPs[i]
			moduleLPP  *RecursedSegmentCompilation
		)

		witnessLPP.InitialFiatShamirState = sharedRandomness

		t.Logf("segment(total)=%v module=%v segment.index=%v", i, witnessLPP.ModuleNames, witnessLPP.ModuleIndex)

		for k := range distWizard.LPPs {

			if !reflect.DeepEqual(distWizard.LPPs[k].ModuleNames(), witnessLPPs[i].ModuleNames) {
				continue
			}

			moduleLPP = compiledLPPs[k]
		}

		if moduleLPP == nil {
			t.Fatalf("module does not exists")
		}

		t.Logf("RUNNING THE LPP PROVER: %v", time.Now())

		run := moduleLPP.ProveSegment(witnessLPP)
		proofLPPs[i] = recursion.ExtractWitness(run)

		t.Logf("RUNNING THE LPP PROVER - DONE: %v", time.Now())
	}

	_ = conglomeration.Prove(proofGLs, proofLPPs)
}
