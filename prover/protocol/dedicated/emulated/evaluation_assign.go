package emulated

import (
	"math/big"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
)

// assignEmulatedColumns assigns the values to the columns used in the emulated evaluation
func (a *Evaluation) assignEmulatedColumns(run *wizard.ProverRuntime) {
	nbRows := a.Terms[0][0].Columns[0].Size()

	var (
		dstQuoLimbs = make([][]field.Element, len(a.Quotient.Columns))
		dstCarry    = make([][]field.Element, len(a.Carry.Columns))
	)
	for i := range a.Quotient.Columns {
		dstQuoLimbs[i] = make([]field.Element, nbRows)
	}
	for i := range a.Carry.Columns {
		dstCarry[i] = make([]field.Element, nbRows)
	}

	parallel.Execute(nbRows, func(start, end int) {
		// initialize all buffers to avoid reallocations
		bufWit := make([]*big.Int, a.nbLimbs)
		for i := range bufWit {
			bufWit[i] = new(big.Int)
		}
		// we allocate LHS twice as we set the result value and we modify the buffer
		bufTermProd1 := make([]*big.Int, len(a.Carry.Columns))
		for i := range bufTermProd1 {
			bufTermProd1[i] = new(big.Int)
		}
		bufTermProd2 := make([]*big.Int, len(a.Carry.Columns))
		for i := range bufTermProd2 {
			bufTermProd2[i] = new(big.Int)
		}
		bufLhs := make([]*big.Int, len(a.Carry.Columns))
		for i := range bufLhs {
			bufLhs[i] = new(big.Int)
		}

		bufMod := make([]*big.Int, len(a.Modulus.Columns))
		for i := range bufMod {
			bufMod[i] = new(big.Int)
		}
		bufQuo := make([]*big.Int, len(a.Quotient.Columns))
		for i := range bufQuo {
			bufQuo[i] = new(big.Int)
		}
		bufRhs := make([]*big.Int, nbMultiplicationResLimbs(len(bufQuo), len(bufMod)))
		for i := range bufRhs {
			bufRhs[i] = new(big.Int)
		}

		wit := new(big.Int)
		witModulus := new(big.Int)

		tmpEval := new(big.Int)
		tmpQuotient := new(big.Int)
		carry := new(big.Int)
		tmpTermProduct := new(big.Int)
		tmpRemainder := new(big.Int) // only for assigning and checking that the eval result is 0

		for i := start; i < end; i++ {
			// clear all buffers
			clearBuffer(bufMod)
			clearBuffer(bufQuo)
			clearBuffer(bufRhs)
			clearBuffer(bufLhs)
			clearBuffer(bufWit)
			tmpEval.SetInt64(0)
			// recompose all terms
			for j := range a.Terms {
				clearBuffer(bufTermProd1)
				bufTermProd1[0].SetInt64(1) // multiplication identity
				clearBuffer(bufTermProd2)
				tmpTermProduct.SetInt64(1)
				termNbLimbs := 1
				for k := range a.Terms[j] {
					nbLimbs := len(a.Terms[j][k].Columns)
					// recompose the term value as integer
					if err := limbsToBigInt(wit, bufWit[:nbLimbs], a.Terms[j][k], i, a.nbBitsPerLimb, run); err != nil {
						utils.Panic("failed to convert witness term [%d][%d]: %v", j, k, err)
					}
					if wit.Sign() != 0 {
						// perform integer multiplication
						tmpTermProduct.Mul(tmpTermProduct, wit)
						// perform limb multiplication
						if err := limbMul(bufTermProd2[:nbMultiplicationResLimbs(termNbLimbs, nbLimbs)], bufTermProd1[:termNbLimbs], bufWit[:nbLimbs]); err != nil {
							utils.Panic("failed to multiply LHS2 and LHS1: %v", err)
						}
					} else {
						// when one term is zero, then the whole term is zero
						tmpTermProduct.SetInt64(0)
						clearBuffer(bufTermProd2)
					}
					termNbLimbs = nbMultiplicationResLimbs(termNbLimbs, nbLimbs)
					bufTermProd2, bufTermProd1 = bufTermProd1, bufTermProd2
				}
				// accumulate the term into the evaluation as integer
				tmpEval.Add(tmpEval, tmpTermProduct)
				// accumulate the term into the evaluation as limbs
				for i := range bufLhs {
					bufLhs[i].Add(bufLhs[i], bufTermProd1[i])
				}
			}
			// recompose the modulus as integer
			if err := limbsToBigInt(witModulus, bufMod, a.Modulus, i, a.nbBitsPerLimb, run); err != nil {
				utils.Panic("failed to convert witness modulus: %v", err)
			}
			// compute the quotient from the accumulated evaluation. NB! We require the evaluation
			// value to be zero.
			switch {
			case witModulus.Sign() != 0 && tmpEval.Sign() != 0:
				// we have both nonzero modulus and eval.
				tmpQuotient.QuoRem(tmpEval, witModulus, tmpRemainder)
				if tmpRemainder.Sign() != 0 {
					utils.Panic("emulated evaluation at row %d: evaluation not divisible by modulus", i)
				}
			case witModulus.Sign() == 0 && tmpEval.Sign() != 0:
				// modulus is zero, eval non zero => invalid
				utils.Panic("modulus cannot be zero when evaluation is non zero")
			default:
				// eval is zero, quotient and remainder are zero. We don't
				// need to reset as the values are zeroed already
			}
			// assign the quotient limbs from the computed quotient
			if err := bigIntToLimbs(tmpQuotient, bufQuo, a.Quotient, dstQuoLimbs, i, a.nbBitsPerLimb); err != nil {
				utils.Panic("failed to convert quotient to limbs: %v", err)
			}
			// compute the carry limbs from the difference
			if tmpQuotient.Sign() != 0 && witModulus.Sign() != 0 {
				if err := limbMul(bufRhs, bufQuo, bufMod); err != nil {
					utils.Panic("failed to compute quotient * modulus: %v", err)
				}
			}
			carry.SetInt64(0)
			for j := range dstCarry {
				if j < len(bufLhs) {
					carry.Add(carry, bufLhs[j])
				}
				if j < len(bufRhs) {
					carry.Sub(carry, bufRhs[j])
				}
				carry.Rsh(carry, uint(a.nbBitsPerLimb))
				dstCarry[j][i].SetBigInt(carry)
			}
		}
	})

	for i := range dstQuoLimbs {
		run.AssignColumn(a.Quotient.Columns[i].GetColID(), smartvectors.NewRegular(dstQuoLimbs[i]))
	}
	for i := range dstCarry {
		run.AssignColumn(a.Carry.Columns[i].GetColID(), smartvectors.NewRegular(dstCarry[i]))
	}
}

func (a *Evaluation) assignChallengePowers(run *wizard.ProverRuntime) {
	chal := run.GetRandomCoinField(a.Challenge.Name)
	nbRows := a.ChallengePowers[0].Size()
	var power field.Element
	power.SetOne()
	for i := range a.ChallengePowers {
		col := vector.Repeat(power, nbRows)
		sv := smartvectors.NewRegular(col)
		run.AssignColumn(a.ChallengePowers[i].GetColID(), sv)
		power.Mul(&power, &chal)
	}
}
