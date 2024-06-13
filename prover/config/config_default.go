package config

import "github.com/spf13/viper"

const _undecidedTraceValue = 1 << 10
const _undecidedTraceValueLarge = 1 << 12

var (
	DefaultDeferToOtherLargeCodes     = []int{137}     // List of exit codes for which the job will put back the job to be reexecuted in large mode.
	DefaultRetryLocallyWithLargeCodes = []int{77, 333} // List of exit codes for which the job will retry in large mode
)

func setDefaultValues() {
	setDefaultTracesLimit()
	setDefaultPaths()

	viper.SetDefault("debug.profiling", false)
	viper.SetDefault("debug.tracing", false)

	viper.SetDefault("controller.enable_execution", true)
	viper.SetDefault("controller.enable_blob_decompression", true)
	viper.SetDefault("controller.enable_aggregation", true)

	// Set the default values for the retry delays
	viper.SetDefault("controller.retry_delays", []int{0, 1, 2, 3, 5, 8, 13, 21, 44, 85})
	viper.SetDefault("controller.defer_to_other_large_codes", DefaultDeferToOtherLargeCodes)
	viper.SetDefault("controller.retry_locally_with_large_codes", DefaultRetryLocallyWithLargeCodes)

	// Set default for cmdTmpl and cmdLargeTmpl
	// TODO @gbotrel binary to run prover is hardcoded here.
	viper.SetDefault("controller.worker_cmd_tmpl", "prover prove --config {{.ConfFile}} --in {{.InFile}} --out {{.OutFile}}")
	viper.SetDefault("controller.worker_cmd_large_tmpl", "prover prove --config {{.ConfFile}} --in {{.InFile}} --out {{.OutFile}} --large")

}

func setDefaultPaths() {
	viper.SetDefault("execution.conflated_traces_dir", "/shared/traces/conflated")
	viper.SetDefault("execution.requests_root_dir", "/shared/prover-execution")
	viper.SetDefault("blob_decompression.requests_root_dir", "/shared/prover-compression")
	viper.SetDefault("aggregation.requests_root_dir", "/shared/prover-aggregation")
}

