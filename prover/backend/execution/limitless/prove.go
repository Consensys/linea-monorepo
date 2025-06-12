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

func (asset *Assest) Prove(cfg *config.Config, req *execution.Request) error {

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

type Assest struct {
	Zkevm      *zkevm.ZkEvm
	Disc       *distributed.StandardModuleDiscoverer
	DistWizard *distributed.DistributedWizard
}

// ReadAndDeser reads and deserializes limitless prover assets from files.
func ReadAndDeser(config *config.Config) (*Assest, error) {
	if config == nil {
		return nil, fmt.Errorf("config is nil")
	}

	filePath := config.PathforLimitlessProverAssets()
	var readBuf bytes.Buffer

	// Initialize result struct
	assets := &Assest{
		Zkevm:      &zkevm.ZkEvm{},
		Disc:       &distributed.StandardModuleDiscoverer{},
		DistWizard: &distributed.DistributedWizard{},
	}

	// Define files to read and deserialize
	files := []struct {
		name   string
		target interface{}
	}{
		// {name: "zkevm.bin", target: &assets.Zkevm},
		// {name: "disc.bin", target: &assets.Disc},
		{name: "dw-raw.bin", target: &assets.DistWizard},
	}

	// Read and deserialize each file
	var readFiles []string
	for _, file := range files {
		readBuf.Reset()
		assetPath := path.Join(filePath, file.name)
		if err := utils.ReadFromFile(assetPath, &readBuf); err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", assetPath, err)
		}
		if err := serialization.Deserialize(readBuf.Bytes(), file.target); err != nil {
			return nil, fmt.Errorf("failed to deserialize %s: %w", file.name, err)
		}
		readFiles = append(readFiles, assetPath)
	}

	logrus.Infof("Read and deserialized limitless prover assets from %v", readFiles)
	return assets, nil
}

func ReadAndDeserCompiledSeg(config *config.Config) (*distributed.DistributedWizard, error) {
	if config == nil {
		return nil, fmt.Errorf("config is nil")
	}

	var dw *distributed.DistributedWizard
	filePath := config.PathforLimitlessProverAssets()
	assetPath := path.Join(filePath, "dw-raw.bin")
	var readBuf bytes.Buffer
	if err := utils.ReadFromFile(assetPath, &readBuf); err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", assetPath, err)
	}
	if err := serialization.Deserialize(readBuf.Bytes(), &dw); err != nil {
		return nil, fmt.Errorf("failed to deserialize %s: %w", assetPath, err)
	}

	readBuf.Reset()

	// Read and deser compiled Default
	var compiledDef *distributed.RecursedSegmentCompilation
	compDefPath := path.Join(filePath, "dw-compiled-default.bin")
	if err := utils.ReadFromFile(compDefPath, &readBuf); err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", compDefPath, err)
	}
	if err := serialization.Deserialize(readBuf.Bytes(), &compiledDef); err != nil {
		return nil, fmt.Errorf("failed to deserialize %s: %w", compDefPath, err)
	}
	dw.CompiledDefault = compiledDef
	compiledDef = nil

	// Read and deser. GL modules
	var compiledGLs []*distributed.RecursedSegmentCompilation
	for i := 0; i < len(dw.ModuleNames); i++ {
		var compiledGL *distributed.RecursedSegmentCompilation
		compGLPath := path.Join(filePath, fmt.Sprintf("dw-compiled-gl-%d.bin", i))
		if err := utils.ReadFromFile(compGLPath, &readBuf); err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", compGLPath, err)
		}
		if err := serialization.Deserialize(readBuf.Bytes(), &compiledGL); err != nil {
			return nil, fmt.Errorf("failed to deserialize %s: %w", compGLPath, err)
		}
		compiledGLs = append(compiledGLs, compiledGL)
	}
	dw.CompiledGLs = compiledGLs
	compiledGLs = nil

	// Read and deser LPP modules
	var compiledLPPs []*distributed.RecursedSegmentCompilation
	for i := 0; i < 3; i++ {
		var compiledLPP *distributed.RecursedSegmentCompilation
		compLPPPath := path.Join(filePath, fmt.Sprintf("dw-compiled-lpp-%d.bin", i))
		if err := utils.ReadFromFile(compLPPPath, &readBuf); err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", compLPPPath, err)
		}
		if err := serialization.Deserialize(readBuf.Bytes(), &compiledLPP); err != nil {
			return nil, fmt.Errorf("failed to deserialize %s: %w", compLPPPath, err)
		}
		compiledLPPs = append(compiledLPPs, compiledLPP)
	}
	dw.CompiledLPPs = compiledLPPs
	compiledLPPs = nil

	// Read and deser conglomeration
	var compiledConglomeration *distributed.ConglomeratorCompilation
	compConglomerationPath := path.Join(filePath, "dw-compiled-conglomeration.bin")
	if err := utils.ReadFromFile(compConglomerationPath, &readBuf); err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", compConglomerationPath, err)
	}
	if err := serialization.Deserialize(readBuf.Bytes(), &compiledConglomeration); err != nil {
		return nil, fmt.Errorf("failed to deserialize %s: %w", compConglomerationPath, err)
	}
	dw.CompiledConglomeration = compiledConglomeration

	return dw, nil
}
