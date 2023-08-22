package plonk_test

import (
	"testing"

	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler/dummy"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/dedicated/plonk"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/consensys/gnark/frontend"
	"github.com/stretchr/testify/require"
)

// Circuit defines a simple circuit for test
// x**3 + x + 5 == y
type MyCircuit struct {
	// struct tags on a variable is optional
	// default uses variable name and secret visibility.
	X frontend.Variable `gnark:"x"`
	Y frontend.Variable `gnark:",public"`
}

// Define declares the circuit constraints
// x**3 + x + 5 == y
func (circuit *MyCircuit) Define(api frontend.API) error {
	x3 := api.Mul(circuit.X, circuit.X, circuit.X)
	a := api.Add(x3, circuit.X, 5)
	api.Println(a)
	api.AssertIsEqual(circuit.Y, api.Add(x3, circuit.X, 5))
	return nil
}

func TestPlonkWizard(t *testing.T) {

	t.SkipNow()

	circuit := &MyCircuit{}

	// Assigner is a function returning an assignment. It is called
	// as part of the prover runtime work, but the function is always
	// the same so it should be defined accounting for that
	assigner := func() frontend.Circuit {
		return &MyCircuit{X: 0, Y: 5}
	}

	//profiling.ProfileTrace("test-example", true, true, func() {
	compiled := wizard.Compile(
		func(build *wizard.Builder) {
			plonk.PlonkCheck(build.CompiledIOP, "PLONK", 0, circuit, []func() frontend.Circuit{assigner})
		},
		dummy.Compile,
	)

	proof := wizard.Prove(compiled, func(assi *wizard.ProverRuntime) {})
	err := wizard.Verify(compiled, proof)
	require.NoError(t, err)
	//})

}
