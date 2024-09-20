package main

import (
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/zkevm/arithmetization"
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
