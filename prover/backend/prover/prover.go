package prover

import (
	"github.com/consensys/accelerated-crypto-monorepo/backend/prover/dummycircuit"
	"github.com/consensys/accelerated-crypto-monorepo/backend/prover/outercircuit"
	"github.com/consensys/accelerated-crypto-monorepo/backend/prover/plonkutil"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/sirupsen/logrus"
)

// MustProveAndPass the prover (in the void). Does not takes a
// prover-step function performing the assignment but a function
// returning such a function. This is important to avoid side-effects
// when calling it twice.
func MustProveAndPass(
	mode string,
	mainProverStep func(*wizard.ProverRuntime),
	options *ProverOptions,
	publicInput field.Element,
) (proofHexString string) {

	switch mode {
	case NOTRACE:
		// We use this mode when, we want to test the integration of the prover
		// with other components without having to run the prover properly speaking
		// In this case, we simply generate a dummy proof using the light-prover's
		// circuit.
		logrus.Infof("Running the prover in no-trace mode")
		// regenerate the setup
		pp := dummycircuit.GenPublicParamsUnsafe()
		return dummycircuit.MakeProof(pp, publicInput)

	case LIGHT:
		logrus.Info("Running the LIGHT prover")
		// If in light-mode, we still want to make sure that the trace
		// was correctly generated even though we do not prove it.
		// Run the checker with all the options
		checker := getChecker(&ProverOptions{})
		proof := wizard.Prove(checker, mainProverStep)
		if err := wizard.Verify(checker, proof); err != nil {
			utils.Panic("(LIGHT-MODE ONLY) the checker did not pass: %v", err)
		}

		// And run the light-prover with only the main steps
		lightIOP := getDevLightIOP()
		proof = wizard.Prove(lightIOP, mainProverStep)

		// sanity-checks the proof so that we do not propagate errors
		if err := wizard.Verify(lightIOP, proof); err != nil {
			utils.Panic("the prover did not pass: %v", err)
		}

		// regenerate the setup
		pp := dummycircuit.GenPublicParamsUnsafe()
		return dummycircuit.MakeProof(pp, publicInput)

	case FULL, FULL_LARGE:
		logrus.Info("Running the FULL prover")

		// Run the full prover to obtain the intermediate proof
		logrus.Info("Get Full IOP")
		fullIOP := GetFullIOP(options)

		// Parse the setup
		logrus.Info("Fetching the outer-proof's public parameters")
		setup := make(chan plonkutil.Setup, 1)
		go plonkutil.ReadPPFromConfig(setup)

		proof := wizard.Prove(fullIOP, options.ApplyProverSteps(mainProverStep))

		logrus.Info("Sanity-checking the inner-proof")
		// sanity-checks the proof so that we do not propagate errors
		if err := wizard.Verify(fullIOP, proof); err != nil {
			utils.Panic("the prover did not pass: %v", err)
		}

		pp := <-setup
		return outercircuit.MakeProof(pp, fullIOP, proof, publicInput)

	case CHECKER:
		logrus.Info("Running the CHECKER prover")
		// Run the checker
		checker := getChecker(options)
		proof := wizard.Prove(checker, options.ApplyProverSteps(mainProverStep))
		if err := wizard.Verify(checker, proof); err != nil {
			utils.Panic("(CHECKER MODE) the checker did not pass: %v", err)
		}

		// The checker is not expected to return a proof
		// so we return the empty string
		return ""
	}

	panic("unreachable")
}

// MustPass the prover (in the void)
func MustPassCheck(p func(*wizard.ProverRuntime), po *ProverOptions) {
	compiled := CompiledCheckerIOP(po)
	proof := wizard.Prove(compiled, p)
	if err := wizard.Verify(compiled, proof); err != nil {
		utils.Panic("the checker did not pass: %v", err)
	}
}
