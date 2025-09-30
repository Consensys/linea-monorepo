package distributed

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"time"

	"github.com/consensys/linea-monorepo/prover/backend/execution"
	"github.com/consensys/linea-monorepo/prover/backend/files"
	"github.com/consensys/linea-monorepo/prover/circuits"
	execCirc "github.com/consensys/linea-monorepo/prover/circuits/execution"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/recursion"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/serialization"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/exit"
	"github.com/consensys/linea-monorepo/prover/zkevm"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	"github.com/consensys/linea-monorepo/prover/config"
)

var (

	// numConcurrentReadingGoroutines governs the goroutine deserializing,
	// decompressing and reading the proofs. The reading part is also controlled
	// by a semaphore on top of this.
	numConcurrentReadingGoroutines = 12
)

func RunConglomerator(cfg *config.Config, req *Metadata) (execResp *execution.Response, err error) {

	logrus.Infof("Starting conglomerator for conflation request %s-%s", req.StartBlock, req.EndBlock)

	// Recover wrapper for panics
	defer func() {
		if r := recover(); r != nil {
			logrus.Errorf("[PANIC] Conglomerator crashed for conflation request %s:%s\n%s", req.StartBlock, req.EndBlock, debug.Stack())
			return
		}
	}()

	// Generic defer for marking files on exit
	defer func() {
		suf := files.OutcomeSuffix(err)
		allFiles := append([]string{}, req.GLProofFiles...)
		allFiles = append(allFiles, req.LPPProofFiles...)
		allFiles = append(allFiles, req.SharedRndFile)
		files.MarkFiles(allFiles, suf)
	}()

	// Launch the shared randomness proces first
	err = runSharedRandomness(cfg, req)
	if err != nil {
		return nil, fmt.Errorf("error running shared randomness: %w", err)
	}

	// Success of shared randomness => Presence of GL sub-proof files
	// Start Deserializing the sub-proofs as they arrive
	logrus.Infoln("Begin deserializing GL-subproofs")

	wgGL, ctxGL := errgroup.WithContext(context.Background())
	wgGL.SetLimit(numConcurrentReadingGoroutines)

	proofGLs := make([]recursion.Witness, req.NumGL)
	for i := 0; i < req.NumGL; i++ {
		i := i
		wgGL.Go(func() error {
			select {
			case <-ctxGL.Done():
				return ctxGL.Err()
			default:
				proofGL := &recursion.Witness{}
				// GL proof loading
				if err := retryDeser(req.GLProofFiles[i], proofGL, true, cfg.LimitlessParams.NumberOfRetries, time.Duration(cfg.LimitlessParams.RetryDelay)*time.Millisecond); err != nil {
					return err
				}

				proofGLs[i] = *proofGL
				proofGL = nil
				return nil
			}
		})
	}

	if err := wgGL.Wait(); err != nil {
		return nil, fmt.Errorf("could not deserialize GL-subproofs: %v", err)
	}

	logrus.Infoln("Deserialized all GL-subproofs and loaded into memory")

	// Wait for all LPP sub-proofs files to arrive with specified timeout
	if err = waitForAllLPPWorkers(cfg, req); err != nil {
		return nil, err
	}

	logrus.Infoln("Begin deserializing LPP-subproofs")
	wgLPP, ctxLPP := errgroup.WithContext(context.Background())
	wgLPP.SetLimit(numConcurrentReadingGoroutines)

	proofLPPs := make([]recursion.Witness, req.NumLPP)
	for i := 0; i < req.NumLPP; i++ {
		i := i
		wgLPP.Go(func() error {
			select {
			case <-ctxLPP.Done():
				return ctxLPP.Err()
			default:
				proofLPP := &recursion.Witness{}
				// LPP proof loading
				if err := retryDeser(req.LPPProofFiles[i], proofLPP, true, cfg.LimitlessParams.NumberOfRetries, time.Duration(cfg.LimitlessParams.RetryDelay)*time.Millisecond); err != nil {
					return err
				}

				proofLPPs[i] = *proofLPP
				proofLPP = nil
				return nil
			}
		})
	}

	if err := wgLPP.Wait(); err != nil {
		return nil, fmt.Errorf("could not deserialize LPP-subproofs: %v", err)
	}

	logrus.Infoln("Deserialized all LPP-subproofs and loaded into memory")

	execReq := &execution.Request{}
	if err := files.ReadRequest(req.ExecutionRequestFile, execReq); err != nil {
		return nil, fmt.Errorf("could not read the execution proof request file (%v): %w", req.ExecutionRequestFile, err)
	}

	out := execution.CraftProverOutput(cfg, execReq)
	execResp = &out
	var (
		witness     = execution.NewWitness(cfg, execReq, execResp)
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

	logrus.Infof("Starting the conglomeration-prover for conflation request %s-%s", req.StartBlock, req.EndBlock)

	proofConglo, congloWIOP, err := runConglomeration(cfg, proofGLs, proofLPPs)
	if err != nil {
		return nil, fmt.Errorf("could not run conglomeration prover: %w", err)
	}

	<-chSetupDone
	if errSetup != nil {
		utils.Panic("could not load setup: %v", errSetup)
	}

	execResp.Proof = execCirc.MakeProof(
		&cfg.TracesLimits,
		setup,
		congloWIOP,
		proofConglo,
		*witness.FuncInp,
	)

	execResp.VerifyingKeyShaSum = setup.VerifyingKeyDigest()
	return execResp, nil
}

func runSharedRandomness(cfg *config.Config, req *Metadata) (err error) {
	// Incase the prev process was interrupted, rm any corrupted file if exists
	_ = os.Remove(req.SharedRndFile)

	// Add defer to mark outcome on all GLCommitFiles
	defer func() {
		suf := files.OutcomeSuffix(err)
		files.MarkFiles(req.GLCommitFiles, suf)
	}()

	// Set timeout for all gl subproofs
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.LimitlessParams.GLSubproofsTimeout)*time.Second)
	defer cancel()

	msg := fmt.Sprintf("Scanning for %d LPP commitments from GL provers with configured timeout:%d sec to generate shared randomness", req.NumGL, cfg.LimitlessParams.GLSubproofsTimeout)
	if err = files.WaitForAllFilesAtPath(ctx, req.GLCommitFiles, true, msg); err != nil {
		return fmt.Errorf("error waiting for all GL workers: %w", err)
	}

	// Once all the gl-lpp-commit.bin files have arrived, we can start deserializing
	lppCommitments := make([]field.Element, len(req.GLCommitFiles))
	for i, path := range req.GLCommitFiles {
		lppCommitment := &field.Element{}
		if derr := retryDeser(path, lppCommitment, true, cfg.LimitlessParams.NumberOfRetries, time.Duration(cfg.LimitlessParams.RetryDelay)*time.Millisecond); derr != nil {
			err = fmt.Errorf("could not load lpp-commitment: %w", derr)
			return err
		}
		lppCommitments[i] = *lppCommitment
	}

	// Generate and store the shared randomness
	sharedRandomness := distributed.GetSharedRandomness(lppCommitments)
	if err = serialization.StoreToDisk(req.SharedRndFile, sharedRandomness, true); err != nil {
		return fmt.Errorf("could not save shared randomness: %w", err)
	}

	logrus.Infof("Generated shared randomness for conflation request %s-%s and stored to path %s",
		req.StartBlock, req.EndBlock, req.SharedRndFile)
	return nil
}

