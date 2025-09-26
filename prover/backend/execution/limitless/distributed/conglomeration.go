package distributed

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"
	"time"

	"github.com/consensys/linea-monorepo/prover/backend/execution"
	"github.com/consensys/linea-monorepo/prover/backend/files"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/serialization"
	"github.com/sirupsen/logrus"

	"github.com/consensys/linea-monorepo/prover/config"
)

func RunConglomerator(cfg *config.Config, req *Metadata) (*execution.Response, error) {

	logrus.Infof("Starting conglomerator for conflation request%s-%s", req.StartBlock, req.EndBlock)

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

	// Start Deserializing the sub-proofs as they arrive
	logrus.Infoln("Begin deserializing gl-subproofs...")

	return nil, nil
}

func runSharedRandomness(cfg *config.Config, req *Metadata) error {

	// Incase the prev process was interrupted, rm any corrupted file if exists
	_ = os.Remove(req.SharedRndFile)

	// Set timeout for all gl subproofs
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.LimitlessParams.GLSubproofsTimeout)*time.Second)
	defer cancel()

	// Important: Though only GL commit files are required to generate shared randomness, we want to always
	// ensure that both GL proof files and GL commit files always exist. This is for the conglomerator to verify atomicity.
	// We dont want a case where GL proof files have arrived but commit files have not arrived due to a crash or vice-versa.
	filesToWatch := make([]string, 0, 2*req.NumGL)
	filesToWatch = append(filesToWatch, req.GLProofFiles...)
	filesToWatch = append(filesToWatch, req.GLCommitFiles...)

	msg := fmt.Sprintf("Scanning for %d GL proof files and %d LPP commitments from GL provers with configured timeout:%d sec to generate shared randomness", req.NumGL, req.NumGL, cfg.LimitlessParams.GLSubproofsTimeout)
	err := files.WaitForAllFilesAtPath(ctx, filesToWatch, true, msg)
	if err != nil {
		return fmt.Errorf("error waiting for all gl-lpp commit files: %w", err)
	}

	// Once all the gl-lpp-commit.bin files have arrived, we can start deserializing
	// the lpp commitments and generate the shared random commitment for LPP workers
	lppCommitments := make([]field.Element, len(req.GLCommitFiles))
	for i, path := range req.GLCommitFiles {
		lppCommitment := &field.Element{}
		if err := serialization.LoadFromDisk(path, lppCommitment, true); err != nil {
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
