package config

import (
	"bytes"
	"encoding/json"
	"slices"
	"strings"

	"github.com/consensys/linea-monorepo/prover/utils"
)

// These are modules that are internal to the prover and for which we have a
// hard requirement that they are defined.
var (
	modulePrecompileEcrecoverEffectiveCalls    = "PRECOMPILE_ECRECOVER_EFFECTIVE_CALLS"
	modulePrecompileSha2Blocks                 = "PRECOMPILE_SHA2_BLOCKS"
	modulePrecompileRipemdBlocks               = "PRECOMPILE_RIPEMD_BLOCKS"
	modulePrecompileModexpEffectiveCalls       = "PRECOMPILE_MODEXP_EFFECTIVE_CALLS"
	modulePrecompileModexpEffectiveCalls8192   = "PRECOMPILE_MODEXP_EFFECTIVE_CALLS_4096"
	modulePrecompileEcaddEffectiveCalls        = "PRECOMPILE_ECADD_EFFECTIVE_CALLS"
	modulePrecompileEcmulEffectiveCalls        = "PRECOMPILE_ECMUL_EFFECTIVE_CALLS"
	modulePrecompileEcpairingEffectiveCalls    = "PRECOMPILE_ECPAIRING_FINAL_EXPONENTIATIONS"
	modulePrecompileEcpairingMillerLoops       = "PRECOMPILE_ECPAIRING_MILLER_LOOPS"
	modulePrecompileEcpairingG2MembershipCalls = "PRECOMPILE_ECPAIRING_G2_MEMBERSHIP_CALLS"
	modulePrecompileBlakeEffectiveCalls        = "PRECOMPILE_BLAKE_EFFECTIVE_CALLS"
	modulePrecompileBlakeRounds                = "PRECOMPILE_BLAKE_ROUNDS"
	moduleBlockKeccak                          = "BLOCK_KECCAK"
	moduleBlockL1Size                          = "BLOCK_L1_SIZE"
	moduleBlockL2L1Logs                        = "BLOCK_L2_L1_LOGS"
	moduleBlockTransactions                    = "BLOCK_TRANSACTIONS"
	moduleShomeiMerkleProofs                   = "SHOMEI_MERKLE_PROOFS"

	modulePrecompileBlsG1AddEffectiveCalls               = "PRECOMPILE_BLS_G1_ADD_EFFECTIVE_CALLS"
	modulePrecompileBlsG2AddEffectiveCalls               = "PRECOMPILE_BLS_G2_ADD_EFFECTIVE_CALLS"
	modulePrecompileBlsG1MsmEffectiveCalls               = "PRECOMPILE_BLS_G1_MSM_EFFECTIVE_CALLS"
	modulePrecompileBlsG2MsmEffectiveCalls               = "PRECOMPILE_BLS_G2_MSM_EFFECTIVE_CALLS"
	modulePrecompileBlsPairingCheckMillerLoops           = "PRECOMPILE_BLS_PAIRING_CHECK_MILLER_LOOPS"
	modulePrecompileBlsFinalExponentiations              = "PRECOMPILE_BLS_FINAL_EXPONENTIATIONS"
	modulePrecompileBlsG1MembershipCalls                 = "PRECOMPILE_BLS_G1_MEMBERSHIP_CALLS"
	modulePrecompileBlsG2MembershipCalls                 = "PRECOMPILE_BLS_G2_MEMBERSHIP_CALLS"
	modulePrecompileBlsMapFpToG1EffectiveCalls           = "PRECOMPILE_BLS_MAP_FP_TO_G1_EFFECTIVE_CALLS"
	modulePrecompileBlsMapFp2ToG2EffectiveCalls          = "PRECOMPILE_BLS_MAP_FP2_TO_G2_EFFECTIVE_CALLS"
	modulePrecompileBlsC1MembershipCalls                 = "PRECOMPILE_BLS_C1_MEMBERSHIP_CALLS"
	modulePrecompileBlsC2MembershipCalls                 = "PRECOMPILE_BLS_C2_MEMBERSHIP_CALLS"
	modulePrecompileBlsPointEvaluationEffectiveCalls     = "PRECOMPILE_BLS_POINT_EVALUATION_EFFECTIVE_CALLS"
	modulePrecompilePointEvaluationFailureEffectiveCalls = "PRECOMPILE_POINT_EVALUATION_FAILURE_EFFECTIVE_CALLS"
	modulePrecompileP256VerifyEffectiveCalls             = "PRECOMPILE_P256_VERIFY_EFFECTIVE_CALLS"
)

