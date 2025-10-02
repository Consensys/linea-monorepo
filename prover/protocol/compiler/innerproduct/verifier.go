package innerproduct

import (
	"fmt"

	"github.com/consensys/gnark-crypto/field/koalabear/vortex"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/common/poly"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"
)

// VerifierForSize implements [wizard.VerifierAction]
type VerifierForSize[T zk.Element] struct {
	// Queries is the list of queries involved with the current verification
	// step.
	Queries []query.InnerProduct[T]
	// SummationOpening is the local opening at the end of Summation.
	SummationOpening query.LocalOpening[T]
	// BatchOpening is the challenge used for the linear combination
	BatchOpening coin.Info[T]
	skipped      bool `serde:"omit"`
}

// Run implements [wizard.VerifierAction]
func (v *VerifierForSize[T]) Run(run wizard.Runtime[T]) error {

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
		expected = vortex.EvalFextPolyHorner(ys, batchingCoin)
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
func (v *VerifierForSize[T]) RunGnark(api frontend.API, run wizard.GnarkRuntime[T]) {

	var (
		// ys stores the list of all the inner-product openings
		ys       = []gnarkfext.E4Gen[T]{}
		expected gnarkfext.E4Gen[T] // expected stores the random linear combinations of the ys by batching
		// coin
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
		expected = poly.EvaluateUnivariateGnarkExt[T](api, ys, batchingCoin)
	}

	if len(ys) <= 1 {
		expected = ys[0]
	}

	api.AssertIsEqual(expected, actual)
}

func (v *VerifierForSize[T]) Skip() {
	v.skipped = true
}

func (v *VerifierForSize[T]) IsSkipped() bool {
	return v.skipped
}
