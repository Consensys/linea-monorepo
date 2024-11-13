package main

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/consensys/linea-monorepo/prover/cmd/prover/cmd"
	allbackend "github.com/consensys/linea-monorepo/prover/integration/all-backend"
	"github.com/consensys/linea-monorepo/prover/utils/test_utils"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var flagAggregationOnly = flag.Bool("aggregation-only", true, "whether to only run aggregation")

func main() {

	flag.Parse()
	var t test_utils.FakeTestingT

	// run all execution tests
	allbackend.CdProver(t)

	const (
		testPath          = "./integration/all-backend"
		decompressionPath = testPath + "/testdata/prover-compression"
		executionPath     = testPath + "/testdata/prover-execution"
		aggregationPath   = testPath + "/testdata/prover-aggregation"
	)

	cmd.FConfigFile = "/home/ubuntu/linea-monorepo/prover/integration/all-backend/config-integration-light.toml"
	cmd.FDictPath = "./lib/compressor/compressor_dict.bin"

	runAllJsonInFolder := func(dirPath string) {
		inFolder := filepath.Join(dirPath, "requests")
		outFolder := filepath.Join(dirPath, "responses")
		dir, err := os.ReadDir(inFolder)
		require.NoError(t, err)
		for _, entry := range dir {
			if filepath.Ext(entry.Name()) != ".json" {
				logrus.Warn("skipping ", entry.Name())
				continue
			}
			cmd.FInput = filepath.Join(inFolder, entry.Name())
			cmd.FOutput = filepath.Join(outFolder, entry.Name())
			assert.NoError(t, cmd.CmdProve("prove", []string{}))
		}
	}

	if !*flagAggregationOnly {
		runAllJsonInFolder(decompressionPath)
		runAllJsonInFolder(executionPath)
	}
	runAllJsonInFolder(aggregationPath)
}
