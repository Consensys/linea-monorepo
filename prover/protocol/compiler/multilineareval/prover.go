package multilineareval

import (
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/protocol/sumcheck"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// transcriptSeed is the fixed label used to initialise the sumcheck transcript
// for this compiler. Both prover and verifier must use the same seed.
const transcriptSeed = "multilineareval-compile-v1"

// proverAction implements [wizard.ProverAction] for one compiled batch.
type proverAction struct {
	ctx *context
}

// Run reads the lambda coin, builds the sumcheck claims from per-poly refs,
// runs the batched sumcheck prover, then assigns the round-poly proof column
// and residual MultilinearEval params.
func (p *proverAction) Run(run *wizard.ProverRuntime) {
	ctx := p.ctx
	n := ctx.NumVars

	lambda := run.GetRandomCoinFieldExt(ctx.LambdaCoin.Name)

	var claims []sumcheck.Claim
	var polys []sumcheck.MultiLin

	for _, ref := range ctx.PolRefs {
		params := run.GetMultilinearParams(ref.QueryName)
		claims = append(claims, sumcheck.Claim{
			Point: params.Points[ref.PolIdx],
			Eval:  params.Ys[ref.PolIdx],
		})
		colVec := run.GetColumn(ref.Col.GetColID()).IntoRegVecSaveAllocExt()
		polys = append(polys, sumcheck.MultiLin(colVec))
	}

	t := sumcheck.NewMockTranscript(transcriptSeed)
	t.Append("lambda", lambda)

	proof, challenges, err := sumcheck.ProveBatchedWith(claims, polys, lambda, t)
	if err != nil {
		panic(err)
	}

	// Flatten round polys into the proof column (padded to power-of-2 length).
	colSize := ctx.RoundPolys.Size()
	flat := make([]fext.Element, colSize)
	for k := 0; k < n; k++ {
		flat[k*3+0] = proof.RoundPolys[k][0]
		flat[k*3+1] = proof.RoundPolys[k][1]
		flat[k*3+2] = proof.RoundPolys[k][2]
	}
	run.AssignColumn(ctx.RoundPolys.GetColID(), smartvectors.NewRegularExt(flat))

	// Residual: all polys evaluated at the shared sumcheck point c.
	residualPoints := make([][]fext.Element, len(ctx.PolRefs))
	for i := range ctx.PolRefs {
		residualPoints[i] = challenges
	}
	run.AssignMultilinearExt(ctx.Residual.Name(), residualPoints, proof.FinalEvals...)
}
