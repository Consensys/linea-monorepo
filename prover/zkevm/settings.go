package zkevm

import (
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/zkevm/arithmetization"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/ecarith"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/ecdsa"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/ecpair"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/keccak"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/sha2"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/modexp"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/publicInput"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/statemanager"
)

// type alias to denote a wizard-compilation suite. This is used when calling
// compile and provides internal parameters for the wizard package.
type compilationSuite = []func(*wizard.CompiledIOP)

// List the options set to initialize the zkEVM
type Settings struct {
	Keccak           keccak.Settings
	Statemanager     statemanager.Settings
	Arithmetization  arithmetization.Settings
	Ecdsa            ecdsa.Settings
	Modexp           modexp.Settings
	Ecadd, Ecmul     ecarith.Limits
	Ecpair           ecpair.Limits
	Sha2             sha2.Settings
	PublicInput      publicInput.Settings
	CompilationSuite compilationSuite
	Metadata         wizard.VersionMetadata
}
