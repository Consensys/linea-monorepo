package multilinvortex

import (
	"runtime"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/protocol/sumcheck"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
)

// ProverAction implements [wizard.ProverAction].
type ProverAction struct {
	Ctx *Context
}

// Run computes UAlpha and RowEvals for each input column, then assigns all
// committed columns and MultilinearEval params.
//
// Row evaluation uses a precomputed eq(cCol, *) table decomposed into SoA
// (structure-of-arrays) form. Per row b:
//   - rowEval = Σ_j eq[j]·row[j] via 4 AVX-512 inner products (base-field)
//   - uAlpha[j] += α^b·row[j]    via 4 AVX-512 ScalarMul+Add passes
//
// Parallelism strategy (adaptive):
//   - K ≥ GOMAXPROCS: outer column parallelism (large-K path)
//   - K < GOMAXPROCS: inner row parallelism, K columns processed serially (small-K path)
//
// The small-K path pre-allocates and reuses the per-thread SoA partial buffers
// across all K serial iterations, avoiding repeated large allocations.
func (p *ProverAction) Run(run *wizard.ProverRuntime) {
	ctx := p.Ctx
	nRow := ctx.NRow
	nCol := ctx.NCol

	alpha := run.GetRandomCoinFieldExt(ctx.AlphaCoin.Name)

	inputParams := run.GetMultilinearParams(ctx.InputQuery.Name())
	cRow := inputParams.Points[0][:nRow]
	cCol := inputParams.Points[0][nRow:]

	nRowSize := 1 << nRow
	alphaPow := make([]fext.Element, nRowSize)
	alphaPow[0].SetOne()
	for b := 1; b < nRowSize; b++ {
		alphaPow[b].Mul(&alphaPow[b-1], &alpha)
	}

	nColSize := 1 << nCol
	K := len(ctx.InputQuery.Pols)

	eqColTable := sumcheck.BuildEqTable(cCol)
	eqRowTable := sumcheck.BuildEqTable(cRow)

	// eqColSoA[c][j] = component c of eq(cCol, j). Shared read-only by all goroutines.
	eqColSoA := buildSoA(eqColTable, nColSize)

	// Small-K path: for K < smallK the row-parallel path (one launch of
	// GOMAXPROCS goroutines per column) gives better utilization than launching
	// only K goroutines; for K ≥ smallK the outer column-parallelism already
	// saturates the machine.
	const smallK = 8

	if K < smallK {
		nthreads := runtime.GOMAXPROCS(0)
		if nthreads > nRowSize {
			nthreads = nRowSize
		}
		rowsPerThread := (nRowSize + nthreads - 1) / nthreads

		// Pre-allocate SoA partial buffers once; cleared and reused across K iterations.
		// rowEvalsVec is allocated fresh per k (it's small and passed to AssignColumn
		// by reference, so it must not be reused).
		partialUASoA := make([][4]field.Vector, nthreads)
		for t := range partialUASoA {
			for c := range 4 {
				partialUASoA[t][c] = make(field.Vector, nColSize)
			}
		}

		var packedUAlpha, packedRowEvals []fext.Element
		if ctx.Packed {
			packedUAlpha = make([]fext.Element, ctx.KPow2*nColSize)
			packedRowEvals = make([]fext.Element, ctx.KPow2*nRowSize)
		}

		for k := 0; k < K; k++ {
			// Clear SoA partials from the previous iteration.
			for t := range partialUASoA {
				for c := range 4 {
					clear(partialUASoA[t][c])
				}
			}

			pol := ctx.InputQuery.Pols[k]
			sv := run.GetColumn(pol.GetColID())

			// Allocate rowEvalsVec fresh per iteration: it's passed to AssignColumn
			// by reference and must not be overwritten by a future iteration's clear.
			rowEvalsVec := make([]fext.Element, nRowSize)

			colBase, isBase := sv.IntoRegVecSaveAllocBase()
			if isBase == nil {
				// Base-field path: SIMD inner product for rowEval + SIMD ScalarMul for uAlpha.
				parallel.Execute(nthreads, func(start, stop int) {
					tmpVec := make(field.Vector, nColSize)
					for t := start; t < stop; t++ {
						bStart := t * rowsPerThread
						bStop := bStart + rowsPerThread
						if bStop > nRowSize {
							bStop = nRowSize
						}
						ua := &partialUASoA[t]
						for b := bStart; b < bStop; b++ {
							rowBase := field.Vector(colBase[b*nColSize : (b+1)*nColSize])
							rowEvalsVec[b] = simdRowEvalBase(eqColSoA, rowBase)
							simdUAlphaAccumBase(ua, rowBase, &alphaPow[b], tmpVec)
						}
					}
				})

				uAlphaVec := mergeSoA(partialUASoA, nColSize)
				vk, actualY := computeVkAndActualY(alphaPow, eqRowTable, rowEvalsVec, nRowSize)

				if ctx.Packed {
					copy(packedUAlpha[k*nColSize:(k+1)*nColSize], uAlphaVec)
					copy(packedRowEvals[k*nRowSize:(k+1)*nRowSize], rowEvalsVec)
					run.AssignMultilinearExt(ctx.UCols[k].Name(), [][]fext.Element{locatorPoint(k, ctx.L, cCol)}, vk)
					run.AssignMultilinearExt(ctx.RowClaims[k].Name(), [][]fext.Element{locatorPoint(k, ctx.L, cRow)}, actualY)
				} else {
					run.AssignColumn(ctx.UAlpha[k].GetColID(), smartvectors.NewRegularExt(uAlphaVec))
					run.AssignColumn(ctx.RowEvals[k].GetColID(), smartvectors.NewRegularExt(rowEvalsVec))
					run.AssignMultilinearExt(ctx.UCols[k].Name(), [][]fext.Element{cCol}, vk)
					run.AssignMultilinearExt(ctx.RowClaims[k].Name(), [][]fext.Element{cRow}, actualY)
				}
			} else {
				// Fext path: columns are already fext (e.g. UAlpha/RowEvals from prior round).
				colDataExt := sv.IntoRegVecSaveAllocExt()
				partialUA := make([][]fext.Element, nthreads)
				for t := range partialUA {
					partialUA[t] = make([]fext.Element, nColSize)
				}

				parallel.Execute(nthreads, func(start, stop int) {
					for t := start; t < stop; t++ {
						bStart := t * rowsPerThread
						bStop := bStart + rowsPerThread
						if bStop > nRowSize {
							bStop = nRowSize
						}
						loc := partialUA[t]
						for b := bStart; b < bStop; b++ {
							rowExt := colDataExt[b*nColSize : (b+1)*nColSize]
							ap := alphaPow[b]
							var rowEval fext.Element
							for col := 0; col < nColSize; col++ {
								v := &rowExt[col]
								eq := &eqColTable[col]
								var r fext.Element
								r.Mul(eq, v)
								rowEval.Add(&rowEval, &r)
								var tmp fext.Element
								tmp.Mul(&ap, v)
								loc[col].Add(&loc[col], &tmp)
							}
							rowEvalsVec[b] = rowEval
						}
					}
				})

				uAlphaVec := make([]fext.Element, nColSize)
				for t := 0; t < nthreads; t++ {
					for col := 0; col < nColSize; col++ {
						uAlphaVec[col].Add(&uAlphaVec[col], &partialUA[t][col])
					}
				}

				vk, actualY := computeVkAndActualY(alphaPow, eqRowTable, rowEvalsVec, nRowSize)

				if ctx.Packed {
					copy(packedUAlpha[k*nColSize:(k+1)*nColSize], uAlphaVec)
					copy(packedRowEvals[k*nRowSize:(k+1)*nRowSize], rowEvalsVec)
					run.AssignMultilinearExt(ctx.UCols[k].Name(), [][]fext.Element{locatorPoint(k, ctx.L, cCol)}, vk)
					run.AssignMultilinearExt(ctx.RowClaims[k].Name(), [][]fext.Element{locatorPoint(k, ctx.L, cRow)}, actualY)
				} else {
					run.AssignColumn(ctx.UAlpha[k].GetColID(), smartvectors.NewRegularExt(uAlphaVec))
					run.AssignColumn(ctx.RowEvals[k].GetColID(), smartvectors.NewRegularExt(rowEvalsVec))
					run.AssignMultilinearExt(ctx.UCols[k].Name(), [][]fext.Element{cCol}, vk)
					run.AssignMultilinearExt(ctx.RowClaims[k].Name(), [][]fext.Element{cRow}, actualY)
				}
			}
		}

		if ctx.Packed {
			run.AssignColumn(ctx.UAlpha[0].GetColID(), smartvectors.NewRegularExt(packedUAlpha))
			run.AssignColumn(ctx.RowEvals[0].GetColID(), smartvectors.NewRegularExt(packedRowEvals))
		}
		return
	}

	// Large-K path: K ≥ GOMAXPROCS — outer parallelism over columns saturates all cores.
	var packedUAlpha, packedRowEvals []fext.Element
	if ctx.Packed {
		packedUAlpha = make([]fext.Element, ctx.KPow2*nColSize)
		packedRowEvals = make([]fext.Element, ctx.KPow2*nRowSize)
	}

	parallel.Execute(K, func(start, stop int) {
		tmpVec := make(field.Vector, nColSize)

		for k := start; k < stop; k++ {
			pol := ctx.InputQuery.Pols[k]
			sv := run.GetColumn(pol.GetColID())

			rowEvalsVec := make([]fext.Element, nRowSize)

			colBase, isBase := sv.IntoRegVecSaveAllocBase()
			if isBase == nil {
				// Base-field path: SoA uAlpha, SIMD inner products per row.
				uAlphaSoA := [4]field.Vector{
					make(field.Vector, nColSize),
					make(field.Vector, nColSize),
					make(field.Vector, nColSize),
					make(field.Vector, nColSize),
				}
				for b := 0; b < nRowSize; b++ {
					rowBase := field.Vector(colBase[b*nColSize : (b+1)*nColSize])
					rowEvalsVec[b] = simdRowEvalBase(eqColSoA, rowBase)
					simdUAlphaAccumBase(&uAlphaSoA, rowBase, &alphaPow[b], tmpVec)
				}
				uAlphaVec := soaToAoS(&uAlphaSoA, nColSize)
				vk, actualY := computeVkAndActualY(alphaPow, eqRowTable, rowEvalsVec, nRowSize)

				if ctx.Packed {
					copy(packedUAlpha[k*nColSize:(k+1)*nColSize], uAlphaVec)
					copy(packedRowEvals[k*nRowSize:(k+1)*nRowSize], rowEvalsVec)
					run.AssignMultilinearExt(ctx.UCols[k].Name(), [][]fext.Element{locatorPoint(k, ctx.L, cCol)}, vk)
					run.AssignMultilinearExt(ctx.RowClaims[k].Name(), [][]fext.Element{locatorPoint(k, ctx.L, cRow)}, actualY)
				} else {
					run.AssignColumn(ctx.UAlpha[k].GetColID(), smartvectors.NewRegularExt(uAlphaVec))
					run.AssignColumn(ctx.RowEvals[k].GetColID(), smartvectors.NewRegularExt(rowEvalsVec))
					run.AssignMultilinearExt(ctx.UCols[k].Name(), [][]fext.Element{cCol}, vk)
					run.AssignMultilinearExt(ctx.RowClaims[k].Name(), [][]fext.Element{cRow}, actualY)
				}
			} else {
				// Fext path: columns are already fext.
				colDataExt := sv.IntoRegVecSaveAllocExt()
				uAlphaVec := make([]fext.Element, nColSize)
				for b := 0; b < nRowSize; b++ {
					rowExt := colDataExt[b*nColSize : (b+1)*nColSize]
					ap := alphaPow[b]
					var rowEval fext.Element
					for col := 0; col < nColSize; col++ {
						v := &rowExt[col]
						eq := &eqColTable[col]
						var r fext.Element
						r.Mul(eq, v)
						rowEval.Add(&rowEval, &r)
						var tmp fext.Element
						tmp.Mul(&ap, v)
						uAlphaVec[col].Add(&uAlphaVec[col], &tmp)
					}
					rowEvalsVec[b] = rowEval
				}
				vk, actualY := computeVkAndActualY(alphaPow, eqRowTable, rowEvalsVec, nRowSize)

				if ctx.Packed {
					copy(packedUAlpha[k*nColSize:(k+1)*nColSize], uAlphaVec)
					copy(packedRowEvals[k*nRowSize:(k+1)*nRowSize], rowEvalsVec)
					run.AssignMultilinearExt(ctx.UCols[k].Name(), [][]fext.Element{locatorPoint(k, ctx.L, cCol)}, vk)
					run.AssignMultilinearExt(ctx.RowClaims[k].Name(), [][]fext.Element{locatorPoint(k, ctx.L, cRow)}, actualY)
				} else {
					run.AssignColumn(ctx.UAlpha[k].GetColID(), smartvectors.NewRegularExt(uAlphaVec))
					run.AssignColumn(ctx.RowEvals[k].GetColID(), smartvectors.NewRegularExt(rowEvalsVec))
					run.AssignMultilinearExt(ctx.UCols[k].Name(), [][]fext.Element{cCol}, vk)
					run.AssignMultilinearExt(ctx.RowClaims[k].Name(), [][]fext.Element{cRow}, actualY)
				}
			}
		}
	})

	if ctx.Packed {
		run.AssignColumn(ctx.UAlpha[0].GetColID(), smartvectors.NewRegularExt(packedUAlpha))
		run.AssignColumn(ctx.RowEvals[0].GetColID(), smartvectors.NewRegularExt(packedRowEvals))
	}
}

