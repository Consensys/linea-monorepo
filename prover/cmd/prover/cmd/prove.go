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
	FInput  string
	FOutput string
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

	proveCmd.Flags().StringVar(&FInput, "in", "", "input file")
	proveCmd.Flags().StringVar(&FOutput, "out", "", "output file")
	proveCmd.Flags().BoolVar(&fLarge, "large", false, "run the large execution circuit")

}

func cmdProve(cmd *cobra.Command, args []string) error {
	return CmdProve(cmd.Name(), args)
}

func CmdProve(cmdName string, args []string) error {
	// TODO @gbotrel with a specific flag, we could compile the circuit and compare with the checksum of the
	// asset we deserialize, to make sure we are using the circuit associated with the compiled binary and the setup.

	// read config
	cfg, err := config.NewConfigFromFile(FConfigFile)
	if err != nil {
		return fmt.Errorf("%s failed to read config file: %w", cmdName, err)
	}

	// discover the type of the job from the input file name
	jobExecution := strings.Contains(FInput, "getZkProof")
	jobBlobDecompression := strings.Contains(FInput, "getZkBlobCompressionProof")
	jobAggregation := strings.Contains(FInput, "getZkAggregatedProof")

	if jobExecution {
		req := &execution.Request{}
		if err := readRequest(FInput, req); err != nil {
			return fmt.Errorf("could not read the input file (%v): %w", FInput, err)
		}
		// we use the large traces in 2 cases;
		// 1. the user explicitly asked for it (fLarge)
		// 2. the job contains the large suffix and we are a large machine (cfg.Execution.CanRunLarge)
		large := fLarge || (strings.Contains(FInput, "large") && cfg.Execution.CanRunFullLarge)

		resp, err := execution.Prove(cfg, req, large)
		if err != nil {
			return fmt.Errorf("could not prove the execution: %w", err)
		}

		return writeResponse(FOutput, resp)
	}

	if jobBlobDecompression {
		req := &blobdecompression.Request{}
		if err := readRequest(FInput, req); err != nil {
			return fmt.Errorf("could not read the input file (%v): %w", FInput, err)
		}

		resp, err := blobdecompression.Prove(cfg, req)
		if err != nil {
			return fmt.Errorf("could not prove the blob decompression: %w", err)
		}

		return writeResponse(FOutput, resp)
	}

	if jobAggregation {
		req := &aggregation.Request{}
		if err := readRequest(FInput, req); err != nil {
			return fmt.Errorf("could not read the input file (%v): %w", FInput, err)
		}

		resp, err := aggregation.Prove(cfg, req)
		if err != nil {
			return fmt.Errorf("could not prove the aggregation: %w", err)
		}

		return writeResponse(FOutput, resp)
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
