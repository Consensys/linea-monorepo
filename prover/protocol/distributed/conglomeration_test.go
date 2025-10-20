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
