package limitless

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/config"
)

// TestSerAndWriteAssets tests the serialization and writing of assets and compiled files
func TestSerAndWriteAssets(t *testing.T) {
	cfg, err := config.NewConfigFromFile("/home/ubuntu/linea-monorepo/prover/config/config-sepolia-full.toml")
	if err != nil {
		t.Fatalf("could not get the config: %v", err)
	}

	err = SerializeAndWriteAssets(cfg)
	if err != nil {
		t.Fatalf("could not serialize and write the assets: %v", err)
	}

	t.Logf("Successfully serialized and wrote the assets")
}

// TestReadAndDeserAssets tests the reading and deserialization of assets and compiled files
func TestReadAndDeserAssets(t *testing.T) {
	cfg, err := config.NewConfigFromFile("/home/ubuntu/linea-monorepo/prover/config/config-sepolia-full.toml")
	if err != nil {
		t.Fatalf("could not get the config: %v", err)
	}

	_, err = ReadAndDeserAssets(cfg)
	if err != nil {
		t.Fatalf("could not read and deserialize the assets: %v", err)
	}

	t.Logf("Successfully read and deserialized the assets")
	//t.Logf("Dist wizard module names: %v", assets.DistWizard.ModuleNames)
}
