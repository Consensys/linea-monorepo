package limitless

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strconv"

	"github.com/consensys/linea-monorepo/prover/backend/execution"
	"github.com/consensys/linea-monorepo/prover/circuits"
	execCirc "github.com/consensys/linea-monorepo/prover/circuits/execution"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/serde"
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
	// jobs for the limitless prover. With ~35GB peak per job, 20 concurrent
	// jobs uses ~700GB peak which fits comfortably in a 1.5TB machine.
	numConcurrentSubProverJobs = 20
	// numConcurrentCongloMerges governs the number of concurrent pairwise
	// conglomeration merges. Each merge is lighter than a GL/LPP proof (~2.4min
	// with ~1.3min initial + ~1.1min recursion), so we can run more in parallel.
	numConcurrentCongloMerges = 10
)

// Prove function for the Assest struct
func Prove(cfg *config.Config, req *execution.Request) (*execution.Response, error) {

	// Set MonitorParams before any proving happens
	profiling.SetMonitorParams(cfg)

	// Setting the issue handler to exit on unsatisfied constraint and missing trace file,
	// but not limit overflow.
	exit.SetIssueHandlingMode(exit.ExitOnUnsatisfiedConstraint | exit.ExitOnMissingTraceFile)

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

	mt, err := zkevm.LoadVerificationKeyMerkleTree(cfg)
	if err != nil {
		return nil, fmt.Errorf("could not load verification key merkle tree: %w", err)
	}

	var (
		numGL, numLPP = RunBootstrapper(cfg, witness.ZkEVM, mt.GetRoot())
	)
	logrus.Infof("Finished running the bootstrapper, generated %d GL modules and %d LPP modules", numGL, numLPP)

	// Use a parent context for the whole proving flow
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Final Conglomeration outcome
	type congResult struct {
		proof *distributed.SegmentProof
		err   error
	}

	var (
		totalProofs = numGL + numLPP

		glPreloads     = make([]*glPreload, numGL)
		glPreloadGroup = &errgroup.Group{}
		proofGLs       = make([]*distributed.SegmentProof, numGL)
		glProveGroup   = &errgroup.Group{}

		lppPreloads     = make([]*lppPreload, numLPP)
		lppPreloadGroup = &errgroup.Group{}
		lppProveGroup   = &errgroup.Group{}

		// Conglomeration proof pipeline setup: have a buffered channel proofStream
		// with capacity large enough so producers won't block
		proofStream = make(chan *distributed.SegmentProof, totalProofs)
		resultCh    = make(chan congResult, 1)

		cong *distributed.RecursedSegmentCompilation
	)

	// Safety cleanup: close any GL preload closers not yet closed (e.g., on error paths)
	defer func() {
		glPreloadGroup.Wait() //nolint:errcheck // ensure preloads finish before closing
		for i, p := range glPreloads {
			if p != nil && p.closer != nil {
				p.closer.Close()
				glPreloads[i] = nil
			}
		}
	}()

	// Safety cleanup: close any LPP preload closers not yet closed (e.g., on error paths)
	defer func() {
		lppPreloadGroup.Wait() //nolint:errcheck // ensure preloads finish before closing
		for i, p := range lppPreloads {
			if p != nil && p.closer != nil {
				p.closer.Close()
				lppPreloads[i] = nil
			}
		}
	}()

	// -- 2. Launch background hierarchical reduction pipeline to recursively conglomerate as 2 or more
	// proofs come in. It will exit when it collects `totalProofs` or when ctx is cancelled.
	go func() {
		var err error
		cong, err = zkevm.LoadCompiledConglomeration(cfg)
		if err != nil || cong == nil {
			panic(fmt.Errorf("could not load compiled conglomeration: %w", err))
		}
		logrus.Infoln("Succesfully loaded the compiled conglomeration and starting to run hierarchical conglomeration")
		proof, err := RunConglomerationHierarchical(ctx, mt, cong, proofStream, totalProofs)
		resultCh <- congResult{proof: proof, err: err}
	}()

	logrus.Infof("Using %d concurrent sub-prover jobs for limitless proving", numConcurrentSubProverJobs)

	// Start loading the outer circuit setup early, in parallel with GL/LPP proving.
	// This avoids a sequential load after conglomeration finishes.
	var (
		setup       circuits.Setup
		errSetup    error
		chSetupDone = make(chan struct{})
	)
	go func() {
		logrus.Infof("Loading setup (early) - circuitID: %s", circuits.ExecutionLimitlessCircuitID)
		setup, errSetup = circuits.LoadSetup(cfg, circuits.ExecutionLimitlessCircuitID)
		close(chSetupDone)
	}()

	// -- 3. Pre-load ALL GL and LPP witnesses and compiled circuits in parallel.
	// Preloading is I/O-bound; launching all goroutines immediately maximizes
	// overlap and avoids disk contention inside the CPU-bound proving pool.
	for i := 0; i < numGL; i++ {
		i := i
		glPreloadGroup.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			var (
				jobErr  error
				preload *glPreload
			)

			func() {
				defer func() {
					if r := recover(); r != nil {
						jobErr = fmt.Errorf("GL preload for witness index=%v panicked: %v", i, r)
						logrus.Error(jobErr)
						debug.PrintStack()
					}
				}()

				var err error
				preload, err = PreloadGL(cfg, i)
				if err != nil {
					jobErr = fmt.Errorf("could not preload GL for witness index=%v: %w", i, err)
				}
			}()

			if jobErr != nil {
				return jobErr
			}

			glPreloads[i] = preload
			return nil
		})
	}

	for i := 0; i < numLPP; i++ {
		i := i
		lppPreloadGroup.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			var (
				jobErr  error
				preload *lppPreload
			)

			func() {
				defer func() {
					if r := recover(); r != nil {
						jobErr = fmt.Errorf("LPP preload for witness index=%v panicked: %v", i, r)
						logrus.Error(jobErr)
						debug.PrintStack()
					}
				}()

				var err error
				preload, err = PreloadLPP(cfg, i)
				if err != nil {
					jobErr = fmt.Errorf("could not preload LPP for witness index=%v: %w", i, err)
				}
			}()

			if jobErr != nil {
				return jobErr
			}

			lppPreloads[i] = preload
			return nil
		})
	}

	// -- 3b. Wait for GL preloads, then launch GL proving with the CPU-bound pool.
	if err := glPreloadGroup.Wait(); err != nil {
		cancel()
		res := <-resultCh
		if res.err != nil {
			return nil, fmt.Errorf("GL preload error: %w (aggregator error: %v)", err, res.err)
		}
		return nil, fmt.Errorf("GL preload error: %w", err)
	}

	logrus.Infof("All %d GL witnesses and circuits pre-loaded, starting GL proving", numGL)

	glProveGroup.SetLimit(numConcurrentSubProverJobs)
	for i := 0; i < numGL; i++ {
		i := i
		glProveGroup.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			var (
				jobErr  error
				proofGL *distributed.SegmentProof
			)

			func() {
				defer func() {
					if r := recover(); r != nil {
						jobErr = fmt.Errorf("GL prover for witness index=%v panicked: %v", i, r)
						logrus.Error(jobErr)
						debug.PrintStack()
					}
				}()

				var err error
				proofGL, err = ProveGL(glPreloads[i])
				if err != nil {
					jobErr = fmt.Errorf("could not prove GL for witness index=%v: %w", i, err)
				}
			}()

			if jobErr != nil {
				return jobErr
			}

			// Store local copy for shared randomness computation
			proofGLs[i] = proofGL

			select {
			case proofStream <- proofGL:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})
	}

	// Wait for GLs
	if err := glProveGroup.Wait(); err != nil {
		cancel()
		res := <-resultCh
		if res.err != nil {
			return nil, fmt.Errorf("GL error: %w (aggregator error: %v)", err, res.err)
		}
		return nil, fmt.Errorf("GL error: %w", err)
	}

	var (
		//--4. Compute shared randomness after GL proofs succeeded
		sharedRandomness = distributed.GetSharedRandomnessFromSegmentProofs(proofGLs)
	)

	// -- 4b. Wait for LPP preloads to complete (likely already done during GL proving)
	if err := lppPreloadGroup.Wait(); err != nil {
		cancel()
		res := <-resultCh
		if res.err != nil {
			return nil, fmt.Errorf("LPP preload error: %w (aggregator error: %v)", err, res.err)
		}
		return nil, fmt.Errorf("LPP preload error: %w", err)
	}

	logrus.Infof("All %d LPP witnesses and circuits pre-loaded, starting LPP proving with sharedRandomness", numLPP)

	// -- 5. Launch LPP proof jobs using pre-loaded witnesses and circuits
	lppProveGroup.SetLimit(numConcurrentSubProverJobs)
	for i := 0; i < numLPP; i++ {
		i := i
		lppProveGroup.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			var (
				jobErr   error
				proofLPP *distributed.SegmentProof
			)

			func() {
				defer func() {
					if r := recover(); r != nil {
						jobErr = fmt.Errorf("LPP prover for witness index=%v panicked: %v", i, r)
						logrus.Error(jobErr)
						debug.PrintStack()
					}
				}()

				var err error
				proofLPP, err = ProveLPP(lppPreloads[i], sharedRandomness)
				if err != nil {
					jobErr = fmt.Errorf("could not prove LPP for witness index=%v: %w", i, err)
				}
			}()

			if jobErr != nil {
				return jobErr
			}

			select {
			case proofStream <- proofLPP:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})
	}

	// Wait for LPPs
	if err := lppProveGroup.Wait(); err != nil {
		cancel()
		res := <-resultCh
		if res.err != nil {
			return nil, fmt.Errorf("LPP error: %w (aggregator error: %v)", err, res.err)
		}
		return nil, fmt.Errorf("LPP error: %w", err)
	}

	// All producers finished successfully: close the proofStream so aggregator can finish
	close(proofStream)

	// Wait for final conglomeration proof
	res := <-resultCh
	if res.err != nil {
		return nil, fmt.Errorf("conglomeration failed: %w", res.err)
	}

	logrus.Infof("HIERARCHICAL CONGLOMERATION SUCCESSFUL!!!")

	congFinalproof := res.proof

	// -- 6. Wait for setup that was started loading early
	<-chSetupDone
	if errSetup != nil {
		utils.Panic("could not load setup: %v", errSetup)
	}

	out.Proof = execCirc.MakeProof(
		&cfg.TracesLimits,
		setup,
		cong.RecursionCompBLS,
		congFinalproof.GetOuterProofInput(),
		*witness.FuncInp,
		witness.ZkEVM.ExecData,
	)

	out.VerifyingKeyShaSum = setup.VerifyingKeyDigest()

	return &out, nil
}

