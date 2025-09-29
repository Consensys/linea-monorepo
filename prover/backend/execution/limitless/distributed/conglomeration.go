package distributed

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"time"

	"github.com/consensys/linea-monorepo/prover/backend/execution"
	"github.com/consensys/linea-monorepo/prover/backend/files"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/recursion"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/serialization"
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

func RunConglomerator(cfg *config.Config, req *Metadata) (*execution.Response, error) {

	logrus.Infof("Starting conglomerator for conflation request %s-%s", req.StartBlock, req.EndBlock)

	// Recover wrapper for panics
	defer func() {
		if r := recover(); r != nil {
			logrus.Errorf("[PANIC] Conglomerator crashed for conflation request %s:%s\n%s", req.StartBlock, req.EndBlock, debug.Stack())
			return
		}
	}()

	// Launch the shared randomness proces first
	err := runSharedRandomness(cfg, req)
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
				if err := serialization.LoadFromDisk(req.GLProofFiles[i], proofGL, true); err != nil {
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

	logrus.Infoln("Deserialized all GL-subproofs and waiting for all LPP workers to finish")

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
				if err := serialization.LoadFromDisk(req.LPPProofFiles[i], proofLPP, true); err != nil {
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

	logrus.Infoln("Finished deserialization of all LPP-subproofs.")

	logrus.Infof("Starting the conglomeration-prover for conflation request %s-%s", req.StartBlock, req.EndBlock)
	return nil, nil
}

func runSharedRandomness(cfg *config.Config, req *Metadata) error {

	// Incase the prev process was interrupted, rm any corrupted file if exists
	_ = os.Remove(req.SharedRndFile)

	// Set timeout for all gl subproofs
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.LimitlessParams.GLSubproofsTimeout)*time.Second)
	defer cancel()

	msg := fmt.Sprintf("Scanning for %d LPP commitments from GL provers with configured timeout:%d sec to generate shared randomness", req.NumGL, cfg.LimitlessParams.GLSubproofsTimeout)
	err := files.WaitForAllFilesAtPath(ctx, req.GLCommitFiles, true, msg)
	if err != nil {
		return fmt.Errorf("error waiting for all GL workers: %w", err)
	}

	// Once all the gl-lpp-commit.bin files have arrived, we can start deserializing
	// the lpp commitments and generate the shared random commitment for LPP workers
	lppCommitments := make([]field.Element, len(req.GLCommitFiles))
	for i, path := range req.GLCommitFiles {
		lppCommitment := &field.Element{}
		if err := serialization.LoadFromDisk(path, lppCommitment, true); err != nil {
			if err == io.EOF {

			}
			return fmt.Errorf("could not load lpp-commitment: %w", err)
		}
		lppCommitments[i] = *lppCommitment
	}

	// Generate and store the shared randomness
	sharedRandomness := distributed.GetSharedRandomness(lppCommitments)
	if err := serialization.StoreToDisk(req.SharedRndFile, sharedRandomness, true); err != nil {
		return fmt.Errorf("could not save shared randomness: %w", err)
	}

	logrus.Infof("Generated shared randomness for conflation request %s-%s and stored to path %s", req.StartBlock, req.EndBlock, req.SharedRndFile)
	return nil
}

func waitForAllLPPWorkers(cfg *config.Config, req *Metadata) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.LimitlessParams.LPPSubproofsTimeout)*time.Second)
	defer cancel()

	msg := fmt.Sprintf("Scanning for %d proof files from LPP provers with configured timeout:%d seconds", req.NumLPP, cfg.LimitlessParams.LPPSubproofsTimeout)
	err := files.WaitForAllFilesAtPath(ctx, req.LPPProofFiles, true, msg)
	if err != nil {
		return fmt.Errorf("error waiting for all LPP workers: %w", err)
	}
	return nil
}
