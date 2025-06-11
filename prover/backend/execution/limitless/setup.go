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

// SerAssestAndWrite serializes and writes prover assets to files.
func SerAssestAndWrite(config *config.Config) error {
	if config == nil {
		return fmt.Errorf("config is nil")
	}

	filePath := config.PathforLimitlessProverAssets()
	if err := os.MkdirAll(filePath, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", filePath, err)
	}

	var (
		zkevm = test_utils.GetZkEVM()
		disc  = &distributed.StandardModuleDiscoverer{
			TargetWeight: 1 << 28,
			Affinities:   test_utils.GetAffinities(zkevm),
			Predivision:  1,
		}

		distWizardRaw = distributed.DistributeWizard(zkevm.WizardIOP, disc)
		// distWizardCompiled = distWizardRaw.CompileSegments()
		// distWizard         = distWizardCompiled.Conglomerate(20)
	)

	// Define assets to serialize and write
	assets := []struct {
		name   string
		object interface{}
	}{
		// {name: "zkevm.bin", object: zkevm},
		// {name: "disc.bin", object: disc},
		{name: "dw-raw.bin", object: distWizardRaw},
		// {name: "dw-compiled.bin", object: distWizardCompiled},
		// {name: "dw.bin", object: distWizard},
	}

	// Serialize and write each asset
	reader := bytes.NewReader(nil) // Initialize with nil slice
	var writtenFiles []string
	for _, asset := range assets {
		data, err := serialization.Serialize(asset.object)
		if err != nil {
			return fmt.Errorf("failed to serialize %s: %w", asset.name, err)
		}
		reader.Reset(data)
		assetPath := path.Join(filePath, asset.name)
		if err := utils.WriteToFile(assetPath, reader); err != nil {
			return fmt.Errorf("failed to write %s: %w", assetPath, err)
		}
		writtenFiles = append(writtenFiles, assetPath)
		runtime.GC()
	}

	logrus.Infof("Written limitless prover assets to %v", writtenFiles)
	return nil
}
