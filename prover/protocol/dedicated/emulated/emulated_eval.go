package emulated

import (
	"fmt"
	"math/big"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/vectorext"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated/bigrange"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
)

type EmulatedEvaluationModule struct {
	Terms    [][]Limbs // \sum_i \prod_j Terms[i][j] == 0
	Modulus  Limbs
	Quotient Limbs
	Carry    Limbs

	Challenge       *coin.Info
	ChallengePowers []ifaces.Column

	nbBitsPerLimb int
	round         int
	name          string
	maxTermDegree int
	nbLimbs       int
}

func (a *EmulatedEvaluationModule) assignEmulatedColumns(run *wizard.ProverRuntime) {
	nbRows := a.Terms[0][0].Columns[0].Size()

	var (
		dstQuoLimbs = make([]*common.VectorBuilder, len(a.Quotient.Columns))
		dstCarry    = make([]*common.VectorBuilder, len(a.Carry.Columns))
	)
	for i := range a.Quotient.Columns {
		dstQuoLimbs[i] = common.NewVectorBuilder(a.Quotient.Columns[i])
	}
	for i := range a.Carry.Columns {
		dstCarry[i] = common.NewVectorBuilder(a.Carry.Columns[i])
	}

	bufTerm := make([][][]*big.Int, len(a.Terms))
	for i := range bufTerm {
		bufTerm[i] = make([][]*big.Int, a.maxTermDegree)
		for j := range bufTerm[i] {
			bufTerm[i][j] = make([]*big.Int, a.nbLimbs)
			for k := range bufTerm[i][j] {
				bufTerm[i][j][k] = new(big.Int)
			}
		}
	}
	nbLhsLimbs := a.nbLimbs
	for range a.maxTermDegree - 1 {
		nbLhsLimbs = nbMultiplicationResLimbs(nbLhsLimbs, a.nbLimbs)
	}
	// we allocate LHS twice as we set the result value and we modify the buffer
	bufTermProd1 := make([]*big.Int, nbLhsLimbs)
	for i := range bufTermProd1 {
		bufTermProd1[i] = new(big.Int)
	}
	bufTermProd2 := make([]*big.Int, nbLhsLimbs)
	for i := range bufTermProd2 {
		bufTermProd2[i] = new(big.Int)
	}
	bufLhs := make([]*big.Int, nbLhsLimbs)
	for i := range bufLhs {
		bufLhs[i] = new(big.Int)
	}

	bufMod := make([]*big.Int, a.nbLimbs)
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

	witTerms := make([]*big.Int, a.maxTermDegree)
	for i := range witTerms {
		witTerms[i] = new(big.Int)
	}
	witModulus := new(big.Int)

	tmpEval := new(big.Int)
	tmpQuotient := new(big.Int)
	carry := new(big.Int)
	tmpTermProduct := new(big.Int)
	tmpRemainder := new(big.Int) // only for assigning and checking that the eval result is 0

	for i := range nbRows {
		clearBuffer(bufMod)
		clearBuffer(bufQuo)
		clearBuffer(bufRhs)
		clearBuffer(bufLhs)
		for i := range bufTerm {
			for j := range bufTerm[i] {
				clearBuffer(bufTerm[i][j])
			}
		}
		tmpEval.SetInt64(0)
		// recompose all terms
		for j := range a.Terms {
			clearBuffer(bufTermProd1)
			bufTermProd1[0].SetInt64(1) // multiplication identity
			clearBuffer(bufTermProd2)
			tmpTermProduct.SetInt64(1)
			termNbLimbs := 1
			for k := range a.Terms[j] {
				if err := limbsToBigInt(witTerms[j], bufTerm[j][k], a.Terms[j][k], i, a.nbBitsPerLimb, run); err != nil {
					utils.Panic("failed to convert witness term [%d][%d]: %v", j, k, err)
				}
				tmpTermProduct.Mul(tmpTermProduct, witTerms[j])
				if err := limbMul(bufTermProd2[:nbMultiplicationResLimbs(termNbLimbs, a.nbLimbs)], bufTermProd1[:termNbLimbs], bufTerm[j][k]); err != nil {
					utils.Panic("failed to multiply LHS2 and LHS1: %v", err)
				}
				termNbLimbs = nbMultiplicationResLimbs(termNbLimbs, a.nbLimbs)
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
		if witModulus.Sign() != 0 {
			tmpQuotient.QuoRem(tmpEval, witModulus, tmpRemainder)
			if tmpRemainder.Sign() != 0 {
				utils.Panic("emulated evaluation at row %d: evaluation not divisible by modulus", i)
			}
		} else {
			utils.Panic("modulus cannot be zero")
		}
		if err := bigIntToLimbs(tmpQuotient, bufQuo, a.Quotient, dstQuoLimbs, a.nbBitsPerLimb); err != nil {
			utils.Panic("failed to convert quotient to limbs: %v", err)
		}
		if err := limbMul(bufRhs, bufQuo, bufMod); err != nil {
			utils.Panic("failed to compute quotient * modulus: %v", err)
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
			var f field.Element
			f.SetBigInt(carry)
			dstCarry[j].PushField(f)
		}
	}

	for i := range dstQuoLimbs {
		dstQuoLimbs[i].PadAndAssign(run, field.Zero())
	}
	for i := range dstCarry {
		dstCarry[i].PadAndAssign(run, field.Zero())
	}
}

func (a *EmulatedEvaluationModule) assignChallengePowers(run *wizard.ProverRuntime) {
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
	nbQuoLimbs := (maxTermDegree - 1) * nbLimbs
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
			true,
		)
	}
	nbCarryLimbs := nbMultiplicationResLimbs(nbQuoLimbs, nbLimbs)
	carry := Limbs{
		Columns: make([]ifaces.Column, nbCarryLimbs),
	}
	for i := range carry.Columns {
		carry.Columns[i] = comp.InsertCommit(
			round,
			ifaces.ColIDf("%s_EMUL_EVAL_CARRY_LIMB_%d", name, i),
			nbRows,
			true,
		)
	}
	challenge := comp.InsertCoin(round+1, coin.Namef("%s_EMUL_CHALLENGE", name), coin.FieldExt)
	challengePowers := make([]ifaces.Column, len(carry.Columns))
	for i := range challengePowers {
		challengePowers[i] = comp.InsertCommit(
			round+1,
			ifaces.ColIDf("%s_EMUL_CHALLENGE_POWER_%d", name, i),
			nbRows,
			false,
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
