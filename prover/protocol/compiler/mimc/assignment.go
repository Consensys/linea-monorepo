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

		// Initialize slices to hold intermediate results and intermediatePow4
		// The first entry is left empty for consistency with ctx.intermediateResult
		// We don't need to assign it because it is assigned already.
		intermediateRes  = make([][]field.Element, len(ctx.intermediateResult))
		intermediatePow4 = make([][]field.Element, len(ctx.intermediateResult))
	)

	// Initialize intermediateRes and intermediatePow4 with correct lengths

	// TODO: @srinathLN7 =>  Possible memory optimization ideas:
	// Compute inplace - Can we use a single working slice and update it in-place for each round, avoiding the need for
	// multiple large allocations? Only viable if the algo. does not require all intermediate results to be retained and
	// check downstream code (e.g., constraints or proof generation) to ensure in-place updates are compatible

	for i := range intermediateRes {
		// For each intermediate result, create a slice of field.Elements with length numRows
		intermediateRes[i] = make([]field.Element, len(oldState))
		intermediatePow4[i] = make([]field.Element, len(oldState))
	}

	// Set the initial intermediate res as the block itself
	intermediateRes[0] = blocks

	// Compute intermediate values for each round
	for i := range ctx.intermediateResult {
		computeIntermediateValues(i, oldState, intermediateRes, intermediatePow4)

	}

	// Assign columns
	for i := range ctx.intermediateResult {
		// Assign computed values to the runtime
		if i > 0 {
			// Skip the first intermediate result
			// Recall that the first intermediate res is the block itself
			run.AssignColumn(
				ctx.intermediateResult[i].GetColID(),
				smartvectors.NewRegular(intermediateRes[i]),
			)
		}

		// Assign intermediatePow4 to the runtime
		run.AssignColumn(
			ctx.intermediatePow4[i].GetColID(),
			smartvectors.NewRegular(intermediatePow4[i]),
		)
	}

}

// computeIntermediateValues computes intermediate values for the given round
func computeIntermediateValues(round int, oldState []field.Element, intermediateRes, intermediatePow4 [][]field.Element) {
	parallel.Execute(len(oldState), func(start, stop int) {
		for k := start; k < stop; k++ {
			if round == 0 {
				// For the first round, compute initial intermediatePow4
				tmp := intermediateRes[0][k]
				tmp.Add(&tmp, &mimc.Constants[0]).Add(&tmp, &oldState[k])
				intermediatePow4[0][k].Square(&tmp).Square(&intermediatePow4[0][k])
			} else {
				// For subsequent rounds, compute intermediate values based on previous results
				ark := mimc.Constants[round-1]
				nextArk := mimc.Constants[round]

				tmp := intermediatePow4[round-1][k]
				tmp.Square(&tmp).Square(&tmp)

				// Compute intermediate result using previous result and oldState
				intermediateRes[round][k] = intermediateRes[round-1][k]
				intermediateRes[round][k].Add(&intermediateRes[round][k], &ark).Add(&intermediateRes[round][k], &oldState[k])
				intermediateRes[round][k].Mul(&intermediateRes[round][k], &tmp)

				// Compute intermediatePow4
				tmp = intermediateRes[round][k]
				tmp.Add(&tmp, &nextArk).Add(&tmp, &oldState[k])
				tmp.Square(&tmp).Square(&tmp)
				intermediatePow4[round][k] = tmp
			}
		}
	})

}
