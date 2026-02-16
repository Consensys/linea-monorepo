package controller

import (
	"context"
	"os"
	"path"

	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// run represents the command to run the prover
var rootCmd = &cobra.Command{
	Use:   "prover-controller",
	Short: "Run the prover",
	Run:   cobraControllerRunCmd,
}

// the arguments of the command
var (
	fConfig  string
	fLocalID string
)

// registers the arguments for the command
func init() {
	// mark the flags as required
	rootCmd.Flags().StringVar(&fConfig, "config", "", "config file")
	rootCmd.Flags().StringVar(&fLocalID, "local-id", "no-local-id-provided", "local ID of the controller container")
	rootCmd.MarkFlagRequired("config")
	rootCmd.MarkFlagRequired("local-id")
}

var (
	limitlessDirs    []string
	witnessReqDirs   []string
	witnessDoneDirs  []string
	subproofReqDirs  []string
	subproofDoneDirs []string

	sharedFailureDir  string
	metadataReqDir    string
	metadataDoneDir   string
	randomnessReqDir  string
	randomnessDoneDir string
)

// cobra command
func cobraControllerRunCmd(c *cobra.Command, args []string) {

	logrus.Infof("provided config files : %v", fConfig)
	logrus.Infof("provided local-id : %s", fLocalID)

	// load the configuration file
	cfg, err := config.NewConfigFromFile(fConfig)
	if err != nil {
		logrus.Fatalf("could not get the config : %v", err)
	}
	cfg.Controller.LocalID = fLocalID

	// Base dirs
	dirs := []string{
		cfg.Execution.DirDone(),
		cfg.Execution.DirFrom(),
		cfg.Execution.DirTo(),
		cfg.DataAvailability.DirDone(),
		cfg.DataAvailability.DirFrom(),
		cfg.DataAvailability.DirTo(),
		cfg.Aggregation.DirDone(),
		cfg.Aggregation.DirFrom(),
		cfg.Aggregation.DirTo(),
	}

	// ===== Limitless Core Directories (static)
	sharedFailureDir = cfg.ExecutionLimitless.SharedFailureDir
	metadataReqDir = path.Join(cfg.ExecutionLimitless.MetadataDir, config.RequestsFromSubDir)
	metadataDoneDir = path.Join(cfg.ExecutionLimitless.MetadataDir, config.RequestsDoneSubDir)

	randomnessReqDir = path.Join(cfg.ExecutionLimitless.SharedRandomnessDir, config.RequestsFromSubDir)
	randomnessDoneDir = path.Join(cfg.ExecutionLimitless.SharedRandomnessDir, config.RequestsDoneSubDir)

	// register static dirs
	limitlessDirs = append(limitlessDirs,
		sharedFailureDir,
		metadataReqDir, metadataDoneDir,
		randomnessReqDir, randomnessDoneDir,
	)

	// ===== Dynamic Module-Based Dirs
	for _, mod := range config.ALL_MODULES {

		// Witness(GL/LPP) dirs
		witnessReqDirs = append(witnessReqDirs, path.Join(cfg.ExecutionLimitless.WitnessDir, "GL", mod, config.RequestsFromSubDir))
		witnessDoneDirs = append(witnessDoneDirs, path.Join(cfg.ExecutionLimitless.WitnessDir, "GL", mod, config.RequestsDoneSubDir))
		witnessReqDirs = append(witnessReqDirs, path.Join(cfg.ExecutionLimitless.WitnessDir, "LPP", mod, config.RequestsFromSubDir))
		witnessDoneDirs = append(witnessDoneDirs, path.Join(cfg.ExecutionLimitless.WitnessDir, "LPP", mod, config.RequestsDoneSubDir))

		// Subproofs(GL/LPP) dirs
		subproofReqDirs = append(subproofReqDirs, path.Join(cfg.ExecutionLimitless.SubproofsDir, "GL", mod, config.RequestsFromSubDir))
		subproofDoneDirs = append(subproofDoneDirs, path.Join(cfg.ExecutionLimitless.SubproofsDir, "GL", mod, config.RequestsDoneSubDir))
		subproofReqDirs = append(subproofReqDirs, path.Join(cfg.ExecutionLimitless.SubproofsDir, "LPP", mod, config.RequestsFromSubDir))
		subproofDoneDirs = append(subproofDoneDirs, path.Join(cfg.ExecutionLimitless.SubproofsDir, "LPP", mod, config.RequestsDoneSubDir))
	}

	// Combine all for creation
	limitlessDirs = append(limitlessDirs, witnessReqDirs...)
	limitlessDirs = append(limitlessDirs, witnessDoneDirs...)
	limitlessDirs = append(limitlessDirs, subproofReqDirs...)
	limitlessDirs = append(limitlessDirs, subproofDoneDirs...)

	// Final list
	dirs = append(dirs, limitlessDirs...)

	// Create
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			logrus.Fatalf("could not create directory %s : %v", dir, err)
		}
	}

	// Start the main loop
	runController(context.Background(), cfg)
}

// Execute the cobra root command
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		logrus.Errorf("Got an error: %v", err)
		os.Exit(1)
	}
}
