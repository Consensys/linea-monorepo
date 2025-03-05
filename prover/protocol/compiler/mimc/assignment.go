package mimc

import (
	"sync"

	"github.com/consensys/linea-monorepo/prover/crypto/mimc"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
	"github.com/sirupsen/logrus"
)

// assign assigns MiMC columns to the prover runtime, optimizing memory with PaddedCircularWindow
func (ctx *mimcCtx) assign(run *wizard.ProverRuntime) {
	var (
		oldState = ctx.oldStates.GetColAssignment(run).IntoRegVecSaveAlloc()
		blocks   = ctx.blocks.GetColAssignment(run).IntoRegVecSaveAlloc()
		n        = len(ctx.intermediateResult)
	)

	// Precompute padding values for all rounds starting from zeroes
	resPadding, pow4Padding := computePaddingValues(n, mimc.Constants)

	// Count zero pairs (block=0, initState=0) for initial sparsity estimate
	countZero := countZeroPairs(oldState, blocks)
	logrus.Infof("Zero pairs: %d out of %d (~%.2f%% sparsity)", countZero, len(oldState), float64(countZero)*100/float64(len(oldState)))

	// Precompute intermediatePow4[0] with PaddedCircularWindow for Round 0
	prevPow4, nonZeroWindowPow4 := precomputePow4Round0(oldState, blocks, countZero)
	pow4Round0 := createSmartVector(nonZeroWindowPow4, pow4Padding[0], 0, len(oldState))
	logrus.Infof("Round 0 pow4 non-zeros: %d, index 0: %v", len(nonZeroWindowPow4), pow4Round0.Get(0))

	// Round 0 => Special case
	run.AssignColumn(ctx.intermediatePow4[0].GetColID(), pow4Round0)
	prevRes := make([]field.Element, len(oldState))
	copy(prevRes, blocks)

	// Temporary slices for computation
	currRes := make([]field.Element, len(oldState))
	currPow4 := make([]field.Element, len(oldState))

	// Compute and assign for rounds > 0
	for i := 1; i < n; i++ {
		computeIntermediateValues(i, oldState, prevRes, currRes, prevPow4, currPow4)

		// Convert currRes to SmartVector with precomputed padding value
		resVector := createSmartVectorFromResults(currRes, countZero, len(oldState), i, resPadding[i])
		// logrus.Infof("Round %d res non-zeros: %d, index 0: %v", i, countNonZeros(currRes), resVector.Get(0))
		run.AssignColumn(ctx.intermediateResult[i].GetColID(), resVector)

		// Convert currPow4 to SmartVector with precomputed padding value
		pow4Vector := createSmartVectorFromResults(currPow4, countZero, len(oldState), i, pow4Padding[i])
		// logrus.Infof("Round %d pow4 non-zeros: %d, index 0: %v", i, countNonZeros(currPow4), pow4Vector.Get(0))
		run.AssignColumn(ctx.intermediatePow4[i].GetColID(), pow4Vector)

		// Swap for next round
		prevRes, currRes = currRes, prevRes
		prevPow4, currPow4 = currPow4, prevPow4
	}
}

// countZeroPairs counts the number of zero pairs in oldState and blocks
func countZeroPairs(oldState, blocks []field.Element) int {
	var countZero int
	for i := range oldState {
		if oldState[i].IsZero() && blocks[i].IsZero() {
			countZero++
		}
	}
	return countZero
}

// precomputePow4Round0 precomputes the intermediatePow4[0] values
func precomputePow4Round0(oldState, blocks []field.Element, countZero int) ([]field.Element, []field.Element) {
	pow4Round0Full := make([]field.Element, len(oldState))
	nonZeroWindowPow4 := make([]field.Element, 0, len(oldState)-countZero)
	var mu sync.Mutex
	parallel.Execute(len(oldState), func(start, stop int) {
		for k := start; k < stop; k++ {
			var tmp field.Element
			tmp.Add(&blocks[k], &mimc.Constants[0]).Add(&tmp, &oldState[k])
			tmp.Square(&tmp).Square(&tmp)
			pow4Round0Full[k] = tmp
			if !blocks[k].IsZero() || !oldState[k].IsZero() {
				mu.Lock()
				nonZeroWindowPow4 = append(nonZeroWindowPow4, tmp)
				mu.Unlock()
			}
		}
	})
	return pow4Round0Full, nonZeroWindowPow4
}

