package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/consensys/linea-monorepo/prover/backend/execution"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/zkevm"
	"github.com/dnlo/struct2csv"
)

type LogStatsArgs struct {
	Input      string
	StatsFile  string
	ConfigFile string
}

func LogStats(_ context.Context, args LogStatsArgs) error {

	const cmdName = "log-stats"

	// Read config
	cfg, err := config.NewConfigFromFile(args.ConfigFile)
	if err != nil {
		return fmt.Errorf("%s failed to read config file: %w", cmdName, err)
	}

	// Read the input file
	req := &execution.Request{}
	if err := readRequest(args.Input, req); err != nil {
		return fmt.Errorf("could not read the input file (%v): %w", args.Input, err)
	}

	// Setup execution witness and output response
	var (
		out     = execution.CraftProverOutput(cfg, req)
		witness = execution.NewWitness(cfg, req, &out)
		lz      = zkevm.NewLimitlessDebugZkEVM(cfg)
		stats   = lz.RunStatRecords(witness.ZkEVM)
	)

	// Open the file in append mode, create it if it doesn't exist
	file, err := os.OpenFile(args.StatsFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("an error occurred while opening the stats file: %w", err)
	}
	defer file.Close() // Ensure the file is closed when the function exits

	// Write the stats to the document. We always append the header to the CSV
	// because it's easier to remove it from the stats-files if needed than the
	// doing the contrary.
	w := struct2csv.NewWriter(file)

	if err := w.WriteColNames(stats[0]); err != nil {
		return fmt.Errorf("an error occurred while writing the stats file header: %w", err)
	}

	for _, stat := range stats {
		if err := w.WriteStruct(stat); err != nil {
			return fmt.Errorf("an error occurred while writing the stats file row: %w", err)
		}
	}

	w.Flush()
	return nil
}
