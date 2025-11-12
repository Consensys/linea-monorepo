package ecpair

import (
	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
)

// Assign assigns the data to the circuit
func (ec *ECPair) Assign(run *wizard.ProverRuntime) {

	// assign data to the pairing check part
	ec.assignPairingData(run)
	// assign data to the membership check part
	ec.assignMembershipData(run)
	// assign the column telling wether the previous and the current row have
	// the same id.
	ec.CptPrevEqualCurrID.Run(run)

	// general assignments
	var (
		srcIsG2Pulling       = ec.UnalignedG2MembershipData.IsPulling.GetColAssignment(run).IntoRegVecSaveAlloc()
		srcIsG2Computed      = ec.UnalignedG2MembershipData.IsComputed.GetColAssignment(run).IntoRegVecSaveAlloc()
		srcIsPairingPulling  = ec.UnalignedPairingData.IsPulling.GetColAssignment(run).IntoRegVecSaveAlloc()
		srcIsPairingComputed = ec.UnalignedPairingData.IsComputed.GetColAssignment(run).IntoRegVecSaveAlloc()
	)

	if len(srcIsG2Pulling) != len(srcIsG2Computed) || len(srcIsG2Pulling) != len(srcIsPairingPulling) || len(srcIsG2Pulling) != len(srcIsPairingComputed) {
		utils.Panic("ECPair: input length mismatch")
	}

	dstIsActive := common.NewVectorBuilder(ec.IsActive)
	for i := 0; i < len(srcIsG2Pulling); i++ {
		if srcIsG2Pulling[i].IsOne() || srcIsG2Computed[i].IsOne() || srcIsPairingPulling[i].IsOne() || srcIsPairingComputed[i].IsOne() {
			dstIsActive.PushOne()
		}
	}
	dstIsActive.PadAndAssign(run, field.Zero())

	// assign the public inputs for gnark membership check circuit
	if ec.AlignedG2MembershipData != nil {
		ec.flattenLimbsG2Membership.Run(run)
		ec.AlignedG2MembershipData.Assign(run)
	}
	// assign the public inputs for gnark Miller loop circuit
	if ec.AlignedMillerLoopCircuit != nil {
		ec.flattenLimbsMillerLoop.Run(run)
		ec.AlignedMillerLoopCircuit.Assign(run)
	}
	// assign the public inputs for gnark final exponentiation circuit
	if ec.AlignedFinalExpCircuit != nil {
		ec.flattenLimbsFinalExp.Run(run)
		ec.AlignedFinalExpCircuit.Assign(run)
	}
}

