package distributed

import (
	"fmt"
	"os"
	"path"
	"runtime/debug"

	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/serde"

	"github.com/consensys/linea-monorepo/prover/utils/profiling"
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

	// Set MonitorParams before GL-proving
	profiling.SetMonitorParams(cfg)

	logrus.Infof("Starting GL-prover from witnessGL file path %v", req.WitnessGLFile)

	// Recover wrapper for panics
	defer func() {
		if r := recover(); r != nil {
			logrus.Errorf("[PANIC] GL prover crashed for witness %s: %v", req.WitnessGLFile, r)
			debug.PrintStack()
			os.Exit(2)
		}
	}()

	witnessGL := &distributed.ModuleWitnessGL{}
	closer, err := serde.LoadFromDisk(req.WitnessGLFile, witnessGL, true)
	if err != nil {
		return fmt.Errorf("could not load witness: %w", err)
	}
	defer closer.Close()

	var (
		glProofFileName = fmt.Sprintf("%s-%s-seg-%d-mod-%d-gl-proof.bin", req.StartBlock, req.EndBlock, req.SegID, witnessGL.ModuleIndex)
		proofGLFile     = path.Join(cfg.ExecutionLimitless.SubproofsDir, "GL", string(witnessGL.ModuleName), config.RequestsFromSubDir, glProofFileName)
	)

	// Incase the prev. process was interrupted, we clear the previous corrupted files (if it exists)
	_ = os.Remove(proofGLFile)

	logrus.Infof("Running the GL-prover for witness module name=%s at index=%d", witnessGL.ModuleName, witnessGL.ModuleIndex)

	compiledGL, err := zkevm.LoadCompiledGL(cfg, witnessGL.ModuleName)
	if err != nil {
		return fmt.Errorf("could not load compiled GL: %w", err)
	}

	logrus.Infof("Loaded the compiled GL for witness module=%v at index=%d", witnessGL.ModuleName, witnessGL.ModuleIndex)

	proofGL := compiledGL.ProveSegment(witnessGL)

	logrus.Infof("Finished running the GL-prover for witness module=%v at index=%d", witnessGL.ModuleName, witnessGL.ModuleIndex)

	if err := serde.StoreToDisk(proofGLFile, *proofGL, true); err != nil {
		return fmt.Errorf("could not store GL proof: %w", err)
	}

	logrus.Infof("Stored GL proof for witness module=%v at index=%d  to disk", witnessGL.ModuleName, witnessGL.ModuleIndex)

	return nil
}