// buildSoA decomposes an fext slice into 4 base-field component vectors.
// soA[0] = B0.A0 components, soA[1] = B0.A1, soA[2] = B1.A0, soA[3] = B1.A1.
func buildSoA(eq sumcheck.MultiLin, size int) [4]field.Vector {
	var soa [4]field.Vector
	for c := range 4 {
		soa[c] = make(field.Vector, size)
	}
	for j, e := range eq[:size] {
		soa[0][j] = e.B0.A0
		soa[1][j] = e.B0.A1
		soa[2][j] = e.B1.A0
		soa[3][j] = e.B1.A1
	}
	return soa
}

// simdRowEvalBase computes rowEval = Σ_j eqSoA[*][j] × rowBase[j]
// using 4 AVX-512 inner products (one per fext component).
func simdRowEvalBase(eqSoA [4]field.Vector, rowBase field.Vector) fext.Element {
	var e fext.Element
	e.B0.A0 = eqSoA[0].InnerProduct(rowBase)
	e.B0.A1 = eqSoA[1].InnerProduct(rowBase)
	e.B1.A0 = eqSoA[2].InnerProduct(rowBase)
	e.B1.A1 = eqSoA[3].InnerProduct(rowBase)
	return e
}

// simdUAlphaAccumBase adds ap × rowBase to each component of ua using
// 4 AVX-512 ScalarMul + Add operations. tmp is a reusable scratch buffer.
func simdUAlphaAccumBase(ua *[4]field.Vector, rowBase field.Vector, ap *fext.Element, tmp field.Vector) {
	tmp.ScalarMul(rowBase, &ap.B0.A0)
	ua[0].Add(ua[0], tmp)
	tmp.ScalarMul(rowBase, &ap.B0.A1)
	ua[1].Add(ua[1], tmp)
	tmp.ScalarMul(rowBase, &ap.B1.A0)
	ua[2].Add(ua[2], tmp)
	tmp.ScalarMul(rowBase, &ap.B1.A1)
	ua[3].Add(ua[3], tmp)
}

