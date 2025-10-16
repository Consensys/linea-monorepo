package limitless

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/consensys/linea-monorepo/prover/backend/execution"
	"github.com/consensys/linea-monorepo/prover/circuits"
	execCirc "github.com/consensys/linea-monorepo/prover/circuits/execution"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/serialization"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/exit"
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
	numConcurrentWitnessWritingGoroutines = 20
	// numConcurrentSubProverJobs governs the number of concurrent sub-prover
	// jobs.
	numConcurrentSubProverJobs = 4
)

// Prove function for the Assest struct
func Prove(cfg *config.Config, req *execution.Request) (*execution.Response, error) {

	// Set MonitorParams before any proving happens
	profiling.SetMonitorParams(cfg)

	// Setting the issue handler to exit on unsatisfied constraint but not limit
	// overflow.
	exit.SetIssueHandlingMode(exit.ExitOnUnsatisfiedConstraint)

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

	if cfg.Execution.LimitlessWithDebug {
		limitlessZkEVM := zkevm.NewLimitlessDebugZkEVM(cfg)
		limitlessZkEVM.RunDebug(cfg, witness.ZkEVM)
		// The return of "out" is to avoid panics later on in the process.
		return &out, nil
	}

	// -- 1. Launch bootstrapper

	logrus.Info("Starting to run the bootstrapper")
	var (
		bootstrapperResp = RunBootstrapper(cfg, witness.ZkEVM)
		numGL, numLPP    = bootstrapperResp.numGL, bootstrapperResp.numLPP
		mt               = bootstrapperResp.mt
		cong             = bootstrapperResp.cong
	)
	logrus.Infof("Finished running the bootstrapper, generated %d GL modules and %d LPP modules", numGL, numLPP)

	var (
		proofGLs = make([]distributed.SegmentProof, numGL)

		glErrGroup  = &errgroup.Group{}
		lppErrGroup = &errgroup.Group{}

		contextGL, cancelGL = context.WithCancel(context.Background())

		err error

		totalProofs = numGL + numLPP

		// Conglomeration proof pipeline setup: have a buffered channel of capacity 2
		proofStream = make(chan distributed.SegmentProof, 2)
		resultCh    = make(chan distributed.SegmentProof, 1)
	)

	// -- 2. Launch background hierarchical reduction pipeline to recursively conglomerate as 2 or more proofs come in
	go func() {
		result := RunConglomerationHierarchical(&mt, cong, proofStream, totalProofs)
		resultCh <- result
	}()

	// -- 3. Launch GL proof jobs, sending each result to proofStream
	glErrGroup.SetLimit(numConcurrentSubProverJobs)
	for i := 0; i < numGL; i++ {

		i := i // local copy for closure
		glErrGroup.Go(func() error {

			if contextGL.Err() != nil {
				return nil
			}

			var (
				jobErr  error
				proofGL *distributed.SegmentProof
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
				proofGL, err = RunGL(cfg, i)
				if err != nil {
					jobErr = fmt.Errorf("could not run GL prover for witness index=%v: %w", i, err)
				}
			}()

			if jobErr != nil {
				cancelGL()
				return jobErr
			}

			proofGLs[i] = *proofGL
			// Send to pipeline as soon as ready
			proofStream <- *proofGL
			return nil
		})
	}

	if err := glErrGroup.Wait(); err != nil {
		close(proofStream)
		return nil, err
	}

	// This context cancellation is here to ensure the context is wiped-out in
	// every branches.
	cancelGL()
	if err != nil {
		return nil, err
	}

	var (
		sharedRandomness      = distributed.GetSharedRandomnessFromSegmentProofs(proofGLs)
		contextLPP, cancelLPP = context.WithCancel(context.Background())
	)

	// -- 4. Launch LPP proof jobs, streaming each proof to pipeline
	lppErrGroup.SetLimit(numConcurrentSubProverJobs)

	for i := 0; i < numLPP; i++ {

		i := i // local copy for closure
		lppErrGroup.Go(func() error {

			if contextLPP.Err() != nil {
				return nil
			}

			var (
				jobErr   error
				proofLPP *distributed.SegmentProof
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

			proofStream <- *proofLPP
			return nil
		})
	}

	if err := lppErrGroup.Wait(); err != nil {
		close(proofStream)
		return nil, err
	}

	// This context cancellation is here to ensure the context is wiped-out in
	// every branches.
	cancelLPP()
	if err != nil {
		return nil, err
	}

	var (
		setup       circuits.Setup
		errSetup    error
		chSetupDone = make(chan struct{})
	)

	// Start loading the setup before starting the conglomeration so that it is
	// ready when we need it.
	go func() {
		logrus.Infof("Loading setup - circuitID: %s", circuits.ExecutionLimitlessCircuitID)
		setup, errSetup = circuits.LoadSetup(cfg, circuits.ExecutionLimitlessCircuitID)
		close(chSetupDone)
	}()

	//-- 8. Wait for hierarchical aggregation to finish (final proof root)
	congProofRoot := <-resultCh

	// -- 9. We wait for setup to be loaded. It should already be loaded normally so we
	// do not expect to actually wait here.
	<-chSetupDone
	if errSetup != nil {
		utils.Panic("could not load setup: %v", errSetup)
	}

	out.Proof = execCirc.MakeProof(
		&cfg.TracesLimits,
		setup,
		cong.HierarchicalConglomeration.Wiop,
		congProofRoot,
		*witness.FuncInp,
	)

	out.VerifyingKeyShaSum = setup.VerifyingKeyDigest()

	return &out, nil
}

type bootstrapResp struct {
	numGL  int
	numLPP int
	mt     distributed.VerificationKeyMerkleTree
	cong   *distributed.RecursedSegmentCompilation
}

// RunBootstrapper loads the assets required to run the bootstrapper and runs it,
// the function then performs the module segmentation and saves each module
// witness in the /tmp directory.
func RunBootstrapper(cfg *config.Config, zkevmWitness *zkevm.Witness,
) *bootstrapResp {

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

	// The function initially attempt to run the bootstrapper directly and will
	// catch "limit-overflow" panic msgs. When they happen, we reattempt running
	// the bootstrapper with higher and higher limits until it works.
	var (
		scalingFactor = 1
		runtimeBoot   *wizard.ProverRuntime
	)

	for runtimeBoot == nil {

		logrus.Infof("Trying to bootstrap with a scaling of %v\n", scalingFactor)

		func() {

			// Since the [exit] package is configured to only send panic messages
			// on overflow. The overflows are catchable.
			defer func() {
				if err := recover(); err != nil {
					oFReport, isOF := err.(exit.LimitOverflowReport)
					if isOF {
						extra := utils.DivCeil(oFReport.RequestedSize, oFReport.Limit)
						scalingFactor *= utils.NextPowerOfTwo(extra)
						return
					}

					panic(err)
				}
			}()

			if scalingFactor == 1 {
				logrus.Infof("Running bootstrapper")
				runtimeBoot = wizard.RunProver(
					assets.DistWizard.Bootstrapper,
					assets.Zkevm.GetMainProverStep(zkevmWitness),
				)
				return
			}

			scaledUpBootstrapper, scaledUpZkEVM := zkevm.GetScaledUpBootstrapper(
				cfg, assets.DistWizard.Disc, scalingFactor,
			)

			runtimeBoot = wizard.RunProver(
				scaledUpBootstrapper,
				scaledUpZkEVM.GetMainProverStep(zkevmWitness),
			)
		}()
	}

	// This frees the memory from the assets that are no longer needed. We don't
	// need to do that for the module GLs and LPPs because they are thrown away
	// when the current function returns.

	var (
		mt   = assets.DistWizard.VerificationKeyMerkleTree
		cong = assets.DistWizard.CompiledConglomeration
	)

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
		assets.DistWizard.VerificationKeyMerkleTree.GetRoot(),
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

	return &bootstrapResp{
		mt:     mt,
		cong:   cong,
		numGL:  len(witnessGLs),
		numLPP: len(witnessLPPs),
	}
}

// RunGL runs the GL prover for the provided witness index
func RunGL(cfg *config.Config, witnessIndex int) (proofGL *distributed.SegmentProof, err error) {

	logrus.Infof("Running the GL-prover for witness index=%v", witnessIndex)

	witness := &distributed.ModuleWitnessGL{}
	witnessFilePath := witnessDir + "/witness-GL-" + strconv.Itoa(witnessIndex)
	if err := serialization.LoadFromDisk(witnessFilePath, witness, true); err != nil {
		return nil, err
	}

	logrus.Infof("Loaded the witness for witness index=%v, module=%v", witnessIndex, witness.ModuleName)

	compiledGL, err := zkevm.LoadCompiledGL(cfg, witness.ModuleName)
	if err != nil {
		return nil, fmt.Errorf("could not load compiled GL: %w", err)
	}

	logrus.Infof("Loaded the compiled GL for witness index=%v, module=%v", witnessIndex, witness.ModuleName)

	_proofGL := compiledGL.ProveSegment(witness)

	logrus.Infof("Finished running the GL-prover for witness index=%v, module=%v", witnessIndex, witness.ModuleName)

	return &_proofGL, nil
}

// RunLPP runs the LPP prover for the provided witness index
func RunLPP(cfg *config.Config, witnessIndex int, sharedRandomness field.Element) (proofLPP *distributed.SegmentProof, err error) {

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

	_proofLPP := compiledLPP.ProveSegment(witness)

	logrus.Infof("Finished running the LPP-prover for witness index=%v, module=%v", witnessIndex, witness.ModuleName)

	return &_proofLPP, nil
}

func RunConglomerationHierarchical(mt *distributed.VerificationKeyMerkleTree, cong *distributed.RecursedSegmentCompilation,
	proofStream <-chan distributed.SegmentProof, totalProofs int) distributed.SegmentProof {

	resultCh := make(chan distributed.SegmentProof, 1)
	// Aggregation workers:
	go func() {
		// Stack is a slice for channel-based pairing logic
		var stack []distributed.SegmentProof
		proofsReceived := 0

		// Continuously pull proofs from the stream or from aggregation itself
		for proofsReceived < totalProofs || len(stack) > 1 {
			select {
			// Receive new proof from main proof stream
			case proof := <-proofStream:
				logrus.Infof("Received proof for module index:%d and segment index:%d", proof.ModuleIndex, proof.SegmentIndex)
				stack = append(stack, proof)
				proofsReceived++
			// Whenever 2 proofs are ready, conglomerate them
			default:
				// Keep aggregating while there are at least 2
				for len(stack) >= 2 {
					_proof1 := stack[len(stack)-1]
					_proof2 := stack[len(stack)-2]
					stack = stack[:len(stack)-2]
					// Conglomerate in background
					aggregated := cong.ProveSegment(&distributed.ModuleWitnessConglo{
						SegmentProofs:             []distributed.SegmentProof{_proof1, _proof2},
						VerificationKeyMerkleTree: *mt,
					})
					// Put back into stack for further aggregation
					stack = append(stack, aggregated)
				}
			}
		}
		// The last item is the final proof
		if len(stack) == 1 {
			resultCh <- stack[0]
		} else {
			panic("RunConglomerationHierarchical: stack does not have 1 final proof")
		}
	}()
	// Wait for result
	final := <-resultCh
	close(resultCh)
	return final
}

/*
// RunConglomeration runs the conglomeration prover for the provided subproofs
func RunConglomeration(cfg *config.Config, proofGLs, proofLPPs []recursion.Witness) (proof wizard.Proof, congloWIOP *wizard.CompiledIOP, err error) {

	logrus.Infof("Running the conglomeration-prover")

	cong, err := zkevm.LoadConglomeration(cfg)
	if err != nil {
		return wizard.Proof{}, nil, fmt.Errorf("could not load compiled conglomeration: %w", err)
	}

	logrus.Infof("Loaded the compiled conglomeration")

	proof = cong.Prove(proofGLs, proofLPPs)

	logrus.Infof("Finished running the conglomeration-prover")
	run, err := wizard.VerifyWithRuntime(cong.Wiop, proof)
	if err != nil {
		zkevm.LogPublicInputs(run)
		exit.OnUnsatisfiedConstraints(err)
	}

	logrus.Infof("Successfully sanity-checked the conglomerator")

	return proof, cong.Wiop, nil
} */
