package cmd

import (
	"os"

	"github.com/consensys/gnark/logger"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var fConfigFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "prover",
	Short: "run pre-compute or compute proofs for Linea circuits",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&fConfigFile, "config", "", "config file")

	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "15:04:05", NoColor: true}
	l := zerolog.New(output).With().Timestamp().Logger()

	// Set global log level for gnark
	logger.Set(l)
}
