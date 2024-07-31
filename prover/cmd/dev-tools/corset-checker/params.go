package main

import (
	"flag"
	"fmt"

	"github.com/consensys/linea-monorepo/prover/config"
)

var (
	configFPathCLI string
	traceFPathCLI  string
)

func init() {
	flag.StringVar(&configFPathCLI, "config", "", "path to the config file. Only the trace limits are read")
	flag.StringVar(&traceFPathCLI, "trace-file", "", "path to the `.lt` trace file")
	flag.Parse()
}

func getParamsFromCLI() (cfg *config.Config, traceFPath string, err error) {

	flag.Parse()

	if len(configFPathCLI) == 0 {
		return nil, "", fmt.Errorf("could not find the config path, got %++v", configFPathCLI)
	}

	if len(traceFPathCLI) == 0 {
		return nil, "", fmt.Errorf("could not find the trace file path, got %++v", traceFPathCLI)
	}

	if cfg, err = config.NewConfigFromFile(configFPathCLI); err != nil {
		return nil, "", fmt.Errorf("could not parse the config: %w", err)
	}

	fmt.Printf("the file = %v\n", traceFPath)

	return cfg, traceFPathCLI, nil
}
