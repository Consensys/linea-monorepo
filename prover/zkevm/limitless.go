package zkevm

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/consensys/linea-monorepo/prover/backend/files"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/serialization"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/publicInput"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

var (
	bootstrapperFile       = "dw-bootstrapper.bin"
	discFile               = "disc.bin"
	zkevmFile              = "zkevm-wiop.bin"
	compiledDefaultFile    = "dw-compiled-default.bin"
	blueprintGLPrefix      = "dw-blueprint-gl"
	blueprintLppPrefix     = "dw-blueprint-lpp"
	blueprintGLTemplate    = blueprintGLPrefix + "-%d.bin"
	blueprintLppTemplate   = blueprintLppPrefix + "-%d.bin"
	compileLppTemplate     = "dw-compiled-lpp-%d.bin"
	compileGlTemplate      = "dw-compiled-gl-%d.bin"
	conglomerationFile     = "dw-compiled-conglomeration.bin"
	executionLimitlessPath = "execution-limitless"
)

// GetTestZkEVM returns a ZkEVM object configured for testing.
func GetTestZkEVM() *ZkEvm {
	return FullZKEVMWithSuite(config.GetTestTracesLimits(), CompilationSuite{}, &config.Config{})
}

// LimitlessZkEVM defines the wizard responsible for proving execution of the EVM
// and the associated wizard circuits for the limitless prover protocol.
type LimitlessZkEVM struct {
	Zkevm      *ZkEvm
	DistWizard *distributed.DistributedWizard
}

// NewLimitlessZkEVM returns a new LimitlessZkEVM object.
func NewLimitlessZkEVM(cfg *config.Config) *LimitlessZkEVM {
	var (
		traceLimits = cfg.TracesLimits
		zkevm       = FullZKEVMWithSuite(&traceLimits, CompilationSuite{}, cfg)
		disc        = &distributed.StandardModuleDiscoverer{
			TargetWeight: 1 << 28,
			Affinities:   GetAffinities(zkevm),
			Predivision:  1,
		}
		dw = distributed.DistributeWizard(zkevm.WizardIOP, disc)
	)

	// These are the slow and expensive operations.
	// dw.CompileSegments().Conglomerate(20)

	// decorateWithPublicInputs(dw.CompiledConglomeration)

	return &LimitlessZkEVM{
		Zkevm:      zkevm,
		DistWizard: dw,
	}
}

// Store writes the limitless prover zkevm into disk in the folder given by
// [cfg.PathforLimitlessProverAssets].
func (lz *LimitlessZkEVM) Store(cfg *config.Config) error {

	// asset is a utility struct used to list the object and the file name
	type asset struct {
		Name   string
		Object any
	}

	if cfg == nil {
		utils.Panic("config is nil")
	}

	// Create directory for assets
	assetDir := cfg.PathForSetup(executionLimitlessPath)
	if err := os.MkdirAll(assetDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", assetDir, err)
	}

	assets := []asset{
		{
			Name:   zkevmFile,
			Object: lz.Zkevm,
		},
		{
			Name: discFile,
			// alex: the conversion is needed because we figured that the
			// serialization was not working well when attempting with the
			// interface object. The reason why is not clear yet, but it works
			// this way.
			Object: *lz.DistWizard.Disc.(*distributed.StandardModuleDiscoverer),
		},
		{
			Name:   bootstrapperFile,
			Object: lz.DistWizard.Bootstrapper,
		},
		{
			Name:   compiledDefaultFile,
			Object: lz.DistWizard.CompiledDefault,
		},
	}

	for i, modGl := range lz.DistWizard.CompiledGLs {
		assets = append(assets, asset{
			Name:   fmt.Sprintf(compileGlTemplate, i),
			Object: modGl,
		})
	}

	for i, blueprintGL := range lz.DistWizard.BlueprintGLs {
		assets = append(assets, asset{
			Name:   fmt.Sprintf(blueprintGLTemplate, i),
			Object: blueprintGL,
		})
	}

	for i, modLpp := range lz.DistWizard.CompiledLPPs {
		assets = append(assets, asset{
			Name:   fmt.Sprintf(compileLppTemplate, i),
			Object: modLpp,
		})
	}

	for i, blueprintLPP := range lz.DistWizard.BlueprintLPPs {
		assets = append(assets, asset{
			Name:   fmt.Sprintf(blueprintLppTemplate, i),
			Object: blueprintLPP,
		})
	}

	assets = append(assets, struct {
		Name   string
		Object any
	}{
		Name:   conglomerationFile,
		Object: lz.DistWizard.CompiledConglomeration,
	})

	for _, asset := range assets {
		logrus.Infof("writing %s to disk", asset.Name)
		if err := writeToDisk(assetDir, asset.Name, asset.Object); err != nil {
			return err
		}
	}

	logrus.Info("limitless prover assets written to disk")
	return nil
}