// TracesLimits defines the limits for each module. The way limit work is by
// letting the user specifies prefixes of modules and the size to map to the
// module. The modules are sorted in reverse alphabetical order. When the user
// provides a module name (string), the string is checked against the prefixes
// in reverse alphabetical order.
//
// For instance, the limit map the followings:
//
// log2			= 1000, 2000
// log2_u256 	= 100, 	200
// log2_u160	= 5, 	10
// log2_u32		= 12, 	24
//
// If the user provides the string "log2_u32", the limit will be 12. And if the
// user provides log2_u16 (for which u160 is NOT a prefix) the limit will be
// 1000 because it will match "log2".
//
// Passing a module whose name is the empty string will be used as the default
// module.
//
// At runtime, the user may switch the trace limit file in "Large" mode by
// calling [SetLargeMode].
type TracesLimits struct {
	// toLargeMode is a flag telling the current struct to return LimitLarge
	// values instead of Limit when a module is requested.
	toLargeMode bool
	// scalingFactor is the factor used to scale the limit when a module is
	// requested. This is useful for the limitless prover when it needs to retry
	// with a larger limit. If the value is 0, this is equivalent to 1.
	scalingFactor int
	// Modules is the list of modules and their limits
	Modules []ModuleLimit `mapstructure:"modules" validate:"required"`
}

// ModuleLimit defines the limit for each module.
type ModuleLimit struct {
	Module     string `mapstructure:"module" validate:"required"`
	Limit      int    `mapstructure:"limit" validate:"required,power_of_2"`
	LimitLarge int    `mapstructure:"limit_large" validate:"required,power_of_2"`
	// IsNotScalable indicates that the module limit should never be scaled up.
	// This can be used for "MAX_L2_BLOCK_SIZE" or "L2L1Logs" whose size is
	// fixed for the proof system.
	IsNotScalable bool `mapstructure:"is_not_scalable"`
}

// normalizeToLowercase normalizes the modules to lowercase
// normalizeToLowercase returns the limits corresponding to the
// name of the method. (auto-generated)
func (tl *TracesLimits) normalizeToLowercase() {
	for i := range tl.Modules {
		tl.Modules[i].Module = strings.ToLower(tl.Modules[i].Module)
	}
}

// sortReverseAlphabetical sorts the modules in reverse alphabetical order
// sortReverseAlphabetical returns the limits corresponding to the
// name of the method. (auto-generated)
func (tl *TracesLimits) sortReverseAlphabetical() {
	slices.SortStableFunc(tl.Modules, func(a, b ModuleLimit) int {
		switch {
		case a.Module < b.Module:
			return 1
		case a.Module > b.Module:
			return -1
		default:
			utils.Panic("module %++v is not unique, also have module %++v", a, b)
		}
		return 0 // unreachable
	})
}

// SetLargeMode the TracesLimits to use the large mode
// SetLargeMode returns the limits corresponding to the
// name of the method. (auto-generated)
func (tl *TracesLimits) SetLargeMode() {
	tl.toLargeMode = true
}

// CheckSum returns the checksum of the TraceLimits. Checksum returns the limits
// corresponding to the name of the method. (auto-generated)
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

// findModuleLimits returns the limits for a module or panics.
// mustFindModuleLimits returns the limits corresponding to the
// name of the method. (auto-generated)
func (tl *TracesLimits) mustFindModuleLimits(module string) ModuleLimit {
	moduleLower := strings.ToLower(module)
	for _, m := range tl.Modules {
		if strings.HasPrefix(moduleLower, m.Module) {
			return m
		}
	}
	utils.Panic("found no module limits for module %q", module)
	return ModuleLimit{}
}

