package distributed

import (
	"fmt"
	"os"
	"path"
	"runtime/debug"

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
	GLProofPath        string `json:"glProofPath"`
	LPPCommitmeentPath string `json:"lppCommitPath"`
}

func RunGL(cfg *config.Config, req *GLRequest) (glResp *GLResponse, err error) {

	logrus.Infof("Starting GL-prover from witnessGL file path %v", req.WitnessGLPath)

	// Recover wrapper for panics
	defer func() {
		if r := recover(); r != nil {
			logrus.Errorf("[PANIC] GL prover crashed for witness %s: %v\n%s", req.WitnessGLPath, r, debug.Stack())
			err = fmt.Errorf("panic in GL prover: %v", r)
		}
	}()

	witnessGL := &distributed.ModuleWitnessGL{}
	if err := serialization.LoadFromDisk(req.WitnessGLPath, witnessGL, true); err != nil {
		return nil, fmt.Errorf("could not load witness: %w", err)
	}

	var (
		glProofFile   = fmt.Sprintf("%s-%s-seg-%d-mod-%d-gl-proof.bin", req.StartBlock, req.EndBlock, req.SegID, witnessGL.ModuleIndex)
		proofGLPath   = path.Join(cfg.LimitlessParams.SubproofsDir, "GL", string(witnessGL.ModuleName), glProofFile)
		LPPCommitFile = fmt.Sprintf("%s-%s-seg-%d-mod-%d-gl-lpp-commit.bin", req.StartBlock, req.EndBlock, req.SegID, witnessGL.ModuleIndex)
		LPPCommitPath = path.Join(cfg.LimitlessParams.CommitsDir, string(witnessGL.ModuleName), LPPCommitFile)
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

	glResp = &GLResponse{
		GLProofPath:        proofGLPath,
		LPPCommitmeentPath: LPPCommitPath,
	}
	return glResp, nil
}