func (ec *ECPair) assignPairingData(run *wizard.ProverRuntime) {
	var (
		srcLimbs [common.NbLimbU128][]field.Element

		srcIsPairing = ec.ECPairSource.CsEcpairing.GetColAssignment(run).IntoRegVecSaveAlloc()
		srcIsData    = ec.ECPairSource.IsEcPairingData.GetColAssignment(run).IntoRegVecSaveAlloc()
		srcIsRes     = ec.ECPairSource.IsEcPairingResult.GetColAssignment(run).IntoRegVecSaveAlloc()
		srcID        = ec.ECPairSource.ID.GetColAssignment(run).IntoRegVecSaveAlloc()
	)
	if len(srcIsPairing) != len(srcIsData) || len(srcIsPairing) != len(srcIsRes) {
		utils.Panic("ECPair: input length mismatch")
	}

	for k := range common.NbLimbU128 {
		srcLimbs[k] = ec.ECPairSource.Limbs[k].GetColAssignment(run).IntoRegVecSaveAlloc()

		if len(srcLimbs[k]) != len(srcIsPairing) {
			utils.Panic("ECPair: LIMB_%d input length mismatch", k)
		}
	}

	var (
		inputResult [common.NbLimbU128]ResultLimbs
		pairingInG1 [common.NbLimbU128][]G1Limbs
		pairingInG2 [common.NbLimbU128][]G2Limbs
	)

	var (
		dstLimb [common.NbLimbU128]*common.VectorBuilder

		dstIsActive     = common.NewVectorBuilder(ec.UnalignedPairingData.IsActive)
		dstToMillerLoop = common.NewVectorBuilder(ec.UnalignedPairingData.ToMillerLoopCircuitMask)
		dstToFinalExp   = common.NewVectorBuilder(ec.UnalignedPairingData.ToFinalExpCircuitMask)
		dstIsPulling    = common.NewVectorBuilder(ec.UnalignedPairingData.IsPulling)
		dstIsComputed   = common.NewVectorBuilder(ec.UnalignedPairingData.IsComputed)
		dstPairId       = common.NewVectorBuilder(ec.UnalignedPairingData.PairID)
		dstTotalPairs   = common.NewVectorBuilder(ec.UnalignedPairingData.TotalPairs)
		dstIsFirstLine  = common.NewVectorBuilder(ec.UnalignedPairingData.IsFirstLineOfInstance)
		dstInstanceId   = common.NewVectorBuilder(ec.UnalignedPairingData.InstanceID)
		dstIsAccPrev    = common.NewVectorBuilder(ec.UnalignedPairingData.IsAccumulatorPrev)
		dstIsAccCurr    = common.NewVectorBuilder(ec.UnalignedPairingData.IsAccumulatorCurr)
		dstIsAccInit    = common.NewVectorBuilder(ec.UnalignedPairingData.IsAccumulatorInit)
		dstIndex        = common.NewVectorBuilder(ec.UnalignedPairingData.Index)
		dstIsFirstPrev  = common.NewVectorBuilder(ec.UnalignedPairingData.IsFirstLineOfPrevAccumulator)
		dstIsFirstCurr  = common.NewVectorBuilder(ec.UnalignedPairingData.IsFirstLineOfCurrAccumulator)
		dstIsResult     = common.NewVectorBuilder(ec.UnalignedPairingData.IsResultOfInstance)
	)

	for k := range common.NbLimbU128 {
		dstLimb[k] = common.NewVectorBuilder(ec.UnalignedPairingData.Limbs[k])
	}

	for currPos := 0; currPos < len(srcLimbs[0]); {
		// we need to check if the current position is a pairing or not. If not,
		// then skip it.
		//
		// this also skips inputs to the current pairing instance which do not
		// require passing to the pairing circuit (full-trivial or half-trivial
		// inputs). But this is fine as it doesn't change the result as trivial
		// result ML is 1. Only non-trivial input pairs affect the pairing check
		// result.
		if srcIsPairing[currPos].IsZero() {
			currPos++
			continue
		}
		// when we reach the data, then the following data is for the pairing
		// input data. We iterate over the data to get the number of pairing
		// inputs.
		nbInputs := 1 // we start with 1 because we always have at least one pairing input
		// we may have trivial input pairs in-between the non-trivial ones. We
		// mark which inputs we are interested in.
		actualInputs := []int{0} // we started when isPairing is 1, this means that we have non-trivial input.
		for {
			if srcIsRes[currPos+nbInputs*(nbG1Limbs+nbG2Limbs)].IsOne() {
				break
			}
			if srcIsPairing[currPos+nbInputs*(nbG1Limbs+nbG2Limbs)].IsOne() {
				actualInputs = append(actualInputs, nbInputs)
			}
			nbInputs++
		}
		nbActualTotalPairs := len(actualInputs)
		// now, we have continous chunk of data that is for the pairing. Prepare it for processing.
		for k := range common.NbLimbU128 {
			pairingInG1[k] = make([]G1Limbs, nbActualTotalPairs)
			pairingInG2[k] = make([]G2Limbs, nbActualTotalPairs)
		}

		for ii, i := range actualInputs {
			for j := 0; j < nbG1Limbs; j++ {
				for k := range common.NbLimbU128 {
					pairingInG1[k][ii][j] = srcLimbs[k][currPos+i*(nbG1Limbs+nbG2Limbs)+j]
				}
			}
			for j := 0; j < nbG2Limbs; j++ {
				for k := range common.NbLimbU128 {
					pairingInG2[k][ii][j] = srcLimbs[k][currPos+i*(nbG1Limbs+nbG2Limbs)+nbG1Limbs+j]
				}
			}
		}

		for k := range common.NbLimbU128 {
			inputResult[k][0] = srcLimbs[k][currPos+nbInputs*(nbG1Limbs+nbG2Limbs)]
			inputResult[k][1] = srcLimbs[k][currPos+nbInputs*(nbG1Limbs+nbG2Limbs)+1]
		}
		limbs := processPairingData(pairingInG1, pairingInG2, inputResult)

		limbsLen := len(limbs[0])

		instanceId := srcID[currPos]
		// processed data has the input limbs, but we have entered the intermediate Gt accumulator values

		// generic assignment. We push the static values for the current instance:
		// - the limbs (interleaved with accumulator values)
		// - indicator for the first row of the whole pairingcheck instance
		// - the instance id
		// - the index of the current limb
		// - activeness of the submodule
		for i := 0; i < limbsLen; i++ {
			for k := range common.NbLimbU128 {
				dstLimb[k].PushField(limbs[k][i])
			}

			if i == 0 {
				dstIsFirstLine.PushOne()
			} else {
				dstIsFirstLine.PushZero()
			}
			dstInstanceId.PushField(instanceId)
			dstIndex.PushInt(i)
			dstIsActive.PushOne()
		}
		// now we push the dynamic values per Miller loop. We send all valid
		// pairs except the last to Miller loop (not Miller loop + finalexp!)
		// circuit. Keep in mind if there is only a single valid pair then this
		// loop is skipped.
		for ii := range actualInputs[:len(actualInputs)-1] {
			// first we indicate for the accumulator if it is the first one, previous or current
			for j := 0; j < nbGtLimbs; j++ {
				dstIsComputed.PushOne()
				dstIsPulling.PushZero()
				if ii == 0 {
					// we handle first accumulator separately to be able to
					// constrain that the initial accumulator value is correct.
					dstIsAccInit.PushOne()
					dstIsAccPrev.PushZero()
					dstIsFirstPrev.PushZero()
				} else {
					// we're not in the first pair of points, so we indicate
					// that the accumulator consistency needs to be checked.
					dstIsAccInit.PushZero()
					if j == 0 {
						dstIsFirstPrev.PushOne()
					} else {
						dstIsFirstPrev.PushZero()
					}
					dstIsAccPrev.PushOne()
				}
				dstIsAccCurr.PushZero()
				dstIsFirstCurr.PushZero()
			}
			// now we push the dynamic values for the actual inputs to the pairing check circuit (coming from the arithmetization).
			// essentially we only mark that this limb came directly from arithmetization.
			for j := nbGtLimbs; j < nbGtLimbs+nbG1Limbs+nbG2Limbs; j++ {
				dstIsPulling.PushOne()
				dstIsComputed.PushZero()
				dstIsAccInit.PushZero()
				dstIsAccPrev.PushZero()
				dstIsAccCurr.PushZero()
				dstIsFirstPrev.PushZero()
				dstIsFirstCurr.PushZero()
			}
			// finally, we need to indicate that the next limbs are for the current accumulator
			for j := nbGtLimbs + nbG1Limbs + nbG2Limbs; j < 2*nbGtLimbs+nbG1Limbs+nbG2Limbs; j++ {
				dstIsComputed.PushOne()
				dstIsPulling.PushZero()
				dstIsAccInit.PushZero()
				dstIsAccPrev.PushZero()
				dstIsFirstPrev.PushZero()
				if j == nbGtLimbs+nbG1Limbs+nbG2Limbs {
					dstIsFirstCurr.PushOne()
				} else {
					dstIsFirstCurr.PushZero()
				}
				dstIsAccCurr.PushOne()
			}
			// we also set the static values for all limbs in this pairs of points. These are:
			// - true mark for miller loop circuit
			// - false mark for the ML+finalexp circuit
			// - the pair ID
			// - the total number of pairs
			// - false mark that current limb is for the result
			for j := 0; j < nbG1Limbs+nbG2Limbs+2*nbGtLimbs; j++ {
				dstToMillerLoop.PushOne()
				dstToFinalExp.PushZero()
				dstPairId.PushInt(ii + 1)
				dstTotalPairs.PushInt(nbActualTotalPairs)
				dstIsResult.PushZero()
			}
		}
		// we need to handle the final pair of points separately. The
		// ML+finalexp circuit does not take the current accumulator as an
		// input, but rather the expected pairing check result.
		//
		// first we set the masks for the accumulator limbs.
		for j := 0; j < nbGtLimbs; j++ {
			dstIsComputed.PushOne()
			dstIsPulling.PushZero()
			dstPairId.PushInt(nbActualTotalPairs)
			// handle separately the case when there is only one valid input
			// pair. In this case, the first valid pair also includes the
			// accumulator initialization.
			if nbActualTotalPairs == 1 {
				dstIsAccInit.PushOne()
				dstIsAccPrev.PushZero()
			} else {
				dstIsAccInit.PushZero()
				dstIsAccPrev.PushOne()
			}
			dstIsAccCurr.PushZero()
			if j == 0 {
				// handle separately the case when there is only one valid
				// input. In this case we don't have the previous accumulator.
				if nbActualTotalPairs == 1 {
					dstIsFirstPrev.PushZero()
				} else {
					dstIsFirstPrev.PushOne()
				}
			} else {
				dstIsFirstPrev.PushZero()
			}
			dstIsFirstCurr.PushZero()
			dstIsResult.PushZero()
		}
		// similarly for the final pair of points, we need to indicate that the
		// G1/G2 points come directly from the arithmetization.
		for j := nbGtLimbs; j < nbGtLimbs+nbG1Limbs+nbG2Limbs; j++ {
			dstIsPulling.PushOne()
			dstIsComputed.PushZero()
			dstPairId.PushInt(nbActualTotalPairs)
			dstIsAccInit.PushZero()
			dstIsAccPrev.PushZero()
			dstIsAccCurr.PushZero()
			dstIsFirstPrev.PushZero()
			dstIsFirstCurr.PushZero()
			dstIsResult.PushZero()
		}
		// finally, we need to indicate that the result of the pairing check
		// comes directly from the arithmetization. this is a bit explicit loop,
		// but it's easier to understand.
		for j := nbGtLimbs + nbG1Limbs + nbG2Limbs; j < nbGtLimbs+nbG1Limbs+nbG2Limbs+2; j++ {
			dstIsPulling.PushOne()
			dstIsComputed.PushZero()
			// NB! for the result the Pair ID is 0. This is important to keep in
			// mind as we set some constraints based on this.
			dstPairId.PushZero()
			dstIsAccInit.PushZero()
			dstIsAccPrev.PushZero()
			dstIsAccCurr.PushZero()
			dstIsFirstPrev.PushZero()
			dstIsFirstCurr.PushZero()
			dstIsResult.PushOne()
		}
		// finally we set the static masks for the final pair of points.
		for j := 0; j < nbG1Limbs+nbG2Limbs+nbGtLimbs+2; j++ {
			dstToFinalExp.PushOne()
			dstToMillerLoop.PushZero()
			dstTotalPairs.PushInt(nbActualTotalPairs)
		}
		currPos += nbInputs*(nbG1Limbs+nbG2Limbs) + 2
	}
	// Finally, we pad and assign the assigned data.
	for i := range common.NbLimbU128 {
		dstLimb[i].PadAndAssign(run, field.Zero())
	}
	dstIsActive.PadAndAssign(run, field.Zero())
	dstPairId.PadAndAssign(run, field.Zero())
	dstTotalPairs.PadAndAssign(run, field.Zero())
	dstToMillerLoop.PadAndAssign(run, field.Zero())
	dstToFinalExp.PadAndAssign(run, field.Zero())
	dstIsPulling.PadAndAssign(run, field.Zero())
	dstIsComputed.PadAndAssign(run, field.Zero())
	dstIsFirstLine.PadAndAssign(run, field.Zero())
	dstInstanceId.PadAndAssign(run, field.Zero())
	dstIsAccInit.PadAndAssign(run, field.Zero())
	dstIsAccPrev.PadAndAssign(run, field.Zero())
	dstIsAccCurr.PadAndAssign(run, field.Zero())
	dstIndex.PadAndAssign(run, field.Zero())
	dstIsFirstPrev.PadAndAssign(run, field.Zero())
	dstIsFirstCurr.PadAndAssign(run, field.Zero())
	dstIsResult.PadAndAssign(run, field.Zero())
}

