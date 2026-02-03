package main

import (
	"github.com/consensys/go-corset/pkg/ir/mir"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/config"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/zkevm/arithmetization"
)

var globalArith *arithmetization.Arithmetization

func MakeDefine(cfg *config.Config, optConfig *mir.OptimisationConfig) wizard.DefineFunc {
	return func(build *wizard.Builder) {
		globalArith = arithmetization.NewArithmetization(build, arithmetization.Settings{
			Limits:                   &cfg.TracesLimits,
			OptimisationLevel:        optConfig,
			IgnoreCompatibilityCheck: &cfg.Execution.IgnoreCompatibilityCheck,
		})

	}
}

func MakeProver(traceFile string) wizard.MainProverStep {
	return func(run *wizard.ProverRuntime) {
		globalArith.Assign(run, traceFile)
	}
}
