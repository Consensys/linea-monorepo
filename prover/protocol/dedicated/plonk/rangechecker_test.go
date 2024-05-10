package plonk_test

import (
	"math/big"
	"testing"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/zkevm-monorepo/prover/protocol/dedicated/plonk"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This is a simple test circuit whose only goal is to utilize the range-checker
// so it triggers the external range-checker of the Wizard. This circuit does
// not use all the witness variables in gates and assumes that the missing gates
// are added.
type testRangeCheckingCircuitIncomplete struct {
	A [10]frontend.Variable
}

func (r *testRangeCheckingCircuitIncomplete) Define(api frontend.API) error {
	rangeChecker := api.(frontend.Rangechecker)

	for i := range r.A {
		rangeChecker.Check(r.A[i], 16)
	}
	return nil
}

// This is a simple test circuit whose only goal is to utilize the range-checker
// so it triggers the external range-checker of the Wizard. This circuit uses
// all the inputs in gates so there should exist a gate for each input.
type testRangeCheckingCircuitComplete struct {
	A [10]frontend.Variable
}

func (r *testRangeCheckingCircuitComplete) Define(api frontend.API) error {
	rangeChecker := api.(frontend.Rangechecker)

	for i := range r.A {
		api.AssertIsDifferent(0, r.A[i])
	}

	for i := range r.A {
		rangeChecker.Check(r.A[i], 16)
	}
	return nil
}

// This is a simple test circuit whose only goal is to utilize the range-checker
// so it triggers the external range-checker of the Wizard. This circuit does
// not use all the internal variables in gates and automatic gate addition
// should error.
type testRangeCheckingCircuitIncompleteInternal struct {
	A [10]frontend.Variable
}

func (r *testRangeCheckingCircuitIncompleteInternal) Define(api frontend.API) error {
	rangeChecker := api.(frontend.Rangechecker)

	for i := range r.A {
		rangeChecker.Check(r.A[i], 16)
	}
	// create an internal variable which is not witness.
	res, err := api.Compiler().NewHint(DummyHint, 1, frontend.Variable(0))
	if err != nil {
		return err
	}
	rangeChecker.Check(res[0], 16)
	return nil
}

func DummyHint(_ *big.Int, inputs []*big.Int, outputs []*big.Int) error {
	outputs[0].SetInt64(1)
	return nil
}

// This tests the correctness of the Plonk in Wizard implementation in a case
// where we use the external range-checker. The variables are in range, but the
// compiler should fail as the witness variables are not used in gates.
func TestRangeCheckIncompleteFails(t *testing.T) {

	circuit := &testRangeCheckingCircuitIncomplete{}

	assigner := func() frontend.Circuit {
		return &testRangeCheckingCircuitIncomplete{
			A: [10]frontend.Variable{
				field.NewElement(1),
				field.NewElement(1),
				field.NewElement(1),
				field.NewElement(1),
				field.NewElement(1),
				field.NewElement(1),
				field.NewElement(1),
				field.NewElement(1),
				field.NewElement(1),
				field.NewElement(1),
			},
		}
	}

	assert.Panics(t, func() {
		wizard.Compile(
			func(build *wizard.Builder) {
				plonk.PlonkCheck(
					build.CompiledIOP,
					"PLONK",
					0,
					circuit, []func() frontend.Circuit{assigner},
					plonk.WithRangecheck(16, 1, false),
				)
			},
			dummy.Compile,
		)
	})
}

// This tests the correctness of the Plonk in Wizard implementation in a case
// where we use the external range-checker. We use the incomplete circuit but
// builder should add the missing gates.
func TestRangeCheckIncompleteSucceeds(t *testing.T) {

	circuit := &testRangeCheckingCircuitIncomplete{}

	assigner := func() frontend.Circuit {
		return &testRangeCheckingCircuitIncomplete{
			A: [10]frontend.Variable{
				field.NewElement(1),
				field.NewElement(1),
				field.NewElement(1),
				field.NewElement(1),
				field.NewElement(1),
				field.NewElement(1),
				field.NewElement(1),
				field.NewElement(1),
				field.NewElement(1),
				field.NewElement(1),
			},
		}
	}

	compiled := wizard.Compile(
		func(build *wizard.Builder) {
			plonk.PlonkCheck(
				build.CompiledIOP,
				"PLONK",
				0,
				circuit, []func() frontend.Circuit{assigner},
				plonk.WithRangecheck(16, 1, true),
			)
		},
		dummy.Compile,
	)

	proof := wizard.Prove(compiled, func(assi *wizard.ProverRuntime) {})
	err := wizard.Verify(compiled, proof)
	require.NoError(t, err)
}

// This tests the correctness of the Plonk in Wizard implementation in a case
// where we use the external range-checker. This is a negative test and the
// prover should not accept the provided witness.
func TestRangeCheckNegative(t *testing.T) {

	circuit := &testRangeCheckingCircuitIncomplete{}

	assigner := func() frontend.Circuit {
		return &testRangeCheckingCircuitIncomplete{
			A: [10]frontend.Variable{
				// 0x10000000000000000000000000 = 2^100
				field.NewFromString("0x10000000000000000000000000"),
				field.NewFromString("0x10000000000000000000000000"),
				field.NewFromString("0x10000000000000000000000000"),
				field.NewFromString("0x10000000000000000000000000"),
				field.NewFromString("0x10000000000000000000000000"),
				field.NewFromString("0x10000000000000000000000000"),
				field.NewFromString("0x10000000000000000000000000"),
				field.NewFromString("0x10000000000000000000000000"),
				field.NewFromString("0x10000000000000000000000000"),
				field.NewFromString("0x10000000000000000000000000"),
			},
		}
	}

	compiled := wizard.Compile(
		func(build *wizard.Builder) {
			plonk.PlonkCheck(
				build.CompiledIOP,
				"PLONK",
				0,
				circuit, []func() frontend.Circuit{assigner},
				plonk.WithRangecheck(16, 1, true),
			)
		},
		dummy.Compile,
	)

	assert.Panics(t, func() {
		_ = wizard.Prove(compiled, func(assi *wizard.ProverRuntime) {})
	},
		"The prover should refuse to build a proof for this assignment since "+
			"the range check is for 2**16 and we provide 2**100",
	)
}

// This tests the correctness of the Plonk in Wizard implementation in a case
// where we use the external range-checker. The variables are in range and the
// prover should accept the provided witness. The circuit is complete as all the
// witness variables are used in gates.
func TestRangeCheckCompleteSucceeds(t *testing.T) {
	circuit := &testRangeCheckingCircuitComplete{}

	assigner := func() frontend.Circuit {
		return &testRangeCheckingCircuitComplete{
			A: [10]frontend.Variable{
				field.NewElement(1),
				field.NewElement(1),
				field.NewElement(1),
				field.NewElement(1),
				field.NewElement(1),
				field.NewElement(1),
				field.NewElement(1),
				field.NewElement(1),
				field.NewElement(1),
				field.NewElement(1),
			},
		}
	}

	compiled := wizard.Compile(
		func(build *wizard.Builder) {
			plonk.PlonkCheck(
				build.CompiledIOP,
				"PLONK",
				0,
				circuit, []func() frontend.Circuit{assigner},
				plonk.WithRangecheck(16, 1, false),
			)
		},
		dummy.Compile,
	)

	proof := wizard.Prove(compiled, func(assi *wizard.ProverRuntime) {})
	err := wizard.Verify(compiled, proof)
	require.NoError(t, err)
}

// This tests the correctness of the Plonk in Wizard implementation in a case
// where we use the external range-checker. The variables are in range, but the
// compiler should fail as internal variable is not used in gates and automatic
// gate addition should error.
func TestRangeCheckIncompleteInternalFails(t *testing.T) {

	circuit := &testRangeCheckingCircuitIncompleteInternal{}

	assigner := func() frontend.Circuit {
		return &testRangeCheckingCircuitIncompleteInternal{
			A: [10]frontend.Variable{
				field.NewElement(1),
				field.NewElement(1),
				field.NewElement(1),
				field.NewElement(1),
				field.NewElement(1),
				field.NewElement(1),
				field.NewElement(1),
				field.NewElement(1),
				field.NewElement(1),
				field.NewElement(1),
			},
		}
	}

	assert.Panics(t, func() {
		wizard.Compile(
			func(build *wizard.Builder) {
				plonk.PlonkCheck(
					build.CompiledIOP,
					"PLONK",
					0,
					circuit, []func() frontend.Circuit{assigner},
					plonk.WithRangecheck(16, 1, true),
				)
			},
			dummy.Compile,
		)
	})
}
