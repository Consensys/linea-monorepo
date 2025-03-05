package mimc

import (
	"github.com/consensys/linea-monorepo/prover/crypto/mimc"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
	"github.com/sirupsen/logrus"
)

// assign assigns the columns to the prover runtime using PaddedCircularWindow
func (ctx *mimcCtx) assign(run *wizard.ProverRuntime) {
	oldStateSV := ctx.oldStates.GetColAssignment(run)
	blocksSV := ctx.blocks.GetColAssignment(run)
	n := oldStateSV.Len()

	O, L := identifyActiveWindow(oldStateSV, blocksSV, n)
	logrus.Infof("After identifying active window 0:%d, L:%d n:%d", O, L, n)

	resPad, pow4Pad := precomputePaddingValues(len(ctx.intermediateResult))
	oldStateWindow, blocksWindow := extractWindowSlices(oldStateSV, blocksSV, O, L)
	resWindow, pow4Window := computeIntermediateValues(len(ctx.intermediateResult), oldStateWindow, blocksWindow, L)

	assignOptimizedVectors(run, ctx, resWindow, pow4Window, resPad, pow4Pad, O, n)
}

// identifyActiveWindow finds the smallest window containing all non-zero values
func identifyActiveWindow(oldStateSV, blocksSV smartvectors.SmartVector, n int) (int, int) {
	firstNonZero, lastNonZero := n, -1
	//zero := field.Element{}

	// Convert to regular vectors to scan all elements
	oldState := smartvectors.IntoRegVec(oldStateSV)
	blocks := smartvectors.IntoRegVec(blocksSV)

	for i := 0; i < n; i++ {
		if !oldState[i].IsZero() || !blocks[i].IsZero() {
			firstNonZero = min(firstNonZero, i)
			lastNonZero = max(lastNonZero, i)
		}
	}

	// Debug input types and window
	switch oldStateSV.(type) {
	case *smartvectors.PaddedCircularWindow:
		logrus.Info("oldStateSV is PaddedCircularWindow")
	case *smartvectors.Regular:
		logrus.Info("oldStateSV is Regular")
	default:
		logrus.Infof("oldStateSV is %T", oldStateSV)
	}
	switch blocksSV.(type) {
	case *smartvectors.PaddedCircularWindow:
		logrus.Info("blocksSV is PaddedCircularWindow")
	case *smartvectors.Regular:
		logrus.Info("blocksSV is Regular")
	default:
		logrus.Infof("blocksSV is %T", blocksSV)
	}

	if firstNonZero <= lastNonZero {
		O := firstNonZero
		L := lastNonZero - firstNonZero + 1
		return O, L
	}
	return 0, 1 // Minimal window if all zeros
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
				logrus.Infof("Assigning Regular smart vector for res4 vector in round:%d", i)
				fullVec := make([]field.Element, n)
				copy(fullVec[O:O+L], resWindow[i])
				run.AssignColumn(
					ctx.intermediateResult[i].GetColID(),
					smartvectors.NewRegular(fullVec),
				)
			} else {
				// Partial window: use PaddedCircularWindow
				logrus.Infof("Assigning PaddedCircularWindow smart vector for res4 vector in round:%d", i)
				run.AssignColumn(
					ctx.intermediateResult[i].GetColID(),
					smartvectors.NewPaddedCircularWindow(resWindow[i], resPad[i], O, n),
				)
			}
		}

		L := len(pow4Window[i])
		if L == n {
			// Full-length window: use Regular vector
			logrus.Infof("Assigning Regular smart vector for pow4 vector in round:%d", i)
			fullVec := make([]field.Element, n)
			copy(fullVec[O:O+L], pow4Window[i])
			run.AssignColumn(
				ctx.intermediatePow4[i].GetColID(),
				smartvectors.NewRegular(fullVec),
			)
		} else {
			// Partial window: use PaddedCircularWindow
			logrus.Infof("Assigning PaddedCircularWindow smart vector for pow4 vector in round:%d", i)
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
