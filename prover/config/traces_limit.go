package config

import (
	"bytes"
	"encoding/json"

	"github.com/consensys/linea-monorepo/prover/utils"
)

// The trace limits define the maximum trace size per module that the prover can handle.
// Raising these limits increases the prover's memory needs and simultaneously decreases the number
// of transactions it can prove in a single go. These traces are vital for the setup generator, so
// any changes in trace limits mean we'll need to run a new setup and update the verifier contracts
// before deploying.
type TracesLimits struct {
	Add               int `mapstructure:"ADD" validate:"power_of_2" corset:"add"`
	Bin               int `mapstructure:"BIN" validate:"power_of_2" corset:"bin"`
	Blake2Fmodexpdata int `mapstructure:"BLAKE_MODEXP_DATA" validate:"power_of_2" corset:"blake2fmodexpdata"`
	Blockdata         int `mapstructure:"BLOCK_DATA" corset:"blockdata"`
	Blockhash         int `mapstructure:"BLOCK_HASH" validate:"power_of_2" corset:"blockhash"`
	Ecdata            int `mapstructure:"EC_DATA" validate:"power_of_2" corset:"ecdata"`
	Euc               int `mapstructure:"EUC" validate:"power_of_2" corset:"euc"`
	Exp               int `mapstructure:"EXP" validate:"power_of_2" corset:"exp"`
	Ext               int `mapstructure:"EXT" validate:"power_of_2" corset:"ext"`
	Gas               int `mapstructure:"GAS" validate:"power_of_2" corset:"gas"`
	Hub               int `mapstructure:"HUB" validate:"power_of_2" corset:"hub"`
	Logdata           int `mapstructure:"LOG_DATA" validate:"power_of_2" corset:"logdata"`
	Loginfo           int `mapstructure:"LOG_INFO" validate:"power_of_2" corset:"loginfo"`
	Mmio              int `mapstructure:"MMIO" validate:"power_of_2" corset:"mmio"`
	Mmu               int `mapstructure:"MMU" validate:"power_of_2" corset:"mmu"`
	Mod               int `mapstructure:"MOD" validate:"power_of_2" corset:"mod"`
	Mul               int `mapstructure:"MUL" validate:"power_of_2" corset:"mul"`
	Mxp               int `mapstructure:"MXP" validate:"power_of_2" corset:"mxp"`
	Oob               int `mapstructure:"OOB" validate:"power_of_2" corset:"oob"`
	Rlpaddr           int `mapstructure:"RLP_ADDR" validate:"power_of_2" corset:"rlpaddr"`
	Rlptxn            int `mapstructure:"RLP_TXN" validate:"power_of_2" corset:"rlptxn"`
	Rlptxrcpt         int `mapstructure:"RLP_TXN_RCPT" validate:"power_of_2" corset:"rlptxrcpt"`
	Rom               int `mapstructure:"ROM" validate:"power_of_2" corset:"rom"`
	Romlex            int `mapstructure:"ROM_LEX" validate:"power_of_2" corset:"romlex"`
	Shakiradata       int `mapstructure:"SHAKIRA_DATA" validate:"power_of_2" corset:"shakiradata"`
	Shf               int `mapstructure:"SHF" validate:"power_of_2" corset:"shf"`
	Stp               int `mapstructure:"STP" validate:"power_of_2" corset:"stp"`
	Trm               int `mapstructure:"TRM" validate:"power_of_2" corset:"trm"`
	Txndata           int `mapstructure:"TXN_DATA" validate:"power_of_2" corset:"txndata"`
	Wcp               int `mapstructure:"WCP" validate:"power_of_2" corset:"wcp"`
	// internal modules
	ByteShl256    int `mapstructure:"BYTE_SHL256" validate:"power_of_2" corset:"byte_shl256"`
	ByteShr256    int `mapstructure:"BYTE_SHR256" validate:"power_of_2" corset:"byte_shr256"`
	ByteSar256    int `mapstructure:"BYTE_SAR256" validate:"power_of_2" corset:"byte_sar256"`
	BitShl256     int `mapstructure:"BIT_SHL256" validate:"power_of_2" corset:"bit_shl256"`
	BitShr256     int `mapstructure:"BIT_SHR256" validate:"power_of_2" corset:"bit_shr256"`
	BitSar256     int `mapstructure:"BIT_SAR256" validate:"power_of_2" corset:"bit_sar256"`
	FillBytesFrom int `mapstructure:"FILL_BYTES_FROM" validate:"power_of_2" corset:"fill_bytes_from"`
	Log2          int `mapstructure:"LOG2" validate:"power_of_2" corset:"log2"`
	Log2U128      int `mapstructure:"LOG2_U128" validate:"power_of_2" corset:"log2_u128"`
	Log2U64       int `mapstructure:"LOG2_U64" validate:"power_of_2" corset:"log2_u64"`
	Log2U32       int `mapstructure:"LOG2_U32" validate:"power_of_2" corset:"log2_u32"`
	Log2U16       int `mapstructure:"LOG2_U16" validate:"power_of_2" corset:"log2_u16"`
	Log2U8        int `mapstructure:"LOG2_U8" validate:"power_of_2" corset:"log2_u8"`
	Log2U4        int `mapstructure:"LOG2_U4" validate:"power_of_2" corset:"log2_u4"`
	Log2U2        int `mapstructure:"LOG2_U2" validate:"power_of_2" corset:"log2_u2"`
	Log256        int `mapstructure:"LOG256" validate:"power_of_2" corset:"log256"`
	Log256U128    int `mapstructure:"LOG256_U128" validate:"power_of_2" corset:"log256_u128"`
	Log256U64     int `mapstructure:"LOG256_U64" validate:"power_of_2" corset:"log256_u64"`
	Log256U32     int `mapstructure:"LOG256_U32" validate:"power_of_2" corset:"log256_u32"`
	Log256U16     int `mapstructure:"LOG256_U16" validate:"power_of_2" corset:"log256_u16"`
	SetByte256    int `mapstructure:"SET_BYTE256" validate:"power_of_2" corset:"set_byte256"`
	SetByte128    int `mapstructure:"SET_BYTE128" validate:"power_of_2" corset:"set_byte128"`
	SetByte64     int `mapstructure:"SET_BYTE64" validate:"power_of_2" corset:"set_byte64"`
	SetByte32     int `mapstructure:"SET_BYTE32" validate:"power_of_2" corset:"set_byte32"`
	SetByte16     int `mapstructure:"SET_BYTE16" validate:"power_of_2" corset:"set_byte16"`
	// limitless typing modules
	U128 int `mapstructure:"U128" validate:"power_of_2" corset:"u128"`
	U127 int `mapstructure:"U127" validate:"power_of_2" corset:"u127"`
	U126 int `mapstructure:"U126" validate:"power_of_2" corset:"u126"`
	U120 int `mapstructure:"U120" validate:"power_of_2" corset:"u120"`
	U119 int `mapstructure:"U119" validate:"power_of_2" corset:"u119"`
	U64  int `mapstructure:"U64" validate:"power_of_2" corset:"u64"`
	U63  int `mapstructure:"U63" validate:"power_of_2" corset:"u63"`
	U62  int `mapstructure:"U62" validate:"power_of_2" corset:"u62"`
	U56  int `mapstructure:"U56" validate:"power_of_2" corset:"u56"`
	U55  int `mapstructure:"U55" validate:"power_of_2" corset:"u55"`
	U36  int `mapstructure:"U36" validate:"power_of_2" corset:"u36"`
	U32  int `mapstructure:"U32" validate:"power_of_2" corset:"u32"`
	U31  int `mapstructure:"U31" validate:"power_of_2" corset:"u31"`
	U30  int `mapstructure:"U30" validate:"power_of_2" corset:"u30"`
	U24  int `mapstructure:"U24" validate:"power_of_2" corset:"u24"`
	U23  int `mapstructure:"U23" validate:"power_of_2" corset:"u23"`
	U20  int `mapstructure:"U20" validate:"power_of_2" corset:"u20"`
	// reference tables
	Binreftable int `mapstructure:"BIN_REFERENCE_TABLE" validate:"power_of_2" corset:"binreftable"`
	Instdecoder int `mapstructure:"INSTRUCTION_DECODER" validate:"power_of_2" corset:"instdecoder"`

	PrecompileEcrecoverEffectiveCalls    int `mapstructure:"PRECOMPILE_ECRECOVER_EFFECTIVE_CALLS"`
	PrecompileSha2Blocks                 int `mapstructure:"PRECOMPILE_SHA2_BLOCKS"`
	PrecompileRipemdBlocks               int `mapstructure:"PRECOMPILE_RIPEMD_BLOCKS"`
	PrecompileModexpEffectiveCalls       int `mapstructure:"PRECOMPILE_MODEXP_EFFECTIVE_CALLS"`
	PrecompileModexpEffectiveCalls4096   int `mapstructure:"PRECOMPILE_MODEXP_EFFECTIVE_CALLS_4096"`
	PrecompileEcaddEffectiveCalls        int `mapstructure:"PRECOMPILE_ECADD_EFFECTIVE_CALLS"`
	PrecompileEcmulEffectiveCalls        int `mapstructure:"PRECOMPILE_ECMUL_EFFECTIVE_CALLS"`
	PrecompileEcpairingEffectiveCalls    int `mapstructure:"PRECOMPILE_ECPAIRING_FINAL_EXPONENTIATIONS"`
	PrecompileEcpairingMillerLoops       int `mapstructure:"PRECOMPILE_ECPAIRING_MILLER_LOOPS"`
	PrecompileEcpairingG2MembershipCalls int `mapstructure:"PRECOMPILE_ECPAIRING_G2_MEMBERSHIP_CALLS"`
	PrecompileBlakeEffectiveCalls        int `mapstructure:"PRECOMPILE_BLAKE_EFFECTIVE_CALLS"`
	PrecompileBlakeRounds                int `mapstructure:"PRECOMPILE_BLAKE_ROUNDS"`

	BlockKeccak       int `mapstructure:"BLOCK_KECCAK"`
	BlockL1Size       int `mapstructure:"BLOCK_L1_SIZE"`
	BlockL2L1Logs     int `mapstructure:"BLOCK_L2_L1_LOGS"`
	BlockTransactions int `mapstructure:"BLOCK_TRANSACTIONS"`

	ShomeiMerkleProofs int `mapstructure:"SHOMEI_MERKLE_PROOFS"`

	// beta v4.0
	PrecompileBlsPointEvaluationEffectiveCalls     int `mapstructure:"PRECOMPILE_BLS_POINT_EVALUATION_EFFECTIVE_CALLS"`
	PrecompilePointEvaluationFailureEffectiveCalls int `mapstructure:"PRECOMPILE_POINT_EVALUATION_FAILURE_EFFECTIVE_CALLS"`
	PrecompileBlsG1AddEffectiveCalls               int `mapstructure:"PRECOMPILE_BLS_G1_ADD_EFFECTIVE_CALLS"`
	PrecompileBlsG1MsmEffectiveCalls               int `mapstructure:"PRECOMPILE_BLS_G1_MSM_EFFECTIVE_CALLS"`
	PrecompileBlsG2AddEffectiveCalls               int `mapstructure:"PRECOMPILE_BLS_G2_ADD_EFFECTIVE_CALLS"`
	PrecompileBlsG2MsmEffectiveCalls               int `mapstructure:"PRECOMPILE_BLS_G2_MSM_EFFECTIVE_CALLS"`
	PrecompileBlsPairingCheckMillerLoops           int `mapstructure:"PRECOMPILE_BLS_PAIRING_CHECK_MILLER_LOOPS"`
	PrecompileBlsFinalExponentiations              int `mapstructure:"PRECOMPILE_BLS_FINAL_EXPONENTIATIONS"`
	PrecompileBlsMapFpToG1EffectiveCalls           int `mapstructure:"PRECOMPILE_BLS_MAP_FP_TO_G1_EFFECTIVE_CALLS"`
	PrecompileBlsMapFp2ToG2EffectiveCalls          int `mapstructure:"PRECOMPILE_BLS_MAP_FP2_TO_G2_EFFECTIVE_CALLS"`
	PrecompileBlsC1MembershipChecks                int `mapstructure:"PRECOMPILE_BLS_C1_MEMBERSHIP_CHECKS"`
	PrecompileBlsC2MembershipCalls                 int `mapstructure:"PRECOMPILE_BLS_C2_MEMBERSHIP_CALLS"`
	PrecompileBlsG1MembershipCalls                 int `mapstructure:"PRECOMPILE_BLS_G1_MEMBERSHIP_CALLS"`
	PrecompileBlsG2MembershipCalls                 int `mapstructure:"PRECOMPILE_BLS_G2_MEMBERSHIP_CALLS"`
	Bls                                            int `mapstructure:"BLS"`
	PointEval                                      int `mapstructure:"POINT_EVAL"`
	BlsData                                        int `mapstructure:"BLS_DATA"`
	RlpUtils                                       int `mapstructure:"RLP_UTILS"`
	PowerReferenceTable                            int `mapstructure:"POWER_REFERENCE_TABLE"`
	BlsReferenceTable                              int `mapstructure:"BLS_REFERENCE_TABLE"`
}

