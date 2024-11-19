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

	cmd.FConfigFile = "/home/ubuntu/linea-monorepo/prover/integration/all-backend/config-integration-light.toml"

	cmd.FCircuits = "aggregation"
	if len(os.Args) > 1 {
		cmd.FCircuits = strings.Join(os.Args[1:], ",")
	}

	assert.NoError(t, cmd.CmdSetup("setup", context.TODO(), []string{}))
}
