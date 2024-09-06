package main

import (
	"github.com/consensys/zkevm-monorepo/prover/config"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/arithmetization"
)

var globalArith *arithmetization.Arithmetization

func MakeDefine(cfg *config.Config) wizard.DefineFunc {
	return func(build *wizard.Builder) {
		globalArith = arithmetization.NewArithmetization(build, arithmetization.Settings{Limits: &cfg.TracesLimits})

	}
}

func MakeProver(traceFile string) wizard.ProverStep {
	return func(run *wizard.ProverRuntime) {
		globalArith.Assign(run, traceFile)
	}
}
