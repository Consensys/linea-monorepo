package plonkinternal_test

import (
	"testing"

	"github.com/consensys/gnark/backend/witness"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/zk"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/plonkinternal"
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

	apiGen, err := zk.NewGenericApi(api)
	if err != nil {
		return err
	}

	x3 := apiGen.Mul(circuit.X, circuit.X)
	x3 = apiGen.Mul(x3, circuit.X)
	wFive := zk.ValueOf(5)
	a := apiGen.Add(x3, circuit.X)
	a = apiGen.Add(a, wFive)

	// compute powers of a:
	powersOfA := []zk.WrappedVariable{zk.WrapFrontendVariable(a)}
	for i := 1; i < 15; i++ {
		powersOfA = append(powersOfA, apiGen.Mul(powersOfA[i-1], powersOfA[i-1]))
	}

	// commit to powers of a
	//committer := api.(frontend.Committer)
	//_, err := committer.Commit(powersOfA...)

	// TODO @thomas fixme
	// committer := api.(frontend.WideCommitter)
	// width is the size of the output slice of WideCommit, this 4 could be later interpreted as a field extension
	// _, err := committer.WideCommit(4, powersOfA...)
	// if err != nil {
	// 	return err
	// }

	apiGen.AssertIsEqual(circuit.Y, a)
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
		pa.Run(run, []witness.Witness{gnarkutil.AsWitnessPublicSmallField([]zk.WrappedVariable{zk.ValueOf(0), zk.ValueOf(5)})})
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
			gnarkutil.AsWitnessPublicSmallField([]zk.WrappedVariable{zk.ValueOf(0), zk.ValueOf(5)}),
			gnarkutil.AsWitnessPublicSmallField([]zk.WrappedVariable{zk.ValueOf(0), zk.ValueOf(5)}),
			gnarkutil.AsWitnessPublicSmallField([]zk.WrappedVariable{zk.ValueOf(0), zk.ValueOf(5)}),
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
			gnarkutil.AsWitnessPublicSmallField([]zk.WrappedVariable{zk.ValueOf(0), zk.ValueOf(5)}),
			gnarkutil.AsWitnessPublicSmallField([]zk.WrappedVariable{zk.ValueOf(0), zk.ValueOf(5)}),
			gnarkutil.AsWitnessPublicSmallField([]zk.WrappedVariable{zk.ValueOf(0), zk.ValueOf(5)}),
		})
	})

	err := wizard.Verify(compiled, proof)
	require.NoError(t, err)
}
