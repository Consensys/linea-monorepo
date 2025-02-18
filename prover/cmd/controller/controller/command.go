package controller

import (
	"context"
	"os"

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

	// Disable Legacy for testing
	cfg.Controller.EnableExecution = true
	cfg.Controller.EnableBlobDecompression = false
	cfg.Controller.EnableAggregation = false

	cfg.Controller.EnableExecBootstrap = false
	cfg.Controller.EnableExecGL = false
	cfg.Controller.EnableExecRndBeacon = false
	cfg.Controller.EnableExecLPP = false
	cfg.Controller.EnableExecConglomeration = false

	// TODO @gbotrel @AlexandreBelling check who is responsible for creating the directories
	// create the sub directories if they do not exist
	dirs := []string{
		cfg.Execution.DirDone(0),
		cfg.Execution.DirFrom(0),
		cfg.Execution.DirTo(0),
		cfg.BlobDecompression.DirDone(0),
		cfg.BlobDecompression.DirFrom(0),
		cfg.BlobDecompression.DirTo(0),
		cfg.Aggregation.DirDone(0),
		cfg.Aggregation.DirFrom(0),
		cfg.Aggregation.DirTo(0),

		// Dirs. for Limitless controller
		cfg.ExecBootstrap.DirFrom(0),
		cfg.ExecBootstrap.DirDone(0),
		cfg.ExecBootstrap.DirTo(0),
		cfg.ExecBootstrap.DirTo(1),

		cfg.ExecGL.DirFrom(0),
		cfg.ExecGL.DirDone(0),
		cfg.ExecGL.DirTo(0),
		cfg.ExecGL.DirTo(1),

		cfg.ExecRndBeacon.DirFrom(0),
		cfg.ExecRndBeacon.DirFrom(1),
		cfg.ExecRndBeacon.DirDone(0),
		cfg.ExecRndBeacon.DirDone(1),
		cfg.ExecRndBeacon.DirTo(0),

		cfg.ExecLPP.DirFrom(0),
		cfg.ExecLPP.DirDone(0),
		cfg.ExecLPP.DirTo(0),

		cfg.ExecConglomeration.DirFrom(0),
		cfg.ExecConglomeration.DirFrom(1),
		cfg.ExecConglomeration.DirFrom(2),
		cfg.ExecConglomeration.DirDone(0),
		cfg.ExecConglomeration.DirDone(1),
		cfg.ExecConglomeration.DirDone(2),
		cfg.ExecConglomeration.DirTo(0),
	}

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