func waitForAllLPPWorkers(cfg *config.Config, req *Metadata) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.LimitlessParams.LPPSubproofsTimeout)*time.Second)
	defer cancel()

	msg := fmt.Sprintf("Waiting for %d proof files from LPP workers with configured timeout:%d seconds", req.NumLPP, cfg.LimitlessParams.LPPSubproofsTimeout)
	err := files.WaitForAllFilesAtPath(ctx, req.LPPProofFiles, true, msg)
	if err != nil {
		return fmt.Errorf("error waiting for all LPP workers: %w", err)
	}
	return nil
}

// runConglomeration runs the conglomeration prover for the provided subproofs
func runConglomeration(cfg *config.Config, proofGLs, proofLPPs []recursion.Witness) (proof wizard.Proof, congloWIOP *wizard.CompiledIOP, err error) {

	logrus.Infof("Running the conglomeration-prover")

	// TODO: Implment mmap optimization here
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
}

// retryDeser: wraps LoadFromDisk with simple flat retries for io.EOF / io.ErrUnexpectedEOF
func retryDeser(path string, v any, compressed bool, retries int, delay time.Duration) error {
	for attempt := 0; attempt <= retries; attempt++ {
		err := serialization.LoadFromDisk(path, v, compressed)
		if err == nil {
			return nil
		}

		// Retry if the underlying cause is EOF-related
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			if attempt < retries {
				logrus.Warnf("Retrying load for %s after %v due to %v (attempt %d/%d)", path, delay, err, attempt+1, retries)
				time.Sleep(delay)
				continue
			}
		}

		return fmt.Errorf("could not deserialize %s: %w", path, err)
	}
	return fmt.Errorf("unreachable retry logic for %s", path)
}
