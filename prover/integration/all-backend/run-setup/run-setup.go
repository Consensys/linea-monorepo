package main

import (
	"context"
	"os"
	"strings"

	"github.com/consensys/linea-monorepo/prover/cmd/prover/cmd"
	allbackend "github.com/consensys/linea-monorepo/prover/integration/all-backend"
	"github.com/consensys/linea-monorepo/prover/utils/test_utils"
	"github.com/stretchr/testify/assert"
)

func main() {
	var t test_utils.FakeTestingT
	allbackend.CdProver(t)
	args := cmd.SetupArgs{
		Force:      true,
		Circuits:   "aggregation",
		DictPath:   "./lib/compressor/compressor_dict.bin",
		AssetsDir:  "",
		ConfigFile: "./integration/all-backend/config-integration-light.toml",
	}

	if len(os.Args) > 1 {
		args.Circuits = strings.Join(os.Args[1:], ",")
	}

	assert.NoError(t, cmd.Setup(context.TODO(), args))
}
