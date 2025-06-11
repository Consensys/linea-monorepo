package cmd

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/config"
)

func TestSerAndWriteAssets(t *testing.T) {
	cfg, err := config.NewConfigFromFile("/home/ubuntu/linea-monorepo/prover/config/config-sepolia-full.toml")
	if err != nil {
		t.Errorf("could not get the config : %v", err)
	}

	err = SerAssestAndWrite(cfg)
	if err != nil {
		t.Errorf("could not write the assets : %v", err)
	}
}

func TestDeserAndReadAssets(t *testing.T) {
	cfg, err := config.NewConfigFromFile("/home/ubuntu/linea-monorepo/prover/config/config-sepolia-full.toml")
	if err != nil {
		t.Errorf("could not get the config : %v", err)
	}

	asset, err := ReadAndDeser(cfg)
	if err != nil {
		t.Errorf("could not read the assets : %v", err)
	}

	t.Logf("Successfully read and deserialized the assets")
	t.Logf("Dist wizard module names: %v", asset.DistWizard.ModuleNames)
}
