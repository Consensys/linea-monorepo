package innerproduct

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/common/poly"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// VerifierForSize implements [wizard.VerifierAction]
type VerifierForSize struct {
	// Queries is the list of queries involved with the current verification
	// step.
	Queries []query.InnerProduct
	// SummationOpening is the local opening at the end of Summation.
	SummationOpening query.LocalOpening
	// BatchOpening is the challenge used for the linear combination
	BatchOpening coin.Info
	skipped      bool `serde:"omit"`
}

// Run implements [wizard.VerifierAction]
func (v *VerifierForSize) Run(run wizard.Runtime) error {

	var (
		// ys stores the list of all the inner-product openings
		ys = []field.Element{}
		// expected stores the random linear combinations of the ys by batching
		// coin
		expected field.Element
		// actual stores the opening value of the last entry of Summation. The
		// verifier checks the equality between it and `expected`.
		actual = run.GetLocalPointEvalParams(v.SummationOpening.ID).Y
	)

	for _, q := range v.Queries {
		ipys := run.GetInnerProductParams(q.ID)
		ys = append(ys, ipys.Ys...)
	}

	if len(ys) > 1 {
		batchingCoin := run.GetRandomCoinField(v.BatchOpening.Name)
		expected = poly.EvalUnivariate(ys, batchingCoin)
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
func (v *VerifierForSize) RunGnark(api frontend.API, run wizard.GnarkRuntime) {

	var (
		// ys stores the list of all the inner-product openings
		ys = []frontend.Variable{}
		// expected stores the random linear combinations of the ys by batching
		// coin
		expected frontend.Variable
		// actual stores the opening value of the last entry of Summation. The
		// verifier checks the equality between it and `expected`.
		actual = run.GetLocalPointEvalParams(v.SummationOpening.ID).Y
	)

	for _, q := range v.Queries {
		ipys := run.GetInnerProductParams(q.ID)
		ys = append(ys, ipys.Ys...)
	}

	if len(ys) > 1 {
		batchingCoin := run.GetRandomCoinField(v.BatchOpening.Name)
		expected = poly.EvaluateUnivariateGnark(api, ys, batchingCoin)
	}

	if len(ys) <= 1 {
		expected = ys[0]
	}

	api.AssertIsEqual(expected, actual)
}

func (v *VerifierForSize) Skip() {
	v.skipped = true
}

func (v *VerifierForSize) IsSkipped() bool {
	return v.skipped
}
