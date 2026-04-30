package multilinvortex

import (
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/protocol/sumcheck"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
)

// proverAction implements [wizard.ProverAction].
type ProverAction struct {
	ctx *context
}

// Run computes UAlpha and RowEvals for each input column, then assigns all
// committed columns and MultilinearEval params.
func (p *ProverAction) Run(run *wizard.ProverRuntime) {
	ctx := p.ctx
	nRow := ctx.NRow
	nCol := ctx.NCol

	alpha := run.GetRandomCoinFieldExt(ctx.AlphaCoin.Name)

	// Read the shared evaluation point from the input query params.
	inputParams := run.GetMultilinearParams(ctx.InputQuery.Name())
	// Points[0] holds the shared evaluation point (all polys in the batch share it).
	// First nRow coordinates are c_row; last nCol coordinates are c_col.
	cRow := inputParams.Points[0][:nRow]
	cCol := inputParams.Points[0][nRow:]

	// Pre-compute alpha powers: alphaPow[b] = α^b for b = 0,...,2^nRow-1.
	nRowSize := 1 << nRow
	alphaPow := make([]fext.Element, nRowSize)
	alphaPow[0].SetOne()
	for b := 1; b < nRowSize; b++ {
		alphaPow[b].Mul(&alphaPow[b-1], &alpha)
	}

	nColSize := 1 << nCol

	// alphaPow, cRow, cCol, and nRowSize/nColSize are read-only once built, so
	// they can be captured by value across parallel goroutines.
	parallel.Execute(len(ctx.InputQuery.Pols), func(start, stop int) {
		for k := start; k < stop; k++ {
			pol := ctx.InputQuery.Pols[k]
			colData := run.GetColumn(pol.GetColID()).IntoRegVecSaveAllocExt()

			// Compute RowEvals[b] = MultilinEval(colData[b*nColSize:(b+1)*nColSize], cCol)
			// and UAlpha[col] = Σ_b α^b · colData[b*nColSize + col].
			rowEvalsVec := make([]fext.Element, nRowSize)
			uAlphaVec := make([]fext.Element, nColSize)

			for b := 0; b < nRowSize; b++ {
				row := colData[b*nColSize : (b+1)*nColSize]
				// Fold in-place (row is a sub-slice of the freshly allocated colData).
				rowML := sumcheck.MultiLin(row).Clone()
				for _, r := range cCol {
					rowML.Fold(r)
				}
				rowEvalsVec[b] = rowML[0]

				// Accumulate into UAlpha: uAlphaVec[col] += alphaPow[b] * row[col]
				for col := 0; col < nColSize; col++ {
					var t fext.Element
					t.Mul(&alphaPow[b], &row[col])
					uAlphaVec[col].Add(&uAlphaVec[col], &t)
				}
			}

			// v_k = Σ_b α^b · RowEvals[b]
			var vk fext.Element
			for b := 0; b < nRowSize; b++ {
				var t fext.Element
				t.Mul(&alphaPow[b], &rowEvalsVec[b])
				vk.Add(&vk, &t)
			}

			run.AssignColumn(ctx.UAlpha[k].GetColID(), smartvectors.NewRegularExt(uAlphaVec))
			run.AssignColumn(ctx.RowEvals[k].GetColID(), smartvectors.NewRegularExt(rowEvalsVec))
			run.AssignMultilinearExt(ctx.UCols[k].Name(), [][]fext.Element{cCol}, vk)
			run.AssignMultilinearExt(ctx.RowClaims[k].Name(), [][]fext.Element{cRow}, inputParams.Ys[k])
		}
	})
}
