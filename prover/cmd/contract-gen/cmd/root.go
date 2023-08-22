/*
Copyright © 2023 Consensys
*/
package cmd

import (
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	fPath    string // path is the path to the directory where assets will be generated
	fNoColor bool   // noColor is a flag to disable colorized output
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "contract-gen",
	Short: "generates assets needed for proof verification - solidity smart contract, circuit, proving and verifying keys",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if fNoColor {
			color.NoColor = true
		}
	},
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
	rootCmd.PersistentFlags().StringVarP(&fPath, "path", "p", "./docker/", "Path to the directory where assets will be generated")
	rootCmd.PersistentFlags().BoolVar(&fNoColor, "no-color", false, "Disable colorized output")
}

func checkError(err error) {
	if err != nil {
		color.Red(err.Error())
		os.Exit(-1)
	}
}

func printWarning(msg string) {
	color.Yellow(msg)
}