// mergeSoA sums all partial SoA uAlpha slices and converts to AoS fext layout.
func mergeSoA(partials [][4]field.Vector, nColSize int) []fext.Element {
	acc := [4]field.Vector{
		make(field.Vector, nColSize),
		make(field.Vector, nColSize),
		make(field.Vector, nColSize),
		make(field.Vector, nColSize),
	}
	for _, p := range partials {
		for c := range 4 {
			acc[c].Add(acc[c], p[c])
		}
	}
	return soaToAoS(&acc, nColSize)
}

// soaToAoS packs 4 component vectors into an fext (AoS) slice.
func soaToAoS(soa *[4]field.Vector, nColSize int) []fext.Element {
	out := make([]fext.Element, nColSize)
	for j := 0; j < nColSize; j++ {
		out[j].B0.A0 = soa[0][j]
		out[j].B0.A1 = soa[1][j]
		out[j].B1.A0 = soa[2][j]
		out[j].B1.A1 = soa[3][j]
	}
	return out
}

// computeVkAndActualY returns v_k = Σ_b α^b·rowEvalsVec[b] and
// y_k = P_k(cRow,cCol) = Σ_b eq(cRow,b)·rowEvalsVec[b] in one pass.
func computeVkAndActualY(alphaPow []fext.Element, eqRow sumcheck.MultiLin, rowEvalsVec []fext.Element, nRowSize int) (vk, actualY fext.Element) {
	for b := 0; b < nRowSize; b++ {
		var ta fext.Element
		ta.Mul(&alphaPow[b], &rowEvalsVec[b])
		vk.Add(&vk, &ta)
		ta.Mul(&eqRow[b], &rowEvalsVec[b])
		actualY.Add(&actualY, &ta)
	}
	return
}

// locatorPoint returns the L-bit big-endian binary encoding of k prepended to
// rest, giving the locator-extended evaluation point for the k-th packed claim.
func locatorPoint(k, L int, rest []fext.Element) []fext.Element {
	pt := make([]fext.Element, L+len(rest))
	for i := 0; i < L; i++ {
		if (k>>(L-1-i))&1 == 1 {
			pt[i].SetOne()
		}
	}
	copy(pt[L:], rest)
	return pt
}
