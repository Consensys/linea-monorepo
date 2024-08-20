package arithmetization

import (
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/arithmetization/define"
)

// Settings specifies the parameters for the arithmetization part of the zkEVM.
type Settings = define.Settings

// Arithmetization exposes all the methods relevant for the user to interact
// with the arithmetization of the zkEVM. It is a sub-component of the whole
// ZkEvm object as it does not includes the precompiles, the keccaks and the
// signature verification.
type Arithmetization struct {
	Settings *Settings
}

// NewArithmetization is the function that declares all the columns and the constraints of
// the zkEVM in the input builder object.
func NewArithmetization(builder *wizard.Builder, settings Settings) *Arithmetization {
	// wrapped works as an adapter between the define.Define and the
	// wizard.Builder. This value is only relevant during the execution of the
	// current function and it should not be reused thereafter.
	wrapped := define.Builder{
		Settings: &settings,
	}
	wrapped.Define(builder)

	registerMissingColumns(builder, &settings)
	return &Arithmetization{
		Settings: &settings,
	}
}

// Assign the arithmetization related columns. Namely, it will open the file
// specified in the witness object, call corset and assign the prover runtime
// columns.
func Assign(run *wizard.ProverRuntime, traceFile string) {
	// @Alex: This opens and reads the conflated trace file
	AssignFromCorset(
		traceFile,
		run,
	)
}

// registerMissingColumn registers columns that exists in the arithmetization
// but are omitted in the define.go as they are unconstrained due to the hub
// being missing.
func registerMissingColumns(b *wizard.Builder, limits *Settings) {
	b.RegisterCommit("shakiradata.LIMB", limits.Traces.Shakiradata)
	b.RegisterCommit("blake2fmodexpdata.LIMB", limits.Traces.Blake2Fmodexpdata)
}
