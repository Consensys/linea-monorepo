package config

import (
	"bytes"
	"encoding/json"

	"github.com/consensys/zkevm-monorepo/prover/utils"
)

const _defaultTraceValue = 1 << 20

// The trace limits define the maximum trace size per module that the prover can handle.
// Raising these limits increases the prover's memory needs and simultaneously decreases the number
// of transactions it can prove in a single go. These traces are vital for the setup generator, so
// any changes in trace limits mean we'll need to run a new setup and update the verifier contracts
// before deploying.
type TracesLimits struct {
	Add                int `mapstructure:"ADD" validate:"power_of_2"`
	Bin                int `mapstructure:"BIN" validate:"power_of_2"`
	BinRt              int `mapstructure:"BIN_RT" validate:"power_of_2"`
	EcData             int `mapstructure:"EC_DATA" validate:"power_of_2"`
	Ext                int `mapstructure:"EXT" validate:"power_of_2"`
	Hub                int `mapstructure:"HUB" validate:"power_of_2"`
	InstructionDecoder int `mapstructure:"INSTRUCTION_DECODER" validate:"power_of_2"`
	Mmio               int `mapstructure:"MMIO" validate:"power_of_2"`
	Mmu                int `mapstructure:"MMU" validate:"power_of_2"`
	MmuID              int `mapstructure:"MMU_ID" validate:"power_of_2"`
	Mod                int `mapstructure:"MOD" validate:"power_of_2"`
	Mul                int `mapstructure:"MUL" validate:"power_of_2"`
	Mxp                int `mapstructure:"MXP" validate:"power_of_2"`
	PhoneyRlp          int `mapstructure:"PHONEY_RLP" validate:"power_of_2"`
	HashData           int `mapstructure:"PUB_HASH" validate:"power_of_2"`
	PubHashInfo        int `mapstructure:"PUB_HASH_INFO" validate:"power_of_2"`
	LogData            int `mapstructure:"PUB_LOG" validate:"power_of_2"`
	PubLogInfo         int `mapstructure:"PUB_LOG_INFO" validate:"power_of_2"`
	Rlp                int `mapstructure:"RLP" validate:"power_of_2"`
	Rom                int `mapstructure:"ROM" validate:"power_of_2"`
	Shf                int `mapstructure:"SHF" validate:"power_of_2"`
	ShfRt              int `mapstructure:"SHF_RT" validate:"power_of_2"`
	Size               int `mapstructure:"SIZE" validate:"power_of_2"`
	TxRlp              int `mapstructure:"TX_RLP" validate:"power_of_2"`
	Wcp                int `mapstructure:"WCP" validate:"power_of_2"`

	// Block specific limits
	BlockTx       int `mapstructure:"BLOCK_TX"`
	BlockL2L1Logs int `mapstructure:"BLOCK_L2L1LOGS"`
	BlockKeccak   int `mapstructure:"BLOCK_KECCAK"`

	// Precompiles limits
	PrecompileEcrecover int `mapstructure:"PRECOMPILE_ECRECOVER"`
	PrecompileSha2      int `mapstructure:"PRECOMPILE_SHA2"`
	PrecompileRipemd    int `mapstructure:"PRECOMPILE_RIPEMD"`
	PrecompileIdentity  int `mapstructure:"PRECOMPILE_IDENTITY"`
	PrecompileModexp    int `mapstructure:"PRECOMPILE_MODEXP"`
	PrecompileEcadd     int `mapstructure:"PRECOMPILE_ECADD"`
	PrecompileEcmul     int `mapstructure:"PRECOMPILE_ECMUL"`
	PrecompileEcpairing int `mapstructure:"PRECOMPILE_ECPAIRING"`
	PrecompileBlake2f   int `mapstructure:"PRECOMPILE_BLAKE2F"`
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
