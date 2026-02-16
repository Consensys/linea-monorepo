package distributed

import (
	"context"
	"fmt"
	"os"
	"path"
	"runtime/debug"
	"time"

	"github.com/consensys/linea-monorepo/prover/backend/files"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/serde"

	"github.com/consensys/linea-monorepo/prover/utils/profiling"
	"github.com/consensys/linea-monorepo/prover/zkevm"
	"github.com/sirupsen/logrus"
)

type LPPRequest struct {
	WitnessLPPFile       string
	SharedRandomnessFile string
	StartBlock           string
	EndBlock             string
	SegID                int
}

func RunLPP(cfg *config.Config, req *LPPRequest) error {

	// Set MonitorParams before LPP-proving
	profiling.SetMonitorParams(cfg)

	logrus.Infof("Starting LPP-prover from witnessLPP file path %v", req.WitnessLPPFile)

	// Recover wrapper for panics
	defer func() {
		if r := recover(); r != nil {
			logrus.Errorf("[PANIC] LPP prover crashed for witness %s: %v", req.WitnessLPPFile, r)
			debug.PrintStack()
			os.Exit(2)
		}
	}()

	witnessLPP := &distributed.ModuleWitnessLPP{}
	closer1, err := serde.LoadFromDisk(req.WitnessLPPFile, witnessLPP, true)
	if err != nil {
		return fmt.Errorf("could not load witness: %w", err)
	}
	defer closer1.Close()

	compiledLPP, err := zkevm.LoadCompiledLPP(cfg, witnessLPP.ModuleName)
	if err != nil {
		return fmt.Errorf("could not load compiled LPP: %w", err)
	}

	logrus.Infof("Loaded the compiled LPP for witness module=%v at index=%d", witnessLPP.ModuleName, witnessLPP.ModuleIndex)

	// We wait for the shared randomness file to arrive
	var (
		sharedRandomnessFileName = fmt.Sprintf("%s-%s-commit.bin", req.StartBlock, req.EndBlock)
		sharedRandomnessFile     = path.Join(cfg.ExecutionLimitless.SharedRandomnessDir, config.RequestsFromSubDir, sharedRandomnessFileName)
	)
	req.SharedRandomnessFile = sharedRandomnessFile
	err = waitForSharedRandomnessFile(cfg, req)
	if err != nil {
		return err
	}

	// Generate the shared randomness
	sharedRandomness := &field.Octuplet{}
	closer2, err := serde.LoadFromDisk(sharedRandomnessFile, sharedRandomness, true)
	if err != nil {
		return fmt.Errorf("could not load shared randomness: %w", err)
	}
	defer closer2.Close()
	witnessLPP.InitialFiatShamirState = *sharedRandomness

	var (
		lppProofFileName = fmt.Sprintf("%s-%s-seg-%d-mod-%d-lpp-proof.bin", req.StartBlock, req.EndBlock, req.SegID, witnessLPP.ModuleIndex)
		proofLPPFile     = path.Join(cfg.ExecutionLimitless.SubproofsDir, "LPP", string(witnessLPP.ModuleName), config.RequestsFromSubDir, lppProofFileName)
	)

	// Incase the prev. process was interrupted, we clear the previous corrupted files (if it exists)
	_ = os.Remove(proofLPPFile)

	logrus.Infof("Running the LPP-prover for witness module name=%s at index=%d", witnessLPP.ModuleName, witnessLPP.ModuleIndex)

	proofLPP := compiledLPP.ProveSegment(witnessLPP)

	logrus.Infof("Finished running the LPP-prover for witness module=%v at index=%d", witnessLPP.ModuleName, witnessLPP.ModuleIndex)

	if err := serde.StoreToDisk(proofLPPFile, *proofLPP, true); err != nil {
		return fmt.Errorf("could not store LPP proof: %w", err)
	}

	logrus.Infof("Stored LPP proof for witness module=%v at index=%d to disk", witnessLPP.ModuleName, witnessLPP.ModuleIndex)

	return nil
}

func waitForSharedRandomnessFile(cfg *config.Config, req *LPPRequest) error {

	// Set timeout for all randomness beacon timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.ExecutionLimitless.Timeout)*time.Second)
	defer cancel()

	msg := fmt.Sprintf("Waiting for shared randomness file with configured timeout:%d sec", cfg.ExecutionLimitless.Timeout)
	err := files.WaitForFileAtPath(ctx, req.SharedRandomnessFile, time.Duration(cfg.ExecutionLimitless.PollInterval), true, msg)
	if err != nil {
		return fmt.Errorf("error waiting for shared randomness file: %w", err)
	}

	return nil
}
