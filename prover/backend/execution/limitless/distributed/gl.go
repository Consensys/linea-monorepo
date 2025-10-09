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
	WitnessGLFile string
	StartBlock    string
	EndBlock      string
	SegID         int
}

func RunGL(cfg *config.Config, req *GLRequest) error {

	logrus.Infof("Starting GL-prover from witnessGL file path %v", req.WitnessGLFile)

	// Recover wrapper for panics
	defer func() {
		if r := recover(); r != nil {
			logrus.Errorf("[PANIC] GL prover crashed for witness %s: %v\n%s", req.WitnessGLFile, r, debug.Stack())
			os.Exit(2)
		}
	}()

	witnessGL := &distributed.ModuleWitnessGL{}
	if err := serialization.LoadFromDisk(req.WitnessGLFile, witnessGL, true); err != nil {
		return fmt.Errorf("could not load witness: %w", err)
	}

	var (
		glProofFileName   = fmt.Sprintf("%s-%s-seg-%d-mod-%d-gl-proof.bin", req.StartBlock, req.EndBlock, req.SegID, witnessGL.ModuleIndex)
		proofGLFile       = path.Join(cfg.ExecutionLimitless.SubproofsDir, "GL", string(witnessGL.ModuleName), config.RequestsFromSubDir, glProofFileName)
		LPPCommitFileName = fmt.Sprintf("%s-%s-seg-%d-mod-%d-gl-lpp-commit.bin", req.StartBlock, req.EndBlock, req.SegID, witnessGL.ModuleIndex)
		LPPCommitFile     = path.Join(cfg.ExecutionLimitless.CommitsDir, string(witnessGL.ModuleName), config.RequestsFromSubDir, LPPCommitFileName)
	)

	// Incase the prev. process was interrupted, we clear the previous corrupted files (if it exists)
	_ = os.Remove(proofGLFile)
	_ = os.Remove(LPPCommitFile)

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

	// It is important to write the GL proof files first and then the LPP commitments. This is because
	// in the conglomeration prover, we have a watcher for all LPP commitments to get shared randomness and
	// hence the presence of LPP commitment files signal that GL proof files also exist. We want to make sure to
	// avoid partial writes - the case where only one file is written.
	_proofGL := recursion.ExtractWitness(run)
	if err := serialization.StoreToDisk(proofGLFile, _proofGL, true); err != nil {
		return fmt.Errorf("could not store GL proof: %w", err)
	}

	logrus.Infof("Generated GL proof for witness module=%v at index=%d and stored to disk", witnessGL.ModuleName, witnessGL.ModuleIndex)

	_lppCommit := distributed.GetLppCommitmentFromRuntime(run)
	if err := serialization.StoreToDisk(LPPCommitFile, _lppCommit, true); err != nil {
		return fmt.Errorf("could not store GL proof: %w", err)
	}

	logrus.Infof("Generated LPP commitment for witness module=%v at index=%d and stored to disk", witnessGL.ModuleName, witnessGL.ModuleIndex)

	return nil
}
