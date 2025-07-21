package limitless

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strconv"

	"github.com/consensys/linea-monorepo/prover/backend/execution"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/recursion"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/serialization"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/profiling"
	"github.com/consensys/linea-monorepo/prover/zkevm"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

var (
	witnessDir = "/tmp/witnesses"
	// numConcurrentWitnessWritingGoroutines governs the goroutine serializing,
	// compressing and writing the  witness. The writing part is also controlled
	// by a semaphore on top of this.
	numConcurrentWitnessWritingGoroutines = runtime.NumCPU()
	// numConcurrentSubProverJobs governs the number of concurrent sub-prover
	// jobs.
	numConcurrentSubProverJobs = 6
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
		out           = execution.CraftProverOutput(cfg, req)
		witness       = execution.NewWitness(cfg, req, &out)
		numGL, numLPP = 0, 0
	)

	if cfg.Execution.LimitlessWithDebug {
		limitlessZkEVM := zkevm.NewLimitlessDebugZkEVM(cfg)
		limitlessZkEVM.RunDebug(witness.ZkEVM)
		// The return of "out" is to avoid panics later on in the process.
		return &out, nil
	}

	logrus.Info("Starting to run the bootstrapper")

	numGL, numLPP = RunBootstrapper(cfg, witness.ZkEVM)

	logrus.Infof("Finished running the bootstrapper, generated %d GL modules and %d LPP modules", numGL, numLPP)

	var (
		proofGLs            []recursion.Witness
		proofLPPs           []recursion.Witness
		lppCommitments      []field.Element
		errGroup            = &errgroup.Group{}
		contextGL, cancelGL = context.WithCancel(context.Background())
	)

	errGroup.SetLimit(numConcurrentSubProverJobs)

	for i := 0; i < numGL; i++ {

		errGroup.Go(func() error {

			if contextGL.Err() != nil {
				return nil
			}

			var (
				jobErr        error
				proofGL       *recursion.Witness
				lppCommitment field.Element
			)

			// RunGL may panic and therefore exit the goroutine without returning
			// an error. Therefore it has to be wrapped in a recoverable function
			// call so that the error handling mechanism work. If an error occur
			// it is assigned to jobErr so that we can check it once the wrapper
			// function returns.
			func() {

				defer func() {
					if err := recover(); err != nil {
						jobErr = fmt.Errorf("GL prover for witness index=%v panicked: %v", i, err)
					}
				}()

				var err error
				proofGL, lppCommitment, err = RunGL(cfg, i)
				if err != nil {
					jobErr = fmt.Errorf("could not run GL prover for witness index=%v: %w", i, err)
				}
			}()

			if jobErr != nil {
				cancelGL()
				return jobErr
			}

			proofGLs = append(proofGLs, *proofGL)
			lppCommitments = append(lppCommitments, lppCommitment)
			return nil
		})
	}

	err := errGroup.Wait()
	// This context cancellation is here to ensure the context is wiped-out in
	// every branches.
	cancelGL()
	if err != nil {
		return nil, err
	}

	var (
		sharedRandomness      = distributed.GetSharedRandomness(lppCommitments)
		contextLPP, cancelLPP = context.WithCancel(context.Background())
	)

	for i := 0; i < numLPP; i++ {

		errGroup.Go(func() error {

			if contextLPP.Err() != nil {
				return nil
			}

			var (
				jobErr   error
				proofLPP *recursion.Witness
			)

			// RunLPP may panic and therefore exit the goroutine without returning
			// an error. Therefore it has to be wrapped in a recoverable function
			// call so that the error handling mechanism work. If an error occur
			// it is assigned to jobErr so that we can check it once the wrapper
			// function returns.
			func() {

				defer func() {
					if err := recover(); err != nil {
						jobErr = fmt.Errorf("LPP prover for witness index=%v panicked: %v", i, err)
					}
				}()

				var err error
				proofLPP, err = RunLPP(cfg, i, sharedRandomness)
				if err != nil {
					jobErr = fmt.Errorf("could not run LPP prover for witness index=%v: %w", i, err)
				}
			}()

			if jobErr != nil {
				cancelLPP()
				return fmt.Errorf("could not run LPP prover for witness index=%v: %w", i, jobErr)
			}

			proofLPPs = append(proofLPPs, *proofLPP)
			return nil
		})
	}

	err = errGroup.Wait()
	// This context cancellation is here to ensure the context is wiped-out in
	// every branches.
	cancelLPP()
	if err != nil {
		return nil, err
	}

	_, err = RunConglomeration(cfg, proofGLs, proofLPPs)
	if err != nil {
		return nil, fmt.Errorf("could not run conglomeration prover: %w", err)
	}

	return &out, nil
}

