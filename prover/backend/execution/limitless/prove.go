package limitless

import (
	"bytes"
	"fmt"
	"path"

	"github.com/consensys/linea-monorepo/prover/backend/execution"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/serialization"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/profiling"
	"github.com/consensys/linea-monorepo/prover/utils/test_utils"
	"github.com/consensys/linea-monorepo/prover/zkevm"
	"github.com/sirupsen/logrus"
)

// Asset struct to hold deserialized assets
type Asset struct {
	Zkevm      *zkevm.ZkEvm
	Disc       *distributed.StandardModuleDiscoverer
	DistWizard *distributed.DistributedWizard
}

// Prove function for the Assest struct
func (asset *Asset) Prove(cfg *config.Config, req *execution.Request) error {
	// Set MonitorParams before any proving happens
	profiling.SetMonitorParams(cfg)

	logrus.Info("Starting to run the bootstrapper")
	var (
		witness     = test_utils.GetZkevmWitness(req, cfg)
		runtimeBoot = wizard.RunProver(asset.DistWizard.Bootstrapper, asset.Zkevm.GetMainProverStep(witness))
		_, _        = distributed.SegmentRuntime(runtimeBoot, asset.DistWizard)
	)
	logrus.Info("Finished running the bootstrapper")

	return nil
}

// Unified function to read and deserialize all assets and compiled files
func ReadAndDeser(config *config.Config) (*Asset, error) {
	if config == nil {
		return nil, fmt.Errorf("config is nil")
	}

	filePath := config.PathforLimitlessProverAssets()
	var readBuf bytes.Buffer

	// Initialize result struct
	assets := &Asset{
		Zkevm:      &zkevm.ZkEvm{},
		Disc:       &distributed.StandardModuleDiscoverer{},
		DistWizard: &distributed.DistributedWizard{},
	}

	// Define all files to read and deserialize
	files := []struct {
		name   string
		target any
	}{
		{name: "zkevm.bin", target: &assets.Zkevm},
		{name: "disc.bin", target: &assets.Disc},
		{name: "dw-raw.bin", target: &assets.DistWizard},
		{name: "dw-compiled-default.bin", target: &assets.DistWizard.CompiledDefault},
	}

	// Read and deserialize each file
	for _, file := range files {
		if err := readAndDeserialize(filePath, file.name, file.target, &readBuf); err != nil {
			return nil, err
		}
	}

	// Read and deserialize GL modules
	for i := 0; i < len(assets.DistWizard.GLs); i++ {
		var compiledGL *distributed.RecursedSegmentCompilation
		fileName := fmt.Sprintf("dw-compiled-gl-%d.bin", i)
		if err := readAndDeserialize(filePath, fileName, &compiledGL, &readBuf); err != nil {
			return nil, err
		}
		assets.DistWizard.CompiledGLs = append(assets.DistWizard.CompiledGLs, compiledGL)
	}

	// Read and deserialize LPP modules
	for i := 0; i < len(assets.DistWizard.LPPs); i++ {
		var compiledLPP *distributed.RecursedSegmentCompilation
		fileName := fmt.Sprintf("dw-compiled-lpp-%d.bin", i)
		if err := readAndDeserialize(filePath, fileName, &compiledLPP, &readBuf); err != nil {
			return nil, err
		}
		assets.DistWizard.CompiledLPPs = append(assets.DistWizard.CompiledLPPs, compiledLPP)
	}

	// Read and deserialize conglomeration compilation
	if err := readAndDeserialize(filePath, "dw-compiled-conglomeration.bin", &assets.DistWizard.CompiledConglomeration, &readBuf); err != nil {
		return nil, err
	}

	return assets, nil
}

// Helper function to read and deserialize an object from a file
func readAndDeserialize(filePath string, fileName string, target any, readBuf *bytes.Buffer) error {
	readBuf.Reset()
	fullPath := path.Join(filePath, fileName)
	if err := utils.ReadFromFile(fullPath, readBuf); err != nil {
		return fmt.Errorf("failed to read %s: %w", fullPath, err)
	}
	if err := serialization.Deserialize(readBuf.Bytes(), target); err != nil {
		return fmt.Errorf("failed to deserialize %s: %w", fileName, err)
	}
	logrus.Infof("Read and deserialized %s from %s", fileName, fullPath)
	return nil
}