func processPairingData(
	pairingInG1 [common.NbLimbU128][]G1Limbs,
	pairingInG2 [common.NbLimbU128][]G2Limbs,
	inputResult [common.NbLimbU128]ResultLimbs,
) [common.NbLimbU128][]field.Element {
	var res [common.NbLimbU128][]field.Element

	var acc bn254.GT
	acc.SetOne()

	pairingLength := len(pairingInG1[0])
	for i := 0; i < pairingLength; i++ {
		accLimbs := convGtGnarkToWizard(acc)

		var toGnarkG1Limbs [common.NbLimbU128]G1Limbs
		var toGnarkG2Limbs [common.NbLimbU128]G2Limbs

		for j := range common.NbLimbU128 {
			res[j] = append(res[j], accLimbs[j][:]...)
			res[j] = append(res[j], pairingInG1[j][i][:]...)
			res[j] = append(res[j], pairingInG2[j][i][:]...)

			toGnarkG1Limbs[j] = pairingInG1[j][i]
			toGnarkG2Limbs[j] = pairingInG2[j][i]
		}

		inG1 := convG1WizardToGnark(toGnarkG1Limbs)
		inG2 := convG2WizardToGnark(toGnarkG2Limbs)

		// Miller loop with and without line precomputations give different
		// results. As in-circuit we're using the variant with precomputation
		// (but implicitly not as here where we're explicit)
		lines := bn254.PrecomputeLines(inG2)
		mlres, err := bn254.MillerLoopFixedQ(
			[]bn254.G1Affine{inG1},
			[][2][len(bn254.LoopCounter)]bn254.LineEvaluationAff{lines},
		)

		if err != nil {
			utils.Panic("ECPair: failed to compute miller loop: %v", err)
		}

		acc.Mul(&acc, &mlres)
		if i != pairingLength-1 {
			accLimbs = convGtGnarkToWizard(acc)
			for j := range common.NbLimbU128 {
				res[j] = append(res[j], accLimbs[j][:]...)
			}
		}
	}

	for j := range common.NbLimbU128 {
		res[j] = append(res[j], inputResult[j][:]...)
	}

	return res
}

