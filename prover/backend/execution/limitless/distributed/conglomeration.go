package distributed

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/consensys/linea-monorepo/prover/backend/execution"
	"github.com/consensys/linea-monorepo/prover/backend/files"
	"github.com/consensys/linea-monorepo/prover/circuits"
	execCirc "github.com/consensys/linea-monorepo/prover/circuits/execution"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/serde"

	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/profiling"
	"github.com/consensys/linea-monorepo/prover/zkevm"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	"github.com/consensys/linea-monorepo/prover/config"
)

func RunConglomerator(cfg *config.Config, req *Metadata) (execResp *execution.Response, err error) {

	// Set MonitorParams before conglomeration
	profiling.SetMonitorParams(cfg)

	logrus.Infof("Starting conglomerator for conflation request %s-%s", req.StartBlock, req.EndBlock)

	// Final Conglomeration outcome
	type congResult struct {
		proof *distributed.SegmentProof
		err   error
	}

	var (
		ctx, cancel = context.WithTimeout(context.Background(), time.Duration(cfg.ExecutionLimitless.Timeout)*time.Second)
		totalProofs = req.NumGL + req.NumLPP

		proofGLs = make([]*distributed.SegmentProof, req.NumGL)

		// Conglomeration proof pipeline setup: have a buffered channel proofStream
		// with capacity large enough so producers won't block
		cong        *distributed.RecursedSegmentCompilation
		proofStream = make(chan *distributed.SegmentProof, totalProofs)
		resultCh    = make(chan congResult, 1)

		// Err groups for GL/LPP sub-provers
		errGroup = errgroup.Group{}
	)

	// ---- 1) Generic defer for marking files on exit ----
	// Declare marking defer first so it runs after recovery (LIFO).
	defer func() {
		suf := files.OutcomeSuffix(err)
		allFiles := append([]string{}, req.GLProofFiles...)
		allFiles = append(allFiles, req.LPPProofFiles...)
		allFiles = append(allFiles, req.SharedRndFile)
		files.MarkAndMoveToDone(cfg, allFiles, suf)
	}()

	// Recover wrapper for panics
	defer func() {
		cancel()
		if r := recover(); r != nil {
			logrus.Errorf("[PANIC] Conglomerator crashed for conflation request %s:%s", req.StartBlock, req.EndBlock)
			debug.PrintStack()
		}
	}()

	// -- 1. Launch background hierarchical reduction pipeline to recursively conglomerate as 2 or more
	// proofs come in. It will exit when it collects `totalProofs` or when ctx is cancelled.
	go func() {
		cong, err = zkevm.LoadCompiledConglomeration(cfg)
		if err != nil || cong == nil {
			panic(fmt.Errorf("could not load compiled conglomeration: %w", err))
		}
		logrus.Infoln("Succesfully loaded the compiled conglomeration and starting to run hierarchical conglomeration")
		proof, err := runConglomerationHierarchical(ctx, cfg, cong, proofStream, totalProofs)
		resultCh <- congResult{proof: proof, err: err}
	}()

	// -- 2. Wait for all GL proofs to arrive from the GL sub-provers and
	// deserialize and send them to the conglomeration pipeline as they arrive
	errGroup.SetLimit(req.NumGL)
	for i := 0; i < req.NumGL; i++ {
		i := i // local copy
		errGroup.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			var (
				jobErr  error
				proofGL = &distributed.SegmentProof{}
			)

			jobErr = files.WaitForFileAtPath(ctx, req.GLProofFiles[i], time.Duration(cfg.ExecutionLimitless.PollInterval), true, fmt.Sprintf("Waiting for GL proof %d at path %s", i, req.GLProofFiles[i]))
			if jobErr != nil {
				logrus.Errorf("error waiting for GL proof %d: %s - Cancelling context", i, jobErr)
				return jobErr
			}

			// Once the GL-proof file has arrived - deserialize it and send it to the pipeline
			closer, err := serde.LoadFromDisk(req.GLProofFiles[i], proofGL, true)
			if err != nil {
				logrus.Errorf("error deserializing GL proof %d: %s - Cancelling context", i, err)
				return err
			}
			defer closer.Close()

			// Store local copy for shared randomness computation
			proofGLs[i] = proofGL

			// Safe send: if ctx cancelled, abort send
			select {
			case proofStream <- proofGL:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})
	}

	// Wait for GLs
	if err := errGroup.Wait(); err != nil {
		// Cancel overall flow (aggregator will observe ctx.Done)
		cancel()
		// Wait for aggregator to finish (so resultCh gets something)
		res := <-resultCh
		if res.err != nil {
			// aggregator returned an error (likely due to cancel); return original error
			return nil, fmt.Errorf("GL error: %w (aggregator error: %v)", err, res.err)
		}
		return nil, fmt.Errorf("GL error: %w", err)
	}

	//--4. Compute shared randomness after GL proofs succeeded and save it to disk
	sharedRandomness := distributed.GetSharedRandomnessFromSegmentProofs(proofGLs)
	proofGLs = nil // Free the slice as we no longer need it
	if err = serde.StoreToDisk(req.SharedRndFile, sharedRandomness, true); err != nil {
		logrus.Errorf("error saving shared randomness: %s. Cancelling context", err)
		cancel()
		return nil, fmt.Errorf("could not save shared randomness: %w", err)
	}
	logrus.Infof("Generated shared randomness for conflation request %s-%s and stored to path %s",
		req.StartBlock, req.EndBlock, req.SharedRndFile)

	// -- 5. Wait for all  proofs to arrive from the LPP sub-provers and
	// deserialize and send them to the conglomeration pipeline as they arrive
	errGroup.SetLimit(req.NumLPP)
	for i := 0; i < req.NumLPP; i++ {
		i := i // local copy
		errGroup.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			var (
				jobErr   error
				proofLPP = &distributed.SegmentProof{}
			)

			jobErr = files.WaitForFileAtPath(ctx, req.LPPProofFiles[i], time.Duration(cfg.ExecutionLimitless.PollInterval), true, fmt.Sprintf("Waiting for LPP proof %d", i))
			if jobErr != nil {
				logrus.Errorf("error waiting for LPP proof %d: %s - Cancelling context", i, jobErr)
				return jobErr
			}

			// Once the LPP-proof file has arrived - deserialize it and send it to the pipeline
			closer, err := serde.LoadFromDisk(req.LPPProofFiles[i], proofLPP, true)
			if err != nil {
				logrus.Errorf("error deserializing LPP proof %d: %s - Cancelling context", i, err)
				return err
			}
			defer closer.Close()

			// Safe send: if ctx cancelled, abort send
			select {
			case proofStream <- proofLPP:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})
	}

	// Wait for all LPP provers
	if err := errGroup.Wait(); err != nil {
		// Cancel overall flow (aggregator will observe ctx.Done)
		cancel()
		// Wait for aggregator to finish (so resultCh gets something)
		res := <-resultCh
		if res.err != nil {
			// aggregator returned an error (likely due to cancel); return original error
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

	logrus.Infof("Successfully aggregated all %d GL and %d LPP execution sub-proofs.", req.NumGL, req.NumLPP)

	// -- 5. Load setup (in background started earlier maybe)
	execReq := &execution.Request{}
	if err := files.ReadRequest(req.BootstrapRequestDoneFile, execReq); err != nil {
		return nil, fmt.Errorf("could not read the execution proof request file (%v): %w", req.BootstrapRequestDoneFile, err)
	}

	out := execution.CraftProverOutput(cfg, execReq)
	execResp = &out
	var (
		witness        = execution.NewWitness(cfg, execReq, execResp)
		congFinalproof = res.proof
		setup          circuits.Setup
		errSetup       error
	)

	logrus.Infof("Loading setup - circuitID: %s", circuits.ExecutionLimitlessCircuitID)
	setup, errSetup = circuits.LoadSetup(cfg, circuits.ExecutionLimitlessCircuitID)

	if errSetup != nil {
		utils.Panic("could not load setup: %v", errSetup)
	}
	out.Proof = execCirc.MakeProof(
		&cfg.TracesLimits,
		setup,
		cong.RecursionComp,
		congFinalproof.GetOuterProofInput(),
		*witness.FuncInp,
		witness.ZkEVM.ExecData,
	)

	out.VerifyingKeyShaSum = setup.VerifyingKeyDigest()
	return &out, nil
}

// runConglomerationHierarchical aggregates segment proofs into a single proof.
// It returns the final proof or an error. It respects the passed context for cancellation.
func runConglomerationHierarchical(ctx context.Context, cfg *config.Config, cong *distributed.RecursedSegmentCompilation,
	proofStream <-chan *distributed.SegmentProof, totalProofs int,
) (*distributed.SegmentProof, error) {

	mt, err := zkevm.LoadVerificationKeyMerkleTree(cfg)
	if err != nil {
		return nil, fmt.Errorf("could not load verification key merkle tree: %w", err)
	}

	// Stack is a slice for channel-based pairing logic
	var stack []*distributed.SegmentProof
	proofsReceived := 0

	// Main loop: block on either new proof, or cancellation.
	for {
		// First, aggregate while we have at least 2 items
		for len(stack) >= 2 {

			// _proof2 normally has already its runtime cleared
			_proof1 := stack[len(stack)-1].ClearRuntime()
			_proof2 := stack[len(stack)-2]
			stack = stack[:len(stack)-2]

			logrus.Infof("Conglomerating sub-proofs for (proofType, moduleIdx, segmentIdx) = (%d, %d, %d) and (%d, %d, %d)",
				_proof1.ProofType, _proof1.ModuleIndex, _proof1.SegmentIndex,
				_proof2.ProofType, _proof2.ModuleIndex, _proof2.SegmentIndex)
			aggregated := cong.ProveSegment(&distributed.ModuleWitnessConglo{
				SegmentProofs:             []distributed.SegmentProof{*_proof1, *_proof2},
				VerificationKeyMerkleTree: *mt,
			})
			stack = append(stack, aggregated)
		}

		// If we've received all proofs and have exactly one on the stack -> done
		if proofsReceived >= totalProofs {
			// Last item is the final proof
			if len(stack) == 1 {
				logrus.Infoln("Successfully finished running conglomeration prover.")
				finalProof := stack[0]
				// Clear stack
				stack = nil
				return finalProof, nil
			}
			return nil, fmt.Errorf("conglomeration finished but stack size=%d (expected 1)", len(stack))
		}

		// Wait for next proof or cancellation
		select {
		case <-ctx.Done():
			return nil, ctx.Err()

		// Receive new proof from main proof stream
		case p, ok := <-proofStream:
			if !ok {
				// sender closed channel prematurely
				if len(stack) == 1 {
					return stack[0], nil
				}
				return nil, fmt.Errorf("proof stream closed prematurely; stack size=%d, proofsReceived=%d, totalProofs=%d", len(stack), proofsReceived, totalProofs)
			}
			logrus.Infof("Received proof (proofType, moduleIdx, segmentIdx) = (%d, %d, %d) for conglomeration", p.ProofType, p.ModuleIndex, p.SegmentIndex)
			stack = append(stack, p)
			proofsReceived++
		}
	}
}
