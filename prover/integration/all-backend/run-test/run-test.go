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

var flagNoExecution = flag.Bool("no-execution", true, "don't create new execution proofs")
var flagNoDecompression = flag.Bool("no-decompression", false, "don't create new decompression proofs")

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

			assert.NoError(t, cmd.Prove(cmd.ProverArgs{
				Input:      filepath.Join(inFolder, entry.Name()),
				Output:     filepath.Join(outFolder, entry.Name()),
				Large:      false,
				ConfigFile: "integration/all-backend/config-integration-light.toml",
			}))
		}
	}

	if !*flagNoDecompression {
		logrus.Info("processing decompression requests")
		runAllJsonInFolder(decompressionPath)
	}
	if !*flagNoExecution {
		runAllJsonInFolder(executionPath)
	}
	runAllJsonInFolder(aggregationPath)
}
