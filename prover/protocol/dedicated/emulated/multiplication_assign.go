package emulated

import (
	"math/big"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/vectorext"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
)

func (a *Multiplication) assignEmulatedColumns(run *wizard.ProverRuntime) {
	nbRows := a.TermL.Columns[0].Size()
	var (
		dstQuoLimbs = make([][]field.Element, len(a.Quotient.Columns))
		dstRemLimbs = make([][]field.Element, len(a.Result.Columns))
		dstCarry    = make([][]field.Element, len(a.Carry.Columns))
	)
	for i := range dstQuoLimbs {
		dstQuoLimbs[i] = make([]field.Element, nbRows)
	}
	for i := range dstRemLimbs {
		dstRemLimbs[i] = make([]field.Element, nbRows)
	}
	for i := range dstCarry {
		dstCarry[i] = make([]field.Element, nbRows)
	}

	parallel.Execute(nbRows, func(start, end int) {
		bufL := make([]uint64, len(a.TermL.Columns))
		bufR := make([]uint64, len(a.TermR.Columns))
		bufMod := make([]uint64, len(a.Modulus.Columns))
		bufQuo := make([]uint64, len(a.Quotient.Columns))
		bufRem := make([]uint64, len(a.Result.Columns))
		// to compute the carries, we need to perform multiplication on limbs
		bufLhs := make([]uint64, nbMultiplicationResLimbs(len(bufL), len(bufR)))
		bufRhs := make([]uint64, nbMultiplicationResLimbs(len(bufQuo), len(bufMod)))

		witTermL := new(big.Int)
		witTermR := new(big.Int)
		witModulus := new(big.Int)

		tmpProduct := new(big.Int)
		tmpQuotient := new(big.Int)
		tmpRemainder := new(big.Int)
		tmpU64ToBig := new(big.Int)
		carry := new(big.Int)
		for i := start; i < end; i++ {
			// we can reuse all the big ints here
			if err := limbsToBigInt(witTermL, bufL, a.TermL, i, a.nbBitsPerLimb, run); err != nil {
				utils.Panic("failed to convert witness term L: %v", err)
			}
			if err := limbsToBigInt(witTermR, bufR, a.TermR, i, a.nbBitsPerLimb, run); err != nil {
				utils.Panic("failed to convert witness term R: %v", err)
			}
			if err := limbsToBigInt(witModulus, bufMod, a.Modulus, i, a.nbBitsPerLimb, run); err != nil {
				utils.Panic("failed to convert witness modulus: %v", err)
			}
			tmpProduct.Mul(witTermL, witTermR)
			switch {
			case witModulus.Sign() != 0 && tmpProduct.Sign() != 0:
				// we have both nonzero modulus and product.
				tmpQuotient.QuoRem(tmpProduct, witModulus, tmpRemainder)
			case witModulus.Sign() == 0 && tmpProduct.Sign() != 0:
				// modulus is zero, product non zero => invalid
				utils.Panic("modulus cannot be zero when product is non zero")
			default:
				// product is zero, quotient and remainder are zero. We don't
				// need to reset as the values are zeroed already
			}
			if err := bigIntToLimbs(tmpQuotient, bufQuo, a.Quotient, dstQuoLimbs, i, a.nbBitsPerLimb); err != nil {
				utils.Panic("failed to convert quotient to limbs: %v", err)
			}
			if err := bigIntToLimbs(tmpRemainder, bufRem, a.Result, dstRemLimbs, i, a.nbBitsPerLimb); err != nil {
				utils.Panic("failed to convert remainder to limbs: %v", err)
			}
			if err := limbMul(bufLhs, bufL, bufR); err != nil {
				utils.Panic("failed to multiply lhs limbs: %v", err)
			}
			if err := limbMul(bufRhs, bufQuo, bufMod); err != nil {
				utils.Panic("failed to multiply rhs limbs: %v", err)
			}
			for j := range bufRem {
				bufRhs[j] += bufRem[j]
			}

			for j := range dstCarry {
				if j < len(bufLhs) {
					carry.Add(carry, tmpU64ToBig.SetUint64(bufLhs[j]))
				}
				if j < len(bufRhs) {
					carry.Sub(carry, tmpU64ToBig.SetUint64(bufRhs[j]))
				}
				carry.Rsh(carry, uint(a.nbBitsPerLimb))
				dstCarry[j][i].SetBigInt(carry)
			}

			clearBuffer(bufL)
			clearBuffer(bufR)
			clearBuffer(bufMod)
			clearBuffer(bufQuo)
			clearBuffer(bufRem)
			clearBuffer(bufLhs)
			clearBuffer(bufRhs)
			tmpProduct.SetUint64(0)
			tmpQuotient.SetUint64(0)
			tmpRemainder.SetUint64(0)
			carry.SetUint64(0)
		}
	})
	for i := range dstQuoLimbs {
		run.AssignColumn(a.Quotient.Columns[i].GetColID(), smartvectors.NewRegular(dstQuoLimbs[i]))
	}
	for i := range dstRemLimbs {
		run.AssignColumn(a.Result.Columns[i].GetColID(), smartvectors.NewRegular(dstRemLimbs[i]))
	}
	for i := range dstCarry {
		run.AssignColumn(a.Carry.Columns[i].GetColID(), smartvectors.NewRegular(dstCarry[i]))
	}
}

func (a *Multiplication) assignChallengePowers(run *wizard.ProverRuntime) {
	chal := run.GetRandomCoinFieldExt(a.Challenge.Name)
	nbRows := a.ChallengePowers[0].Size()
	var power fext.Element
	power.SetOne()
	for i := range a.ChallengePowers {
		col := vectorext.Repeat(power, nbRows)
		sv := smartvectors.NewRegularExt(col)
		run.AssignColumn(a.ChallengePowers[i].GetColID(), sv)
		power.Mul(&power, &chal)
	}
}
