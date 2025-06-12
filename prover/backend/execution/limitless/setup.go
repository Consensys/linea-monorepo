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

func SerAndWriteCompiledSeg(config *config.Config) error {
	if config == nil {
		return fmt.Errorf("config is nil")
	}

	filePath := config.PathforLimitlessProverAssets()
	if err := os.MkdirAll(filePath, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", filePath, err)
	}

	// dwRawPath := path.Join(filePath, "dw-raw.bin")
	// var readBuf bytes.Buffer
	// if err := utils.ReadFromFile(dwRawPath, &readBuf); err != nil {
	// 	return fmt.Errorf("failed to read %s: %w", dwRawPath, err)
	// }

	// var dwRaw *distributed.DistributedWizard
	// if err := serialization.Deserialize(readBuf.Bytes(), &dwRaw); err != nil {
	// 	return fmt.Errorf("failed to deserialize %s: %w", dwRawPath, err)
	// }
	// readBuf.Reset()

	var (
		zkevm = test_utils.GetZkEVM()
		disc  = &distributed.StandardModuleDiscoverer{
			TargetWeight: 1 << 28,
			Affinities:   test_utils.GetAffinities(zkevm),
			Predivision:  1,
		}
	)

	logrus.Info("Starting to compile distributed wizard")
	dw := distributed.DistributeWizard(zkevm.WizardIOP, disc).CompileSegments().Conglomerate(20)
	logrus.Info("Finished compiling distributed wizard")

	compiledDefault := dw.CompiledDefault
	compiledGLs := dw.CompiledGLs
	compiledLPPs := dw.CompiledLPPs
	congo := dw.CompiledConglomeration
	dw = nil

	runtime.GC()

	// Serialize compiled default module and write to file
	compDefBytes, err := serialization.Serialize(compiledDefault)
	if err != nil {
		return fmt.Errorf("failed to serialize compiled default module: %w", err)
	}

	compDefPath := path.Join(filePath, "dw-compiled-default.bin")
	reader := bytes.NewReader(compDefBytes)
	if err := utils.WriteToFile(compDefPath, reader); err != nil {
		return fmt.Errorf("failed to write %s: %w", compDefPath, err)
	}

	logrus.Infof("Dist. wizard compiled default module written to %s", compDefPath)

	// Serialiaze each GL module seperately and write to file
	for i, compiledGL := range compiledGLs {
		glBytes, err := serialization.Serialize(compiledGL)
		if err != nil {
			return fmt.Errorf("failed to serialize compiled GL module %d: %w", i, err)
		}
		glPath := path.Join(filePath, fmt.Sprintf("dw-compiled-gl-%d.bin", i))
		reader.Reset(glBytes)
		if err := utils.WriteToFile(glPath, reader); err != nil {
			return fmt.Errorf("failed to write %s: %w", glPath, err)
		}
		logrus.Infof("Dist. wizard compiled GL module %d written to %s", i, glPath)
	}

	// Serialize each LPP module and write to file seperately
	for i, compiledLPP := range compiledLPPs {
		lppBytes, err := serialization.Serialize(compiledLPP)
		if err != nil {
			return fmt.Errorf("failed to serialize compiled LPP module %d: %w", i, err)
		}
		lppPath := path.Join(filePath, fmt.Sprintf("dw-compiled-lpp-%d.bin", i))
		reader.Reset(lppBytes)
		if err := utils.WriteToFile(lppPath, reader); err != nil {
			return fmt.Errorf("failed to write %s: %w", lppPath, err)
		}
		logrus.Infof("Dist. Wizard compiled LPP module %d written to %s", i, lppPath)
	}

	// Serialize conglomeration compilation and write to file
	congoBytes, err := serialization.Serialize(congo)
	if err != nil {
		return fmt.Errorf("failed to serialize compiled conglomeration: %w", err)
	}
	congoPath := path.Join(filePath, "dw-compiled-conglomeration.bin")
	reader.Reset(congoBytes)
	if err := utils.WriteToFile(congoPath, reader); err != nil {
		return fmt.Errorf("failed to write %s: %w", congoPath, err)
	}
	logrus.Infof("Dist. Wizard compiled conglomeration written to %s", congoPath)

	return nil
}
