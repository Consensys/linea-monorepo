package selector

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/variables"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// Tests that a small table is obtained from subsampling a larger column with a given offset
func CheckSubsample(comp *wizard.CompiledIOP, name string, large, small []ifaces.Column, offset int) {

	// Check that there are as many columns in smalls and in large
	if len(small) != len(large) {
		utils.Panic("small and large should have the same number of columns (%v and %v)", len(small), len(large))
	}

	// Check that there is a non-zero number of columns (testing
	// only in small is enough since we already test that large == small)
	if len(small) == 0 {
		utils.Panic("small is empty")
	}

	numCol := len(small)
	round := 0

	// sanity-check : all the large have the same size and the same for the smalls
	lenSmall := small[0].Size()
	lenLarge := large[0].Size()
	period := lenLarge / lenSmall

	for i := 0; i < numCol; i++ {
		round = utils.Max(round, small[i].Round(), large[i].Round())

		if small[i].Size() != lenSmall {
			utils.Panic("small[%v].Size()!= lenSmall (%v!= %v)", i, small[i].Size(), lenSmall)
		}

		if large[i].Size() != lenLarge {
			utils.Panic("large[%v].Size()!= lenLarge (%v!= %v)", i, large[i].Size(), lenLarge)
		}
	}

	// If the number of columns exceeds one, then we need to
	// add a random coin.
	needGamma := numCol > 1
	var gamma coin.Info

	if needGamma {
		gamma = comp.InsertCoin(round+1, coin.Namef("%v_GAMMA", name), coin.Field)
	}

	alpha := comp.InsertCoin(round+1, coin.Namef("%v_ALPHA", name), coin.Field)

	// Registers the two accumulators
	accSmall := comp.InsertCommit(
		round+1,
		ifaces.ColIDf("%v_ACCUMULATOR_SMALL", name),
		lenSmall,
	)

	accLarge := comp.InsertCommit(
		round+1,
		ifaces.ColIDf("%v_ACCUMULATOR_LARGE", name),
		lenLarge,
	)

	// Also declares the queries on ResAcc and ExpectedResAcc

	r := ifaces.ColumnAsVariable(large[0])
	rPrime := ifaces.ColumnAsVariable(small[0])

	if needGamma {
		// Assign r and rPrime to linear operations
		largeVar := []*symbolic.Expression{}
		smallVar := []*symbolic.Expression{}

		for i := range large {
			largeVar = append(largeVar, ifaces.ColumnAsVariable(large[i]))
			smallVar = append(smallVar, ifaces.ColumnAsVariable(small[i]))
		}

		r = symbolic.NewPolyEval(gamma.AsVariable(), largeVar)
		rPrime = symbolic.NewPolyEval(gamma.AsVariable(), smallVar)
	}

	//
	// Accumulation of the hashes
	//
	// ResAcc[i] = ResAcc[i-1] + Zn,T,−1[(α−1)ResAcc[i-1] + NewState[i]]
	comp.InsertGlobal(
		round+1,
		ifaces.QueryIDf("RESULT_ACCUMULATION_%v", name),
		alpha.AsVariable().
			Sub(symbolic.NewConstant(1)).
			Mul(ifaces.ColumnAsVariable(column.Shift(accLarge, -1))).
			Add(r).
			Mul(variables.NewPeriodicSample(period, offset)).
			Add(ifaces.ColumnAsVariable(column.Shift(accLarge, -1))).
			Sub(ifaces.ColumnAsVariable(accLarge)),
	)

	//
	// Accumulation of the expected results
	//
	// ExpectedResAcc[i] = αExpectedResAcc[i-1] + ExpectedHash[i]
	comp.InsertGlobal(
		round+1,
		ifaces.QueryIDf("SUMSAMPLED_ACCUMULATION_%v", name),
		alpha.AsVariable().
			Mul(ifaces.ColumnAsVariable(column.Shift(accSmall, -1))).
			Add(rPrime).
			Sub(ifaces.ColumnAsVariable(accSmall)),
	)

	//
	// At zero, the ResAcc[i] is equal to 0 if the offset is not zero
	// or equal to r if the offset is 0.
	//
	locExpr := ifaces.ColumnAsVariable(accLarge)
	if offset == 0 {
		locExpr = locExpr.Sub(r)
	}
	comp.InsertLocal(
		round+1,
		ifaces.QueryIDf("RES_ACC_ZERO_%v", name),
		locExpr,
	)

	//
	// At zero, expectedResAcc[i] is equal to expectedHash
	//
	comp.InsertLocal(
		round+1,
		ifaces.QueryIDf("EXPECTED_RES_ACC_ZERO_%v", name),
		ifaces.ColumnAsVariable(accSmall).Sub(rPrime),
	)

	// And we also want to extract the last values of ResAcc and ExpectedResAcc
	accLargeLast := comp.InsertLocalOpening(
		round+1,
		ifaces.QueryIDf("ACC_LARGE_LAST_%v", name),
		column.Shift(accLarge, -1),
	)

	accSmallLast := comp.InsertLocalOpening(
		round+1,
		ifaces.QueryIDf("ACC_SMALL_LAST_%v", name),
		column.Shift(accSmall, -1),
	)

	// And assign them
	comp.SubProvers.AppendToInner(round+1, func(run *wizard.ProverRuntime) {

		r := large[0].GetColAssignment(run)
		if needGamma {
			// Then we need to compute the linear combination explicitly
			largeWit := make([]smartvectors.SmartVector, numCol)
			largeWit[0] = r
			for i := 1; i < numCol; i++ {
				largeWit[i] = large[i].GetColAssignment(run)
			}

			gamma := run.GetRandomCoinField(gamma.Name)
			r = smartvectors.PolyEval(largeWit, gamma)
		}

		// AccLarge
		prev := field.Zero()
		accLargeWit := make([]field.Element, lenLarge)
		alpha_ := run.GetRandomCoinField(alpha.Name)

		for hashID := 0; hashID < lenSmall; hashID++ {
			for i := 0; i < period; i++ {
				pos := hashID*period + i

				// Depending on whether the newstate is at the end
				// of the chunk we compute the next value differently.
				if i != offset {
					// reuse the previous value
					accLargeWit[pos] = prev
					continue
				}

				// prev <- prev * alpha + newState
				currentNewState := r.Get(pos)
				accLargeWit[pos].Mul(&alpha_, &prev)
				accLargeWit[pos].Add(&accLargeWit[pos], &currentNewState)
				prev = accLargeWit[pos]
			}
		}

		run.AssignColumn(accLarge.GetColID(), smartvectors.NewRegular(accLargeWit))
		run.AssignLocalPoint(accLargeLast.ID, prev)

		// rPrime

		rPrime := small[0].GetColAssignment(run)
		if needGamma {
			// Then we need to compute the linear combination explicitly
			smallWit := make([]smartvectors.SmartVector, numCol)
			smallWit[0] = rPrime
			for i := 1; i < numCol; i++ {
				smallWit[i] = small[i].GetColAssignment(run)
			}

			gamma := run.GetRandomCoinField(gamma.Name)
			rPrime = smartvectors.PolyEval(smallWit, gamma)
		}

		accSmallWit := make([]field.Element, lenSmall)
		prev = field.Zero()

		for hashID := 0; hashID < lenSmall; hashID++ {
			// prev <- prev * alpha + newState
			currExpectedHash := rPrime.Get(hashID)
			accSmallWit[hashID].Mul(&alpha_, &prev)
			accSmallWit[hashID].Add(&accSmallWit[hashID], &currExpectedHash)
			prev = accSmallWit[hashID]
		}

		run.AssignColumn(accSmall.GetColID(), smartvectors.NewRegular(accSmallWit))
		run.AssignLocalPoint(accSmallLast.ID, prev)
	})

	comp.InsertVerifier(
		round+1,
		func(run wizard.Runtime) error {
			resAccLast := run.GetLocalPointEvalParams(accLargeLast.ID)
			expectedResAccLast := run.GetLocalPointEvalParams(accSmallLast.ID)
			if resAccLast.Y != expectedResAccLast.Y {
				return fmt.Errorf("linear hashing failed : the ResAcc and ExpectedResAcc do not match on their last inputs %v, %v", resAccLast.Y.String(), expectedResAccLast.Y.String())
			}
			return nil
		},
		func(a frontend.API, run wizard.GnarkRuntime) {
			resAccLast := run.GetLocalPointEvalParams(accLargeLast.ID)
			expectedResAccLast := run.GetLocalPointEvalParams(accSmallLast.ID)
			a.AssertIsEqual(resAccLast.Y, expectedResAccLast.Y)
		},
	)

}
