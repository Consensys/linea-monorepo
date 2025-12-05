package emulated

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
)

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