func (tl *TracesLimits) Checksum() string {
	// encode the struct to json, then hash it
	encoded, err := json.Marshal(tl)
	if err != nil {
		panic(err) // should never happen
	}

	digest, err := utils.Digest(bytes.NewReader(encoded))
	if err != nil {
		panic(err) // should never happen
	}

	return digest
}

func (tl *TracesLimits) ScaleUp(by int) {

	tl.Add *= by
	tl.Bin *= by
	tl.Blake2Fmodexpdata *= by
	tl.Blockdata *= by
	tl.Blockhash *= by
	tl.Ecdata *= by
	tl.Euc *= by
	tl.Exp *= by
	tl.Ext *= by
	tl.Gas *= by
	tl.Hub *= by
	tl.Logdata *= by
	tl.Loginfo *= by
	tl.Mmio *= by
	tl.Mmu *= by
	tl.Mod *= by
	tl.Mul *= by
	tl.Mxp *= by
	tl.Oob *= by
	tl.Rlpaddr *= by
	tl.Rlptxn *= by
	tl.Rlptxrcpt *= by
	tl.Rom *= by
	tl.Romlex *= by
	tl.Shakiradata *= by
	tl.Shf *= by
	tl.Stp *= by
	tl.Trm *= by
	tl.Txndata *= by
	tl.Wcp *= by
	tl.PrecompileEcrecoverEffectiveCalls *= by
	tl.PrecompileSha2Blocks *= by
	tl.PrecompileModexpEffectiveCalls *= by
	tl.PrecompileModexpEffectiveCalls4096 *= by
	tl.PrecompileEcaddEffectiveCalls *= by
	tl.PrecompileEcmulEffectiveCalls *= by
	tl.PrecompileEcpairingEffectiveCalls *= by
	tl.PrecompileEcpairingMillerLoops *= by
	tl.PrecompileEcpairingG2MembershipCalls *= by
	tl.BlockKeccak *= by
	tl.BlockTransactions *= by
	tl.ShomeiMerkleProofs *= by
	tl.U128 *= by
	tl.U20 *= by
	tl.U32 *= by
	tl.U36 *= by
	tl.U64 *= by
	// beta v4.0
	tl.PrecompileBlsPointEvaluationEffectiveCalls *= by
	tl.PrecompilePointEvaluationFailureEffectiveCalls *= by
	tl.PrecompileBlsG1AddEffectiveCalls *= by
	tl.PrecompileBlsG1MsmEffectiveCalls *= by
	tl.PrecompileBlsG2AddEffectiveCalls *= by
	tl.PrecompileBlsG2MsmEffectiveCalls *= by
	tl.PrecompileBlsPairingCheckMillerLoops *= by
	tl.PrecompileBlsFinalExponentiations *= by
	tl.PrecompileBlsMapFpToG1EffectiveCalls *= by
	tl.PrecompileBlsMapFp2ToG2EffectiveCalls *= by
	tl.PrecompileBlsC1MembershipChecks *= by
	tl.PrecompileBlsC2MembershipCalls *= by
	tl.PrecompileBlsG1MembershipCalls *= by
	tl.PrecompileBlsG2MembershipCalls *= by
	tl.Bls *= by
	tl.PointEval *= by
	tl.BlsData *= by
	tl.RlpUtils *= by
}

