/*
Copyright © 2023 Consensys
*/
package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	fS3Bucket string // s3Bucket is the name of the s3 bucket where assets will be copied
	fManifest string // manifest is the path to the manifest file
)

// copyCmd represents the copy command
var copyCmd = &cobra.Command{
	Use:   "copy",
	Short: "Copies assets into destination s3 bucket",
	Run:   copyAssets,
}

func copyAssets(cmd *cobra.Command, args []string) {
	if err := fileExists(fManifest); err != nil {
		checkError(err)
	}

	// read manifest
	var manifest Manifest
	err := manifest.Read(fManifest)
	checkError(err)

	// ensure that all files in manifest exist
	baseDir := filepath.Dir(fManifest)
	for _, file := range manifest.Files {
		err = fileExists(filepath.Join(baseDir, file))
		checkError(err)
	}

	// copy files to s3
	keyPrefix := manifest.Key()
	s3Client := getAWSS3Client()

	for _, file := range manifest.Files {
		key := keyPrefix + "/" + file
		err := uploadToS3(s3Client, fS3Bucket, key, filepath.Join(baseDir, file))
		checkError(err)
	}

	color.Green("done ✅")

}

func fileExists(path string) error {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("file %s does not exist", path)
	}
	return nil
}

func init() {
	rootCmd.AddCommand(copyCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// copyCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// copyCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	copyCmd.Flags().StringVar(&fS3Bucket, "s3-bucket", "linea-prover-assets", "Name of the s3 bucket where assets will be copied")
	copyCmd.Flags().StringVarP(&fManifest, "manifest", "m", "manifest.json", "Path to the manifest file")
}
