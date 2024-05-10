package arithmetization

import (
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/arithmetization/define"
)

// Settings specifies the parameters for the arithmetization part of the zk-EVM.
type Settings = define.Settings

// Arithmetization exposes all the methods relevant for the user to interact
// with the arithmetization of the zkEVM. It is a sub-component of the whole
// ZkEvm object as it does not includes the precompiles, the keccaks and the
// signature verification.
type Arithmetization struct {
	Settings *Settings
}

// Define is the function that declares all the columns and the constraints of
// the zkEVM in the input builder object.
func (a *Arithmetization) Define(builder *wizard.Builder) {
	// wrapped works as an adapter between the define.Define and the
	// wizard.Builder. This value is only relevant during the execution of the
	// current function and it should not be reused thereafter.
	wrapped := define.Builder{
		Settings: a.Settings,
	}
	wrapped.Define(builder)
}

// Assign the arithmetization related columns. Namely, it will open the file
// specified in the witness object, call corset and assign the prover runtime
// columns.
func Assign(run *wizard.ProverRuntime, getTraces TraceGetter) {
	// @Alex: This opens and reads the conflated trace file
	traceReader := getTraces()
	defer traceReader.Close()
	AssignFromCorset(
		traceReader,
		run,
	)
}