func GetTestTracesLimits() *TracesLimits {

	// This are the config trace-limits from sepolia. All multiplied by 16.
	traceLimits := &TracesLimits{
		Add:                                  1 << 19,
		Bin:                                  1 << 18,
		Blake2Fmodexpdata:                    1 << 14,
		Blockdata:                            1 << 12,
		Blockhash:                            1 << 12,
		Ecdata:                               1 << 18,
		Euc:                                  1 << 16,
		Exp:                                  1 << 14,
		Ext:                                  1 << 20,
		Gas:                                  1 << 16,
		Hub:                                  1 << 21,
		Logdata:                              1 << 16,
		Loginfo:                              1 << 12,
		Mmio:                                 1 << 21,
		Mmu:                                  1 << 21,
		Mod:                                  1 << 17,
		Mul:                                  1 << 16,
		Mxp:                                  1 << 19,
		Oob:                                  1 << 18,
		Rlpaddr:                              1 << 12,
		Rlptxn:                               1 << 17,
		Rlptxrcpt:                            1 << 17,
		Rom:                                  1 << 22,
		Romlex:                               1 << 12,
		Shakiradata:                          1 << 15,
		Shf:                                  1 << 16,
		Stp:                                  1 << 14,
		Trm:                                  1 << 15,
		Txndata:                              1 << 14,
		Wcp:                                  1 << 18,
		Binreftable:                          1 << 20,
		Instdecoder:                          1 << 9,
		PrecompileEcrecoverEffectiveCalls:    1 << 9,
		PrecompileSha2Blocks:                 1 << 9,
		PrecompileRipemdBlocks:               0,
		PrecompileModexpEffectiveCalls:       1 << 10,
		PrecompileModexpEffectiveCalls4096:   1 << 4,
		PrecompileEcaddEffectiveCalls:        1 << 6,
		PrecompileEcmulEffectiveCalls:        1 << 6,
		PrecompileEcpairingEffectiveCalls:    1 << 4,
		PrecompileEcpairingMillerLoops:       1 << 4,
		PrecompileEcpairingG2MembershipCalls: 1 << 4,
		PrecompileBlakeEffectiveCalls:        0,
		PrecompileBlakeRounds:                0,
		BlockKeccak:                          1 << 13,
		BlockL1Size:                          100_000,
		BlockL2L1Logs:                        16,
		BlockTransactions:                    1 << 8,
		ShomeiMerkleProofs:                   1 << 14,
		U128:                                 1 << 17,
		U20:                                  1 << 17,
		U32:                                  1 << 17,
		U36:                                  1 << 17,
		U64:                                  1 << 17,
		// beta v4.0
		PrecompileBlsPointEvaluationEffectiveCalls:     0,
		PrecompilePointEvaluationFailureEffectiveCalls: 0,
		PrecompileBlsG1AddEffectiveCalls:               0,
		PrecompileBlsG1MsmEffectiveCalls:               0,
		PrecompileBlsG2AddEffectiveCalls:               0,
		PrecompileBlsG2MsmEffectiveCalls:               0,
		PrecompileBlsPairingCheckMillerLoops:           0,
		PrecompileBlsFinalExponentiations:              0,
		PrecompileBlsMapFpToG1EffectiveCalls:           0,
		PrecompileBlsMapFp2ToG2EffectiveCalls:          0,
		PrecompileBlsC1MembershipChecks:                0,
		PrecompileBlsC2MembershipCalls:                 0,
		PrecompileBlsG1MembershipCalls:                 0,
		PrecompileBlsG2MembershipCalls:                 0,
		Bls:                                            0,
		PointEval:                                      0,
		BlsData:                                        4096,
		RlpUtils:                                       131072,
		PowerReferenceTable:                            16,
		BlsReferenceTable:                              256,
	}

	return traceLimits
}
