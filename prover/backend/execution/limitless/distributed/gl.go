package distributed

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/recursion"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/serialization"
	"github.com/consensys/linea-monorepo/prover/zkevm"
	"github.com/sirupsen/logrus"
)

type GLResp struct {
	ProofGL       *recursion.Witness `cbor:"proof_gl"`
	LPPCommitment field.Element      `cbor:"lpp_commit"`
}

func RunGL(cfg *config.Config, witnessGLPath string, proofGLPath string) error {

	logrus.Infof("Initating the GL-prover from witnessGL file path %v", witnessGLPath)

	witnessGL := &distributed.ModuleWitnessGL{}
	if err := serialization.LoadFromDisk(witnessGLPath, witnessGL, true); err != nil {
		return fmt.Errorf("could not load witness: %w", err)
	}

	logrus.Infof("Running the GL-prover for witness module name=%s at index=%d", witnessGL.ModuleName, witnessGL.ModuleIndex)

	// TODO: Loads static prover asset. This call will go away after impl.
	// `mmap` optimization in the respective GL worker controller
	compiledGL, err := zkevm.LoadCompiledGL(cfg, witnessGL.ModuleName)
	if err != nil {
		return fmt.Errorf("could not load compiled GL: %w", err)
	}

	logrus.Infof("Loaded the compiled GL for witness module=%v at index=%d", witnessGL.ModuleName, witnessGL.ModuleIndex)

	run := compiledGL.ProveSegment(witnessGL)

	logrus.Infof("Finished running the GL-prover for witness module=%v at index=%d", witnessGL.ModuleName, witnessGL.ModuleIndex)

	var (
		_proofGL   = recursion.ExtractWitness(run)
		_lppCommit = distributed.GetLppCommitmentFromRuntime(run)
	)

	logrus.Infof("Extracted the witness for witness module=%v at index=%d", witnessGL.ModuleName, witnessGL.ModuleIndex)

	glResp := &GLResp{
		ProofGL:       &_proofGL,
		LPPCommitment: _lppCommit,
	}

	if err := serialization.StoreToDisk(proofGLPath, *glResp, true); err != nil {
		return fmt.Errorf("could not store GL proof: %w", err)
	}

	logrus.Infof("Stored the GL proof and LPP commitment to path=%s", proofGLPath)

	// Free memory immediately for next GL job
	glResp = nil

	return nil
}
