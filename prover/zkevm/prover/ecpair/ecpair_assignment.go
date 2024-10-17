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
		ec.AlignedG2MembershipData.Assign(run)
	}
	// assign the public inputs for gnark Miller loop circuit
	if ec.AlignedMillerLoopCircuit != nil {
		ec.AlignedMillerLoopCircuit.Assign(run)
	}
	// assign the public inputs for gnark final exponentiation circuit
	if ec.AlignedFinalExpCircuit != nil {
		ec.AlignedFinalExpCircuit.Assign(run)
	}
}

func (ec *ECPair) assignPairingData(run *wizard.ProverRuntime) {
	var (
		srcIsPairing = ec.ECPairSource.CsEcpairing.GetColAssignment(run).IntoRegVecSaveAlloc()
		srcLimbs     = ec.ECPairSource.Limb.GetColAssignment(run).IntoRegVecSaveAlloc()
		srcIsData    = ec.ECPairSource.IsEcPairingData.GetColAssignment(run).IntoRegVecSaveAlloc()
		srcIsRes     = ec.ECPairSource.IsEcPairingResult.GetColAssignment(run).IntoRegVecSaveAlloc()
		srcID        = ec.ECPairSource.ID.GetColAssignment(run).IntoRegVecSaveAlloc()
	)
	if len(srcIsPairing) != len(srcLimbs) || len(srcIsPairing) != len(srcIsData) || len(srcIsPairing) != len(srcIsRes) {
		utils.Panic("ECPair: input length mismatch")
	}

	var (
		inputResult [2]field.Element
		pairingInG1 [][nbG1Limbs]field.Element
		pairingInG2 [][nbG2Limbs]field.Element
	)

	var (
		dstIsActive     = common.NewVectorBuilder(ec.UnalignedPairingData.IsActive)
		dstLimb         = common.NewVectorBuilder(ec.UnalignedPairingData.Limb)
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
	)

	for currPos := 0; currPos < len(srcLimbs); {
		// we need to check if the current position is a pairing or not. If not, then skip it.
		if srcIsPairing[currPos].IsZero() {
			currPos++
			continue
		}
		// when we reach the data, then the following data is for the pairing
		// input data. We iterate over the data to get the number of pairing
		// inputs.
		nbInputs := 1 // we start with 1 because we always have at least one pairing input
		for {
			if srcIsRes[currPos+nbInputs*(nbG1Limbs+nbG2Limbs)].IsOne() {
				break
			}
			nbInputs++
		}
		// now, we have continous chunk of data that is for the pairing. Prepare it for processing.
		pairingInG1 = make([][nbG1Limbs]field.Element, nbInputs)
		pairingInG2 = make([][nbG2Limbs]field.Element, nbInputs)
		for i := 0; i < nbInputs; i++ {
			for j := 0; j < nbG1Limbs; j++ {
				pairingInG1[i][j] = srcLimbs[currPos+i*(nbG1Limbs+nbG2Limbs)+j]
			}
			for j := 0; j < nbG2Limbs; j++ {
				pairingInG2[i][j] = srcLimbs[currPos+i*(nbG1Limbs+nbG2Limbs)+nbG1Limbs+j]
			}
			inputResult[0] = srcLimbs[currPos+(i+1)*(nbG1Limbs+nbG2Limbs)]
			inputResult[1] = srcLimbs[currPos+(i+1)*(nbG1Limbs+nbG2Limbs)+1]
		}
		limbs := processPairingData(pairingInG1, pairingInG2, inputResult)
		instanceId := srcID[currPos]
		// processed data has the input limbs, but we have entered the intermediate Gt accumulator values
		for i := 0; i < len(limbs); i++ {
			dstLimb.PushField(limbs[i])
			if i == 0 {
				dstIsFirstLine.PushOne()
			} else {
				dstIsFirstLine.PushZero()
			}
			dstInstanceId.PushField(instanceId)
			dstIndex.PushInt(i)
			dstIsActive.PushOne()
		}
		for i := 0; i < nbInputs-1; i++ {
			for j := 0; j < nbGtLimbs; j++ {
				dstIsComputed.PushOne()
				dstIsPulling.PushZero()
				if i == 0 {
					dstIsAccInit.PushOne()
					dstIsAccPrev.PushZero()
					dstIsFirstPrev.PushZero()
				} else {
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
			for j := nbGtLimbs; j < nbGtLimbs+nbG1Limbs+nbG2Limbs; j++ {
				dstIsPulling.PushOne()
				dstIsComputed.PushZero()
				dstIsAccInit.PushZero()
				dstIsAccPrev.PushZero()
				dstIsAccCurr.PushZero()
				dstIsFirstPrev.PushZero()
				dstIsFirstCurr.PushZero()
			}
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
			for j := 0; j < nbG1Limbs+nbG2Limbs+2*nbGtLimbs; j++ {
				dstToMillerLoop.PushOne()
				dstToFinalExp.PushZero()
				dstPairId.PushInt(i + 1)
				dstTotalPairs.PushInt(nbInputs)
			}
		}
		for j := 0; j < nbGtLimbs; j++ {
			dstIsComputed.PushOne()
			dstIsPulling.PushZero()
			dstPairId.PushInt(nbInputs)
			dstIsAccInit.PushZero()
			dstIsAccPrev.PushOne()
			dstIsAccCurr.PushZero()
			if j == 0 {
				dstIsFirstPrev.PushOne()
			} else {
				dstIsFirstPrev.PushZero()
			}
			dstIsFirstCurr.PushZero()
		}
		for j := nbGtLimbs; j < nbGtLimbs+nbG1Limbs+nbG2Limbs; j++ {
			dstIsPulling.PushOne()
			dstIsComputed.PushZero()
			dstPairId.PushInt(nbInputs)
			dstIsAccInit.PushZero()
			dstIsAccPrev.PushZero()
			dstIsAccCurr.PushZero()
			dstIsFirstPrev.PushZero()
			dstIsFirstCurr.PushZero()
		}
		for j := nbGtLimbs + nbG1Limbs + nbG2Limbs; j < nbGtLimbs+nbG1Limbs+nbG2Limbs+2; j++ {
			dstIsPulling.PushOne()
			dstIsComputed.PushZero()
			dstPairId.PushZero()
			dstIsAccInit.PushZero()
			dstIsAccPrev.PushZero()
			dstIsAccCurr.PushZero()
			dstIsFirstPrev.PushZero()
			dstIsFirstCurr.PushZero()
		}
		for j := 0; j < nbG1Limbs+nbG2Limbs+nbGtLimbs+2; j++ {
			dstToFinalExp.PushOne()
			dstToMillerLoop.PushZero()
			dstTotalPairs.PushInt(nbInputs)
		}
		currPos += nbInputs*(nbG1Limbs+nbG2Limbs) + 2
	}
	dstIsActive.PadAndAssign(run, field.Zero())
	dstLimb.PadAndAssign(run, field.Zero())
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
}

func processPairingData(pairingInG1 [][nbG1Limbs]field.Element, pairingInG2 [][nbG2Limbs]field.Element, inputResult [2]field.Element) []field.Element {
	var res []field.Element

	var acc bn254.GT
	acc.SetOne()
	for i := 0; i < len(pairingInG1); i++ {
		accLimbs := convGtGnarkToWizard(acc)
		res = append(res, accLimbs[:]...)
		res = append(res, pairingInG1[i][:]...)
		res = append(res, pairingInG2[i][:]...)
		inG1 := convG1WizardToGnark(pairingInG1[i])
		inG2 := convG2WizardToGnark(pairingInG2[i])

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
		if i != len(pairingInG1)-1 {
			accLimbs = convGtGnarkToWizard(acc)
			res = append(res, accLimbs[:]...)
		}
	}
	res = append(res, inputResult[:]...)
	return res
}

func (ec *ECPair) assignMembershipData(run *wizard.ProverRuntime) {
	var (
		srcIsG2       = ec.ECPairSource.CsG2Membership.GetColAssignment(run).IntoRegVecSaveAlloc()
		srcLimbs      = ec.ECPairSource.Limb.GetColAssignment(run).IntoRegVecSaveAlloc()
		srcSuccessBit = ec.ECPairSource.SuccessBit.GetColAssignment(run).IntoRegVecSaveAlloc()
	)
	if len(srcIsG2) != len(srcLimbs) || len(srcIsG2) != len(srcSuccessBit) {
		utils.Panic("ECPair: input length mismatch")
	}
	var (
		dstLimb       = common.NewVectorBuilder(ec.UnalignedG2MembershipData.Limb)
		dstMask       = common.NewVectorBuilder(ec.UnalignedG2MembershipData.ToG2MembershipCircuitMask)
		dstIsPulling  = common.NewVectorBuilder(ec.UnalignedG2MembershipData.IsPulling)
		dstIsComputed = common.NewVectorBuilder(ec.UnalignedG2MembershipData.IsComputed)
		dstSuccessBit = common.NewVectorBuilder(ec.UnalignedG2MembershipData.SuccessBit)
	)

	for currPos := 0; currPos < len(srcLimbs); {
		if srcIsG2[currPos].IsZero() {
			currPos++
			continue
		}
		// push the G2 limbs
		for i := 0; i < nbG2Limbs; i++ {
			dstLimb.PushField(srcLimbs[currPos+i])
			dstSuccessBit.PushField(srcSuccessBit[currPos])
			dstMask.PushOne()
			dstIsPulling.PushOne()
			dstIsComputed.PushZero()
		}
		// push the expected success bit
		dstLimb.PushField(srcSuccessBit[currPos])
		dstSuccessBit.PushField(srcSuccessBit[currPos])
		dstMask.PushOne()
		dstIsPulling.PushZero()
		dstIsComputed.PushOne()

		currPos += nbG2Limbs
	}

	dstLimb.PadAndAssign(run, field.Zero())
	dstSuccessBit.PadAndAssign(run, field.Zero())
	dstMask.PadAndAssign(run, field.Zero())
	dstIsPulling.PadAndAssign(run, field.Zero())
	dstIsComputed.PadAndAssign(run, field.Zero())
}
