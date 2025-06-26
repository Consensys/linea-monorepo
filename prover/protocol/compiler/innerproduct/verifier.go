package innerproduct

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/common/poly"
	"github.com/consensys/linea-monorepo/prover/maths/common/polyext"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// verifierForSize implements [wizard.VerifierAction]
type verifierForSize struct {
	// Queries is the list of queries involved with the current verification
	// step.
	Queries []query.InnerProduct
	// SummationOpening is the local opening at the end of Summation.
	SummationOpening query.LocalOpening
	// BatchOpening is the challenge used for the linear combination
	BatchOpening coin.Info
	skipped      bool
}

// Run implements [wizard.VerifierAction]
func (v *verifierForSize) Run(run wizard.Runtime) error {

	var (
		// ys stores the list of all the inner-product openings
		ys = []fext.Element{}
		// expected stores the random linear combinations of the ys by batching
		// coin
		expected fext.Element
		// actual stores the opening value of the last entry of Summation. The
		// verifier checks the equality between it and `expected`.
		actual = run.GetLocalPointEvalParams(v.SummationOpening.ID).ExtY
	)

	for _, q := range v.Queries {
		ipys := run.GetInnerProductParams(q.ID)
		ys = append(ys, ipys.Ys...)
	}

	if len(ys) > 1 {
		batchingCoin := run.GetRandomCoinFieldExt(v.BatchOpening.Name)
		expected = polyext.Eval(ys, batchingCoin)
	}

	if len(ys) <= 1 {
		expected = ys[0]
	}

	if actual != expected {
		return fmt.Errorf("inner-product verification failed %v != %v", actual.String(), expected.String())
	}

	return nil
}

// RunGnark implements the [wizard.VerifierAction] interface
func (v *verifierForSize) RunGnark(api frontend.API, run wizard.GnarkRuntime) {

	var (
		// ys stores the list of all the inner-product openings
		ys = []gnarkfext.Element{}
		// expected stores the random linear combinations of the ys by batching
		// coin
		expected frontend.Variable
		// actual stores the opening value of the last entry of Summation. The
		// verifier checks the equality between it and `expected`.
		actual = run.GetLocalPointEvalParams(v.SummationOpening.ID).ExtY
	)

	for _, q := range v.Queries {
		ipys := run.GetInnerProductParams(q.ID)
		ys = append(ys, ipys.Ys...)
	}

	if len(ys) > 1 {
		batchingCoin := run.GetRandomCoinFieldExt(v.BatchOpening.Name)
		expected = poly.EvaluateUnivariateGnarkExt(api, ys, batchingCoin)
	}

	if len(ys) <= 1 {
		expected = ys[0]
	}

	api.AssertIsEqual(expected, actual)
}

func (v *verifierForSize) Skip() {
	v.skipped = true
}

func (v *verifierForSize) IsSkipped() bool {
	return v.skipped
}
