package multilinvortex

import (
	"runtime"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
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

	// Detect a shared-input batch: all K polys reference the same column AND
	// all K per-pol points share the same cCol suffix. This pattern is emitted
	// by InsertBootstrapperOpeningsPacked, which inserts K duplicate Q polys
	// each at (b_k ‖ master[L_k:]) — since the locator length L_k is always
	// ≤ nv_k ≤ nRow in the prefix-exclusive scheme, cCol = master[nRow:] is
	// shared. When the pattern matches we run ONE matrix pass parallelised
	// across all cores and fan out cheap per-k assignments.
	if K > 1 {
		if shared, sharedV, sharedY := p.tryRunSharedInput(run, inputParams, alphaPow, cCol, nRow, nCol, nRowSize, nColSize); shared {
			_, _ = sharedV, sharedY
			return
		}
	}

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

			// Padded fast path (base-field): only iterate rows that touch
			// the non-zero window. Rows entirely in the zero suffix contribute
			// 0 to both rowEvalsVec and uAlpha, so they can be skipped.
			if window, ok := rightZeroPaddedBase(sv); ok {
				activeRows := (len(window) + nColSize - 1) / nColSize
				activeBase := materializeActiveBase(window, activeRows, nColSize)
				parallel.Execute(nthreads, func(start, stop int) {
					tmpVec := make(field.Vector, nColSize)
					for t := start; t < stop; t++ {
						bStart := t * rowsPerThread
						bStop := bStart + rowsPerThread
						if bStop > activeRows {
							bStop = activeRows
						}
						if bStart >= bStop {
							continue
						}
						ua := &partialUASoA[t]
						for b := bStart; b < bStop; b++ {
							rowBase := field.Vector(activeBase[b*nColSize : (b+1)*nColSize])
							rowEvalsVec[b] = simdRowEvalBase(eqColSoA, rowBase)
							simdUAlphaAccumBase(ua, rowBase, &alphaPow[b], tmpVec)
						}
					}
				})
				uAlphaVec := mergeSoA(partialUASoA, nColSize)
				vk, actualY := computeVkAndActualY(alphaPow, eqRowTable, rowEvalsVec, activeRows)
				p.assignProverResults(run, k, uAlphaVec, rowEvalsVec, activeRows, nRowSize, nColSize, vk, actualY, cCol, cRow, packedUAlpha, packedRowEvals)
				continue
			}
			// Padded fast path (fext): same logic for already-fext columns
			// (e.g. RowEvals from a prior round whose first round was padded).
			if windowExt, ok := rightZeroPaddedExt(sv); ok {
				activeRows := (len(windowExt) + nColSize - 1) / nColSize
				activeExt := materializeActiveExt(windowExt, activeRows, nColSize)
				partialUA := make([][]fext.Element, nthreads)
				for t := range partialUA {
					partialUA[t] = make([]fext.Element, nColSize)
				}
				parallel.Execute(nthreads, func(start, stop int) {
					for t := start; t < stop; t++ {
						bStart := t * rowsPerThread
						bStop := bStart + rowsPerThread
						if bStop > activeRows {
							bStop = activeRows
						}
						if bStart >= bStop {
							continue
						}
						loc := partialUA[t]
						for b := bStart; b < bStop; b++ {
							rowExt := activeExt[b*nColSize : (b+1)*nColSize]
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
				vk, actualY := computeVkAndActualY(alphaPow, eqRowTable, rowEvalsVec, activeRows)
				p.assignProverResults(run, k, uAlphaVec, rowEvalsVec, activeRows, nRowSize, nColSize, vk, actualY, cCol, cRow, packedUAlpha, packedRowEvals)
				continue
			}

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
				p.assignProverResults(run, k, uAlphaVec, rowEvalsVec, nRowSize, nRowSize, nColSize, vk, actualY, cCol, cRow, packedUAlpha, packedRowEvals)
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
				p.assignProverResults(run, k, uAlphaVec, rowEvalsVec, nRowSize, nRowSize, nColSize, vk, actualY, cCol, cRow, packedUAlpha, packedRowEvals)
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

			// Padded fast path (base-field): only iterate active rows.
			if window, ok := rightZeroPaddedBase(sv); ok {
				activeRows := (len(window) + nColSize - 1) / nColSize
				activeBase := materializeActiveBase(window, activeRows, nColSize)
				uAlphaSoA := [4]field.Vector{
					make(field.Vector, nColSize),
					make(field.Vector, nColSize),
					make(field.Vector, nColSize),
					make(field.Vector, nColSize),
				}
				for b := 0; b < activeRows; b++ {
					rowBase := field.Vector(activeBase[b*nColSize : (b+1)*nColSize])
					rowEvalsVec[b] = simdRowEvalBase(eqColSoA, rowBase)
					simdUAlphaAccumBase(&uAlphaSoA, rowBase, &alphaPow[b], tmpVec)
				}
				uAlphaVec := soaToAoS(&uAlphaSoA, nColSize)
				vk, actualY := computeVkAndActualY(alphaPow, eqRowTable, rowEvalsVec, activeRows)
				p.assignProverResults(run, k, uAlphaVec, rowEvalsVec, activeRows, nRowSize, nColSize, vk, actualY, cCol, cRow, packedUAlpha, packedRowEvals)
				continue
			}
			// Padded fast path (fext).
			if windowExt, ok := rightZeroPaddedExt(sv); ok {
				activeRows := (len(windowExt) + nColSize - 1) / nColSize
				activeExt := materializeActiveExt(windowExt, activeRows, nColSize)
				uAlphaVec := make([]fext.Element, nColSize)
				for b := 0; b < activeRows; b++ {
					rowExt := activeExt[b*nColSize : (b+1)*nColSize]
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
				vk, actualY := computeVkAndActualY(alphaPow, eqRowTable, rowEvalsVec, activeRows)
				p.assignProverResults(run, k, uAlphaVec, rowEvalsVec, activeRows, nRowSize, nColSize, vk, actualY, cCol, cRow, packedUAlpha, packedRowEvals)
				continue
			}

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
				p.assignProverResults(run, k, uAlphaVec, rowEvalsVec, nRowSize, nRowSize, nColSize, vk, actualY, cCol, cRow, packedUAlpha, packedRowEvals)
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
				p.assignProverResults(run, k, uAlphaVec, rowEvalsVec, nRowSize, nRowSize, nColSize, vk, actualY, cCol, cRow, packedUAlpha, packedRowEvals)
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

// rightZeroPaddedBase returns (window, true) iff sv is a base-field
// PaddedCircularWindow with offset 0 and a zero padding value — i.e. the
// non-zero data lies entirely in [0, len(window)) and everything from
// len(window) up to sv.Len() is zero.
func rightZeroPaddedBase(sv smartvectors.SmartVector) ([]field.Element, bool) {
	w, ok := sv.(*smartvectors.PaddedCircularWindow)
	if !ok || w.Offset_ != 0 || !w.PaddingVal_.IsZero() {
		return nil, false
	}
	return w.Window_, true
}

// rightZeroPaddedExt is the fext analogue of rightZeroPaddedBase.
func rightZeroPaddedExt(sv smartvectors.SmartVector) ([]fext.Element, bool) {
	w, ok := sv.(*smartvectors.PaddedCircularWindowExt)
	if !ok || w.Offset != 0 || !w.PaddingVal_.IsZero() {
		return nil, false
	}
	return w.Window_, true
}

// materializeActiveBase returns a contiguous base-field slice of length
// activeRows*nColSize whose first len(window) entries are window and the rest
// are zero. When window already aligns with activeRows*nColSize the slice is
// aliased to window directly (no copy).
func materializeActiveBase(window []field.Element, activeRows, nColSize int) []field.Element {
	if len(window) == activeRows*nColSize {
		return window
	}
	out := make([]field.Element, activeRows*nColSize)
	copy(out, window)
	return out
}

// materializeActiveExt is the fext analogue of materializeActiveBase.
func materializeActiveExt(window []fext.Element, activeRows, nColSize int) []fext.Element {
	if len(window) == activeRows*nColSize {
		return window
	}
	out := make([]fext.Element, activeRows*nColSize)
	copy(out, window)
	return out
}

// assignProverResults wraps the AssignColumn / AssignMultilinearExt calls that
// emit one column's worth of UAlpha / RowEvals / UCols / RowClaims results.
// When activeRows < nRowSize the RowEvals output is wrapped in a
// PaddedCircularWindowExt so the next round's ProverAction can detect the
// sparsity and skip the zero suffix again — propagating the O(actual) saving
// across recursion levels.
func (p *ProverAction) assignProverResults(
	run *wizard.ProverRuntime,
	k int,
	uAlphaVec, rowEvalsVec []fext.Element,
	activeRows, nRowSize, nColSize int,
	vk, actualY fext.Element,
	cCol, cRow []fext.Element,
	packedUAlpha, packedRowEvals []fext.Element,
) {
	ctx := p.Ctx
	if ctx.Packed {
		copy(packedUAlpha[k*nColSize:(k+1)*nColSize], uAlphaVec)
		copy(packedRowEvals[k*nRowSize:(k+1)*nRowSize], rowEvalsVec)
		run.AssignMultilinearExt(ctx.UCols[k].Name(), [][]fext.Element{locatorPoint(k, ctx.L, cCol)}, vk)
		run.AssignMultilinearExt(ctx.RowClaims[k].Name(), [][]fext.Element{locatorPoint(k, ctx.L, cRow)}, actualY)
		return
	}
	run.AssignColumn(ctx.UAlpha[k].GetColID(), smartvectors.NewRegularExt(uAlphaVec))
	if activeRows > 0 && activeRows < nRowSize {
		var zero fext.Element
		run.AssignColumn(ctx.RowEvals[k].GetColID(),
			smartvectors.NewPaddedCircularWindowExt(rowEvalsVec[:activeRows], zero, 0, nRowSize))
	} else {
		run.AssignColumn(ctx.RowEvals[k].GetColID(), smartvectors.NewRegularExt(rowEvalsVec))
	}
	run.AssignMultilinearExt(ctx.UCols[k].Name(), [][]fext.Element{cCol}, vk)
	run.AssignMultilinearExt(ctx.RowClaims[k].Name(), [][]fext.Element{cRow}, actualY)
}

// tryRunSharedInput implements the shared-input fast path: when every input
// poly is the SAME column and every per-pol cCol slice is equal, UAlpha and
// RowEvals depend only on (Q, α) and (Q, cCol) respectively, so they are the
// same vector for every k. We compute each ONCE — using a single full-machine
// parallel matrix pass over Q — and fan out cheap O(nRowSize) per-k work for
// the K downstream y_k and the (shared) v_k. Returns (false, _, _) if the
// pattern doesn't match; the caller falls back to the general K-loop.
//
// Outputs (when matched):
//   - K UAlpha proof columns assigned identical content (uAlphaVec_shared)
//   - K RowEvals proof columns assigned identical content (rowEvalsVec_shared)
//   - K UCols query params at (cCol_shared, vk_shared)
//   - K RowClaims query params at (cRow_k, y_k)
func (p *ProverAction) tryRunSharedInput(
	run *wizard.ProverRuntime,
	inputParams query.MultilinearEvalParams,
	alphaPow []fext.Element,
	cCol []fext.Element,
	nRow, nCol, nRowSize, nColSize int,
) (bool, fext.Element, fext.Element) {
	ctx := p.Ctx
	K := len(ctx.InputQuery.Pols)
	if K < 2 || ctx.Packed {
		return false, fext.Element{}, fext.Element{}
	}

	// Require uniform input column. SharedInput contexts always satisfy
	// this; legacy contexts may benefit when callers happen to point K
	// queries at the same column.
	col0ID := ctx.InputQuery.Pols[0].GetColID()
	for k := 1; k < K; k++ {
		if ctx.InputQuery.Pols[k].GetColID() != col0ID {
			return false, fext.Element{}, fext.Element{}
		}
	}

	// Group polys by cCol. For each unique cCol_g we'll do one matrix pass.
	// UAlpha is independent of cCol and so is computed exactly once.
	cColEqual := func(a, b []fext.Element) bool {
		for j := 0; j < nCol; j++ {
			if !a[j].Equal(&b[j]) {
				return false
			}
		}
		return true
	}
	uniqueIdx := make([]int, K)             // cCol-group of poly k
	uniqueCCol := [][]fext.Element{cCol}     // representative cCol per group
	for k := 1; k < K; k++ {
		pt := inputParams.Points[k]
		if len(pt) != nRow+nCol {
			return false, fext.Element{}, fext.Element{}
		}
		cColK := pt[nRow:]
		found := -1
		for g, c := range uniqueCCol {
			if cColEqual(c, cColK) {
				found = g
				break
			}
		}
		if found < 0 {
			found = len(uniqueCCol)
			uniqueCCol = append(uniqueCCol, cColK)
		}
		uniqueIdx[k] = found
	}
	// Legacy (non-SharedInput) contexts have K UAlpha/RowEvals columns;
	// the fast path only helps when cCol is uniform — otherwise its K-pass
	// would emit duplicate UAlpha to K cols, but we have no compile-time
	// guarantee about K_RowEvals layout. Bail out unless we can claim the
	// full optimisation OR the context already declares SharedInput.
	if !ctx.SharedInput && len(uniqueCCol) > 1 {
		return false, fext.Element{}, fext.Element{}
	}

	nthreads := runtime.GOMAXPROCS(0)
	if nthreads > nRowSize {
		nthreads = nRowSize
	}
	rowsPerThread := (nRowSize + nthreads - 1) / nthreads

	sv := run.GetColumn(col0ID)
	rowEvalsByGroup := make([][]fext.Element, len(uniqueCCol))
	for g := range rowEvalsByGroup {
		rowEvalsByGroup[g] = make([]fext.Element, nRowSize)
	}

	colBase, isBase := sv.IntoRegVecSaveAllocBase()
	var uAlphaVec []fext.Element

	// Per-group eqColSoA tables. The first group also drives UAlpha
	// accumulation (UAlpha is independent of cCol — compute it once).
	eqSoAs := make([][4]field.Vector, len(uniqueCCol))
	for g, c := range uniqueCCol {
		eqSoAs[g] = buildSoA(sumcheck.BuildEqTable(c), nColSize)
	}

	partialUASoA := make([][4]field.Vector, nthreads)
	for t := range partialUASoA {
		for c := range 4 {
			partialUASoA[t][c] = make(field.Vector, nColSize)
		}
	}

	if isBase == nil {
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
					for g := range uniqueCCol {
						rowEvalsByGroup[g][b] = simdRowEvalBase(eqSoAs[g], rowBase)
					}
					simdUAlphaAccumBase(ua, rowBase, &alphaPow[b], tmpVec)
				}
			}
		})
		uAlphaVec = mergeSoA(partialUASoA, nColSize)
	} else {
		// Fext fallback (rare for the bootstrapper).
		colDataExt := sv.IntoRegVecSaveAllocExt()
		partialUA := make([][]fext.Element, nthreads)
		for t := range partialUA {
			partialUA[t] = make([]fext.Element, nColSize)
		}
		// Per-group raw eq tables (we need the fext eq vector here, not SoA).
		eqColTables := make([]sumcheck.MultiLin, len(uniqueCCol))
		for g, c := range uniqueCCol {
			eqColTables[g] = sumcheck.BuildEqTable(c)
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
					rowEvalsG := make([]fext.Element, len(uniqueCCol))
					for col := 0; col < nColSize; col++ {
						v := &rowExt[col]
						for g := range uniqueCCol {
							var r fext.Element
							r.Mul(&eqColTables[g][col], v)
							rowEvalsG[g].Add(&rowEvalsG[g], &r)
						}
						var tmp fext.Element
						tmp.Mul(&ap, v)
						loc[col].Add(&loc[col], &tmp)
					}
					for g := range uniqueCCol {
						rowEvalsByGroup[g][b] = rowEvalsG[g]
					}
				}
			}
		})
		uAlphaVec = make([]fext.Element, nColSize)
		for t := 0; t < nthreads; t++ {
			for col := 0; col < nColSize; col++ {
				uAlphaVec[col].Add(&uAlphaVec[col], &partialUA[t][col])
			}
		}
	}

	uAlphaSV := smartvectors.NewRegularExt(uAlphaVec)

	// Assign UAlpha: ONE col when SharedInput, K identical cols when legacy.
	if ctx.SharedInput {
		run.AssignColumn(ctx.UAlpha[0].GetColID(), uAlphaSV)
	} else {
		for k := 0; k < K; k++ {
			run.AssignColumn(ctx.UAlpha[k].GetColID(), uAlphaSV)
		}
	}

	// Pre-build smartvectors for each unique RowEvals group.
	rowEvalsSVs := make([]smartvectors.SmartVector, len(rowEvalsByGroup))
	for g, vec := range rowEvalsByGroup {
		rowEvalsSVs[g] = smartvectors.NewRegularExt(vec)
	}

	// SharedRowEvals: single RowEvals column, single UCols claim, single
	// RowClaims query holding K (point, y) pairs. Possible only when
	// len(uniqueCCol) == 1 (which the build path guarantees by contract).
	if ctx.SharedRowEvals {
		if len(uniqueCCol) != 1 {
			panic("multilinvortex: SharedRowEvals context received non-uniform cCol at runtime")
		}
		run.AssignColumn(ctx.RowEvals[0].GetColID(), rowEvalsSVs[0])

		// vk_shared = Σ_b α^b · RowEvals[b] — identical for every k.
		var vkShared fext.Element
		for b := 0; b < nRowSize; b++ {
			var t fext.Element
			t.Mul(&alphaPow[b], &rowEvalsByGroup[0][b])
			vkShared.Add(&vkShared, &t)
		}
		run.AssignMultilinearExt(ctx.UCols[0].Name(),
			[][]fext.Element{uniqueCCol[0]}, vkShared)

		// Single RowClaims query with K (point, y) pairs.
		rowPoints := make([][]fext.Element, K)
		rowYs := make([]fext.Element, K)
		for k := 0; k < K; k++ {
			cRowK := inputParams.Points[k][:nRow]
			eqRowK := sumcheck.BuildEqTable(cRowK)
			var yk fext.Element
			for b := 0; b < nRowSize; b++ {
				var t fext.Element
				t.Mul(&eqRowK[b], &rowEvalsByGroup[0][b])
				yk.Add(&yk, &t)
			}
			rowPoints[k] = cRowK
			rowYs[k] = yk
		}
		run.AssignMultilinearExt(ctx.RowClaims[0].Name(), rowPoints, rowYs...)
		return true, fext.Element{}, fext.Element{}
	}

	// SharedInput (no RowEvals collapse) / legacy fast-path fan-out: K
	// UCols + K RowClaims, each assigned per-k.
	for k := 0; k < K; k++ {
		g := uniqueIdx[k]
		run.AssignColumn(ctx.RowEvals[k].GetColID(), rowEvalsSVs[g])

		// v_k = Σ_b α^b · RowEvals_k[b]. Same across k whenever cCol_k
		// shares a group; recomputed here for clarity (cheap O(nRowSize)).
		var vk fext.Element
		for b := 0; b < nRowSize; b++ {
			var t fext.Element
			t.Mul(&alphaPow[b], &rowEvalsByGroup[g][b])
			vk.Add(&vk, &t)
		}

		cRowK := inputParams.Points[k][:nRow]
		eqRowK := sumcheck.BuildEqTable(cRowK)
		var yk fext.Element
		for b := 0; b < nRowSize; b++ {
			var t fext.Element
			t.Mul(&eqRowK[b], &rowEvalsByGroup[g][b])
			yk.Add(&yk, &t)
		}

		run.AssignMultilinearExt(ctx.UCols[k].Name(),
			[][]fext.Element{uniqueCCol[g]}, vk)
		run.AssignMultilinearExt(ctx.RowClaims[k].Name(),
			[][]fext.Element{cRowK}, yk)
	}

	return true, fext.Element{}, fext.Element{}
}