// ScaleUp increases the scaling factor
func (tl *TracesLimits) ScaleUp(by int) {
	if tl.scalingFactor == 0 {
		tl.scalingFactor = 1
	}
	tl.scalingFactor *= by
}

// GetLimit returns the limits of a module
func (tl *TracesLimits) GetLimit(module string) int {

	var (
		ml  = tl.mustFindModuleLimits(module)
		res = ml.Limit
	)

	if tl.toLargeMode {
		res = ml.LimitLarge
	}

	if tl.scalingFactor == 0 {
		tl.scalingFactor = 1
	}

	if !ml.IsNotScalable {
		res *= tl.scalingFactor
	}

	return res
}

// PrecompileEcrecoverEffectiveCalls returns the limits corresponding to the
// name of the method. (auto-generated)
func (tl *TracesLimits) PrecompileEcrecoverEffectiveCalls() int {
	name := modulePrecompileEcrecoverEffectiveCalls
	return tl.GetLimit(name)
}

// PrecompileSha2Blocks returns the limits corresponding to the
// name of the method. (auto-generated)
func (tl *TracesLimits) PrecompileSha2Blocks() int {
	name := modulePrecompileSha2Blocks
	return tl.GetLimit(name)
}

// PrecompileRipemdBlocks returns the limits corresponding to the
// name of the method. (auto-generated)
func (tl *TracesLimits) PrecompileRipemdBlocks() int {
	name := modulePrecompileRipemdBlocks
	return tl.GetLimit(name)
}

// PrecompileModexpEffectiveCalls returns the limits corresponding to the
// name of the method. (auto-generated)
func (tl *TracesLimits) PrecompileModexpEffectiveCalls() int {
	name := modulePrecompileModexpEffectiveCalls
	return tl.GetLimit(name)
}

// PrecompileModexpEffectiveCalls8192 returns the limits corresponding to the
// name of the method. (auto-generated)
func (tl *TracesLimits) PrecompileModexpEffectiveCalls8192() int {
	name := modulePrecompileModexpEffectiveCalls8192
	return tl.GetLimit(name)
}

// PrecompileEcaddEffectiveCalls returns the limits corresponding to the
// name of the method. (auto-generated)
func (tl *TracesLimits) PrecompileEcaddEffectiveCalls() int {
	name := modulePrecompileEcaddEffectiveCalls
	return tl.GetLimit(name)
}

// PrecompileEcmulEffectiveCalls returns the limits corresponding to the
// name of the method. (auto-generated)
func (tl *TracesLimits) PrecompileEcmulEffectiveCalls() int {
	name := modulePrecompileEcmulEffectiveCalls
	return tl.GetLimit(name)
}

// PrecompileEcpairingEffectiveCalls returns the limits corresponding to the
// name of the method. (auto-generated)
func (tl *TracesLimits) PrecompileEcpairingEffectiveCalls() int {
	name := modulePrecompileEcpairingEffectiveCalls
	return tl.GetLimit(name)
}

// PrecompileEcpairingMillerLoops returns the limits corresponding to the
// name of the method. (auto-generated)
func (tl *TracesLimits) PrecompileEcpairingMillerLoops() int {
	name := modulePrecompileEcpairingMillerLoops
	return tl.GetLimit(name)
}

// PrecompileEcpairingG2MembershipCalls returns the limits corresponding to the
// name of the method. (auto-generated)
func (tl *TracesLimits) PrecompileEcpairingG2MembershipCalls() int {
	name := modulePrecompileEcpairingG2MembershipCalls
	return tl.GetLimit(name)
}

// PrecompileBlakeEffectiveCalls returns the limits corresponding to the
// name of the method. (auto-generated)
func (tl *TracesLimits) PrecompileBlakeEffectiveCalls() int {
	name := modulePrecompileBlakeEffectiveCalls
	return tl.GetLimit(name)
}

