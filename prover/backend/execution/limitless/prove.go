package limitless

import (
	"fmt"
	"os"
	"strconv"

	"github.com/consensys/linea-monorepo/prover/backend/execution"
	"github.com/consensys/linea-monorepo/prover/backend/files"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/serialization"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/profiling"
	"github.com/consensys/linea-monorepo/prover/zkevm"
	"github.com/sirupsen/logrus"
)

const (
	witnessDir = "/tmp/witnesses"
)

// Prove function for the Assest struct
func Prove(cfg *config.Config, req *execution.Request) (*execution.Response, error) {

	// Set MonitorParams before any proving happens
	profiling.SetMonitorParams(cfg)

	// Clean up witness directory to be sure it is empty when we start the
	// process. This helps addressing the situation where a previous process
	// have been interrupted.
	os.RemoveAll(witnessDir)
	defer os.RemoveAll(witnessDir)

	// Setup execution witness and output response
	var (
		out     = execution.CraftProverOutput(cfg, req)
		witness = execution.NewWitness(cfg, req, &out)
	)

	logrus.Info("Starting to run the bootstrapper")
	_, _ = RunBootstrapper(cfg, witness.ZkEVM)

	logrus.Info("Finished running the bootstrapper")

	return &out, nil
}

// RunBootstrapper loads the assets required to run the bootstrapper and runs it,
// the function then performs the module segmentation and saves each module
// witness in the /tmp directory.
func RunBootstrapper(cfg *config.Config, zkevmWitness *zkevm.Witness,
) (numWitnessGL, numWitnessLPP int) {

	logrus.Infof("Loading bootstrapper and zkevm")
	assets := &zkevm.LimitlessZkEVM{}
	if err := assets.LoadBootstrapper(cfg); err != nil {
		utils.Panic("could not load bootstrapper: %v", err)
	}

	if err := assets.LoadZkEVM(cfg); err != nil {
		utils.Panic("could not load zkevm: %v", err)
	}

	// The GL and LPP modules are loaded in the background immediately but we
	// only need them for the [distributed.SegmentRuntime] call.
	distDone := make(chan error)
	go func() {
		err := assets.LoadModuleGLsAndLPPs(cfg)
		distDone <- err
	}()

	logrus.Infof("Running bootstrapper")
	runtimeBoot := wizard.RunProver(
		assets.DistWizard.Bootstrapper,
		assets.Zkevm.GetMainProverStep(zkevmWitness),
	)

	// This frees the memory from the assets that are no longer needed. We don't
	// need to do that for the module GLs and LPPs because they are thrown away
	// when the current function returns.
	assets.Zkevm = nil
	assets.DistWizard.Bootstrapper = nil

	if err := <-distDone; err != nil {
		utils.Panic("could not load GL and LPP modules: %v", err)
	}

	logrus.Info("Segmenting the runtime")
	witnessGLs, witnessLPPs := distributed.SegmentRuntime(runtimeBoot, assets.DistWizard)

	logrus.Info("Saving the witnesses")
	for i := range witnessGLs {

		filePath := witnessDir + "/witness-GL-" + strconv.Itoa(i)
		if err := writeToDisk(filePath, witnessGLs[i]); err != nil {
			utils.Panic("could not save witnessGL: %v", err)
		}

		// This frees the memory from the witness that is no longer needed.
		witnessGLs[i] = nil
	}

	for i := range witnessLPPs {

		filePath := witnessDir + "witness-LPP-" + strconv.Itoa(i)
		if err := writeToDisk(filePath, witnessLPPs[i]); err != nil {
			utils.Panic("could not save witnessLPP: %v", err)
		}

		// This frees the memory from the witness that is no longer needed.
		witnessLPPs[i] = nil
	}

	return len(witnessGLs), len(witnessLPPs)
}

// writeToDisk writes the provided assets to disk using the
// [serialization.Serialize] function.
func writeToDisk(filePath string, asset any) error {

	f := files.MustOverwrite(filePath)
	defer f.Close()

	buf, serr := serialization.Serialize(asset)
	if serr != nil {
		return fmt.Errorf("could not serialize %s: %w", filePath, serr)
	}

	if _, werr := f.Write(buf); werr != nil {
		return fmt.Errorf("could not write to file %s: %w", filePath, werr)
	}

	return nil
}