// RunBootstrapper loads the assets required to run the bootstrapper and runs it,
// the function then performs the module segmentation and saves each module
// witness in the /tmp directory.
func RunBootstrapper(cfg *config.Config, zkevmWitness *zkevm.Witness, merkleTreeRoot field.Octuplet) (int, int) {

	logrus.Infof("Loading bootstrapper and zkevm")
	assets := &zkevm.LimitlessZkEVM{}
	if err := assets.LoadBootstrapper(cfg); err != nil || assets.DistWizard.Bootstrapper == nil {
		utils.Panic("could not load bootstrapper: %v", err)
	}

	if err := assets.LoadZkEVM(cfg); err != nil || assets.Zkevm == nil {
		utils.Panic("could not load zkevm: %v", err)
	}

	// Override the compatibility check setting from the runtime config, since
	// the serialized asset may have been built with a different value.
	assets.Zkevm.Arithmetization.Settings.IgnoreCompatibilityCheck = &cfg.Execution.IgnoreCompatibilityCheck

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
					false,
				)
				return
			}

			scaledUpBootstrapper, scaledUpZkEVM := zkevm.GetScaledUpBootstrapper(
				cfg, assets.DistWizard.Disc, scalingFactor,
			)

			runtimeBoot = wizard.RunProver(
				scaledUpBootstrapper,
				scaledUpZkEVM.GetMainProverStep(zkevmWitness),
				false,
			)
		}()
	}

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
		merkleTreeRoot,
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
			if err := serde.StoreToDisk(filePath, *witnessGLs[i], true); err != nil {
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
			if err := serde.StoreToDisk(filePath, *witnessLPPs[i], true); err != nil {
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

// glPreload holds pre-loaded GL witness and compiled circuit data.
// The closer must remain open until the proof is generated.
type glPreload struct {
	witnessIndex int
	witness      *distributed.ModuleWitnessGL
	compiled     *distributed.RecursedSegmentCompilation
	closer       io.Closer
}

// PreloadGL loads the GL witness and compiled circuit from disk without
// running the prover. This allows overlapping I/O with other work.
func PreloadGL(cfg *config.Config, witnessIndex int) (*glPreload, error) {
	logrus.Infof("Preloading GL witness and circuit for index=%v", witnessIndex)

	witness := &distributed.ModuleWitnessGL{}
	witnessFilePath := witnessDir + "/witness-GL-" + strconv.Itoa(witnessIndex)
	closer, err := serde.LoadFromDisk(witnessFilePath, witness, true)
	if err != nil {
		return nil, fmt.Errorf("could not load GL witness for index=%v: %w", witnessIndex, err)
	}

	logrus.Infof("Loaded GL witness for index=%v, module=%v", witnessIndex, witness.ModuleName)

	compiledGL, err := zkevm.LoadCompiledGL(cfg, witness.ModuleName)
	if err != nil {
		closer.Close()
		return nil, fmt.Errorf("could not load compiled GL for index=%v: %w", witnessIndex, err)
	}

	logrus.Infof("Preloaded GL circuit for index=%v, module=%v", witnessIndex, witness.ModuleName)

	return &glPreload{
		witnessIndex: witnessIndex,
		witness:      witness,
		compiled:     compiledGL,
		closer:       closer,
	}, nil
}

// ProveGL generates the GL proof using pre-loaded data.
func ProveGL(preload *glPreload) (*distributed.SegmentProof, error) {
	defer func() {
		preload.closer.Close()
		preload.closer = nil
	}()

	logrus.Infof("Running GL proof for index=%v, module=%v", preload.witnessIndex, preload.witness.ModuleName)

	proof := preload.compiled.ProveSegmentKoala(preload.witness).ClearRuntime()

	logrus.Infof("Finished GL proof for index=%v, module=%v", preload.witnessIndex, preload.witness.ModuleName)

	return proof, nil
}

// lppPreload holds pre-loaded LPP witness and compiled circuit data.
// The closer must remain open until the proof is generated.
type lppPreload struct {
	witnessIndex int
	witness      *distributed.ModuleWitnessLPP
	compiled     *distributed.RecursedSegmentCompilation
	closer       io.Closer
}

// PreloadLPP loads the LPP witness and compiled circuit from disk without
// requiring sharedRandomness. This allows overlapping I/O with GL proving.
func PreloadLPP(cfg *config.Config, witnessIndex int) (*lppPreload, error) {
	logrus.Infof("Preloading LPP witness and circuit for index=%v", witnessIndex)

	witness := &distributed.ModuleWitnessLPP{}
	witnessFilePath := witnessDir + "/witness-LPP-" + strconv.Itoa(witnessIndex)
	closer, err := serde.LoadFromDisk(witnessFilePath, witness, true)
	if err != nil {
		return nil, fmt.Errorf("could not load LPP witness for index=%v: %w", witnessIndex, err)
	}

	logrus.Infof("Loaded LPP witness for index=%v, module=%v", witnessIndex, witness.ModuleName)

	compiledLPP, err := zkevm.LoadCompiledLPP(cfg, witness.ModuleName)
	if err != nil {
		closer.Close()
		return nil, fmt.Errorf("could not load compiled LPP for index=%v: %w", witnessIndex, err)
	}

	logrus.Infof("Preloaded LPP circuit for index=%v, module=%v", witnessIndex, witness.ModuleName)

	return &lppPreload{
		witnessIndex: witnessIndex,
		witness:      witness,
		compiled:     compiledLPP,
		closer:       closer,
	}, nil
}

// ProveLPP generates the LPP proof using pre-loaded data and sharedRandomness.
func ProveLPP(preload *lppPreload, sharedRandomness field.Octuplet) (*distributed.SegmentProof, error) {
	defer func() {
		preload.closer.Close()
		preload.closer = nil
	}()

	preload.witness.InitialFiatShamirState = sharedRandomness

	logrus.Infof("Running LPP proof for index=%v, module=%v", preload.witnessIndex, preload.witness.ModuleName)

	proof := preload.compiled.ProveSegmentKoala(preload.witness).ClearRuntime()

	logrus.Infof("Finished LPP proof for index=%v, module=%v", preload.witnessIndex, preload.witness.ModuleName)

	return proof, nil
}

// RunConglomerationHierarchical aggregates segment proofs into a single proof
// using a parallel binary-tree reduction. Leaf proofs arrive via proofStream;
// pairs are merged concurrently up to numConcurrentCongloMerges workers. Each
// merge produces an intermediate proof that feeds back into the pool. The last
// of the totalProofs-1 merges uses BLS for the outer circuit.
func RunConglomerationHierarchical(ctx context.Context,
	mt *distributed.VerificationKeyMerkleTree,
	cong *distributed.RecursedSegmentCompilation,
	proofStream <-chan *distributed.SegmentProof, totalProofs int,
) (*distributed.SegmentProof, error) {

	if totalProofs <= 0 {
		return nil, fmt.Errorf("totalProofs must be > 0, got %d", totalProofs)
	}

	// Special case: single proof, no merges needed.
	if totalProofs == 1 {
		select {
		case p, ok := <-proofStream:
			if !ok {
				return nil, fmt.Errorf("proof stream closed before receiving the single proof")
			}
			return p, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	// Pool channel holds all proofs available for pairing: both leaf proofs
	// from proofStream and intermediate results from merge workers.
	pool := make(chan *distributed.SegmentProof, 2*totalProofs)

	// Feed leaf proofs from proofStream into the pool.
	go func() {
		for p := range proofStream {
			pool <- p
		}
	}()

	var (
		totalMerges = totalProofs - 1
		eg          = &errgroup.Group{}
		resultCh    = make(chan *distributed.SegmentProof, 1)
	)
	eg.SetLimit(numConcurrentCongloMerges)

	logrus.Infof("Starting parallel conglomeration: %d proofs, %d merges, up to %d concurrent",
		totalProofs, totalMerges, numConcurrentCongloMerges)

	// Dispatcher loop: for each of the N-1 merges, pull a pair from the pool
	// and launch a merge worker. eg.SetLimit blocks when the concurrency cap
	// is reached, providing natural back-pressure. The binary-tree structure
	// emerges organically: when many proofs are available, multiple pairs are
	// merged in parallel at each level; results feed back into the pool for
	// the next level.
	for i := 0; i < totalMerges; i++ {
		var p1, p2 *distributed.SegmentProof

		select {
		case p1 = <-pool:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
		select {
		case p2 = <-pool:
		case <-ctx.Done():
			return nil, ctx.Err()
		}

		isLast := (i == totalMerges-1)

		logrus.Infof("Conglomerating sub-proofs [merge %d/%d%s] for (proofType, moduleIdx, segmentIdx) = (%d, %d, %d) and (%d, %d, %d)",
			i+1, totalMerges, func() string {
				if isLast {
					return " FINAL/BLS"
				}
				return ""
			}(),
			p1.ProofType, p1.ModuleIndex, p1.SegmentIndex,
			p2.ProofType, p2.ModuleIndex, p2.SegmentIndex)

		eg.Go(func() error {
			// Check for cancellation before starting expensive work
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			var (
				merged *distributed.SegmentProof
				jobErr error
			)

			// Recover from panics in the prover
			func() {
				defer func() {
					if r := recover(); r != nil {
						jobErr = fmt.Errorf("conglo merge %d panicked: %v", i+1, r)
						logrus.Error(jobErr)
						debug.PrintStack()
					}
				}()

				wit := &distributed.ModuleWitnessConglo{
					SegmentProofs:             []distributed.SegmentProof{*p1.ClearRuntime(), *p2},
					VerificationKeyMerkleTree: *mt,
				}

				if isLast {
					merged = cong.ProveSegmentBLS(wit)
				} else {
					merged = cong.ProveSegmentKoala(wit)
				}
			}()

			if jobErr != nil {
				return jobErr
			}

			if isLast {
				resultCh <- merged
			} else {
				pool <- merged
			}
			return nil
		})
	}

	// Wait for all merge workers to complete
	if err := eg.Wait(); err != nil {
		return nil, fmt.Errorf("conglomeration failed: %w", err)
	}

	select {
	case finalProof := <-resultCh:
		logrus.Infoln("Successfully finished running parallel conglomeration prover.")
		return finalProof, nil
	default:
		return nil, fmt.Errorf("conglomeration completed but no final proof produced")
	}
}
