package config

import (
	"time"

	v1 "github.com/consensys/linea-monorepo/prover/lib/compressor/blob/v1"
	"github.com/spf13/viper"
)

var (
	DefaultDeferToOtherLargeCodes     = []int{137}        // List of exit codes for which the job will put back the job to be reexecuted in large mode.
	DefaultRetryLocallyWithLargeCodes = []int{77, 333, 2} // List of exit codes for which the job will retry in large mode
)

func setDefaultValues() {

	setDefaultPaths()
	setDefaultExecLimitlessParams()
	setDefaultDebugFlags()
	setDefaultControllerParams()
	setDefaultDAParams()
	viper.SetDefault("execution.ignore_compatibility_check", false)
}

func setDefaultControllerParams() {
	viper.SetDefault("controller.enable_execution", false)
	viper.SetDefault("controller.enable_blob_decompression", false)
	viper.SetDefault("controller.enable_aggregation", false)

	viper.SetDefault("controller.limitless_jobs.enable_bootstrapper", false)
	viper.SetDefault("controller.limitless_jobs.enable_conglomerator", false)

	// Set the default values for the retry delays
	viper.SetDefault("controller.retry_delays", []int{0, 1, 2, 3, 5, 8, 13, 21, 44, 85})
	viper.SetDefault("controller.defer_to_other_large_codes", DefaultDeferToOtherLargeCodes)
	viper.SetDefault("controller.retry_locally_with_large_codes", DefaultRetryLocallyWithLargeCodes)
	viper.SetDefault("controller.spot_instance_reclaim_time_seconds", 120)
	viper.SetDefault("controller.termination_grace_period_seconds", 1800)

	viper.SetDefault("controller.spot_instance_reclaim_time_seconds", 120)
	viper.SetDefault("controller.termination_grace_period_seconds", 2700)

	// Set default for cmdTmpl and cmdLargeTmpl
	// TODO @gbotrel binary to run prover is hardcoded here.
	viper.SetDefault("controller.worker_cmd_tmpl", "prover prove --config {{.ConfFile}} --in {{.InFile}} --out {{.OutFile}}")
	viper.SetDefault("controller.worker_cmd_large_tmpl", "prover prove --config {{.ConfFile}} --in {{.InFile}} --out {{.OutFile}} --large")

	setDefaultProverPhaseCmds()
}

func setDefaultProverPhaseCmds() {

	// Set default cmds for limitless prover invoking the --phase flag
	var (
		bootstrapCmd      = "prover prove --phase=bootstrap --config {{.ConfFile}} --in {{.InFile}} --out {{.OutFile}}"
		glCmd             = "prover prove --phase=gl --config {{.ConfFile}} --in {{.InFile}} --out /dev/null"
		lppCmd            = "prover prove --phase=lpp --config {{.ConfFile}} --in {{.InFile}} --out /dev/null"
		conglomerationCmd = "prover prove --phase=conglomeration --config {{.ConfFile}} --in {{.InFile}} --out {{.OutFile}}"
	)

	viper.SetDefault("controller.prover_phase.bootstrap_cmd", bootstrapCmd)
	viper.SetDefault("controller.prover_phase.gl_cmd", glCmd)
	viper.SetDefault("controller.prover_phase.lpp_cmd", lppCmd)
	viper.SetDefault("controller.prover_phase.conglomeration_cmd", conglomerationCmd)
}

func setDefaultExecLimitlessParams() {
	viper.SetDefault("exec_limitless.preload_assets", true)
	viper.SetDefault("exec_limitless.shared_failure_dir", "/shared/tmp/exec-limitless/failure")
	viper.SetDefault("exec_limitless.metadata_dir", "/shared/tmp/exec-limitless/metadata")
	viper.SetDefault("exec_limitless.witness_dir", "/shared/tmp/exec-limitless/witness")
	viper.SetDefault("exec_limitless.subproofs_dir", "/shared/tmp/exec-limitless/subproofs")
	viper.SetDefault("exec_limitless.shared_rnd_dir", "/shared/tmp/exec-limitless/commit")
	viper.SetDefault("exec_limitless.poll_interval_seconds", 60)
	viper.SetDefault("exec_limitless.timeout_seconds", 2400)
}

func setDefaultDebugFlags() {
	viper.SetDefault("debug.profiling", false)
	viper.SetDefault("debug.tracing", false)
	viper.SetDefault("debug.performance_monitor.active", false)
	viper.SetDefault("debug.performance_monitor.sample_duration", 1*time.Second)
	viper.SetDefault("debug.performance_monitor.profile", "prover-rounds")
}

func setDefaultPaths() {
	viper.SetDefault("execution.conflated_traces_dir", "/shared/traces/conflated")
	viper.SetDefault("execution.requests_root_dir", "/shared/prover-execution")
	viper.SetDefault("data_availability.requests_root_dir", "/shared/prover-compression")
	viper.SetDefault("aggregation.requests_root_dir", "/shared/prover-aggregation")
	viper.SetDefault("debug.performance_monitor.profile_dir", "/shared/prover-execution/profiling")
}

func setDefaultDAParams() {
	viper.SetDefault("data_availability.max_nb_batches", 100)
	viper.SetDefault("data_availability.max_uncompressed_nb_bytes", v1.MaxUncompressedBytes)
	viper.SetDefault("data_availability.dict_nb_bytes", 65536)
}
