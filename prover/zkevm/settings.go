package zkevm

import (
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/arithmetization"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/hash/keccak"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/statemanager"
)

// type alias to denote a wizard-compilation suite. This is used when calling
// compile and provides internal parameters for the wizard package.
type compilationSuite = []func(*wizard.CompiledIOP)

// List the options set to initialize the zk-EVM
type Settings struct {
	Keccak       keccak.Settings
	Statemanager statemanager.SettingsLegacy
	// Settings object for the arithmetization
	Arithmetization  arithmetization.Settings
	CompilationSuite compilationSuite
}
