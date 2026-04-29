package multilinvortex

import (
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/protocol/sumcheck"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// proverAction implements [wizard.ProverAction].
type proverAction struct {
	ctx *context
}

// Run computes UAlpha and RowEvals for each input column, then assigns all
// committed columns and MultilinearEval params.
func (p *proverAction) Run(run *wizard.ProverRuntime) {
	ctx := p.ctx
	nRow := ctx.NRow
	nCol := ctx.NCol

	alpha := run.GetRandomCoinFieldExt(ctx.AlphaCoin.Name)

	// Read the shared evaluation point from the input query params.
	inputParams := run.GetMultilinearParams(ctx.InputQuery.Name())
	// Point has n elements: first nRow are c_row, last nCol are c_col.
	cRow := inputParams.Point[:nRow]
	cCol := inputParams.Point[nRow:]

	// Pre-compute alpha powers: alphaPow[b] = α^b for b = 0,...,2^nRow-1.
	nRowSize := 1 << nRow
	alphaPow := make([]fext.Element, nRowSize)
	alphaPow[0].SetOne()
	for b := 1; b < nRowSize; b++ {
		alphaPow[b].Mul(&alphaPow[b-1], &alpha)
	}

	nColSize := 1 << nCol

	for k, pol := range ctx.InputQuery.Pols {
		colData := run.GetColumn(pol.GetColID()).IntoRegVecSaveAllocExt()

		// Compute RowEvals[b] = MultilinEval(colData[b*nColSize:(b+1)*nColSize], cCol)
		// and UAlpha[col] = Σ_b α^b · colData[b*nColSize + col].
		rowEvalsVec := make([]fext.Element, nRowSize)
		uAlphaVec := make([]fext.Element, nColSize)

		for b := 0; b < nRowSize; b++ {
			row := colData[b*nColSize : (b+1)*nColSize]
			rowML := sumcheck.MultiLin(row)
			rowEvalsVec[b] = rowML.Evaluate(cCol)

			// Accumulate into UAlpha: uAlphaVec[col] += alphaPow[b] * row[col]
			for col := 0; col < nColSize; col++ {
				var t fext.Element
				t.Mul(&alphaPow[b], &row[col])
				uAlphaVec[col].Add(&uAlphaVec[col], &t)
			}
		}

		// v_k = Σ_b α^b · RowEvals[b] (= UAlpha's multilinear eval at cCol)
		var vk fext.Element
		for b := 0; b < nRowSize; b++ {
			var t fext.Element
			t.Mul(&alphaPow[b], &rowEvalsVec[b])
			vk.Add(&vk, &t)
		}
		// Alternatively: vk = MultilinEval(uAlphaVec, cCol) — both must agree.

		run.AssignColumn(ctx.UAlpha[k].GetColID(), smartvectors.NewRegularExt(uAlphaVec))
		run.AssignColumn(ctx.RowEvals[k].GetColID(), smartvectors.NewRegularExt(rowEvalsVec))

		run.AssignMultilinearExt(ctx.UCols[k].Name(), cCol, vk)
		run.AssignMultilinearExt(ctx.RowClaims[k].Name(), cRow, inputParams.Ys[k])
	}
}
