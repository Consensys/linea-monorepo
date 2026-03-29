package zkevm

import (
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/zkevm/arithmetization"
)

// type alias to denote a wizard-compilation suite. This is used when calling
// compile and provides internal parameters for the wizard package.
type CompilationSuite = []func(*wizard.CompiledIOP)

// List the options set to initialize the zkEVM
type Settings struct {
	// General parameters
	PreRecursionCompilationSuite  CompilationSuite
	PostRecursionCompilationSuite *CompilationSuite
	Metadata                      wizard.VersionMetadata
	Arithmetization               arithmetization.Settings
}
