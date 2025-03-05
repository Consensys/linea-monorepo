package mimc

import (
	"github.com/consensys/linea-monorepo/prover/crypto/mimc"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
)

// assign assigns the columns to the prover runtime using PaddedCircularWindow
func (ctx *mimcCtx) assign(run *wizard.ProverRuntime) {
	oldStateSV := ctx.oldStates.GetColAssignment(run)
	blocksSV := ctx.blocks.GetColAssignment(run)
	n := oldStateSV.Len()

	O, L := identifyActiveWindow(oldStateSV, blocksSV, n)
	resPad, pow4Pad := precomputePaddingValues(len(ctx.intermediateResult))
	oldStateWindow, blocksWindow := extractWindowSlices(oldStateSV, blocksSV, O, L)
	resWindow, pow4Window := computeIntermediateValues(len(ctx.intermediateResult), oldStateWindow, blocksWindow, L)

	assignOptimizedVectors(run, ctx, resWindow, pow4Window, resPad, pow4Pad, O, n)
}

// identifyActiveWindow identifies the active window (non-zero regions)
func identifyActiveWindow(oldStateSV, blocksSV smartvectors.SmartVector, n int) (int, int) {
	minO, maxEnd := n, 0

	// Check oldStateSV
	if pcw, ok := oldStateSV.(*smartvectors.PaddedCircularWindow); ok {
		paddingVal := pcw.PaddingVal()
		if (&paddingVal).IsZero() {
			window := smartvectors.Window(oldStateSV)
			// Assuming offset is available or inferred; adjust if you have an Offset() method
			o := n - len(window) // Default to left-padded assumption
			end := o + len(window)
			minO = min(minO, o)
			maxEnd = max(maxEnd, end)
		}
	}

	// Check blocksSV
	if pcw, ok := blocksSV.(*smartvectors.PaddedCircularWindow); ok {
		paddingVal := pcw.PaddingVal()
		if (&paddingVal).IsZero() {
			window := smartvectors.Window(blocksSV)
			// Assuming offset is 0 for right-padded; adjust if you have an Offset() method
			o := 0
			end := len(window)
			minO = min(minO, o)
			maxEnd = max(maxEnd, end)
		}
	}

	if maxEnd > minO {
		O := minO
		L := maxEnd - minO
		return O, L
	}
	// Default to full vector if no valid window is found or inputs aren’t optimized
	return 0, n
}

// precomputePaddingValues precomputes padding values for constant regions
func precomputePaddingValues(numRounds int) ([]field.Element, []field.Element) {
	resPad := make([]field.Element, numRounds)
	pow4Pad := make([]field.Element, numRounds)
	resPad[0].SetZero()
	var tmp field.Element
	tmp.Add(&resPad[0], &mimc.Constants[0])
	pow4Pad[0].Square(&tmp).Square(&pow4Pad[0])

	for r := 1; r < numRounds; r++ {
		tmp.Square(&pow4Pad[r-1]).Square(&tmp)
		resPad[r].Add(&resPad[r-1], &mimc.Constants[r-1])
		resPad[r].Mul(&resPad[r], &tmp)
		tmp.Add(&resPad[r], &mimc.Constants[r])
		pow4Pad[r].Square(&tmp).Square(&pow4Pad[r])
	}

	return resPad, pow4Pad
}

// extractWindowSlices extracts window slices from the smart vectors
func extractWindowSlices(oldStateSV, blocksSV smartvectors.SmartVector, O, L int) ([]field.Element, []field.Element) {
	oldStateWindow := smartvectors.IntoRegVec(oldStateSV)[O : O+L]
	blocksWindow := smartvectors.IntoRegVec(blocksSV)[O : O+L]
	return oldStateWindow, blocksWindow
}

// computeIntermediateValues computes intermediate values for the window
func computeIntermediateValues(numRounds int, oldStateWindow, blocksWindow []field.Element, L int) ([][]field.Element, [][]field.Element) {
	resWindow := make([][]field.Element, numRounds)
	pow4Window := make([][]field.Element, numRounds)
	for i := range resWindow {
		resWindow[i] = make([]field.Element, L)
		pow4Window[i] = make([]field.Element, L)
	}
	copy(resWindow[0], blocksWindow)

	for r := 0; r < numRounds; r++ {
		parallel.Execute(L, func(start, stop int) {
			for k := start; k < stop; k++ {
				if r == 0 {
					tmp := resWindow[0][k]
					tmp.Add(&tmp, &mimc.Constants[0]).Add(&tmp, &oldStateWindow[k])
					pow4Window[0][k].Square(&tmp).Square(&pow4Window[0][k])
				} else {
					ark := mimc.Constants[r-1]
					nextArk := mimc.Constants[r]
					tmp := pow4Window[r-1][k]
					tmp.Square(&tmp).Square(&tmp)
					resWindow[r][k] = resWindow[r-1][k]
					resWindow[r][k].Add(&resWindow[r][k], &ark).Add(&resWindow[r][k], &oldStateWindow[k])
					resWindow[r][k].Mul(&resWindow[r][k], &tmp)
					tmp = resWindow[r][k]
					tmp.Add(&tmp, &nextArk).Add(&tmp, &oldStateWindow[k])
					pow4Window[r][k].Square(&tmp).Square(&pow4Window[r][k])
				}
			}
		})
	}

	return resWindow, pow4Window
}

// assignOptimizedVectors assigns optimized vectors to the runtime
func assignOptimizedVectors(run *wizard.ProverRuntime, ctx *mimcCtx, resWindow, pow4Window [][]field.Element, resPad, pow4Pad []field.Element, O, n int) {
	for i := range ctx.intermediateResult {
		if i > 0 {
			L := len(resWindow[i])
			if L == n {
				// Full-length window: use Regular vector
				fullVec := make([]field.Element, n)
				copy(fullVec[O:O+L], resWindow[i])
				run.AssignColumn(
					ctx.intermediateResult[i].GetColID(),
					smartvectors.NewRegular(fullVec),
				)
			} else {
				// Partial window: use PaddedCircularWindow
				run.AssignColumn(
					ctx.intermediateResult[i].GetColID(),
					smartvectors.NewPaddedCircularWindow(resWindow[i], resPad[i], O, n),
				)
			}
		}

		L := len(pow4Window[i])
		if L == n {
			// Full-length window: use Regular vector
			fullVec := make([]field.Element, n)
			copy(fullVec[O:O+L], pow4Window[i])
			run.AssignColumn(
				ctx.intermediatePow4[i].GetColID(),
				smartvectors.NewRegular(fullVec),
			)
		} else {
			// Partial window: use PaddedCircularWindow
			run.AssignColumn(
				ctx.intermediatePow4[i].GetColID(),
				smartvectors.NewPaddedCircularWindow(pow4Window[i], pow4Pad[i], O, n),
			)
		}
	}
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
