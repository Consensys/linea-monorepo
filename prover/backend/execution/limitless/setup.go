package limitless

import (
	"os"
	"path"

	"github.com/consensys/linea-monorepo/prover/backend/execution"
	"github.com/consensys/linea-monorepo/prover/circuits"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/utils"
)

func loadCktSetupAsync(cfg *config.Config) (*circuits.Setup, chan struct{}, error) {
	var (
		setup       circuits.Setup
		errSetup    error
		chSetupDone = make(chan struct{})
	)
	go func() {
		setup, errSetup = circuits.LoadSetup(cfg, circuits.ExecutionLimitlessCircuitID)
		close(chSetupDone)
	}()

	return &setup, chSetupDone, errSetup
}

// Helper function to finalize setup and validate checksum
func finalizeCktSetup(cfg *config.Config, chSetupDone <-chan struct{},
	setup *circuits.Setup, errSetup error) error {
	<-chSetupDone
	if errSetup != nil {
		utils.Panic("could not load setup: %v", errSetup)
	}
	execution.ValidateSetupChecksum(*setup, &cfg.TracesLimits)
	return nil
}

// Helper function to clean up witness directory
func cleanWitnessDirectory(cfg *config.Config) {
	filepath := cfg.PathforLimitlessProverAssets()
	filepath = path.Join(filepath, "witness")
	os.RemoveAll(filepath)
}
