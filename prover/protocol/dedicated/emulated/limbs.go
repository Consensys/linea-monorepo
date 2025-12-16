package emulated

import (
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// Limbs represents a set of columns used to store the limbs of an emulated
// integer in little-endian order (smallest limb first).
type Limbs struct {
	Columns []ifaces.Column
}

// String returns a string representation of the limbs. Useful for mapping the
// grouped limbs in maps.
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

// NewLimbs creates a new Limbs structure with nbLimbs columns at the given
// round and with the given name as prefix for the column IDs.
func NewLimbs(comp *wizard.CompiledIOP, round int, name string, nbLimbs int, nbRows int) Limbs {
	// TODO: add options for range checking inputs. Right now we expect
	// everything comes range checking
	limbs := Limbs{
		Columns: make([]ifaces.Column, nbLimbs),
	}
	for i := range nbLimbs {
		limbs.Columns[i] = comp.InsertCommit(
			round,
			ifaces.ColIDf("%s_LIMB_%d", name, i),
			utils.NextPowerOfTwo(nbRows),
			true,
		)
	}
	return limbs
}

// csPolyEval constructs a symbolic expression that evaluates the polynomial
// defined by the limbs and the challenge powers.
func csPolyEval(val Limbs, challengePowers []ifaces.Column) *symbolic.Expression {
	// TODO: should store the value in column?
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

// limbsToBigInt recomposes the limbs at loc into res using buf as a temporary
// buffer.
func limbsToBigInt(res *big.Int, buf []uint64, limbs [][]field.Element, loc int, nbBitsPerLimb int) error {
	if res == nil {
		return fmt.Errorf("result not initialized")
	}
	nbLimbs := len(limbs)
	if len(buf) < nbLimbs {
		buf = append(buf, make([]uint64, nbLimbs-len(buf))...)
	}
	for j := range nbLimbs {
		buf[j] = limbs[j][loc].Uint64()
	}
	if err := IntLimbRecompose(buf, nbBitsPerLimb, res); err != nil {
		return err
	}
	return nil
}

// bigIntToLimbs decomposes input into limbs and sets the outputCols at loc
// accordingly.
func bigIntToLimbs(input *big.Int, buf []uint64, limbs Limbs, outputCols [][]field.Element, loc int, nbBitsPerLimb int) error {
	if len(buf) != len(limbs.Columns) {
		return fmt.Errorf("mismatched size between limbs and buffer")
	}
	if err := IntLimbDecompose(input, nbBitsPerLimb, buf); err != nil {
		return fmt.Errorf("failed to decompose big.Int into limbs: %v", err)
	}
	for i := range limbs.Columns {
		outputCols[i][loc].SetUint64(buf[i])
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
func IntLimbDecompose(input *big.Int, nbBits int, res []uint64) error {
	// limb modulus
	if input.BitLen() > len(res)*int(nbBits) {
		return errors.New("decomposed integer does not fit into res")
	}
	base := new(big.Int).Lsh(big.NewInt(1), uint(nbBits))
	base.Sub(base, big.NewInt(1))
	tmp := new(big.Int).Set(input)
	tmp2 := new(big.Int)
	for i := range res {
		tmp2.And(tmp, base)
		res[i] = tmp2.Uint64()
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
func IntLimbRecompose(inputs []uint64, nbBits int, res *big.Int) error {
	if res == nil {
		return errors.New("result not initialized")
	}
	res.SetUint64(0)
	tmp := new(big.Int)
	for i := range inputs {
		res.Lsh(res, uint(nbBits))
		tmp.SetUint64(inputs[len(inputs)-i-1])
		res.Add(res, tmp)
	}
	// we do not mod-reduce here as the result is mod-reduced by the caller if
	// needed. In some places we need non-reduced results.
	return nil
}

// limbMul performs limb multiplication between lhs and rhs storing the result
// in res.
func limbMul(res, lhs, rhs []uint64) error {
	if len(res) != nbMultiplicationResLimbs(len(lhs), len(rhs)) {
		return errors.New("result slice length mismatch")
	}
	clearBuffer(res)
	for i := range lhs {
		for j := range rhs {
			res[i+j] += lhs[i] * rhs[j]
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
func clearBuffer[E interface{ uint64 | *big.Int }, S []E](buf S) {
	switch any(buf).(type) {
	case []*big.Int:
		clearBufferBigInt(any(buf).([]*big.Int))
	case []uint64:
		clearBufferUint64(any(buf).([]uint64))
	default:
		panic("unsupported type")
	}
}

func clearBufferBigInt(buf []*big.Int) {
	for i := range buf {
		buf[i].SetInt64(0)
	}
}

func clearBufferUint64(buf []uint64) {
	for i := range buf {
		buf[i] = 0
	}
}
