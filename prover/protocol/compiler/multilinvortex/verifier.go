package multilinvortex

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// verifierAction implements [wizard.VerifierAction] and checks Check 3:
//
//	Σ_b α^b · RowEvals_k[b] == UCols_k.Y
//
// for each column k. The MultilinearEval correctness of UCols_k and RowClaims_k
// is delegated to the next compiler pass.
type verifierAction struct {
	ctx     *context
	skipped bool `serde:"omit"`
}

// Run verifies the α-combination consistency between RowEvals and U_α.
func (v *verifierAction) Run(run wizard.Runtime) error {
	ctx := v.ctx
	nRowSize := 1 << ctx.NRow

	alpha := run.GetRandomCoinFieldExt(ctx.AlphaCoin.Name)

	// Pre-compute alpha powers.
	alphaPow := make([]fext.Element, nRowSize)
	alphaPow[0].SetOne()
	for b := 1; b < nRowSize; b++ {
		alphaPow[b].Mul(&alphaPow[b-1], &alpha)
	}

	for k := range ctx.InputQuery.Pols {
		// Compute Σ_b α^b · RowEvals[b] from the committed column.
		var computed fext.Element
		for b := 0; b < nRowSize; b++ {
			re := run.GetColumnAtExt(ctx.RowEvals[k].GetColID(), b)
			var t fext.Element
			t.Mul(&alphaPow[b], &re)
			computed.Add(&computed, &t)
		}

		// Read the claimed v_k from the U_α MultilinearEval params.
		uColParams := run.GetMultilinearParams(ctx.UCols[k].Name())
		if len(uColParams.Ys) == 0 {
			return fmt.Errorf("multilinvortex: UCols[%d] has no Ys", k)
		}
		claimed := uColParams.Ys[0]

		if !computed.Equal(&claimed) {
			return fmt.Errorf("multilinvortex: Check 3 failed for column %d: "+
				"Σ_b α^b·RowEvals[b] = %v, claimed UCols.Y = %v",
				k, computed.String(), claimed.String())
		}
	}
	return nil
}

// RunGnark is not yet implemented.
func (v *verifierAction) RunGnark(_ frontend.API, _ wizard.GnarkRuntime) {
	panic("multilinvortex.RunGnark: not yet implemented")
}

func (v *verifierAction) Skip()           { v.skipped = true }
func (v *verifierAction) IsSkipped() bool { return v.skipped }

// terminalVerifierAction handles 1-variable MultilinearEval queries directly:
// it checks P(r) = (1-r)*col[0] + r*col[1] for each polynomial in the query.
type terminalVerifierAction struct {
	q       query.MultilinearEval
	skipped bool `serde:"omit"`
}

func (v *terminalVerifierAction) Run(run wizard.Runtime) error {
	params := run.GetMultilinearParams(v.q.QueryID)

	for k, pol := range v.q.Pols {
		r := params.Points[k][0]
		var oneMinusR fext.Element
		oneMinusR.SetOne()
		oneMinusR.Sub(&oneMinusR, &r)

		val0 := run.GetColumnAtExt(pol.GetColID(), 0)
		val1 := run.GetColumnAtExt(pol.GetColID(), 1)

		var t0, t1, computed fext.Element
		t0.Mul(&oneMinusR, &val0)
		t1.Mul(&r, &val1)
		computed.Add(&t0, &t1)

		if !computed.Equal(&params.Ys[k]) {
			return fmt.Errorf("multilinvortex: terminal check failed for column %d: "+
				"(1-r)*col[0]+r*col[1] = %v, claimed y = %v",
				k, computed.String(), params.Ys[k].String())
		}
	}
	return nil
}

func (v *terminalVerifierAction) RunGnark(_ frontend.API, _ wizard.GnarkRuntime) {
	panic("multilinvortex.terminalVerifierAction.RunGnark: not yet implemented")
}

func (v *terminalVerifierAction) Skip()           { v.skipped = true }
func (v *terminalVerifierAction) IsSkipped() bool { return v.skipped }
