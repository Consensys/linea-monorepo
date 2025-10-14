package distributed_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/sirupsen/logrus"
)

// TestConglomerationBasic generates a conglomeration proof and checks if it is valid
func TestConglomerationBasic(t *testing.T) {

	var (
		numRow = 1 << 10
		tc     = DistributeTestCase{numRow: numRow}
		disc   = &distributed.StandardModuleDiscoverer{
			TargetWeight: 3 * numRow / 2,
			Predivision:  1,
		}
		comp = wizard.Compile(func(build *wizard.Builder) {
			tc.Define(build.CompiledIOP)
		})

		// This tests the compilation of the compiled-IOP
		distWizard = distributed.DistributeWizard(comp, disc).
				CompileSegments().
				Conglomerate()

		runtimeBoot             = wizard.RunProver(distWizard.Bootstrapper, tc.Assign)
		witnessGLs, witnessLPPs = distributed.SegmentRuntime(
			runtimeBoot,
			distWizard.Disc,
			distWizard.BlueprintGLs,
			distWizard.BlueprintLPPs,
			distWizard.VerificationKeyMerkleTree.GetRoot(),
		)
		glProofs         = runProverGLs(t, distWizard, witnessGLs)
		sharedRandomness = distributed.GetSharedRandomnessFromSegmentProofs(glProofs)
		runLPPs          = runProverLPPs(t, distWizard, sharedRandomness, witnessLPPs)
	)

	runConglomerationProver(
		t,
		&distWizard.VerificationKeyMerkleTree,
		distWizard.CompiledConglomeration,
		glProofs,
		runLPPs,
	)
}

// // TestConglomeration generates a conglomeration proof and checks if it is valid
// func TestConglomeration(t *testing.T) {

// 	t.Skipf("the test is a development/debug/integration test. It is not needed for CI")

// 	var (
// 		z    = zkevm.GetTestZkEVM()
// 		disc = &distributed.StandardModuleDiscoverer{
// 			TargetWeight: 1 << 28,
// 			Affinities:   zkevm.GetAffinities(z),
// 			Predivision:  1,
// 		}

// 		// This tests the compilation of the compiled-IOP
// 		distWizard = distributed.DistributeWizard(z.WizardIOP, disc).
// 				CompileSegments().
// 				Conglomerate()
// 	)

// 	var (
// 		reqFile      = files.MustRead("/home/ubuntu/beta-v2-rc11/10556002-10556002-etv0.2.0-stv2.2.2-getZkProof.json")
// 		cfgFilePath  = "/home/ubuntu/zkevm-monorepo/prover/config/config-sepolia-full.toml"
// 		req          = &execution.Request{}
// 		reqDecodeErr = json.NewDecoder(reqFile).Decode(req)
// 		cfg, cfgErr  = config.NewConfigFromFileUnchecked(cfgFilePath)
// 	)

// 	if reqDecodeErr != nil {
// 		t.Fatalf("could not read the request file: %v", reqDecodeErr)
// 	}

// 	if cfgErr != nil {
// 		t.Fatalf("could not read the config file: err=%v, cfg=%++v", cfgErr, cfg)
// 	}

// 	t.Logf("loaded config: %++v", cfg)

// 	t.Logf("[%v] running the bootstrapper\n", time.Now())

// 	var (
// 		_, witness  = test_utils.GetZkevmWitness(req, cfg)
// 		runtimeBoot = wizard.RunProver(distWizard.Bootstrapper, z.GetMainProverStep(witness))
// 	)

// 	t.Logf("[%v] done running the bootstrapper\n", time.Now())

// 	var (
// 		witnessGLs, witnessLPPs = distributed.SegmentRuntime(
// 			runtimeBoot,
// 			distWizard.Disc,
// 			distWizard.BlueprintGLs,
// 			distWizard.BlueprintLPPs,
// 		)
// 		runGLs = runProverGLs(t, distWizard, witnessGLs)
// 	)

// 	for i := range runGLs {
// 		t.Logf("sanity-checking runGLs[%d]\n", i)
// 		sanityCheckConglomeration(t, distWizard.CompiledConglomeration, runGLs[i])
// 	}

// 	var (
// 		sharedRandomness = getSharedRandomness(runGLs)
// 		runLPPs          = runProverLPPs(t, distWizard, sharedRandomness, witnessLPPs)
// 	)

// 	for i := range runLPPs {
// 		t.Logf("sanity-checking runLPPs[%d]\n", i)
// 		sanityCheckConglomeration(t, distWizard.CompiledConglomeration, runLPPs[i])
// 	}

// 	runConglomerationProver(t, distWizard.CompiledConglomeration, runGLs, runLPPs)
// }

// // Sanity-check for conglomeration compilation.
// func sanityCheckConglomeration(t *testing.T, cong *distributed.ConglomeratorCompilation, run *wizard.ProverRuntime) {

// 	t.Logf("sanity-check for conglomeration")
// 	stopRound := recursion.VortexQueryRound(cong.ModuleGLIops[0])
// 	err := wizard.VerifyUntilRound(cong.ModuleGLIops[0], run.ExtractProof(), stopRound+1)

// 	if err != nil {
// 		t.Fatalf("could not verify proof: %v", err)
// 	}
// }

// This function runs a prover for a conglomerator compilation. It takes in a ConglomeratorCompilation
// object and two slices of ProverRuntime objects, runGLs and runLPPs. It extracts witnesses from
// these runtimes, then uses the ConglomeratorCompilation object to prove the conglomerator,
// logging the start and end times of the proof process.
func runConglomerationProver(
	t *testing.T,
	mt *distributed.VerificationKeyMerkleTree,
	cong *distributed.RecursedSegmentCompilation,
	runGLs, runLPPs []distributed.SegmentProof,
) distributed.SegmentProof {

	// The channel is used as a FIFO queue to store the remaining proofs to be
	// aggregated.
	var (
		remainingProofs = make(chan distributed.SegmentProof, len(runGLs)+len(runLPPs))
	)

	// This populates the queue
	for i := range runGLs {
		remainingProofs <- runGLs[i]
	}

	for i := range runLPPs {
		remainingProofs <- runLPPs[i]
	}

	// TryPopQueue attempts to consume a proof from the queue or return false
	// if the queue is empty.
	tryPopQueue := func() (distributed.SegmentProof, bool) {
		select {
		case proof := <-remainingProofs:
			return proof, true
		default:
			return distributed.SegmentProof{}, false
		}
	}

	// This is the actual proof aggregation loop.
	for {
		a, ok := tryPopQueue()
		if !ok {
			panic("the queue cannot be empty here")
		}

		// If b cannot be found, it means that the queue contained only a single
		// proof to aggregate which means it was the result of the function.
		b, ok := tryPopQueue()
		if !ok {
			return a
		}

		logrus.Infof("AGGREGATING PROOF, remaining %v\n", len(remainingProofs))

		new := cong.ProveSegment(&distributed.ModuleWitnessConglo{
			SegmentProofs:             []distributed.SegmentProof{a, b},
			VerificationKeyMerkleTree: *mt,
		})

		logrus.Infof("AGGREGATED PROOFS\n")

		remainingProofs <- new
	}
}