// RunBootstrapper loads the assets required to run the bootstrapper and runs it,
// the function then performs the module segmentation and saves each module
// witness in the /tmp directory.
func RunBootstrapper(cfg *config.Config, zkevmWitness *zkevm.Witness,
) (numWitnessGL, numWitnessLPP int) {

	logrus.Infof("Loading bootstrapper and zkevm")
	assets := &zkevm.LimitlessZkEVM{}
	if err := assets.LoadBootstrapper(cfg); err != nil || assets.DistWizard.Bootstrapper == nil {
		utils.Panic("could not load bootstrapper: %v", err)
	}

	if err := assets.LoadZkEVM(cfg); err != nil || assets.Zkevm == nil {
		utils.Panic("could not load zkevm: %v", err)
	}

	if err := assets.LoadDisc(cfg); err != nil || assets.DistWizard.Disc == nil {
		utils.Panic("could not load disc: %v", err)
	}

	// The GL and LPP modules are loaded in the background immediately but we
	// only need them for the [distributed.SegmentRuntime] call.
	distDone := make(chan error)
	go func() {
		err := assets.LoadBlueprints(cfg)
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

	logrus.Infof("Segmenting the runtime")
	witnessGLs, witnessLPPs := distributed.SegmentRuntime(
		runtimeBoot,
		assets.DistWizard.Disc,
		assets.DistWizard.BlueprintGLs,
		assets.DistWizard.BlueprintLPPs,
	)

	logrus.Info("Saving the witnesses")

	eg := &errgroup.Group{}
	eg.SetLimit(numConcurrentWitnessWritingGoroutines)

	for i := range witnessGLs {

		// This saves the value of i in the closure to ensure that the right
		// value is passed. It should be obsolete with newer version of Go.
		i := i
		eg.Go(func() error {

			filePath := witnessDir + "/witness-GL-" + strconv.Itoa(i)
			if err := serialization.StoreToDisk(filePath, *witnessGLs[i], true); err != nil {
				return fmt.Errorf("could not save witnessGL: %v", err)
			}

			// This frees the memory from the witness that is no longer needed.
			witnessGLs[i] = nil
			return nil
		})

	}

	for i := range witnessLPPs {

		// This saves the value of i in the closure to ensure that the right
		// value is passed. It should be obsolete with newer version of Go.
		i := i

		eg.Go(func() error {

			filePath := witnessDir + "/witness-LPP-" + strconv.Itoa(i)
			if err := serialization.StoreToDisk(filePath, *witnessLPPs[i], true); err != nil {
				return fmt.Errorf("could not save witnessLPP: %v", err)
			}

			// This frees the memory from the witness that is no longer needed.
			witnessLPPs[i] = nil
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		utils.Panic("could not save witnesses: %v", err)
	}

	return len(witnessGLs), len(witnessLPPs)
}

// RunGL runs the GL prover for the provided witness index
func RunGL(cfg *config.Config, witnessIndex int) (proofGL *recursion.Witness, lppCommitments field.Element, err error) {

	logrus.Infof("Running the GL-prover for witness index=%v", witnessIndex)

	witness := &distributed.ModuleWitnessGL{}
	witnessFilePath := witnessDir + "/witness-GL-" + strconv.Itoa(witnessIndex)
	if err := serialization.LoadFromDisk(witnessFilePath, witness, true); err != nil {
		return nil, field.Element{}, err
	}

	logrus.Infof("Loaded the witness for witness index=%v, module=%v", witnessIndex, witness.ModuleName)

	compiledGL, err := zkevm.LoadCompiledGL(cfg, witness.ModuleName)
	if err != nil {
		return nil, field.Element{}, fmt.Errorf("could not load compiled GL: %w", err)
	}

	logrus.Infof("Loaded the compiled GL for witness index=%v, module=%v", witnessIndex, witness.ModuleName)

	run := compiledGL.ProveSegment(witness)

	logrus.Infof("Finished running the GL-prover for witness index=%v, module=%v", witnessIndex, witness.ModuleName)

	_proofGL := recursion.ExtractWitness(run)

	logrus.Infof("Extracted the witness for witness index=%v, module=%v", witnessIndex, witness.ModuleName)

	return &_proofGL, distributed.GetLppCommitmentFromRuntime(run), nil
}

// RunLPP runs the LPP prover for the provided witness index
func RunLPP(cfg *config.Config, witnessIndex int, sharedRandomness field.Element) (proofLPP *recursion.Witness, err error) {

	logrus.Infof("Running the LPP-prover for witness index=%v", witnessIndex)

	witness := &distributed.ModuleWitnessLPP{}
	witnessFilePath := witnessDir + "/witness-LPP-" + strconv.Itoa(witnessIndex)
	if err := serialization.LoadFromDisk(witnessFilePath, witness, true); err != nil {
		return nil, err
	}

	witness.InitialFiatShamirState = sharedRandomness

	logrus.Infof("Loaded the witness for witness index=%v, module=%v", witnessIndex, witness.ModuleName)

	compiledLPP, err := zkevm.LoadCompiledLPP(cfg, witness.ModuleName)
	if err != nil {
		return nil, fmt.Errorf("could not load compiled LPP: %w", err)
	}

	logrus.Infof("Loaded the compiled LPP for witness index=%v, module=%v", witnessIndex, witness.ModuleName)

	run := compiledLPP.ProveSegment(witness)

	logrus.Infof("Finished running the LPP-prover for witness index=%v, module=%v", witnessIndex, witness.ModuleName)

	_proofLPP := recursion.ExtractWitness(run)

	logrus.Infof("Extracted the witness for witness index=%v, module=%v", witnessIndex, witness.ModuleName)

	return &_proofLPP, nil
}

// RunConglomeration runs the conglomeration prover for the provided subproofs
func RunConglomeration(cfg *config.Config, proofGLs, proofLPPs []recursion.Witness) (proof wizard.Proof, err error) {

	logrus.Infof("Running the conglomeration-prover")

	cong, err := zkevm.LoadConglomeration(cfg)
	if err != nil {
		return wizard.Proof{}, fmt.Errorf("could not load compiled conglomeration: %w", err)
	}

	logrus.Infof("Loaded the compiled conglomeration")

	proof = cong.Prove(proofGLs, proofLPPs)

	logrus.Infof("Finished running the conglomeration-prover")

	return proof, nil
}
