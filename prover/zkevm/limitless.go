package zkevm

import (
	"fmt"
	"io"
	"math/rand/v2"
	"os"
	"path"
	"reflect"
	"strings"

	"github.com/consensys/linea-monorepo/prover/backend/files"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/serialization"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
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
	compileLppTemplate     = "dw-compiled-lpp-%v.bin"
	compileGlTemplate      = "dw-compiled-gl-%v.bin"
	debugLppTemplate       = "dw-debug-lpp-%v.bin"
	debugGlTemplate        = "dw-debug-gl-%v.bin"
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

// DiscoveryAdvices is a list of advice for the discovery of the modules. These
// values have been obtained thanks to a statistical analysis of the traces
// assignments involving correlation of the modules and hierarchical clustering.
// The advices are optimized to minimize the number of segments generated when
// producing an EVM proof.
var DiscoveryAdvices = []distributed.ModuleDiscoveryAdvice{
	{BaseSize: 4096, Cluster: "0", Column: "PUBLIC_INPUT_L2L1LOGS_EXTRACTED_HI"},
	{BaseSize: 32, Cluster: "0", Column: "LookUp_Num"},
	{BaseSize: 32, Cluster: "0", Column: "REPEATED_PATTERN_4741_24_PATTERN"},
	{BaseSize: 32, Cluster: "0", Column: "REPEATED_PATTERN_6890_20_PATTERN"},
	{BaseSize: 32, Cluster: "0", Column: "LOOKUP_TABLE_RANGE_1_30"},
	{BaseSize: 256, Cluster: "0", Column: "LOOKUP_TABLE_RANGE_1_136"},
	{BaseSize: 256, Cluster: "0", Column: "LOOKUP_TABLE_RANGE_1_144"},
	{BaseSize: 512, Cluster: "0", Column: "TABLE_instdecoder.ALPHA,instdecoder.DELTA,instdecoder.FAMILY_ACCOUNT,instdecoder.FAMILY_ADD,instdecoder.FAMILY_BATCH,instdecoder.FAMILY_BIN,instdecoder.FAMILY_CALL,instdecoder.FAMILY_CONTEXT,instdecoder.FAMILY_COPY,instdecoder.FAMILY_CREATE,instdecoder.FAMILY_DUP,instdecoder.FAMILY_EXT,instdecoder.FAMILY_HALT,instdecoder.FAMILY_INVALID,instdecoder.FAMILY_JUMP,instdecoder.FAMILY_KEC,instdecoder.FAMILY_LOG,instdecoder.FAMILY_MACHINE_STATE,instdecoder.FAMILY_MOD,instdecoder.FAMILY_MUL,instdecoder.FAMILY_PUSH_POP,instdecoder.FAMILY_SHF,instdecoder.FAMILY_STACK_RAM,instdecoder.FAMILY_STORAGE,instdecoder.FAMILY_SWAP,instdecoder.FAMILY_TRANSACTION,instdecoder.FAMILY_WCP,instdecoder.FLAG_1,instdecoder.FLAG_2,instdecoder.FLAG_3,instdecoder.FLAG_4,instdecoder.MXP_FLAG,instdecoder.OPCODE,instdecoder.STATIC_FLAG,instdecoder.STATIC_GAS,instdecoder.TWO_LINE_INSTRUCTION_0_LOGDERIVATIVE_M"},
	{BaseSize: 4096, Cluster: "0", Column: "TABLE_shfreftable.BYTE1,shfreftable.IOMF,shfreftable.LAS,shfreftable.MSHP,shfreftable.ONES,shfreftable.RAP_0_LOGDERIVATIVE_M"},
	{BaseSize: 16384, Cluster: "0", Column: "KECCAKF_BASE1_CLEAN_"},
	{BaseSize: 32768, Cluster: "0", Column: "KECCAKF_BASE1_DIRTY_"},
	{BaseSize: 262144, Cluster: "0", Column: "TABLE_binreftable.INPUT_BYTE_1,binreftable.INPUT_BYTE_2,binreftable.INST,binreftable.RESULT_BYTE_0_LOGDERIVATIVE_M"},
	{BaseSize: 32, Cluster: "1", Column: "TABLE_rlpaddr.ADDR_HI,rlpaddr.ADDR_LO,rlpaddr.DEP_ADDR_HI,rlpaddr.DEP_ADDR_LO,rlpaddr.KEC_HI,rlpaddr.KEC_LO,rlpaddr.NONCE,rlpaddr.RECIPE,rlpaddr.SALT_HI,rlpaddr.SALT_LO_0_LOGDERIVATIVE_M"},
	{BaseSize: 8, Cluster: "1", Column: "TABLE_blockhash.BLOCKHASH_ARG_HI_xor_EXO_ARG_1_HI,blockhash.BLOCKHASH_ARG_LO_xor_EXO_ARG_1_LO,blockhash.BLOCKHASH_RES_HI_xor_EXO_ARG_2_HI,blockhash.BLOCKHASH_RES_LO_xor_EXO_ARG_2_LO,blockhash.REL_BLOCK_0_LOGDERIVATIVE_M"},
	{BaseSize: 4096, Cluster: "1", Column: "PUBLIC_INPUT_ROLLING_MSG_EXTRACTED_HI"},
	{BaseSize: 4096, Cluster: "1", Column: "PUBLIC_INPUT_ROLLING_HASH_EXTRACTED_HI"},
	{BaseSize: 8192, Cluster: "1", Column: "BLOCK_TX_METADATA_BLOCK_ID"},
	{BaseSize: 512, Cluster: "6", Column: "MODEXP_INPUT_IS_MODEXP"},
	{BaseSize: 16384, Cluster: "6", Column: "TABLE_ext.ARG_1_HI,ext.ARG_1_LO,ext.ARG_2_HI,ext.ARG_2_LO,ext.ARG_3_HI,ext.ARG_3_LO,ext.INST,ext.RES_HI,ext.RES_LO_0_LOGDERIVATIVE_M"},
	{BaseSize: 1024, Cluster: "6", Column: "MODEXP_IS_ACTIVE"},
	{BaseSize: 256, Cluster: "6", Column: "MODEXP_256_BITS_IS_ACTIVE"},
	{BaseSize: 512, Cluster: "2", Column: "TABLE_ecdata.ID,ecdata.INDEX,ecdata.LIMB,ecdata.PHASE,ecdata.SUCCESS_BIT,ecdata.TOTAL_SIZE_0_LOGDERIVATIVE_M"},
	{BaseSize: 1024, Cluster: "2", Column: "ECADD_INTEGRATION_ALIGNMENT_PI"},
	{BaseSize: 256, Cluster: "2", Column: "ECMUL_INTEGRATION_ALIGNMENT_IS_ACTIVE"},
	{BaseSize: 128, Cluster: "2", Column: "ECPAIR_IS_ACTIVE"},
	{BaseSize: 128, Cluster: "2", Column: "ECPAIR_ALIGNMENT_ML_PI"},
	{BaseSize: 128, Cluster: "2", Column: "CYCLIC_COUNTER_6762_64_1024_COUNTER"},
	{BaseSize: 256, Cluster: "7", Column: "CLEANING_SHA2_CleanLimb"},
	{BaseSize: 256, Cluster: "7", Column: "ACCUMULATE_UP_TO_MAX_6864_Accumulator"},
	{BaseSize: 256, Cluster: "7", Column: "BLOCK_SHA2_AccNumLane"},
	{BaseSize: 256, Cluster: "7", Column: "SHA2_OVER_BLOCK_HASH_HI"},
	{BaseSize: 512, Cluster: "7", Column: "SHA2_OVER_BLOCK_SHA2_COMPRESSION_CIRCUIT_PI"},
	{BaseSize: 2048, Cluster: "3", Column: "IS_ZERO_13075_INVERSE_OR_ZERO_4096"},
	{BaseSize: 65536, Cluster: "3", Column: "PUBLIC_INPUT_RLP_TXN_FETCHER_CHAIN_ID"},
	{BaseSize: 32768, Cluster: "3", Column: "TABLE_rlptxrcpt.ABS_LOG_NUM,rlptxrcpt.ABS_LOG_NUM_MAX,rlptxrcpt.ABS_TX_NUM,rlptxrcpt.ABS_TX_NUM_MAX,rlptxrcpt.INPUT_1,rlptxrcpt.INPUT_2,rlptxrcpt.PHASE_ID_0_LOGDERIVATIVE_M"},
	{BaseSize: 4194304, Cluster: "3", Column: "rom.IS_PUSH"},
	{BaseSize: 256, Cluster: "3", Column: "TABLE_romlex.ADDRESS_HI,romlex.ADDRESS_LO,romlex.CODE_FRAGMENT_INDEX,romlex.CODE_HASH_HI,romlex.CODE_HASH_LO,romlex.CODE_SIZE,romlex.DEPLOYMENT_NUMBER,romlex.DEPLOYMENT_STATUS_0_LOGDERIVATIVE_M"},
	{BaseSize: 2048, Cluster: "3", Column: "TABLE_trm.IS_PRECOMPILE,trm.RAW_ADDRESS_HI,trm.RAW_ADDRESS_LO,trm.TRM_ADDRESS_HI_0_LOGDERIVATIVE_M"},
	{BaseSize: 1024, Cluster: "3", Column: "BLOCK_TX_METADATA_FILTER_ARITH"},
	{BaseSize: 4096, Cluster: "3", Column: "TABLE_logdata.ABS_LOG_NUM,logdata.ABS_LOG_NUM_MAX,logdata.SIZE_TOTAL_0_LOGDERIVATIVE_M"},
	{BaseSize: 4096, Cluster: "8", Column: "ECDSA_ANTICHAMBER_ADDRESSES_ADDRESS_HI"},
	{BaseSize: 4096, Cluster: "8", Column: "ECDSA_ANTICHAMBER_GNARK_DATA_IS_ACTIVE"},
	{BaseSize: 1024, Cluster: "3", Column: "IS_ZERO_13054_INVERSE_OR_ZERO_4096"},
	{BaseSize: 262144, Cluster: "3", Column: "MIMC_CODE_HASH_CFI"},
	{BaseSize: 256, Cluster: "3", Column: "CMP_MULTI_LIMB_4302_IS_GREATER"},
	{BaseSize: 64, Cluster: "3", Column: "PUBLIC_INPUT_TIMESTAMP_FETCHER_DATA"},
	{BaseSize: 128, Cluster: "3", Column: "PUBLIC_INPUT_TXN_DATA_FETCHER_ABS_TX_NUM"},
	{BaseSize: 8192, Cluster: "3", Column: "IS_ZERO_13155_INVERSE_OR_ZERO_131072"},
	{BaseSize: 8192, Cluster: "3", Column: "EXECUTION_DATA_COLLECTOR_ABS_TX_ID"},
	{BaseSize: 8192, Cluster: "3", Column: "CLEANING_EXECUTION_DATA_MIMC_CleanLimb"},
	{BaseSize: 8192, Cluster: "3", Column: "ACCUMULATE_UP_TO_MAX_7190_Accumulator"},
	{BaseSize: 2048, Cluster: "3", Column: "BLOCK_EXECUTION_DATA_MIMC_AccNumLane"},
	{BaseSize: 4096, Cluster: "3", Column: "CYCLIC_COUNTER_7207_524288_524288_COUNTER"},
	{BaseSize: 2097152, Cluster: "4", Column: "FILTER_CONNECTOR_HUB_STATE_SUMMARY_ACCOUNT_EPHEMERAL_FILTER"},
	{BaseSize: 16384, Cluster: "4", Column: "gas.INPUTS_AND_OUTPUTS_ARE_MEANINGFUL"},
	{BaseSize: 8388608, Cluster: "4", Column: "hub.(inv (- (shift hub:stkcp_CN_POW_4 1) hub:stkcp_CN_POW_4))"},
	{BaseSize: 524288, Cluster: "4", Column: "TABLE_mmu.AUX_ID_xor_CN_S_xor_EUC_A,mmu.EXO_SUM_xor_EXO_ID,mmu.INST_xor_INST_xor_CT,mmu.KEC_ID,mmu.LIMB_1_xor_LIMB_xor_WCP_ARG_1_HI,mmu.MICRO,mmu.MMIO_STAMP,mmu.PHASE,mmu.PHASE_xor_EXO_SUM,mmu.REF_OFFSET_xor_CN_T_xor_EUC_B,mmu.REF_SIZE_xor_SLO_xor_EUC_CEIL,mmu.SBO_xor_WCP_INST,mmu.SIZE,mmu.SIZE_xor_TLO_xor_EUC_QUOT,mmu.SRC_ID_xor_TOTAL_SIZE_xor_EUC_REM,mmu.SUCCESS_BIT_xor_SUCCESS_BIT_xor_EUC_FLAG,mmu.TBO_0_LOGDERIVATIVE_M"},
	{BaseSize: 262144, Cluster: "4", Column: "TABLE_oob.DATA_1,oob.DATA_2,oob.DATA_3,oob.DATA_4,oob.DATA_5,oob.DATA_6,oob.DATA_7,oob.DATA_8,oob.DATA_9,oob.OOB_INST_0_LOGDERIVATIVE_M"},
	{BaseSize: 262144, Cluster: "4", Column: "TABLE_mxp.CN,mxp.DEPLOYS,mxp.GAS_MXP,mxp.INST,mxp.MTNTOP,mxp.MXPX,mxp.OFFSET_1_HI,mxp.OFFSET_1_LO,mxp.OFFSET_2_HI,mxp.OFFSET_2_LO,mxp.SIZE_1_HI,mxp.SIZE_1_LO,mxp.SIZE_1_NONZERO_NO_MXPX,mxp.SIZE_2_HI,mxp.SIZE_2_LO,mxp.SIZE_2_NONZERO_NO_MXPX,mxp.STAMP,mxp.WORDS_0_LOGDERIVATIVE_M"},
	{BaseSize: 16384, Cluster: "4", Column: "TABLE_shakiradata.(shift shakiradata:LIMB -1),shakiradata.ID,shakiradata.INDEX,shakiradata.LIMB,shakiradata.PHASE_0_LOGDERIVATIVE_M"},
	{BaseSize: 8192, Cluster: "4", Column: "TABLE_stp.EXISTS,stp.GAS_ACTUAL,stp.GAS_HI,stp.GAS_LO,stp.GAS_MXP,stp.GAS_OUT_OF_POCKET,stp.GAS_STIPEND,stp.GAS_UPFRONT,stp.INSTRUCTION,stp.OUT_OF_GAS_EXCEPTION,stp.VAL_HI,stp.VAL_LO,stp.WARM_0_LOGDERIVATIVE_M"},
	{BaseSize: 32768, Cluster: "4", Column: "GENERIC_ACCUMULATOR_IsActive"},
	{BaseSize: 4096, Cluster: "4", Column: "GENERIC_ACCUMULATOR_Hash_Hi"},
	{BaseSize: 65536, Cluster: "4", Column: "CLEANING_KECCAK_CleanLimb"},
	{BaseSize: 131072, Cluster: "4", Column: "ACCUMULATE_UP_TO_MAX_4667_Accumulator"},
	{BaseSize: 65536, Cluster: "4", Column: "BASE_CONVERSION_IsFromBlockBaseB"},
	{BaseSize: 131072, Cluster: "4", Column: "CYCLIC_COUNTER_4744_24_262144_COUNTER"},
	{BaseSize: 4096, Cluster: "4", Column: "HASH_OUTPUT_Hash_Hi"},
	{BaseSize: 8192, Cluster: "4", Column: "TABLE_euc.DIVIDEND,euc.DIVISOR,euc.DONE,euc.QUOTIENT_0_LOGDERIVATIVE_M"},
	{BaseSize: 1048576, Cluster: "5", Column: "TABLE_mmio.MMIO_STAMP_0_LOGDERIVATIVE_M"},
	{BaseSize: 65536, Cluster: "5", Column: "TABLE_mod.ARG_1_HI,mod.ARG_1_LO,mod.ARG_2_HI,mod.ARG_2_LO,mod.INST,mod.RES_HI,mod.RES_LO_0_LOGDERIVATIVE_M"},
	{BaseSize: 2097152, Cluster: "5", Column: "mmio.(inv (- (shift mmio:CN_ABC_SORTED 1) mmio:CN_ABC_SORTED))"},
	{BaseSize: 32768, Cluster: "5", Column: "TABLE_mul.ARG_1_HI,mul.ARG_1_LO,mul.ARG_2_HI,mul.ARG_2_LO,mul.INSTRUCTION,mul.RES_HI,mul.RES_LO_0_LOGDERIVATIVE_M"},
	{BaseSize: 65536, Cluster: "5", Column: "TABLE_add.ARG_1_HI,add.ARG_1_LO,add.ARG_2_HI,add.ARG_2_LO,add.INST,add.RES_HI,add.RES_LO_0_LOGDERIVATIVE_M"},
	{BaseSize: 131072, Cluster: "5", Column: "TABLE_bin.ARGUMENT_1_HI,bin.ARGUMENT_1_LO,bin.ARGUMENT_2_HI,bin.ARGUMENT_2_LO,bin.INST,bin.RESULT_HI,bin.RESULT_LO_0_LOGDERIVATIVE_M"},
	{BaseSize: 131072, Cluster: "5", Column: "TABLE_shf.ARG_1_HI,shf.ARG_1_LO,shf.ARG_2_HI,shf.ARG_2_LO,shf.INST,shf.RES_HI,shf.RES_LO_0_LOGDERIVATIVE_M"},
	{BaseSize: 262144, Cluster: "5", Column: "TABLE_wcp.ARGUMENT_1_HI,wcp.ARGUMENT_1_LO,wcp.ARGUMENT_2_HI,wcp.ARGUMENT_2_LO,wcp.INST,wcp.RESULT_0_LOGDERIVATIVE_M"},
	{BaseSize: 4096, Cluster: "5", Column: "CMP_MULTI_LIMB_2493_IS_GREATER"},
	{BaseSize: 8192, Cluster: "5", Column: "ACCUMULATOR_COUNTER"},
	{BaseSize: 1048576, Cluster: "5", Column: "MIMC_ROUND_0_RESULT_0_7219"},
	{BaseSize: 16384, Cluster: "5", Column: "TABLE_exp.DATA_3_xor_WCP_ARG_2_HI,exp.DATA_4_xor_WCP_ARG_2_LO,exp.DATA_5,exp.EXP_INST,exp.MACRO,exp.RAW_ACC_xor_DATA_1_xor_WCP_ARG_1_HI,exp.TRIM_ACC_xor_DATA_2_xor_WCP_ARG_1_LO_0_LOGDERIVATIVE_M"},
	{BaseSize: 128, Cluster: "RARE", Column: "MODEXP_4096_BITS_PI"},
	{BaseSize: 1024, Cluster: "RARE", Column: "ECPAIR_ALIGNMENT_G2_IS_ACTIVE"},
	{BaseSize: 4096, Cluster: "RARE", Column: "PUBLIC_INPUT_ROLLING_SEL_EXISTS_MSG"},
	{BaseSize: 16, Cluster: "RARE", Column: "LOOKUP_TABLE_RANGE_1_16"},
	{BaseSize: 16, Cluster: "RARE", Column: "REPEATED_PATTERN_6743_16_PATTERN"},
	{BaseSize: 64, Cluster: "RARE", Column: "REPEATED_PATTERN_2417_64_PATTERN"},
	{BaseSize: 64, Cluster: "RARE", Column: "REPEATED_PATTERN_6710_64_PATTERN"},
	{BaseSize: 64, Cluster: "RARE", Column: "REPEATED_PATTERN_6751_64_PATTERN"},
	{BaseSize: 64, Cluster: "RARE", Column: "REPEATED_PATTERN_6759_64_PATTERN"},
	{BaseSize: 64, Cluster: "RARE", Column: "REPEATED_PATTERN_6902_64_PATTERN"},
	{BaseSize: 128, Cluster: "RARE", Column: "REPEATED_PATTERN_6686_128_PATTERN"},
	{BaseSize: 128, Cluster: "RARE", Column: "REPEATED_PATTERN_6694_128_PATTERN"},
	{BaseSize: 512, Cluster: "RARE", Column: "REPEATED_PATTERN_6702_512_PATTERN"},
	{BaseSize: 16384, Cluster: "RARE", Column: "LOOKUP_BaseBDirty"},
	{BaseSize: 65536, Cluster: "RARE", Column: "LOOKUP_BaseA"},
}

// NewLimitlessZkEVM returns a new LimitlessZkEVM object.
func NewLimitlessZkEVM(cfg *config.Config) *LimitlessZkEVM {
	var (
		traceLimits = cfg.TracesLimits
		zkevm       = FullZKEVMWithSuite(&traceLimits, CompilationSuite{}, cfg)
		disc        = &distributed.StandardModuleDiscoverer{
			TargetWeight: 1 << 28,
			Predivision:  1,
			Advices:      DiscoveryAdvices,
		}
		dw = distributed.DistributeWizard(zkevm.WizardIOP, disc)
	)

	// These are the slow and expensive operations.
	dw.CompileSegments().Conglomerate(50)

	decorateWithPublicInputs(dw.CompiledConglomeration)

	return &LimitlessZkEVM{
		Zkevm:      zkevm,
		DistWizard: dw,
	}
}

// NewLimitlessRawZkEVM returns a new LimitlessZkEVM object without any
// compilation.
func NewLimitlessRawZkEVM(cfg *config.Config) *LimitlessZkEVM {

	var (
		traceLimits = cfg.TracesLimits
		zkevm       = FullZKEVMWithSuite(&traceLimits, CompilationSuite{}, cfg)
		disc        = &distributed.StandardModuleDiscoverer{
			TargetWeight: 1 << 28,
			Predivision:  1,
			Advices:      DiscoveryAdvices,
		}
		dw = distributed.DistributeWizard(zkevm.WizardIOP, disc)
	)

	return &LimitlessZkEVM{
		Zkevm:      zkevm,
		DistWizard: dw,
	}
}

// NewLimitlessDebugZkEVM returns a new LimitlessZkEVM with only the debugging
// components. The resulting object is not meant to be stored on disk and should
// be used right away to debug the prover. The return object can run the
// bootstrapper (with added) sanity-checks, the segmentation and then sanity-
// checking all the segments.
func NewLimitlessDebugZkEVM(cfg *config.Config) *LimitlessZkEVM {

	var (
		traceLimits = cfg.TracesLimits
		zkevm       = FullZKEVMWithSuite(&traceLimits, CompilationSuite{}, cfg)
		disc        = &distributed.StandardModuleDiscoverer{
			TargetWeight: 1 << 28,
			Predivision:  1,
			Advices:      DiscoveryAdvices,
		}
		dw             = distributed.DistributeWizard(zkevm.WizardIOP, disc)
		limitlessZkEVM = &LimitlessZkEVM{
			Zkevm:      zkevm,
			DistWizard: dw,
		}
	)

	// This adds debugging to the bootstrapper which are normally not present by
	// default.
	wizard.ContinueCompilation(
		limitlessZkEVM.DistWizard.Bootstrapper,
		dummy.CompileAtProverLvl(dummy.WithMsg("bootstrapper")),
	)

	return limitlessZkEVM
}

// RunStatRecords runs only the bootstrapper and returns a list of stat records
func (lz *LimitlessZkEVM) RunStatRecords(witness *Witness) []distributed.QueryBasedAssignmentStatsRecord {

	var (
		runtimeBoot = wizard.RunProver(
			lz.DistWizard.Bootstrapper,
			lz.Zkevm.GetMainProverStep(witness),
		)

		res  = []distributed.QueryBasedAssignmentStatsRecord{}
		disc = lz.DistWizard.Disc.(*distributed.StandardModuleDiscoverer)
	)

	for _, mod := range disc.Modules {
		res = append(res, mod.RecordAssignmentStats(runtimeBoot)...)
	}

	return res
}

// RunDebug runs the LimitlessZkEVM on debug mode. It will run the boostrapper,
// the segmentation and then the sanity checks for all the segments. The
// check of the LPP module is done using a deterministic pseudo-random number
// generator and will yield the same result every time.
func (lz *LimitlessZkEVM) RunDebug(witness *Witness) {

	runtimeBoot := wizard.RunProver(
		lz.DistWizard.Bootstrapper,
		lz.Zkevm.GetMainProverStep(witness),
	)

	witnessGLs, witnessLPPs := distributed.SegmentRuntime(
		runtimeBoot,
		lz.DistWizard.Disc,
		lz.DistWizard.BlueprintGLs,
		lz.DistWizard.BlueprintLPPs,
	)

	for _, witness := range witnessGLs {

		var (
			moduleToFind = witness.ModuleName
			debugGL      *distributed.ModuleGL
		)

		for i := range lz.DistWizard.DebugGLs {
			if lz.DistWizard.DebugGLs[i].DefinitionInput.ModuleName == moduleToFind {
				debugGL = lz.DistWizard.DebugGLs[i]
				break
			}
		}

		if debugGL == nil {
			utils.Panic("debugGL not found")
		}

		var (
			mainProverStep = debugGL.GetMainProverStep(witness)
			compiledIOP    = debugGL.Wiop
		)

		// The debugGLs is compiled with the CompileAtProverLevel routine so we
		// don't need the proof to complete the sanity checks: everything is
		// done at the prover level.
		_ = wizard.Prove(compiledIOP, mainProverStep)
	}

	// Here, we can't we can't just use 0 or a dummy small value because there
	// is a risk of creating false-positives with the grand-products and the
	// horner (as if one of the term of the product cancels, the product is
	// zero and we want to prevent that) or false negative due to inverting
	// zeroes in the log-derivative sums.
	rng := rand.New(utils.NewRandSource(42))
	sharedRandomness := field.PseudoRand(rng)

	for _, witness := range witnessLPPs {

		var (
			moduleToFind = witness.ModuleName
			debugLPP     *distributed.ModuleLPP
		)

		for i := range lz.DistWizard.DebugLPPs {
			if reflect.DeepEqual(lz.DistWizard.DebugLPPs[i].ModuleNames(), moduleToFind) {
				debugLPP = lz.DistWizard.DebugLPPs[i]
				break
			}
		}

		if debugLPP == nil {
			utils.Panic("debugLPP not found")
		}

		witness.InitialFiatShamirState = sharedRandomness

		var (
			mainProverStep = debugLPP.GetMainProverStep(witness)
			compiledIOP    = debugLPP.Wiop
		)

		// The debugLPP is compiled with the CompileAtProverLevel routine so we
		// don't need the proof to complete the sanity checks: everything is
		// done at the prover level.
		_ = wizard.Prove(compiledIOP, mainProverStep)
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

	for _, modGl := range lz.DistWizard.CompiledGLs {
		assets = append(assets, asset{
			Name:   fmt.Sprintf(compileGlTemplate, modGl.ModuleGL.DefinitionInput.ModuleName),
			Object: *modGl,
		})
	}

	for i, blueprintGL := range lz.DistWizard.BlueprintGLs {
		assets = append(assets, asset{
			Name:   fmt.Sprintf(blueprintGLTemplate, i),
			Object: blueprintGL,
		})
	}

	for _, debugGL := range lz.DistWizard.DebugGLs {
		assets = append(assets, asset{
			Name:   fmt.Sprintf(debugGlTemplate, debugGL.DefinitionInput.ModuleName),
			Object: debugGL,
		})
	}

	for _, modLpp := range lz.DistWizard.CompiledLPPs {
		assets = append(assets, asset{
			Name:   fmt.Sprintf(compileLppTemplate, modLpp.ModuleLPP.ModuleNames()),
			Object: *modLpp,
		})
	}

	for i, blueprintLPP := range lz.DistWizard.BlueprintLPPs {
		assets = append(assets, asset{
			Name:   fmt.Sprintf(blueprintLppTemplate, i),
			Object: blueprintLPP,
		})
	}

	for _, debugLPP := range lz.DistWizard.DebugLPPs {
		assets = append(assets, asset{
			Name:   fmt.Sprintf(debugLppTemplate, debugLPP.ModuleNames()),
			Object: debugLPP,
		})
	}

	assets = append(assets, struct {
		Name   string
		Object any
	}{
		Name:   conglomerationFile,
		Object: *lz.DistWizard.CompiledConglomeration,
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

// LoadCompiledGL loads the compiled GL from disk
func LoadCompiledGL(cfg *config.Config, moduleName distributed.ModuleName) (*distributed.RecursedSegmentCompilation, error) {

	var (
		assetDir = cfg.PathForSetup(executionLimitlessPath)
		filePath = path.Join(assetDir, fmt.Sprintf(compileGlTemplate, moduleName))
		res      = &distributed.RecursedSegmentCompilation{}
	)

	if err := loadFromFile(filePath, res); err != nil {
		return nil, err
	}

	return res, nil
}

// LoadCompiledLPP loads the compiled LPP from disk
func LoadCompiledLPP(cfg *config.Config, moduleNames []distributed.ModuleName) (*distributed.RecursedSegmentCompilation, error) {

	var (
		assetDir = cfg.PathForSetup(executionLimitlessPath)
		filePath = path.Join(assetDir, fmt.Sprintf(compileLppTemplate, moduleNames))
		res      = &distributed.RecursedSegmentCompilation{}
	)

	if err := loadFromFile(filePath, res); err != nil {
		return nil, err
	}

	return res, nil
}

// LoadDebugGL loads the debug GL from disk
func LoadDebugGL(cfg *config.Config, moduleName distributed.ModuleName) (*distributed.ModuleGL, error) {

	var (
		assetDir = cfg.PathForSetup(executionLimitlessPath)
		filePath = path.Join(assetDir, fmt.Sprintf(debugGlTemplate, moduleName))
		res      = &distributed.ModuleGL{}
	)

	if err := loadFromFile(filePath, res); err != nil {
		return nil, err
	}

	return res, nil
}

// LoadDebugLPP loads the debug LPP from disk
func LoadDebugLPP(cfg *config.Config, moduleName []distributed.ModuleName) (*distributed.ModuleLPP, error) {

	var (
		assetDir = cfg.PathForSetup(executionLimitlessPath)
		filePath = path.Join(assetDir, fmt.Sprintf(debugLppTemplate, moduleName))
		res      = &distributed.ModuleLPP{}
	)

	if err := loadFromFile(filePath, res); err != nil {
		return nil, err
	}

	return res, nil
}

// LoadConglomeration loads the conglomeration assets from disk
func LoadConglomeration(cfg *config.Config) (*distributed.ConglomeratorCompilation, error) {

	var (
		assetDir = cfg.PathForSetup(executionLimitlessPath)
		filePath = path.Join(assetDir, conglomerationFile)
		res      = &distributed.ConglomeratorCompilation{}
	)

	if err := loadFromFile(filePath, res); err != nil {
		return nil, err
	}

	return res, nil
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
