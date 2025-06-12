package limitless

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"runtime"

	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/serialization"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/test_utils"
	"github.com/sirupsen/logrus"
)

// SerAsset represents a file name and the object to be serialized
type SerAsset struct {
	Name   string
	Object interface{}
}

// Unified function to serialize and write all assets and compiled files
func SerializeAndWrite(config *config.Config) error {
	if config == nil {
		return fmt.Errorf("config is nil")
	}

	filePath := config.PathforLimitlessProverAssets()
	if err := os.MkdirAll(filePath, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", filePath, err)
	}

	// Shared initialization for zkevm and disc
	var (
		zkevm = test_utils.GetZkEVM()
		disc  = &distributed.StandardModuleDiscoverer{
			TargetWeight: 1 << 28,
			Affinities:   test_utils.GetAffinities(zkevm),
			Predivision:  1,
		}

		dwRaw = distributed.DistributeWizard(zkevm.WizardIOP, disc)
	)

	// Define all assets to serialize and write
	assets := []SerAsset{
		{Name: "zkevm.bin", Object: zkevm},
		{Name: "disc.bin", Object: disc},
		{Name: "dw-raw.bin", Object: dwRaw},
	}

	// Compile distributed wizard
	logrus.Info("Starting to compile distributed wizard")
	dw := dwRaw.CompileSegments().Conglomerate(20)
	logrus.Info("Finished compiling distributed wizard")

	compiledDefault := dw.CompiledDefault
	compiledGLs := dw.CompiledGLs
	compiledLPPs := dw.CompiledLPPs
	congo := dw.CompiledConglomeration
	dw = nil

	runtime.GC()

	// Add compiled default module to assets
	assets = append(assets, SerAsset{
		Name:   "dw-compiled-default.bin",
		Object: compiledDefault,
	})

	// Add compiled GL modules to assets
	for i, compiledGL := range compiledGLs {
		assets = append(assets, SerAsset{
			Name:   fmt.Sprintf("dw-compiled-gl-%d.bin", i),
			Object: compiledGL,
		})
	}

	// Add compiled LPP modules to assets
	for i, compiledLPP := range compiledLPPs {
		assets = append(assets, SerAsset{
			Name:   fmt.Sprintf("dw-compiled-lpp-%d.bin", i),
			Object: compiledLPP,
		})
	}

	// Add conglomeration compilation to assets
	assets = append(assets, SerAsset{
		Name:   "dw-compiled-conglomeration.bin",
		Object: congo,
	})

	// Initialize with nil slice
	reader := bytes.NewReader(nil)

	// Serialize and write each asset
	for _, asset := range assets {
		if err := serializeAndWrite(filePath, asset.Name, asset.Object, reader); err != nil {
			return err
		}
	}

	return nil
}

// Helper function to serialize and write an object to a file
func serializeAndWrite(filePath string, fileName string, object any, reader *bytes.Reader) error {
	data, err := serialization.Serialize(object)
	if err != nil {
		return fmt.Errorf("failed to serialize %s: %w", fileName, err)
	}
	reader.Reset(data)
	fullPath := path.Join(filePath, fileName)
	if err := utils.WriteToFile(fullPath, reader); err != nil {
		return fmt.Errorf("failed to write %s: %w", fullPath, err)
	}
	logrus.Infof("Written %s to %s", fileName, fullPath)
	return nil
}
