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

	// TODO @gbotrel @AlexandreBelling check who is responsible for creating the directories
	// create the sub directories if they do not exist
	dirs := []string{
		cfg.Execution.DirDone(),
		cfg.Execution.DirFrom(),
		cfg.Execution.DirTo(),
		cfg.BlobDecompression.DirDone(),
		cfg.BlobDecompression.DirFrom(),
		cfg.BlobDecompression.DirTo(),
		cfg.Aggregation.DirDone(),
		cfg.Aggregation.DirFrom(),
		cfg.Aggregation.DirTo(),
	}

	// Limitless specific dirs
	limitlessDirs := []string{
		path.Join(cfg.ExecutionLimitless.MetadataDir, config.RequestsFromSubDir),
		path.Join(cfg.ExecutionLimitless.MetadataDir, config.RequestsDoneSubDir),
		path.Join(cfg.ExecutionLimitless.MetadataDir, config.RequestsToSubDir),
		path.Join(cfg.ExecutionLimitless.SharedRandomnessDir, config.RequestsFromSubDir),
		path.Join(cfg.ExecutionLimitless.SharedRandomnessDir, config.RequestsDoneSubDir),
		path.Join(cfg.ExecutionLimitless.SharedRandomnessDir, config.RequestsToSubDir),
	}

	for _, mod := range config.ALL_MODULES {
		limitlessDirs = append(limitlessDirs, path.Join(cfg.ExecutionLimitless.WitnessDir, "GL", mod, config.RequestsFromSubDir))
		limitlessDirs = append(limitlessDirs, path.Join(cfg.ExecutionLimitless.WitnessDir, "GL", mod, config.RequestsDoneSubDir))
		limitlessDirs = append(limitlessDirs, path.Join(cfg.ExecutionLimitless.WitnessDir, "GL", mod, config.RequestsToSubDir))

		limitlessDirs = append(limitlessDirs, path.Join(cfg.ExecutionLimitless.WitnessDir, "LPP", mod, config.RequestsFromSubDir))
		limitlessDirs = append(limitlessDirs, path.Join(cfg.ExecutionLimitless.WitnessDir, "LPP", mod, config.RequestsDoneSubDir))
		limitlessDirs = append(limitlessDirs, path.Join(cfg.ExecutionLimitless.WitnessDir, "LPP", mod, config.RequestsToSubDir))

		limitlessDirs = append(limitlessDirs, path.Join(cfg.ExecutionLimitless.SubproofsDir, "GL", mod, config.RequestsFromSubDir))
		limitlessDirs = append(limitlessDirs, path.Join(cfg.ExecutionLimitless.SubproofsDir, "GL", mod, config.RequestsDoneSubDir))
		limitlessDirs = append(limitlessDirs, path.Join(cfg.ExecutionLimitless.SubproofsDir, "GL", mod, config.RequestsToSubDir))

		limitlessDirs = append(limitlessDirs, path.Join(cfg.ExecutionLimitless.SubproofsDir, "LPP", mod, config.RequestsFromSubDir))
		limitlessDirs = append(limitlessDirs, path.Join(cfg.ExecutionLimitless.SubproofsDir, "LPP", mod, config.RequestsDoneSubDir))
		limitlessDirs = append(limitlessDirs, path.Join(cfg.ExecutionLimitless.SubproofsDir, "LPP", mod, config.RequestsToSubDir))

		limitlessDirs = append(limitlessDirs, path.Join(cfg.ExecutionLimitless.CommitsDir, mod, config.RequestsFromSubDir))
		limitlessDirs = append(limitlessDirs, path.Join(cfg.ExecutionLimitless.CommitsDir, mod, config.RequestsDoneSubDir))
		limitlessDirs = append(limitlessDirs, path.Join(cfg.ExecutionLimitless.CommitsDir, mod, config.RequestsToSubDir))
	}

	dirs = append(dirs, limitlessDirs...)
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			logrus.Fatalf("could not create the directory %s : %v", dir, err)
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