// PrecompileBlakeRounds returns the limits corresponding to the
// name of the method. (auto-generated)
func (tl *TracesLimits) PrecompileBlakeRounds() int {
	name := modulePrecompileBlakeRounds
	return tl.GetLimit(name)
}

// BlockKeccak returns the limits corresponding to the
// name of the method. (auto-generated)
func (tl *TracesLimits) BlockKeccak() int {
	name := moduleBlockKeccak
	return tl.GetLimit(name)
}

// BlockL1Size returns the limits corresponding to the
// name of the method. (auto-generated)
func (tl *TracesLimits) BlockL1Size() int {
	name := moduleBlockL1Size
	return tl.GetLimit(name)
}

// BlockL2L1Logs returns the limits corresponding to the
// name of the method. (auto-generated)
func (tl *TracesLimits) BlockL2L1Logs() int {
	name := moduleBlockL2L1Logs
	return tl.GetLimit(name)
}

// BlockTransactions returns the limits corresponding to the
// name of the method. (auto-generated)
func (tl *TracesLimits) BlockTransactions() int {
	name := moduleBlockTransactions
	return tl.GetLimit(name)
}

// ShomeiMerkleProofs returns the limits corresponding to the
// name of the method. (auto-generated)
func (tl *TracesLimits) ShomeiMerkleProofs() int {
	name := moduleShomeiMerkleProofs
	return tl.GetLimit(name)
}

// PrecompileBlsG1AddEffectiveCalls returns the limits corresponding to the
// name of the method. (auto-generated)
func (tl *TracesLimits) PrecompileBlsG1AddEffectiveCalls() int {
	name := modulePrecompileBlsG1AddEffectiveCalls
	return tl.GetLimit(name)
}

// PrecompileBlsG2AddEffectiveCalls returns the limits corresponding to the
// name of the method. (auto-generated)
func (tl *TracesLimits) PrecompileBlsG2AddEffectiveCalls() int {
	name := modulePrecompileBlsG2AddEffectiveCalls
	return tl.GetLimit(name)
}

// PrecompileBlsG1MsmEffectiveCalls returns the limits corresponding to the
// name of the method. (auto-generated)
func (tl *TracesLimits) PrecompileBlsG1MsmEffectiveCalls() int {
	name := modulePrecompileBlsG1MsmEffectiveCalls
	return tl.GetLimit(name)
}

// PrecompileBlsG2MsmEffectiveCalls returns the limits corresponding to the
// name of the method. (auto-generated)
func (tl *TracesLimits) PrecompileBlsG2MsmEffectiveCalls() int {
	name := modulePrecompileBlsG2MsmEffectiveCalls
	return tl.GetLimit(name)
}

// PrecompileBlsPairingCheckMillerLoops returns the limits corresponding to the
// name of the method. (auto-generated)
func (tl *TracesLimits) PrecompileBlsPairingCheckMillerLoops() int {
	name := modulePrecompileBlsPairingCheckMillerLoops
	return tl.GetLimit(name)
}

// PrecompileBlsFinalExponentiations returns the limits corresponding to the
// name of the method. (auto-generated)
func (tl *TracesLimits) PrecompileBlsFinalExponentiations() int {
	name := modulePrecompileBlsFinalExponentiations
	return tl.GetLimit(name)
}

// PrecompileBlsG1MembershipCalls returns the limits corresponding to the
// name of the method. (auto-generated)
func (tl *TracesLimits) PrecompileBlsG1MembershipCalls() int {
	name := modulePrecompileBlsG1MembershipCalls
	return tl.GetLimit(name)
}

// PrecompileBlsG2MembershipCalls returns the limits corresponding to the
// name of the method. (auto-generated)
func (tl *TracesLimits) PrecompileBlsG2MembershipCalls() int {
	name := modulePrecompileBlsG2MembershipCalls
	return tl.GetLimit(name)
}

// PrecompileBlsMapFpToG1EffectiveCalls returns the limits corresponding to the
// name of the method. (auto-generated)
func (tl *TracesLimits) PrecompileBlsMapFpToG1EffectiveCalls() int {
	name := modulePrecompileBlsMapFpToG1EffectiveCalls
	return tl.GetLimit(name)
}

