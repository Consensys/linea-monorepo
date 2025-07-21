package distributed_test

import (
	"fmt"
	"testing"

	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/zkevm"
)

func TestStandardDiscoveryOnZkEVM(t *testing.T) {

	var (
		z    = zkevm.GetTestZkEVM()
		disc = &distributed.StandardModuleDiscoverer{
			TargetWeight: 1 << 28,
			Affinities:   zkevm.GetAffinities(z),
			Predivision:  16,
		}
	)

	distributed.PrecompileInitialWizard(z.WizardIOP, disc)

	// The test is to make sure that this function returns
	disc.Analyze(z.WizardIOP)

	fmt.Printf("%++v\n", disc)

	allCols := z.WizardIOP.Columns.AllKeys()
	for _, colName := range allCols {
		col := z.WizardIOP.Columns.GetHandle(colName)

		var (
			nat     = col.(column.Natural)
			newSize = disc.NewSizeOf(nat)
			module  = disc.ModuleOf(nat)
		)

		if module == "" {
			t.Errorf("module of %v is empty", colName)
		}

		if newSize == 0 {
			t.Errorf("new-size of %v is 0", colName)
		}
	}

	for _, col := range z.WizardIOP.Columns.AllKeys() {

		var (
			nat     = z.WizardIOP.Columns.GetHandle(col).(column.Natural)
			modules = []distributed.ModuleName{}
		)

		for i := range disc.Modules {
			mod := disc.Modules[i]
			for k := range mod.SubModules {
				if mod.SubModules[k].Ds.Has(nat.ID) {
					modules = append(modules, mod.ModuleName)
				}
			}
		}

		if len(modules) == 0 {
			t.Errorf("could not match any module for %v", col)
		}

		if len(modules) > 1 {
			t.Errorf("could match more than one module for %v: %v", col, modules)
		}
	}

	t.Logf("totalNumber of columns: %v", len(z.WizardIOP.Columns.AllKeys()))

	for _, mod := range disc.Modules {
		t.Logf("module=%v weight=%v numcol=%v\n", mod.ModuleName, mod.Weight(z.WizardIOP), disc.NumColumnOf(mod.ModuleName))
	}
}

