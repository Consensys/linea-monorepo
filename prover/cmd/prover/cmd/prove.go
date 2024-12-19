package cmd

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/consensys/linea-monorepo/prover/backend/aggregation"
	"github.com/consensys/linea-monorepo/prover/backend/blobdecompression"
	"github.com/consensys/linea-monorepo/prover/backend/execution"
	"github.com/consensys/linea-monorepo/prover/backend/files"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/sirupsen/logrus"
)

type ProverArgs struct {
	Input      string
	Output     string
	Large      bool
	ConfigFile string
}

func Prove(args ProverArgs) error {
	const cmdName = "prove"
	// TODO @gbotrel with a specific flag, we could compile the circuit and compare with the checksum of the
	// asset we deserialize, to make sure we are using the circuit associated with the compiled binary and the setup.

	// read config
	cfg, err := config.NewConfigFromFile(args.ConfigFile)
	if err != nil {
		return fmt.Errorf("%s failed to read config file: %w", cmdName, err)
	}

	// discover the type of the job from the input file name
	jobExecution := strings.Contains(args.Input, "getZkProof")
	jobBlobDecompression := strings.Contains(args.Input, "getZkBlobCompressionProof")
	jobAggregation := strings.Contains(args.Input, "getZkAggregatedProof")

	if jobExecution {
		req := &execution.Request{}
		if err := readRequest(args.Input, req); err != nil {
			return fmt.Errorf("could not read the input file (%v): %w", args.Input, err)
		}

		// we use the large traces in 2 cases;
		// 1. the user explicitly asked for it (args.Large)
		// 2. the job contains the large suffix and we are a large machine (cfg.Execution.CanRunLarge)
		large := args.Large || (strings.Contains(args.Input, "large") && cfg.Execution.CanRunFullLarge)

		// check the arithmetization version used to generate the trace is contained in the prover request
		// and fail fast if the constraint version is not supported
		if err := checkArithmetizationVersion(req.ConflatedExecutionTracesFile, req.TracesEngineVersion, "./constraints-versions.txt"); err != nil {
			return err
		}

		resp, err := execution.Prove(cfg, req, large)
		if err != nil {
			return fmt.Errorf("could not prove the execution: %w", err)
		}

		return writeResponse(args.Output, resp)
	}

	if jobBlobDecompression {
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

	if jobAggregation {
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

	return errors.New("unknown job type")
}

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

func writeResponse(path string, from any) error {
	f := files.MustOverwrite(path)
	defer f.Close()

	if err := json.NewEncoder(f).Encode(from); err != nil {
		return fmt.Errorf("could not encode output file: %w", err)
	}

	return nil
}

// verifies the arithmetization version used to generate the trace file against the list of versions
// specified by the constraints in the file path.
func checkArithmetizationVersion(traceFileName, tracesEngineVersion, filepath string) error {
	logrus.Info("Verifying the arithmetization version for generating the trace file is supported by the constraints version")
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	traceFileVersion, err := validateAndExtractVersion(traceFileName)
	if err != nil {
		return err
	}

	if strings.Compare(traceFileVersion, tracesEngineVersion) != 0 {
		return fmt.Errorf("version specified in the conflated trace file: %s does not match with the trace engine version: %s", traceFileVersion, tracesEngineVersion)
	}

	scanner := bufio.NewScanner(file)
	if err := scanner.Err(); err != nil {
		return err
	}

	for scanner.Scan() {
		version := strings.TrimSpace(scanner.Text())
		if version != "" && strings.Compare(traceFileVersion, version) == 0 && strings.Compare(tracesEngineVersion, version) == 0 {
			return nil
		}
	}
	return fmt.Errorf("unsupported arithmetization version found in the conflated trace file: %s", traceFileName)
}

func validateAndExtractVersion(traceFileName string) (string, error) {
	logrus.Info("Validating and extracting the version from conflated trace files")
	// Define the regex pattern with a capturing group for the version part
	traceFilePattern := `^\d+-\d+\.conflated\.(v\d+\.\d+\.\d+-[^.]+)\.lt$`
	re := regexp.MustCompile(traceFilePattern)

	// Check if the file name matches the pattern and extract the version part
	matches := re.FindStringSubmatch(traceFileName)
	if len(matches) > 1 {
		return matches[1], nil
	}
	return "", fmt.Errorf("conflated trace file: %s not in the appropriate format or version not found", traceFileName)
}