// PrecompileBlsMapFp2ToG2EffectiveCalls returns the limits corresponding to the
// name of the method. (auto-generated)
func (tl *TracesLimits) PrecompileBlsMapFp2ToG2EffectiveCalls() int {
	name := modulePrecompileBlsMapFp2ToG2EffectiveCalls
	return tl.GetLimit(name)
}

// PrecompileBlsC1MembershipCalls returns the limits corresponding to the
// name of the method. (auto-generated)
func (tl *TracesLimits) PrecompileBlsC1MembershipCalls() int {
	name := modulePrecompileBlsC1MembershipCalls
	return tl.GetLimit(name)
}

// PrecompileBlsC2MembershipCalls returns the limits corresponding to the
// name of the method. (auto-generated)
func (tl *TracesLimits) PrecompileBlsC2MembershipCalls() int {
	name := modulePrecompileBlsC2MembershipCalls
	return tl.GetLimit(name)
}

// PrecompileBlsPointEvaluationEffectiveCalls returns the limits corresponding
// to the name of the method. (auto-generated)
func (tl *TracesLimits) PrecompileBlsPointEvaluationEffectiveCalls() int {
	name := modulePrecompileBlsPointEvaluationEffectiveCalls
	return tl.GetLimit(name)
}

// PrecompilePointEvaluationFailureEffectiveCalls returns the limits
// corresponding to the name of the method. (auto-generated)
func (tl *TracesLimits) PrecompilePointEvaluationFailureEffectiveCalls() int {
	name := modulePrecompilePointEvaluationFailureEffectiveCalls
	return tl.GetLimit(name)
}

// PrecompileP256VerifyEffectiveCalls returns the limits corresponding to the
// name of the method. (auto-generated)
func (tl *TracesLimits) PrecompileP256VerifyEffectiveCalls() int {
	name := modulePrecompileP256VerifyEffectiveCalls
	return tl.GetLimit(name)
}

