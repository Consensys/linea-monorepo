package emulated

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated/bigrange"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
)

type Limbs struct {
	Columns []ifaces.Column
}

func (l Limbs) String() string {
	if len(l.Columns) == 0 {
		return "[]"
	}
	names := make([]string, len(l.Columns))
	for i := range l.Columns {
		names[i] = string(l.Columns[i].GetColID())
	}
	return fmt.Sprintf("[%s]", strings.Join(names, ":"))
}

func NewLimbs(comp *wizard.CompiledIOP, round int, name string, nbLimbs int, nbRows int) Limbs {
	limbs := Limbs{
		Columns: make([]ifaces.Column, nbLimbs),
	}
	for i := range nbLimbs {
		limbs.Columns[i] = comp.InsertCommit(
			round,
			ifaces.ColIDf("%s_LIMB_%d", name, i),
			nbRows,
		)
	}
	return limbs
}

type EmulatedMultiplicationModule struct {
	// TODO: use expression board instead, but for now keep it simple

	// TODO: later on we want to do multiple emulated operations in one query, but for now we test with multiplication only
	// Terms [][]Limbs // \sum_i prod_j limbs_{i,j} == 0
	TermL Limbs
	TermR Limbs

	Modulus  Limbs
	Result   Limbs
	Quotient Limbs
	Carry    Limbs

	Challenge       *coin.Info
	ChallengePowers []ifaces.Column

	nbBitsPerLimb int
	round         int
	name          string
}

type WrappedProverAction struct {
	fn func(run *wizard.ProverRuntime)
}

func (a *WrappedProverAction) Run(run *wizard.ProverRuntime) {
	a.fn(run)
}

