package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/consensys/linea-monorepo/prover/backend/aggregation"
	"github.com/consensys/linea-monorepo/prover/backend/blobdecompression"
	"github.com/consensys/linea-monorepo/prover/backend/execution"
	"github.com/consensys/linea-monorepo/prover/backend/execution/limitless"
	"github.com/consensys/linea-monorepo/prover/backend/execution/limitless/distributed"
	"github.com/consensys/linea-monorepo/prover/backend/files"
	"github.com/consensys/linea-monorepo/prover/config"
)

type ProverArgs struct {
	Input      string
	Output     string
	Large      bool
	Phase      string
	ConfigFile string
}

// Prove orchestrates the proving process based on the job type
func Prove(args ProverArgs) error {

	// TODO @gbotrel with a specific flag, we could compile the circuit and compare with the checksum of the
	// asset we deserialize, to make sure we are using the circuit associated with the compiled
	// binary and the setup.
	const cmdName = "prove"

	// Read config
	cfg, err := config.NewConfigFromFile(args.ConfigFile)
	if err != nil {
		return fmt.Errorf("%s failed to read config file: %w", cmdName, err)
	}

	// Determine job type from input file name
	var (
		jobExecution         = strings.Contains(args.Input, "getZkProof")
		jobBlobDecompression = strings.Contains(args.Input, "getZkBlobCompressionProof")
		jobAggregation       = strings.Contains(args.Input, "getZkAggregatedProof")

		// Limitless prover jobs
		jobBootstrap      = strings.Contains(args.Input, "getZkProof") && strings.EqualFold(args.Phase, "bootstrap")
		jobGL             = strings.Contains(args.Input, "wit.bin") && strings.EqualFold(args.Phase, "gl")
		jobConglomeration = strings.Contains(args.Input, "metadata-getZkProof") && strings.EqualFold(args.Phase, "conglomeration")
	)

	// Handle job type
	switch {
	case jobBootstrap:
		return handleBootstrapJob(cfg, args)
	case jobGL:
		return handleGLJob(cfg, args)
	case jobConglomeration:
		return handleConglomerationJob(cfg, args)
	case jobExecution:
		return handleExecutionJob(cfg, args)
	case jobBlobDecompression:
		return handleBlobDecompressionJob(cfg, args)
	case jobAggregation:
		return handleAggregationJob(cfg, args)
	default:
		return errors.New("unknown job type")
	}
}

func handleBootstrapJob(cfg *config.Config, args ProverArgs) error {

	if cfg.Execution.ProverMode != config.ProverModeLimitless {
		return fmt.Errorf("--phase flag can be invoked only in the %v mode", config.ProverModeLimitless)
	}

	req := &execution.Request{}
	if err := readRequest(args.Input, req); err != nil {
		return fmt.Errorf("could not read input file (%v): %w", args.Input, err)
	}

	// Extract start and end block from the req file
	sbr, ebr, err := files.ParseReqFile(args.Input)
	if err != nil {
		return err
	}

	// Build metadata from the request (example, you can adjust)
	metadata := &distributed.Metadata{
		StartBlock: sbr,
		EndBlock:   ebr,
	}

	metadata, err = distributed.RunBootstrapper(cfg, req, metadata)
	if err != nil {
		return fmt.Errorf("bootstrapper phase failed: %w", err)
	}

	return writeResponse(args.Output, metadata)
}

func handleGLJob(cfg *config.Config, args ProverArgs) error {

	if cfg.Execution.ProverMode != config.ProverModeLimitless {
		return fmt.Errorf("--phase flag can be invoked only in the %v mode", config.ProverModeLimitless)
	}

	sb, eb, segID, err := files.ParseWitnessFile(args.Input)
	if err != nil {
		return fmt.Errorf("unable to parse sb/eb/segID from %s: %w", args.Input, err)
	}
	req := &distributed.GLRequest{
		WitnessGLPath: args.Input,
		StartBlock:    sb,
		EndBlock:      eb,
		SegID:         segID,
	}
	resp, err := distributed.RunGL(cfg, req)
	if err != nil {
		return fmt.Errorf("GL phase failed: %w", err)
	}

	return writeResponse(args.Output, resp)
}

func handleConglomerationJob(cfg *config.Config, args ProverArgs) error {
	if cfg.Execution.ProverMode != config.ProverModeLimitless {
		return fmt.Errorf("--phase flag can be invoked only in the %v mode", config.ProverModeLimitless)
	}

	req := &distributed.Metadata{}
	if err := readRequest(args.Input, req); err != nil {
		return fmt.Errorf("could not read the input file (%v): %w", args.Input, err)
	}

	_, err := distributed.RunConglomerator(cfg, req)
	if err != nil {
		return fmt.Errorf("error while running the conglomerator: %w", err)
	}

	return nil
}

// handleExecutionJob processes an execution job
func handleExecutionJob(cfg *config.Config, args ProverArgs) error {
	req := &execution.Request{}
	if err := readRequest(args.Input, req); err != nil {
		return fmt.Errorf("could not read the input file (%v): %w", args.Input, err)
	}

	var resp *execution.Response
	var err error

	if cfg.Execution.ProverMode == config.ProverModeLimitless {
		// Limitless execution mode
		resp, err = limitless.Prove(cfg, req)
		if err != nil {
			return fmt.Errorf("could not prove the execution in limitless mode: %w", err)
		}
	} else {
		// Standard execution mode
		large := args.Large || (strings.Contains(args.Input, "large") && cfg.Execution.CanRunFullLarge)
		resp, err = execution.Prove(cfg, req, large)
		if err != nil {
			return fmt.Errorf("could not prove the execution: %w", err)
		}
	}

	return writeResponse(args.Output, resp)
}

// handleBlobDecompressionJob processes a blob decompression job
func handleBlobDecompressionJob(cfg *config.Config, args ProverArgs) error {
	req := &blobdecompression.Request{}
	if err := readRequest(args.Input, req); err != nil {
		return fmt.Errorf("could not read the input file (%v): %w", args.Input, err)
	}

	resp, err := blobdecompression.Prove(cfg, req)
	if err != nil {
		return fmt.Errorf("could not prove the blob decompression: %w", err)
	}

	return writeResponse(args.Output, resp)
}

// handleAggregationJob processes an aggregation job
func handleAggregationJob(cfg *config.Config, args ProverArgs) error {
	req := &aggregation.Request{}
	if err := readRequest(args.Input, req); err != nil {
		return fmt.Errorf("could not read the input file (%v): %w", args.Input, err)
	}

	resp, err := aggregation.Prove(cfg, req)
	if err != nil {
		return fmt.Errorf("could not prove the aggregation: %w", err)
	}

	return writeResponse(args.Output, resp)
}

// readRequest reads and decodes a request from a file
func readRequest(path string, into any) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("could not open file: %w", err)
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(into); err != nil {
		return fmt.Errorf("could not decode input file: %w", err)
	}

	return nil
}

// writeResponse encodes and writes a response to a file
func writeResponse(path string, from any) error {
	f := files.MustOverwrite(path)
	defer f.Close()

	if err := json.NewEncoder(f).Encode(from); err != nil {
		return fmt.Errorf("could not encode output file: %w", err)
	}

	return nil
}
