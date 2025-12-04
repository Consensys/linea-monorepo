package emulated

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
)

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

	Challenge       ifaces.Column
	ChallengePowers []ifaces.Column

	nbBits uint
}

func (a *EmulatedProverAction) limbsToBigInt(res *big.Int, buf []*big.Int, limbs Limbs, loc int, run *wizard.ProverRuntime) error {
	if res == nil {
		return fmt.Errorf("result not initialized")
	}
	nbLimbs := len(limbs.Columns)
	if len(buf) < nbLimbs {
		buf = append(buf, make([]*big.Int, nbLimbs-len(buf))...)
	}
	for i := range buf {
		if buf[i] == nil {
			buf[i] = new(big.Int)
		}
	}
	for j := range nbLimbs {
		limb := limbs.Columns[j].GetColAssignmentAt(run, loc)
		limb.BigInt(buf[j])
	}
	if err := recompose(buf, a.nbBits, res); err != nil {
		return err
	}
	return nil
}

func (a *EmulatedProverAction) bigIntToLimbs(input *big.Int, buf []*big.Int, limbs Limbs, vb []*common.VectorBuilder) error {
	if len(buf) != len(limbs.Columns) {
		return fmt.Errorf("mismatched size between limbs and buffer")
	}
	if err := decompose(input, a.nbBits, buf); err != nil {
		return fmt.Errorf("failed to decompose big.Int into limbs: %v", err)
	}
	for i := range limbs.Columns {
		var f field.Element
		f.SetBigInt(buf[i])
		vb[i].PushField(f)
	}
	return nil
}

func (a *EmulatedProverAction) Run(run *wizard.ProverRuntime) {
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
			fmt.Println("carry limb ", i, ": ", tmpCarries[i].String())
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

// Recompose takes the limbs in inputs and combines them into res. It errors if
// inputs is uninitialized or zero-length and if the result is uninitialized.
//
// The following holds
//
//	res = \sum_{i=0}^{len(inputs)} inputs[i] * 2^{nbBits * i}
func recompose(inputs []*big.Int, nbBits uint, res *big.Int) error {
	if res == nil {
		return errors.New("result not initialized")
	}
	res.SetUint64(0)
	for i := range inputs {
		res.Lsh(res, nbBits)
		res.Add(res, inputs[len(inputs)-i-1])
	}
	// we do not mod-reduce here as the result is mod-reduced by the caller if
	// needed. In some places we need non-reduced results.
	return nil
}

// Decompose decomposes the input into res as integers of width nbBits. It
// errors if the decomposition does not fit into res or if res is uninitialized.
//
// The following holds
//
//	input = \sum_{i=0}^{len(res)} res[i] * 2^{nbBits * i}
func decompose(input *big.Int, nbBits uint, res []*big.Int) error {
	// limb modulus
	if input.BitLen() > len(res)*int(nbBits) {
		return errors.New("decomposed integer does not fit into res")
	}
	for _, r := range res {
		if r == nil {
			return errors.New("result slice element uninitialized")
		}
	}
	base := new(big.Int).Lsh(big.NewInt(1), nbBits)
	tmp := new(big.Int).Set(input)
	for i := 0; i < len(res); i++ {
		res[i].Mod(tmp, base)
		tmp.Rsh(tmp, nbBits)
	}
	return nil
}

func limbMul(res, lhs, rhs []*big.Int) error {
	tmp := new(big.Int)
	if len(res) != nbMultiplicationResLimbs(len(lhs), len(rhs)) {
		return errors.New("result slice length mismatch")
	}
	for i := 0; i < len(lhs); i++ {
		for j := 0; j < len(rhs); j++ {
			res[i+j].Add(res[i+j], tmp.Mul(lhs[i], rhs[j]))
		}
	}
	return nil
}

// nbMultiplicationResLimbs returns the number of limbs which fit the
// multiplication result.
func nbMultiplicationResLimbs(lenLeft, lenRight int) int {
	res := lenLeft + lenRight - 1
	if res < 0 {
		res = 0
	}
	return res
}

func clearBuffer(buf []*big.Int) {
	for i := range buf {
		buf[i].SetUint64(0)
	}
}

func EmulatedMultiplication(comp *wizard.CompiledIOP, left, right, modulus Limbs, nbBits uint) *EmulatedProverAction {
	// TODO: add range checks on hinted limbs
	// TODO: use wizardutils.LastRoundToEval
	const round_nr = 0
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

	pa := &EmulatedProverAction{
		TermL:    left,
		TermR:    right,
		Modulus:  modulus,
		Result:   remainder,
		Quotient: quotient,
		Carry:    carries,
		nbBits:   nbBits,
	}
	pa.csMultiplication(comp)
	// comp.RegisterProverAction(round_nr, pa)
	return pa
}

func (cs *EmulatedProverAction) csMultiplication(comp *wizard.CompiledIOP) {

}
