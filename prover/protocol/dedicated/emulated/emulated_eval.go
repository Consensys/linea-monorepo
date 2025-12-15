package emulated

import (
	"fmt"
	"math/big"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated/bigrange"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
)

type EmulatedEvaluationModule struct {
	Terms    [][]Limbs // \sum_i \prod_j Terms[i][j] == 0
	Modulus  Limbs
	Quotient Limbs
	Carry    Limbs

	Challenge       *coin.Info
	ChallengePowers []ifaces.Column

	// nbBitsPerLimb is the number of bits per limb
	nbBitsPerLimb int
	// round is the maximum round number of the input columns
	round int
	// name of the module
	name string
	// maxTermDegree is the maximum degree of all terms
	maxTermDegree int
	// nbLimbs is the maximum number of limbs seen over terms and modulus
	nbLimbs int
}

func (a *EmulatedEvaluationModule) assignEmulatedColumns(run *wizard.ProverRuntime) {
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
					if err := limbsToBigInt(wit, bufWit[:nbLimbs], a.Terms[j][k], i, a.nbBitsPerLimb, run); err != nil {
						utils.Panic("failed to convert witness term [%d][%d]: %v", j, k, err)
					}
					if wit.Sign() != 0 {
						tmpTermProduct.Mul(tmpTermProduct, wit)
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
				tmpEval.Add(tmpEval, tmpTermProduct)
				for i := range bufLhs {
					bufLhs[i].Add(bufLhs[i], bufTermProd1[i])
				}
			}
			if err := limbsToBigInt(witModulus, bufMod, a.Modulus, i, a.nbBitsPerLimb, run); err != nil {
				utils.Panic("failed to convert witness modulus: %v", err)
			}
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
			if err := bigIntToLimbs(tmpQuotient, bufQuo, a.Quotient, dstQuoLimbs, i, a.nbBitsPerLimb); err != nil {
				utils.Panic("failed to convert quotient to limbs: %v", err)
			}
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

func (a *EmulatedEvaluationModule) assignChallengePowers(run *wizard.ProverRuntime) {
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

func EmulatedEvaluation(comp *wizard.CompiledIOP, name string, nbBitsPerLimb int, modulus Limbs, terms [][]Limbs) *EmulatedEvaluationModule {
	round := 0
	nbRows := modulus.Columns[0].Size()
	maxTermDegree := 0
	nbLimbs := len(modulus.Columns)
	nbRangecheckBits := 16
	nbRangecheckLimbs := (nbBitsPerLimb + nbRangecheckBits - 1) / nbRangecheckBits
	for i := range terms {
		maxTermDegree = max(maxTermDegree, len(terms[i]))
		for j := range terms[i] {
			for k := range terms[i][j].Columns {
				round = max(round, terms[i][j].Columns[k].Round())
			}
		}
	}
	nbQuoBits := 0
	for i := range terms {
		nbTermQuoLimbsBits := len(terms[i][0].Columns) * nbBitsPerLimb
		for j := 1; j < len(terms[i]); j++ {
			nbTermQuoLimbsBits = nbMultiplicationResLimbs(nbTermQuoLimbsBits, len(terms[i][j].Columns)*nbBitsPerLimb)
			nbLimbs = max(nbLimbs, len(terms[i][j].Columns))
		}
		nbQuoBits = max(nbQuoBits, nbTermQuoLimbsBits)
	}
	nbQuoBits += utils.DivCeil(utils.Log2Ceil(len(terms)), nbBitsPerLimb) // add some slack for the addition of terms
	nbCarryBits := nbQuoBits
	nbQuoBits = max(0, nbQuoBits-len(modulus.Columns)*nbBitsPerLimb+1) // we divide by modulus of nbLimbs size
	nbQuoLimbs := utils.DivCeil(nbQuoBits, nbBitsPerLimb)
	for i := range modulus.Columns {
		round = max(round, modulus.Columns[i].Round())
	}

	quotient := Limbs{
		Columns: make([]ifaces.Column, nbQuoLimbs),
	}
	for i := range quotient.Columns {
		quotient.Columns[i] = comp.InsertCommit(
			round,
			ifaces.ColIDf("%s_EMUL_EVAL_QUO_LIMB_%d", name, i),
			nbRows,
		)
	}
	nbCarryLimbs := utils.DivCeil(nbCarryBits, nbBitsPerLimb)
	carry := Limbs{
		Columns: make([]ifaces.Column, nbCarryLimbs),
	}
	for i := range carry.Columns {
		carry.Columns[i] = comp.InsertCommit(
			round,
			ifaces.ColIDf("%s_EMUL_EVAL_CARRY_LIMB_%d", name, i),
			nbRows,
		)
	}
	challenge := comp.InsertCoin(round+1, coin.Namef("%s_EMUL_CHALLENGE", name), coin.Field)
	challengePowers := make([]ifaces.Column, len(carry.Columns))
	for i := range challengePowers {
		challengePowers[i] = comp.InsertCommit(
			round+1,
			ifaces.ColIDf("%s_EMUL_CHALLENGE_POWER_%d", name, i),
			nbRows,
		)
	}

	pa := &EmulatedEvaluationModule{
		Terms:           terms,
		Modulus:         modulus,
		Quotient:        quotient,
		Carry:           carry,
		Challenge:       &challenge,
		ChallengePowers: challengePowers,
		nbBitsPerLimb:   nbBitsPerLimb,
		round:           round,
		name:            name,
		maxTermDegree:   maxTermDegree,
		nbLimbs:         nbLimbs,
	}

	comp.RegisterProverAction(round, &WrappedProverAction{pa.assignEmulatedColumns})
	comp.RegisterProverAction(round+1, &WrappedProverAction{pa.assignChallengePowers})

	for i := range quotient.Columns {
		bigrange.BigRange(
			comp,
			ifaces.ColumnAsVariable(pa.Quotient.Columns[i]), int(nbRangecheckLimbs), nbRangecheckBits,
			fmt.Sprintf("%s_EMUL_QUOTIENT_LIMB_RANGE_%d", name, i),
		)
	}

	pa.csEval(comp)
	csChallengePowers(comp, pa.Challenge, pa.ChallengePowers, round, name)
	return pa
}

func (cs *EmulatedEvaluationModule) csEval(comp *wizard.CompiledIOP) {
	// TODO: should we write the evaluation results in a limb. Then we get smaller-degree polynomials
	uniqueLimbs := make(map[string]*symbolic.Expression)
	for i := range cs.Terms {
		for j := range cs.Terms[i] {
			name := cs.Terms[i][j].String()
			if _, ok := uniqueLimbs[name]; !ok {
				uniqueLimbs[name] = csPolyEval(cs.Terms[i][j], cs.ChallengePowers)
			}
		}
	}
	evalSum := symbolic.NewConstant(0)
	for i := range cs.Terms {
		termProd := symbolic.NewConstant(1)
		for j := range cs.Terms[i] {
			name := cs.Terms[i][j].String()
			termProd = symbolic.Mul(termProd, uniqueLimbs[name])
		}
		evalSum = symbolic.Add(evalSum, termProd)
	}
	modulusEval := csPolyEval(cs.Modulus, cs.ChallengePowers)
	quotientEval := csPolyEval(cs.Quotient, cs.ChallengePowers)
	carryEval := csPolyEval(cs.Carry, cs.ChallengePowers)
	coef := big.NewInt(0).Lsh(big.NewInt(1), uint(cs.nbBitsPerLimb))
	carryCoef := symbolic.Sub(
		symbolic.NewConstant(coef),
		cs.Challenge.AsVariable(),
	)

	carryCoefEval := symbolic.Mul(carryEval, carryCoef)
	qmEval := symbolic.Mul(quotientEval, modulusEval)
	comp.InsertGlobal(
		cs.round+1,
		ifaces.QueryIDf("%s_EMUL_EVAL", cs.name),
		symbolic.Sub(
			evalSum,
			qmEval,
			carryCoefEval,
		),
	)
}
