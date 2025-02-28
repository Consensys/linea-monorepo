package testing

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// Testcase is an object specifying how a wizard testcase protocol should
// both be defined and assigned.
type Testcase interface {
	Define(comp *wizard.CompiledIOP)
	Assign(run *wizard.ProverRuntime)
	MustFail() bool
}

// AnonymousTestcase is an implementation of testcase allowing the caller to
// explicitly provide his own functions for define, assign and mustFail.
type AnonymousTestcase struct {
	DefineFunc   func(comp *wizard.CompiledIOP)
	AssignFunc   func(run *wizard.ProverRuntime)
	MustFailFlag bool
}

// RunTestcase compiles and runs a testcase using the provided compilation
// suite. The function will attempt to run the prover and verify a proof
// then it will either succeed of fail the test depending on the requirements
// of the test.
func RunTestcase(t *testing.T, tc Testcase, suite []func(comp *wizard.CompiledIOP)) {

	define := func(b *wizard.Builder) {
		tc.Define(b.CompiledIOP)
	}

	comp := wizard.Compile(define, suite...)

	if tc.MustFail() {
		runTestShouldFail(t, comp, tc.Assign)
	}

	if !tc.MustFail() {
		runTestShouldPass(t, comp, tc.Assign)
	}
}

// RunAlteratedTestcase compiles and runs a testcase using the provided
// compilation suite. The function will attempt to run the prover and verify
// a randomly alterated version of the proof expecting the verifier to either
// reject or panic.
func RunAlteredTestcase(t *testing.T, tc Testcase, nbIterations int, suite []func(comp *wizard.CompiledIOP)) {

	if tc.MustFail() {
		t.Fatalf("[RunAlteredTestcase] needs a testcase where tc.MustFail() == false")
	}

	var (
		define = func(b *wizard.Builder) {
			tc.Define(b.CompiledIOP)
		}

		comp = wizard.Compile(define, suite...)
	)

	for iteration := 0; iteration < nbIterations; iteration++ {

		var (
			verErr, panicErr error
		)

		proof := wizard.Prove(comp, tc.Assign)
		AlterProof(&proof)

		panicErr = utils.RecoverPanic(func() {
			verErr = wizard.Verify(comp, proof)
		})

		if panicErr == nil && verErr == nil {
			t.Error("test was expected to fail but did not")
		}
	}

}

func runTestShouldPass(t *testing.T, comp *wizard.CompiledIOP, prover wizard.ProverStep) {

	var (
		proof = wizard.Prove(comp, prover)
		err   = wizard.Verify(comp, proof)
	)

	if err != nil {
		t.Errorf("verifier failed: %v", err)
	}
}

func runTestShouldFail(t *testing.T, comp *wizard.CompiledIOP, prover wizard.ProverStep) {

	var (
		verErr, panicErr error
		proof            wizard.Proof
	)

	panicErr = utils.RecoverPanic(func() {
		proof = wizard.Prove(comp, prover)
	})

	if panicErr != nil {
		return
	}

	panicErr = utils.RecoverPanic(func() {
		verErr = wizard.Verify(comp, proof)
	})

	if panicErr == nil && verErr == nil {
		t.Error("test was expected to fail but did not")
	}
}

func (a *AnonymousTestcase) Define(comp *wizard.CompiledIOP) {
	a.DefineFunc(comp)
}

func (a *AnonymousTestcase) Assign(run *wizard.ProverRuntime) {
	a.AssignFunc(run)
}

func (a *AnonymousTestcase) MustFail() bool {
	return a.MustFailFlag
}