func loadFromFile(assetFilePath string, obj any) error {

	logrus.Infof("Loading %s\n", assetFilePath)

	var (
		f        = files.MustRead(assetFilePath)
		buf, err = io.ReadAll(f)
	)

	if err != nil {
		return fmt.Errorf("could not read file %s: %w", assetFilePath, err)
	}

	if err := serialization.Deserialize(buf, obj); err != nil {
		return fmt.Errorf("could not deserialize file %s: %w", assetFilePath, err)
	}

	return nil
}

// LoadBootstrapperAsync loads the bootstrapper from disk.
func (lz *LimitlessZkEVM) LoadBootstrapper(cfg *config.Config) error {
	if lz.DistWizard == nil {
		lz.DistWizard = &distributed.DistributedWizard{}
	}
	return loadFromFile(cfg.PathForSetup(executionLimitlessPath)+"/"+bootstrapperFile, &lz.DistWizard.Bootstrapper)
}

// LoadZkEVM loads the zkevm from disk
func (lz *LimitlessZkEVM) LoadZkEVM(cfg *config.Config) error {
	return loadFromFile(cfg.PathForSetup(executionLimitlessPath)+"/"+zkevmFile, &lz.Zkevm)
}

// LoadDisc loads the discoverer from disk
func (lz *LimitlessZkEVM) LoadDisc(cfg *config.Config) error {
	if lz.DistWizard == nil {
		lz.DistWizard = &distributed.DistributedWizard{}
	}

	// The discoverer is not directly deserialized as an interface object as we
	// figured that it does not work very well and the reason is unclear. This
	// conversion step is a workaround for the problem.
	res := &distributed.StandardModuleDiscoverer{}

	err := loadFromFile(cfg.PathForSetup(executionLimitlessPath)+"/"+discFile, res)
	if err != nil {
		return err
	}

	lz.DistWizard.Disc = res
	return nil
}

