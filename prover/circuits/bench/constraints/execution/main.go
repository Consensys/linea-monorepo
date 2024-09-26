package main

import (
	"path/filepath"

	"github.com/consensys/zkevm-monorepo/prover/circuits/execution"
	"github.com/consensys/zkevm-monorepo/prover/config"
	"github.com/consensys/zkevm-monorepo/prover/utils"
	"github.com/consensys/zkevm-monorepo/prover/utils/test_utils"
	"github.com/consensys/zkevm-monorepo/prover/zkevm"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func main() {
	var t test_utils.FakeTestingT
	root, err := utils.GetRepoRootPath()
	assert.NoError(t, err)
	root = filepath.Join(root, "prover")

	viper.Set("assets_dir", filepath.Join(root, "prover-assets"))
	// TODO Make sure this is the correct config file
	cfg, err := config.NewConfigFromFile(filepath.Join(root, "config", "config-integration-full.toml"))
	assert.NoError(t, err)

	//extraFlags["cfg_checksum"] = limits.Checksum()
	// The builder itself will create a profile called "profiling-execution.pprof"
	_, err = execution.NewBuilder(
		zkevm.FullZkEvm(&cfg.TracesLimits), // or TracesLimitsLarge?
	).Compile()
	assert.NoError(t, err)
}
