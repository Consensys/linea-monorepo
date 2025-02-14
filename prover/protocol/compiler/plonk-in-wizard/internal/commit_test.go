package plonk_test

import (
	"testing"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated/plonk"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/stretchr/testify/require"
)

// Circuit defines a simple circuit for test
// x**3 + x + 5 == y
type TestCommitCircuit struct {
	// struct tags on a variable is optional
	// default uses variable name and secret visibility.
	X frontend.Variable
	Y frontend.Variable `gnark:",public"`
}

// Define declares the circuit constraints
// x**3 + x + 5 == y
func (circuit *TestCommitCircuit) Define(api frontend.API) error {
	x3 := api.Mul(circuit.X, circuit.X, circuit.X)
	a := api.Add(x3, circuit.X, 5)

	// compute powers of a:
	powersOfA := []frontend.Variable{a}
	for i := 1; i < 15; i++ {
		powersOfA = append(powersOfA, api.Mul(powersOfA[i-1], powersOfA[i-1]))
	}

	// commit to powers of a
	committer := api.(frontend.Committer)
	_, err := committer.Commit(powersOfA...)
	if err != nil {
		return err
	}

	api.AssertIsEqual(circuit.Y, a)
	return nil
}

func TestPlonkWizardCircuitWithCommit(t *testing.T) {

	circuit := &TestCommitCircuit{}

	// Assigner is a function returning an assignment. It is called
	// as part of the prover runtime work, but the function is always
	// the same so it should be defined accounting for that
	assigner := func() frontend.Circuit {
		return &TestCommitCircuit{X: 0, Y: 5}
	}

	witnessAssigner := plonk.NewSafeCircuitAssigner(circuit, assigner)

	var pa plonk.PlonkInWizardProverAction

	compiled := wizard.Compile(
		func(build *wizard.Builder) {
			ctx := plonk.PlonkCheck(build.CompiledIOP, "PLONK", 0, circuit, 1)
			pa = ctx.GetPlonkProverAction()
		},
		dummy.Compile,
	)

	proof := wizard.Prove(compiled, func(run *wizard.ProverRuntime) {
		pa.Run(run, witnessAssigner)
	})

	err := wizard.Verify(compiled, proof)
	require.NoError(t, err)
}

func TestPlonkWizardCircuitWithCommitMultiInstance(t *testing.T) {

	circuit := &TestCommitCircuit{}

	// Assigner is a function returning an assignment. It is called
	// as part of the prover runtime work, but the function is always
	// the same so it should be defined accounting for that
	assigner := func() frontend.Circuit {
		return &TestCommitCircuit{X: 0, Y: 5}
	}

	witnessAssigner := plonk.NewSafeCircuitAssigner(circuit, assigner, assigner, assigner, assigner, assigner)

	var pa plonk.PlonkInWizardProverAction

	compiled := wizard.Compile(
		func(build *wizard.Builder) {
			ctx := plonk.PlonkCheck(build.CompiledIOP, "PLONK", 0, circuit, 5)
			pa = ctx.GetPlonkProverAction()
		},
		dummy.Compile,
	)

	proof := wizard.Prove(compiled, func(run *wizard.ProverRuntime) {
		pa.Run(run, witnessAssigner)
	})

	err := wizard.Verify(compiled, proof)
	require.NoError(t, err)
}