// LoadBlueprints loads the segmentation blueprints from disk for all the modules
// LPP and GL.
func (lz *LimitlessZkEVM) LoadBlueprints(cfg *config.Config) error {

	var (
		assetDir        = cfg.PathForSetup(executionLimitlessPath)
		cntLpps, cntGLs int
	)

	if lz.DistWizard == nil {
		lz.DistWizard = &distributed.DistributedWizard{}
	}

	files, err := os.ReadDir(assetDir)
	if err != nil {
		return fmt.Errorf("could not read directory %s: %w", assetDir, err)
	}

	for _, file := range files {

		if strings.HasPrefix(file.Name(), blueprintGLPrefix) {
			cntGLs++
		}

		if strings.HasPrefix(file.Name(), blueprintLppPrefix) {
			cntLpps++
		}
	}

	lz.DistWizard.BlueprintGLs = make([]distributed.ModuleSegmentationBlueprint, cntGLs)
	lz.DistWizard.BlueprintLPPs = make([]distributed.ModuleSegmentationBlueprint, cntLpps)

	eg := &errgroup.Group{}

	for i := 0; i < cntGLs; i++ {
		eg.Go(func() error {
			filePath := path.Join(assetDir, fmt.Sprintf(blueprintGLTemplate, i))
			if err := loadFromFile(filePath, &lz.DistWizard.BlueprintGLs[i]); err != nil {
				return err
			}
			return nil
		})
	}

	for i := 0; i < cntLpps; i++ {
		eg.Go(func() error {
			filePath := path.Join(assetDir, fmt.Sprintf(blueprintLppTemplate, i))
			if err := loadFromFile(filePath, &lz.DistWizard.BlueprintLPPs[i]); err != nil {
				return err
			}
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	return nil
}

// writeToDisk writes the provided assets to disk using the
// [serialization.Serialize] function.
func writeToDisk(dir, fileName string, asset any) error {

	var (
		filepath = path.Join(dir, fileName)
		f        = files.MustOverwrite(filepath)
	)

	defer f.Close()

	buf, serr := serialization.Serialize(asset)
	if serr != nil {
		return fmt.Errorf("could not serialize %s: %w", filepath, serr)
	}

	if _, werr := f.Write(buf); werr != nil {
		return fmt.Errorf("could not write to file %s: %w", filepath, werr)
	}

	return nil
}

// GetAffinities returns a list of affinities for the following modules. This
// affinities regroup how the modules are grouped.
//
//	ecadd / ecmul / ecpairing
//	hub / hub.scp / hub.acp
//	everything related to keccak
func GetAffinities(z *ZkEvm) [][]column.Natural {

	return [][]column.Natural{
		{
			z.Ecmul.AlignedGnarkData.IsActive.(column.Natural),
			z.Ecadd.AlignedGnarkData.IsActive.(column.Natural),
			z.Ecpair.AlignedFinalExpCircuit.IsActive.(column.Natural),
			z.Ecpair.AlignedG2MembershipData.IsActive.(column.Natural),
			z.Ecpair.AlignedMillerLoopCircuit.IsActive.(column.Natural),
		},
		{
			z.WizardIOP.Columns.GetHandle("hub.HUB_STAMP").(column.Natural),
			z.WizardIOP.Columns.GetHandle("hub.scp_ADDRESS_HI").(column.Natural),
			z.WizardIOP.Columns.GetHandle("hub.acp_ADDRESS_HI").(column.Natural),
			z.WizardIOP.Columns.GetHandle("hub.ccp_HUB_STAMP").(column.Natural),
			z.WizardIOP.Columns.GetHandle("hub.envcp_HUB_STAMP").(column.Natural),
			z.WizardIOP.Columns.GetHandle("hub.stkcp_PEEK_AT_STACK_POW_4").(column.Natural),
		},
		{
			z.WizardIOP.Columns.GetHandle("KECCAK_IMPORT_PAD_HASH_NUM").(column.Natural),
			z.WizardIOP.Columns.GetHandle("CLEANING_KECCAK_CleanLimb").(column.Natural),
			z.WizardIOP.Columns.GetHandle("DECOMPOSITION_KECCAK_Decomposed_Len_0").(column.Natural),
			z.WizardIOP.Columns.GetHandle("KECCAK_FILTERS_SPAGHETTI").(column.Natural),
			z.WizardIOP.Columns.GetHandle("LANE_KECCAK_Lane").(column.Natural),
			z.WizardIOP.Columns.GetHandle("KECCAKF_IS_ACTIVE_").(column.Natural),
			z.WizardIOP.Columns.GetHandle("KECCAKF_BLOCK_BASE_2_0").(column.Natural),
			z.WizardIOP.Columns.GetHandle("KECCAK_OVER_BLOCKS_TAGS_0").(column.Natural),
			z.WizardIOP.Columns.GetHandle("HASH_OUTPUT_Hash_Lo").(column.Natural),
		},
		{
			z.WizardIOP.Columns.GetHandle("SHA2_IMPORT_PAD_HASH_NUM").(column.Natural),
			z.WizardIOP.Columns.GetHandle("DECOMPOSITION_SHA2_Decomposed_Len_0").(column.Natural),
			z.WizardIOP.Columns.GetHandle("LENGTH_CONSISTENCY_SHA2_BYTE_LEN_0_0").(column.Natural),
			z.WizardIOP.Columns.GetHandle("SHA2_FILTERS_SPAGHETTI").(column.Natural),
			z.WizardIOP.Columns.GetHandle("LANE_SHA2_Lane").(column.Natural),
			z.WizardIOP.Columns.GetHandle("Coefficient_SHA2").(column.Natural),
			z.WizardIOP.Columns.GetHandle("SHA2_OVER_BLOCK_IS_ACTIVE").(column.Natural),
			z.WizardIOP.Columns.GetHandle("SHA2_OVER_BLOCK_SHA2_COMPRESSION_CIRCUIT_IS_ACTIVE").(column.Natural),
		},
		{
			z.WizardIOP.Columns.GetHandle("mmio.CN_ABC").(column.Natural),
			z.WizardIOP.Columns.GetHandle("mmio.MMIO_STAMP").(column.Natural),
			z.WizardIOP.Columns.GetHandle("mmu.STAMP").(column.Natural),
		},
	}
}

// decorateWithPublicInputs decorates the [LimitlessZkEVM] with the public inputs from
// the initial zkevm.
func decorateWithPublicInputs(cong *distributed.ConglomeratorCompilation) {

	publicInputList := []string{
		publicInput.DataNbBytes,
		publicInput.DataChecksum,
		publicInput.L2MessageHash,
		publicInput.InitialStateRootHash,
		publicInput.FinalStateRootHash,
		publicInput.InitialBlockNumber,
		publicInput.FinalBlockNumber,
		publicInput.InitialBlockTimestamp,
		publicInput.FinalBlockTimestamp,
		publicInput.FirstRollingHashUpdate_0,
		publicInput.FirstRollingHashUpdate_1,
		publicInput.LastRollingHashUpdate_0,
		publicInput.LastRollingHashUpdate_1,
		publicInput.FirstRollingHashUpdateNumber,
		publicInput.LastRollingHashNumberUpdate,
		publicInput.ChainID,
		publicInput.NBytesChainID,
		publicInput.L2MessageServiceAddrHi,
		publicInput.L2MessageServiceAddrLo,
	}

	for _, name := range publicInputList {
		cong.BubbleUpPublicInput(name)
	}
}