// computePaddingValues precomputes padding values for all rounds starting from zeroes
func computePaddingValues(n int, constants []field.Element) ([]field.Element, []field.Element) {
	resPadding := make([]field.Element, n)
	pow4Padding := make([]field.Element, n)
	var zero field.Element // Zero in the field

	// Round 0: Starting with blocks = 0
	pow4Padding[0].Add(&zero, &constants[0]).Square(&pow4Padding[0]).Square(&pow4Padding[0])
	// resPadding[0] remains zero, as intermediateRes[0] = blocks = 0

	for i := 1; i < n; i++ {
		ark := constants[i-1]
		nextArk := constants[i]

		// Compute resPadding[i] = (resPadding[i-1] + ark + 0) * pow4Padding[i-1]^4
		var tmp field.Element
		resPadding[i].Add(&resPadding[i-1], &ark) // +0 is implicit
		tmp.Square(&pow4Padding[i-1]).Square(&tmp)
		resPadding[i].Mul(&resPadding[i], &tmp)

		// Compute pow4Padding[i] = (resPadding[i] + nextArk + 0)^4
		pow4Padding[i].Add(&resPadding[i], &nextArk)
		pow4Padding[i].Square(&pow4Padding[i]).Square(&pow4Padding[i])

		// logrus.Infof("Round %d padding res: %v, pow4: %v", i, resPadding[i], pow4Padding[i])
		LogRoundZero := true
		if LogRoundZero {
			// logrus.Infof("Round 0 padding res: %v, pow4: %v", resPadding[0], pow4Padding[0])
			LogRoundZero = false
		}
	}
	return resPadding, pow4Padding
}

// createSmartVector creates a SmartVector from a slice of field.Element
func createSmartVector(elements []field.Element, paddingVal field.Element, offset, totalLen int) smartvectors.SmartVector {
	if len(elements) == totalLen {
		return smartvectors.NewRegular(elements)
	}
	logrus.Infof("Creating NewPaddedCircularWindow with offset:%d, paddingVal:%v", offset, paddingVal)
	return smartvectors.NewPaddedCircularWindow(elements, paddingVal, offset, totalLen)
}

// createSmartVectorFromResults creates a SmartVector from computation results
func createSmartVectorFromResults(results []field.Element, countZero, totalLen, round int, paddingVal field.Element) smartvectors.SmartVector {
	expectedNonZeros := totalLen - countZero
	if round > 1 {
		expectedNonZeros = totalLen / 2 // Adjust based on observed sparsity decrease
	}
	nonZeroWindow := make([]field.Element, 0, expectedNonZeros)
	var offset int
	var initialOffset bool
	for i, res := range results {
		if !res.IsZero() {
			if !initialOffset {
				offset, initialOffset = i, true
			}
			nonZeroWindow = append(nonZeroWindow, res)
		}
	}
	return createSmartVector(nonZeroWindow, paddingVal, offset, totalLen)
}

// computeIntermediateValues computes intermediate values for rounds > 0
func computeIntermediateValues(round int, oldState, prevRes, currRes, prevPow4, currPow4 []field.Element) {
	parallel.Execute(len(oldState), func(start, stop int) {
		for k := start; k < stop; k++ {
			// For subsequent rounds (>0), compute intermediate values based on previous results
			ark := mimc.Constants[round-1]
			nextArk := mimc.Constants[round]

			tmp := prevPow4[k]
			tmp.Square(&tmp).Square(&tmp)

			// Compute intermediate result using previous result and oldState
			currRes[k].Add(&prevRes[k], &ark).Add(&currRes[k], &oldState[k])
			currRes[k].Mul(&currRes[k], &tmp)

			// Compute intermediatePow4
			tmp = currRes[k]
			tmp.Add(&tmp, &nextArk).Add(&tmp, &oldState[k])
			tmp.Square(&tmp).Square(&tmp)
			currPow4[k] = tmp
		}
	})
}
