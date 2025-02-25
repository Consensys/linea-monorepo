package main

import (
	"fmt"
	"os"

	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/crypto/ringsis"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/cleanup"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/mimc"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/selfrecursion"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/vortex"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

func main() {

	cfg, optConfig, traceFile, pErr := getParamsFromCLI()
	if pErr != nil {
		fmt.Printf("FATAL\n")
		fmt.Printf("err = %v\n", pErr)
		os.Exit(1)
	}

	suite := []func(*wizard.CompiledIOP){
		dummy.Compile,
	}

	if cfg.Execution.ProverMode == config.ProverModeBench {
		suite = fullSuite()
	}

	var (
		comp  = wizard.Compile(MakeDefine(cfg, optConfig), suite...)
		proof = wizard.Prove(comp, MakeProver(traceFile))
		vErr  = wizard.Verify(comp, proof)
	)

	if vErr == nil {
		fmt.Printf("PASSED\n")
	}

	if vErr != nil {
		fmt.Printf("FAILED\n")
		fmt.Printf("err = %v\n", vErr.Error())
		os.Exit(1)
	}

}

func fullSuite() []func(*wizard.CompiledIOP) {

	sisInstance := ringsis.Params{LogTwoBound: 16, LogTwoDegree: 6}

	return []func(*wizard.CompiledIOP){
		mimc.CompileMiMC,
		compiler.Arcane(1<<10, 1<<19, false),
		vortex.Compile(
			2,
			vortex.ForceNumOpenedColumns(256),
			vortex.WithSISParams(&sisInstance),
		),

		selfrecursion.SelfRecurse,
		cleanup.CleanUp,
		mimc.CompileMiMC,
		compiler.Arcane(1<<10, 1<<18, false),
		vortex.Compile(
			2,
			vortex.ForceNumOpenedColumns(256),
			vortex.WithSISParams(&sisInstance),
		),

		selfrecursion.SelfRecurse,
		cleanup.CleanUp,
		mimc.CompileMiMC,
		compiler.Arcane(1<<10, 1<<16, false),
		vortex.Compile(
			8,
			vortex.ForceNumOpenedColumns(64),
			vortex.WithSISParams(&sisInstance),
		),

		selfrecursion.SelfRecurse,
		cleanup.CleanUp,
		mimc.CompileMiMC,
		compiler.Arcane(1<<10, 1<<13, false),
		vortex.Compile(
			8,
			vortex.ForceNumOpenedColumns(64),
			vortex.ReplaceSisByMimc(),
		),
	}
}
