package lookup

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/zkevm-monorepo/prover/maths/common/vector"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/query"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
)

// finalEvaluationCheck implements the [wizard.VerifierAction] interface. It
// represents the consistency check between SigmaT and the SigmaS[i]. This
// corresponds in the check #1 in the doc, where the verifier ensures that
// summing the ending values of the SigmaS[i]s yields the same result as the
// ending value of SigmaT.
type finalEvaluationCheck struct {
	// Name is a string repr of the verifier action. It is used to format errors
	// so that we can easily know which verifier action is at fault during the
	// verification is at fault.
	Name string
	// SigmaSOpenings is an [singleTableCtx.SigmaSOpenings]
	SigmaSOpenings []query.LocalOpening
	// SigmaTOpening is an [singleTableCtx.SigmaTOpening]
	SigmaTOpening []query.LocalOpening
}

// Run implements the [wizard.VerifierAction]
func (f *finalEvaluationCheck) Run(run *wizard.VerifierRuntime) error {

	var (
		sigmaTs = []field.Element{}
		sigmaSs = []field.Element{}
	)

	// SigmaSKSum stores the sum of the ending values of the SigmaSs as queried
	// in the protocol via the
	sigmaSKSum := field.Zero()
	for k := range f.SigmaSOpenings {
		temp := run.GetLocalPointEvalParams(f.SigmaSOpenings[k].ID).Y
		sigmaSs = append(sigmaSs, temp)
		sigmaSKSum.Add(&sigmaSKSum, &temp)
	}

	// SigmaTSum stores the ending value
	sigmaTSum := field.Zero()
	for frag := range f.SigmaTOpening {
		temp := run.GetLocalPointEvalParams(f.SigmaTOpening[frag].ID).Y
		sigmaTs = append(sigmaTs, temp)
		sigmaTSum.Add(&sigmaTSum, &temp)
	}

	if sigmaSKSum != sigmaTSum {
		fmt.Printf("sigmaTs = %v\nsigmaSs = %v\n", vector.Prettify(sigmaTs), vector.Prettify(sigmaSs))
		return fmt.Errorf("log-derivate lookup, the final evaluation check failed")
	}

	return nil
}

// RunGnark implements the [wizard.VerifierAction]
func (f *finalEvaluationCheck) RunGnark(api frontend.API, run *wizard.WizardVerifierCircuit) {

	// SigmaSKSum stores the sum of the ending values of the SigmaSs as queried
	// in the protocol via the
	sigmaSKSum := frontend.Variable(field.Zero())
	for k := range f.SigmaSOpenings {
		temp := run.GetLocalPointEvalParams(f.SigmaSOpenings[k].ID).Y
		sigmaSKSum = api.Add(sigmaSKSum, temp)
	}

	// SigmaTSum stores the ending value
	sigmaTSum := frontend.Variable(field.Zero())
	for frag := range f.SigmaTOpening {
		temp := run.GetLocalPointEvalParams(f.SigmaTOpening[frag].ID).Y
		sigmaTSum = api.Add(sigmaTSum, temp)
	}

	api.AssertIsEqual(sigmaSKSum, sigmaTSum)
}