// GetTestTracesLimits returns a sample of the trace limits.
func GetTestTracesLimits() *TracesLimits {

	// This are the config trace-limits from sepolia. All multiplied by 16.
	traceLimits := &TracesLimits{
		Modules: []ModuleLimit{
			{Module: "Add", Limit: 1 << 19, LimitLarge: 1 << 19},
			{Module: "Bin", Limit: 1 << 18, LimitLarge: 1 << 18},
			{Module: "Blake_modexp_data", Limit: 1 << 14, LimitLarge: 1 << 14},
			{Module: "Block_data", Limit: 1 << 12, LimitLarge: 1 << 12},
			{Module: "Block_hash", Limit: 1 << 12, LimitLarge: 1 << 12},
			{Module: "Ec_data", Limit: 1 << 18, LimitLarge: 1 << 18},
			{Module: "Euc", Limit: 1 << 16, LimitLarge: 1 << 16},
			{Module: "Exp", Limit: 1 << 14, LimitLarge: 1 << 14},
			{Module: "Ext", Limit: 1 << 20, LimitLarge: 1 << 20},
			{Module: "Gas", Limit: 1 << 16, LimitLarge: 1 << 16},
			{Module: "Hub", Limit: 1 << 21, LimitLarge: 1 << 21},
			{Module: "Log_data", Limit: 1 << 16, LimitLarge: 1 << 16},
			{Module: "Log_info", Limit: 1 << 12, LimitLarge: 1 << 12},
			{Module: "Mmio", Limit: 1 << 21, LimitLarge: 1 << 21},
			{Module: "Mmu", Limit: 1 << 21, LimitLarge: 1 << 21},
			{Module: "Mod", Limit: 1 << 17, LimitLarge: 1 << 17},
			{Module: "Mul", Limit: 1 << 16, LimitLarge: 1 << 16},
			{Module: "Mxp", Limit: 1 << 19, LimitLarge: 1 << 19},
			{Module: "Oob", Limit: 1 << 18, LimitLarge: 1 << 18},
			{Module: "Rlp_addr", Limit: 1 << 12, LimitLarge: 1 << 12},
			{Module: "Rlp_txn", Limit: 1 << 17, LimitLarge: 1 << 17},
			{Module: "Rlp_txn_rcpt", Limit: 1 << 17, LimitLarge: 1 << 17},
			{Module: "Rom", Limit: 1 << 22, LimitLarge: 1 << 22},
			{Module: "Rom_lex", Limit: 1 << 12, LimitLarge: 1 << 12},
			{Module: "Shakira_data", Limit: 1 << 15, LimitLarge: 1 << 15},
			{Module: "Shf", Limit: 1 << 16, LimitLarge: 1 << 16},
			{Module: "Stp", Limit: 1 << 14, LimitLarge: 1 << 14},
			{Module: "Trm", Limit: 1 << 15, LimitLarge: 1 << 15},
			{Module: "Txn_data", Limit: 1 << 14, LimitLarge: 1 << 14},
			{Module: "Wcp", Limit: 1 << 18, LimitLarge: 1 << 18},
			{Module: "BIN_REFERENCE_TABLE", Limit: 1 << 20, LimitLarge: 1 << 20, IsNotScalable: true},
			{Module: "SHF_REFERENCE_TABLE", Limit: 1 << 12, LimitLarge: 1 << 12, IsNotScalable: true},
			{Module: "INSTRUCTION_DECODER", Limit: 1 << 9, LimitLarge: 1 << 9, IsNotScalable: true},
			{Module: "Precompile_Ecrecover_Effective_Calls", Limit: 1 << 9, LimitLarge: 1 << 9},
			{Module: "Precompile_Sha2_Blocks", Limit: 1 << 9, LimitLarge: 1 << 9},
			{Module: "Precompile_Ripemd_Blocks", Limit: 0, LimitLarge: 0},
			{Module: "Precompile_Modexp_Effective_Calls", Limit: 1 << 10, LimitLarge: 1 << 10},
			{Module: "Precompile_Modexp_Effective_Calls8192", Limit: 1 << 4, LimitLarge: 1 << 4},
			{Module: "Precompile_Ecadd_Effective_Calls", Limit: 1 << 6, LimitLarge: 1 << 6},
			{Module: "Precompile_Ecmul_Effective_Calls", Limit: 1 << 6, LimitLarge: 1 << 6},
			{Module: "Precompile_Ecpairing_Effective_Calls", Limit: 1 << 4, LimitLarge: 1 << 4},
			{Module: "Precompile_Ecpairing_Miller_Loops", Limit: 1 << 4, LimitLarge: 1 << 4},
			{Module: "Precompile_Ecpairing_G2_Membership_Calls", Limit: 1 << 4, LimitLarge: 1 << 4},
			{Module: "Precompile_Blake_Effective_Calls", Limit: 0, LimitLarge: 0},
			{Module: "Precompile_Blake_Rounds", Limit: 0, LimitLarge: 0},
			{Module: "Block_Keccak", Limit: 1 << 13, LimitLarge: 1 << 13},
			{Module: "Block_L1_Size", Limit: 100_000, LimitLarge: 100_000, IsNotScalable: true},
			{Module: "Block_L2_L1_Logs", Limit: 16, LimitLarge: 16, IsNotScalable: true},
			{Module: "Block_Transactions", Limit: 1 << 8, LimitLarge: 1 << 8},
			{Module: "Shomei_Merkle_Proofs", Limit: 1 << 14, LimitLarge: 1 << 14},
			{Module: "U128", Limit: 1 << 17, LimitLarge: 1 << 17},
			{Module: "U20", Limit: 1 << 17, LimitLarge: 1 << 17},
			{Module: "U32", Limit: 1 << 17, LimitLarge: 1 << 17},
			{Module: "U36", Limit: 1 << 17, LimitLarge: 1 << 17},
			{Module: "U64", Limit: 1 << 17, LimitLarge: 1 << 17},
		},
	}

	return traceLimits
}
