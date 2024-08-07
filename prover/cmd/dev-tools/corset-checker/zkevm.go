package main

import (
	"github.com/consensys/zkevm-monorepo/prover/config"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/arithmetization"
)

func MakeDefine(cfg *config.Config) wizard.DefineFunc {
	return func(b *wizard.Builder) {
		_ = arithmetization.NewArithmetization(
			b,
			arithmetization.Settings{Traces: &cfg.TracesLimits},
		)
	}
}

func MakeProver(traceFile string) wizard.ProverStep {
	return func(run *wizard.ProverRuntime) {
		arithmetization.Assign(run, traceFile)
	}
}
