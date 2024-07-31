package config

import (
	"bytes"
	"encoding/json"

	"github.com/consensys/zkevm-monorepo/prover/utils"
)

// The trace limits define the maximum trace size per module that the prover can handle.
// Raising these limits increases the prover's memory needs and simultaneously decreases the number
// of transactions it can prove in a single go. These traces are vital for the setup generator, so
// any changes in trace limits mean we'll need to run a new setup and update the verifier contracts
// before deploying.
type TracesLimits struct {
	Add               int `mapstructure:"ADD" validate:"power_of_2"`
	Bin               int `mapstructure:"BIN" validate:"power_of_2"`
	Blake2Fmodexpdata int `mapstructure:"BLAKE_MODEXP_DATA" validate:"power_of_2"`
	Blockdata         int `mapstructure:"BLOCK_DATA"`
	Blockhash         int `mapstructure:"BLOCK_HASH" validate:"power_of_2"`
	Ecdata            int `mapstructure:"EC_DATA" validate:"power_of_2"`
	Euc               int `mapstructure:"EUC" validate:"power_of_2"`
	Exp               int `mapstructure:"EXP" validate:"power_of_2"`
	Ext               int `mapstructure:"EXT" validate:"power_of_2"`
	Gas               int `mapstructure:"GAS" validate:"power_of_2"`
	Hub               int `mapstructure:"HUB" validate:"power_of_2"`
	Logdata           int `mapstructure:"LOG_DATA" validate:"power_of_2"`
	Loginfo           int `mapstructure:"LOG_INFO" validate:"power_of_2"`
	Mmio              int `mapstructure:"MMIO" validate:"power_of_2"`
	Mmu               int `mapstructure:"MMU" validate:"power_of_2"`
	Mod               int `mapstructure:"MOD" validate:"power_of_2"`
	Mul               int `mapstructure:"MUL" validate:"power_of_2"`
	Mxp               int `mapstructure:"MXP" validate:"power_of_2"`
	Oob               int `mapstructure:"OOB" validate:"power_of_2"`
	Rlpaddr           int `mapstructure:"RLP_ADDR" validate:"power_of_2"`
	Rlptxn            int `mapstructure:"RLP_TXN" validate:"power_of_2"`
	Rlptxrcpt         int `mapstructure:"RLP_TXN_RCPT" validate:"power_of_2"`
	Rom               int `mapstructure:"ROM" validate:"power_of_2"`
	Romlex            int `mapstructure:"ROM_LEX" validate:"power_of_2"`
	Shakiradata       int `mapstructure:"SHAKIRA_DATA" validate:"power_of_2"`
	Shf               int `mapstructure:"SHF" validate:"power_of_2"`
	Stp               int `mapstructure:"STP" validate:"power_of_2"`
	Trm               int `mapstructure:"TRM" validate:"power_of_2"`
	Txndata           int `mapstructure:"TXN_DATA" validate:"power_of_2"`
	Wcp               int `mapstructure:"WCP" validate:"power_of_2"`

	Binreftable int `mapstructure:"BIN_REFERENCE_TABLE" validate:"power_of_2"`
	Shfreftable int `mapstructure:"SHF_REFERENCE_TABLE" validate:"power_of_2"`
	Instdecoder int `mapstructure:"INSTRUCTION_DECODER" validate:"power_of_2"`

	PrecompileEcrecoverEffectiveCalls    int `mapstructure:"PRECOMPILE_ECRECOVER_EFFECTIVE_CALLS"`
	PrecompileSha2Blocks                 int `mapstructure:"PRECOMPILE_SHA2_BLOCKS"`
	PrecompileRipemdBlocks               int `mapstructure:"PRECOMPILE_RIPEMD_BLOCKS"`
	PrecompileModexpEffectiveCalls       int `mapstructure:"PRECOMPILE_MODEXP_EFFECTIVE_CALLS"`
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
