package distributed

import (
	"fmt"
	"os"
	"path"

	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/recursion"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/serialization"
	"github.com/consensys/linea-monorepo/prover/zkevm"
	"github.com/sirupsen/logrus"
)

type GLRequest struct {
	WitnessGLPath string
	StartBlock    string
	EndBlock      string
	SegID         int
}

type GLResponse struct {
	GLProofPath        string `json:"gl_proof_path"`
	LPPCommitmeentPath string `json:"lpp_commit_path"`
}

func RunGL(cfg *config.Config, req *GLRequest) (*GLResponse, error) {

	logrus.Infof("Initating the GL-prover from witnessGL file path %v", req.WitnessGLPath)

	witnessGL := &distributed.ModuleWitnessGL{}
	if err := serialization.LoadFromDisk(req.WitnessGLPath, witnessGL, true); err != nil {
		return nil, fmt.Errorf("could not load witness: %w", err)
	}

	var (
		glProofFile   = fmt.Sprintf("%s-%s-seg-%d-mod-%d-gl-proof.bin", req.StartBlock, req.EndBlock, req.SegID, witnessGL.ModuleIndex)
		proofGLPath   = path.Join(config.SubProofsGLDirPrefix, string(witnessGL.ModuleName), glProofFile)
		LPPCommitFile = fmt.Sprintf("%s-%s-seg-%d-mod-%d-gl-commit.bin", req.StartBlock, req.EndBlock, req.SegID, witnessGL.ModuleIndex)
		LPPCommitPath = path.Join(config.LPPCommitPrefix, string(witnessGL.ModuleName), LPPCommitFile)
	)

	// Incase the prev. process was interrupted, we clear the previous corrupted files (if it exists)
	_ = os.Remove(proofGLPath)
	_ = os.Remove(LPPCommitPath)

	logrus.Infof("Running the GL-prover for witness module name=%s at index=%d", witnessGL.ModuleName, witnessGL.ModuleIndex)

	// TODO: Loads static prover asset. This call will go away after impl.
	// `mmap` optimization in the respective GL worker controller
	compiledGL, err := zkevm.LoadCompiledGL(cfg, witnessGL.ModuleName)
	if err != nil {
		return nil, fmt.Errorf("could not load compiled GL: %w", err)
	}

	logrus.Infof("Loaded the compiled GL for witness module=%v at index=%d", witnessGL.ModuleName, witnessGL.ModuleIndex)

	run := compiledGL.ProveSegment(witnessGL)

	logrus.Infof("Finished running the GL-prover for witness module=%v at index=%d", witnessGL.ModuleName, witnessGL.ModuleIndex)

	_proofGL := recursion.ExtractWitness(run)
	if err := serialization.StoreToDisk(proofGLPath, _proofGL, true); err != nil {
		return nil, fmt.Errorf("could not store GL proof: %w", err)
	}

	logrus.Infof("Generated GL proof for witness module=%v at index=%d and stored to disk", witnessGL.ModuleName, witnessGL.ModuleIndex)

	_lppCommit := distributed.GetLppCommitmentFromRuntime(run)
	if err := serialization.StoreToDisk(LPPCommitPath, _lppCommit, true); err != nil {
		return nil, fmt.Errorf("could not store GL proof: %w", err)
	}

	logrus.Infof("Generated LPP commitment for witness module=%v at index=%d and stored to disk", witnessGL.ModuleName, witnessGL.ModuleIndex)

	return &GLResponse{
		GLProofPath:        proofGLPath,
		LPPCommitmeentPath: LPPCommitPath,
	}, nil
}
