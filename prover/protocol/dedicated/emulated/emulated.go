package emulated

import (
	"math/big"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
)

const round_nr = 0

type Limbs struct {
	Columns []ifaces.Column
}

type EmulatedProverAction struct {
	// TODO: use expression board instead, but for now keep it simple

	// TODO: later on we want to do multiple emulated operations in one query, but for now we test with multiplication only
	// Terms [][]Limbs // \sum_i prod_j limbs_{i,j}
	TermL Limbs
	TermR Limbs

	Modulus  Limbs
	Result   Limbs
	Quotient Limbs
	Carry    Limbs

	Challenge       coin.Info
	ChallengePowers []ifaces.Column

	nbBits uint
}

func (a *EmulatedProverAction) Run(run *wizard.ProverRuntime) {
	a.assignEmulatedColumns(run)
	a.assignChallengePowers(run)
}

func (a *EmulatedProverAction) assignEmulatedColumns(run *wizard.ProverRuntime) {
	nbRows := a.TermL.Columns[0].Size()
	nbLimbs := len(a.TermL.Columns)
	bufL := make([]*big.Int, nbLimbs)
	for i := range bufL {
		bufL[i] = new(big.Int)
	}
	bufR := make([]*big.Int, nbLimbs)
	for i := range bufR {
		bufR[i] = new(big.Int)
	}
	bufMod := make([]*big.Int, len(a.Modulus.Columns))
	for i := range bufMod {
		bufMod[i] = new(big.Int)
	}
	bufQuo := make([]*big.Int, len(a.Quotient.Columns))
	for i := range bufQuo {
		bufQuo[i] = new(big.Int)
	}
	bufRem := make([]*big.Int, len(a.Result.Columns))
	for i := range bufRem {
		bufRem[i] = new(big.Int)
	}

	witTermL := new(big.Int)
	witTermR := new(big.Int)
	witModulus := new(big.Int)
	var (
		dstQuoLimbs = make([]*common.VectorBuilder, len(a.Quotient.Columns))
		dstRemLimbs = make([]*common.VectorBuilder, len(a.Result.Columns))
		dstCarry    = make([]*common.VectorBuilder, len(a.Carry.Columns))
	)
	for i := range dstQuoLimbs {
		dstQuoLimbs[i] = common.NewVectorBuilder(a.Quotient.Columns[i])
	}
	for i := range dstRemLimbs {
		dstRemLimbs[i] = common.NewVectorBuilder(a.Result.Columns[i])
	}
	for i := range dstCarry {
		dstCarry[i] = common.NewVectorBuilder(a.Carry.Columns[i])
	}

	tmpProduct := new(big.Int)
	tmpQuotient := new(big.Int)
	tmpRemainder := new(big.Int)
	tmpCarries := make([]*big.Int, len(a.Carry.Columns))
	for i := range tmpCarries {
		tmpCarries[i] = new(big.Int)
	}
	for i := range nbRows {
		// we can reuse all the big ints here
		if err := a.limbsToBigInt(witTermL, bufL, a.TermL, i, run); err != nil {
			utils.Panic("failed to convert witness term L: %v", err)
		}
		if err := a.limbsToBigInt(witTermR, bufR, a.TermR, i, run); err != nil {
			utils.Panic("failed to convert witness term R: %v", err)
		}
		if err := a.limbsToBigInt(witModulus, bufMod, a.Modulus, i, run); err != nil {
			utils.Panic("failed to convert witness modulus: %v", err)
		}
		tmpProduct.Mul(witTermL, witTermR)
		if witModulus.Sign() != 0 {
			tmpQuotient.QuoRem(tmpProduct, witModulus, tmpRemainder)
		} else {
			// TODO: panic?
			utils.Panic("modulus cannot be zero")
		}
		if err := a.bigIntToLimbs(tmpQuotient, bufQuo, a.Quotient, dstQuoLimbs); err != nil {
			utils.Panic("failed to convert quotient to limbs: %v", err)
		}
		if err := a.bigIntToLimbs(tmpRemainder, bufRem, a.Result, dstRemLimbs); err != nil {
			utils.Panic("failed to convert remainder to limbs: %v", err)
		}
		// to compute the carries, we need to perform multiplication on limbs
		bufLhs := make([]*big.Int, nbMultiplicationResLimbs(len(bufL), len(bufR)))
		for i := range bufLhs {
			bufLhs[i] = new(big.Int)
		}
		bufRhs := make([]*big.Int, nbMultiplicationResLimbs(len(bufQuo), len(bufMod)))
		for i := range bufRhs {
			bufRhs[i] = new(big.Int)
		}
		if err := limbMul(bufLhs, bufL, bufR); err != nil {
			utils.Panic("failed to multiply lhs limbs: %v", err)
		}
		if err := limbMul(bufRhs, bufQuo, bufMod); err != nil {
			utils.Panic("failed to multiply rhs limbs: %v", err)
		}
		// add the remainder to the rhs, it now only has k*p. This is only for very
		// edge cases where by adding the remainder we get additional bits in the
		// carry.
		for i := range bufRem {
			if i < len(bufRhs) {
				bufRhs[i].Add(bufRhs[i], bufRem[i])
			} else {
				bufRhs = append(bufRhs, new(big.Int).Set(bufRem[i]))
			}
		}
		for i := range tmpCarries {
			if i < len(bufLhs) {
				tmpCarries[i].Add(tmpCarries[i], bufLhs[i])
			}
			if i < len(bufRhs) {
				tmpCarries[i].Sub(tmpCarries[i], bufRhs[i])
			}
			tmpCarries[i].Rsh(tmpCarries[i], uint(a.nbBits))
			var f field.Element
			f.SetBigInt(tmpCarries[i])
			dstCarry[i].PushField(f)
		}

		clearBuffer(bufL)
		clearBuffer(bufR)
		clearBuffer(bufMod)
		clearBuffer(bufQuo)
		clearBuffer(bufRem)
		clearBuffer(bufLhs)
		clearBuffer(bufRhs)
		clearBuffer(tmpCarries)
		tmpProduct.SetUint64(0)
		tmpQuotient.SetUint64(0)
		tmpRemainder.SetUint64(0)
	}
	for i := range dstQuoLimbs {
		dstQuoLimbs[i].PadAndAssign(run, field.Zero())
	}
	for i := range dstRemLimbs {
		dstRemLimbs[i].PadAndAssign(run, field.Zero())
	}
	for i := range dstCarry {
		dstCarry[i].PadAndAssign(run, field.Zero())
	}
}