func (ec *ECPair) assignMembershipData(run *wizard.ProverRuntime) {
	var (
		srcLimbs [common.NbLimbU128][]field.Element

		srcIsG2       = ec.ECPairSource.CsG2Membership.GetColAssignment(run).IntoRegVecSaveAlloc()
		srcSuccessBit = ec.ECPairSource.SuccessBit.GetColAssignment(run).IntoRegVecSaveAlloc()
	)
	if len(srcIsG2) != len(srcSuccessBit) {
		utils.Panic("ECPair: input length mismatch")
	}

	for k := range common.NbLimbU128 {
		srcLimbs[k] = ec.ECPairSource.Limbs[k].GetColAssignment(run).IntoRegVecSaveAlloc()

		if len(srcLimbs[k]) != len(srcIsG2) {
			utils.Panic("ECPair: LIMB_%d input length mismatch", k)
		}
	}

	var (
		dstLimb [common.NbLimbU128]*common.VectorBuilder

		dstMask       = common.NewVectorBuilder(ec.UnalignedG2MembershipData.ToG2MembershipCircuitMask)
		dstIsPulling  = common.NewVectorBuilder(ec.UnalignedG2MembershipData.IsPulling)
		dstIsComputed = common.NewVectorBuilder(ec.UnalignedG2MembershipData.IsComputed)
		dstSuccessBit = common.NewVectorBuilder(ec.UnalignedG2MembershipData.SuccessBit)
	)

	for k := range common.NbLimbU128 {
		dstLimb[k] = common.NewVectorBuilder(ec.UnalignedG2MembershipData.Limbs[k])
	}

	for currPos := 0; currPos < len(srcLimbs[0]); {
		if srcIsG2[currPos].IsZero() {
			currPos++
			continue
		}
		// push the G2 limbs
		for i := 0; i < nbG2Limbs; i++ {
			for k := range common.NbLimbU128 {
				dstLimb[k].PushField(srcLimbs[k][currPos+i])
			}

			dstSuccessBit.PushField(srcSuccessBit[currPos])
			dstMask.PushOne()
			dstIsPulling.PushOne()
			dstIsComputed.PushZero()
		}
		// push the expected success bit to the last limb column,
		// while the rest of the columns are zero.
		for k := range common.NbLimbU128 - 1 {
			dstLimb[k].PushZero()
		}
		dstLimb[common.NbLimbU128-1].PushField(srcSuccessBit[currPos])

		dstSuccessBit.PushField(srcSuccessBit[currPos])
		dstMask.PushOne()
		dstIsPulling.PushZero()
		dstIsComputed.PushOne()

		currPos += nbG2Limbs
	}

	for k := range common.NbLimbU128 {
		dstLimb[k].PadAndAssign(run, field.Zero())
	}
	dstSuccessBit.PadAndAssign(run, field.Zero())
	dstMask.PadAndAssign(run, field.Zero())
	dstIsPulling.PadAndAssign(run, field.Zero())
	dstIsComputed.PadAndAssign(run, field.Zero())
}
