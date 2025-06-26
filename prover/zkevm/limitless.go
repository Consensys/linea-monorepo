package zkevm

import (
	"fmt"
	"os"
	"path"

	"github.com/consensys/linea-monorepo/prover/backend/files"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/serialization"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// GetTestZkEVM returns a ZkEVM object configured for testing.
func GetTestZkEVM() *ZkEvm {
	return FullZKEVMWithSuite(config.GetTestTracesLimits(), CompilationSuite{}, &config.Config{})
}

// LimitlessZkEVM defines the wizard responsible for proving execution of the EVM
// and the associated wizard circuits for the limitless prover protocol.
type LimitlessZkEVM struct {
	Zkevm      *ZkEvm
	Disc       *distributed.StandardModuleDiscoverer
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
		dw = distributed.DistributeWizard(zkevm.WizardIOP, disc).CompileSegments().Conglomerate(20)
	)

	return &LimitlessZkEVM{
		Zkevm:      zkevm,
		Disc:       disc,
		DistWizard: dw,
	}
}

// Store writes the limitless prover zkevm into disk in the folder given by
// [cfg.PathforLimitlessProverAssets].
func (lz *LimitlessZkEVM) Store(cfg *config.Config) error {

	if cfg == nil {
		utils.Panic("config is nil")
	}

	// Create directory for assets
	assetDir := cfg.PathforLimitlessProverAssets()
	if err := os.MkdirAll(assetDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", assetDir, err)
	}

	assets := []struct {
		Name   string
		Object any
	}{
		{
			Name:   "zkevm-wiop.bin",
			Object: lz.Zkevm,
		},
		{
			Name:   "disc.bin",
			Object: lz.Disc,
		},
		{
			Name:   "bootstrapper.bin",
			Object: lz.DistWizard.Bootstrapper,
		},
		{
			Name:   "compile-default.bin",
			Object: lz.DistWizard.CompiledDefault,
		},
	}

	for i, modGl := range lz.DistWizard.CompiledGLs {
		assets = append(assets, struct {
			Name   string
			Object any
		}{
			Name:   fmt.Sprintf("compile-gl-%d.bin", i),
			Object: modGl,
		})
	}

	for i, modLpp := range lz.DistWizard.CompiledLPPs {
		assets = append(assets, struct {
			Name   string
			Object any
		}{
			Name:   fmt.Sprintf("compile-lpp-%d.bin", i),
			Object: modLpp,
		})
	}

	assets = append(assets, struct {
		Name   string
		Object any
	}{
		Name:   "dw-raw.bin",
		Object: lz.DistWizard.CompiledConglomeration,
	})

	for _, asset := range assets {
		if err := writeToDisk(assetDir, asset.Name, asset.Object); err != nil {
			return err
		}
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
