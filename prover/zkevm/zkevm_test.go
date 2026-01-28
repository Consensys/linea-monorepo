package zkevm

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/config"
)

func TestCanGenerateFullZkEVM(t *testing.T) {

	cfg, err := config.NewConfigFromFileUnchecked("../config/config-mainnet-limitless.toml")
	if err != nil {
		t.Fatal(err)
	}

	_ = FullZkEvm(&cfg.TracesLimits, cfg)
}
