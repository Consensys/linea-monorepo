package plonkinternal_test

import (
	"testing"

	"github.com/consensys/gnark/backend/witness"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/internal/plonkinternal"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils/gnarkutil"
	"github.com/stretchr/testify/require"
)

// Circuit defines a simple circuit for test
// x**3 + x + 5 == y
type TestCommitCircuit struct {
	// struct tags on a variable is optional
	// default uses variable name and secret visibility.
	X zk.WrappedVariable `gnark:",public"`
	Y zk.WrappedVariable `gnark:",public"`
}

// Define declares the circuit constraints
// x**3 + x + 5 == y
func (circuit *TestCommitCircuit) Define(api frontend.API) error {
	x3 := api.Mul(circuit.X, circuit.X, circuit.X)
	a := api.Add(x3, circuit.X, 5)

	// compute powers of a:
	powersOfA := []zk.WrappedVariable{a}
	for i := 1; i < 15; i++ {
		powersOfA = append(powersOfA, api.Mul(powersOfA[i-1], powersOfA[i-1]))
	}

	// commit to powers of a
	//committer := api.(frontend.Committer)
	//_, err := committer.Commit(powersOfA...)

	committer := api.(frontend.WideCommitter)
	// width is the size of the output slice of WideCommit, this 4 could be later interpreted as a field extension
	_, err := committer.WideCommit(4, powersOfA...)
	if err != nil {
		return err
	}

	api.AssertIsEqual(circuit.Y, a)
	return nil
}

func TestPlonkWizardCircuitWithCommit(t *testing.T) {

	circuit := &TestCommitCircuit{}

	var pa plonkinternal.PlonkInWizardProverAction

	compiled := wizard.Compile(
		func(build *wizard.Builder) {
			ctx := plonkinternal.PlonkCheck(build.CompiledIOP, "PLONK", 0, circuit, 1)
			pa = ctx.GetPlonkProverAction()
		},
		dummy.Compile,
	)

	proof := wizard.Prove(compiled, func(run *wizard.ProverRuntime) {
		pa.Run(run, []witness.Witness{gnarkutil.AsWitnessPublicSmallField([]zk.WrappedVariable{0, 5})})
	})

	err := wizard.Verify(compiled, proof)
	require.NoError(t, err)
}

func TestPlonkWizardCircuitWithCommitMultiInstance(t *testing.T) {

	circuit := &TestCommitCircuit{}

	var pa plonkinternal.PlonkInWizardProverAction

	compiled := wizard.Compile(
		func(build *wizard.Builder) {
			ctx := plonkinternal.PlonkCheck(build.CompiledIOP, "PLONK", 0, circuit, 5)
			pa = ctx.GetPlonkProverAction()
		},
		dummy.Compile,
	)

	proof := wizard.Prove(compiled, func(run *wizard.ProverRuntime) {
		pa.Run(run, []witness.Witness{
			gnarkutil.AsWitnessPublicSmallField([]zk.WrappedVariable{0, 5}),
			gnarkutil.AsWitnessPublicSmallField([]zk.WrappedVariable{0, 5}),
			gnarkutil.AsWitnessPublicSmallField([]zk.WrappedVariable{0, 5}),
		})
	})

	err := wizard.Verify(compiled, proof)
	require.NoError(t, err)
}

func TestPlonkWizardCircuitWithCommitMultiInstanceFixedNbRow(t *testing.T) {

	circuit := &TestCommitCircuit{}

	var pa plonkinternal.PlonkInWizardProverAction

	compiled := wizard.Compile(
		func(build *wizard.Builder) {
			ctx := plonkinternal.PlonkCheck(build.CompiledIOP, "PLONK", 0, circuit, 5, plonkinternal.WithFixedNbRows(256))
			pa = ctx.GetPlonkProverAction()
		},
		dummy.Compile,
	)

	proof := wizard.Prove(compiled, func(run *wizard.ProverRuntime) {
		pa.Run(run, []witness.Witness{
			gnarkutil.AsWitnessPublic([]zk.WrappedVariable{0, 5}),
			gnarkutil.AsWitnessPublic([]zk.WrappedVariable{0, 5}),
			gnarkutil.AsWitnessPublic([]zk.WrappedVariable{0, 5}),
		})
	})

	err := wizard.Verify(compiled, proof)
	require.NoError(t, err)
}
