package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/consensys/linea-monorepo/prover/backend/aggregation"
	"github.com/consensys/linea-monorepo/prover/backend/blobdecompression"
	"github.com/consensys/linea-monorepo/prover/backend/execution"
	"github.com/consensys/linea-monorepo/prover/backend/execution/limitless"
	"github.com/consensys/linea-monorepo/prover/backend/files"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/pkg/profile"
)

type ProverArgs struct {
	Input      string
	Output     string
	Large      bool
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
	)

	// Handle job type
	switch {
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

// handleExecutionJob processes an execution job
func handleExecutionJob(cfg *config.Config, args ProverArgs) error {
	req := &execution.Request{}

	p := profile.Start(profile.CPUProfile, profile.ProfilePath("."), profile.NoShutdownHook)
	startReadReq := time.Now()
	if err := readRequest(args.Input, req); err != nil {
		return fmt.Errorf("could not read the input file (%v): %w", args.Input, err)
	}
	fmt.Printf("Time to read request: %s\n", time.Since(startReadReq))

	// start a go routine that wakes up every minute to print current number of go routines and heap usage
	go func() {
		for {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			var size float64
			var unit string
			switch {
			case m.HeapAlloc >= 1<<30:
				size = float64(m.HeapAlloc) / (1 << 30)
				unit = "GiB"
			case m.HeapAlloc >= 1<<20:
				size = float64(m.HeapAlloc) / (1 << 20)
				unit = "MiB"
			case m.HeapAlloc >= 1<<10:
				size = float64(m.HeapAlloc) / (1 << 10)
				unit = "KiB"
			default:
				size = float64(m.HeapAlloc)
				unit = "B"
			}
			fmt.Printf("Number of go routines: %d, HeapAlloc = %.2f %s\n", runtime.NumGoroutine(), size, unit)
			time.Sleep(10 * time.Second)

			if runtime.NumGoroutine() <= 6 {
				buf := make([]byte, 1<<20) // 1 MB buffer
				n := runtime.Stack(buf, true)
				os.Stderr.Write(buf[:n])
			}
		}
	}()

	var resp *execution.Response
	var err error
	startProve := time.Now()
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
	fmt.Printf("Time to prove: %s\n", time.Since(startProve))

	startWrite := time.Now()
	err = writeResponse(args.Output, resp)
	fmt.Printf("Time to write response: %s\n", time.Since(startWrite))

	p.Stop()
	return err
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