func (a *EmulatedProverAction) assignChallengePowers(run *wizard.ProverRuntime) {
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

func EmulatedMultiplication(comp *wizard.CompiledIOP, left, right, modulus Limbs, nbBits uint) *EmulatedProverAction {
	// TODO: add range checks on hinted limbs
	// TODO: use wizardutils.LastRoundToEval

	nbRows := left.Columns[0].Size()
	nbLimbs := len(modulus.Columns)
	quotient := Limbs{
		Columns: make([]ifaces.Column, nbLimbs),
	}
	remainder := Limbs{
		Columns: make([]ifaces.Column, nbLimbs),
	}
	for i := 0; i < nbLimbs; i++ {
		quotient.Columns[i] = comp.InsertCommit(
			round_nr,
			ifaces.ColIDf("EMUL_QUOTIENT_LIMB_%d", i),
			nbRows,
		)
		remainder.Columns[i] = comp.InsertCommit(
			round_nr,
			ifaces.ColIDf("EMUL_REMAINDER_LIMB_%d", i),
			nbRows,
		)
	}
	carries := Limbs{Columns: make([]ifaces.Column, 2*nbLimbs-1)}
	for i := 0; i < 2*nbLimbs-1; i++ {
		carries.Columns[i] = comp.InsertCommit(
			round_nr,
			ifaces.ColIDf("EMUL_CARRY_%d", i),
			nbRows,
		)
	}
	challenge := comp.InsertCoin(round_nr+1, coin.Namef("EMUL_CHALLENGE"), coin.Field)
	challengePowers := make([]ifaces.Column, len(carries.Columns))
	for i := range challengePowers {
		challengePowers[i] = comp.InsertCommit(
			round_nr+1,
			ifaces.ColIDf("EMUL_CHALLENGE_POWER_%d", i),
			nbRows,
		)
	}

	pa := &EmulatedProverAction{
		TermL:           left,
		TermR:           right,
		Modulus:         modulus,
		Result:          remainder,
		Quotient:        quotient,
		Carry:           carries,
		Challenge:       challenge,
		ChallengePowers: challengePowers,
		nbBits:          nbBits,
	}
	pa.csMultiplication(comp)
	pa.csChallengePowers(comp)
	// comp.RegisterProverAction(round_nr, pa)
	return pa
}

func (cs *EmulatedProverAction) csChallengePowers(comp *wizard.CompiledIOP) {
	ch := cs.Challenge.AsVariable()
	comp.InsertGlobal(
		round_nr,
		"EMUL_CHALLENGE_POWER_0",
		symbolic.Sub(
			cs.ChallengePowers[0],
			1,
		),
	)
	for i := 1; i < len(cs.ChallengePowers); i++ {
		comp.InsertGlobal(
			round_nr+1,
			ifaces.QueryIDf("EMUL_CHALLENGE_POWER_CONSISTENCY_%d", i),
			symbolic.Sub(
				ifaces.ColumnAsVariable(cs.ChallengePowers[i]),
				symbolic.Mul(
					ifaces.ColumnAsVariable(cs.ChallengePowers[i-1]),
					ch,
				),
			),
		)
	}
}

func (cs *EmulatedProverAction) csMultiplication(comp *wizard.CompiledIOP) {
	leftEval := cs.csPolyEval(comp, cs.TermL)
	rightEval := cs.csPolyEval(comp, cs.TermR)

	modulusEval := cs.csPolyEval(comp, cs.Modulus)
	quotientEval := cs.csPolyEval(comp, cs.Quotient)

	resultEval := cs.csPolyEval(comp, cs.Result)

	carryEval := cs.csPolyEval(comp, cs.Carry)
	coef := big.NewInt(0).Lsh(big.NewInt(1), uint(cs.nbBits))
	carryCoef := symbolic.Sub(
		cs.Challenge.AsVariable(),
		symbolic.NewConstant(coef),
	)

	mulEval := symbolic.Mul(leftEval, rightEval)
	carryCoefEval := symbolic.Mul(carryEval, carryCoef)
	qmEval := symbolic.Mul(modulusEval, quotientEval)

	// Enforce: left * right - modulus * quotient - result - (2^nbits - challenge) * carry = 0
	comp.InsertGlobal(
		round_nr+1,
		"EMUL_MULTIPLICATION",
		symbolic.Sub(
			mulEval,
			qmEval,
			resultEval,
			carryCoefEval,
		),
	)
}

func (cs *EmulatedProverAction) csPolyEval(comp *wizard.CompiledIOP, val Limbs) *symbolic.Expression {
	// should write down?
	res := symbolic.NewConstant(0)
	for i := range val.Columns {
		res = symbolic.Add(
			symbolic.Mul(
				val.Columns[i],
				cs.ChallengePowers[i],
			),
		)
	}
	return res
}
