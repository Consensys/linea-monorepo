package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/consensys/linea-monorepo/prover/backend/aggregation"
	"github.com/consensys/linea-monorepo/prover/backend/dataavailability"
	"github.com/consensys/linea-monorepo/prover/backend/execution"
	"github.com/consensys/linea-monorepo/prover/backend/execution/limitless"
	"github.com/consensys/linea-monorepo/prover/backend/execution/limitless/distributed"
	"github.com/consensys/linea-monorepo/prover/backend/files"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/utils"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/utils/signal"
	"github.com/sirupsen/logrus"
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

	// This allows the user to dump stacktraces by sending a SIGUSR1 to the
	// current process.
	signal.RegisterStackTraceDumpHandler()

	// TODO @gbotrel with a specific flag, we could compile the circuit and compare with the checksum of the
	// asset we deserialize, to make sure we are using the circuit associated with the compiled
	// binary and the setup.
	const cmdName = "prove"

	// Read config
	cfg, err := config.NewConfigFromFile(args.ConfigFile)
	if err != nil {
		return fmt.Errorf("%s failed to read config file at %v: %w", cmdName, args.ConfigFile, err)
	}

	// Determine job type from input file name
	var (
		jobExecution        = strings.Contains(args.Input, "getZkProof")
		jobDataAvailability = strings.Contains(args.Input, "getZkBlobCompressionProof")
		jobAggregation      = strings.Contains(args.Input, "getZkAggregatedProof")

		// Limitless prover jobs
		jobBootstrap      = strings.Contains(args.Input, "getZkProof") && strings.EqualFold(args.Phase, "bootstrap")
		jobGL             = strings.Contains(args.Input, "wit.bin") && strings.EqualFold(args.Phase, "gl")
		jobLPP            = strings.Contains(args.Input, "wit.bin") && strings.EqualFold(args.Phase, "lpp")
		jobConglomeration = strings.Contains(args.Input, "metadata-getZkProof") && strings.EqualFold(args.Phase, "conglomeration")
	)

	// Handle job type
	switch {
	case jobBootstrap:
		return handleBootstrapJob(cfg, args)
	case jobGL:
		return handleGLJob(cfg, args)
	case jobLPP:
		return handleLPPJob(cfg, args)
	case jobConglomeration:
		return handleConglomerationJob(cfg, args)
	case jobExecution:
		return handleExecutionJob(cfg, args)
	case jobDataAvailability:
		return handleDataAvailabilityJob(cfg, args)
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
	if err := files.ReadRequest(args.Input, req); err != nil {
		return fmt.Errorf("could not read input file (%v): %w", args.Input, err)
	}

	// Extract start and end block from the req file
	sbr, ebr, err := files.ParseReqFile(args.Input)
	if err != nil {
		return err
	}

	// local helper: takes the input filename and appends the bootstrap partial success suffix
	// after trimming the inprogress suffix
	makeBootstrapDoneFile := func(input string) string {
		file := filepath.Base(input)
		// find `.inprogress` + suffix
		if idx := strings.Index(file, config.InProgressSuffix); idx != -1 {
			return file[:idx] + config.BootstrapPartialSucessSuffix
		}

		utils.Panic("failed to find %q in %q", config.InProgressSuffix, file)
		return "" // unreachable
	}

	metadata := &distributed.Metadata{
		BootstrapRequestDoneFile: filepath.Join(filepath.Dir(args.Input), makeBootstrapDoneFile(args.Input)),
		StartBlock:               sbr,
		EndBlock:                 ebr,
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
		WitnessGLFile: args.Input,
		StartBlock:    sb,
		EndBlock:      eb,
		SegID:         segID,
	}
	return distributed.RunGL(cfg, req)
}

func handleLPPJob(cfg *config.Config, args ProverArgs) error {

	if cfg.Execution.ProverMode != config.ProverModeLimitless {
		return fmt.Errorf("--phase flag can be invoked only in the %v mode", config.ProverModeLimitless)
	}

	sb, eb, segID, err := files.ParseWitnessFile(args.Input)
	if err != nil {
		return fmt.Errorf("unable to parse sb/eb/segID from %s: %w", args.Input, err)
	}
	req := &distributed.LPPRequest{
		WitnessLPPFile: args.Input,
		StartBlock:     sb,
		EndBlock:       eb,
		SegID:          segID,
	}
	return distributed.RunLPP(cfg, req)
}

func handleConglomerationJob(cfg *config.Config, args ProverArgs) error {
	if cfg.Execution.ProverMode != config.ProverModeLimitless {
		return fmt.Errorf("--phase flag can be invoked only in the %v mode", config.ProverModeLimitless)
	}

	req := &distributed.Metadata{}
	if err := files.ReadRequest(args.Input, req); err != nil {
		return fmt.Errorf("could not read the input file (%v): %w", args.Input, err)
	}

	// Get local id of conglomerator
	var (
		parts           = strings.Split(args.Input, "."+config.InProgressSuffix+".")
		localID         = parts[1] // gauranteed by controller to have the *inprogress suffix (wont panic)
		marker          = fmt.Sprintf("%s.", config.InProgressSuffix)
		newInprogSuffix = marker + localID
	)

	if splits := strings.SplitN(req.BootstrapRequestDoneFile, config.BootstrapPartialSucessSuffix, 2); len(splits) == 2 {
		var (
			oldFile = req.BootstrapRequestDoneFile
			newFile = splits[0] + newInprogSuffix
		)

		// Rename the original Bootstrap file with inprogress appended by localID so that we know
		// which pod is going to deal with the conglomeration job
		if err := os.Rename(oldFile, newFile); err != nil {
			return fmt.Errorf("failed to rename file while handling conglomeration job: %v", err)
		} else {
			logrus.Printf("Renamed file: %s -> %s\n", oldFile, newFile)
			req.BootstrapRequestDoneFile = newFile
		}
	}

	resp, err := distributed.RunConglomerator(cfg, req)
	if err != nil {
		return fmt.Errorf("error while running the conglomerator: %w", err)
	}

	return writeResponse(args.Output, resp)
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

// handleDataAvailabilityJob processes a data availability proof job
func handleDataAvailabilityJob(cfg *config.Config, args ProverArgs) error {
	req := &dataavailability.Request{}
	if err := readRequest(args.Input, req); err != nil {
		return fmt.Errorf("could not read the input file (%v): %w", args.Input, err)
	}

	resp, err := dataavailability.Prove(cfg, req)
	if err != nil {
		return fmt.Errorf("could not prove the data availability: %w", err)
	}

	return writeResponse(args.Output, resp)
}
