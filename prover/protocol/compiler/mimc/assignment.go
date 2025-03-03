package mimc

import (
	"github.com/consensys/linea-monorepo/prover/crypto/mimc"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
)

// assign assigns the columns to the prover runtime
func (ctx *mimcCtx) assign(run *wizard.ProverRuntime) {

	var (
		oldState = ctx.oldStates.GetColAssignment(run).IntoRegVecSaveAlloc()
		blocks   = ctx.blocks.GetColAssignment(run).IntoRegVecSaveAlloc()

		// Initialize slices for intermediate results and powers
		intermediatePrevRes  = make([]field.Element, len(oldState))
		intermediateCurrRes  = make([]field.Element, len(oldState))
		intermediatePrevPow4 = make([]field.Element, len(oldState))
		intermediateCurrPow4 = make([]field.Element, len(oldState))
	)

	// Set initial intermediate result as the blocks
	copy(intermediatePrevRes, blocks)

	// Compute and assign intermediate values for each round
	for i := range ctx.intermediateResult {
		computeIntermediateValues(i, oldState, intermediatePrevRes, intermediateCurrRes, intermediatePrevPow4, intermediateCurrPow4)

		// Assign computed values with independent copies
		if i == 0 {
			// Round 0: Assign intermediatePow4[0] using intermediatePrevPow4
			run.AssignColumn(
				ctx.intermediatePow4[0].GetColID(),
				smartvectors.NewRegular(append([]field.Element{}, intermediatePrevPow4...)),
			)
		} else {
			// Rounds i > 0: Assign intermediateResult[i] and intermediatePow4[i]
			run.AssignColumn(
				ctx.intermediateResult[i].GetColID(),
				smartvectors.NewRegular(append([]field.Element{}, intermediateCurrRes...)),
			)
			run.AssignColumn(
				ctx.intermediatePow4[i].GetColID(),
				smartvectors.NewRegular(append([]field.Element{}, intermediateCurrPow4...)),
			)
		}

		// Swap slices for the next round (after round 0)
		if i > 0 {
			intermediatePrevRes, intermediateCurrRes = intermediateCurrRes, intermediatePrevRes
			intermediatePrevPow4, intermediateCurrPow4 = intermediateCurrPow4, intermediatePrevPow4
		}
	}
}

// computeIntermediateValues computes intermediate values for the given round
func computeIntermediateValues(round int, oldState []field.Element, intermediatePrevRes, intermediateCurrRes, intermediatePrevPow4, intermediateCurrPow4 []field.Element) {
	parallel.Execute(len(oldState), func(start, stop int) {
		for k := start; k < stop; k++ {
			if round == 0 {
				// For the first round, compute initial intermediatePow4
				tmp := intermediatePrevRes[k]
				tmp.Add(&tmp, &mimc.Constants[0]).Add(&tmp, &oldState[k])
				intermediatePrevPow4[k].Square(&tmp).Square(&intermediatePrevPow4[k])
			} else {
				// For subsequent rounds, compute intermediate values based on previous results
				ark := mimc.Constants[round-1]
				nextArk := mimc.Constants[round]

				tmp := intermediatePrevPow4[k]
				tmp.Square(&tmp).Square(&tmp)

				// Compute intermediate result using previous result and oldState
				intermediateCurrRes[k].Add(&intermediatePrevRes[k], &ark).Add(&intermediateCurrRes[k], &oldState[k])
				intermediateCurrRes[k].Mul(&intermediateCurrRes[k], &tmp)

				// Compute intermediatePow4
				tmp = intermediateCurrRes[k]
				tmp.Add(&tmp, &nextArk).Add(&tmp, &oldState[k])
				tmp.Square(&tmp).Square(&tmp)
				intermediateCurrPow4[k] = tmp
			}
		}
	})
}