func TestStandardDiscoveryOnZkEVMWithAdvices(t *testing.T) {

	var (
		z    = zkevm.GetTestZkEVM()
		disc = &distributed.StandardModuleDiscoverer{
			TargetWeight: 1 << 28,
			Affinities:   zkevm.GetAffinities(z),
			Predivision:  16,
			Advices: []distributed.ModuleDiscoveryAdvice{
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
				{BaseSize: 512, Cluster: "2", Column: "MODEXP_INPUT_IS_MODEXP"},
				{BaseSize: 512, Cluster: "2", Column: "TABLE_ecdata.ID,ecdata.INDEX,ecdata.LIMB,ecdata.PHASE,ecdata.SUCCESS_BIT,ecdata.TOTAL_SIZE_0_LOGDERIVATIVE_M"},
				{BaseSize: 1024, Cluster: "2", Column: "MODEXP_IS_ACTIVE"},
				{BaseSize: 128, Cluster: "2", Column: "MODEXP_256_BITS_IS_ACTIVE"},
				{BaseSize: 1024, Cluster: "2", Column: "ECADD_INTEGRATION_ALIGNMENT_PI"},
				{BaseSize: 512, Cluster: "2", Column: "ECMUL_INTEGRATION_ALIGNMENT_IS_ACTIVE"},
				{BaseSize: 128, Cluster: "2", Column: "ECPAIR_IS_ACTIVE"},
				{BaseSize: 64, Cluster: "2", Column: "ECPAIR_ALIGNMENT_ML_PI"},
				{BaseSize: 64, Cluster: "2", Column: "CYCLIC_COUNTER_6762_64_1024_COUNTER"},
				{BaseSize: 64, Cluster: "2", Column: "CLEANING_SHA2_CleanLimb"},
				{BaseSize: 256, Cluster: "2", Column: "ACCUMULATE_UP_TO_MAX_6864_Accumulator"},
				{BaseSize: 256, Cluster: "2", Column: "BLOCK_SHA2_AccNumLane"},
				{BaseSize: 256, Cluster: "2", Column: "SHA2_OVER_BLOCK_HASH_HI"},
				{BaseSize: 512, Cluster: "2", Column: "SHA2_OVER_BLOCK_SHA2_COMPRESSION_CIRCUIT_PI"},
				{BaseSize: 16384, Cluster: "2", Column: "TABLE_ext.ARG_1_HI,ext.ARG_1_LO,ext.ARG_2_HI,ext.ARG_2_LO,ext.ARG_3_HI,ext.ARG_3_LO,ext.INST,ext.RES_HI,ext.RES_LO_0_LOGDERIVATIVE_M"},
				{BaseSize: 2048, Cluster: "3", Column: "IS_ZERO_13075_INVERSE_OR_ZERO_4096"},
				{BaseSize: 65536, Cluster: "3", Column: "PUBLIC_INPUT_RLP_TXN_FETCHER_CHAIN_ID"},
				{BaseSize: 32768, Cluster: "3", Column: "TABLE_rlptxrcpt.ABS_LOG_NUM,rlptxrcpt.ABS_LOG_NUM_MAX,rlptxrcpt.ABS_TX_NUM,rlptxrcpt.ABS_TX_NUM_MAX,rlptxrcpt.INPUT_1,rlptxrcpt.INPUT_2,rlptxrcpt.PHASE_ID_0_LOGDERIVATIVE_M"},
				{BaseSize: 4194304, Cluster: "3", Column: "rom.IS_PUSH"},
				{BaseSize: 256, Cluster: "3", Column: "TABLE_romlex.ADDRESS_HI,romlex.ADDRESS_LO,romlex.CODE_FRAGMENT_INDEX,romlex.CODE_HASH_HI,romlex.CODE_HASH_LO,romlex.CODE_SIZE,romlex.DEPLOYMENT_NUMBER,romlex.DEPLOYMENT_STATUS_0_LOGDERIVATIVE_M"},
				{BaseSize: 2048, Cluster: "3", Column: "TABLE_trm.IS_PRECOMPILE,trm.RAW_ADDRESS_HI,trm.RAW_ADDRESS_LO,trm.TRM_ADDRESS_HI_0_LOGDERIVATIVE_M"},
				{BaseSize: 1024, Cluster: "3", Column: "BLOCK_TX_METADATA_FILTER_ARITH"},
				{BaseSize: 4096, Cluster: "3", Column: "TABLE_logdata.ABS_LOG_NUM,logdata.ABS_LOG_NUM_MAX,logdata.SIZE_TOTAL_0_LOGDERIVATIVE_M"},
				{BaseSize: 4096, Cluster: "3", Column: "ECDSA_ANTICHAMBER_ADDRESSES_ADDRESS_HI"},
				{BaseSize: 4096, Cluster: "3", Column: "ECDSA_ANTICHAMBER_GNARK_DATA_IS_ACTIVE"},
				{BaseSize: 1024, Cluster: "3", Column: "IS_ZERO_13054_INVERSE_OR_ZERO_4096"},
				{BaseSize: 262144, Cluster: "3", Column: "IS_ZERO_10824_INVERSE_OR_ZERO_4194304"},
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
			},
		}
	)

	distributed.PrecompileInitialWizard(z.WizardIOP, disc)

	// The test is to make sure that this function returns
	disc.Analyze(z.WizardIOP)

	fmt.Printf("%++v\n", disc)

	allCols := z.WizardIOP.Columns.AllKeys()
	for _, colName := range allCols {
		col := z.WizardIOP.Columns.GetHandle(colName)

		var (
			nat     = col.(column.Natural)
			newSize = disc.NewSizeOf(nat)
			module  = disc.ModuleOf(nat)
		)

		if module == "" {
			t.Errorf("module of %v is empty", colName)
		}

		if newSize == 0 {
			t.Errorf("new-size of %v is 0", colName)
		}
	}

	for _, col := range z.WizardIOP.Columns.AllKeys() {

		var (
			nat     = z.WizardIOP.Columns.GetHandle(col).(column.Natural)
			modules = []distributed.ModuleName{}
		)

		for i := range disc.Modules {
			mod := disc.Modules[i]
			for k := range mod.SubModules {
				if mod.SubModules[k].Ds.Has(nat.ID) {
					modules = append(modules, mod.ModuleName)
				}
			}
		}

		if len(modules) == 0 {
			t.Errorf("could not match any module for %v", col)
		}

		if len(modules) > 1 {
			t.Errorf("could match more than one module for %v: %v", col, modules)
		}
	}

	t.Logf("totalNumber of columns: %v", len(z.WizardIOP.Columns.AllKeys()))

	for _, mod := range disc.Modules {
		t.Logf("module=%v weight=%v numcol=%v\n", mod.ModuleName, mod.Weight(z.WizardIOP), disc.NumColumnOf(mod.ModuleName))
	}
}
