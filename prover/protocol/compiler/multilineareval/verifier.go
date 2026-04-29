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

	// Reconstruct claims from per-poly refs.
	var claims []sumcheck.Claim
	for _, ref := range ctx.PolRefs {
		params := run.GetMultilinearParams(ref.QueryName)
		claims = append(claims, sumcheck.Claim{
			Point: params.Points[ref.PolIdx],
			Eval:  params.Ys[ref.PolIdx],
		})
	}

	// Read the round polys from the proof column.
	roundPolys := make([][3]fext.Element, n)
	for k := 0; k < n; k++ {
		roundPolys[k][0] = run.GetColumnAtExt(ctx.RoundPolys.GetColID(), k*3+0)
		roundPolys[k][1] = run.GetColumnAtExt(ctx.RoundPolys.GetColID(), k*3+1)
		roundPolys[k][2] = run.GetColumnAtExt(ctx.RoundPolys.GetColID(), k*3+2)
	}

	// Read the residual params.
	residualParams := run.GetMultilinearParams(ctx.Residual.Name())

	proof := sumcheck.BatchedProof{
		RoundPolys: roundPolys,
		FinalEvals: residualParams.Ys,
	}

	t := sumcheck.NewMockTranscript(transcriptSeed)
	t.Append("lambda", lambda)

	challenges, _, err := sumcheck.VerifyBatchedWith(claims, proof, lambda, t)
	if err != nil {
		return fmt.Errorf("multilineareval verifier: %w", err)
	}

	// Confirm every residual poly's point equals the sumcheck challenges.
	for i := range ctx.PolRefs {
		if len(residualParams.Points[i]) != len(challenges) {
			return fmt.Errorf("multilineareval verifier: poly %d residual point length mismatch", i)
		}
		for j, c := range challenges {
			if !c.Equal(&residualParams.Points[i][j]) {
				return fmt.Errorf("multilineareval verifier: poly %d residual point[%d] mismatch", i, j)
			}
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