func setDefaultTracesLimit() {
	// TODO @gbotrel @AlexandreBelling check if we can remove this
	viper.SetDefault("traces_limits.MMU_ID", _defaultTraceValue)
	viper.SetDefault("traces_limits_large.MMU_ID", _defaultTraceValue)

	// Set the default values for the traces limits
	// TODO: check exact limits (@arithmetization team)

	viper.SetDefault("traces_limits.ADD", 524288)
	viper.SetDefault("traces_limits.BIN", 262144)
	viper.SetDefault("traces_limits.BIN_REF_TABLE", _undecidedTraceValue)
	viper.SetDefault("traces_limits.BIN_RT", 262144)
	viper.SetDefault("traces_limits.BLAKE2F_MOD_EXP_DATA", _undecidedTraceValue)
	viper.SetDefault("traces_limits.EC_DATA", 4096)
	viper.SetDefault("traces_limits.EXT", 131072)
	viper.SetDefault("traces_limits.EUC", _undecidedTraceValue)
	viper.SetDefault("traces_limits.HUB", 2097152)
	viper.SetDefault("traces_limits.INST_DECODER", 512)
	viper.SetDefault("traces_limits.LOG_DATA", _undecidedTraceValue)
	viper.SetDefault("traces_limits.LOG_INFO", _undecidedTraceValue)
	viper.SetDefault("traces_limits.MMIO", 131072)
	viper.SetDefault("traces_limits.MMU", 131072)
	viper.SetDefault("traces_limits.MOD", 131072)
	viper.SetDefault("traces_limits.MUL", 65536)
	viper.SetDefault("traces_limits.MXP", 524288)
	viper.SetDefault("traces_limits.PHONEY_RLP", 32768)
	viper.SetDefault("traces_limits.PUB_HASH", 32768)
	viper.SetDefault("traces_limits.PUB_HASH_INFO", 32768)
	viper.SetDefault("traces_limits.PUB_LOG_INFO", 16384)
	viper.SetDefault("traces_limits.RLP", 512)
	viper.SetDefault("traces_limits.RLP_ADDR", _undecidedTraceValue)
	viper.SetDefault("traces_limits.RLP_TXN", _undecidedTraceValue)
	viper.SetDefault("traces_limits.RLP_TX_RCPT", _undecidedTraceValue)
	viper.SetDefault("traces_limits.ROM", 4194304)
	viper.SetDefault("traces_limits.ROMLEX", _undecidedTraceValue)
	viper.SetDefault("traces_limits.SHAKIRA_DATA", _undecidedTraceValue)
	viper.SetDefault("traces_limits.SHF", 65536)
	viper.SetDefault("traces_limits.SHF_REF_TABLE", _undecidedTraceValue)
	viper.SetDefault("traces_limits.SHF_RT", 4096)
	viper.SetDefault("traces_limits.SIZE", 4194304)
	viper.SetDefault("traces_limits.STP", _undecidedTraceValue)
	viper.SetDefault("traces_limits.TRM", _undecidedTraceValue)
	viper.SetDefault("traces_limits.TX_RLP", 131072)
	viper.SetDefault("traces_limits.TXN_DATA", _undecidedTraceValue)
	viper.SetDefault("traces_limits.WCP", 262144)

	viper.SetDefault("traces_limits.BLOCK_DATA", _undecidedTraceValue)
	viper.SetDefault("traces_limits.BLOCK_HASH", _undecidedTraceValue)
	viper.SetDefault("traces_limits.BLOCK_KECCAK", 8192)
	viper.SetDefault("traces_limits.BLOCK_L2L1LOGS", 16)
	viper.SetDefault("traces_limits.BLOCK_TX", _undecidedTraceValue)

	viper.SetDefault("traces_limits.PRECOMPILE_ECRECOVER", 10000)
	viper.SetDefault("traces_limits.PRECOMPILE_SHA2", 10000)
	viper.SetDefault("traces_limits.PRECOMPILE_RIPEMD", 10000)
	viper.SetDefault("traces_limits.PRECOMPILE_IDENTITY", 10000)
	viper.SetDefault("traces_limits.PRECOMPILE_MODEXP", 10000)
	viper.SetDefault("traces_limits.PRECOMPILE_ECADD", 10000)
	viper.SetDefault("traces_limits.PRECOMPILE_ECMUL", 10000)
	viper.SetDefault("traces_limits.PRECOMPILE_ECPAIRING", 10000)
	viper.SetDefault("traces_limits.PRECOMPILE_BLAKE2F", 512)

	viper.SetDefault("traces_limits_large.ADD", 1048576)
	viper.SetDefault("traces_limits_large.BIN", 524288)
	viper.SetDefault("traces_limits_large.BIN_REF_TABLE", _undecidedTraceValueLarge)
	viper.SetDefault("traces_limits_large.BIN_RT", 524288)
	viper.SetDefault("traces_limits_large.BLAKE2F_MOD_EXP_DATA", _undecidedTraceValueLarge)
	viper.SetDefault("traces_limits_large.EC_DATA", 4096)
	viper.SetDefault("traces_limits_large.EXT", 262144)
	viper.SetDefault("traces_limits_large.EUC", _undecidedTraceValueLarge)
	viper.SetDefault("traces_limits_large.HUB", 4194304)
	viper.SetDefault("traces_limits_large.INST_DECODER", 512)
	viper.SetDefault("traces_limits_large.LOG_DATA", _undecidedTraceValueLarge)
	viper.SetDefault("traces_limits_large.LOG_INFO", _undecidedTraceValueLarge)
	viper.SetDefault("traces_limits_large.MMIO", 262144)
	viper.SetDefault("traces_limits_large.MMU", 262144)
	viper.SetDefault("traces_limits_large.MOD", 131072)
	viper.SetDefault("traces_limits_large.MUL", 131072)
	viper.SetDefault("traces_limits_large.MXP", 1048576)
	viper.SetDefault("traces_limits_large.PHONEY_RLP", 65536)
	viper.SetDefault("traces_limits_large.PUB_HASH", 65536)
	viper.SetDefault("traces_limits_large.PUB_HASH_INFO", 65536)
	viper.SetDefault("traces_limits_large.PUB_LOG_INFO", 32768)
	viper.SetDefault("traces_limits_large.RLP", 1024)
	viper.SetDefault("traces_limits_large.RLP_ADDR", _undecidedTraceValueLarge)
	viper.SetDefault("traces_limits_large.RLP_TXN", _undecidedTraceValueLarge)
	viper.SetDefault("traces_limits_large.RLP_TX_RCPT", _undecidedTraceValueLarge)
	viper.SetDefault("traces_limits_large.ROM", 8388608)
	viper.SetDefault("traces_limits_large.ROMLEX", _undecidedTraceValueLarge)
	viper.SetDefault("traces_limits_large.SHAKIRA_DATA", _undecidedTraceValueLarge)
	viper.SetDefault("traces_limits_large.SHF", 131072)
	viper.SetDefault("traces_limits_large.SHF_REF_TABLE", _undecidedTraceValueLarge)
	viper.SetDefault("traces_limits_large.SHF_RT", 4096)
	viper.SetDefault("traces_limits_large.SIZE", 8388608)
	viper.SetDefault("traces_limits_large.STP", _undecidedTraceValueLarge)
	viper.SetDefault("traces_limits_large.TRM", _undecidedTraceValueLarge)
	viper.SetDefault("traces_limits_large.TX_RLP", 524288)
	viper.SetDefault("traces_limits_large.TXN_DATA", _undecidedTraceValueLarge)
	viper.SetDefault("traces_limits_large.WCP", 524288)

	viper.SetDefault("traces_limits_large.BLOCK_DATA", _undecidedTraceValueLarge)
	viper.SetDefault("traces_limits_large.BLOCK_HASH", _undecidedTraceValueLarge)
	viper.SetDefault("traces_limits_large.BLOCK_KECCAK", 8192)
	viper.SetDefault("traces_limits_large.BLOCK_L2L1LOGS", 16)
	viper.SetDefault("traces_limits_large.BLOCK_TX", _undecidedTraceValueLarge)

	viper.SetDefault("traces_limits_large.PRECOMPILE_ECRECOVER", 10000)
	viper.SetDefault("traces_limits_large.PRECOMPILE_SHA2", 10000)
	viper.SetDefault("traces_limits_large.PRECOMPILE_RIPEMD", 10000)
	viper.SetDefault("traces_limits_large.PRECOMPILE_IDENTITY", 10000)
	viper.SetDefault("traces_limits_large.PRECOMPILE_MODEXP", 10000)
	viper.SetDefault("traces_limits_large.PRECOMPILE_ECADD", 10000)
	viper.SetDefault("traces_limits_large.PRECOMPILE_ECMUL", 10000)
	viper.SetDefault("traces_limits_large.PRECOMPILE_ECPAIRING", 10000)
	viper.SetDefault("traces_limits_large.PRECOMPILE_BLAKE2F", 512)

}
