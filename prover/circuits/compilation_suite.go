package circuits

import (
	"github.com/consensys/linea-monorepo/prover/crypto/ringsis"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/cleanup"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/logdata"
	mimcComp "github.com/consensys/linea-monorepo/prover/protocol/compiler/mimc"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/selfrecursion"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/vortex"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// type alias to denote a wizard-compilation suite. This is used when calling
// compile and provides internal parameters for the wizard package.
type compilationSuite = []func(*wizard.CompiledIOP)

// it return the compilation suite for wizard hashing.
func WizardCompilationParameters() []func(iop *wizard.CompiledIOP) {
	var (
		sisInstance = ringsis.Params{LogTwoBound: 16, LogTwoDegree: 6}

		fullCompilationSuite = compilationSuite{

			compiler.Arcane(
				compiler.WithTargetColSize(1<<18),
				compiler.WithStitcherMinSize(1<<8),
			),
			logdata.Log("after vortex"),
			vortex.Compile(
				2,
				vortex.ForceNumOpenedColumns(256),
				vortex.WithSISParams(&sisInstance),
			),

			selfrecursion.SelfRecurse,
			cleanup.CleanUp,
			mimcComp.CompileMiMC,
			compiler.Arcane(
				compiler.WithTargetColSize(1<<16),
				compiler.WithStitcherMinSize(1<<8),
			),
			vortex.Compile(
				8,
				vortex.ForceNumOpenedColumns(64),
				vortex.WithSISParams(&sisInstance),
			),

			selfrecursion.SelfRecurse,
			cleanup.CleanUp,
			mimcComp.CompileMiMC,
			compiler.Arcane(
				compiler.WithTargetColSize(1<<13),
				compiler.WithStitcherMinSize(1<<8),
			),
			vortex.Compile(
				8,
				vortex.ForceNumOpenedColumns(64),
				vortex.WithOptionalSISHashingThreshold(1<<20),
			),
		}
	)

	return fullCompilationSuite

}
