package xcomp

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/accessors"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

type GlobalVerifier struct {
	// PublicInputs are the inputs to the global verifier.
	PublicInputs []wizard.PublicInput
	skip         bool
}

func (v *GlobalVerifier) Run(run *wizard.VerifierRuntime) error {
	fmt.Printf("in the verifier\n")
	var (
		logDerivSum, grandProduct, grandSum field.Element
	)
	for _, pi := range v.PublicInputs {
		switch v := pi.Acc.(type) {

		case *accessors.FromLogDerivSumAccessor:
			fmt.Printf("logDerivSum %v\n", logDerivSum.String())
			logDerivCurr := v.GetVal(run)
			logDerivSum.Add(&logDerivSum, &logDerivCurr)

		case *accessors.FromGrandProductAccessor:
			grandProductCurr := v.GetVal(run)
			grandProduct.Mul(&grandProduct, &grandProductCurr)

		case *accessors.FromGrandSumAccessor:
			grandSumCurr := v.GetVal(run)
			grandSum.Mul(&grandSum, &grandSumCurr)
		}
	}

	if logDerivSum != field.Zero() {
		panic("the global sum over LogDerivSumParams is not zero")
	}

	if grandProduct != field.Zero() {
		panic("the global product overGrandProductParams is not zero")
	}

	if grandSum != field.Zero() {
		panic("the global sum over GrandSumParams is not zero")
	}
	return nil

}

// RunGnark implements the [wizard.VerifierAction]
func (v *GlobalVerifier) RunGnark(api frontend.API, run *wizard.WizardVerifierCircuit) {

	var (
		logDerivSum, grandProduct, grandSum frontend.Variable
	)
	for _, pi := range v.PublicInputs {
		switch v := pi.Acc.(type) {

		case *accessors.FromLogDerivSumAccessor:
			logDerivCurr := v.GetFrontendVariable(api, run)
			api.Add(logDerivSum, logDerivCurr)

		case *accessors.FromGrandProductAccessor:
			grandProductCurr := v.GetFrontendVariable(api, run)
			api.Mul(grandProduct, grandProductCurr)

		case *accessors.FromGrandSumAccessor:
			grandSumCurr := v.GetFrontendVariable(api, run)
			api.Mul(grandSum, grandSumCurr)
		}
	}

	api.AssertIsEqual(logDerivSum, field.Zero())
	api.AssertIsEqual(grandProduct, field.Zero())
	api.AssertIsEqual(grandSum, field.Zero())
}

func (v *GlobalVerifier) Skip() {
	v.skip = true
}

func (v *GlobalVerifier) IsSkipped() bool {
	return v.skip
}
