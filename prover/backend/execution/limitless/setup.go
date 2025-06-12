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
		{name: "zkevm.bin", object: zkevm},
		{name: "disc.bin", object: disc},
		{name: "dw-raw.bin", object: distWizardRaw},
	}

	// Serialize and write each asset
	reader := bytes.NewReader(nil) // Initialize with nil slice
	var writtenFiles []string
	for _, asset := range assets {

		switch asset.name {
		case "dw-compiled.bin":
			var readBuf bytes.Buffer
			dwRawPath := path.Join(filePath, "dw-raw.bin")
			if err := utils.ReadFromFile(dwRawPath, &readBuf); err != nil {
				return fmt.Errorf("failed to read %s: %w", dwRawPath, err)
			}

			var dwRaw *distributed.DistributedWizard
			if err := serialization.Deserialize(readBuf.Bytes(), &dwRaw); err != nil {
				return fmt.Errorf("failed to deserialize %s: %w", dwRawPath, err)
			}

			asset.object = dwRaw.CompileSegments()
		case "dw.bin":
			var readBuf bytes.Buffer
			dwCompiledPath := path.Join(filePath, "dw-compiled.bin")
			if err := utils.ReadFromFile(dwCompiledPath, &readBuf); err != nil {
				return fmt.Errorf("failed to read %s: %w", dwCompiledPath, err)
			}
			var dwCompiled *distributed.DistributedWizard
			if err := serialization.Deserialize(readBuf.Bytes(), &dwCompiled); err != nil {
				return fmt.Errorf("failed to deserialize %s: %w", dwCompiledPath, err)
			}
			asset.object = dwCompiled.Conglomerate(20)
		}

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
