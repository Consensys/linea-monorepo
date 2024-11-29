package prover

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/consensys/linea-monorepo/prover/cmd/prover/cmd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetup(t *testing.T) {
	assert.NoError(t, cmd.Setup(context.TODO(), cmd.SetupArgs{
		Force:      true,
		Circuits:   "public-input-interconnection",
		DictPath:   "lib/compressor/compressor_dict.bin",
		AssetsDir:  "integration/all_backend/assets",
		ConfigFile: "integration/all-backend/config-test-sepolia-v0.8.0-rc3.toml",
	}))
}

func TestProveExecution(t *testing.T) {
	args := cmd.ProverArgs{
		Large:      false,
		ConfigFile: "integration/all-backend/config-test-sepolia-v0.8.0-rc3.toml",
	}

	const baseDir = "integration/all-backend/testdata/sepolia-v0.8.0-rc3/prover-execution"
	const inDir = baseDir + "/requests"
	const outDir = baseDir + "/responses"

	files, err := os.ReadDir(inDir)
	require.NoError(t, err)

	for _, file := range files {

		if file.IsDir() || !strings.HasSuffix(file.Name(), "getZkProof.json") {
			continue
		}
		fmt.Println("processing file: ", file.Name())
		elems := strings.Split(file.Name(), "-")

		args.Input = filepath.Join(inDir, file.Name())
		args.Output = filepath.Join(outDir, fmt.Sprintf("%s-%s-getZkProof.json", elems[0], elems[1]))

		require.NoError(t, cmd.Prove(args))
	}

}

func TestProveDecompression(t *testing.T) {
	args := cmd.ProverArgs{
		Large:      false,
		ConfigFile: "integration/all-backend/config-test-sepolia-v0.8.0-rc3.toml",
	}

	const baseDir = "integration/all-backend/testdata/sepolia-v0.8.0-rc3/prover-compression"
	const inDir = baseDir + "/requests"
	const outDir = baseDir + "/responses"

	files, err := os.ReadDir(inDir)
	require.NoError(t, err)

	for _, file := range files {

		if file.IsDir() || !strings.HasSuffix(file.Name(), "getZkBlobCompressionProof.json") {
			fmt.Println("skipping file: ", file.Name())
			continue
		}
		fmt.Println("processing file: ", file.Name())

		split := strings.Split(file.Name(), "-")

		args.Input = filepath.Join(inDir, file.Name())
		args.Output = filepath.Join(outDir, strings.Join(slices.Delete(split, 2, 4), "-"))

		require.NoError(t, cmd.Prove(args))
	}

}

func TestAggregation(t *testing.T) {
	args := cmd.ProverArgs{
		Input:      "integration/all-backend/testdata/sepolia-v0.8.0-rc3/prover-aggregation/requests/4454961-4476909-645da7ecbd6d96b7c23c2a5a337b90e6060b36c0aea3957174253fef01c10f16-getZkAggregatedProof.json",
		Output:     "",
		Large:      false,
		ConfigFile: "integration/all-backend/config-test-sepolia-v0.8.0-rc3.toml",
	}

	require.NoError(t, cmd.Prove(args))
}

func TestWhichIsNumber79(t *testing.T) {
	const dir = "integration/all-backend/testdata/sepolia-v0.8.0-rc3/prover-execution/responses"
	files, err := os.ReadDir(dir)
	require.NoError(t, err)
	names := make([]string, 0, len(files))
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), "getZkProof.json") {
			continue
		}
		names = append(names, file.Name())
	}
	slices.Sort(names)
	fmt.Println("file #79 is: ", names[79])
}

func TestPrettifyExecs(t *testing.T) {
	const dir = "integration/all-backend/testdata/sepolia-v0.8.0-rc3/prover-execution/responses"
	files, err := os.ReadDir(dir)
	require.NoError(t, err)
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), "getZkProof.json") {
			continue
		}
		f, err := os.Open(filepath.Join(dir, file.Name()))
		require.NoError(t, err)
		var o any
		require.NoError(t, json.NewDecoder(f).Decode(&o))
		require.NoError(t, f.Close())
		f, err = os.OpenFile(filepath.Join(dir, file.Name()), os.O_CREATE|os.O_WRONLY, 0600)
		require.NoError(t, err)
		enc := json.NewEncoder(f)
		enc.SetIndent("", "  ")
		require.NoError(t, errors.Join(enc.Encode(o), f.Close()))
	}
}
