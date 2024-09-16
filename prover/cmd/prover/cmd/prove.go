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
	"github.com/consensys/linea-monorepo/prover/backend/files"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/spf13/cobra"
)

var (
	fInput  string
	fOutput string
	fLarge  bool
)

// proveCmd represents the prove command
var proveCmd = &cobra.Command{
	Use:   "prove",
	Short: "prove process a request, creates a proof with the adequate circuit and writes the proof to a file",
	RunE:  cmdProve,
}

func init() {
	rootCmd.AddCommand(proveCmd)

	proveCmd.Flags().StringVar(&fInput, "in", "", "input file")
	proveCmd.Flags().StringVar(&fOutput, "out", "", "output file")
	proveCmd.Flags().BoolVar(&fLarge, "large", false, "run the large execution circuit")

}

func cmdProve(cmd *cobra.Command, args []string) error {
	// TODO @gbotrel with a specific flag, we could compile the circuit and compare with the checksum of the
	// asset we deserialize, to make sure we are using the circuit associated with the compiled binary and the setup.

	// read config
	cfg, err := config.NewConfigFromFile(fConfigFile)
	if err != nil {
		return fmt.Errorf("%s failed to read config file: %w", cmd.Name(), err)
	}

	// discover the type of the job from the input file name
	jobExecution := strings.Contains(fInput, "getZkProof")
	jobBlobDecompression := strings.Contains(fInput, "getZkBlobCompressionProof")
	jobAggregation := strings.Contains(fInput, "getZkAggregatedProof")

	if jobExecution {
		req := &execution.Request{}
		if err := readRequest(fInput, req); err != nil {
			return fmt.Errorf("could not read the input file (%v): %w", fInput, err)
		}
		// we use the large traces in 2 cases;
		// 1. the user explicitly asked for it (fLarge)
		// 2. the job contains the large suffix and we are a large machine (cfg.Execution.CanRunLarge)
		large := fLarge || (strings.Contains(fInput, "large") && cfg.Execution.CanRunFullLarge)

		resp, err := execution.Prove(cfg, req, large)
		if err != nil {
			return fmt.Errorf("could not prove the execution: %w", err)
		}

		return writeResponse(fOutput, resp)
	}

	if jobBlobDecompression {
		req := &blobdecompression.Request{}
		if err := readRequest(fInput, req); err != nil {
			return fmt.Errorf("could not read the input file (%v): %w", fInput, err)
		}

		resp, err := blobdecompression.Prove(cfg, req)
		if err != nil {
			return fmt.Errorf("could not prove the blob decompression: %w", err)
		}

		return writeResponse(fOutput, resp)
	}

	if jobAggregation {
		req := &aggregation.Request{}
		if err := readRequest(fInput, req); err != nil {
			return fmt.Errorf("could not read the input file (%v): %w", fInput, err)
		}

		resp, err := aggregation.Prove(cfg, req)
		if err != nil {
			return fmt.Errorf("could not prove the aggregation: %w", err)
		}

		return writeResponse(fOutput, resp)
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