func (a *EmulatedMultiplicationModule) assignEmulatedColumns(run *wizard.ProverRuntime) {
	// TODO: parallelize
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
	// to compute the carries, we need to perform multiplication on limbs
	bufLhs := make([]*big.Int, nbMultiplicationResLimbs(len(bufL), len(bufR)))
	for i := range bufLhs {
		bufLhs[i] = new(big.Int)
	}
	bufRhs := make([]*big.Int, nbMultiplicationResLimbs(len(bufQuo), len(bufMod)))
	for i := range bufRhs {
		bufRhs[i] = new(big.Int)
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
	carry := new(big.Int)
	for i := range nbRows {
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
		if witModulus.Sign() != 0 {
			tmpQuotient.QuoRem(tmpProduct, witModulus, tmpRemainder)
		} else {
			// TODO: panic?
			utils.Panic("modulus cannot be zero")
		}
		if err := bigIntToLimbs(tmpQuotient, bufQuo, a.Quotient, dstQuoLimbs, a.nbBitsPerLimb); err != nil {
			utils.Panic("failed to convert quotient to limbs: %v", err)
		}
		if err := bigIntToLimbs(tmpRemainder, bufRem, a.Result, dstRemLimbs, a.nbBitsPerLimb); err != nil {
			utils.Panic("failed to convert remainder to limbs: %v", err)
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
		for j := range bufRem {
			if j < len(bufRhs) {
				bufRhs[j].Add(bufRhs[j], bufRem[j])
			} else {
				bufRhs = append(bufRhs, new(big.Int).Set(bufRem[j]))
			}
		}

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

func (a *EmulatedMultiplicationModule) assignChallengePowers(run *wizard.ProverRuntime) {
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

func EmulatedMultiplication(comp *wizard.CompiledIOP, name string, left, right, modulus Limbs, nbBitsPerLimb int) *EmulatedMultiplicationModule {
	// TODO: add options for range checking inputs
	// TODO: add options for including permutation on inputs/outputs
	// TODO: add option to have activator column. When it is given then we can avoid assigning zeros when the multiplication is not active
	// TODO: check all limbs are same width
	round := 0
	nbRows := left.Columns[0].Size()
	nbLimbs := len(modulus.Columns)
	nbRangecheckBits := 16
	nbRangecheckLimbs := (nbBitsPerLimb + nbRangecheckBits - 1) / nbRangecheckBits
	for i := range left.Columns {
		round = max(round, left.Columns[i].Round())
	}
	for i := range right.Columns {
		round = max(round, right.Columns[i].Round())
	}
	for i := range modulus.Columns {
		round = max(round, modulus.Columns[i].Round())
	}

	result := Limbs{
		Columns: make([]ifaces.Column, nbLimbs),
	}
	quotient := Limbs{
		Columns: make([]ifaces.Column, nbLimbs),
	}
	for i := range nbLimbs {
		result.Columns[i] = comp.InsertCommit(
			round,
			ifaces.ColIDf("%s_EMUL_REMAINDER_LIMB_%d", name, i),
			nbRows,
		)
		quotient.Columns[i] = comp.InsertCommit(
			round,
			ifaces.ColIDf("%s_EMUL_QUOTIENT_LIMB_%d", name, i),
			nbRows,
		)
	}
	carry := Limbs{Columns: make([]ifaces.Column, 2*nbLimbs-1)}
	for i := range carry.Columns {
		carry.Columns[i] = comp.InsertCommit(
			round,
			ifaces.ColIDf("%s_EMUL_CARRY_%d", name, i),
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

	pa := &EmulatedMultiplicationModule{
		TermL:           left,
		TermR:           right,
		Modulus:         modulus,
		Result:          result,
		Quotient:        quotient,
		Carry:           carry,
		Challenge:       &challenge,
		ChallengePowers: challengePowers,
		nbBitsPerLimb:   nbBitsPerLimb,
		round:           round,
		name:            name,
	}
	// we need to register prover action already here to ensure it is called
	// before bigrange prover actions
	comp.RegisterProverAction(round, &WrappedProverAction{pa.assignEmulatedColumns})
	comp.RegisterProverAction(round+1, &WrappedProverAction{pa.assignChallengePowers})

	for i := range nbLimbs {
		bigrange.BigRange(
			comp,
			ifaces.ColumnAsVariable(pa.Quotient.Columns[i]), int(nbRangecheckLimbs), nbRangecheckBits,
			fmt.Sprintf("%s_EMUL_QUOTIENT_LIMB_RANGE_%d", name, i),
		)
		bigrange.BigRange(
			comp,
			ifaces.ColumnAsVariable(pa.Result.Columns[i]), int(nbRangecheckLimbs), nbRangecheckBits,
			fmt.Sprintf("%s_EMUL_REMAINDER_LIMB_RANGE_%d", name, i),
		)
	}

	// define the global constraints
	pa.csMultiplication(comp)
	csChallengePowers(comp, pa.Challenge, pa.ChallengePowers, pa.round, pa.name)

	return pa
}

func csChallengePowers(comp *wizard.CompiledIOP, chinfo *coin.Info, challengePowers []ifaces.Column, round int, name string) {
	ch := chinfo.AsVariable()
	comp.InsertGlobal(
		round,
		ifaces.QueryIDf("%s_EMUL_CHALLENGE_POWER_0", name),
		symbolic.Sub(
			challengePowers[0],
			1,
		),
	)
	for i := 1; i < len(challengePowers); i++ {
		comp.InsertGlobal(
			round+1,
			ifaces.QueryIDf("%s_EMUL_CHALLENGE_POWER_CONSISTENCY_%d", name, i),
			symbolic.Sub(
				ifaces.ColumnAsVariable(challengePowers[i]),
				symbolic.Mul(
					ifaces.ColumnAsVariable(challengePowers[i-1]),
					ch,
				),
			),
		)
	}
}

func (cs *EmulatedMultiplicationModule) csMultiplication(comp *wizard.CompiledIOP) {
	leftEval := csPolyEval(cs.TermL, cs.ChallengePowers)
	rightEval := csPolyEval(cs.TermR, cs.ChallengePowers)

	modulusEval := csPolyEval(cs.Modulus, cs.ChallengePowers)
	quotientEval := csPolyEval(cs.Quotient, cs.ChallengePowers)

	resultEval := csPolyEval(cs.Result, cs.ChallengePowers)

	carryEval := csPolyEval(cs.Carry, cs.ChallengePowers)
	coef := big.NewInt(0).Lsh(big.NewInt(1), uint(cs.nbBitsPerLimb))
	carryCoef := symbolic.Sub(
		symbolic.NewConstant(coef),
		cs.Challenge.AsVariable(),
	)

	mulEval := symbolic.Mul(leftEval, rightEval)
	carryCoefEval := symbolic.Mul(carryEval, carryCoef)
	qmEval := symbolic.Mul(modulusEval, quotientEval)

	// Enforce: left * right - modulus * quotient - result - (2^nbits - challenge) * carry = 0
	comp.InsertGlobal(
		cs.round+1,
		ifaces.QueryIDf("%s_EMUL_MULTIPLICATION", cs.name),
		symbolic.Sub(
			mulEval,
			qmEval,
			resultEval,
			carryCoefEval,
		),
	)
}

func csPolyEval(val Limbs, challengePowers []ifaces.Column) *symbolic.Expression {
	// TODO: should store in column?
	res := symbolic.NewConstant(0)
	for i := range val.Columns {
		res = symbolic.Add(
			res,
			symbolic.Mul(
				val.Columns[i],
				challengePowers[i],
			),
		)
	}
	return res
}
