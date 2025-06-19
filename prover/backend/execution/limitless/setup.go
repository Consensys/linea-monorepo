package limitless

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"runtime"

	"github.com/consensys/linea-monorepo/prover/backend/execution"
	"github.com/consensys/linea-monorepo/prover/circuits"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/serialization"
	"github.com/consensys/linea-monorepo/prover/utils"
	utils_limitless "github.com/consensys/linea-monorepo/prover/utils/limitless"
	"github.com/consensys/linea-monorepo/prover/zkevm"
	"github.com/sirupsen/logrus"
)

// SerAsset represents a file name and the object to be serialized
type SerAsset struct {
	Name   string
	Object interface{}
}

// Unified function to serialize and write all assets and compiled files
func SerializeAndWriteAssets(config *config.Config) error {
	if config == nil {
		return fmt.Errorf("config is nil")
	}

	// Create directory for assets
	filePath := config.PathforLimitlessProverAssets()
	if err := os.MkdirAll(filePath, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", filePath, err)
	}

	// Shared initialization for zkevm and disc
	var (
		zkevm = zkevm.FullZKEVMWithSuite(&config.TracesLimits, zkevm.CompilationSuite{}, config)
		disc  = &distributed.StandardModuleDiscoverer{
			TargetWeight: 1 << 28,
			Affinities:   utils_limitless.GetAffinities(zkevm),
			Predivision:  1,
		}
		dwRaw = distributed.DistributeWizard(zkevm.WizardIOP, disc)
	)

	// Serialize and write initial assets
	initialAssets := []SerAsset{
		{Name: "zkevm.bin", Object: zkevm},
		{Name: "disc.bin", Object: disc},
		{Name: "dw-raw.bin", Object: dwRaw},
	}
	if err := serializeAssets(filePath, initialAssets); err != nil {
		return err
	}

	// Clean up memory
	zkevm = nil
	disc = nil

	// Compile distributed wizard
	logrus.Info("Starting to compile distributed wizard")
	dw := dwRaw.CompileSegments().Conglomerate(20)
	logrus.Info("Finished compiling distributed wizard")

	dwRaw = nil

	// Serialize and write compiled default module
	defaultAsset := []SerAsset{
		{Name: "dw-compiled-default.bin", Object: dw.CompiledDefault},
	}
	if err := serializeAssets(filePath, defaultAsset); err != nil {
		return err
	}
	dw.CompiledDefault = nil
	runtime.GC()

	// Serialize and write compiled GL modules
	glAssets := make([]SerAsset, len(dw.CompiledGLs))
	for i, compiledGL := range dw.CompiledGLs {
		glAssets[i] = SerAsset{
			Name:   fmt.Sprintf("dw-compiled-gl-%d.bin", i),
			Object: compiledGL,
		}
	}
	if err := serializeAssets(filePath, glAssets); err != nil {
		return err
	}
	dw.CompiledGLs = nil
	runtime.GC()

	// Serialize and write compiled LPP modules
	lppAssets := make([]SerAsset, len(dw.CompiledLPPs))
	for i, compiledLPP := range dw.CompiledLPPs {
		lppAssets[i] = SerAsset{
			Name:   fmt.Sprintf("dw-compiled-lpp-%d.bin", i),
			Object: compiledLPP,
		}
	}
	if err := serializeAssets(filePath, lppAssets); err != nil {
		return err
	}
	dw.CompiledLPPs = nil
	runtime.GC()

	// Serialize and write conglomeration compilation
	conglomerationAsset := []SerAsset{
		{Name: "dw-compiled-conglomeration.bin", Object: dw.CompiledConglomeration},
	}
	if err := serializeAssets(filePath, conglomerationAsset); err != nil {
		return err
	}
	dw.CompiledConglomeration = nil
	return nil
}

// Helper function to serialize and write assets
func serializeAssets(filePath string, assets []SerAsset) error {
	reader := bytes.NewReader(nil)
	for _, asset := range assets {
		if err := serialization.SerializeAndWrite(filePath, asset.Name, asset.Object, reader); err != nil {
			return err
		}
	}
	return nil
}

func loadCktSetupAsync(cfg *config.Config) (*circuits.Setup, chan struct{}, error) {
	var (
		setup       circuits.Setup
		errSetup    error
		chSetupDone = make(chan struct{})
	)
	go func() {
		setup, errSetup = circuits.LoadSetup(cfg, circuits.ExecutionLimitlessCircuitID)
		close(chSetupDone)
	}()

	return &setup, chSetupDone, errSetup
}

// Helper function to finalize setup and validate checksum
func finalizeCktSetup(cfg *config.Config, chSetupDone <-chan struct{},
	setup *circuits.Setup, errSetup error) error {
	<-chSetupDone
	if errSetup != nil {
		utils.Panic("could not load setup: %v", errSetup)
	}
	execution.ValidateSetupChecksum(*setup, &cfg.TracesLimits)
	return nil
}

// Helper function to clean up witness directory
func cleanWitnessDirectory(cfg *config.Config) {
	filepath := cfg.PathforLimitlessProverAssets()
	filepath = path.Join(filepath, "witness")
	os.RemoveAll(filepath)
}

func SetupLimitlessAssest(config *config.Config) *Asset {
	var (
		traceLimits = utils_limitless.GetLimitlessTraceLimits()
		zkevm       = zkevm.FullZKEVMWithSuite(traceLimits, zkevm.CompilationSuite{}, config)
		disc        = &distributed.StandardModuleDiscoverer{
			TargetWeight: 1 << 28,
			Affinities:   utils_limitless.GetAffinities(zkevm),
			Predivision:  1,
		}
		dw = distributed.DistributeWizard(zkevm.WizardIOP, disc).CompileSegments().Conglomerate(20)
	)

	return &Asset{
		Zkevm:      zkevm,
		Disc:       disc,
		DistWizard: dw,
	}
}
