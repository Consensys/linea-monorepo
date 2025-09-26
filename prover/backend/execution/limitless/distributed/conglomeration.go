package distributed

import (
	"context"
	"fmt"
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
	// Set timeout for all gl subproofs
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.LimitlessParams.GLSubproofsTimeout)*time.Second)
	defer cancel()

	msg := fmt.Sprintf("Waiting and scanning for %d LPP commitments from GL provers with configured timeout:%d sec to generate shared randomness", req.NumGL, cfg.LimitlessParams.GLSubproofsTimeout)
	err := files.WaitForAllFilesAtPath(ctx, req.GLCommitFiles, true, msg)
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
