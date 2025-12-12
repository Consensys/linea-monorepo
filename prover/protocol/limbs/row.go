package limbs

import (
	"math/big"
	"slices"

	"github.com/consensys/linea-monorepo/prover/maths/field"
)

// row represents a single row in the assignment of a limb register. The type
// may only be constructed by converting explicit types.
type row[E Endianness] []field.Element

// NumLimbs returns the number of limbs in the row
func (r row[E]) NumLimbs() int {
	return len(r)
}

// RowFromBigInt converts a big.Int into a row
func RowFromBigInt[E Endianness](bi *big.Int, bitSize int) row[E] {
	return bigIntToLimbs[E](bi, bitSize)
}

// RowFromInt converts an int into a row
func RowFromInt[E Endianness](x int, bitSize int) row[E] {
	return bigIntToLimbs[E](big.NewInt(int64(x)), bitSize)
}

// RowFromBytes converts bytes into a row
func RowFromBytes[E Endianness](x []byte) row[E] {
	return bytesToLimbs[E](x)
}

// ToBigEndian returns the row in big endian form
func (r row[E]) ToBigEndian() row[BigEndian] {
	r = slices.Clone(r)
	if isLittleEndian[E]() {
		slices.Reverse(r)
	}
	return row[BigEndian](r)
}

// ToLittleEndian returns the row in little endian form
func (r row[E]) ToLittleEndian() row[LittleEndian] {
	r = slices.Clone(r)
	if isBigEndian[E]() {
		slices.Reverse(r)
	}
	return row[LittleEndian](r)
}
