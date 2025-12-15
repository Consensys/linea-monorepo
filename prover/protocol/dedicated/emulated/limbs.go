package emulated

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// limbsToBigInt recomposes the limbs at loc into res using buf as a temporary
// buffer.
func limbsToBigInt(res *big.Int, buf []*big.Int, limbs Limbs, loc int, nbBitsPerLimb int, run *wizard.ProverRuntime) error {
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
	if err := IntLimbRecompose(buf, nbBitsPerLimb, res); err != nil {
		return err
	}
	return nil
}

// bigIntToLimbs decomposes input into limbs and sets the outputCols at loc
// accordingly.
func bigIntToLimbs(input *big.Int, buf []*big.Int, limbs Limbs, outputCols [][]field.Element, loc int, nbBitsPerLimb int) error {
	if len(buf) != len(limbs.Columns) {
		return fmt.Errorf("mismatched size between limbs and buffer")
	}
	if err := IntLimbDecompose(input, nbBitsPerLimb, buf); err != nil {
		return fmt.Errorf("failed to decompose big.Int into limbs: %v", err)
	}
	for i := range limbs.Columns {
		outputCols[i][loc].SetBigInt(buf[i])
	}
	return nil
}

// IntLimbDecompose decomposes the input into buffer as integers of width
// nbBits. The decomposition is stored in res in little-endian order. It errors
// if the decomposition does not fit into res or if res is uninitialized.
//
// The following holds
//
//	input = \sum_{i=0}^{len(res)} res[i] * 2^{nbBits * i}
func IntLimbDecompose(input *big.Int, nbBits int, res []*big.Int) error {
	// limb modulus
	if input.BitLen() > len(res)*int(nbBits) {
		return errors.New("decomposed integer does not fit into res")
	}
	for _, r := range res {
		if r == nil {
			return errors.New("result slice element uninitialized")
		}
	}
	base := new(big.Int).Lsh(big.NewInt(1), uint(nbBits))
	tmp := new(big.Int).Set(input)
	for i := range res {
		res[i].Mod(tmp, base)
		tmp.Rsh(tmp, uint(nbBits))
	}
	return nil
}

// Recompose takes the little-endian limbs in inputs and combines them into res.
// It errors if inputs is uninitialized or zero-length and if the result is
// uninitialized.
//
// The following holds
//
//	res = \sum_{i=0}^{len(inputs)} inputs[i] * 2^{nbBits * i}
func IntLimbRecompose(inputs []*big.Int, nbBits int, res *big.Int) error {
	if res == nil {
		return errors.New("result not initialized")
	}
	res.SetUint64(0)
	for i := range inputs {
		res.Lsh(res, uint(nbBits))
		res.Add(res, inputs[len(inputs)-i-1])
	}
	// we do not mod-reduce here as the result is mod-reduced by the caller if
	// needed. In some places we need non-reduced results.
	return nil
}

// limbMul performs limb multiplication between lhs and rhs storing the result
// in res.
func limbMul(res, lhs, rhs []*big.Int) error {
	tmp := new(big.Int)
	if len(res) != nbMultiplicationResLimbs(len(lhs), len(rhs)) {
		return errors.New("result slice length mismatch")
	}
	clearBuffer(res)
	for i := range lhs {
		for j := range rhs {
			res[i+j].Add(res[i+j], tmp.Mul(lhs[i], rhs[j]))
		}
	}
	return nil
}

// nbMultiplicationResLimbs returns the number of limbs which fit the
// multiplication result.
func nbMultiplicationResLimbs(lenLeft, lenRight int) int {
	res := max(lenLeft+lenRight-1, 0)
	return res
}

// clearBuffer buffers the big.Int slice by setting all elements to zero.
func clearBuffer(buf []*big.Int) {
	for i := range buf {
		buf[i].SetUint64(0)
	}
}
