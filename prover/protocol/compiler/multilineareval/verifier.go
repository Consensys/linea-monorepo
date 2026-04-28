package multilineareval

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/protocol/sumcheck"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// verifierAction implements [wizard.VerifierAction] for one compiled batch.
type verifierAction struct {
	ctx     *context
	skipped bool `serde:"omit"`
}

// Run verifies the batched sumcheck proof against the input claims and checks
// that the residual query params are consistent with the sumcheck output.
func (v *verifierAction) Run(run wizard.Runtime) error {
	ctx := v.ctx
	n := ctx.NumVars

	lambda := run.GetRandomCoinFieldExt(ctx.LambdaCoin.Name)

	// Reconstruct claims from input query params.
	var claims []sumcheck.Claim
	for _, q := range ctx.InputQueries {
		params := run.GetMultilinearParams(q.Name())
		for j := range q.Pols {
			claims = append(claims, sumcheck.Claim{
				Point: params.Point,
				Eval:  params.Ys[j],
			})
		}
	}

	// Read the round polys from the proof column.
	roundPolys := make([][3]fext.Element, n)
	for k := 0; k < n; k++ {
		roundPolys[k][0] = run.GetColumnAtExt(ctx.RoundPolys.GetColID(), k*3+0)
		roundPolys[k][1] = run.GetColumnAtExt(ctx.RoundPolys.GetColID(), k*3+1)
		roundPolys[k][2] = run.GetColumnAtExt(ctx.RoundPolys.GetColID(), k*3+2)
	}

	// Read the residual params — the final evals and residual point.
	residualParams := run.GetMultilinearParams(ctx.Residual.Name())

	proof := sumcheck.BatchedProof{
		RoundPolys: roundPolys,
		FinalEvals: residualParams.Ys,
	}

	// Verify the sumcheck using the same transcript as the prover.
	t := sumcheck.NewMockTranscript(transcriptSeed)
	t.Append("lambda", lambda)

	challenges, _, err := sumcheck.VerifyBatchedWith(claims, proof, lambda, t)
	if err != nil {
		return fmt.Errorf("multilineareval verifier: %w", err)
	}

	// Confirm the residual point matches the sumcheck challenges.
	if len(challenges) != len(residualParams.Point) {
		return fmt.Errorf("multilineareval verifier: challenge length mismatch")
	}
	for i, c := range challenges {
		if !c.Equal(&residualParams.Point[i]) {
			return fmt.Errorf("multilineareval verifier: residual point mismatch at index %d", i)
		}
	}

	return nil
}

// RunGnark is not yet implemented for the prototype.
func (v *verifierAction) RunGnark(_ frontend.API, _ wizard.GnarkRuntime) {
	panic("multilineareval.RunGnark: not yet implemented")
}

func (v *verifierAction) Skip() {
	v.skipped = true
}

func (v *verifierAction) IsSkipped() bool {
	return v.skipped
}
